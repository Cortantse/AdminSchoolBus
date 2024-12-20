package utils

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// RegularizeTimeForMySQL 适合mysql的标准时间格式
func RegularizeTimeForMySQL(input string) (string, error) {
	// 定义输入的时间格式（包括时区）
	const inputFormat = "2006-01-02 15:04:05 -0700 MST"
	// 定义输出的 MySQL DATETIME 格式
	const outputFormat = "2006-01-02 15:04:05"

	if len(input) < 19 {
		log.Printf("intput string is too short")
		return "", fmt.Errorf("error in RemoveTimezone: input string is too short")
	}

	// 只保留前几个字符

	return input[:19], nil
}

// TrimExtraSpaces 用于去除字符串前后的多余空格
func TrimExtraSpaces(input string) string {
	if len(input) == 0 {
		return ""
	}
	if strings.HasPrefix(input, " ") {
		input = strings.TrimPrefix(input, " ")
		return TrimExtraSpaces(input)
	}
	if strings.HasSuffix(input, " ") {
		input = strings.TrimSuffix(input, " ")
	}
	return input
}

// GetFormattedCurrentTime 获取当前时间，并格式化，方便日志工具使用
func GetFormattedCurrentTime() string {
	return time.Now().String() + ": "
}

// ConvertStringsToInterface 将string[]换位db_api接收的[]interface{};
func ConvertStringsToInterface(intput []string) []interface{} {
	var interfaces []interface{}
	for _, param := range intput {
		interfaces = append(interfaces, param)
	}
	return interfaces
}
