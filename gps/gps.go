package gps

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"login/websocket" // 引入 WebSocket API 模块
	"sync"
	"time"
)

type Driver struct {
	Type     string   `json:"type"`     // 消息类型
	ID       string   `json:"id"`       // 驾驶员唯一标识
	Location Location `json:"location"` // 地理位置（例如GPS定位）
}

// 地理位置结构体
type Location struct {
	Latitude  float64 `json:"latitude"`  // 纬度
	Longitude float64 `json:"longitude"` // 经度
}

type GPSModule struct {
	drivers         map[string]*Driver // 存储驾驶员信息
	driversMutex    sync.Mutex         // 用于保护驾驶员数据
	passengers      map[string]*Passenger
	passengersMutex sync.Mutex
	webSocketAPI    *websocket.WebSocketAPI // WebSocket API 实例
}

// Passenger 代表乘客的基本信息
type Passenger struct {
	ID string `json:"id"` // 乘客唯一标识
}

// NewGPSModule 创建一个 GPSModule 实例
func NewGPSModule(webSocketAPI *websocket.WebSocketAPI) *GPSModule {
	return &GPSModule{
		drivers:      make(map[string]*Driver),
		passengers:   make(map[string]*Passenger),
		webSocketAPI: webSocketAPI,
	}
}

// CreateDriver 创建一个新的驾驶员对象
func (g *GPSModule) CreateDriver(id string) (*Driver, error) {
	if id == "" {
		return nil, errors.New("driver ID cannot be empty")
	}

	g.driversMutex.Lock()
	defer g.driversMutex.Unlock()

	if _, exists := g.drivers[id]; exists {
		return nil, errors.New("driver already exists")
	}

	driver := &Driver{Type: "driver_gps", ID: id}
	g.drivers[id] = driver
	return driver, nil
}

// DeleteDriver 删除一个驾驶员对象
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

	return nil
}

// 定时广播驾驶员位置信息
func (g *GPSModule) StartBroadcast() {
	go func() { // 使用 goroutine 实现异步广播
		ticker := time.NewTicker(2 * time.Second) // 每两秒触发
		defer ticker.Stop()

		for range ticker.C {
			g.broadcastDriverLocations()
		}
	}()
}

// 广播所有驾驶员的位置信息
func (g *GPSModule) broadcastDriverLocations() {
	// g.driversMutex.Lock()

	if len(g.drivers) == 0 {
		log.Printf("1asdas")
		// g.driversMutex.Unlock()
		return // 如果没有驾驶员，不广播
	}

	// 序列化驾驶员位置信息
	driverData, err := json.Marshal(g.GetAllDrivers())
	// g.driversMutex.Unlock()
	if err != nil {
		return // 如果序列化失败，直接跳过
	}
	// log.Printf(string(driverData))

	// 广播消息到所有客户端
	log.Println(g.GetAllDrivers())
	g.webSocketAPI.SendMessage(driverData, "")
}

// UpdateDriverLocation 更新驾驶员的位置信息
func (g *GPSModule) UpdateDriverLocation(id string, latitude, longitude float64) error {
	g.driversMutex.Lock()
	defer g.driversMutex.Unlock()

	driver, exists := g.drivers[id]
	if !exists {
		return errors.New("driver not found")
	}

	driver.Location.Latitude = latitude
	driver.Location.Longitude = longitude

	return nil
}

// GetAllDrivers 获取所有驾驶员的位置信息
func (g *GPSModule) GetAllDrivers() []*Driver {
	g.driversMutex.Lock()
	defer g.driversMutex.Unlock()

	drivers := make([]*Driver, 0, len(g.drivers))
	for _, driver := range g.drivers {
		drivers = append(drivers, driver)
	}

	return drivers
}

// CreatePassenger 创建一个新的乘客对象
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
	return nil
}
