package main

import (
	"fmt"
	"login/api"
	"os"

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
	err = db.InitDB(config.RoleDriver)
	if err != nil {
		fmt.Println("driver数据库连接失败，错误信息为：", err)
		return fmt.Errorf("driver数据库连接失败，错误信息为：%v", err)
	}
	fmt.Println("driver数据库连接成功")

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

	// 验证码
	mux.HandleFunc("/api/register", auth.HandleRegistry)
}

func main() {
	// 初始化全局参数 ======
	err := config.LoadConfig("config.yaml")
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "C:\\Users\\27785\\GolandProjects\\login\\service-account-file.json")
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
	mux.HandleFunc("/api/login", api.LoginHandler)
	mux.HandleFunc("/api/logout", api.LogoutHandler)               // 设置登出处理路由
	mux.HandleFunc("/api/validateToken", api.ValidateTokenHandler) // 设置登出处理路由

	//用于处理驾驶员上下班
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		driverShift.HandleShiftStart(w, r, gps_api)
	})
	mux.HandleFunc("/modifyDriverInfo", driverShift.HandleShiftInfo)

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
