# API 文档 - 接口注释说明

# GPS API 文档

以下是 `gps_api.go` 文件中提供的接口文档，涵盖了所有注册的 HTTP 路由及其功能。

---

## **1. 创建驾驶员**
- **接口地址**: `/create_driver`
- **请求方法**: `POST`
- **请求头**: 
  - `Content-Type: application/json`
- **请求体**:
  ```json
  {
    "id": "string"
  }
  ```
  - `id`: 驾驶员的唯一标识符。

- **响应**:
  - 成功:
    ```json
    {
      "id": "string",
      "latitude": 0,
      "longitude": 0
    }
    ```
    - `id`: 创建的驾驶员 ID。
    - `latitude`: 驾驶员的初始纬度（默认为 0）。
    - `longitude`: 驾驶员的初始经度（默认为 0）。
  - 错误:
    - 状态码 `400`: 请求体无效。
    - 状态码 `500`: 创建驾驶员失败（如 ID 已存在）。

---

## **2. 删除驾驶员**
- **接口地址**: `/delete_driver`
- **请求方法**: `DELETE`
- **请求头**: 
  - `Content-Type: application/json`
- **请求体**:
  ```json
  {
    "id": "string"
  }
  ```
  - `id`: 要删除的驾驶员的唯一标识符。

- **响应**:
  - 成功: 状态码 `200`，消息 `"Driver deleted successfully"`。
  - 错误:
    - 状态码 `400`: 请求体无效。
    - 状态码 `404`: 驾驶员未找到。

---

## **3. 创建乘客**
- **接口地址**: `/create_passenger`
- **请求方法**: `POST`
- **请求头**: 
  - `Content-Type: application/json`
- **请求体**:
  ```json
  {
    "id": "string"
  }
  ```
  - `id`: 乘客的唯一标识符。

- **响应**:
  - 成功:
    ```json
    {
      "id": "string"
    }
    ```
    - `id`: 创建的乘客 ID。
  - 错误:
    - 状态码 `400`: 请求体无效。
    - 状态码 `500`: 创建乘客失败（如 ID 已存在）。

---

## **4. 删除乘客**
- **接口地址**: `/delete_passenger`
- **请求方法**: `DELETE`
- **请求头**: 
  - `Content-Type: application/json`
- **请求体**:
  ```json
  {
    "id": "string"
  }
  ```
  - `id`: 要删除的乘客的唯一标识符。

- **响应**:
  - 成功: 状态码 `200`，消息 `"Passenger deleted successfully"`。
  - 错误:
    - 状态码 `400`: 请求体无效。
    - 状态码 `404`: 乘客未找到。

---

## **5. WebSocket 连接**
- **接口地址**: `/ws`
- **请求方法**: `GET`
- **功能描述**:
  - 实现驾驶员位置的实时更新和广播。
  - 前端通过 WebSocket 连接该接口，可以：
    - 发送驾驶员的位置信息更新。
    - 实时接收所有驾驶员的位置信息。

- **发送数据格式**:
  ```json
  {
    "id": "string",
    "latitude": float,
    "longitude": float
  }
  ```
  - `id`: 驾驶员的唯一标识符。
  - `latitude`: 驾驶员的纬度。
  - `longitude`: 驾驶员的经度。

- **接收数据格式**:
  - 广播所有驾驶员的位置信息：
    ```json
    [
      {
        "id": "string",
        "latitude": float,
        "longitude": float
      }
    ]
    ```


---



# GPS 模块 文档

## 概述

`gps.go` 文件实现了一个 GPS 模块，用于管理驾驶员和乘客的信息，以及通过 WebSocket 通信广播驾驶员位置信息。主要功能包括驾驶员和乘客的创建、删除、位置信息的更新和广播。

---

## 数据结构说明

### 1. `Driver`
代表驾驶员的基本信息。
- **字段：**
  - `ID` (string): 驾驶员的唯一标识。
  - `Latitude` (float64): 驾驶员的纬度位置。
  - `Longitude` (float64): 驾驶员的经度位置。

### 2. `Passenger`
代表乘客的基本信息。
- **字段：**
  - `ID` (string): 乘客的唯一标识。

### 3. `GPSModule`
GPS 模块的核心结构，负责管理驾驶员和乘客的信息，处理 WebSocket 通信及广播数据。
- **字段：**
  - `drivers` (map[string]*Driver): 存储驾驶员信息。
  - `passengers` (map[string]*Passenger): 存储乘客信息。
  - `driversMutex` (sync.Mutex): 用于保护驾驶员信息的并发访问。
  - `passengersMutex` (sync.Mutex): 用于保护乘客信息的并发访问。
  - `clients` (map[*websocket.Conn]bool): 存储 WebSocket 客户端连接。
  - `broadcast` (chan []*Driver): 用于广播驾驶员位置信息的通道。
  - `upgrader` (websocket.Upgrader): 用于升级 HTTP 请求为 WebSocket。

---

## 核心函数说明

### 数据操作函数

1. **`NewGPSModule()`**
   - **描述：** 初始化并返回一个新的 `GPSModule` 实例。

2. **`CreateDriver(id string)`**
   - **描述：** 创建一个新的驾驶员对象。
   - **输入：** 驾驶员 ID。
   - **返回：** 驾驶员对象或错误。

3. **`DeleteDriver(id string)`**
   - **描述：** 删除指定 ID 的驾驶员对象。
   - **输入：** 驾驶员 ID。
   - **返回：** 错误信息。

4. **`CreatePassenger(id string)`**
   - **描述：** 创建一个新的乘客对象。
   - **输入：** 乘客 ID。
   - **返回：** 乘客对象或错误。

5. **`DeletePassenger(id string)`**
   - **描述：** 删除指定 ID 的乘客对象。
   - **输入：** 乘客 ID。
   - **返回：** 错误信息。

6. **`UpdateDriverLocation(id string, latitude, longitude float64)`**
   - **描述：** 更新驾驶员的位置信息。
   - **输入：** 驾驶员 ID、纬度、经度。
   - **返回：** 错误信息。

7. **`GetAllDrivers()`**
   - **描述：** 获取所有驾驶员的位置信息。
   - **返回：** 驾驶员信息列表。

---

### WebSocket 通信相关函数

1. **`HandleWebSocket(w http.ResponseWriter, r *http.Request)`**
   - **描述：** 处理新的 WebSocket 客户端连接。
   - **功能：** 为每个客户端启动监听和广播协程。

2. **`listenClientMessages(conn *websocket.Conn)`**
   - **描述：** 监听客户端发送的消息，并更新驾驶员位置信息。

3. **`broadcastDriverUpdates()`**
   - **描述：** 广播驾驶员位置信息给所有 WebSocket 客户端。

4. **`handleHeartbeat()`**
   - **描述：** 定期发送心跳消息检测客户端连接是否存活。

---

### 辅助功能

1. **`jsonBufferPool`**
   - **描述：** 使用 `sync.Pool` 提供 JSON 编码缓冲区复用，提升性能。

---

