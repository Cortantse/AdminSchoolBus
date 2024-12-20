package exception

import (
	"fmt"
	"log"
	"reflect"
	"runtime"

	"time"


	"login/log_service"
)

// PrintError 打印错误发生在什么函数中
func PrintError(fn interface{}, err error) {
	// 定义颜色 ANSI 转义序列
	red := "\033[31m"  // 红色字体
	bold := "\033[1m"  // 加粗
	reset := "\033[0m" // 重置样式

	// 获取函数名
	pc := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())



	addErrorOrWarning(true)


	var str string
	if pc != nil {
		str = fmt.Sprintf("%s%sError occurs in %s: %s%s\n", bold, red, pc.Name(), err.Error(), reset)
	} else {
		str = fmt.Sprintf("%s%sError occurs in unknown function: %s%s\n", bold, red, err.Error(), reset)
	}
	log_service.WriteToBoth([]string{str})
}

// PrintWarning 打印警告发生在什么函数中
func PrintWarning(fn interface{}, err error) {
	// 定义颜色 ANSI 转义序列
	yellow := "\033[33m" // 黄色字体
	bold := "\033[1m"    // 加粗
	reset := "\033[0m"   // 重置样式

	addErrorOrWarning(false)

	// 获取函数名
	pc := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())

	if pc != nil {
		log.Printf("%s%sWarning in %s: %s%s\n", bold, yellow, pc.Name(), err.Error(), reset)
	} else {
		log.Printf("%s%sWarning in unknown function: %s%s\n", bold, yellow, err.Error(), reset)
	}
}

func addErrorOrWarning(ifError bool) {
	// 获取当前时间
	currentTime := time.Now()

	// 提取小时 0-23
	currentHour := currentTime.Hour()

	if ifError {
		log_service.HourlyErrorsNum += 1
		log_service.ADayErrors[currentHour] += 1
	} else {
		log_service.HourlyWarningNum += 1
		log_service.ADayWarnings[currentHour] += 1
	}
}
