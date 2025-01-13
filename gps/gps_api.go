package gps

import (
	"encoding/json"
	"fmt"
	"login/websocket" // 引入 WebSocket API 模块
	"net/http"
)

// GPSAPI 提供对 GPS 模块的 HTTP 接口
type GPSAPI struct {
	module *GPSModule // 通过 GPSModule 间接操作 WebSocket API
}

// InitGPSAPI 初始化 GPS API 模块
func InitGPSAPI(webSocketAPI *websocket.WebSocketAPI) *GPSAPI {
	module := NewGPSModule(webSocketAPI) // 创建 GPSModule 实例
	return &GPSAPI{module: module}
}

// RegisterRoutes 注册 HTTP 路由
func (api *GPSAPI) RegisterRoutes(mux *http.ServeMux) {
	// 注册 HTTP API 路由
	mux.HandleFunc("/create_driver", api.HandleCreateDriver)
	mux.HandleFunc("/delete_driver", api.HandleDeleteDriver)
	mux.HandleFunc("/create_passenger", api.HandleCreatePassenger)
	mux.HandleFunc("/delete_passenger", api.HandleDeletePassenger)
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

	driver, err := api.module.CreateDriver(requestData.ID) // 通过 GPSModule 调用创建方法
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(driver)
}

// HandleDeleteDriver 处理删除驾驶员的请求
func (api *GPSAPI) HandleDeleteDriver(w http.ResponseWriter, r *http.Request) {
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

	err = api.module.DeleteDriver(requestData.ID) // 通过 GPSModule 调用删除方法
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

	passenger, err := api.module.CreatePassenger(requestData.ID) // 通过 GPSModule 调用创建方法
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

	err = api.module.DeletePassenger(requestData.ID) // 通过 GPSModule 调用删除方法
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Passenger deleted successfully"))
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

// updateDriverLocation 内部调用的更新駕駛員位置方法
func (api *GPSAPI) UpdateDriverLocation(id string, latitude, longitude float64, car_id string) error {
	if id == "" {
		return fmt.Errorf("")
	}

	// 调用 GPSModule 中的方法来删除乘客
	err := api.module.UpdateDriverLocation(id, latitude, longitude, car_id)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

// CreateDriver 供内部模块调用来创建驾驶员
func (api *GPSAPI) StartBroadcast() {
	api.module.StartBroadcast()
}
