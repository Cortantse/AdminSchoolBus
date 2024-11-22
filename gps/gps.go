package gps

import (
	// "bytes"
	// "encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Driver 代表驾驶员的基本信息
// 用于存储每位驾驶员的唯一ID和当前位置（经纬度）
type Driver struct {
	ID        string  `json:"id"`        // 驾驶员唯一标识
	Latitude  float64 `json:"latitude"`  // 驾驶员当前纬度
	Longitude float64 `json:"longitude"` // 驾驶员当前经度
}

// Passenger 代表乘客的基本信息
// 用于存储每位乘客的唯一ID
type Passenger struct {
	ID string `json:"id"` // 乘客唯一标识
}

// GPSModule 是核心管理模块
// 负责管理驾驶员和乘客的信息、处理 WebSocket 通信及广播数据
type GPSModule struct {
	drivers         map[string]*Driver              // 存储驾驶员信息
	passengers      map[string]*Passenger           // 存储乘客信息
	driversMutex    sync.Mutex                      // 用于保护驾驶员数据
	passengersMutex sync.Mutex                      // 用于保护乘客数据
	clients         map[*websocket.Conn]*sync.Mutex // 存储客户端连接及其对应的写锁
	clientToDriver  map[*websocket.Conn]string      // 将客户端连接映射到驾驶员 ID
	broadcast       chan []*Driver                  // 广播通道
	upgrader        websocket.Upgrader              // WebSocket 升级器
}

// NewGPSModule 创建一个 GPSModule 实例
// 初始化所有内部字段，准备接受请求和管理数据
func NewGPSModule() *GPSModule {
	return &GPSModule{
		drivers:    make(map[string]*Driver),
		passengers: make(map[string]*Passenger),
		clients:    make(map[*websocket.Conn]*sync.Mutex), // 初始化连接与写锁映射
		broadcast:  make(chan []*Driver),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有跨域连接
			},
		},
	}
}

// CreateDriver 创建一个新的驾驶员对象
// 输入：驾驶员ID
// 返回：创建成功的 Driver 对象，或错误信息
func (g *GPSModule) CreateDriver(id string) (*Driver, error) {
	if id == "" {
		return nil, errors.New("driver ID cannot be empty")
	}

	g.driversMutex.Lock()
	defer g.driversMutex.Unlock()

	if _, exists := g.drivers[id]; exists {
		return nil, errors.New("driver already exists")
	}

	driver := &Driver{ID: id}
	g.drivers[id] = driver
	return driver, nil
}

// DeleteDriver 删除一个驾驶员对象
// 输入：驾驶员ID
// 返回：删除成功或失败的错误信息
func (g *GPSModule) DeleteDriver(id string) error {
	if id == "" {
		return errors.New("driver ID cannot be empty")
	}

	g.driversMutex.Lock()
	defer g.driversMutex.Unlock()

	if _, exists := g.drivers[id]; !exists {
		return errors.New("driver not found")
	}
	// 找到对应的 WebSocket 连接
	var client *websocket.Conn
	for conn, driverID := range g.clientToDriver {
		if driverID == id {
			client = conn
			break
		}
	}

	// 如果找到对应的连接，进行清理
	if client != nil {
		go g.cleanupClient(client)
	}
	delete(g.drivers, id)
	fmt.Printf("Driver with ID %s has been deleted\n", id)

	// 广播最新的驾驶员数据
	g.broadcast <- g.GetAllDrivers()
	return nil
}

// CreatePassenger 创建一个新的乘客对象
// 输入：乘客ID
// 返回：创建成功的 Passenger 对象，或错误信息
func (g *GPSModule) CreatePassenger(id string) (*Passenger, error) {
	if id == "" {
		return nil, errors.New("passenger ID cannot be empty")
	}

	g.passengersMutex.Lock()
	defer g.passengersMutex.Unlock()

	if _, exists := g.passengers[id]; exists {
		return nil, errors.New("passenger already exists")
	}

	passenger := &Passenger{ID: id}
	g.passengers[id] = passenger
	return passenger, nil
}

// DeletePassenger 删除一个乘客对象
// 输入：乘客ID
// 返回：删除成功或失败的错误信息
func (g *GPSModule) DeletePassenger(id string) error {
	if id == "" {
		return errors.New("passenger ID cannot be empty")
	}

	g.passengersMutex.Lock()
	defer g.passengersMutex.Unlock()

	if _, exists := g.passengers[id]; !exists {
		return errors.New("passenger not found")
	}

	delete(g.passengers, id)
	fmt.Printf("Passenger with ID %s has been deleted\n", id)
	return nil
}

// UpdateDriverLocation 更新驾驶员的位置信息
// 输入：驾驶员ID、纬度、经度
// 返回：更新成功或失败的错误信息
func (g *GPSModule) UpdateDriverLocation(id string, latitude, longitude float64) error {
	// if latitude < -90 || latitude > 90 {
	// 	return errors.New("invalid latitude: must be between -90 and 90")
	// }
	// if longitude < -180 || longitude > 180 {
	// 	return errors.New("invalid longitude: must be between -180 and 180")
	// }

	g.driversMutex.Lock()
	defer g.driversMutex.Unlock()

	driver, exists := g.drivers[id]
	if !exists {
		return errors.New("driver not found")
	}

	driver.Latitude = latitude
	driver.Longitude = longitude
	return nil
}

// GetAllDrivers 获取所有驾驶员的位置信息
// 返回：驾驶员信息的切片
func (g *GPSModule) GetAllDrivers() []*Driver {
	g.driversMutex.Lock()
	defer g.driversMutex.Unlock()

	drivers := make([]*Driver, 0, len(g.drivers))
	for _, driver := range g.drivers {
		drivers = append(drivers, driver)
	}
	return drivers
}

// handleHeartbeat 定期发送心跳消息检测客户端连接是否存活
func (g *GPSModule) handleHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		g.driversMutex.Lock()
		for client, lock := range g.clients {
			lock.Lock()
			err := client.WriteMessage(websocket.PingMessage, []byte("ping"))
			lock.Unlock()
			if err != nil {
				fmt.Printf("Heartbeat failed for client: %v, removing client\n", err)
				go g.cleanupClient(client)
			}
		}
		g.driversMutex.Unlock()
	}
}

// HandleWebSocket 处理 WebSocket 连接
// 为每个客户端启动监听和广播协程
func (g *GPSModule) HandleWebSocket(w http.ResponseWriter, r *http.Request, driverID string) {
	conn, err := g.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}

	// 初始化客户端的写锁并绑定驾驶员 ID
	g.clients[conn] = &sync.Mutex{}
	g.clientToDriver[conn] = driverID

	fmt.Printf("New WebSocket client connected for Driver ID: %s\n", driverID)

	// 启动监听消息和广播的协程
	go g.listenClientMessages(conn)
	go g.handleHeartbeat()
}

// listenClientMessages 监听 WebSocket 客户端发送的消息
// 并更新驾驶员的位置信息
func (g *GPSModule) listenClientMessages(conn *websocket.Conn) {
	defer func() {
		conn.Close()
		g.clients[conn] = &sync.Mutex{} // 为每个 WebSocket 连接分配一个互斥锁
	}()

	for {
		var requestData struct {
			ID        string  `json:"id"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		}

		err := conn.ReadJSON(&requestData)
		if err != nil {
			fmt.Printf("Client message error: %v\n", err)
			break
		}

		_ = g.UpdateDriverLocation(requestData.ID, requestData.Latitude, requestData.Longitude)
		g.broadcast <- g.GetAllDrivers()
	}
}

func (g *GPSModule) cleanupClient(client *websocket.Conn) {
	// 获取对应的驾驶员 ID
	driverID := g.clientToDriver[client]

	// 移除客户端
	delete(g.clients, client)
	delete(g.clientToDriver, client)

	// 删除驾驶员
	if driverID != "" {
		if err := g.DeleteDriver(driverID); err != nil {
			fmt.Printf("Failed to delete driver %s: %v\n", driverID, err)
		}
	}

	// 关闭连接
	client.Close()
}

// broadcastDriverUpdates 广播驾驶员位置信息给所有 WebSocket 客户端
// func (g *GPSModule) broadcastDriverUpdates() {
// 	for drivers := range g.broadcast {
// 		for client, lock := range g.clients {
// 			buf := jsonBufferPool.Get().(*bytes.Buffer)
// 			buf.Reset()

// 			err := json.NewEncoder(buf).Encode(drivers)
// 			if err != nil {
// 				fmt.Printf("Encoding error: %v\n", err)
// 				client.Close()
// 				delete(g.clients, client)
// 				continue
// 			}

// 			lock.Lock() // 加锁保护写操作
// 			err = client.WriteMessage(websocket.TextMessage, buf.Bytes())
// 			lock.Unlock() // 解锁

// 			if err != nil {
// 				fmt.Printf("Write error: %v\n", err)
// 				client.Close()
// 				delete(g.clients, client)
// 			}

// 			jsonBufferPool.Put(buf)
// 		}
// 	}
// }

// // 使用 sync.Pool 复用 JSON 编码缓冲区
// var jsonBufferPool = sync.Pool{
// 	New: func() interface{} {
// 		return &bytes.Buffer{}
// 	},
// }
