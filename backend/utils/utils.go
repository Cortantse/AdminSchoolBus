package utils

import (
	"fmt"
	"login/exception"
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
