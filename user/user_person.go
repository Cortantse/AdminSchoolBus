package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"login/config"
	"login/db"
	"login/exception"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Response 定义返回的 JSON 数据结构
type Response struct {
	StudentName string `json:"Student_Name"`
}

// GetUserNameHandler 根据用户 ID 返回姓名
func GetUserNameHandler(w http.ResponseWriter, r *http.Request) {
	// 设置 CORS 头部
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理 OPTIONS 预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		exception.PrintWarning(GetUserNameHandler, fmt.Errorf("option err"))
		return
	}
	// 校验请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST", http.StatusMethodNotAllowed)
		exception.PrintWarning(GetUserNameHandler, fmt.Errorf("post err"))
		return
	}

	// 解析请求参数
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		exception.PrintWarning(GetUserNameHandler, fmt.Errorf("parsefrom err"))
		return
	}

	userID := r.FormValue("userID") // 获取 userID
	if userID == "" {
		exception.PrintWarning(GetUserNameHandler, fmt.Errorf("UserID is empty"))
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	// 校验 userID 是否为数字
	if _, err := strconv.Atoi(userID); err != nil {
		http.Error(w, "Invalid userID format. Must be a number.", http.StatusBadRequest)
		exception.PrintWarning(GetUserNameHandler, err)
		return
	}
	// 执行 SQL 查询
	sqlQuery := "SELECT student_name FROM student_information WHERE user_id = ?"
	result, err := db.ExecuteSQL(config.RolePassenger, sqlQuery, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		exception.PrintWarning(GetUserNameHandler, err)
		return
	}

	// 解析查询结果
	rows, ok := result.(*sql.Rows)
	if !ok {
		http.Error(w, "Unexpected result type", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 提取姓名
	var name string
	if rows.Next() {
		if err := rows.Scan(&name); err != nil {
			http.Error(w, "Failed to retrieve user data", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 返回 JSON 数据
	response := Response{StudentName: name}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}

// FullStudent 定义完整的学生信息，包括头像
type FullStudent struct {
	StudentAccount string `json:"student_account"` // 学生账号，11位数字
	StudentNumber  int    `json:"student_number"`  // 学号
	StudentName    string `json:"student_name"`    // 姓名
	Grade          int    `json:"grade"`           // 年级
	Major          string `json:"major"`           // 专业
	Phone          string `json:"phone"`           // 电话号码
	Avatar         string `json:"avatar"`          // 确保有 Avatar 字段
}

// GetUserInfoHandler 返回完整的用户信息，包括头像
func GetUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 设置 CORS 头部
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理 OPTIONS 预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 校验请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求参数
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("userID") // 获取 userID
	if userID == "" {
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	// 校验 userID 是否为数字
	if _, err := strconv.Atoi(userID); err != nil {
		http.Error(w, "Invalid userID format. Must be a number.", http.StatusBadRequest)
		exception.PrintError(GetUserInfoHandler, err)
		return
	}
	// 执行 SQL 查询
	sqlQuery := "SELECT student_account, student_number, student_name, grade, major, phone, avatar FROM student_information WHERE user_id = ?"

	result, err := db.ExecuteSQL(config.RolePassenger, sqlQuery, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		exception.PrintError(GetUserInfoHandler, err)
		return
	}

	// 解析查询结果
	rows, ok := result.(*sql.Rows)
	if !ok {
		http.Error(w, "Unexpected result type", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 提取数据
	var student FullStudent
	if rows.Next() {
		if err := rows.Scan(&student.StudentAccount, &student.StudentNumber, &student.StudentName,
			&student.Grade, &student.Major, &student.Phone, &student.Avatar); err != nil { // 添加 &student.Avatar
			http.Error(w, "Failed to retrieve user data", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 返回 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(student); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}

// UpdateUserRequest 定义接收的数据结构
type UpdateUserRequest struct {
	StudentAccount string `json:"student_account"`
	Name           string `json:"name"`
	Grade          int    `json:"grade"`
	Major          string `json:"major"`
	Phone          string `json:"phone"`
	Avatar         string `json:"avatar"` // 确保有 Avatar 字段
	UserID         int    `json:"user_id"`
}

// UpdateUserInfoHandler 更新用户信息
func UpdateUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		exception.PrintError(GetUserNameHandler, err)
		return
	}

	// 执行 SQL 更新
	sqlQuery := "UPDATE student_information SET student_name = ?, grade = ?, major = ?, phone = ?, avatar = ? WHERE user_id = ?"
	_, err := db.ExecuteSQL(config.RolePassenger, sqlQuery, req.Name, req.Grade, req.Major, req.Phone, req.Avatar, req.UserID)
	if err != nil {
		http.Error(w, "Failed to update user information", http.StatusInternalServerError)
		return
	}
	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User information updated successfully"})
}

// UploadAvatarHandler 处理头像上传
func UploadAvatarHandler(w http.ResponseWriter, r *http.Request) {
	// 设置 CORS 头部
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理 OPTIONS 预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 校验请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST", http.StatusMethodNotAllowed)
		return
	}

	// 解析 multipart form
	err := r.ParseMultipartForm(5 << 20) // 限制上传大小为 5MB
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// 获取用户ID（假设从 JWT 中解析）
	userID := r.FormValue("userID")
	if userID == "" {
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	// 获取文件
	file, handler, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 检查文件类型
	if !strings.HasPrefix(handler.Header.Get("Content-Type"), "image/") {
		http.Error(w, "Only image files are allowed", http.StatusBadRequest)
		return
	}

	// 生成唯一的文件名
	fileExtension := filepath.Ext(handler.Filename)
	newFileName := fmt.Sprintf("avatar_%s_%d%s", userID, time.Now().Unix(), fileExtension)
	savePath := filepath.Join("uploads", "avatars", newFileName)

	// 创建目录（如果不存在）
	err = os.MkdirAll(filepath.Dir(savePath), os.ModePerm)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 保存文件
	dst, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 生成头像 URL（假设您的服务器可以通过 /uploads/avatars/ 访问）
	avatarURL := fmt.Sprintf("/uploads/avatars/%s", newFileName)

	// 更新数据库中的 avatar 字段
	sqlQuery := "UPDATE student_information SET avatar = ? WHERE user_id = ?"
	_, err = db.ExecuteSQL(config.RolePassenger, sqlQuery, avatarURL, userID)
	if err != nil {
		http.Error(w, "Failed to update user information", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": avatarURL})
}

type RideCoupon struct {
	RideCouponID int    `json:"ride_coupon_id"`
	ExpiryDate   string `json:"expiry_date"`
	UseStatus    string `json:"use_status"`
}

type DiscountCoupon struct {
	CouponID       int     `json:"coupon_id"`
	DiscountAmount float64 `json:"discount_amount"`
	ExpiryDate     string  `json:"expiry_date"`
	UseStatus      string  `json:"use_status"`
}

type UserCouponsResponse struct {
	RideCoupons     []RideCoupon     `json:"rideCoupons"`
	DiscountCoupons []DiscountCoupon `json:"discountCoupons"`
}

func GetUserCouponsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		exception.PrintError(GetUserCouponsHandler, err)
		return
	}

	userID := r.FormValue("userID")
	if userID == "" {
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	// 校验 userID
	if _, err := strconv.Atoi(userID); err != nil {
		http.Error(w, "Invalid userID format. Must be a number.", http.StatusBadRequest)
		exception.PrintError(GetUserCouponsHandler, err)
		return
	}

	// 根据 student_account 查 student_number
	sqlGetStudentNumber := "SELECT student_number FROM student_information WHERE user_id = ?"
	result, err := db.ExecuteSQL(config.RolePassenger, sqlGetStudentNumber, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		exception.PrintError(GetUserCouponsHandler, err)
		return
	}

	rows, ok := result.(*sql.Rows)
	if !ok {
		http.Error(w, "Unexpected result type", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var studentNumber int
	if rows.Next() {
		if err := rows.Scan(&studentNumber); err != nil {
			http.Error(w, "Failed to retrieve user data", http.StatusInternalServerError)
			exception.PrintError(GetUserCouponsHandler, err)
			return
		}
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 获取当前时间
	currentDate := time.Now().Format("2006-01-02")

	// 查询 ride_coupon
	sqlRide := "SELECT ride_coupon_id, expiry_date, use_status FROM ride_coupon WHERE student_number = ?"
	rideResult, err := db.ExecuteSQL(config.RolePassenger, sqlRide, studentNumber)
	if err != nil {
		http.Error(w, "Failed to fetch ride coupons", http.StatusInternalServerError)
		exception.PrintError(GetUserCouponsHandler, err)
		return
	}
	rideRows, ok := rideResult.(*sql.Rows)
	if !ok {
		http.Error(w, "Unexpected result type", http.StatusInternalServerError)
		return
	}
	defer rideRows.Close()

	var rideCoupons []RideCoupon
	for rideRows.Next() {
		var rc RideCoupon
		var expiryDate string
		var useStatusInt int
		if err := rideRows.Scan(&rc.RideCouponID, &expiryDate, &useStatusInt); err != nil {
			http.Error(w, "Failed to scan ride coupons", http.StatusInternalServerError)
			return
		}

		// 检查是否过期
		if expiryDate < currentDate {
			rc.UseStatus = "已过期"
		} else {
			if useStatusInt == 0 {
				rc.UseStatus = "未使用"
			} else if useStatusInt == 1 {
				rc.UseStatus = "已使用"
			} else {
				rc.UseStatus = "未知状态"
			}
		}
		rc.ExpiryDate = expiryDate
		rideCoupons = append(rideCoupons, rc)
	}

	// 查询 discount_coupon
	sqlDiscount := "SELECT coupon_id, discount_amount, expiry_date, use_status FROM discount_coupon WHERE student_number = ?"
	discountResult, err := db.ExecuteSQL(config.RolePassenger, sqlDiscount, studentNumber)
	if err != nil {
		http.Error(w, "Failed to fetch discount coupons", http.StatusInternalServerError)
		exception.PrintError(GetUserCouponsHandler, err)
		return
	}
	discountRows, ok := discountResult.(*sql.Rows)
	if !ok {
		http.Error(w, "Unexpected result type", http.StatusInternalServerError)
		return
	}
	defer discountRows.Close()

	var discountCoupons []DiscountCoupon
	for discountRows.Next() {
		var dc DiscountCoupon
		var expiryDate string
		var useStatusInt int
		if err := discountRows.Scan(&dc.CouponID, &dc.DiscountAmount, &expiryDate, &useStatusInt); err != nil {
			http.Error(w, "Failed to scan discount coupons", http.StatusInternalServerError)
			exception.PrintError(GetUserCouponsHandler, err)
			return
		}

		// 检查是否过期
		if expiryDate < currentDate {
			dc.UseStatus = "已过期"
		} else {
			if useStatusInt == 0 {
				dc.UseStatus = "未使用"
			} else if useStatusInt == 1 {
				dc.UseStatus = "已使用"
			} else {
				dc.UseStatus = "未知状态"
			}
		}
		dc.ExpiryDate = expiryDate
		discountCoupons = append(discountCoupons, dc)
	}

	// 返回 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	response := UserCouponsResponse{
		RideCoupons:     rideCoupons,
		DiscountCoupons: discountCoupons,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		exception.PrintError(GetUserCouponsHandler, err)
	}
}

type Feedback struct {
	FeedbackID      int    `json:"feedback_id"`
	StudentNumber   string `json:"student_number"`
	OrderID         int    `json:"order_id"`
	Rating          int    `json:"rating"`
	FeedbackContent string `json:"feedback_content"`
	FeedbackTime    string `json:"feedback_time"`
}

func GetFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		exception.PrintError(GetFeedbackHandler, err)
		return
	}

	userID := r.FormValue("userID")
	if userID == "" {
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	// 校验 userID
	if _, err := strconv.Atoi(userID); err != nil {
		http.Error(w, "Invalid userID format. Must be a number.", http.StatusBadRequest)
		exception.PrintError(GetFeedbackHandler, err)
		return
	}
	// 根据 student_account 查 student_number
	sqlGetStudentNumber := "SELECT student_number FROM student_information WHERE user_id = ?"
	result, err := db.ExecuteSQL(config.RolePassenger, sqlGetStudentNumber, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		exception.PrintError(GetFeedbackHandler, err)
		return
	}

	rows, ok := result.(*sql.Rows)
	if !ok {
		http.Error(w, "Unexpected result type", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var studentNumber int
	if rows.Next() {
		if err := rows.Scan(&studentNumber); err != nil {
			http.Error(w, "Failed to retrieve user data", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 查询 feedback
	sqlFeedback := "SELECT feedback_id, student_number, order_id, rating, feedback_content, feedback_time FROM feedback WHERE student_number = ?"
	feedbackResult, err := db.ExecuteSQL(config.RolePassenger, sqlFeedback, studentNumber)
	if err != nil {
		http.Error(w, "Failed to fetch feedback data", http.StatusInternalServerError)
		exception.PrintError(GetFeedbackHandler, err)
		return
	}

	feedbackRows, ok := feedbackResult.(*sql.Rows)
	if !ok {
		http.Error(w, "Unexpected result type", http.StatusInternalServerError)
		return
	}
	defer feedbackRows.Close()

	var feedbacks []Feedback
	for feedbackRows.Next() {
		var feedback Feedback
		if err := feedbackRows.Scan(&feedback.FeedbackID, &feedback.StudentNumber, &feedback.OrderID, &feedback.Rating, &feedback.FeedbackContent, &feedback.FeedbackTime); err != nil {
			http.Error(w, "Failed to scan feedback data", http.StatusInternalServerError)
			exception.PrintError(GetFeedbackHandler, err)
			return
		}
		feedbacks = append(feedbacks, feedback)
	}

	// 返回 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(feedbacks); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		exception.PrintError(GetFeedbackHandler, err)
	}
}

func AddFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理 OPTIONS 请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 校验请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体
	var feedback Feedback
	if err := json.NewDecoder(r.Body).Decode(&feedback); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		exception.PrintError(AddFeedbackHandler, err)
		return
	}
	//feedbackID := rand.Intn(900000) + 100000 // 生成六位随机数

	// 根据 student_account 查 student_number
	sqlGetStudentNumber := "SELECT student_number FROM student_information WHERE student_account = ?"
	result, err := db.ExecuteSQL(config.RolePassenger, sqlGetStudentNumber, feedback.StudentNumber)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		exception.PrintError(GetFeedbackHandler, err)
		return
	}

	rows, ok := result.(*sql.Rows)
	if !ok {
		http.Error(w, "Unexpected result type", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var studentNumber int
	if rows.Next() {
		if err := rows.Scan(&studentNumber); err != nil {
			http.Error(w, "Failed to retrieve user data", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 插入评价数据到数据库
	insertQuery := "INSERT INTO feedback (student_number, order_id, rating, feedback_content, feedback_time) VALUES (?, ?, ?, ?, ?)"

	_, err = db.ExecuteSQL(config.RolePassenger, insertQuery,
		studentNumber, feedback.OrderID,
		feedback.Rating, feedback.FeedbackContent, feedback.FeedbackTime)

	if err != nil {
		http.Error(w, "Failed to add feedback", http.StatusInternalServerError)
		exception.PrintError(AddFeedbackHandler, err)
		return
	}

	updateQuery := "UPDATE order_information SET is_rated = 1 WHERE order_id = ?"
	_, err = db.ExecuteSQL(config.RolePassenger, updateQuery, feedback.OrderID)
	if err != nil {
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		exception.PrintError(AddFeedbackHandler, err)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Feedback added successfully"})
}

// 定义订单信息结构体
type Order struct {
	OrderID            int    `json:"order_id"`
	StudentAccount     string `json:"student_account"`
	DriverID           int    `json:"driver_id"`
	CarID              int    `json:"car_id"`
	PickupStationName  string `json:"pickup_station_name"`
	DropoffStationName string `json:"dropoff_station_name"`
	PickupTime         string `json:"pickup_time"`
	Status             string `json:"status"`
	PaymentID          int    `json:"payment_id"`
	IsRated            bool   `json:"is_rated"`
}

// 定义支付信息结构体
type Payment struct {
	PaymentID     int     `json:"payment_id"`
	OrderID       int     `json:"order_id"`
	VehicleID     int     `json:"vehicle_id"`
	PaymentAmount float64 `json:"payment_amount"`
	PaymentMethod string  `json:"payment_method"`
	PaymentTime   string  `json:"payment_time"`
	PaymentStatus string  `json:"payment_status"`
}

// 定义返回的完整响应结构体
type UserOrdersResponse struct {
	Orders   []Order   `json:"orders"`
	Payments []Payment `json:"payments"`
}

func GetUserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理 OPTIONS 请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 校验请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求参数
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		exception.PrintError(GetUserOrdersHandler, err)
		return
	}

	userID := r.FormValue("userID")
	if userID == "" {
		http.Error(w, "Missing userID parameter", http.StatusBadRequest)
		return
	}

	// 校验 userID
	if _, err := strconv.Atoi(userID); err != nil {
		http.Error(w, "Invalid userID format. Must be a number.", http.StatusBadRequest)
		exception.PrintError(GetUserOrdersHandler, err)
		return
	}
	// 根据 userID 查 student_account
	sqlGetStudentNumber := "SELECT student_account FROM student_information WHERE user_id = ?"
	result, err := db.ExecuteSQL(config.RolePassenger, sqlGetStudentNumber, userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		exception.PrintError(GetFeedbackHandler, err)
		return
	}

	rows, ok := result.(*sql.Rows)
	if !ok {
		http.Error(w, "Unexpected result type", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var rawStudentAccount []byte
	var studentAccount string
	if rows.Next() {
		if err := rows.Scan(&rawStudentAccount); err != nil {
			http.Error(w, "Failed to retrieve user data", http.StatusInternalServerError)
			return
		}
		studentAccount = string(rawStudentAccount)
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 查询订单信息
	orderQuery := "SELECT order_id, student_account, driver_id, car_id, pickup_station_name, dropoff_station_name, pickup_time, status, payment_id, is_rated FROM order_information WHERE student_account = ?"
	orderResult, err := db.ExecuteSQL(config.RolePassenger, orderQuery, studentAccount)
	if err != nil {
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		exception.PrintError(GetUserOrdersHandler, err)
		return
	}
	orderRows, ok := orderResult.(*sql.Rows)
	if !ok {
		http.Error(w, "Unexpected result type", http.StatusInternalServerError)
		return
	}
	defer orderRows.Close()

	var orders []Order
	for orderRows.Next() {
		var order Order
		var StatusInt int
		if err := orderRows.Scan(&order.OrderID, &order.StudentAccount, &order.DriverID, &order.CarID,
			&order.PickupStationName, &order.DropoffStationName, &order.PickupTime, &StatusInt, &order.PaymentID, &order.IsRated); err != nil {

			http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
			exception.PrintError(GetUserOrdersHandler, err)
			return
		}
		if StatusInt == 0 {
			order.Status = "进行中"
		} else if StatusInt == 1 {
			order.Status = "已完成"
		} else {
			order.Status = "未知状态"
		}
		orders = append(orders, order)
	}

	// 查询支付信息
	paymentQuery := "SELECT payment_id, order_id, vehicle_id, payment_amount, payment_method, payment_time, payment_status FROM payment_record WHERE order_id IN (SELECT order_id FROM order_information WHERE student_account = ?)"
	paymentResult, err := db.ExecuteSQL(config.RolePassenger, paymentQuery, studentAccount)
	if err != nil {
		http.Error(w, "Failed to fetch discount coupons", http.StatusInternalServerError)
		exception.PrintError(GetUserOrdersHandler, err)
		return
	}
	paymentRows, ok := paymentResult.(*sql.Rows)
	if !ok {
		http.Error(w, "Unexpected result type", http.StatusInternalServerError)
		return
	}
	defer paymentRows.Close()

	var payments []Payment
	var payStatusInt int
	var payMethodInt int

	for paymentRows.Next() {
		var payment Payment
		if err := paymentRows.Scan(&payment.PaymentID, &payment.OrderID, &payment.VehicleID, &payment.PaymentAmount,
			&payMethodInt, &payment.PaymentTime, &payStatusInt); err != nil {
			http.Error(w, "Failed to fetch payments", http.StatusInternalServerError)
			return
		}
		if payStatusInt == 0 {
			payment.PaymentStatus = "失败"
		} else if payStatusInt == 1 {
			payment.PaymentStatus = "成功"
		} else {
			payment.PaymentStatus = "未知"
		}

		if payMethodInt == 0 {
			payment.PaymentMethod = "微信"
		} else if payMethodInt == 1 {
			payment.PaymentMethod = "支付宝"
		} else {
			payment.PaymentMethod = "未知"
		}
		payments = append(payments, payment)
	}

	// 返回 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	response := UserOrdersResponse{
		Orders:   orders,
		Payments: payments,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		exception.PrintError(GetUserOrdersHandler, err)
	}
}

// RegisterUser 注册路由
func RegisterUser(mux *http.ServeMux) {
	mux.HandleFunc("/api/getUserName", GetUserNameHandler)
	mux.HandleFunc("/api/getUserAll", GetUserInfoHandler)
	mux.HandleFunc("/api/updateUser", UpdateUserInfoHandler)
	mux.HandleFunc("/api/uploadAvatar", UploadAvatarHandler) // 新增头像上传路由
	mux.HandleFunc("/api/getUserCoupons", GetUserCouponsHandler)
	mux.HandleFunc("/api/getFeedback", GetFeedbackHandler)
	mux.HandleFunc("/api/getOrders", GetUserOrdersHandler)
	mux.HandleFunc("/api/addFeedback", AddFeedbackHandler)

}
