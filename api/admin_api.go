package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"login/auth"
	"login/config"
	"login/db"
	"login/exception"
	"login/log_service"
	"login/utils"
	"time"

	"net/http"
	"strconv"
)

// LoginRequest 用来解析前端传来的 JSON 数据
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type changeDataRequest struct {
	Dataset   string   `json:"dataset"`
	TableName string   `json:"table_name"`
	DataNames []string `json:"data_names"`
	Params    []string `json:"params"`
	Condition string   `json:"condition"`
	Token     string   `json:"token"`
}

// LoginResponse 用来返回给前端的 JSON 数据
type ApiResponse struct {
	Code           int    `json:"code"`
	Message        string `json:"message"`
	Data           string `json:"data,omitempty"`
	Role           int    `json:"role"`
	AdditionalInfo int    `json:"additional_info,omitempty"`
}

// @Summary 管理员修改信息
// @Description 接收前端post的请求，修改数据库中的信息
// @Tags admins
// @Accept json
// @Produce json
// @Param user body User true "User data"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Router /unknownnow [post]
func ChangeDataRequest(w http.ResponseWriter, r *http.Request) {
	// 获取用户请求数据
	var request changeDataRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		// 无法解析
		w.WriteHeader(http.StatusBadRequest)
		exception.PrintError(ChangeDataRequest, err)
		return
	}

	// 验证权限
	_, role, err := auth.VerifyAToken(request.Token)
	if err != nil {
		exception.PrintError(ChangeDataRequest, err)
		// 无权限
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 判断权限
	if role != config.RoleAdmin {
		exception.PrintError(ChangeDataRequest, err)
		// 无权限
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 选择数据库
	var dataset config.Role
	switch request.Dataset {
	case "passenger":
		dataset = config.RolePassenger
	case "driver":
		dataset = config.RoleDriver
	case "admin":
		dataset = config.RoleAdmin
	default:
		// 无效的数据集
		w.WriteHeader(http.StatusBadRequest)
		exception.PrintError(ChangeDataRequest, err)
		return
	}

	// 构造sql语句

	sqlStatement := "UPDATE " + request.TableName + " SET "

	if len(request.DataNames) != len(request.Params) || 0 == len(request.DataNames) {
		// 参数数量不正确
		w.WriteHeader(http.StatusBadRequest)
		exception.PrintError(ChangeDataRequest, err)
		return
	}

	for i := 0; i < len(request.DataNames); i++ {
		if i == 0 {
			sqlStatement += fmt.Sprintf("%s = ?", request.DataNames[i])
		} else {
			sqlStatement += fmt.Sprintf(", %s = ?", request.DataNames[i])
		}
	}

	sqlStatement += " WHERE " + request.Condition

	interfaces := utils.ConvertStringsToInterface(request.Params)

	// 执行sql语句
	_, err = db.ExecuteSQL(
		dataset,
		sqlStatement,
		interfaces...,
	)
	if err != nil {
		// 执行sql语句失败
		w.WriteHeader(http.StatusInternalServerError)
		exception.PrintError(ChangeDataRequest, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type DashBoardStatus struct {
	TotalUsers   int `json:"total_users"`
	TotalDrivers int `json:"total_drivers"`
	TotalAdmins  int `json:"total_admins"`
	// 错误数据
	SystemErrors   int `json:"system_errors"`
	SystemWarnings int `json:"system_warnings"`
	// 用户满意度
	UserSatisfaction       float64   `json:"user_satisfaction"`
	UserSatisfactionSeries []float64 `json:"user_satisfaction_series"`
	// 活跃用户数据
	DailyActiveSeries  []int `json:"daily_active_series"`
	WeeklyActiveSeries []int `json:"weekly_active_series"`
	// 收入数据
	RevenueSeries []float64 `json:"revenue_series"`
	// 健康度数据
	HealthIndexScore int `json:"health_index_score"`
	// 时间
	LastSevenDays       []string `json:"last_seven_days"`
	LastTwentyFourHours []string `json:"last_twenty_four_hours"`
	// 扣分原因
	DeductionReasons []map[string]string `json:"deduction_reasons"`
}

// @Summary 发送dashboard需要的信息
// @Description 接收前端fetch请求，返回dashboard需要信息
// @Tags users
// @Accept json
// @Produce json
// @Param user body User true "User data"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Router /users [post]
func GiveDashBoardInfo(w http.ResponseWriter, r *http.Request) {

	var totalUsers int
	var totalDrivers int
	var totalAdmins int
	// 获取当前所有数字
	resultUser, _ := db.ExecuteSQL(config.RoleAdmin, "SELECT COUNT(*) FROM userspass WHERE user_type = ?", 1)
	resultDrivers, _ := db.ExecuteSQL(config.RoleAdmin, "SELECT COUNT(*) FROM userspass WHERE user_type = ?", 2)
	resultAdmins, _ := db.ExecuteSQL(config.RoleAdmin, "SELECT COUNT(*) FROM userspass WHERE user_type = ?", 0)
	// 断言类型
	result1, _ := resultUser.(*sql.Rows)
	result2, _ := resultDrivers.(*sql.Rows)
	result3, _ := resultAdmins.(*sql.Rows)

	if result1.Next() && result2.Next() && result3.Next() {
		_ = result1.Scan(&totalUsers)
		_ = result2.Scan(&totalDrivers)
		_ = result3.Scan(&totalAdmins)
	}

	aDayErrors, aDayWarnings := log_service.GetADayErrorsAndWarnings()

	userSatisfaction, userSatisfactionSeries := calUserSatisfaction()

	dailyActiveUsers, weeklyActiveUsers := GetActiveUsers()

	data := DashBoardStatus{
		TotalUsers:             totalUsers,
		TotalDrivers:           totalDrivers,
		TotalAdmins:            totalAdmins,
		SystemErrors:           aDayErrors,
		SystemWarnings:         aDayWarnings,
		UserSatisfaction:       userSatisfaction,
		UserSatisfactionSeries: userSatisfactionSeries,
		DailyActiveSeries:      dailyActiveUsers,
		WeeklyActiveSeries:     weeklyActiveUsers,
		RevenueSeries:          calRevenue(),
		LastSevenDays:          utils.GenerateDateArray(),
		LastTwentyFourHours:    utils.GenerateTimeArray(),
	}

	data.HealthIndexScore, data.DeductionReasons = calculateHealthIndexScore(data)

	defer result1.Close()
	defer result2.Close()
	defer result3.Close()

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// 将数据转化为JSON格式并写入响应体
	json.NewEncoder(w).Encode(data)
}

func calculateHealthIndexScore(data DashBoardStatus) (int, []map[string]string) {
	// 每个error扣5分，每个warning扣1分，最高分别为60分和5分
	var score int
	var reasons []map[string]string

	score = 100

	if data.SystemErrors > 0 {

		minusScore := 0
		if data.SystemErrors >= 12 {
			minusScore = 60
		} else {
			minusScore = data.SystemErrors * 5
		}
		score -= minusScore

		reasons = append(reasons, map[string]string{"description": fmt.Sprintf("扣除%d分：系统错误", minusScore), "highlight": fmt.Sprintf("%d次", data.SystemErrors)})
	}

	if data.SystemWarnings > 0 {
		minusScore := 0
		if data.SystemWarnings >= 5 {
			minusScore = 5
		} else {
			minusScore = data.SystemWarnings
		}
		score -= minusScore
		reasons = append(reasons, map[string]string{"description": fmt.Sprintf("扣除%d分：系统警告", minusScore), "highlight": fmt.Sprintf("%d次", data.SystemWarnings)})
	}

	if data.UserSatisfaction < 60 {
		score -= 25
		reasons = append(reasons, map[string]string{"description": "扣除25分：用户满意度极低", "highlight": fmt.Sprintf("%.2f", data.UserSatisfaction)})
	} else if data.UserSatisfaction < 80 {
		score -= 15
		reasons = append(reasons, map[string]string{"description": "扣除15分：用户满意度较低", "highlight": fmt.Sprintf("%.2f", data.UserSatisfaction)})
	}

	return score, reasons
}

func GetActiveUsers() ([]int, []int) {
	config.AllowWarning = false

	var dailyActiveUsers []int
	var hourlyActiveUsers []int

	sqlStatement := getWeeklyActiveUsersSQL()
	result, err := db.ExecuteSQL(config.RoleAdmin, sqlStatement)
	if err != nil {
		exception.PrintError(GetActiveUsers, err)
		panic("could not get active")
	}
	rows, _ := result.(*sql.Rows)
	for rows.Next() {
		var loginDate string
		var uniqueUserCount int
		err = rows.Scan(&loginDate, &uniqueUserCount)
		if err != nil {
			exception.PrintError(GetActiveUsers, err)
			panic("could not get active")
		}
		dailyActiveUsers = append(dailyActiveUsers, uniqueUserCount)
	}
	defer rows.Close()

	sqlStatement = getDailyActiveUsersSQL()
	result, err = db.ExecuteSQL(config.RoleAdmin, sqlStatement)
	if err != nil {
		exception.PrintError(GetActiveUsers, err)
		panic("could not get active")
	}
	rows, _ = result.(*sql.Rows)
	for rows.Next() {
		var loginHour string
		var uniqueUserCount int
		err = rows.Scan(&loginHour, &uniqueUserCount)
		if err != nil {
			exception.PrintError(GetActiveUsers, err)
			panic("could not get active")
		}
		hourlyActiveUsers = append(hourlyActiveUsers, uniqueUserCount)
	}
	defer rows.Close()

	config.AllowWarning = true

	return dailyActiveUsers, hourlyActiveUsers
}

func getDailyActiveUsersSQL() string {
	return `SELECT
    calendar.hour AS login_hour,
    IFNULL(COUNT(DISTINCT ls.user_id), 0) AS unique_user_count
FROM
    (
        SELECT DATE_FORMAT(DATE_SUB(NOW(), INTERVAL n HOUR), '%Y-%m-%d %H:00:00') AS hour
        FROM (
            SELECT 0 AS n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4 UNION ALL SELECT 5
            UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9 UNION ALL SELECT 10 UNION ALL SELECT 11
        ) numbers
    ) calendar
LEFT JOIN
    loginsessions ls
ON
    DATE_FORMAT(ls.login_time, '%Y-%m-%d %H:00:00') = calendar.hour
GROUP BY
    calendar.hour
ORDER BY
    calendar.hour ASC`
}

func getWeeklyActiveUsersSQL() string {
	return `SELECT
    calendar.date AS login_date,
    IFNULL(COUNT(DISTINCT ls.user_id), 0) AS unique_user_count
FROM
    (
        SELECT CURDATE() AS date
        UNION ALL SELECT DATE_SUB(CURDATE(), INTERVAL 1 DAY)
        UNION ALL SELECT DATE_SUB(CURDATE(), INTERVAL 2 DAY)
        UNION ALL SELECT DATE_SUB(CURDATE(), INTERVAL 3 DAY)
        UNION ALL SELECT DATE_SUB(CURDATE(), INTERVAL 4 DAY)
        UNION ALL SELECT DATE_SUB(CURDATE(), INTERVAL 5 DAY)
        UNION ALL SELECT DATE_SUB(CURDATE(), INTERVAL 6 DAY)
    ) calendar
LEFT JOIN
    loginsessions ls
ON
    DATE(ls.login_time) = calendar.date
GROUP BY
    calendar.date
ORDER BY
    calendar.date ASC;
`
}

func calRevenue() []float64 {
	config.AllowWarning = false

	sqlStatement := GeneratePaymentSQL()
	result, err := db.ExecuteSQL(config.RolePassenger, sqlStatement)
	if err != nil {
		exception.PrintError(calRevenue, err)
		panic("could not give rate")
	}
	rows, _ := result.(*sql.Rows)

	var revenueSeries []float64

	for rows.Next() {
		var paymentDate string
		var totalPayment float64
		err = rows.Scan(&paymentDate, &totalPayment)
		if err != nil {
			exception.PrintError(calRevenue, err)
			panic("could not give rate")
		}
		revenueSeries = append(revenueSeries, totalPayment)

	}

	defer rows.Close()

	config.AllowWarning = true

	return revenueSeries
}

// 计算用户满意度
func calUserSatisfaction() (float64, []float64) {
	// 注意资源管理

	config.AllowWarning = false

	sqlStatement := "SELECT AVG(rating) from feedback"

	result, err := db.ExecuteSQL(config.RolePassenger, sqlStatement)
	if err != nil {
		exception.PrintError(calUserSatisfaction, err)
		panic("could not give rate")
	}
	rows, _ := result.(*sql.Rows)
	var userSatisfaction float64
	if rows.Next() {
		err = rows.Scan(&userSatisfaction)
		if err != nil {
			exception.PrintError(calUserSatisfaction, err)
			panic("could not give rate")
		}
	}
	// 注意翻倍
	userSatisfaction *= 20

	defer rows.Close()

	sqlStatement = returnSqlStaForUserSatis()

	result, err = db.ExecuteSQL(config.RolePassenger, sqlStatement)
	if err != nil {
		exception.PrintError(calUserSatisfaction, err)
		panic("could not give rate")
	}
	rows, _ = result.(*sql.Rows)

	var userSatisfactionSeries []float64

	if rows == nil {
		panic("no rows returned from query")
	}

	for rows.Next() {
		var avgRating float64
		var time string
		err = rows.Scan(&time, &avgRating)
		if err != nil {
			exception.PrintError(calUserSatisfaction, err)
			panic("could not give rate")
		}
		userSatisfactionSeries = append(userSatisfactionSeries, avgRating*20)
	}

	defer rows.Close()

	config.AllowWarning = true

	// 保留两位小数
	userSatisfaction = utils.Round(userSatisfaction, 2)
	for i, v := range userSatisfactionSeries {
		userSatisfactionSeries[i] = utils.Round(v, 2)
	}

	return userSatisfaction, userSatisfactionSeries
}

func GeneratePaymentSQL() string {
	return `SELECT
    calendar.date AS payment_date,
    IFNULL(SUM(pr.payment_amount), 0) AS total_payment
FROM
    (
        SELECT CURDATE() AS date
        UNION ALL SELECT DATE_SUB(CURDATE(), INTERVAL 1 DAY)
        UNION ALL SELECT DATE_SUB(CURDATE(), INTERVAL 2 DAY)
        UNION ALL SELECT DATE_SUB(CURDATE(), INTERVAL 3 DAY)
        UNION ALL SELECT DATE_SUB(CURDATE(), INTERVAL 4 DAY)
        UNION ALL SELECT DATE_SUB(CURDATE(), INTERVAL 5 DAY)
        UNION ALL SELECT DATE_SUB(CURDATE(), INTERVAL 6 DAY)
    ) calendar
LEFT JOIN
    payment_record pr
ON
    DATE(pr.payment_time) = calendar.date AND pr.payment_status = '1'
GROUP BY
    calendar.date
ORDER BY
    calendar.date ASC
`
}

func returnSqlStaForUserSatis() string {
	return `SELECT
    calendar.date AS feedback_date,
    IFNULL(AVG(feedback.rating), 5) AS avg_rating
FROM (
    SELECT CURDATE() - INTERVAL n DAY AS date
    FROM (SELECT 0 AS n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4 UNION ALL SELECT 5 UNION ALL SELECT 6) numbers
) calendar
LEFT JOIN feedback ON DATE(feedback.feedback_time) = calendar.date
GROUP BY calendar.date
ORDER BY calendar.date`
}

// AnswerHeartBeat 接收心跳检测请求
func AnswerHeartBeat(w http.ResponseWriter, r *http.Request) {
	// 正常就回复200即可
	w.WriteHeader(http.StatusOK)
}

// loginHandler 处理用户的登录请求。
//
// 此函数首先检查请求方法是否为POST，以确保是有效的登录请求。
// 接着解析请求体并对请求的JSON数据进行解码。
// 如果用户名和密码正确（在此示例中硬编码为"admin"/"admin"），则返回成功响应；否则返回登录失败响应。
//
// @param w http.ResponseWriter 用于将响应写回给客户端。
// @param r *http.Request 包含客户端请求的详细信息。
//
// @returns void 该函数无返回值，所有响应直接写入http.ResponseWriter。
//
// @throws error 当请求方法不为POST或请求解码失败时，会返回相应的HTTP错误响应。

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	// 允许跨域请求
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	// 如果是 OPTIONS 请求，直接返回成功，处理预检请求。因为会默认发预检请求，所以要保证不会当成错误请求处理
	if r.Method == http.MethodOptions {
		exception.PrintError(LoginHandler, fmt.Errorf("Options err"))
		w.WriteHeader(http.StatusOK)
		return
	}
	// 确保请求是post请求
	if r.Method != http.MethodPost {
		exception.PrintError(LoginHandler, fmt.Errorf("post err"))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 保证格式
	w.Header().Set("Content-Type", "application/json")

	// 解析请求
	var loginReq LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	// 释放资源
	defer r.Body.Close()

	if err != nil {
		log.Printf("请求解码失败: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "The request cannot be decoded",
		})
		return
	}
	// 获取userID => 查询密码表中 对应userID 和 密码 是否有内容
	// 1.获取userID
	var userID int
	result, err := db.ExecuteSQL(config.RoleAdmin, "SELECT user_id FROM usersaliases WHERE user_name = ?", loginReq.Username)
	if err != nil {
		exception.PrintError(LoginHandler, err)
		return
	}
	rows := result.(*sql.Rows)
	if rows.Next() {
		rows.Scan(&userID)
	} else {
		// 没有这个aliases，发送401
		response := ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "账户错误",
			Data:    "",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 2.联合查询
	result, err = db.ExecuteSQL(config.RoleAdmin, "SELECT * FROM userspass WHERE user_id = ? AND user_password_hash = ?", userID, loginReq.Password)
	if err != nil {
		exception.PrintError(LoginHandler, err)
		return
	}
	rows = result.(*sql.Rows)
	if rows.Next() {
		// 登陆成功
		var client auth.UserPass
		err = rows.Scan(&client.UserID, &client.UserPassword, &client.Role, &client.UserStatus)
		if err != nil {
			exception.PrintError(LoginHandler, err)
			return
		}
		// 检测用户 status
		if client.UserStatus != "active" {
			response := ApiResponse{
				Code:    http.StatusUnauthorized,
				Message: "账户状态异常",
				Data:    "",
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
			return
		}
		// 获取客户端信息
		clientInfo := GetClientInfo(r)
		role := determineRole(client.Role)
		// 这个地方要把int换成string
		userID := fmt.Sprintf("%d", userID)
		GenerateAndSendToken(w, role, userID, clientInfo, loginReq.Username)
	} else {
		// 密码错误，发送401
		response := ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "密码错误",
			Data:    "",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

}

// GenerateAndSendToken  公有函数，用于生成令牌并将其发送给客户端
func GenerateAndSendToken(w http.ResponseWriter, role config.Role, userId string, clientInfo string, userName string) {
	token, err := auth.GiveAToken(role, userId, clientInfo)
	if err != nil {
		exception.PrintError(GenerateAndSendToken, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ApiResponse{
			Code:    http.StatusInternalServerError,
			Message: "Token generation failed",
		})
		return
	}

	if role != config.RoleDriver {
		// 如果不是司机
		response := ApiResponse{
			Code:    http.StatusOK,
			Message: "Login success",
			Data:    token,
			Role:    role.Int(),
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	} else {
		// 如果是司机

		var driverID int
		result, err := db.ExecuteSQL(config.RoleDriver, "SELECT driver_id FROM driver_table WHERE driver_nickname = ?", userName)
		if err != nil {
			exception.PrintError(GenerateAndSendToken, err)
			exception.PrintError(GenerateAndSendToken, fmt.Errorf("无法找到该username对应的driverID"))
		}
		rows := result.(*sql.Rows)
		if rows.Next() {
			err := rows.Scan(&driverID)
			if err != nil {
				exception.PrintError(GenerateAndSendToken, err)
				return
			}

		} else {
			exception.PrintError(GenerateAndSendToken, fmt.Errorf("您尝试登陆了一个无法找到对应driverID的用户名"))
			response := ApiResponse{
				Code:    http.StatusBadRequest,
				Message: "您尝试登陆了一个无法找到对应driverID的用户名",
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := ApiResponse{
			Code:           http.StatusOK,
			Message:        "Login success",
			Data:           token,
			Role:           role.Int(),
			AdditionalInfo: driverID,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

}

// GetClientInfo 获取请求中的 User-Agent 信息
func GetClientInfo(r *http.Request) string {
	userAgent := r.Header.Get("User-Agent")
	return userAgent
}

// determineRole 根据 userType 返回对应的角色
func determineRole(userType int) config.Role {
	switch userType {
	case 0:
		return config.RoleAdmin
	case 1:
		return config.RolePassenger
	case 2:
		return config.RoleDriver
	default:
		return config.RolePassenger // 默认返回普通乘客角色
	}
}

// LogoutHandler 处理用户的登出请求
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// 允许跨域请求
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// 如果是 OPTIONS 请求，直接返回成功，处理预检请求
	if r.Method == http.MethodOptions {
		exception.PrintError(LoginHandler, fmt.Errorf("Options err"))
		w.WriteHeader(http.StatusOK)
		return
	}
	// 确保请求是post请求
	if r.Method != http.MethodPost {
		exception.PrintError(LoginHandler, fmt.Errorf("post err"))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从请求头获取令牌
	token := r.Header.Get("Authorization")
	if token == "" {
		exception.PrintError(LoginHandler, fmt.Errorf("GetToekn err"))
		http.Error(w, "Token is missing", http.StatusBadRequest)
		return
	}

	// 验证令牌
	userID, _, err := auth.VerifyAToken(token)
	if err != nil {
		exception.PrintWarning(LoginHandler, fmt.Errorf("VerifyAToken err"))
		exception.PrintWarning(LogoutHandler, err)
		return
	}

	// 更新token_revoked
	_, err = db.ExecuteSQL(config.RoleAdmin, "UPDATE tokens SET token_revoked = 1 WHERE user_id = ? and token_hash = ?", userID, token)
	if err != nil {
		exception.PrintWarning(LoginHandler, fmt.Errorf("VerifyAToken err"))
		exception.PrintWarning(LogoutHandler, err)
		return
	}

	// 更新数据库中的 token_revoked 字段

	// 返回登出成功的响应
	response := ApiResponse{
		Code:    http.StatusOK,
		Message: "Logout success",
	}
	json.NewEncoder(w).Encode(response)
}

func ValidateTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Token is missing", http.StatusBadRequest)
		return
	}

	userID, role, err := auth.VerifyAToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// 成功后作为active_user
	InsertActiveUser(userID, token, r)

	// 返回验证成功响应
	json.NewEncoder(w).Encode(ApiResponse{
		Code:    http.StatusOK,
		Message: "Token is valid",
		Data:    fmt.Sprintf("UserID: %s, Role: %s", userID, role),
		Role:    role.Int(),
	})
}

func InsertActiveUser(userID string, token string, r *http.Request) {
	// 获取发送请求方的 IP 地址
	ip := utils.GetClientIP(r)

	sqlStatement := db.ConstructInsertSQL("loginsessions", []string{"login_status", "login_time", "login_ip_address",
		"user_id", "token_id"})

	login_time, _ := utils.RegularizeTimeForMySQL(time.Now().String())

	// get token id

	tokenID, err := db.ExecuteSQL(config.RoleAdmin, "SELECT token_id FROM tokens WHERE token_hash = ?", token)
	if err != nil {
		exception.PrintError(InsertActiveUser, err)
		return
	}
	rows := tokenID.(*sql.Rows)
	var tokenIDInt int
	if rows.Next() {
		rows.Scan(&tokenIDInt)
	}

	defer rows.Close()

	_, err = db.ExecuteSQL(config.RoleAdmin, sqlStatement, 1, login_time, ip, userID, tokenIDInt)
	if err != nil {
		exception.PrintError(InsertActiveUser, err)
		return
	}

}

// @Summary 获得表格数据
// @Description 根据具体内容获得数据
// @Tags admins
// @Accept json
// @Produce json
// @Param user body User true "User data"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Router /unknownnow [post]
func GetTableData(w http.ResponseWriter, r *http.Request) {
	// 定义用户结构
	type User struct {
		ID               string   `json:"id"`
		Aliases          []string `json:"aliases"`
		AccountType      string   `json:"accountType"`
		AccountStatus    string   `json:"accountStatus"`
		UnlockTime       string   `json:"unlockTime,omitempty"`
		RegistrationTime string   `json:"registrationTime"`
	}

	//// 模拟数据
	//var users = []User{
	//	{"1001", []string{"alias1", "alias2"}, "admin", "active", "", "2024-01-01 08:00:00"},
	//	{"1002", []string{"win"}, "user", "locked", "2024-12-20 12:00:00", "2023-05-15 14:20:00"},
	//	{"1003", []string{"guest"}, "user", "active", "", "2024-06-10 10:30:00"},
	//}

	// 获取查询参数
	keyword := r.URL.Query().Get("keyword")
	accountType := r.URL.Query().Get("accountType")
	accountStatus := r.URL.Query().Get("accountStatus")
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	// 先获得aliases
	sqlStatementBefore := "SELECT user_id, user_name FROM usersaliases WHERE user_id lIKE ?"
	resultBefore, err := db.ExecuteSQL(config.RoleAdmin, sqlStatementBefore, "%"+keyword+"%")
	if err != nil {
		exception.PrintError(GetTableData, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 创建map方便后续O(1)查找
	myMap := make(map[string][]string)
	rowsBefore := resultBefore.(*sql.Rows)
	for rowsBefore.Next() {
		var userID string
		var userAlias string
		err = rowsBefore.Scan(&userID, &userAlias)
		if err != nil {
			exception.PrintError(GetTableData, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		myMap[userID] = append(myMap[userID], userAlias)
	}

	// 获取查询结果，其中如果有keyword那么模糊查询，不包含aliases, aliases单独处理
	sqlStatement := "SELECT u.user_id, u.user_type, u.user_status, l.user_locked_time, i.user_registry_date FROM userspass u LEFT JOIN userslocked l ON u.user_id = l.user_id LEFT JOIN usersinfo i ON u.user_id = i.user_id WHERE u.user_id lIKE ?"
	result, err := db.ExecuteSQL(config.RoleAdmin, sqlStatement, "%"+keyword+"%")
	if err != nil {
		exception.PrintError(GetTableData, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rows := result.(*sql.Rows)
	var users []User

	for rows.Next() {
		var user User
		var userLockedTime sql.NullString
		err = rows.Scan(&user.ID, &user.AccountType, &user.AccountStatus, &userLockedTime, &user.RegistrationTime)
		if err != nil {
			exception.PrintError(GetTableData, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// 处理none
		if userLockedTime.Valid {
			user.UnlockTime = userLockedTime.String
		} else {
			user.UnlockTime = "-"
		}

		// 处理身份
		switch user.AccountType {
		case "0":
			user.AccountType = "admin"
		case "1":
			user.AccountType = "user"
		case "2":
			user.AccountType = "driver"
		}

		// 加入aliases
		user.Aliases = myMap[user.ID]
		// 放入结果数组
		users = append(users, user)
	}

	// 默认分页参数
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(sizeStr)
	if size < 1 {
		size = 10
	}

	// 过滤用户数据
	var filteredUsers []User
	for _, user := range users {
		if accountType != "" && user.AccountType != accountType {
			continue
		}
		if accountStatus != "" && user.AccountStatus != accountStatus {
			continue
		}
		filteredUsers = append(filteredUsers, user)
	}

	// 分页处理
	start := (page - 1) * size
	end := start + size
	if start > len(filteredUsers) {
		start = len(filteredUsers)
	}
	if end > len(filteredUsers) {
		end = len(filteredUsers)
	}
	paginatedUsers := filteredUsers[start:end]

	// 构造返回数据
	response := map[string]interface{}{
		"data":  paginatedUsers,
		"total": len(filteredUsers),
	}

	// 设置响应头并返回 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type deleteDataRequest struct {
	Dataset   string `json:"dataset"`
	TableName string `json:"table_name"`
	Condition string `json:"condition"`
	Token     string `json:"token"`
}

// @Summary 管理员删除信息
// @Description 接收前端post的请求，删除数据库的信息
// @Tags admins
// @Accept json
// @Produce json
// @Param user body User true "User data"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Router /unknownnow [post]
func DeleteDataRequest(w http.ResponseWriter, r *http.Request) {
	// 获取用户请求数据
	var request deleteDataRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		// 无法解析
		w.WriteHeader(http.StatusBadRequest)
		exception.PrintError(ChangeDataRequest, err)
		return
	}

	// 验证权限
	_, role, err := auth.VerifyAToken(request.Token)
	if err != nil {
		exception.PrintError(ChangeDataRequest, err)
		// 无权限
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 判断权限
	if role != config.RoleAdmin {
		exception.PrintError(ChangeDataRequest, err)
		// 无权限
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 选择数据库
	var dataset config.Role
	switch request.Dataset {
	case "passenger":
		dataset = config.RolePassenger
	case "driver":
		dataset = config.RoleDriver
	case "admin":
		dataset = config.RoleAdmin
	default:
		// 无效的数据集
		w.WriteHeader(http.StatusBadRequest)
		exception.PrintError(ChangeDataRequest, err)
		return
	}

	// 构造sql语句

	sqlStatement := fmt.Sprintf("DELETE FROM %s WHERE %s", request.TableName, request.Condition)

	// 执行sql语句
	_, err = db.ExecuteSQL(
		dataset,
		sqlStatement,
	)
	if err != nil {
		// 执行sql语句失败
		w.WriteHeader(http.StatusInternalServerError)
		exception.PrintError(ChangeDataRequest, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type insertDataRequest struct {
	Dataset   string   `json:"dataset"`
	TableName string   `json:"table_name"`
	DataNames []string `json:"data_names"`
	Params    []string `json:"params"`
	Token     string   `json:"token"`
}

// @Summary 管理员添加信息
// @Description 接收前端post的请求，添加数据库的信息
// @Tags admins
// @Accept json
// @Produce json
// @Param user body User true "User data"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Router /unknownnow [post]
func InsertDataRequest(w http.ResponseWriter, r *http.Request) {
	// 获取用户请求数据
	var request insertDataRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		// 无法解析
		w.WriteHeader(http.StatusBadRequest)
		exception.PrintError(ChangeDataRequest, err)
		return
	}

	// 对于有<time>参数的项，替换成实际时间
	for i, v := range request.Params {
		if v == "<time>" {
			reqTime, _ := utils.RegularizeTimeForMySQL(time.Now().String())
			request.Params[i] = reqTime
		}
	}

	// 验证权限
	_, role, err := auth.VerifyAToken(request.Token)
	if err != nil {
		exception.PrintError(ChangeDataRequest, err)
		// 无权限
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 判断权限
	if role != config.RoleAdmin {
		exception.PrintError(ChangeDataRequest, err)
		// 无权限
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 选择数据库
	var dataset config.Role
	switch request.Dataset {
	case "passenger":
		dataset = config.RolePassenger
	case "driver":
		dataset = config.RoleDriver
	case "admin":
		dataset = config.RoleAdmin
	default:
		// 无效的数据集
		w.WriteHeader(http.StatusBadRequest)
		exception.PrintError(ChangeDataRequest, err)
		return
	}

	// 构造sql语句
	sqlStatement := db.ConstructInsertSQL(request.TableName, request.DataNames)

	// 转换数据
	interfaces := utils.ConvertStringsToInterface(request.Params)

	// 执行sql语句

	// 执行sql语句
	_, err = db.ExecuteSQL(
		dataset,
		sqlStatement,
		interfaces...,
	)
	if err != nil {
		// 执行sql语句失败
		w.WriteHeader(http.StatusInternalServerError)
		exception.PrintError(ChangeDataRequest, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
