package log_service

import (
	"log"
	"os"
)

// 定义两个日志实例
var (
	WebSocketLogger *log.Logger
	GPSLogger       *log.Logger
)

// 初始化日志
func InitLoggers() error {
	// 创建统一的日志文件
	logFile, err := os.OpenFile("application.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// 为 WebSocket 模块创建日志实例
	WebSocketLogger = log.New(logFile, "WEBSOCKET: ", log.Ldate|log.Ltime|log.Lshortfile)

	// 为 GPS 模块创建日志实例（与 WebSocket 共享同一个日志文件）
	GPSLogger = log.New(logFile, "GPS: ", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}
