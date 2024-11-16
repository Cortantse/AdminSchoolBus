package main

import (
	"encoding/json"
	"fmt"
	"log"
	"login/driverShift" // 引入 driverShift 包
	"login/gps"         // 引入 gps 包
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

// // GPSData 结构体，用于接收前端的 GPS 数据
// type GPSData struct {
// 	Latitude  float64 `json:"latitude"`
// 	Longitude float64 `json:"longitude"`
// 	Timestamp string  `json:"timestamp"` // 或 time.Time, 取决于数据格式
// }

// CORS 中间件
// func enableCORS(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Access-Control-Allow-Origin", "*")                   // 允许所有来源
// 		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS") // 允许的请求方法
// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")       // 允许的请求头

// 		// 如果是预检请求（OPTIONS），则直接返回
// 		if r.Method == "OPTIONS" {
// 			w.WriteHeader(http.StatusOK)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

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

	// 验证账号密码，这里还是用明文传输的，后续加密啊$$$$$￥￥￥
	// 创建数组存储查询结果
	var results []auth.UserPass
	// 创建变量数组方便传递变量
	var params = []interface{}{
		loginReq.Username, loginReq.Password,
	}
	// 查询结果
	err = db.Select(config.RoleAdmin, "usersPass", &results, true,
		[]string{}, []string{"user_id = ? AND user_password_hash = ?"}, params, "", 1, 0, "", "")
	if err != nil {
		exception.PrintError(loginHandler, err)
		return
	}

	if len(results) != 0 {
		// 返回成功
		response := ApiResponse{
			Code:    http.StatusOK,
			Message: "Login success",
			Data:    "pass", //这里传令牌，最好要加密啊$$$$$￥￥￥
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
		w.Header().Set("Access-Control-Allow-Origin", "*")                   // 允许所有来源
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS") // 允许的请求方法
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")       // 允许的请求头

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

// func test () {
// 	// 示例：添加一个驾驶员对象
// 	gpsModule.CreateDriver("driver1", 34.0522, -118.2437)
// }

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

	// 初始化 GPS 模块 =====
	gpsModule := gps.NewGPSModule()

	// 启动令牌服务 ======
	err = auth.InitTokenService()
	if err != nil {
		print(err.Error())
	}

	// 创建 ServeMux 路由
	mux := http.NewServeMux()

	// 设置路由，处理工作流相关的请求
	mux.HandleFunc("/driverShift/provideInfo", driverShift.ProvideInfo) // 提供选择信息
	mux.HandleFunc("/driverShift/start", driverShift.HandleShiftStart)  // 处理上班信息
	mux.HandleFunc("/driverShift/end", driverShift.HandleShiftEnd)      // 处理下班信息
	mux.HandleFunc("/drivers", gpsModule.GetAllDriversHandler)          // 用于获取所有驾驶员位置信息的接口
	mux.HandleFunc("/updateLocation", gpsModule.Handler)                // 用于接收并处理 GPS 信息的接口
	mux.HandleFunc("/api/login", loginHandler)                          // 设置登录处理路由
	// 设置路由，接收 GPS 数据的端点
	// mux.HandleFunc("/api/gps", handleGPSData) // 接收 GPS 数据
	mux.HandleFunc("/createDriver", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "missing driver id", http.StatusBadRequest)
			return
		}
		latitude := 34.0522
		longitude := -118.2437
		driver, err := gpsModule.CreateDriver(id, latitude, longitude)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(driver)
	})

	// 使用 CORS 中间件
	corsHandler := enableCORS(mux)

	// 启动连接服务 ======
	err = initServer(corsHandler)
	if err != nil {
		return
	}

	// // 使用 CORS 中间件
	// corsHandler := enableCORS(mux)

	// // 启动服务器
	// const port = ":8080"
	// fmt.Println("Service is running on port", port)
	// err := http.ListenAndServe(port, corsHandler)
	// if err != nil {
	// 	fmt.Println("Service is not running properly, with error: ", err)
	// }
}
