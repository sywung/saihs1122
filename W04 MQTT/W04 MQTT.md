## 臺北市立松山工農112學年度第二學期-智慧居家監控實習學習單

<center><font size=6>【MQTT】</font></center>

<div style="text-align: right">班級：______________ 座號：________姓名：________________</div>


#### MQTT explorer
請安裝 電腦端程式
https://mqtt-explorer.com/


#### 相關資料

```
先安裝 library Nick O'Leary 的 pubsubclient
使用 DHT11 須再安裝 SimpleDHT

程式先包含 library:
#include <WiFi.h>
#include <PubSubClient.h>

broker (MQTTServer) 為 192.168.221.5
MQTTPort 為 1883
MQTTUser 和
MQTTPassword 為 空字串("")

設定 publish 的 topic
設定 subscribe 的 topic
設定 publish 的間隔時間(15秒以內)，並以 MQTTClient.loop() 更新訂閱狀態

建立 WiFiClient 物件
WiFiClient WifiClient;

基於 WiFiClient 物件，建立 MQTTClient 物件
PubSubClient MQTTClient(WifiClient);

連接 broker (MQTTServer):
MQTTClient.setServer(MQTTServer, MQTTPort);
MQTTClient.setCallback(MQTTCallback);

[當 subscribe 的 topic 有更新時，執行MQTTCallback 副程式]

訂閱 topic 為 MQTTClient.subscribe(主題)
發佈 topic 資料為 MQTTClient.publish(主題, 字元)

```

#### 程式碼

``` c
#include <WiFi.h>
#include <WiFiMulti.h>    //多重連線
WiFiMulti wifiMulti;      //宣告多重連線

#include <PubSubClient.h> //請先安裝PubSubClient程式庫
#include <SimpleDHT.h>

// ------ 設定WiFi帳號密碼 ------
char ssid[] = "ssid";      //請改名
char password[] = "pw";  //請改名

//------ 設定DHT11腳位 ------
int pinDHT11 = 23;//
SimpleDHT11 dht11(pinDHT11);

char* MQTTServer = "broker.emqx.io";//免註冊MQTT伺服器
int MQTTPort = 1883;      //MQTT Port
char* MQTTUser = "";
char* MQTTPassword = "";

//推播主題1:推播溫度(記得改Topic)
char* MQTTPubTopic1 = "SAIHS_EE/學號/temp";
//推播主題2:推播濕度(記得改Topic)
char* MQTTPubTopic2 = "SAIHS_EE/學號/humi";
//訂閱主題1:改變LED燈號(記得改Topic)
char* MQTTSubTopic1 = "SAIHS_EE/學號/led";
long MQTTLastPublishTime;//此變數用來記錄推播時間
long MQTTPublishInterval = 10000;//每10秒推撥一次

WiFiClient WifiClient; // 建立 WiFiClient 物件
PubSubClient MQTTClient(WifiClient); // 基於 WiFiClient 物件，建立 MQTTClient 物件

void setup() {
  Serial.begin(115200);
  pinMode(15, OUTPUT);  //綠色LED燈

  //開始WiFiMulti連線
  WifiMultiConnecte();

  //開始MQTT連線
  MQTTConnecte();
}

void loop() {
  //如果WiFi連線中斷，則重啟WiFi連線
  if (WiFi.status() != WL_CONNECTED) { WifiMultiConnecte(); }

  //如果MQTT連線中斷，則重啟MQTT連線
  if (!MQTTClient.connected()) {  MQTTConnecte(); }

  //如果距離上次傳輸已經超過10秒，則Publish溫溼度
  if ((millis() - MQTTLastPublishTime) >= MQTTPublishInterval ) {
    //讀取溫濕度
    byte temperature = 0;
    byte humidity = 0;
    ReadDHT(&temperature, &humidity);
    // ------ 將DHT11溫濕度發佈到MQTT主題 ------
    MQTTClient.publish(MQTTPubTopic1, String(temperature).c_str());
    MQTTClient.publish(MQTTPubTopic2, String(humidity).c_str());
    Serial.println("溫溼度已發佈到MQTT Broker");
    MQTTLastPublishTime = millis();   //更新最後傳輸時間
  }
  MQTTClient.loop();//更新訂閱狀態
  delay(50);
}

//自建函式，讀取DHT11溫濕度
void ReadDHT(byte * temperature, byte * humidity) {
  int err = SimpleDHTErrSuccess;
  if ((err = dht11.read(temperature, humidity, NULL)) !=
      SimpleDHTErrSuccess) {
    Serial.print("讀取失敗,錯誤訊息="); 
    Serial.print(SimpleDHTErrCode(err));
    Serial.print(","); 
    Serial.println(SimpleDHTErrDuration(err)); 
    delay(1000);
    return;
  }
  Serial.print("DHT讀取成功：");
  Serial.print((int)*temperature); 
  Serial.print(" *C, ");
  Serial.print((int)*humidity); 
  Serial.println(" H");
}

//自建函式，開始WiFiMulti連線
void WifiMultiConnecte() {
  //開始WiFiMulti連線
  wifiMulti.addAP(ssid, password);
  wifiMulti.addAP(ssid1, password1);
  while(wifiMulti.run() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("WiFiMulti連線成功");
  Serial.print("IP Address: ");
  Serial.println(WiFi.localIP());
}

//自建函式，開始MQTT連線
void MQTTConnecte() {
  MQTTClient.setServer(MQTTServer, MQTTPort);
  MQTTClient.setCallback(MQTTCallback);
  while (!MQTTClient.connected()) {
    //以亂數為ClietID
    String  MQTTClientid = "esp32-" + String(random(1000000, 9999999));
    if (MQTTClient.connect(MQTTClientid.c_str(), MQTTUser, MQTTPassword)) {
      //連結成功，顯示「已連線」。
      Serial.println("MQTT已連線");
      //訂閱SubTopic1主題
      MQTTClient.subscribe(MQTTSubTopic1);
    } else {
      //若連線不成功，則顯示錯誤訊息，並重新連線
      Serial.print("MQTT連線失敗,狀態碼=");
      Serial.println(MQTTClient.state());
      Serial.println("十五秒後重新連線");
      delay(15000);
    }
  }
}

//自建函式，接收到訂閱時
void MQTTCallback(char* topic, byte* payload, unsigned int length) {
  Serial.print(topic); Serial.print("訂閱通知:");
  String payloadString; //將接收的payload轉成字串
  //顯示訂閱內容
  for (int i = 0; i < length; i++) {
    payloadString = payloadString + (char)payload[i];
  }
  Serial.println(payloadString);
  //比對主題是否為訂閱主題1
  if (strcmp(topic, MQTTSubTopic1) == 0) {
    Serial.println("改變燈號：" + payloadString);
    if (payloadString == "ON") {
      digitalWrite(15, HIGH);
    }
    if (payloadString == "OFF") {
      digitalWrite(15, LOW);
    }
  }
}
```

#### 參考資料:
https://hackmd.io/@richardychen/IoT_wk15mqtt

#### 自我練習

請利 mqtt 電腦端控制 esp32 上的 LED 。
