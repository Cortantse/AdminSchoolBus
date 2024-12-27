package utils

import (
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
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

// GenerateDateArray returns an array of 7 strings representing dates.
// The 7th element is today's date, and the preceding 6 elements are the dates of the previous 6 days.
func GenerateDateArray() []string {
	var dates [7]string

	// Get today's date
	now := time.Now()

	// Fill the array with the previous 6 days and today's date
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, i-6)                                  // Calculate the date for each day
		dates[6-i] = fmt.Sprintf("%02d.%02d", date.Month(), date.Day()) // Format as MM-DD
	}
	returnDates := make([]string, 0)
	for _, date := range dates {
		if date != "" {
			returnDates = append(returnDates, date)
		}
	}
	// reverse
	for i, j := 0, len(returnDates)-1; i < j; i, j = i+1, j-1 {
		returnDates[i], returnDates[j] = returnDates[j], returnDates[i]
	}

	return returnDates
}

// GenerateTimeArray returns an array of 24 strings representing hours and minutes.
// Each element represents the time from 23 hours ago up to the current hour.
func GenerateTimeArray() []string {
	var times [12]string

	// Get the current time
	now := time.Now()

	// Fill the array with the times for the last 23 hours and the current hour
	for i := 11; i >= 0; i-- {
		timePoint := now.Add(-time.Duration(i) * time.Hour)                          // Calculate the time for each hour
		times[11-i] = fmt.Sprintf("%02d:%02d", timePoint.Hour(), timePoint.Minute()) // Format as HH:MM
	}
	newTimes := make([]string, 0)
	for _, timePoint := range times {
		if timePoint != "" {
			newTimes = append(newTimes, timePoint)
		}
	}

	return newTimes
}

// GetClientIP 尝试获取客户端真实 IP 地址
func GetClientIP(r *http.Request) string {
	// 首先检查 X-Forwarded-For 头
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For 可能包含多个逗号分隔的 IP 地址，取第一个
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// 如果没有 X-Forwarded-For，则检查 X-Real-IP
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// 如果都没有，使用 RemoteAddr
	// RemoteAddr 格式为 "IP:Port"，需要提取 IP 部分
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // 如果解析失败，直接返回 RemoteAddr
	}
	return ip
}

// Round 四舍五入
func Round(satisfaction float64, i int) float64 {
	// 计算出需要舍入的位数
	pow := math.Pow(10, float64(i))
	// 将小数点后的位数乘以 pow，然后进行四舍五入，再除以 pow
	return math.Floor(satisfaction*pow+0.5) / pow
}
