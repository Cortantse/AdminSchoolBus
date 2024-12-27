package user

import (
	"bytes"
	"database/sql"
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
	StudentAccount     int    `json:"student_account"`
	DriverID           int    `json:"driver_id"`
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

type WorkShift struct {
	DriverID    string `json:"driver_id"`    // 駕駛員編號
	VehicleNo   string `json:"car_id"`       // 車牌號
	ShiftStart  string `json:"work_stime"`   // 上班時間
	ShiftEnd    string `json:"work_etime"`   // 下班時間
	CurrentTime string `json:"current_time"` // 上班時間
}
type Comment struct {
	Studentname    string `json:"studentname"`
	Commentcontent string `json:"commentcontent"`
	Commenttime    string `json:"commenttime"`
	Avatar         string `json:"avatar"`
}

func setCORSHeaders(w http.ResponseWriter, methods string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", methods)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func submitOrder(tempOrderInfo OrderInfo) error {
	_, err := db.ExecuteSQL(config.RolePassenger, "INSERT into Order_Information(student_account,driver_id,car_id,pickup_station_id,dropoff_station_id,pickup_station_name,dropoff_station_name,pickup_time,status,payment_id) values (?,?,?,?,?,?,?,?,?,?)", tempOrderInfo.StudentAccount, tempOrderInfo.DriverID, tempOrderInfo.CarID, tempOrderInfo.PickupStationId, tempOrderInfo.DropoffStationId, tempOrderInfo.PickupStationName, tempOrderInfo.DropoffStationName, tempOrderInfo.PickupTime, tempOrderInfo.Status, tempOrderInfo.PaymentID)
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
func updateOrderStatus(order_id int, payment_id int, new_status string) error {
	_, err := db.ExecuteSQL(config.RolePassenger, "UPDATE Order_Information SET status = ?,payment_id=? WHERE order_id = ?", new_status, payment_id, order_id)
	if err != nil {
		return fmt.Errorf("更新订单信息失败: %w", err)
	}
	return nil
}

func submitComment(tempComentInfo Comment) error {
	_, err := db.ExecuteSQL(config.RolePassenger, "INSERT INTO passenger_comment (student_name, comment_content, comment_time, avatar) VALUES (?,?,?,?)", tempComentInfo.Studentname, tempComentInfo.Commentcontent, tempComentInfo.Commenttime, tempComentInfo.Avatar)
	if err != nil {
		return fmt.Errorf("添加评论失败: %w", err)
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
func respondWithSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
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
	if err := updateOrderStatus(shift.OrderID, shift.PaymentID, shift.Status); err != nil {
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
	log.Printf("接收到的原始pay数据: %s", string(bodyBytes))

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

func GetjourneyRecord(w http.ResponseWriter, r *http.Request) {
	//fmt.Print("getjourney被调用----------------------------")
	// 从数据库中获取行程记录
	type JourneyRecord struct {
		Originsite      string `json:"originsite"`
		Destinationsite string `json:"destinationsite"`
		UpTime          string `json:"uptime"`
		DownTime        string `json:"downtime"`
		Status          string `json:"status"`
	}
	rows, err := db.ExecuteSQL(config.RolePassenger, "SELECT pickup_station_name,dropoff_station_name,pickup_time,dropoff_time,status FROM order_information WHERE order_id>? ", 0)
	if err != nil {
		fmt.Print(err)
	}
	res, _ := rows.(*sql.Rows)
	var results []JourneyRecord
	for res.Next() {
		var journey JourneyRecord
		var downTime sql.NullString // 使用 sql.NullString 处理可能为空的时间字段
		err := res.Scan(&journey.Originsite, &journey.Destinationsite, &journey.UpTime, &downTime, &journey.Status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if downTime.Valid {
			//journey.DownTime = new(string)
			journey.DownTime = downTime.String
		} else {
			journey.DownTime = "---"
		}
		if journey.Status == "0" {
			journey.Status = "进行中"
		} else {
			journey.Status = "结束"
		}
		results = append(results, journey)
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// 将数据转化为JSON格式并写入响应体
	json.NewEncoder(w).Encode(results)
}

func GetComment(w http.ResponseWriter, r *http.Request) {
	type Comment struct {
		Commentid      string `json:"commentid"`
		Studentname    string `json:"studentname"`
		Commentcontent string `json:"commentcontent"`
		Commenttime    string `json:"commenttime"`
		Avatar         string `json:"avatar"`
	}
	rows, err := db.ExecuteSQL(config.RolePassenger, "SELECT * FROM Passenger_Comment WHERE comment_id > ?", 0)
	if err != nil {
		fmt.Print(err)
	}
	res, _ := rows.(*sql.Rows)
	var results []Comment
	for res.Next() {
		var result Comment
		err := res.Scan(&result.Commentid, &result.Studentname, &result.Commentcontent, &result.Commenttime, &result.Avatar)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, result)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

func WriteComment(w http.ResponseWriter, r *http.Request) {
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
	// 重置 Body 并解码为结构体
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	var shift Comment
	err = json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		log.Printf("JSON 解码失败: %v", err)
		respondWithError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}
	if err := submitComment(shift); err != nil {
		respondWithError(w, http.StatusInternalServerError, "添加订单信息失败")
		return
	}
	respondWithSuccess(w, "添加评论成功")
}

func GetNotice(w http.ResponseWriter, r *http.Request) {
	type Notice struct {
		Noticeid    string `json:"noticeid"`
		Title       string `json:"title"`
		Content     string `json:"content"`
		Publishdate string `json:"publishdate"`
	}
	rows, err := db.ExecuteSQL(config.RolePassenger, "SELECT * FROM passenger_notice where notice_id > ?", 0)
	if err != nil {
		fmt.Print(err)
	}
	res, _ := rows.(*sql.Rows)
	var results []Notice
	for res.Next() {
		var result Notice
		err := res.Scan(&result.Noticeid, &result.Title, &result.Content, &result.Publishdate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, result)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}
func HandleGetCurrentOrder(w http.ResponseWriter, r *http.Request) {
	log.Println("GetCurrentOrder 被触发")
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

	// 执行查询获取订单信息
	result, err := db.ExecuteSQL(config.RolePassenger,
		"SELECT order_id FROM order_information WHERE student_account = ? AND pickup_time = ?;", shift.StudentAccount, shift.PickupTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "查询订单信息失败")
		return
	}

	// 类型断言：确保 result 是 *sql.Rows 类型
	rows, ok := result.(*sql.Rows)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "数据库返回结果格式错误")
		return
	}
	defer rows.Close()

	// 假设查询只有一行数据，映射到 OrderInfo 结构体
	var orderInfo OrderInfo
	if rows.Next() {
		err := rows.Scan(&orderInfo.OrderID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "解析订单信息失败")
			return
		}
	} else {
		respondWithError(w, http.StatusNotFound, "未找到该订单信息")
		return
	}

	respondWithSuccess(w, orderInfo)
}
func HandleGetCurrentPayment(w http.ResponseWriter, r *http.Request) {
	log.Println("GetCurrentPayment 被触发")
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

	// 执行查询获取订单信息
	result, err := db.ExecuteSQL(config.RolePassenger,
		"SELECT payment_id FROM payment_record WHERE order_id = ? AND payment_time = ?;", shift.OrderID, shift.PaymentTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "查询支付信息失败")
		return
	}

	// 类型断言：确保 result 是 *sql.Rows 类型
	rows, ok := result.(*sql.Rows)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "数据库返回结果格式错误")
		return
	}
	defer rows.Close()

	// 假设查询只有一行数据，映射到 OrderInfo 结构体
	var paymentInfo PaymentInfo
	if rows.Next() {
		err := rows.Scan(&paymentInfo.PaymentID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "解析支付信息失败")
			return
		}
	} else {
		respondWithError(w, http.StatusNotFound, "未找到该支付信息")
		return
	}

	respondWithSuccess(w, paymentInfo)
}
func HandleGetWorkShift(w http.ResponseWriter, r *http.Request) {
	log.Println("GetWorkShift 被触发")
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

	// 执行查询获取订单信息
	result, err := db.ExecuteSQL(config.RoleDriver,
		"SELECT work_stime,work_etime,driver_id,car_id FROM work_table WHERE work_stime<= ? and work_etime>= ?;", shift.CurrentTime, shift.CurrentTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "查询工作信息失败")
		return
	}

	// 类型断言：确保 result 是 *sql.Rows 类型
	rows, ok := result.(*sql.Rows)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "数据库返回结果格式错误")
		return
	}
	defer rows.Close()

	// 映射到 WorkShift 结构体
	var workShifts []WorkShift
	for rows.Next() {
		var workShift WorkShift
		err := rows.Scan(&workShift.ShiftStart, &workShift.ShiftEnd, &workShift.DriverID, &workShift.VehicleNo)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "解析工作信息失败")
			return
		}
		workShifts = append(workShifts, workShift)

	}
	respondWithSuccess(w, workShifts)
}
