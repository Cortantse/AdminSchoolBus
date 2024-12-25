package api

import (
	"encoding/json"
	"fmt"
	"login/exception"
	"math"
	"net/http"
	"strconv"
)

// 响应结构体，用来传递结果、错误或警告信息
type Response struct {
	Result float64 `json:"result"`
}

// 某种机制
func divide(a float32, b float32) (error, float32) {
	// 除0报错
	if b == 0 {
		exception.PrintError(divide, fmt.Errorf("division by zero"))
		return fmt.Errorf("division by zero"), 0
	}
	result := a / b
	// 溢出检测
	if math.IsInf(float64(result), 0) {
		exception.PrintWarning(divide, fmt.Errorf("division overflow"))
		return fmt.Errorf("division overflow"), 0
	}

	return nil, result
}

// 处理函数
func ReceiveDivisionRequest(w http.ResponseWriter, r *http.Request) {
	// 从查询参数获取两个数字
	// 假设请求是类似于 /divide?a=10&b=2 的形式
	aStr := r.URL.Query().Get("a")
	bStr := r.URL.Query().Get("b")

	// 尝试转换为浮动类型
	a, err := strconv.Atoi(aStr)
	if err != nil {
		http.Error(w, "Invalid parameter 'a'", http.StatusBadRequest)
		return
	}

	b, err := strconv.Atoi(bStr)
	if err != nil {
		http.Error(w, "Invalid parameter 'b'", http.StatusBadRequest)
		return
	}

	err, result := divide(float32(a), float32(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 正常返回200
	w.WriteHeader(http.StatusOK)

	// 构造响应结构体
	response := Response{
		Result: float64(result),
	}

	// 返回JSON格式的响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
