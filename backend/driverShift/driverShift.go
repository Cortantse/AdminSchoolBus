package driverShift

import (
	"login/gps" // 引入 gps 模块
	"encoding/json"
	"fmt"
	"net/http"
)

// 工作班次信息结构体
type WorkShift struct {
	DriverID   string `json:"driver_id"`   // 工号
	VehicleNo  string `json:"vehicle_no"`  // 车牌号
	Route      string `json:"route"`       // 路线
	ShiftStart string `json:"shift_start"` // 上班时间
	ShiftEnd   string `json:"shift_end"`   // 下班时间
	NumPeople  int    `json:"num_people"`  // 下班人数
}

// 提供信息的返回结构体
type SelectInfo struct {
	driverID   int
	VehicleNos []string `json:"vehicle_no_list"`
	Routes     []string `json:"route_list"`
}

// 模拟数据库存储的车辆和路线信息
var driverID = 123456
var vehicleList = []string{"A12345", "B67890", "C11223"}
var routeList = []string{"Route1", "Route2", "Route3"}

// 提供信息：为上班窗口提供合法的信息
func ProvideInfo(w http.ResponseWriter, r *http.Request) {
	// 设置CORS响应头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 仅支持GET请求
	if r.Method != http.MethodGet {
		http.Error(w, "仅支持 GET 请求", http.StatusMethodNotAllowed)
		return
	}

	// 返回合法的车牌号和路线列表
	selectInfo := SelectInfo{
		driverID:   driverID,
		VehicleNos: vehicleList,
		Routes:     routeList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(selectInfo)
}

// 处理上班：验证信息并修改车辆状态、创建GPS驾驶员对象
func HandleShiftStart(w http.ResponseWriter, r *http.Request) {
	// 设置CORS响应头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 如果是预检请求，直接返回状态200
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 仅支持POST请求
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持 POST 请求", http.StatusMethodNotAllowed)
		return
	}

	var shift WorkShift
	err := json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		http.Error(w, "请求数据解析失败", http.StatusBadRequest)
		return
	}

	// 验证必需字段
	if shift.DriverID == "" || shift.VehicleNo == "" || shift.Route == "" {
		http.Error(w, "缺少必要字段", http.StatusBadRequest)
		return
	}

	// 模拟数据库操作：修改车辆状态为使用中
	if err := updateVehicleStatus(shift.VehicleNo, "In Use"); err != nil {
		http.Error(w, "车辆状态更新失败", http.StatusInternalServerError)
		return
	}

	// 在GPS模块中创建一个驾驶员对象
	_, err = gps.NewGPSModule().CreateDriver(shift.DriverID, 34.0522, -118.2437)
	if err != nil {
		http.Error(w, "创建驾驶员失败", http.StatusInternalServerError)
		return
	}

	// 返回成功消息
	response := map[string]string{"message": "上班信息处理成功"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 处理下班：验证信息并修改车辆状态、删除GPS驾驶员对象
func HandleShiftEnd(w http.ResponseWriter, r *http.Request) {
	// 设置CORS响应头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 如果是预检请求，直接返回状态200
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 仅支持POST请求
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持 POST 请求", http.StatusMethodNotAllowed)
		return
	}

	var shift WorkShift
	err := json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		http.Error(w, "请求数据解析失败", http.StatusBadRequest)
		return
	}

	// 验证必需字段
	if shift.NumPeople == 0 || shift.VehicleNo == "" {
		http.Error(w, "缺少必要字段", http.StatusBadRequest)
		return
	}

	// 模拟数据库操作：修改车辆状态为未使用
	if err := updateVehicleStatus(shift.VehicleNo, "Not In Use"); err != nil {
		http.Error(w, "车辆状态更新失败", http.StatusInternalServerError)
		return
	}

	// 在GPS模块中删除驾驶员对象
	if err := gps.NewGPSModule().DeleteDriver(shift.DriverID); err != nil {
		http.Error(w, "删除驾驶员失败", http.StatusInternalServerError)
		return
	}

	// 返回成功消息
	response := map[string]string{"message": "下班信息处理成功"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 模拟更新车辆状态的函数
func updateVehicleStatus(vehicleNo, status string) error {
	// 在数据库中更新车辆状态
	// 此处仅为示例，实际应操作数据库
	fmt.Printf("车辆 %s 状态已更新为 %s\n", vehicleNo, status)
	return nil
}
