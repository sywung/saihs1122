## 臺北市立松山工農114學年度第二學期-智慧居家監控實習學習單

<center><font size=6>【BLE 連線與按鈕事件】</font></center>

<div style="text-align: right">班級：______________ 座號：________姓名：________________</div>


#### 一、GATT 架構

BLE 連線後，資料的交換依照 **GATT（Generic Attribute Profile）** 規範來進行，架構如下：

```
手機（Central）
└── Service（服務，用 UUID 識別）
    ├── Characteristic RX（接收，Write）← App 寫入，ESP32 讀取
    └── Characteristic TX（傳送，Notify）→ ESP32 寫入，App 接收通知
```

| 名詞 | 說明 |
|------|------|
| **Service** | 一組相關功能的集合，類似資料夾，用 UUID 識別 |
| **Characteristic** | 實際儲存資料的單位，類似檔案 |
| **UUID** | 全球唯一識別碼，用來辨識 Service 或 Characteristic |
| **Write（RX）** | App → ESP32，App 主動寫入資料 |
| **Notify（TX）** | ESP32 → App，ESP32 主動推送資料給 App |

`bcbp` 函式庫已內建一組預設 UUID，不需要手動設定即可與 BLETestor App 溝通。


#### 二、BCBP 封包格式

每次 App 傳送資料，都使用固定 **6 bytes** 的 BCBP v1 封包格式：

| Byte 0 | Byte 1 | Byte 2 | Byte 3 | Byte 4 | Byte 5 |
|:------:|:------:|:------:|:------:|:------:|:------:|
| Version | Command | TargetID | Action | Sequence | CRC8 |
| 版本號 | 指令類型 | 按鈕編號 | 動作類型 | 序號 | 檢查碼 |

本課使用的按鈕指令 `CMD_BUTTON = 0x01`，由 App 傳送給 ESP32。

**App 傳送的按鈕動作（Action）：**

| 常數 | 數值 | 說明 |
|------|------|------|
| `ACT_RELEASE` | `0x00` | 按鈕放開 |
| `ACT_SHORT` | `0x01` | 按鈕按下 |

> BLETestor App 只送出「按下（`ACT_SHORT`）」與「放開（`ACT_RELEASE`）」兩種原始事件。
> **長按、雙擊、短按的判斷由 ESP32 自行計時完成。**


#### 三、ESP32 自行判斷按鈕手勢

App 只告訴 ESP32「現在按下了」或「現在放開了」，ESP32 根據時間差決定手勢：

```
時序圖：

  按下           放開
   │←── 持續時間 ──→│
   │                │
   │  ≥ 800 ms      │  → 長按（ACT_LONG）
   │  < 800 ms      │  → 等待 300 ms，看有沒有第二下
                    │
                    │← 300 ms 內再次按下放開 → 雙擊（ACT_DOUBLE）
                    │← 300 ms 內無第二下     → 短按（ACT_SHORT）
```

| 手勢 | 判斷條件 |
|------|---------|
| 長按 | 按下到放開的時間 ≥ 800 ms |
| 雙擊 | 兩次按下，兩次放開之間間隔 < 300 ms |
| 短按 | 放開後 300 ms 內沒有第二次按下 |


#### 四、範例：接收按鈕事件控制 LED

``` c {.line-numbers}
#include <BleManager.h>

const int LED_PIN = 2; // 板載 LED

// 手勢判斷時間常數
const unsigned long LONG_PRESS_MS = 800; // 長按門檻 (ms)
const unsigned long DOUBLE_GAP_MS = 300; // 雙擊判斷視窗 (ms)

// 按鈕狀態機變數
bool isPressed  = false;
bool waitDouble = false;
unsigned long pressStart  = 0;
unsigned long releaseTime = 0;

// 手勢確認後的處理
void handleGesture(uint8_t gesture) {
  switch (gesture) {
    case ACT_SHORT:
      Serial.println("短按：開燈");
      digitalWrite(LED_PIN, HIGH);
      break;

    case ACT_LONG:
      Serial.println("長按：閃爍 3 次");
      for (int i = 0; i < 3; i++) {
        digitalWrite(LED_PIN, HIGH); delay(150);
        digitalWrite(LED_PIN, LOW);  delay(150);
      }
      break;

    case ACT_DOUBLE:
      Serial.println("雙擊：關燈");
      digitalWrite(LED_PIN, LOW);
      break;
  }
}

// 收到封包回呼：只處理按下與放開的原始事件
void onBlePacketReceived(const BcbpPacketV1* packet) {
  if (packet->command != CMD_BUTTON) return;

  if (packet->action == ACT_SHORT) {
    // 按鈕按下：記錄時間
    pressStart = millis();
    isPressed  = true;

  } else if (packet->action == ACT_RELEASE) {
    // 按鈕放開：計算按住時間並判斷手勢
    if (!isPressed) return;
    isPressed = false;

    unsigned long dur = millis() - pressStart;

    if (dur >= LONG_PRESS_MS) {
      // 長按
      handleGesture(ACT_LONG);
      waitDouble = false;
    } else {
      // 短按或雙擊第一下
      if (waitDouble) {
        // 第二次放開 → 確認雙擊
        handleGesture(ACT_DOUBLE);
        waitDouble = false;
      } else {
        // 等待看是否有第二次按下
        waitDouble    = true;
        releaseTime   = millis();
      }
    }
  }
}

// 連線狀態回呼
void onBleConnection(bool connected) {
  if (connected) {
    Serial.println("BLE 已連線！");
  } else {
    Serial.println("BLE 已斷線，重新廣播中...");
    isPressed  = false;
    waitDouble = false;
    digitalWrite(LED_PIN, LOW);
  }
}

void setup() {
  Serial.begin(115200);
  pinMode(LED_PIN, OUTPUT);

  BleManager::getInstance().setConnectionCallback(onBleConnection);
  BleManager::getInstance().setPacketCallback(onBlePacketReceived);
  BleManager::getInstance().begin("ESP32-Button");

  Serial.println("等待手機連線...");
}

void loop() {
  BleManager::getInstance().update();

  // 雙擊視窗逾時 → 確認為短按
  if (waitDouble && (millis() - releaseTime >= DOUBLE_GAP_MS)) {
    handleGesture(ACT_SHORT);
    waitDouble = false;
  }
}
```

**操作步驟：**
1. 上傳程式，開啟 Serial Monitor（115200 baud）
2. BLETestor App 連線到 `ESP32-Button`
3. 在 App 上按下按鈕 1 後立即放開（短按），觀察 LED 亮起
4. 按住按鈕 1 超過 0.8 秒再放開（長按），觀察 LED 閃爍 3 次
5. 快速連按兩下按鈕 1（雙擊），觀察 LED 熄滅


#### 五、BcbpPacketV1 封包欄位說明

程式中 `onBlePacketReceived` 收到的 `packet` 是一個指標，指向解析好的封包結構：

| 欄位 | 型別 | 說明 | 按鈕指令時的內容 |
|------|------|------|----------------|
| `packet->command` | uint8_t | 指令類型 | `CMD_BUTTON`（0x01）|
| `packet->targetId` | uint8_t | 目標編號 | App 上的按鈕編號（1、2、3…）|
| `packet->action` | uint8_t | 動作 | `ACT_SHORT`（按下）/ `ACT_RELEASE`（放開）|
| `packet->sequence` | uint8_t | 封包序號 | 自動遞增，用於除錯 |


#### 六、自我練習

1. 將裝置名稱改為自己的座號，確認 App 可以找到並連線

2. 新增按鈕 2 的處理：短按讓 LED 快速閃爍 5 次

``` c
// 提示：用 packet->targetId 判斷是哪個按鈕
// 按鈕 1 與按鈕 2 共用同一組狀態機變數，
// 可宣告兩組獨立的 pressStart / waitDouble 分別處理
if (packet->targetId == 2) {
  // 在這裡加入按鈕 2 的手勢狀態機
}
```

3. 思考：為什麼短按要等待 300 ms 才確認，而不是放開就立刻觸發？
