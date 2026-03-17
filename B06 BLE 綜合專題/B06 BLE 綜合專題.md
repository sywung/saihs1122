## 臺北市立松山工農114學年度第二學期-智慧居家監控實習學習單

<center><font size=6>【BLE 綜合專題】</font></center>

<div style="text-align: right">班級：______________ 座號：________姓名：________________</div>


#### 一、系統架構

本課整合前五課所學，製作一個完整的 **BLE 智慧監控裝置**：

```
┌─────────────────────────────┐
│           ESP32             │
│                             │
│  DHT11 ──→ 溫度/濕度        │──BLE Notify──→ App 顯示數值
│  HC-SR04 ──→ 距離           │
│  PIR ──→ 人體偵測           │──BLE Notify──→ App 振動+音效警報
│                             │
│  OLED 顯示：                │
│    - BLE 連線狀態            │
│    - 溫度 / 濕度             │
│    - 距離                   │
│                             │
│  搖桿（App）──→ 伺服馬達    │←BLE Write── App 搖桿控制
└─────────────────────────────┘
```

**接線總表：**

| 模組 | 腳位 | ESP32 |
|------|------|-------|
| OLED SH1106 | SDA | GPIO 21 |
| OLED SH1106 | SCL | GPIO 22 |
| DHT11 | DATA | GPIO 23 |
| HC-SR04 | TRIG | GPIO 5 |
| HC-SR04 | ECHO | GPIO 18 |
| PIR HC-SR501 | OUT | GPIO 19 |
| SG90 伺服馬達 | Signal | GPIO 25 |


#### 二、OLED 顯示規劃

OLED 128×64 畫面分為四行：

```
┌────────────────────┐  ← 第 1 行（y=12）：BLE 狀態
│ BLE: Connected     │
│ T:26.5C  H:65%     │  ← 第 2 行（y=28）：溫度/濕度
│ Dist: 35 cm        │  ← 第 3 行（y=44）：距離
│ PIR: Clear         │  ← 第 4 行（y=60）：PIR 狀態
└────────────────────┘
```

使用 U8g2 函式庫（`U8G2_SH1106_128X64_NONAME_1_HW_I2C`）：
``` c
void updateOled(bool bleConnected) {
  char buf[32];
  u8g2.firstPage();
  do {
    u8g2.setFont(u8g2_font_6x12_tf);

    u8g2.setCursor(0, 12);
    u8g2.print(bleConnected ? "BLE: Connected" : "BLE: Waiting...");

    sprintf(buf, "T:%.1fC  H:%.0f%%", temperature, humidity);
    u8g2.setCursor(0, 28);
    u8g2.print(buf);

    sprintf(buf, "Dist: %.0f cm", distance);
    u8g2.setCursor(0, 44);
    u8g2.print(buf);

    u8g2.setCursor(0, 60);
    u8g2.print(pirDetected ? "PIR: Motion!" : "PIR: Clear");
  } while (u8g2.nextPage());
}
```


#### 三、完整程式碼

``` c {.line-numbers}
#include <BleManager.h>
#include <Wire.h>
#include <U8g2lib.h>
#include "DHT.h"
#include <ESP32Servo.h>

// --- 腳位定義 ---
#define DHT_PIN    23
#define TRIG_PIN    5
#define ECHO_PIN   18
#define PIR_PIN    19
#define SERVO_PIN  25

// --- 物件宣告 ---
U8G2_SH1106_128X64_NONAME_1_HW_I2C u8g2(U8G2_R0, U8X8_PIN_NONE);
DHT dht(DHT_PIN, DHT11);
Servo myServo;

// --- 全域變數 ---
float temperature = 0, humidity = 0, distance = 0;
bool pirDetected = false;
bool lastPirState = LOW;
unsigned long lastSensorTime = 0;
unsigned long lastAlertTime  = 0;
const unsigned long sensorInterval = 2000;
const unsigned long alertCooldown  = 5000;

// --- OLED 更新 ---
void updateOled(bool bleConnected) {
  char buf[32];
  u8g2.firstPage();
  do {
    u8g2.setFont(u8g2_font_6x12_tf);

    u8g2.setCursor(0, 12);
    u8g2.print(bleConnected ? "BLE: Connected" : "BLE: Waiting...");

    sprintf(buf, "T:%.1fC  H:%.0f%%", temperature, humidity);
    u8g2.setCursor(0, 28);
    u8g2.print(buf);

    sprintf(buf, "Dist: %.0f cm", distance);
    u8g2.setCursor(0, 44);
    u8g2.print(buf);

    u8g2.setCursor(0, 60);
    u8g2.print(pirDetected ? "PIR: Motion!" : "PIR: Clear");
  } while (u8g2.nextPage());
}

// --- BLE 回呼：連線狀態 ---
void onBleConnection(bool connected) {
  Serial.println(connected ? "BLE 已連線！" : "BLE 已斷線...");
  if (connected) {
    BleManager::getInstance().sendSoundFeedback(SOUND_SUCCESS);
  }
  updateOled(connected);
}

// --- BLE 回呼：收到封包 ---
void onBlePacketReceived(const BcbpPacketV1* packet) {
  if (packet->command == CMD_JOYSTICK) {
    int8_t x = (int8_t)packet->targetId;
    if (abs(x) < 5) x = 0;
    int angle = map(x, -100, 100, 0, 180);
    myServo.write(constrain(angle, 0, 180));
    Serial.printf("搖桿 X:%d → 角度:%d°\n", x, angle);
  }
}

// --- HC-SR04 距離量測 ---
float measureDistance() {
  digitalWrite(TRIG_PIN, LOW);  delayMicroseconds(2);
  digitalWrite(TRIG_PIN, HIGH); delayMicroseconds(10);
  digitalWrite(TRIG_PIN, LOW);
  long dur = pulseIn(ECHO_PIN, HIGH, 30000);
  if (dur == 0) return 999;
  return dur * 0.034 / 2;
}

void setup() {
  Serial.begin(115200);

  u8g2.begin();
  dht.begin();
  myServo.attach(SERVO_PIN, 500, 2400);
  myServo.write(90);
  pinMode(TRIG_PIN, OUTPUT);
  pinMode(ECHO_PIN, INPUT);
  pinMode(PIR_PIN, INPUT);

  // 啟動畫面
  u8g2.firstPage();
  do {
    u8g2.setFont(u8g2_font_8x13_tf);
    u8g2.setCursor(10, 30);
    u8g2.print("BLE Starting...");
  } while (u8g2.nextPage());

  // 初始化 BLE（裝置名稱請改為自己的座號）
  BleManager& ble = BleManager::getInstance();
  ble.setConnectionCallback(onBleConnection);
  ble.setPacketCallback(onBlePacketReceived);
  ble.setBatteryLevel(100);
  ble.begin("ESP32-Monitor"); // 請改為自己的座號，例如 ESP32-18

  Serial.println("系統啟動完成，等待手機連線...");
}

void loop() {
  BleManager::getInstance().update();

  bool bleConnected = BleManager::getInstance().isConnected();

  // 每 2 秒更新感測器數值
  if (millis() - lastSensorTime >= sensorInterval) {
    lastSensorTime = millis();

    float t = dht.readTemperature();
    float h = dht.readHumidity();
    if (!isnan(t) && !isnan(h)) {
      temperature = t;
      humidity    = h;
    }

    distance = measureDistance();
    updateOled(bleConnected);

    if (bleConnected) {
      BleManager::getInstance().sendAnalogReport(0, (uint16_t)(temperature * 10));
      BleManager::getInstance().sendAnalogReport(1, (uint16_t)(humidity * 10));
      BleManager::getInstance().sendAnalogReport(2, (uint16_t)distance);
    }

    Serial.printf("T:%.1f H:%.1f D:%.1f\n", temperature, humidity, distance);
  }

  // PIR 偵測（上升緣觸發）
  bool pirState = digitalRead(PIR_PIN);
  if (pirState == HIGH && lastPirState == LOW) {
    pirDetected = true;
    Serial.println("PIR：偵測到人！");
    if (bleConnected && (millis() - lastAlertTime >= alertCooldown)) {
      lastAlertTime = millis();
      BleManager::getInstance().sendCombinedFeedback(HAPTIC_WARNING, SOUND_ALERT);
    }
    updateOled(bleConnected);
  }
  if (pirState == LOW && lastPirState == HIGH) {
    pirDetected = false;
    updateOled(bleConnected);
  }
  lastPirState = pirState;

  delay(20);
}
```


#### 四、功能驗收清單

完成程式後，請逐項勾選確認：

- [ ] OLED 顯示「BLE: Waiting...」等待連線
- [ ] BLETestor App 找到裝置並連線
- [ ] OLED 切換為「BLE: Connected」，手機發出成功音效
- [ ] OLED 每 2 秒更新溫度、濕度、距離數值
- [ ] App 上 Analog 0/1/2 顯示對應數值
- [ ] 在 PIR 前移動，手機振動並播放警報音效，OLED 顯示「PIR: Motion!」
- [ ] 撥動 App 搖桿，伺服馬達跟著轉動


#### 五、進階挑戰（選做）

1. 在 OLED 加入「警報次數」計數器，每次 PIR 觸發後遞增顯示

2. 距離 < 20cm 時在 OLED 顯示警告符號 `!!!`，並傳送 `HAPTIC_ERROR` 給 App

3. 加入 App 按鈕功能：短按（`ACT_SHORT`）切換 PIR 警報開/關，OLED 顯示目前狀態（App 只送 `ACT_SHORT` 與 `ACT_RELEASE`，在按下瞬間切換即可）
