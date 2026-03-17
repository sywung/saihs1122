## 臺北市立松山工農114學年度第二學期-智慧居家監控實習學習單

<center><font size=6>【BLE 雙向通訊與回饋】</font></center>

<div style="text-align: right">班級：______________ 座號：________姓名：________________</div>


#### 一、雙向通訊架構

前幾課分別學了「App → ESP32」（按鈕、搖桿）和「ESP32 → App」（感測器數值）。本課將兩個方向合在一起，並加入 **回饋指令**：ESP32 偵測到事件後，主動通知 App 振動或播放音效。

```
       App → ESP32：按鈕、搖桿控制
       ESP32 → App：感測器數值（Analog/Digital Report）
       ESP32 → App：振動回饋（CMD_HAPTIC）
       ESP32 → App：音效回饋（CMD_SOUND）
       ESP32 → App：同時振動＋音效（CMD_FEEDBACK）
```

這種設計讓 ESP32 在偵測到異常（如有人闖入、溫度過高）時，能直接通知使用者的手機，不需要使用者盯著畫面。


#### 二、回饋指令說明

**振動回饋（CMD_HAPTIC）**

| 常數 | 說明 |
|------|------|
| `HAPTIC_SHORT` | 短震（~50ms）|
| `HAPTIC_LONG` | 長震（~300ms）|
| `HAPTIC_DOUBLE` | 雙震 |
| `HAPTIC_SUCCESS` | 成功（短-長）|
| `HAPTIC_ERROR` | 錯誤（長-長-長）|
| `HAPTIC_WARNING` | 警告（短-短）|

**音效回饋（CMD_SOUND）**

| 常數 | 說明 |
|------|------|
| `SOUND_BEEP` | 單響 Beep |
| `SOUND_SUCCESS` | 成功音效 |
| `SOUND_ERROR` | 錯誤音效 |
| `SOUND_ALERT` | 警報音效 |
| `SOUND_DOUBLE` | 雙響 |

**發送函式：**

| 函式 | 說明 |
|------|------|
| `sendHapticFeedback(pattern, intensity)` | 發送振動回饋給 App；`intensity` 為強度（0~255），省略時使用最強（255）|
| `sendSoundFeedback(soundId, volume)` | 發送音效回饋給 App；`volume` 為音量（0~255），省略時使用最大聲（255）|
| `sendCombinedFeedback(pattern, soundId)` | 同時發送振動＋音效 |


#### 三、範例：PIR 感測到人→通知 App 警報

**接線：**

| 模組 | 腳位 | ESP32 |
|------|------|-------|
| PIR HC-SR501 | OUT | GPIO 23 |
| PIR HC-SR501 | VCC | 5V |
| PIR HC-SR501 | GND | GND |

``` c {.line-numbers}
#include <BleManager.h>

const int PIR_PIN = 23;

bool lastPirState = LOW;
unsigned long lastAlertTime = 0;
const unsigned long alertCooldown = 5000; // 警報冷卻 5 秒，避免連續觸發

void onBleConnection(bool connected) {
  if (connected) {
    Serial.println("BLE 已連線！");
    // 連線成功時發送成功音效
    BleManager::getInstance().sendSoundFeedback(SOUND_SUCCESS);
  } else {
    Serial.println("BLE 已斷線，重新廣播中...");
  }
}

// 按鈕手勢判斷（App 只送 ACT_SHORT 按下 / ACT_RELEASE 放開）
const unsigned long LONG_PRESS_MS = 800;
bool btnPressed   = false;
unsigned long pressStart = 0;

void onBlePacketReceived(const BcbpPacketV1* packet) {
  if (packet->command != CMD_BUTTON) return;

  if (packet->action == ACT_SHORT) {
    // 按下：記錄時間
    pressStart = millis();
    btnPressed = true;

  } else if (packet->action == ACT_RELEASE && btnPressed) {
    // 放開：判斷是短按或長按
    btnPressed = false;
    unsigned long dur = millis() - pressStart;

    if (dur >= LONG_PRESS_MS) {
      Serial.println("長按：測試音效回饋");
      BleManager::getInstance().sendSoundFeedback(SOUND_ALERT);
    } else {
      Serial.println("短按：測試振動回饋");
      BleManager::getInstance().sendHapticFeedback(HAPTIC_DOUBLE);
    }
  }
}

void setup() {
  Serial.begin(115200);
  pinMode(PIR_PIN, INPUT);

  BleManager::getInstance().setConnectionCallback(onBleConnection);
  BleManager::getInstance().setPacketCallback(onBlePacketReceived);
  BleManager::getInstance().begin("ESP32-Alert");

  Serial.println("等待手機連線...");
}

void loop() {
  BleManager::getInstance().update();

  bool pirState = digitalRead(PIR_PIN);

  // 偵測到人（LOW → HIGH 上升緣）且冷卻時間已過
  if (pirState == HIGH && lastPirState == LOW) {
    Serial.println("偵測到人！");

    if (BleManager::getInstance().isConnected()) {
      unsigned long now = millis();
      if (now - lastAlertTime >= alertCooldown) {
        lastAlertTime = now;
        // 同時觸發振動警告 + 警報音效
        BleManager::getInstance().sendCombinedFeedback(HAPTIC_WARNING, SOUND_ALERT);
        Serial.println("已發送警報給 App！");
      }
    }
  }

  lastPirState = pirState;
  delay(50);
}
```

**操作步驟：**
1. 上傳程式，BLETestor App 連線到 `ESP32-Alert`
2. 連線時手機會發出成功音效確認連線
3. 在 PIR 感測範圍內移動，手機應同時振動並播放警報音效
4. 在 App 上短按按鈕測試振動，長按測試音效


#### 四、冷卻時間（Cooldown）

PIR 感測器偵測到人後，輸出訊號會維持數秒才歸零，若不加冷卻時間會在短時間內連續發送大量封包。以 `millis()` 記錄上次觸發時間，確保兩次警報之間至少間隔 5 秒。

``` c
if (now - lastAlertTime >= alertCooldown) {
    lastAlertTime = now;
    // 執行警報
}
```


#### 五、自我練習

設計距離警報系統：使用 HC-SR04 超音波感測器，當偵測到物體距離小於 15cm 時，發送對應的回饋給 App。

**接線：** HC-SR04 TRIG 接 GPIO 5，ECHO 接 GPIO 18

| 距離 | 回饋類型 |
|------|---------|
| < 15cm | `sendCombinedFeedback(HAPTIC_ERROR, SOUND_ERROR)` |
| 15cm ～ 30cm | `sendHapticFeedback(HAPTIC_WARNING)` |
| > 30cm | 無回饋 |

``` c
// 提示：距離量測
long duration = pulseIn(ECHO_PIN, HIGH);
float distance = duration * 0.034 / 2;

if (distance < 15) {
  // 加入冷卻時間後發送 Combined Feedback
}
```

思考：`sendHapticFeedback()` 與 `sendCombinedFeedback()` 各適合用在哪些場景？
