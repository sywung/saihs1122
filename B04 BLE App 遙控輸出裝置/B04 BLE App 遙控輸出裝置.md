## 臺北市立松山工農114學年度第二學期-智慧居家監控實習學習單

<center><font size=6>【BLE App 遙控輸出裝置】</font></center>

<div style="text-align: right">班級：______________ 座號：________姓名：________________</div>


#### 一、App 遙控資料流

本課學習用 BLETestor App 上的搖桿控制 ESP32 連接的輸出裝置（伺服馬達）。

```
BLETestor App 搖桿
     ↓ BLE Write（RX Characteristic）
   ESP32
     ↓ 解析 CMD_JOYSTICK 封包
 伺服馬達轉動角度
```

搖桿封包使用 `CMD_JOYSTICK (0x02)`，X / Y 軸數值各以 **int8_t（-100 ～ +100）** 儲存於封包中。


#### 二、搖桿封包欄位對應

| 欄位 | 型別 | 說明 |
|------|------|------|
| `packet->command` | uint8_t | `CMD_JOYSTICK`（0x02）|
| `packet->targetId` | int8_t | X 軸數值（-100 ～ +100，左負右正）|
| `packet->action` | int8_t | Y 軸數值（-100 ～ +100，下負上正）|

> 注意：`targetId` 與 `action` 欄位宣告為 `uint8_t`，但搖桿數值為有號整數，讀取時須強制轉型為 `(int8_t)`。

**死區（Dead Zone）** 處理：搖桿放開後不一定回到完美的 0，通常會有 ±5 左右的殘值。設定死區可避免裝置在手指離開後仍微幅動作：

``` c
int8_t x = (int8_t)packet->targetId;
if (abs(x) < 5) x = 0; // 死區：±5 以內視為 0
```


#### 三、map() 函式

`map()` 可以將某個範圍的數值，線性對應到另一個範圍，非常適合用來轉換搖桿數值。

| 語法 | 說明 |
|------|------|
| `map(value, fromLow, fromHigh, toLow, toHigh)` | 將 value 從來源範圍對應到目標範圍 |

```
搖桿 X 軸：-100 ～ +100
伺服馬達：0 ～ 180 度

map(x, -100, 100, 0, 180)
  x = -100 → 0°
  x =    0 → 90°
  x = +100 → 180°
```


#### 四、範例：搖桿 X 軸控制伺服馬達角度

**函式庫需求：** 安裝 `ESP32Servo`

**接線：** SG90 伺服馬達 Signal（橙色）接 GPIO 18，VCC（紅色）接 5V，GND（棕色）接 GND

``` c {.line-numbers}
#include <BleManager.h>
#include <ESP32Servo.h>

const int SERVO_PIN = 18;
Servo myServo;

int currentAngle = 90; // 初始角度：置中

void onBleConnection(bool connected) {
  if (connected) {
    Serial.println("BLE 已連線！");
  } else {
    Serial.println("BLE 已斷線，伺服馬達回中...");
    myServo.write(90); // 斷線時回到中間位置
    currentAngle = 90;
  }
}

void onBlePacketReceived(const BcbpPacketV1* packet) {
  if (packet->command != CMD_JOYSTICK) return;

  // 讀取 X 軸（強制轉型為有號整數）
  int8_t x = (int8_t)packet->targetId;
  int8_t y = (int8_t)packet->action;

  // 死區處理：±5 以內視為 0
  if (abs(x) < 5) x = 0;
  if (abs(y) < 5) y = 0;

  // 將 X 軸 (-100 ~ +100) 對應到伺服角度 (0 ~ 180)
  int angle = map(x, -100, 100, 0, 180);

  // 角度限制（安全保護）
  angle = constrain(angle, 0, 180);

  myServo.write(angle);
  currentAngle = angle;

  Serial.printf("搖桿 X:%d Y:%d → 角度:%d°\n", x, y, angle);
}

void setup() {
  Serial.begin(115200);
  myServo.attach(SERVO_PIN, 500, 2400);
  myServo.write(90); // 初始置中

  BleManager::getInstance().setConnectionCallback(onBleConnection);
  BleManager::getInstance().setPacketCallback(onBlePacketReceived);
  BleManager::getInstance().begin("ESP32-Servo");

  Serial.println("等待手機連線...");
}

void loop() {
  BleManager::getInstance().update();
  delay(10);
}
```

**操作步驟：**
1. 上傳程式，伺服馬達應先停在 90°（中間）
2. BLETestor App 連線到 `ESP32-Servo`
3. 撥動 App 上的搖桿，觀察伺服馬達跟著轉動
4. 放開搖桿，確認死區設定是否讓馬達穩定停止


#### 五、搖桿數值對照表

| 搖桿位置 | X 數值 | 轉換後角度 |
|---------|--------|----------|
| 完全向左 | -100 | 0° |
| 中間（放開）| 0 | 90° |
| 完全向右 | +100 | 180° |


#### 六、自我練習

1. 修改程式，改用 **Y 軸**控制伺服馬達，並觀察搖桿上推/下推時馬達的轉動方向

2. 加入 **按鈕事件**（`CMD_BUTTON`），短按讓伺服馬達回到 90°（中間位置）

``` c
// 提示：App 只送 ACT_SHORT（按下）/ ACT_RELEASE（放開）
// 在按下瞬間觸發回中即可
if (packet->command == CMD_BUTTON && packet->action == ACT_SHORT) {
  myServo.write(90);
  Serial.println("伺服馬達回中！");
}
```

3. 思考：為什麼斷線時要讓伺服馬達回到中間位置？在什麼應用場景下這很重要？
