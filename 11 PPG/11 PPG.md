## 臺北市立松山工農112學年度第二學期電子科-感測器實習學習單 

<center><font size=6>【心率感測器 PPG Sensor (Hardware)】</font></center>

<div style="text-align: right">班級：______________ 座號：________姓名：________________</div>

#### 一、Photoplethysmography, PPG

<center>
<img src="assets/clip_image001.png" alt="image" width="auto" height="330"> <img src="assets/clip_image002.png" alt="image" width="auto" height="320">
</center>

光體積變化描記圖法 (英語：Photoplethysmography, PPG) 是以光學的方式取得的器官體積描記圖（英語：plethysmogram），一般通過脈搏血氧儀（英語：pulse oximetry）來照射皮膚並測量光吸收的變化量來實現。一般的脈膊血氧儀是用來偵測血液灌注到真皮與皮下組織的狀況。(from: wiki)

#### 二、PPG感測器電路原理

```mermaid
graph LR
A[紅外線接收器] -->B[高通濾波器]
    B --> C[低通濾波器]
    C --> D[放大器]
    D --> E["微控制器 (MCU)"]
```

![image-20240603131739046](assets/image-20240603131739046.png)

![image-20240603131757249](assets/image-20240603131757249.png)


1. 紅外線發射時，光線穿透手指得知目前血管內血液含量，將反射訊號由紅外線接收器接收。

2. 將接收的電壓變化經過高通濾波器**濾除人體低頻訊號**，例如手指移動、呼吸訊號等等。

3. 接著經過低通濾波器，**濾除人體高頻訊號**，例如市電雜訊。

4. 因原訊號電壓較小，必須經過放大器才能將其訊號放大，較容易觀察及程式演算。

5. 左方半電壓適用於將PPG訊號維持在電源電壓的一半，如此才會有較大不失真的訊號。

#### 三、元件介紹 
<center>
<img src="assets/clip_image004.png" alt="image" width="auto" height="320"> <img src="assets/clip_image005.png" alt="image" width="auto" height="320">
</center>
#### 四、程式說明

<center>
<img src="assets/clip_image006.png" alt="image" width="auto" height="340"> <img src="assets/clip_image007.png" alt="image" width="auto" height="340">
<img src="assets/clip_image008.png" alt="image" width="auto" height="360"> <img src="assets/clip_image009.png" alt="image" width="auto" height="360">
</center>
 

#### 五、自我練習

1. 請完成電路銲接，手指輕微放置於CNY-70紅外線感測器上，使用示波器查看Vo波形，是否有產生如上波形之PPG圖。
