package gps

import (
	"bytes"
	"encoding/json"
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
	drivers         map[string]*Driver       // 存储驾驶员信息，键为驾驶员ID
	passengers      map[string]*Passenger    // 存储乘客信息，键为乘客ID
	driversMutex    sync.Mutex               // 用于保护对驾驶员数据的并发访问
	passengersMutex sync.Mutex               // 用于保护对乘客数据的并发访问
	clients         map[*websocket.Conn]bool // 存储 WebSocket 客户端连接
	broadcast       chan []*Driver           // 广播通道，用于发送所有驾驶员位置信息
	upgrader        websocket.Upgrader       // 用于升级 HTTP 连接为 WebSocket
}

// NewGPSModule 创建一个 GPSModule 实例
// 初始化所有内部字段，准备接受请求和管理数据
func NewGPSModule() *GPSModule {
	return &GPSModule{
		drivers:    make(map[string]*Driver),
		passengers: make(map[string]*Passenger),
		clients:    make(map[*websocket.Conn]bool),
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
	if latitude < -90 || latitude > 90 {
		return errors.New("invalid latitude: must be between -90 and 90")
	}
	if longitude < -180 || longitude > 180 {
		return errors.New("invalid longitude: must be between -180 and 180")
	}

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
		for client := range g.clients {
			err := client.WriteMessage(websocket.PingMessage, []byte("ping"))
			if err != nil {
				fmt.Printf("Heartbeat failed for client: %v, removing client\n", err)
				client.Close()
				delete(g.clients, client)
			}
		}
		g.driversMutex.Unlock()
	}
}

// HandleWebSocket 处理 WebSocket 连接
// 为每个客户端启动监听和广播协程
func (g *GPSModule) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := g.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}

	g.clients[conn] = true
	fmt.Println("New WebSocket client connected")

	// 启动监听消息和广播的协程
	go g.listenClientMessages(conn)
	go g.broadcastDriverUpdates()
	go g.handleHeartbeat() // 启动心跳检测
}

// listenClientMessages 监听 WebSocket 客户端发送的消息
// 并更新驾驶员的位置信息
func (g *GPSModule) listenClientMessages(conn *websocket.Conn) {
	defer func() {
		conn.Close()
		g.clients[conn] = false
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

// broadcastDriverUpdates 广播驾驶员位置信息给所有 WebSocket 客户端
func (g *GPSModule) broadcastDriverUpdates() {
	for drivers := range g.broadcast {
		for client := range g.clients {
			buf := jsonBufferPool.Get().(*bytes.Buffer)
			buf.Reset()

			err := json.NewEncoder(buf).Encode(drivers)
			if err != nil {
				client.Close()
				delete(g.clients, client)
			} else {
				_ = client.WriteMessage(websocket.TextMessage, buf.Bytes())
			}

			jsonBufferPool.Put(buf)
		}
	}
}

// 使用 sync.Pool 复用 JSON 编码缓冲区
var jsonBufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}
