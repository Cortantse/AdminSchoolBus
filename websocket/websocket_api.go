package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocketAPI 提供对 WebSocket 管理的外部接口
type WebSocketAPI struct {
	manager  *WebSocketManager
	upgrader websocket.Upgrader // WebSocket 升级器
}

// NewWebSocketAPI 创建 WebSocket API 实例
func NewWebSocketAPI() *WebSocketAPI {
	return &WebSocketAPI{
		manager: NewWebSocketManager(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// 允许所有跨域请求，实际应用中可根据需求限制
				return true
			},
		},
	}
}

// SetUpdater 设置 WebSocketManager 的 Updater
func (api *WebSocketAPI) SetUpdater(updater DriverLocationUpdater) {
	api.manager.Updater = updater
}

// HandleWebSocket 处理 WebSocket 请求
func (api *WebSocketAPI) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级 HTTP 连接为 WebSocket 连接
	conn, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		http.Error(w, "Failed to upgrade connection", http.StatusBadRequest)
		return
	}

	// 默认将新连接标记为乘客类型，可以根据业务需求调整
	api.HandleConnection(conn, ClientTypePassenger)
}

// HandleConnection 处理 WebSocket 连接并指定客户端类型
func (api *WebSocketAPI) HandleConnection(conn *websocket.Conn, clientType string) {
	api.manager.HandleWebSocketConnection(conn, clientType)
}

// SendMessage 向所有客户端或特定类型的客户端发送消息
func (api *WebSocketAPI) SendMessage(message []byte, clientType string) {
	api.manager.SendMessageToClients(message, clientType)
}

// Start 启动 WebSocket 服务器并开始监听注册、注销和广播消息
func (api *WebSocketAPI) Start() {
	go api.manager.Start()
}

// RegisterRoutes 注册 WebSocket 路由到 HTTP 路由器
func (api *WebSocketAPI) RegisterRoutes(mux *http.ServeMux) {
	// 将 WebSocket 的路径 "/ws" 注册为路由
	mux.HandleFunc("/ws", api.HandleWebSocket)
}

// RegisterClient 注册一个 WebSocket 连接
func (api *WebSocketAPI) RegisterClient(conn *websocket.Conn) {
	api.manager.Register <- conn
}

// UnregisterClient 注销一个 WebSocket 连接
func (api *WebSocketAPI) UnregisterClient(conn *websocket.Conn) {
	api.manager.Unregister <- conn
}
