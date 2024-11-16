package main

import (
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
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// 如果是 OPTIONS 请求，直接返回成功，处理预检请求。因为会默认发预检请求，所以要保证不会当成错误请求处理
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	// 确保请求是post请求
	if r.Method != http.MethodPost {
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

	// 简单验证账号密码，这里做硬编码，没有数据库还
	if loginReq.Username == "admin" && loginReq.Password == "admin" {
		// 返回成功
		response := ApiResponse{
			Code:    http.StatusOK,
			Message: "Login success",
			Data:    "pass",
		}
		json.NewEncoder(w).Encode(response)
	} else {
		// 返回失败
		response := ApiResponse{
			Code:    http.StatusUnauthorized,
			Message: "Login failed",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
	}
}

// 创造数据库连接实例
//
// Parameters:
//
//	null
//
// Returns:
// - error 如果连接失败，则返回错误信息
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

// 启动服务器外部链接
//
// Parameters:
//
//	null
//
// Returns:
//
//	The sum of a and b.
func initServer() error {
	port := config.AppConfig.Server.Port

	fmt.Println("Service is running on port", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("Service is not running properly, with error: ", err)
		return fmt.Errorf("service is not running properly, with error: %v", err)
	}
	return nil
}

// 1、优化所有报错**函数的位置**
// 2、报错时打印出错误信息

// 端口、密码等静态全局变量请去config.yaml修改
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

	return

	// 设置路由 =======
	http.HandleFunc("/api/login", loginHandler)

	// 启动连接服务 ======
	err = initServer()
	if err != nil {
		return
	}
}
