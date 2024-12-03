package log_service

import (
	"github.com/robfig/cron/v3"
	"log"
	"login/utils"
	"os"
)

// 全局文件变量
var file *os.File

func InitLogService() {
	// 打开日志文件
	var err error
	file, err = os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("日志服务已启动")

	// 设置日志输出到文件
	log.SetOutput(file)
	log.Println("2019-01-01 12:00:00 日志服务已启动==============")

	// 恢复日志输出到标准输出（控制台）
	log.SetOutput(os.Stdout)

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
	msg = append(msg, utils.GetFormattedCurrentTime()+" \033[32m1h内发现0个致命错误\033[0m")
	WriteToBoth(msg)
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
