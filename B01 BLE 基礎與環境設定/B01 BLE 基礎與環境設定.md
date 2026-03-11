## 臺北市立松山工農114學年度第二學期-智慧居家監控實習學習單

<center><font size=6>【BLE 基礎與環境設定】</font></center>

<div style="text-align: right">班級：______________ 座號：________姓名：________________</div>


#### 一、BLE 是什麼？

**BLE（Bluetooth Low Energy，低功耗藍牙）** 是一種短距離無線通訊技術，專為低功耗應用設計。與 WiFi 不同，BLE 不需要路由器，裝置之間可以直接溝通。

| 比較項目 | BLE | 傳統藍牙 | WiFi |
|---------|-----|---------|------|
| 耗電量 | 極低 | 中 | 高 |
| 傳輸距離 | ~10m | ~10m | ~50m |
| 傳輸速率 | 低（適合小量資料）| 高（音樂串流）| 非常高 |
| 需要路由器 | 否 | 否 | 是 |
| 常見應用 | 感測器、穿戴裝置 | 耳機、喇叭 | 網際網路 |

ESP32 內建 BLE 硬體，不需要額外模組就可以直接使用。


#### 二、BLE 核心概念

BLE 通訊中，有兩種角色：

| 角色 | 說明 | 本課範例 |
|------|------|---------|
| **Peripheral（周邊裝置）** | 廣播自己的存在，等待連線 | ESP32 |
| **Central（中央裝置）** | 掃描並主動連線周邊裝置 | 手機（BLETestor App）|

**廣播（Advertising）**：ESP32 會持續發送廣播封包，讓附近的手機能掃描到它，就像在喊「我在這裡！我叫 ESP32-Basic！」

**GATT / Service / Characteristic**：連線後，資料透過「服務（Service）」和「特徵值（Characteristic）」來交換，類似資料夾（Service）裡放著不同的檔案（Characteristic）。本課使用的 `bcbp` 函式庫已將這些細節封裝好，不需要手動處理。

**RSSI（信號強度）**：單位 dBm，數值越接近 0 代表訊號越強。例如 -40 比 -80 強。


#### 三、安裝 BLETestor App

本課使用 **BLETestor** App 作為手機端控制介面。

| 平台 | QR Code | 下載連結 |
|------|:-------:|---------|
| iOS (iPhone/iPad) | ![iOS QR](assets/qr_ios.png) | https://apps.apple.com/tw/app/bletestor/id6758536675 |
| Android | ![Android QR](assets/qr_android.png) | https://play.google.com/store/apps/details?id=cc.twater.saihs.esp32.bletestor |

安裝後開啟 App，畫面會顯示附近的 BLE 裝置列表，可以看到裝置名稱與 RSSI 信號強度。


#### 四、安裝函式庫

本課需要安裝兩個函式庫：

**步驟 1：安裝 NimBLE-Arduino**
Arduino IDE → 工具 → 管理程式庫 → 搜尋 `NimBLE-Arduino`（作者：h2zero）→ 安裝

**步驟 2：安裝 bcbp（BCBP 協定函式庫）**
1. 前往 https://github.com/sywung/bcbp
2. 點選 `Code` → `Download ZIP`
3. Arduino IDE → 草稿碼 → 匯入程式庫 → 加入 .ZIP 程式庫 → 選擇剛下載的 ZIP

安裝完成後，於程式碼開頭加入：
```c
#include <BleManager.h>
```


#### 五、範例：ESP32 開始廣播並等待連線

以下程式讓 ESP32 廣播名稱 `ESP32-Basic`，手機 App 掃描到後可以連線，Serial Monitor 會顯示連線狀態。

``` c {.line-numbers}
#include <BleManager.h>

// 當連線/斷線時執行此函式
void onBleConnection(bool connected) {
  if (connected) {
    Serial.println("BLE 已連線！");
    digitalWrite(LED_BUILTIN, HIGH); // 連線時亮燈
  } else {
    Serial.println("BLE 已斷線，重新廣播中...");
    digitalWrite(LED_BUILTIN, LOW);  // 斷線時關燈
  }
}

// 當收到 App 傳來的封包時執行此函式
void onBlePacketReceived(const BcbpPacketV1* packet) {
  Serial.printf("收到封包 - 指令: 0x%02X, 目標: %d, 動作: %d\n",
                packet->command, packet->targetId, packet->action);
}

void setup() {
  Serial.begin(115200);
  pinMode(LED_BUILTIN, OUTPUT);

  // 設定回呼函式
  BleManager::getInstance().setConnectionCallback(onBleConnection);
  BleManager::getInstance().setPacketCallback(onBlePacketReceived);

  // 啟動 BLE，設定裝置廣播名稱
  BleManager::getInstance().begin("ESP32-Basic");

  Serial.println("BLE 初始化完成，等待手機連線...");
}

void loop() {
  // 必須在 loop 中呼叫 update，以維持 BLE 運作
  BleManager::getInstance().update();
  delay(10);
}
```

**操作步驟：**
1. 上傳程式到 ESP32
2. 開啟 Serial Monitor（115200 baud）
3. 開啟手機的 BLETestor App
4. 點選 `ESP32-Basic` 裝置連線
5. Serial Monitor 應顯示「BLE 已連線！」，板載 LED 亮起


#### 六、BleManager 常用函式

| 函式 | 說明 |
|------|------|
| `BleManager::getInstance()` | 取得 BleManager 物件（Singleton 模式）|
| `.begin("裝置名稱")` | 啟動 BLE 並開始廣播 |
| `.update()` | 維持 BLE 運作，**必須在 loop() 中呼叫** |
| `.isConnected()` | 回傳是否有手機連線中（true/false）|
| `.setConnectionCallback(函式)` | 設定連線/斷線時的回呼函式 |
| `.setPacketCallback(函式)` | 設定收到封包時的回呼函式 |


#### 七、自我練習

1. 將廣播名稱改為自己的座號（例如 `ESP32-18`），用 BLETestor 掃描確認
2. 手機靠近與遠離 ESP32，觀察 App 顯示的 RSSI 數值變化
3. 連線後拔除 USB 再重新插上，觀察 ESP32 重啟後 BLE 是否自動重新廣播

``` c
// 提示：修改這一行的裝置名稱
BleManager::getInstance().begin("ESP32-Basic"); // 改成你的座號
```
