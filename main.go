package main

import (
	"encoding/json"
	"fmt"
	"log" // 引入 driverShift 包
	"login/api"

	"login/driverShift"
	"login/gps" // 引入 gps 包
	"net/http"

	// "log"
	"login/auth"
	"login/config"
	"login/db"
	"login/exception"
	// "net/http"
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

func loginHandler(w http.ResponseWriter, r *http.Request) {

	// 允许跨域请求
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	// 如果是 OPTIONS 请求，直接返回成功，处理预检请求。因为会默认发预检请求，所以要保证不会当成错误请求处理
	if r.Method == http.MethodOptions {
		exception.PrintError(loginHandler, fmt.Errorf("Options err"))
		w.WriteHeader(http.StatusOK)
		return
	}
	// 确保请求是post请求
	if r.Method != http.MethodPost {
		exception.PrintError(loginHandler, fmt.Errorf("post err"))
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

	// 验证账号密码，这里还是用明文传输的，后续加密啊$$$$$￥￥￥
	// 创建数组存储查询结果
	var results []auth.UserPass
	// 创建变量数组方便传递变量
	var params = []interface{}{
		loginReq.Username, loginReq.Password,
	}
	// 查询结果
	err = db.SelectEasy(config.RoleAdmin, "usersPass", &results, true,
		[]string{}, []string{"user_id = ? AND user_password_hash = ?"}, params, "", 1, 0, "", "")
	if err != nil {
		exception.PrintError(loginHandler, err)
		return
	}

	if len(results) != 0 {
		// 获取客户端信息
		clientInfo := GetClientInfo(r)
		userType := results[0].Role
		role := determineRole(userType)
		userID := results[0].UserID
		GenerateAndSendToken(w, role, userID, clientInfo)
	} else {
		// 返回失败
		response := ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "账户或密码错误",
			Data:    "",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
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
	}
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

// logoutHandler 处理用户的登出请求
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// 允许跨域请求
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// 如果是 OPTIONS 请求，直接返回成功，处理预检请求
	if r.Method == http.MethodOptions {
		exception.PrintError(loginHandler, fmt.Errorf("Options err"))
		w.WriteHeader(http.StatusOK)
		return
	}
	// 确保请求是post请求
	if r.Method != http.MethodPost {
		exception.PrintError(loginHandler, fmt.Errorf("post err"))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从请求头获取令牌
	token := r.Header.Get("Authorization")
	if token == "" {
		exception.PrintError(loginHandler, fmt.Errorf("GetToekn err"))
		http.Error(w, "Token is missing", http.StatusBadRequest)
		return
	}

	// 验证令牌
	userID, _, err := auth.VerifyAToken(token)
	if err != nil {
		exception.PrintError(loginHandler, fmt.Errorf("VerifyAToken err"))
		exception.PrintError(logoutHandler, err)
		return
	}

	// 更新token_revoked
	_, err = db.ExecuteSQL(config.RoleAdmin, "UPDATE tokens SET token_revoked = 1 WHERE user_id = ? and token_hash = ?", userID, token)
	if err != nil {
		exception.PrintError(loginHandler, fmt.Errorf("VerifyAToken err"))
		exception.PrintError(logoutHandler, err)
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

func validateTokenHandler(w http.ResponseWriter, r *http.Request) {
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

// 创造数据库连接实例
func initDatasetCon() error {
	err := db.InitDB(config.RoleAdmin)
	if err != nil {
		fmt.Println("admin数据库连接失败，错误信息为：", err)
		return fmt.Errorf("admin数据库连接失败，错误信息为：%v", err)
	}
	fmt.Println("admin数据库连接成功")

	// ** 由于还没有你们的数据库，暂时先注释下面了，你们用的时候记得开开 **
	//err = db.InitDB(config.RolePassenger)
	//if err != nil {
	//	fmt.Println("passenger数据库连接失败，错误信息为：", err)
	//	return fmt.Errorf("passenger数据库连接失败，错误信息为：%v", err)
	//}
	//fmt.Println("passenger数据库连接成功")
	//
	//err = db.InitDB(config.RoleDriver)
	//if err != nil {
	//	fmt.Println("driver数据库连接失败，错误信息为：", err)
	//	return fmt.Errorf("driver数据库连接失败，错误信息为：%v", err)
	//}
	//fmt.Println("driver数据库连接成功")

	return nil
}

// CORS 中间件
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")                            // 允许所有来源
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")          // 允许的请求方法
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // 允许的请求头

		// 如果是预检请求（OPTIONS），则直接返回
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 启动服务器外部链接
func initServer(cors http.Handler) error {
	port := config.AppConfig.Server.Port

	fmt.Println("Service is running on port", port)
	err := http.ListenAndServe(port, cors)
	if err != nil {
		fmt.Println("Service is not running properly, with error: ", err)
		return fmt.Errorf("service is not running properly, with error: %v", err)
	}
	return nil
}

func testForToken(err error) {
	// 获取一个令牌
	token, err := auth.GiveAToken(config.RoleDriver, "2", "")
	if err != nil {
		print(err.Error())
	}
	// 验证令牌，并获得令牌所有者的信息
	userID, role, err := auth.VerifyAToken(token)
	if err != nil {
		exception.PrintWarning(auth.VerifyAToken, err)
	}

	fmt.Printf("UserID is %s, role is %s\n", userID, role)
}

// RegisterAdmin 注册管理员服务
func RegisterAdmin(mux *http.ServeMux) {
	// 注册 HTTP API 路由
	mux.HandleFunc("/admin_home/dashboard", api.GiveDashBoardInfo)
	mux.HandleFunc("/heartbeat", api.AnswerHeartBeat)
}

func main() {
	// 初始化全局参数 ======
	err := config.LoadConfig("config.yaml")
	if err != nil {
		print(err.Error())
	}

	// 设置数据库连接 =====
	err = initDatasetCon()
	if err != nil {
		print(err.Error())
	}

	// 启动令牌服务 ======
	err = auth.InitTokenService()
	if err != nil {
		print(err.Error())
	}

	// 创建一个 GPSAPI 实例，用于将 GPSModule 的核心逻辑对外提供为 HTTP 接口
	gps_api := gps.InitGPSAPI()

	// 创建 ServeMux 路由
	mux := http.NewServeMux()

	// 验证url
	mux.HandleFunc("/api/login", loginHandler)
	mux.HandleFunc("/api/logout", logoutHandler)               // 设置登出处理路由
	mux.HandleFunc("/api/validateToken", validateTokenHandler) // 设置登出处理路由

	//用于处理驾驶员上下班
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		gpsModule := &gps.GPSModule{}
		driverShift.HandleShiftStart(w, r, gpsModule)
	})

	// 注册 GPSAPI 提供的 HTTP 接口到路由器中。
	gps_api.RegisterRoutes(mux)

	// 注册后端服务器服务
	RegisterAdmin(mux)

	// 使用 CORS 中间件
	corsHandler := enableCORS(mux)

	// 启动连接服务 ======
	err = initServer(corsHandler)
	if err != nil {
		return
	}

}
