package gps

import (
	"encoding/json"
	"fmt"
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
	api.module.HandleWebSocket(w, r)
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

	driver, err := api.CreateDriver(requestData.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(driver)
}

// CreateDriver 供内部模块调用来创建驾驶员
func (api *GPSAPI) CreateDriver(ID string) (*Driver, error) {
	if ID == "" {
		return nil, fmt.Errorf("driver ID cannot be empty")
	}
	// 调用 GPSModule 中的方法来创建驾驶员
	driver, err := api.module.CreateDriver(ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %v", err)
	}
	return driver, nil
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

	err = api.DeleteDriver(requestData.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Driver deleted successfully"))
}

// DeleteDriver 供内部模块调用来删除驾驶员
func (api *GPSAPI) DeleteDriver(ID string) error {
	if ID == "" {
		return fmt.Errorf("driver ID cannot be empty")
	}

	// 调用 GPSModule 中的方法来删除驾驶员
	err := api.module.DeleteDriver(ID)
	if err != nil {
		return fmt.Errorf("failed to delete driver: %v", err)
	}
	return nil
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

	passenger, err := api.CreatePassenger(requestData.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(passenger)
}

// CreatePassenger 内部调用的创建乘客方法
func (api *GPSAPI) CreatePassenger(ID string) (*Passenger, error) {
	if ID == "" {
		return nil, fmt.Errorf("passenger ID cannot be empty")
	}

	// 调用 GPSModule 中的方法来创建乘客
	passenger, err := api.module.CreatePassenger(ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create passenger: %v", err)
	}
	return passenger, nil
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

	err = api.DeletePassenger(requestData.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Passenger deleted successfully"))
}

// DeletePassenger 内部调用的删除乘客方法
func (api *GPSAPI) DeletePassenger(ID string) error {
	if ID == "" {
		return fmt.Errorf("passenger ID cannot be empty")
	}

	// 调用 GPSModule 中的方法来删除乘客
	err := api.module.DeletePassenger(ID)
	if err != nil {
		return fmt.Errorf("failed to delete passenger: %v", err)
	}
	return nil
}
