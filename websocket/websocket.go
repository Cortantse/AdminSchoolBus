// package websocket

// import (
// 	"encoding/json"
// 	"fmt"
// 	"login/log_service"
// 	"runtime"

// 	"github.com/gorilla/websocket"
// )

// type DriverLocationUpdater interface {
// 	UpdateDriverLocation(driverID string, latitude, longitude float64) error
// }

// // WebSocketManager 管理WebSocket连接，支持不同类型的客户端
// type WebSocketManager struct {
// 	Clients    map[*websocket.Conn]string // 存储连接池及其对应的客户端类型（"driver", "passenger", "admin"）
// 	Broadcast  chan []byte                // 用于广播消息
// 	Register   chan *websocket.Conn       // 注册连接
// 	Unregister chan *websocket.Conn       // 注销连接
// 	Updater    DriverLocationUpdater      // 引入接口
// }

// // NewWebSocketManager 创建WebSocketManager实例
// func NewWebSocketManager() *WebSocketManager {
// 	log_service.WebSocketLogger.Println("WebSocket Manager initialized.")
// 	return &WebSocketManager{
// 		Clients:    make(map[*websocket.Conn]string),
// 		Broadcast:  make(chan []byte),
// 		Register:   make(chan *websocket.Conn),
// 		Unregister: make(chan *websocket.Conn),
// 	}
// }

// // 客户端类型常量
// const (
// 	ClientTypeDriver    = "driver"
// 	ClientTypePassenger = "passenger"
// 	ClientTypeAdmin     = "admin"
// )

// // getCallerInfo 获取调用者的信息
// func getCallerInfo() string {
// 	pc, file, line, ok := runtime.Caller(2) // 获取上两级调用栈信息
// 	if !ok {
// 		return "Unknown caller"
// 	}
// 	funcName := runtime.FuncForPC(pc).Name()
// 	return fmt.Sprintf("%s:%d %s", file, line, funcName)
// }

// // HandleWebSocketConnection 处理每个WebSocket连接
// func (wm *WebSocketManager) HandleWebSocketConnection(conn *websocket.Conn, clientType string) {
// 	callerInfo := getCallerInfo()
// 	log_service.WebSocketLogger.Printf("Called by: %s - Handling WebSocket connection: %v, type: %s", callerInfo, conn.RemoteAddr(), clientType)

// 	// 注册连接并指定客户端类型
// 	wm.Register <- conn
// 	defer func() {
// 		wm.Unregister <- conn
// 		conn.Close()
// 		log_service.WebSocketLogger.Printf("Called by: %s - Connection closed: %v", callerInfo, conn.RemoteAddr())
// 	}()

// 	wm.Clients[conn] = clientType

// 	for {
// 		_, message, err := conn.ReadMessage()
// 		if err != nil {
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
// 				log_service.WebSocketLogger.Printf("Called by: %s - Unexpected close error: %v", callerInfo, err)
// 			} else {
// 				log_service.WebSocketLogger.Printf("Called by: %s - Connection closed: %v", callerInfo, err)
// 			}
// 			break
// 		}

// 		var msg WebSocketMessage
// 		if err := json.Unmarshal(message, &msg); err != nil {
// 			log_service.WebSocketLogger.Printf("Called by: %s - Error unmarshalling message: %v", callerInfo, err)
// 			continue
// 		}

// 		switch msg.Type {
// 		case "driver_gps":
// 			log_service.WebSocketLogger.Printf("Called by: %s - Received driver GPS: %v", callerInfo, msg)
// 			if wm.Updater != nil {
// 				err := wm.Updater.UpdateDriverLocation(msg.DriverID, msg.Location.Latitude, msg.Location.Longitude)
// 				if err != nil {
// 					log_service.WebSocketLogger.Printf("Called by: %s - Failed to update driver location: %v", callerInfo, err)
// 				} else {
// 					log_service.WebSocketLogger.Printf("Called by: %s - Driver location updated successfully: ID=%s", callerInfo, msg.DriverID)
// 				}
// 			}
// 		default:
// 			log_service.WebSocketLogger.Printf("Called by: %s - Unknown message type: %v", callerInfo, msg.Type)
// 		}
// 	}
// }

// // SendMessageToClients 向所有客户端或特定类型的客户端发送消息
// func (wm *WebSocketManager) SendMessageToClients(message []byte, clientType string) {
// 	callerInfo := getCallerInfo()
// 	for conn, cType := range wm.Clients {
// 		// 如果指定了客户端类型，则仅发送给匹配的类型
// 		if clientType == "" || cType == clientType {
// 			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
// 				log_service.WebSocketLogger.Printf("Called by: %s - Error sending message: %v", callerInfo, err)
// 				conn.Close()
// 				delete(wm.Clients, conn)
// 			} else {
// 				log_service.WebSocketLogger.Printf("Called by: %s - Message sent to client: %v", callerInfo, conn.RemoteAddr())
// 			}
// 		}
// 	}
// }

// // Start 启动WebSocket服务器，监听注册、注销和广播消息
// func (wm *WebSocketManager) Start() {
// 	log_service.WebSocketLogger.Println("WebSocket server started.")
// 	for {
// 		select {
// 		case conn := <-wm.Register:
// 			callerInfo := getCallerInfo()
// 			wm.Clients[conn] = ClientTypePassenger
// 			log_service.WebSocketLogger.Printf("Called by: %s - Client registered: %v", callerInfo, conn.RemoteAddr())
// 		case conn := <-wm.Unregister:
// 			callerInfo := getCallerInfo()
// 			delete(wm.Clients, conn)
// 			log_service.WebSocketLogger.Printf("Called by: %s - Client unregistered: %v", callerInfo, conn.RemoteAddr())
// 		case message := <-wm.Broadcast:
// 			callerInfo := getCallerInfo()
// 			log_service.WebSocketLogger.Printf("Called by: %s - Broadcasting message to all clients.", callerInfo)
// 			wm.SendMessageToClients(message, "") // 发送给所有客户端
// 		}
// 	}
// }

package websocket

import (
	"encoding/json"
	"log"

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
	defer func() {
		wm.Unregister <- conn
		conn.Close()
	}()

	wm.Clients[conn] = clientType

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("Unexpected close error: %v", err)
			} else {
				log.Printf("Connection closed: %v", err)
			}
			break
		}

		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Error unmarshalling message:", err)
			continue
		}

		switch msg.Type {
		case "driver_gps":
			log.Printf("Received driver GPS: %v", msg)
			if wm.Updater != nil {
				err := wm.Updater.UpdateDriverLocation(msg.DriverID, msg.Location.Latitude, msg.Location.Longitude)
				if err != nil {
					log.Println("Failed to update driver location:", err)
				}
			}
		default:
			log.Printf("Unknown message type: %v", msg.Type)
		}

		// wm.Broadcast <- message
	}
}

// SendMessageToClients 向所有客户端或特定类型的客户端发送消息
func (wm *WebSocketManager) SendMessageToClients(message []byte, clientType string) {
	for conn, cType := range wm.Clients {
		// 如果指定了客户端类型，则仅发送给匹配的类型
		if clientType == "" || cType == clientType {
			// log.Printf("safdsadf")
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("Error sending message:", err)
				conn.Close()
				delete(wm.Clients, conn)
			}
		}
	}
}

// Start 启动WebSocket服务器，监听注册、注销和广播消息
func (wm *WebSocketManager) Start() {

	for {
		select {
		case conn := <-wm.Register:
			// 默认注册为"passenger"类型
			wm.Clients[conn] = ClientTypePassenger
		case conn := <-wm.Unregister:
			delete(wm.Clients, conn)
		case message := <-wm.Broadcast:
			wm.SendMessageToClients(message, "") // 发送给所有客户端
		}
	}
}
