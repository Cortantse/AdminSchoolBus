package utils

import (
	"fmt"
	"login/exception"
	"strings"
)

// RegularizeTimeForMySQL 适合mysql的标准时间格式
func RegularizeTimeForMySQL(input string) (string, error) {
	// 定义输入的时间格式（包括时区）
	const inputFormat = "2006-01-02 15:04:05 -0700 MST"
	// 定义输出的 MySQL DATETIME 格式
	const outputFormat = "2006-01-02 15:04:05"

	if len(input) < 19 {
		exception.PrintError(RegularizeTimeForMySQL, fmt.Errorf("intput string is too short"))
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
