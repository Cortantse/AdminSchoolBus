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
	"net/http"
)

// LoginRequest 用来解析前端传来的 JSON 数据
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 用来返回给前端的 JSON 数据
type ApiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
	Role    int    `json:"role"`
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

	type DashBoardStatus struct {
		TotalUsers   int `json:"total_users"`
		TotalDrivers int `json:"total_drivers"`
		TotalAdmins  int `json:"total_admins"`
	}
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

	data := DashBoardStatus{
		TotalUsers:   totalUsers,
		TotalDrivers: totalDrivers,
		TotalAdmins:  totalAdmins,
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// 将数据转化为JSON格式并写入响应体
	json.NewEncoder(w).Encode(data)
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
		// 获取客户端信息
		clientInfo := GetClientInfo(r)
		role := determineRole(client.Role)
		// 这个地方要把int换成string
		userID := fmt.Sprintf("%d", userID)
		GenerateAndSendToken(w, role, userID, clientInfo)
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
func GenerateAndSendToken(w http.ResponseWriter, role config.Role, userId string, clientInfo string) {
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

	response := ApiResponse{
		Code:    http.StatusOK,
		Message: "Login success",
		Data:    token,
		Role:    role.Int(),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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
		exception.PrintError(LoginHandler, fmt.Errorf("VerifyAToken err"))
		exception.PrintError(LogoutHandler, err)
		return
	}

	// 更新token_revoked
	_, err = db.ExecuteSQL(config.RoleAdmin, "UPDATE tokens SET token_revoked = 1 WHERE user_id = ? and token_hash = ?", userID, token)
	if err != nil {
		exception.PrintError(LoginHandler, fmt.Errorf("VerifyAToken err"))
		exception.PrintError(LogoutHandler, err)
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

	// 返回验证成功响应
	json.NewEncoder(w).Encode(ApiResponse{
		Code:    http.StatusOK,
		Message: "Token is valid",
		Data:    fmt.Sprintf("UserID: %s, Role: %s", userID, role),
	})
}
