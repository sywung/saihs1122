## 臺北市立松山工農112學年度第二學期-智慧居家監控實習學習單

<center><font size=6>【Wifi 上傳資料】</font></center>

<div style="text-align: right">班級：______________ 座號：________姓名：________________</div>



esp32 利用 http post 傳送資料控制網頁上LED. 

``` c
#include <WiFi.h>
#include <HTTPClient.h>

const char* ssid = "REPLACE_WITH_YOUR_SSID";
const char* password = "REPLACE_WITH_YOUR_PASSWORD";

//Your Domain name with URL path or IP address with path
String serverName = "http://192.168.221.71:8032/api/";

void setup() {
  Serial.begin(115200); 

  WiFi.begin(ssid, password);
  Serial.println("Connecting");
  while(WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("");
  Serial.print("Connected to WiFi network with IP Address: ");
  Serial.println(WiFi.localIP());
 
  Serial.println("Timer set to 5 seconds (timerDelay variable), it will take 5 seconds before publishing the first reading.");
}

void loop() {
  //Send an HTTP POST request every 10 minutes
  if ((millis() - lastTime) > timerDelay) {
    //Check WiFi connection status
    if(WiFi.status()== WL_CONNECTED){
      HTTPClient http;

      String serverPath = serverName + "ring-led/1";
      
      // Your Domain name with URL path or IP address with path
      http.begin(serverPath.c_str());
      
      // Send HTTP POST request
      http.addHeader("Content-Type", "application/json");

      int httpResponseCode = http.POST("{ \"red\":255, \"green\":0, \"blue\":0 }");
      
      if (httpResponseCode>0) {
        Serial.print("HTTP Response code: ");
        Serial.println(httpResponseCode);
        String payload = http.getString();
        Serial.println(payload);
      }
      else {
        Serial.print("Error code: ");
        Serial.println(httpResponseCode);
      }
      // Free resources
      http.end();
    }
    else {
      Serial.println("WiFi Disconnected");
    }
    lastTime = millis();
  }
}
```



打包 json 資料
``` c
JsonDocument doc;

doc["red"] = "gps";
doc["time"] = 1351824120;
doc["data"][0] = 48.756080;
doc["data"][1] = 2.302038;

serializeJson(doc, Serial);
// This prints:
// {"sensor":"gps","time":1351824120,"data":[48.756080,2.302038]}


doc["red"] = 255;
doc["green"] = 0;
doc["blue"] = 0;

// { "red":255, "green":0, "blue":0 }

```

### 自我練習
1. 用 esp32 控制網頁上環形 LED ，成為單一LED 的跑馬燈。

2. 用 esp32 控制網頁上環形 LED ，成為廣告燈。
