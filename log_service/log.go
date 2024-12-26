package log_service

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"log"
	"login/utils"
	"os"
)

// 全局文件变量
var file *os.File

var HourlyErrorsNum int
var HourlyWarningNum int

var ADayErrors []int
var ADayWarnings []int

func InitLogService() {
	// 打开日志文件
	var err error
	file, err = os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("日志服务已启动")

	HourlyErrorsNum = 0

	// 设置日志输出到文件
	log.SetOutput(file)
	log.Println("2019-01-01 12:00:00 日志服务已启动==============")

	// 恢复日志输出到标准输出（控制台）
	log.SetOutput(os.Stdout)

	// 初始化数组
	ADayErrors = make([]int, 24)
	ADayWarnings = make([]int, 24)

	// 顺带启动Cron日志
	go newCron()
}

// 创建一个 Cron 实例，并启动 Cron 调度器，每小时整点执行函数
func newCron() {
	// 创建支持 6 个域的 Cron 实例
	c := cron.New(cron.WithSeconds())

	// 添加任务，每小时整点执行
	_, err := c.AddFunc("0 0 * * * *", printCurrentTime) // 6 个域
	if err != nil {
		log.Printf("添加任务时出错: %v\n", err)
		return
	}

	// 启动 Cron 调度器
	c.Start()

	// 打印消息，表示服务已启动
	log.Println("Cron 服务已启动，服务将在每小时自动运行")

	// 阻止程序退出
	select {}
}

func printCurrentTime() {
	msg := make([]string, 0)
	msg = append(msg, "=====准点报时======")
	msg = append(msg, utils.GetFormattedCurrentTime()+" 服务器正在运行")

	msg = append(msg, utils.GetFormattedCurrentTime()+" \033[32m1h内发现"+fmt.Sprintf("%d", HourlyErrorsNum)+"个致命错误\033[0m")
	WriteToBoth(msg)

	// 重置操作
	HourlyErrorsNum = 0
	HourlyErrorsNum = 0

	// 将当前这个时刻的数组清空，因为才进入
	// 获取当前时间
	currentTime := time.Now()

	// 提取小时 0-23
	currentHour := currentTime.Hour()

	ADayErrors[currentHour] = 0
	ADayWarnings[currentHour] = 0
}

func WriteToBoth(msg []string) {
	StartWriteLog()
	for _, v := range msg {
		log.Println(v)
	}
	EndWriteLog()
	for _, v := range msg {
		log.Println(v)
	}
}

func StartWriteLog() {
	log.SetOutput(file)
}

func EndWriteLog() {
	log.SetOutput(os.Stdout)
}

func GetADayErrorsAndWarnings() (int, int) {
	errors := 0
	warnings := 0
	for _, v := range ADayErrors {
		errors += v
	}
	for _, v := range ADayWarnings {
		warnings += v
	}
	return errors, warnings
}
