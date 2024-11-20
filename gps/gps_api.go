package gps

import (
	"encoding/json"
	"net/http"
)

// GPSAPI 提供对 GPS 模块的 HTTP 接口
type GPSAPI struct {
	module *GPSModule
}

// InitGPSAPI 初始化 GPS API 模块
func InitGPSAPI() *GPSAPI {
	module := NewGPSModule() // 创建 GPSModule 实例
	return &GPSAPI{module: module}
}

// NewGPSAPI 创建一个 GPSAPI 实例
func NewGPSAPI(module *GPSModule) *GPSAPI {
	return &GPSAPI{module: module}
}

// RegisterRoutes 注册 HTTP 路由，包括 WebSocket 路由
func (api *GPSAPI) RegisterRoutes(mux *http.ServeMux) {
	// 注册 HTTP API 路由
	mux.HandleFunc("/create_driver", api.HandleCreateDriver)
	mux.HandleFunc("/delete_driver", api.HandleDeleteDriver)
	mux.HandleFunc("/create_passenger", api.HandleCreatePassenger)
	mux.HandleFunc("/delete_passenger", api.HandleDeletePassenger)

	// 注册 WebSocket 路由
	mux.HandleFunc("/ws", api.HandleWebSocket)
}

// HandleWebSocket 对外暴露 WebSocket 功能
func (api *GPSAPI) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	api.module.HandleWebSocket(w, r) // 调用 GPSModule 中的 WebSocket 处理逻辑
}

// HandleCreateDriver 处理创建驾驶员的请求
func (api *GPSAPI) HandleCreateDriver(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		ID string `json:"id"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.ID == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	driver, err := api.module.CreateDriver(requestData.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(driver)
}

// HandleDeleteDriver 处理删除驾驶员的请求
func (api *GPSAPI) HandleDeleteDriver(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		ID string `json:"id"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.ID == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = api.module.DeleteDriver(requestData.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Driver deleted successfully"))
}

// HandleCreatePassenger 处理创建乘客的请求
func (api *GPSAPI) HandleCreatePassenger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		ID string `json:"id"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.ID == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	passenger, err := api.module.CreatePassenger(requestData.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(passenger)
}

// HandleDeletePassenger 处理删除乘客的请求
func (api *GPSAPI) HandleDeletePassenger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		ID string `json:"id"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil || requestData.ID == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = api.module.DeletePassenger(requestData.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Passenger deleted successfully"))
}
