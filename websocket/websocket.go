package websocket

import (
	"encoding/json"
	"login/log_service"

	"github.com/gorilla/websocket"
)

type DriverLocationUpdater interface {
	UpdateDriverLocation(driverID string, latitude, longitude float64) error
}

// WebSocketManager 管理WebSocket连接，支持不同类型的客户端
type WebSocketManager struct {
	Clients    map[*websocket.Conn]string // 存储连接池及其对应的客户端类型（"driver", "passenger", "admin"）
	Broadcast  chan []byte                // 用于广播消息
	Register   chan *websocket.Conn       // 注册连接
	Unregister chan *websocket.Conn       // 注销连接
	Updater    DriverLocationUpdater      // 引入接口
}

// NewWebSocketManager 创建WebSocketManager实例
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		Clients:    make(map[*websocket.Conn]string),
		Broadcast:  make(chan []byte),
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}
}

// 客户端类型常量
const (
	ClientTypeDriver    = "driver"
	ClientTypePassenger = "passenger"
	ClientTypeAdmin     = "admin"
)

// HandleWebSocketConnection 处理每个WebSocket连接
func (wm *WebSocketManager) HandleWebSocketConnection(conn *websocket.Conn, clientType string) {
	// 注册连接并指定客户端类型
	wm.Register <- conn
	log_service.WebSocketLogger.Printf("新连接建立，客户端类型：%s\n", clientType)
	defer func() {
		wm.Unregister <- conn
		conn.Close()
		log_service.WebSocketLogger.Printf("连接关闭，客户端类型：%s\n", clientType)
	}()

	wm.Clients[conn] = clientType

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log_service.WebSocketLogger.Printf("连接意外关闭，错误：%v\n", err)
			} else {
				log_service.WebSocketLogger.Printf("连接关闭，错误：%v\n", err)
			}
			break
		}

		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log_service.WebSocketLogger.Printf("消息解析失败，错误：%v\n", err)
			continue
		}

		switch msg.Type {
		case "driver_gps":
			// log_service.WebSocketLogger.Printf("收到驾驶员 GPS 信息：%v\n", msg)
			if wm.Updater != nil {
				err := wm.Updater.UpdateDriverLocation(msg.DriverID, msg.Location.Latitude, msg.Location.Longitude)
				if err != nil {
					log_service.WebSocketLogger.Printf("更新驾驶员位置失败：%v\n", err)
				}
			}
		default:
			log_service.WebSocketLogger.Printf("未知消息类型：%s\n", msg.Type)
		}
	}
}

// SendMessageToClients 向所有客户端或特定类型的客户端发送消息
func (wm *WebSocketManager) SendMessageToClients(message []byte, clientType string) {
	for conn, cType := range wm.Clients {
		// 如果指定了客户端类型，则仅发送给匹配的类型
		if clientType == "" || cType == clientType {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log_service.WebSocketLogger.Printf("向客户端发送消息失败，错误：%v\n", err)
				conn.Close()
				delete(wm.Clients, conn)
				log_service.WebSocketLogger.Printf("客户端连接移除，类型：%s\n", cType)
			}
		}
	}
}

// Start 启动WebSocket服务器，监听注册、注销和广播消息
func (wm *WebSocketManager) Start() {
	log_service.WebSocketLogger.Println("WebSocket 服务器已启动")
	for {
		select {
		case conn := <-wm.Register:
			// 默认注册为"passenger"类型
			wm.Clients[conn] = ClientTypePassenger
			log_service.WebSocketLogger.Println("新客户端已注册，类型：passenger")

		case conn := <-wm.Unregister:
			delete(wm.Clients, conn)
			log_service.WebSocketLogger.Println("客户端已注销")

		case message := <-wm.Broadcast:
			log_service.WebSocketLogger.Printf("广播消息：%s\n", string(message))
			wm.SendMessageToClients(message, "") // 发送给所有客户端
		}
	}
}
