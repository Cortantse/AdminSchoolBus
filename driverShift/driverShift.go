package driverShift

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"login/config"
	"login/db"
	"login/gps" // 引入 gps 模块
	"net/http"
)

// 工作班次信息结构体
type WorkShift struct {
	DriverID      string        `json:"driver_id"`    // 駕駛員編號
	VehicleNo     string        `json:"car_id"`       // 車牌號
	VehicleStatus string        `json:"car_isusing"`  // 车辆状态
	RouteID       int           `json:"route_id"`     // 路線編號
	ShiftStart    string        `json:"work_stime"`   // 上班時間
	ShiftEnd      string        `json:"work_etime"`   // 下班時間
	Feedback      string        `json:"remark"`       // 意見反饋
	RouteRecord   []RouteRecord `json:"record_route"` // 路徑記錄，包含時間和GPS坐標
}

// 路径记录结构体
type RouteRecord struct {
	Time string `json:"time"`  // 时间戳
	GPSX int    `json:"gps_x"` // GPS X 坐標
	GPSY int    `json:"gps_y"` // GPS Y 坐標
}

// var module := gps.NewGPSModule()

func updateVehicleStatus(carID string, newStatus string) error {
	_, err := db.ExecuteSQL(config.RoleDriver, "UPDATE car_isusing SET car_isusing = ? WHERE car_id = ?", newStatus, carID)
	if err != nil {
		return fmt.Errorf("更新车辆状态失败: %w", err)
	}
	return nil
}

// 通用CORS設置
func setCORSHeaders(w http.ResponseWriter, methods string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", methods)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// 通用錯誤響應函數
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// 通用成功響應函數
func respondWithSuccess(w http.ResponseWriter, message string) {
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

// 处理上班：验证信息并创建 GPS 驾驶员对象
func HandleShiftStart(w http.ResponseWriter, r *http.Request, gpsModule *gps.GPSModule) {
	log.Printf("接收到信息")
	setCORSHeaders(w, "POST, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "仅支持 POST 请求")
		return
	}

	bodyBytes, err := io.ReadAll(r.Body) // 读取原始请求体
	if err != nil {
		log.Printf("无法读取请求体: %v", err)
		respondWithError(w, http.StatusBadRequest, "无法读取请求体")
		return
	}
	log.Printf("接收到的原始数据: %s", string(bodyBytes))

	// 重置 Body 并解码为结构体
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	var shift WorkShift
	err = json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		log.Printf("JSON 解码失败: %v", err)
		respondWithError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}
	log.Printf("接收到的解码后数据: %+v", shift)

	if shift.DriverID == "" || shift.VehicleNo == "" || shift.RouteID == 0 || shift.VehicleStatus == "" {
		respondWithError(w, http.StatusBadRequest, "缺少必要字段")
		return
	}

	// 更新车辆状态
	if err := updateVehicleStatus(shift.VehicleNo, shift.VehicleStatus); err != nil {
		respondWithError(w, http.StatusInternalServerError, "车辆状态更新失败")
		return
	}

	// 创建驾驶员对象

	_, err = gpsModule.CreateDriver(shift.DriverID) // 初始纬度和经度为 0
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "创建驾驶员失败")
		return
	}

	respondWithSuccess(w, "上班信息处理成功")
}

// 处理下班：验证信息并删除 GPS 驾驶员对象
func HandleShiftEnd(w http.ResponseWriter, r *http.Request, gpsModule *gps.GPSModule) {
	setCORSHeaders(w, "POST, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "仅支持 POST 请求")
		return
	}

	var shift WorkShift
	err := json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}

	if shift.DriverID == "" || shift.VehicleNo == "" {
		respondWithError(w, http.StatusBadRequest, "缺少必要字段")
		return
	}

	// 更新车辆状态
	// if err := updateVehicleStatus(shift.VehicleNo, "Not In Use"); err != nil {
	// 	respondWithError(w, http.StatusInternalServerError, "车辆状态更新失败")
	// 	return
	// }

	// 删除驾驶员对象
	err = gpsModule.DeleteDriver(shift.DriverID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "删除驾驶员失败")
		return
	}

	respondWithSuccess(w, "下班信息处理成功")
}

// 模拟更新车辆状态的函数
// func updateVehicleStatus(vehicleNo, status string) error {
// 	fmt.Printf("车辆 %s 状态已更新为 %s\n", vehicleNo, status)
// 	return nil
// }
