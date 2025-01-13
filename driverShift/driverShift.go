package driverShift

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"login/config"
	"login/db"
	"login/gps" // 引入 gps 模块
	"net/http"
	"time"
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

type DriverInfo struct {
	Driver_id        string `json:"driver_id"`        // 驾驶员编号
	Driver_avatar    string `json:"driver_avatar"`    //驾驶员头像url
	Driver_name      string `json:"driver_name"`      //驾驶员名称
	Driver_sex       string `json:"driver_sex"`       //驾驶员性别
	Driver_tel       string `json:"driver_tel"`       //驾驶员电话
	Driver_wages     string `json:"driver_wages"`     //驾驶员电话
	Driver_isworking string `json:"driver_isworking"` //驾驶员状态
}

type Comments struct {
	Name    string `json:"student_name"`    // 评论昵称
	Content string `json:"comment_content"` // 评论内容
	Ctime   string `json:"comment_time"`    // 评论时间
	Avatar  string `json:"avatar"`          // 评论人头像url
}

type CommentResponse struct {
	Comments []Comments `json:"comments"`
}

// 路径记录结构体
type RouteRecord struct {
	Time string `json:"time"`  // 时间戳
	GPSX int    `json:"gps_x"` // GPS X 坐標
	GPSY int    `json:"gps_y"` // GPS Y 坐標
}

// var module := gps.NewGPSModule()

// 用于更新车辆运行状态
func updateVehicleStatus(carID string, newStatus string) error {

	status := 0
	if newStatus == "正常运营" {
		status = 1
	}
	if newStatus == "试通行" {
		status = 2
	}
	if newStatus == "休息" {
		status = 3
	}
	_, err := db.ExecuteSQL(config.RoleDriver, "UPDATE car_table SET car_isusing = ? WHERE car_id = ?", status, carID)
	if err != nil {
		return fmt.Errorf("更新车辆状态失败: %w", err)
	}
	return nil
}

func updateDriverStatus(driverID string, status bool) error {

	_, err := db.ExecuteSQL(config.RoleDriver, "UPDATE driver_table SET driver_isworking = ? WHERE driver_id = ?", status, driverID)
	if err != nil {
		return fmt.Errorf("更新司机上班状态失败: %w", err)
	}
	return nil
}

func createWorkTable(driverID string, carID string, routeID int) error {
	timeNow := time.Now().Format("2006-01-02 15:04:05")
	sql := "INSERT INTO work_table (work_stime,driver_id,route_id,car_id) VALUES (?,?,?,?)"
	_, err := db.ExecuteSQL(config.RoleDriver, sql, timeNow, driverID, routeID, carID)
	if err != nil {
		return fmt.Errorf("创建工作表失败: %w", err)
	}
	return nil
}

func modifyWorkTable(driverID string, carID string) error {
	timeNow := time.Now().Format("2006-01-02 15:04:05")
	sql := "UPDATE work_table SET work_etime = ? WHERE driver_id = ? AND car_id = ? AND (work_etime IS NULL ) "
	_, err := db.ExecuteSQL(config.RoleDriver, sql, timeNow, driverID, carID)
	if err != nil {
		return fmt.Errorf("更新工作表失败: %w", err)
	}
	return nil
}

func modifyDriverInfo(tempDriverInfo DriverInfo) error {
	d_id := tempDriverInfo.Driver_id
	d_sex := 0
	d_status := 0
	if tempDriverInfo.Driver_sex == "男" {
		d_sex = 1
	}
	if tempDriverInfo.Driver_isworking == "1" {
		d_status = 1
	} else if tempDriverInfo.Driver_isworking == "2" {
		d_status = 2
	}

	_, err := db.ExecuteSQL(config.RoleDriver, "UPDATE driver_table SET driver_name = ?, driver_sex = ?, driver_tel = ?, driver_isworking = ? WHERE driver_id = ?", tempDriverInfo.Driver_name, d_sex, tempDriverInfo.Driver_tel, d_status, d_id)

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

func respondWithSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// 通用成功響應函數
// func respondWithSuccess(w http.ResponseWriter, message string) {
// 	json.NewEncoder(w).Encode(map[string]string{"message": message})
// }

// 处理上班：验证信息并创建 GPS 驾驶员对象
func HandleShiftStart(w http.ResponseWriter, r *http.Request, gps_api *gps.GPSAPI) {
	log.Printf("接收到上班信息")
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
	if err := updateDriverStatus(shift.DriverID, true); err != nil {
		respondWithError(w, http.StatusInternalServerError, "司机上班状态更新失败")
		return
	}

	if err := createWorkTable(shift.DriverID, shift.VehicleNo, shift.RouteID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "创建工作表失败")
		return
	}

	// 创建驾驶员对象
	log.Printf("driverid = %s\n", shift.DriverID)
	_, err = gps_api.CreateDriver(shift.DriverID) // 初始纬度和经度为 0
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "创建驾驶员失败")
		return
	}

	respondWithSuccess(w, "上班信息处理成功")
}

// 处理下班：验证信息并删除 GPS 驾驶员对象
func HandleShiftEnd(w http.ResponseWriter, r *http.Request, gps_api *gps.GPSAPI) {
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
	if err := updateVehicleStatus(shift.VehicleNo, shift.VehicleStatus); err != nil {
		respondWithError(w, http.StatusInternalServerError, "车辆下班状态更新失败")
		return
	}

	if err := updateDriverStatus(shift.DriverID, false); err != nil {
		respondWithError(w, http.StatusInternalServerError, "司机下班状态更新失败")
		return
	}

	if err := modifyWorkTable(shift.DriverID, shift.VehicleNo); err != nil {
		respondWithError(w, http.StatusInternalServerError, "工作表下班状态更新失败")
		return
	}

	// 删除驾驶员对象
	err = gps_api.DeleteDriver(shift.DriverID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "删除工作中的驾驶员失败")
		return
	}

	respondWithSuccess(w, "下班信息处理成功")
}

// 模拟更新车辆状态的函数
// func updateVehicleStatus(vehicleNo, status string) error {
// 	fmt.Printf("车辆 %s 状态已更新为 %s\n", vehicleNo, status)
// 	return nil
// }

func HandleShiftInfo(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, "POST, OPTIONS")
	log.Printf("接收到修改个人信息")
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
	var shift DriverInfo
	err = json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		log.Printf("JSON 解码失败: %v", err)
		respondWithError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}
	log.Printf("接收到的解码后数据: %+v", shift)

	// 更新车辆状态
	if err := modifyDriverInfo(shift); err != nil {
		respondWithError(w, http.StatusInternalServerError, "司机状态更新失败")
		return
	}

	respondWithSuccess(w, "司机信息修改成功")
}

func GetDriverData(w http.ResponseWriter, r *http.Request) {
	log.Println("GetDriverData 被触发")
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
	var shift DriverInfo
	err = json.NewDecoder(r.Body).Decode(&shift)
	if err != nil {
		log.Printf("JSON 解码失败: %v", err)
		respondWithError(w, http.StatusBadRequest, "请求数据解析失败")
		return
	}
	log.Printf("接收到的解码后数据: %+v", shift)

	// 执行查询获取司机信息
	result, err := db.ExecuteSQL(config.RoleDriver, "SELECT driver_id, driver_name, driver_avatar, driver_sex, driver_tel, driver_wages ,driver_isworking FROM driver_table WHERE driver_id = ?", shift.Driver_id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "查询司机信息失败")
		return
	}

	// 类型断言：确保 result 是 *sql.Rows 类型
	rows, ok := result.(*sql.Rows)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "数据库返回结果格式错误")
		return
	}
	defer rows.Close()

	// 假设查询只有一行数据，映射到 DriverInfo 结构体
	var driverInfo DriverInfo
	if rows.Next() {
		err := rows.Scan(&driverInfo.Driver_id, &driverInfo.Driver_name, &driverInfo.Driver_avatar, &driverInfo.Driver_sex, &driverInfo.Driver_tel, &driverInfo.Driver_wages, &driverInfo.Driver_isworking)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "解析司机信息失败")
			return
		}
	} else {
		respondWithError(w, http.StatusNotFound, "未找到该司机信息")
		return
	}

	respondWithSuccess(w, driverInfo)
}

func GetComments(w http.ResponseWriter, r *http.Request) {
	log.Println("GetComments 被触发")
	setCORSHeaders(w, "POST, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "仅支持 POST 请求")
		return
	}

	result, err := db.ExecuteSQL(config.RolePassenger, "SELECT student_name,comment_content,comment_time,avatar FROM passenger_comment")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "读取评论失败")
		return
	}

	// 类型断言：确保 result 是 *sql.Rows 类型
	rows, ok := result.(*sql.Rows)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "数据库返回结果格式错误")
		return
	}
	defer rows.Close()

	var comments []Comments
	for rows.Next() {
		var comment Comments
		err := rows.Scan(&comment.Name, &comment.Content, &comment.Ctime, &comment.Avatar)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "解析评论失败")
			return
		}
		log.Printf("查询到的评论数据: Name: %s, Content: %s, Time: %s, Avatar: %s",
			comment.Name, comment.Content, comment.Ctime, comment.Avatar)
		comments = append(comments, comment)
	}

	if len(comments) == 0 {
		respondWithError(w, http.StatusNotFound, "未找到评论信息")
		return
	}

	// 返回评论数据
	response := CommentResponse{Comments: comments}
	respondWithSuccess(w, response)
}
