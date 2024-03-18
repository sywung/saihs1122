## 臺北市立松山工農112學年度第二學期-智慧居家監控實習學習單

<center><font size=6>【Wifi 連線】</font></center>

<div style="text-align: right">班級：______________ 座號：________姓名：________________</div>

#### 一、ESP32 Wifi 介紹

在 ESP32 無線網路使用到的 WiFi.h 函式庫，WiFi 可以選擇四種模式

|名稱 | 說明 | 語法|
|-|-|-|
|WIFI_STA | 以工作站(Station)模式啟動，用來 上網讀取資料，此為預設模式| WiFi.mode(WIFI_STA); 
|WIFI_AP|以熱點(Access Point)模式啟動，讓 其他裝置連入 ESP32|WiFi.mode(AP);
|WIFI_AP_STA|混合模式，同時當熱點也當作工作站|WiFi.mode(WIFI_AP_STA);
|WIFI_OFF|關閉網路，也可於網路不正常時，重啟網路時使用|WiFi.mode(WIFI_OFF);|

本節將使用的是 WIFI_STA 模式，讓 ESP32 就像是一台手機，能讀取網路上的資料，或者將資料傳到某個網站，設定好模式後，可以利用 WiFi.


#### 二、Wifi 函式庫

WiFi.mode(WIFI_STA);

WiFi.begin(ssid, password);

WiFi.status();

Value | Constant | Meaning
-|-|-
0 | WL_IDLE_STATUS | temporary status assigned when WiFi.begin() is called
1 | WL_NO_SSID_AVAIL | when no SSID are available
2 | WL_SCAN_COMPLETED | scan networks is completed
3 | WL_CONNECTED | when connected to a WiFi network
4 | WL_CONNECT_FAILED | when the connection fails for all the attempts
5 | WL_CONNECTION_LOST | when the connection is lost
6 | WL_DISCONNECTED | when disconnected from a network

WiFi.localIP()

WiFi.RSSI()

WiFi.reconnect()

WiFi.disconnect();


#### 三、範例

scanNetworks() 指令掃描附 近的無線網路，除了顯示無線網路的名稱 SSID 之外，也會顯示訊號強度 RSSI，RSSI 是 負數表示，越接近 0 代表訊號越強，另外就是有設定密碼的則會標示「*」。

``` c {.line-numbers}
#include "WiFi.h"

void setup() {
  Serial.begin(115200);
  // 設定網路模式為 STA 工作站模式
  WiFi.mode(WIFI_STA);
  WiFi.disconnect();
  delay(100);
  Serial.println("Setup done");
}

void loop() {
  Serial.println("scan start");
  // 開始掃描附近網路
  int n = WiFi.scanNetworks();
  Serial.println("scan done");
  if (n == 0) {
    Serial.println("no networks found");
  } else {
    ial.print(n);
    Serial.println("networks found");
    // 開始印出所有網路
    for (int i = 0; i < n; ++i) {
      // Print SSID and RSSI for each network found
      Serial.print(i + 1);
      Serial.print(":");
      Serial.print(WiFi.SSID(i));
      Serial.print("(");
      Serial.print(WiFi.RSSI(i));
      Serial.print(")");
      Serial.println((WiFi.encryptionType(i) == WIFI_AUTH_OPEN) ? " " : "*");
      delay(10);
    }
  }
  Serial.println("");
  // 暫停 5 秒
  delay(5000);
}
```


#### 四、自我練習


