// OpenAI Chat API 调用 — Go 纯标准库实现
// 本示例演示如何使用 net/http 调用 OpenAI Chat Completions API，
// 包含普通模式和流式响应（SSE）两种方式。
// 支持配置 API 端点，可切换到 Ollama 本地模型（免费替代方案）。
//
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：
//   Part A（模拟模式，直接运行）：go run ./openai-chat/
//   Part B（真实 API）：
//     export OPENAI_API_KEY="sk-your-key"
//     go run ./openai-chat/ real
//   Ollama 本地模型（免费）：
//     ollama serve && ollama pull qwen2.5:7b
//     export OPENAI_BASE_URL="http://localhost:11434/v1"
//     export OPENAI_API_KEY="ollama"
//     export OPENAI_MODEL="qwen2.5:7b"
//     go run ./openai-chat/ real

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ============================================================
// OpenAI API 请求/响应结构体（匹配 OpenAI API 规范）
// ============================================================

// ChatRequest 表示 Chat Completions API 的请求体
type ChatRequest struct {
	Model       string    `json:"model"`                 // 模型名称
	Messages    []Message `json:"messages"`              // 对话消息列表
	Temperature float64   `json:"temperature,omitempty"` // 随机性控制，0~2
	MaxTokens   int       `json:"max_tokens,omitempty"`  // 最大生成 token 数
	Stream      bool      `json:"stream,omitempty"`      // 是否启用流式响应
}

// Message 表示对话中的一条消息
type Message struct {
	Role    string `json:"role"`    // 角色：system / user / assistant
	Content string `json:"content"` // 消息内容
}

// ChatResponse 表示 Chat Completions API 的完整响应
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 表示一个生成选项
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage 表示 token 使用量
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk 表示流式响应中的一个数据块
type StreamChunk struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice 表示流式响应中的一个选项
type StreamChoice struct {
	Index        int          `json:"index"`
	Delta        MessageDelta `json:"delta"`
	FinishReason *string      `json:"finish_reason"`
}

// MessageDelta 表示流式响应中的增量消息
type MessageDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// ============================================================
// OpenAI 客户端
// ============================================================

// Client 是 OpenAI API 客户端
type Client struct {
	BaseURL    string       // API 基础 URL
	APIKey     string       // API 密钥
	Model      string       // 默认模型
	HTTPClient *http.Client // HTTP 客户端
}

// NewClient 创建一个新的 OpenAI 客户端
func NewClient(baseURL, apiKey, model string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second, // 设置超时，避免长时间阻塞
		},
	}
}

// ChatCompletion 发送普通（非流式）Chat Completions 请求
func (c *Client) ChatCompletion(messages []Message) (*ChatResponse, error) {
	reqBody := ChatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1000,
		Stream:      false,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := c.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回错误 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &chatResp, nil
}

// ChatCompletionStream 发送流式 Chat Completions 请求，通过回调逐 token 输出
func (c *Client) ChatCompletionStream(messages []Message, onToken func(token string)) error {
	reqBody := ChatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1000,
		Stream:      true, // 启用流式响应
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	url := c.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	// 流式请求不设置超时（或设置更长的超时）
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API 返回错误 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// 使用 bufio.Scanner 逐行读取 SSE 响应
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// SSE 协议：数据行以 "data: " 前缀开头
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		payload := strings.TrimPrefix(line, "data: ")

		// "[DONE]" 表示流式响应结束
		if payload == "[DONE]" {
			break
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue // 跳过解析失败的行
		}

		// 提取增量内容并通过回调输出
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			onToken(chunk.Choices[0].Delta.Content)
		}
	}

	return scanner.Err()
}

// ============================================================
// Part A：模拟模式（无需 API Key，直接运行理解原理）
// ============================================================

func partA() {
	fmt.Println("========================================")
	fmt.Println("Part A：模拟模式（无需 API Key）")
	fmt.Println("========================================")
	fmt.Println()

	// 演示 1：构建请求结构体
	fmt.Println("--- 1. 构建 Chat Completions 请求 ---")
	request := ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []Message{
			{Role: "system", Content: "你是一个 Go 语言专家，回答简洁准确。"},
			{Role: "user", Content: "什么是 goroutine？"},
		},
		Temperature: 0.7,
		MaxTokens:   500,
		Stream:      false,
	}
	reqJSON, _ := json.MarshalIndent(request, "", "  ")
	fmt.Println("请求体：")
	fmt.Println(string(reqJSON))
	fmt.Println()

	// 演示 2：模拟普通响应
	fmt.Println("--- 2. 模拟普通模式响应 ---")
	mockResponse := ChatResponse{
		ID:      "chatcmpl-mock-001",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gpt-4o-mini",
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: "Goroutine 是 Go 语言中的轻量级线程，由 Go 运行时管理。它比操作系统线程更轻量（初始栈仅 2KB），可以轻松创建数十万个并发执行单元。使用 `go` 关键字启动：`go func() { ... }()`。Goroutine 通过 channel 进行通信，遵循 CSP（Communicating Sequential Processes）并发模型。",
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     25,
			CompletionTokens: 80,
			TotalTokens:      105,
		},
	}
	fmt.Printf("模型: %s\n", mockResponse.Model)
	fmt.Printf("回答: %s\n", mockResponse.Choices[0].Message.Content)
	fmt.Printf("Token 用量: 提示 %d + 生成 %d = 总计 %d\n",
		mockResponse.Usage.PromptTokens,
		mockResponse.Usage.CompletionTokens,
		mockResponse.Usage.TotalTokens)
	fmt.Println()

	// 演示 3：模拟流式响应（SSE）
	fmt.Println("--- 3. 模拟流式响应（SSE 协议） ---")
	streamTokens := []string{
		"Goroutine", " 是", " Go", " 语言", "中的",
		"轻量级", "线程", "，", "由", " Go",
		" 运行时", "管理", "。",
	}

	fmt.Print("流式输出: ")
	for _, token := range streamTokens {
		fmt.Print(token)
		time.Sleep(50 * time.Millisecond) // 模拟逐 token 输出的延迟
	}
	fmt.Println()
	fmt.Println()

	// 演示 4：SSE 数据格式解析
	fmt.Println("--- 4. SSE 数据格式示例 ---")
	sseLines := []string{
		`data: {"id":"chatcmpl-001","choices":[{"delta":{"role":"assistant"},"index":0}]}`,
		`data: {"id":"chatcmpl-001","choices":[{"delta":{"content":"Goroutine"},"index":0}]}`,
		`data: {"id":"chatcmpl-001","choices":[{"delta":{"content":" 是"},"index":0}]}`,
		`data: {"id":"chatcmpl-001","choices":[{"delta":{"content":"轻量级线程"},"index":0}]}`,
		`data: [DONE]`,
	}

	fmt.Println("原始 SSE 数据流：")
	for _, line := range sseLines {
		fmt.Println(line)
	}
	fmt.Println()

	// 解析 SSE 数据
	fmt.Print("解析后的内容: ")
	for _, line := range sseLines {
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			fmt.Print(" [结束]")
			break
		}
		var chunk StreamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err == nil {
			if len(chunk.Choices) > 0 {
				fmt.Print(chunk.Choices[0].Delta.Content)
			}
		}
	}
	fmt.Println()
	fmt.Println()

	// 演示 5：多轮对话
	fmt.Println("--- 5. 多轮对话消息管理 ---")
	conversation := []Message{
		{Role: "system", Content: "你是一个 Go 语言专家"},
		{Role: "user", Content: "什么是 channel？"},
		{Role: "assistant", Content: "Channel 是 Go 中 goroutine 之间通信的管道。"},
		{Role: "user", Content: "有缓冲和无缓冲有什么区别？"},
	}
	fmt.Println("对话历史：")
	for _, msg := range conversation {
		fmt.Printf("  [%s] %s\n", msg.Role, msg.Content)
	}
	fmt.Println()

	// 演示 6：API 端点配置
	fmt.Println("--- 6. API 端点配置（支持 Ollama） ---")
	configs := []struct {
		name    string
		baseURL string
		model   string
	}{
		{"OpenAI", "https://api.openai.com/v1", "gpt-4o-mini"},
		{"Ollama（本地免费）", "http://localhost:11434/v1", "qwen2.5:7b"},
		{"Azure OpenAI", "https://your-resource.openai.azure.com/openai/deployments/your-model/v1", "gpt-4o-mini"},
	}
	for _, cfg := range configs {
		fmt.Printf("  %s: %s (模型: %s)\n", cfg.name, cfg.baseURL, cfg.model)
	}
	fmt.Println()

	fmt.Println("✅ Part A 完成！所有演示均为模拟数据，无需 API Key。")
	fmt.Println("💡 要连接真实 API，请运行: go run ./openai-chat/ real")
}

// ============================================================
// Part B：连接真实 API（需要 API Key 或 Ollama）
// ============================================================

func partB() {
	fmt.Println("========================================")
	fmt.Println("Part B：连接真实 API")
	fmt.Println("========================================")
	fmt.Println()

	// 从环境变量读取配置
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ 未设置 OPENAI_API_KEY 环境变量")
		fmt.Println("使用 OpenAI：export OPENAI_API_KEY=\"sk-your-key\"")
		fmt.Println("使用 Ollama：export OPENAI_API_KEY=\"ollama\"")
		return
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}

	fmt.Printf("API 端点: %s\n", baseURL)
	fmt.Printf("模型: %s\n", model)
	fmt.Println()

	client := NewClient(baseURL, apiKey, model)

	// 1. 普通模式调用
	fmt.Println("--- 1. 普通模式调用 ---")
	messages := []Message{
		{Role: "system", Content: "你是一个 Go 语言专家，回答简洁准确，使用中文。"},
		{Role: "user", Content: "用一句话解释什么是 goroutine。"},
	}

	resp, err := client.ChatCompletion(messages)
	if err != nil {
		fmt.Printf("❌ 普通模式调用失败: %v\n", err)
	} else {
		fmt.Printf("回答: %s\n", resp.Choices[0].Message.Content)
		fmt.Printf("Token: %d\n", resp.Usage.TotalTokens)
	}
	fmt.Println()

	// 2. 流式模式调用
	fmt.Println("--- 2. 流式模式调用 ---")
	streamMessages := []Message{
		{Role: "system", Content: "你是一个 Go 语言专家，回答简洁准确，使用中文。"},
		{Role: "user", Content: "简要介绍 Go 的 channel 机制。"},
	}

	fmt.Print("流式回答: ")
	err = client.ChatCompletionStream(streamMessages, func(token string) {
		fmt.Print(token) // 逐 token 实时输出
	})
	if err != nil {
		fmt.Printf("\n❌ 流式调用失败: %v\n", err)
	}
	fmt.Println()
	fmt.Println()

	fmt.Println("✅ Part B 完成！")
}

// ============================================================
// 主函数
// ============================================================

func main() {
	fmt.Println("🤖 OpenAI Chat API 调用示例")
	fmt.Println("Go 纯标准库实现（net/http + encoding/json）")
	fmt.Println()

	// Part A：模拟模式，直接运行
	partA()

	// Part B：连接真实 API，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		fmt.Println()
		partB()
	}
}
