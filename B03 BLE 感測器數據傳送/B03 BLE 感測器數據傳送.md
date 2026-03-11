## 臺北市立松山工農114學年度第二學期-智慧居家監控實習學習單

<center><font size=6>【BLE 感測器數據傳送】</font></center>

<div style="text-align: right">班級：______________ 座號：________姓名：________________</div>


#### 一、資料流方向：Device → App

前兩課學習了 App 控制 ESP32（App → Device），本課反過來，讓 ESP32 將感測器數值主動回報給 App（Device → App）。

```
DHT11 感測器
     ↓ 讀取
   ESP32
     ↓ BLE Notify（TX Characteristic）
  BLETestor App
     ↓ 顯示數值
```

ESP32 使用 `sendAnalogReport()` 或 `sendDigitalReport()` 將數值透過 TX Characteristic 推送給 App，App 收到後自動更新畫面。


#### 二、傳送函式說明

| 函式 | 說明 | 數值範圍 |
|------|------|---------|
| `sendAnalogReport(channel, value)` | 傳送類比數值 | channel: 0~255，value: 0~4095 |
| `sendDigitalReport(channel, state)` | 傳送數位狀態 | channel: 0~255，state: true/false |
| `setBatteryLevel(level)` | 更新電池電量 | 0~100（%）|

**通道（Channel）** 的概念：同一個 ESP32 可以同時回報多個感測器數值，用通道編號來區分。

| 通道 | 本課用途 |
|------|---------|
| 0 | 溫度（Temperature）|
| 1 | 濕度（Humidity）|


#### 三、範例：DHT11 溫濕度透過 BLE 傳送 App

**函式庫需求：** 安裝 `Adafruit DHT Sensor` 及 `Adafruit Unified Sensor`

**接線：** DHT11 DATA 腳接 GPIO 23

``` c {.line-numbers}
#include <BleManager.h>
#include "DHT.h"

#define DHTPIN  23
#define DHTTYPE DHT11

DHT dht(DHTPIN, DHTTYPE);

unsigned long lastReportTime = 0;
const unsigned long reportInterval = 2000; // 每 2 秒回報一次

void onBleConnection(bool connected) {
  if (connected) {
    Serial.println("BLE 已連線！");
  } else {
    Serial.println("BLE 已斷線，重新廣播中...");
  }
}

void setup() {
  Serial.begin(115200);
  dht.begin();

  BleManager::getInstance().setConnectionCallback(onBleConnection);
  BleManager::getInstance().setBatteryLevel(100); // 設定初始電量
  BleManager::getInstance().begin("ESP32-Sensor");

  Serial.println("等待手機連線...");
}

void loop() {
  BleManager::getInstance().update();

  // 每 2 秒回報一次感測器數值
  if (millis() - lastReportTime >= reportInterval) {
    lastReportTime = millis();

    // 讀取溫濕度
    float t = dht.readTemperature(); // 攝氏溫度
    float h = dht.readHumidity();    // 濕度（%）

    // 若讀取失敗則跳過
    if (isnan(t) || isnan(h)) {
      Serial.println("DHT11 讀取失敗！");
      return;
    }

    Serial.printf("溫度: %.1f°C  濕度: %.1f%%\n", t, h);

    // 只有連線中才發送
    if (BleManager::getInstance().isConnected()) {
      // 將浮點數乘以 10 轉為整數（避免小數點）
      // 例如 25.3°C → 253，App 顯示後再除以 10 還原
      BleManager::getInstance().sendAnalogReport(0, (uint16_t)(t * 10)); // 通道 0：溫度
      BleManager::getInstance().sendAnalogReport(1, (uint16_t)(h * 10)); // 通道 1：濕度
    }
  }
}
```

**操作步驟：**
1. 上傳程式，開啟 Serial Monitor（115200 baud）
2. BLETestor App 連線到 `ESP32-Sensor`
3. 觀察 App 上「Analog 0」與「Analog 1」數值更新
4. 用手指捏住 DHT11，觀察溫度與濕度數值變化


#### 四、數值換算說明

由於 `sendAnalogReport` 只能傳整數（uint16_t），傳送前將浮點數乘以 10 保留一位小數：

| 實際值 | 傳送值（×10）| App 顯示 |
|--------|------------|---------|
| 25.3°C | 253 | 253（需÷10 還原）|
| 68.5% | 685 | 685（需÷10 還原）|

BLETestor App 直接顯示原始整數值。如需在 OLED 或其他顯示器上還原，再除以 10 即可：
``` c
float temp = receivedValue / 10.0;
```


#### 五、電池電量回報

BLE 標準內建電池服務（Battery Service），App 可以直接讀取電量百分比。

``` c
// 在 setup() 設定初始電量
BleManager::getInstance().setBatteryLevel(100);

// 在 loop() 定期更新（例如讀取實際電壓後換算）
BleManager::getInstance().setBatteryLevel(batteryPercent);
```


#### 六、自我練習

加入 HC-SR04 超音波感測器，將距離值透過 **通道 2** 回報給 App。

**接線：** HC-SR04 TRIG 接 GPIO 5，ECHO 接 GPIO 18

``` c
// 提示：HC-SR04 距離量測
long duration;
float distance;

digitalWrite(TRIG_PIN, LOW);  delayMicroseconds(2);
digitalWrite(TRIG_PIN, HIGH); delayMicroseconds(10);
digitalWrite(TRIG_PIN, LOW);

duration = pulseIn(ECHO_PIN, HIGH);
distance = duration * 0.034 / 2; // 單位：cm

// 將距離（cm）透過通道 2 傳送給 App
BleManager::getInstance().sendAnalogReport(2, (uint16_t)distance);
```

思考：若感測器讀值異常（isnan 或距離 > 400cm），應該怎麼處理？
