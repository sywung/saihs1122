## 臺北市立松山工農114學年度第二學期-智慧居家監控實習學習單

<center><font size=6>【BLE 自訂 UUID】</font></center>

<div style="text-align: right">班級：______________ 座號：________姓名：________________</div>


#### 一、為什麼需要自訂 UUID？

在前幾課中，所有同學的 ESP32 都使用相同的預設 UUID。這在教室中會造成問題：

| 問題 | 說明 |
|------|------|
| 裝置辨識困難 | BLETestor App 掃描時會看到許多名稱相似的裝置 |
| 連線到別人的裝置 | 可能誤連到旁邊同學的 ESP32 |
| 資料混亂 | 不同裝置的感測器數值出現在同一個 App 上 |

透過設定**專屬 UUID**，BLETestor App 可以鎖定只顯示特定 UUID 的裝置，解決上述問題。

> **注意：** 自訂 UUID 功能為 BLETestor App 的**付費進階功能**，需在 App 內解鎖後才能使用。


#### 二、UUID 是什麼？

**UUID（Universally Unique Identifier，全球唯一識別碼）** 是一組 128 位元的數字，用來識別 BLE 裝置上的 Service 與 Characteristic。

格式：`xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`（8-4-4-4-12，共 32 個十六進位數字）

範例：
```
73616968-7334-2722-6616-000000000018
```

每個 BLE 裝置可以有自己專屬的 UUID，只要不與他人重複即可。

**bcbp 函式庫中需要設定三組 UUID：**

| 名稱 | 用途 |
|------|------|
| Service UUID | 識別整個服務（資料夾） |
| RX UUID | App → ESP32 的寫入通道 |
| TX UUID | ESP32 → App 的通知通道 |


#### 三、以座號產生專屬 UUID

以固定前綴搭配座號，確保每位同學的 UUID 不重複：

| 座號 | Service UUID |
|------|-------------|
| 01 | `73616968-7334-2722-6616-000000000001` |
| 18 | `73616968-7334-2722-6616-000000000018` |
| 35 | `73616968-7334-2722-6616-000000000035` |

規則：將最後 12 碼（`000000000001`）的末兩位改為自己的**座號**（補零至兩位數）。

三組 UUID 的前綴相同，僅第五段第五碼不同（6、7、8）：

```
Service: 73616968-7334-2722-6616-0000000000XX
RX:      73616968-7334-2722-6617-0000000000XX
TX:      73616968-7334-2722-6618-0000000000XX
                             ↑
                          6/7/8 區分三組
```


#### 四、程式設定方式

在 `begin()` 之前呼叫 `setCustomUUIDs()`，**必須先設定 UUID 再啟動 BLE**：

``` c {.line-numbers}
#include <BleManager.h>

void setup() {
  Serial.begin(115200);

  BleManager& ble = BleManager::getInstance();

  // 請將 XX 改為自己的座號（兩位數，不足補零）
  // 例如座號 7 → 07，座號 18 → 18
  ble.setCustomUUIDs(
    "73616968-7334-2722-6616-000000000018",  // Service（改座號）
    "73616968-7334-2722-6617-000000000018",  // RX     （改座號）
    "73616968-7334-2722-6618-000000000018"   // TX     （改座號）
  );

  // 裝置名稱也加上座號，方便辨識
  ble.begin("ESP32-18"); // 改座號

  Serial.println("BLE 啟動，UUID 已設定");
}

void loop() {
  BleManager::getInstance().update();
}
```

> `setCustomUUIDs()` 若在 `begin()` 之後才呼叫將不會生效，設定順序不能顛倒。


#### 五、在 BLETestor App 設定自訂 UUID

1. 開啟 BLETestor App
2. 進入「設定」→「自訂 UUID」（需付費解鎖）
3. 輸入自己的 Service UUID
4. 儲存後，App 掃描頁面只會顯示符合此 UUID 的裝置


#### 六、將自訂 UUID 套用至 B06 綜合專題

將 B06 的完整程式 `setup()` 中，在 `ble.begin()` 前加入 UUID 設定：

``` c
BleManager& ble = BleManager::getInstance();

// 加入這三行（改成自己的座號）
ble.setCustomUUIDs(
  "73616968-7334-2722-6616-000000000018",
  "73616968-7334-2722-6617-000000000018",
  "73616968-7334-2722-6618-000000000018"
);

ble.setConnectionCallback(onBleConnection);
ble.setPacketCallback(onBlePacketReceived);
ble.setBatteryLevel(100);
ble.begin("ESP32-18"); // 改座號
```


#### 七、自我練習

1. 將自訂 UUID 套用至 B06 的完整程式，確認 App 能以 UUID 篩選找到自己的裝置

2. 與旁邊同學互相確認：使用各自的 UUID 後，App 是否只顯示自己的裝置？

3. 思考：若兩位同學使用相同座號，UUID 會衝突嗎？會發生什麼事？
