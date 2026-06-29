package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	// "github.com/davecgh/go-spew/spew"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var f embed.FS

type DHTData struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

type Pixel struct {
	R int `json:"red"`
	G int `json:"green"`
	B int `json:"blue"`
}

type LEDControl struct {
	Color []Pixel `json:"color"`
}

type ButtonStatus struct {
	Button1 bool `json:"button1"`
	Button2 bool `json:"button2"`
}

type LEDStatus struct {
	LED bool `json:"led"`
}

var (
	Host         string
	dhtReading   DHTData
	ledControl   LEDControl
	buttonStatus ButtonStatus
	ledStatus    LEDStatus
	enableLog    bool
)

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d%02d/%02d", year, month, day)
}

func init() {
	// 定義命令列旗標
	flag.BoolVar(&enableLog, "enablelog", false, "enable connection logs")
	// 解析命令列參數
	flag.Parse()
}

func main() {
	println("\nStarting ...")
	initializeLEDControl()
	getip()
	gin.SetMode("release")

	if enableLog {
		// 設定連線日誌輸出
		gin.DefaultWriter = os.Stdout
		// 顯示日誌
		log.Println("Connection logs enabled")
	} else {
		// 停用 GIN 的 Logger，將日誌輸出導向 io.Discard
		gin.DefaultWriter = io.Discard
	}

	r := gin.Default()
	// r.Use(gin.LoggerWithWriter(io.Discard))
	// r.Use(gin.Recovery())

	var funcMap = template.FuncMap{
		"uHTML":        func(s string) template.HTML { return template.HTML(s) },
		"unHTML":       func(s string) string { return s },
		"formatAsDate": formatAsDate,
	}

	templ := template.Must(template.New("").Funcs(funcMap).ParseFS(f, "templates/*.htm"))
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.SetHTMLTemplate(templ)
	r.GET("/", func(c *gin.Context) {
		urlprefix := "http://" + Host + ":8032/api/"
		dht := "<a href=\"" + urlprefix + "dht\"  target=\"_dht\">" + urlprefix + "dht</a>"
		led := "<a href=\"" + urlprefix + "ring-led/1\" target=\"_led\">" + urlprefix + "ring-led/1</a>"
		ringLED := "<a href=\"" + urlprefix + "ring-led\" target=\"_ringled\">" + urlprefix + "ring-led</a>"
		c.HTML(http.StatusOK, "index.htm", gin.H{"URL": urlprefix, "DHT": dht, "ringLED": ringLED, "LED": led})
	})

	api := r.Group("/api")
	{
		api.POST("/ping", ping)
		api.GET("/ping", ping)
		api.GET("/ip", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ip": Host})
		})

		// Endpoint to get DHT11 data
		api.GET("/dht", func(c *gin.Context) {
			c.JSON(http.StatusOK, dhtReading)
		})

		// Endpoint to set DHT11 data
		api.POST("/dht", func(c *gin.Context) {
			if err := c.ShouldBindJSON(&dhtReading); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "DHT data updated successfully"})
		})

		// Endpoint to get button status data
		api.GET("/button", func(c *gin.Context) {
			c.JSON(http.StatusOK, buttonStatus)
		})

		// Endpoint to set button status data
		api.POST("/button", func(c *gin.Context) {
			if err := c.ShouldBindJSON(&buttonStatus); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Button status data updated successfully"})
		})

		// 路由處理函式，接收 JSON 資料並綁定至 ledControl 結構
		api.POST("/ring-led", func(c *gin.Context) {
			var receivedLEDControl LEDControl
			if err := c.ShouldBindJSON(&receivedLEDControl); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// 更新全域 ledControl
			ledControl = receivedLEDControl

			c.JSON(http.StatusOK, gin.H{"message": "ring-LED control updated successfully"})
		})

		// 路由處理函式，回傳目前 ledControl 狀態
		api.GET("/ring-led", func(c *gin.Context) {
			c.JSON(http.StatusOK, ledControl)
		})

		api.GET("/ring-led/:n", func(c *gin.Context) {
			n, err := strconv.ParseInt(c.Param("n"), 10, 0)
			if n >= 0 && n < 16 && err == nil {
				c.JSON(http.StatusOK, ledControl.Color[n])
			} else {
				c.JSON(http.StatusOK, gin.H{"message": "ERROR!!"})
			}
		})

		//
		api.GET("/ring-led/:n/:r/:g/:b", func(c *gin.Context) {
			n, err := strconv.ParseInt(c.Param("n"), 10, 0)
			if err != nil || n < 0 || n >= 16 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "ERROR!!"})
				return
			}

			r, err := strconv.ParseInt(c.Param("r"), 10, 0)
			if err != nil || r < 0 || r >= 256 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "ERROR!!"})
				return
			}

			g, err := strconv.ParseInt(c.Param("g"), 10, 0)
			if err != nil || g < 0 || g >= 256 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "ERROR!!"})
				return
			}

			b, err := strconv.ParseInt(c.Param("b"), 10, 0)
			if err != nil || b < 0 || b >= 256 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "ERROR!!"})
				return
			}

			ledControl.Color[n].R = int(r)
			ledControl.Color[n].G = int(g)
			ledControl.Color[n].B = int(b)
			c.JSON(http.StatusOK, gin.H{"message": "ring-LED control updated successfully"})
		})

		//
		api.POST("/ring-led/:n", func(c *gin.Context) {
			n, err := strconv.ParseInt(c.Param("n"), 10, 0)
			if err != nil || n < 0 || n >= 16 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "ERROR!!"})
				return
			}

			var receivedLED Pixel
			if err := c.ShouldBindJSON(&receivedLED); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			r := receivedLED.R
			if r < 0 || r >= 256 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "ERROR!!"})
				return
			}

			g := receivedLED.G
			if g < 0 || g >= 256 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "ERROR!!"})
				return
			}

			b := receivedLED.B
			if b < 0 || b >= 256 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "ERROR!!"})
				return
			}

			ledControl.Color[n].R = r
			ledControl.Color[n].G = g
			ledControl.Color[n].B = b
			c.JSON(http.StatusOK, gin.H{"message": "ring-LED control updated successfully"})
		})

	}
	println("\nFor Service ...")
	// Run the server
	r.Run(":8032")
	println("\nEnd!!\n")
}

func initializeLEDControl() {
	ledControl.Color = make([]Pixel, 16)
	for i := range ledControl.Color {
		ledControl.Color[i] = Pixel{R: 0, G: 0, B: 0}
	}
	ledStatus.LED = false
}

// ping godoc
// @Summary 測試反應時間
// @Success 200 {string} json "{"ok": boolen, "data":{"elapsed": int}}"
// @Router /ping [get]
func ping(c *gin.Context) {
	start := time.Now().UnixNano()
	end := time.Now().UnixNano()
	elapsed := end - start
	resp := map[string]interface{}{"ok": true, "method": c.Request.Method, "data": map[string]interface{}{"elapsed": elapsed}}
	c.JSON(http.StatusOK, resp)
}

func getip() {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, address := range addrs {

		// 檢查 IP 位址是否為迴環位址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				Host = ipnet.IP.String()
				//fmt.Println(ipnet.IP.String())
				return
			}

		}
	}

}
