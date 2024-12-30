package api

import (
	"encoding/json"
	"fmt"
	"log"
	"login/exception"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// Message 结构体表示单条消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

type Data struct {
	Text string `json:"text"`
}

type Request struct {
	Command string `json:"command"`
}

func ReceiveAIRequest(w http.ResponseWriter, r *http.Request) {
	// 获取command
	var command Request
	err := json.NewDecoder(r.Body).Decode(&command)
	if err != nil {
		exception.PrintError(ReceiveAIRequest, err)
		return
	}

	log.Println("command:", command.Command)

	response := AIResponse{
		Status:  "success",
		Message: "可选提示",
		Data: Data{
			Text: "AI文本内容",
		},
	}

	json.NewEncoder(w).Encode(response)
}

// sendMessagesToDeepSeek 函数发送消息数组到 DeepSeek API 并返回最新的结果字符串
func sendMessagesToDeepSeek(messages []Message, apiKey string) (string, error) {
	// 创建 Resty 客户端
	client := resty.New()

	// 设置请求头
	client.SetHeader("Authorization", "Bearer "+apiKey)
	client.SetHeader("Content-Type", "application/json")

	// 定义请求体
	requestBody := map[string]interface{}{
		"model":    "deepseek-chat",
		"messages": messages,
		"stream":   false,
	}

	// 发送 POST 请求
	resp, err := client.R().
		SetBody(requestBody).
		Post("https://api.deepseek.com/chat/completions")

	if err != nil {
		return "", fmt.Errorf("请求失败: %v", err)
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("请求失败，状态码: %d，响应: %s", resp.StatusCode(), resp.String())
	}

	// 解析响应
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("未收到有效的响应")
	}

	// 返回最新的结果字符串
	return result.Choices[0].Message.Content, nil
}

func Test() {
	// 示例消息数组
	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello!"},
	}

	// 您的 DeepSeek API 密钥
	apiKey := "sk-06a86b52f77142dc92d2d1ddf6861c27"

	// 调用函数获取结果
	result, err := sendMessagesToDeepSeek(messages, apiKey)
	if err != nil {
		log.Fatalf("错误: %v", err)
	}

	// 输出结果
	fmt.Println("最新的结果:", result)
}
