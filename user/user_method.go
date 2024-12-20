package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"login/config"
	"login/db"
	"net/http"
)

type OrderInfo struct {
	OrderID            int    `json:"order_id"`
	StudentID          int    `json:"student_id"`
	CarID              string `json:"car_id"`
	PickupStationId    int    `json:"pickup_station_id"`
	DropoffStationId   int    `json:"dropoff_station_id"`
	PickupStationName  string `json:"pickup_station_name"`
	DropoffStationName string `json:"dropoff_station_name"`
	PickupTime         string `json:"pickup_time"`
	DropoffTime        string `json:"dropoff_time"`
	Status             string `json:"status"`
	PaymentID          int    `json:"payment_id"`
}
type PaymentInfo struct {
	PaymentID     int     `json:"payment_id"`
	OrderID       int     `json:"order_id"`
	VehicleID     string  `json:"vehicle_id"`
	PaymentAmount float32 `json:"payment_amount"`
	PaymentMethod string  `json:"payment_method"`
	PaymentTime   string  `json:"payment_time"`
	PaymentStatus string  `json:"payment_status"`
}

func setCORSHeaders(w http.ResponseWriter, methods string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", methods)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func submitOrder(tempOrderInfo OrderInfo) error {
	_, err := db.ExecuteSQL(config.RolePassenger, "INSERT into Order_Information(student_id,car_id,pickup_station_id,dropoff_station_id,pickup_station_name,dropoff_station_name,pickup_time,status,payment_id) values (?,?,?,?,?,?,?,?,?)", tempOrderInfo.StudentID, tempOrderInfo.CarID, tempOrderInfo.PickupStationId, tempOrderInfo.DropoffStationId, tempOrderInfo.PickupStationName, tempOrderInfo.DropoffStationName, tempOrderInfo.PickupTime, tempOrderInfo.Status, tempOrderInfo.PaymentID)
	if err != nil {
		return fmt.Errorf("添加订单信息失败: %w", err)
	}
	return nil
}
func submitPayment(tempPaymentInfo PaymentInfo) error {
	_, err := db.ExecuteSQL(config.RolePassenger, "INSERT into Payment_Record(order_id,vehicle_id,payment_amount,payment_method,payment_time,payment_status) values (?,?,?,?,?,?)", tempPaymentInfo.OrderID, tempPaymentInfo.VehicleID, tempPaymentInfo.PaymentAmount, tempPaymentInfo.PaymentMethod, tempPaymentInfo.PaymentTime, tempPaymentInfo.PaymentStatus)
	if err != nil {
		return fmt.Errorf("添加支付信息失败: %w", err)
	}
	return nil
}
func updateOrderStatus(order_id int, new_status string) error {
	_, err := db.ExecuteSQL(config.RolePassenger, "UPDATE Order_Information SET status = ? WHERE order_id = ?", new_status, order_id)
	if err != nil {
		return fmt.Errorf("更新订单信息失败: %w", err)
	}
	return nil
}
func updateLeaveTime(order_id int, leave_time string) error {
	_, err := db.ExecuteSQL(config.RolePassenger, "UPDATE Order_Information SET dropoff_time = ? WHERE order_id = ?", leave_time, order_id)
	if err != nil {
		return fmt.Errorf("更新订单信息失败: %w", err)
	}
	return nil
}
func updatePaymentStatus(payment_id int, new_status string) error {
	_, err := db.ExecuteSQL(config.RolePassenger, "UPDATE Payment_Record SET payment_status = ? WHERE payment_id = ?", new_status, payment_id)
	if err != nil {
		return fmt.Errorf("更新支付信息失败: %w", err)
	}
	return nil
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

func HandleChangeOrder(w http.ResponseWriter, r *http.Request) {
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
	var shift OrderInfo
	err = json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		log.Printf("JSON 解码失败: %v", err)
		respondWithError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}
	log.Printf("接收到的解码后数据: %+v", shift)

	if shift.OrderID == 0 || shift.Status == "" {
		respondWithError(w, http.StatusBadRequest, "缺少必要字段")
		return
	}

	// 更新车辆状态
	if err := updateOrderStatus(shift.OrderID, shift.Status); err != nil {
		respondWithError(w, http.StatusInternalServerError, "订单状态更新失败")
		return
	}
	respondWithSuccess(w, "订单状态修改成功")
}
func HandleChangeLeaveTime(w http.ResponseWriter, r *http.Request) {
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
	var shift OrderInfo
	err = json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		log.Printf("JSON 解码失败: %v", err)
		respondWithError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}
	log.Printf("接收到的解码后数据: %+v", shift)

	if shift.OrderID == 0 || shift.DropoffTime == "" {
		respondWithError(w, http.StatusBadRequest, "缺少必要字段")
		return
	}

	// 更新车辆状态
	if err := updateLeaveTime(shift.OrderID, shift.DropoffTime); err != nil {
		respondWithError(w, http.StatusInternalServerError, "下车时间更新失败")
		return
	}
	respondWithSuccess(w, "下车时间修改成功")
}
func HandleChangePayment(w http.ResponseWriter, r *http.Request) {
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
	var shift PaymentInfo
	err = json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		log.Printf("JSON 解码失败: %v", err)
		respondWithError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}
	log.Printf("接收到的解码后数据: %+v", shift)

	if shift.PaymentID == 0 || shift.PaymentStatus == "" {
		respondWithError(w, http.StatusBadRequest, "缺少必要字段")
		return
	}

	// 更新车辆状态
	if err := updatePaymentStatus(shift.PaymentID, shift.PaymentStatus); err != nil {
		respondWithError(w, http.StatusInternalServerError, "支付状态更新失败")
		return
	}
	respondWithSuccess(w, "支付状态修改成功")
}

func HandleSubmitOrder(w http.ResponseWriter, r *http.Request) {
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
	var shift OrderInfo
	err = json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		log.Printf("JSON 解码失败: %v", err)
		respondWithError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}
	log.Printf("接收到的解码后数据: %+v", shift)

	// 更新车辆状态
	if err := submitOrder(shift); err != nil {
		respondWithError(w, http.StatusInternalServerError, "添加订单信息失败")
		return
	}

	respondWithSuccess(w, "添加订单信息成功")
}
func HandleSubmitPayment(w http.ResponseWriter, r *http.Request) {
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
	var shift PaymentInfo
	err = json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		log.Printf("JSON 解码失败: %v", err)
		respondWithError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}
	log.Printf("接收到的解码后数据: %+v", shift)

	// 更新车辆状态
	if err := submitPayment(shift); err != nil {
		respondWithError(w, http.StatusInternalServerError, "添加支付信息失败")
		return
	}

	respondWithSuccess(w, "添加支付信息成功")
}
