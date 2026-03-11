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

每次 App 或 ESP32 傳送資料，都使用固定 **6 bytes** 的 BCBP v1 封包格式：

| Byte 0 | Byte 1 | Byte 2 | Byte 3 | Byte 4 | Byte 5 |
|:------:|:------:|:------:|:------:|:------:|:------:|
| Version | Command | TargetID | Action | Sequence | CRC8 |
| 版本號 | 指令類型 | 按鈕編號 | 動作類型 | 序號 | 檢查碼 |

**指令方向（Command）：**

| 數值範圍 | 方向 | 說明 |
|---------|------|------|
| `0x01 ~ 0x1F` | App → ESP32 | 控制指令（按鈕、搖桿）|
| `0x21 ~ 0x3F` | ESP32 → App | 回饋指令（振動、音效）|

本課使用的按鈕指令 `CMD_BUTTON = 0x01`，由 App 傳送給 ESP32。

**按鈕動作（Action）：**

| 常數 | 數值 | 說明 |
|------|------|------|
| `ACT_SHORT` | `0x01` | 短按 |
| `ACT_LONG` | `0x02` | 長按 |
| `ACT_DOUBLE` | `0x03` | 雙擊 |


#### 三、範例：接收按鈕事件控制 LED

BLETestor App 上的按鈕按下後，ESP32 接收 `CMD_BUTTON` 封包，根據動作類型控制板載 LED。

``` c {.line-numbers}
#include <BleManager.h>

const int LED_PIN = 2; // 板載 LED（部分板子為 LED_BUILTIN）
bool ledState = false;

// 連線狀態回呼
void onBleConnection(bool connected) {
  if (connected) {
    Serial.println("BLE 已連線！");
  } else {
    Serial.println("BLE 已斷線，重新廣播中...");
    ledState = false;
    digitalWrite(LED_PIN, LOW); // 斷線時關燈
  }
}

// 收到封包回呼
void onBlePacketReceived(const BcbpPacketV1* packet) {
  // 只處理按鈕指令
  if (packet->command != CMD_BUTTON) return;

  Serial.printf("按鈕 %d 被按下，動作：", packet->targetId);

  switch (packet->action) {
    case ACT_SHORT:  // 短按 → 開燈
      Serial.println("短按");
      ledState = true;
      digitalWrite(LED_PIN, HIGH);
      break;

    case ACT_LONG:   // 長按 → 閃爍 3 次
      Serial.println("長按");
      for (int i = 0; i < 3; i++) {
        digitalWrite(LED_PIN, HIGH); delay(150);
        digitalWrite(LED_PIN, LOW);  delay(150);
      }
      break;

    case ACT_DOUBLE: // 雙擊 → 關燈
      Serial.println("雙擊");
      ledState = false;
      digitalWrite(LED_PIN, LOW);
      break;
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
  delay(10);
}
```

**操作步驟：**
1. 上傳程式，開啟 Serial Monitor（115200 baud）
2. BLETestor App 連線到 `ESP32-Button`
3. 在 App 上點選按鈕 1，測試短按 / 長按 / 雙擊
4. 觀察 LED 與 Serial Monitor 的反應


#### 四、BcbpPacketV1 封包欄位說明

程式中 `onBlePacketReceived` 收到的 `packet` 是一個指標，指向解析好的封包結構：

| 欄位 | 型別 | 說明 | 按鈕指令時的內容 |
|------|------|------|----------------|
| `packet->command` | uint8_t | 指令類型 | `CMD_BUTTON`（0x01）|
| `packet->targetId` | uint8_t | 目標編號 | App 上的按鈕編號（1、2、3…）|
| `packet->action` | uint8_t | 動作 | `ACT_SHORT` / `ACT_LONG` / `ACT_DOUBLE` |
| `packet->sequence` | uint8_t | 封包序號 | 自動遞增，用於除錯 |


#### 五、自我練習

1. 將裝置名稱改為自己的座號，確認 App 可以找到並連線
2. 新增按鈕 2 的處理：短按讓 LED 快速閃爍 5 次

``` c
// 提示：用 packet->targetId 判斷是哪個按鈕
if (packet->command == CMD_BUTTON && packet->targetId == 2) {
  // 在這裡加入按鈕 2 的處理
}
```

3. 思考：如果不在 `loop()` 裡呼叫 `BleManager::getInstance().update()`，會發生什麼事？
