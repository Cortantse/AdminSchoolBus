package exception

import (
	"log"
	"reflect"
	"runtime"
)

// PrintError 打印错误发生在什么函数中
func PrintError(fn interface{}, err error) {
	// 定义颜色 ANSI 转义序列
	red := "\033[31m"  // 红色字体
	bold := "\033[1m"  // 加粗
	reset := "\033[0m" // 重置样式

	// 获取函数名
	pc := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	if pc != nil {
		log.Printf("%s%sError occurs in %s: %s%s\n", bold, red, pc.Name(), err.Error(), reset)
	} else {
		log.Printf("%s%sError occurs in unknown function: %s%s\n", bold, red, err.Error(), reset)
	}
}

// PrintWarning 打印警告发生在什么函数中
func PrintWarning(fn interface{}, err error) {
	// 定义颜色 ANSI 转义序列
	yellow := "\033[33m" // 黄色字体
	bold := "\033[1m"    // 加粗
	reset := "\033[0m"   // 重置样式

	// 获取函数名
	pc := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	if pc != nil {
		log.Printf("%s%sWarning in %s: %s%s\n", bold, yellow, pc.Name(), err.Error(), reset)
	} else {
		log.Printf("%s%sWarning in unknown function: %s%s\n", bold, yellow, err.Error(), reset)
	}
}
