// AI Agent Function Calling — Go 纯标准库实现
// 本示例演示完整的 AI Agent 开发：工具定义 → 工具注册 → LLM 决策 →
// 工具执行 → 结果反馈 → 循环直到完成。
// 实现了 ReAct（Reasoning + Acting）模式的 Agent 执行循环。
//
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：
//   Part A（模拟模式，直接运行）：go run ./agent-example/
//   Part B（真实 API）：
//     export OPENAI_API_KEY="sk-your-key"
//     go run ./agent-example/ real
//   Ollama 本地模型（免费）：
//     ollama serve && ollama pull qwen2.5:7b
//     export OPENAI_BASE_URL="http://localhost:11434/v1"
//     export OPENAI_API_KEY="ollama"
//     export OPENAI_MODEL="qwen2.5:7b"
//     go run ./agent-example/ real

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// ============================================================
// 工具系统定义
// ============================================================

// ToolDefinition 工具定义（用于告知 LLM 可用工具）
type ToolDefinition struct {
	Type     string       `json:"type"` // 固定为 "function"
	Function FunctionDef  `json:"function"`
}

// FunctionDef 函数定义
type FunctionDef struct {
	Name        string                 `json:"name"`        // 函数名
	Description string                 `json:"description"` // 功能描述（LLM 据此决定是否调用）
	Parameters  map[string]interface{} `json:"parameters"`  // 参数 JSON Schema
}

// ToolCall LLM 返回的工具调用请求
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function FunctionCall `json:"function"`
}

// FunctionCall 函数调用详情
type FunctionCall struct {
	Name      string `json:"name"`      // 要调用的函数名
	Arguments string `json:"arguments"` // 参数 JSON 字符串
}

// Tool 工具实现（注册到 Agent 的可执行工具）
type Tool struct {
	Definition ToolDefinition
	Execute    func(args map[string]interface{}) (string, error) // 执行函数
}

// ============================================================
// 工具注册表
// ============================================================

// ToolRegistry 工具注册表，管理所有可用工具
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry 创建工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register 注册一个工具
func (r *ToolRegistry) Register(tool Tool) {
	r.tools[tool.Definition.Function.Name] = tool
}

// Get 获取工具
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// GetDefinitions 获取所有工具定义（传给 LLM）
func (r *ToolRegistry) GetDefinitions() []ToolDefinition {
	defs := make([]ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		defs = append(defs, tool.Definition)
	}
	return defs
}

// Execute 执行指定工具
func (r *ToolRegistry) Execute(name string, argsJSON string) (string, error) {
	tool, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("未知工具: %s", name)
	}

	// 解析参数 JSON
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("解析工具参数失败: %w", err)
	}

	return tool.Execute(args)
}

// ============================================================
// 工具实现：计算器
// ============================================================

func newCalculatorTool() Tool {
	return Tool{
		Definition: ToolDefinition{
			Type: "function",
			Function: FunctionDef{
				Name:        "calculate",
				Description: "执行数学计算，支持加减乘除和常见数学函数",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"expression": map[string]interface{}{
							"type":        "string",
							"description": "数学表达式，如 '2 + 3'、'sqrt(16)'、'15 * 3.14'",
						},
					},
					"required": []string{"expression"},
				},
			},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			expr, ok := args["expression"].(string)
			if !ok {
				return "", fmt.Errorf("缺少 expression 参数")
			}
			result := evaluateExpression(expr)
			return fmt.Sprintf("计算结果: %s = %s", expr, result), nil
		},
	}
}

// evaluateExpression 简单的数学表达式求值器
func evaluateExpression(expr string) string {
	expr = strings.TrimSpace(expr)

	// 处理 sqrt 函数
	if strings.HasPrefix(expr, "sqrt(") && strings.HasSuffix(expr, ")") {
		inner := expr[5 : len(expr)-1]
		val, err := strconv.ParseFloat(inner, 64)
		if err != nil {
			return "表达式解析错误"
		}
		return fmt.Sprintf("%.4f", math.Sqrt(val))
	}

	// 处理 pow 函数
	if strings.HasPrefix(expr, "pow(") && strings.HasSuffix(expr, ")") {
		inner := expr[4 : len(expr)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 2 {
			base, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			exp, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err1 == nil && err2 == nil {
				return fmt.Sprintf("%.4f", math.Pow(base, exp))
			}
		}
	}

	// 处理基本四则运算
	operators := []string{" + ", " - ", " * ", " / "}
	for _, op := range operators {
		if idx := strings.LastIndex(expr, op); idx > 0 {
			leftStr := strings.TrimSpace(expr[:idx])
			rightStr := strings.TrimSpace(expr[idx+len(op):])
			left, err1 := strconv.ParseFloat(leftStr, 64)
			right, err2 := strconv.ParseFloat(rightStr, 64)
			if err1 != nil || err2 != nil {
				continue
			}
			var result float64
			switch strings.TrimSpace(op) {
			case "+":
				result = left + right
			case "-":
				result = left - right
			case "*":
				result = left * right
			case "/":
				if right == 0 {
					return "错误: 除以零"
				}
				result = left / right
			}
			// 如果结果是整数，不显示小数点
			if result == math.Trunc(result) {
				return fmt.Sprintf("%.0f", result)
			}
			return fmt.Sprintf("%.4f", result)
		}
	}

	return "无法解析表达式: " + expr
}

// ============================================================
// 工具实现：天气查询
// ============================================================

func newWeatherTool() Tool {
	// 模拟天气数据
	weatherData := map[string]map[string]interface{}{
		"北京": {"temp": 25, "weather": "晴", "humidity": 40, "wind": "北风3级"},
		"上海": {"temp": 28, "weather": "多云", "humidity": 65, "wind": "东南风2级"},
		"深圳": {"temp": 32, "weather": "阵雨", "humidity": 80, "wind": "南风4级"},
		"杭州": {"temp": 27, "weather": "晴转多云", "humidity": 55, "wind": "西风2级"},
		"成都": {"temp": 22, "weather": "阴", "humidity": 70, "wind": "微风"},
	}

	return Tool{
		Definition: ToolDefinition{
			Type: "function",
			Function: FunctionDef{
				Name:        "get_weather",
				Description: "查询指定城市的当前天气信息，包括温度、天气状况、湿度和风力",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"city": map[string]interface{}{
							"type":        "string",
							"description": "城市名称，如 '北京'、'上海'",
						},
					},
					"required": []string{"city"},
				},
			},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			city, ok := args["city"].(string)
			if !ok {
				return "", fmt.Errorf("缺少 city 参数")
			}

			data, exists := weatherData[city]
			if !exists {
				return fmt.Sprintf("未找到城市 '%s' 的天气数据，支持的城市：北京、上海、深圳、杭州、成都", city), nil
			}

			return fmt.Sprintf("%s天气：%s，温度 %d°C，湿度 %d%%，%s",
				city, data["weather"], data["temp"], data["humidity"], data["wind"]), nil
		},
	}
}

// ============================================================
// 工具实现：知识搜索
// ============================================================

func newSearchTool() Tool {
	// 模拟搜索结果
	searchDB := map[string]string{
		"goroutine": "Goroutine 是 Go 的轻量级线程，初始栈 2KB，由 Go 运行时调度。可创建数十万个并发执行。",
		"channel":   "Channel 是 Go 中 goroutine 间通信的管道，遵循 CSP 模型。分为无缓冲和有缓冲两种。",
		"docker":    "Docker 是容器化平台，Go 语言编写。支持多阶段构建，Go 应用可编译为 scratch 基础镜像。",
		"kubernetes": "Kubernetes 是容器编排平台，Go 语言编写。核心概念：Pod、Deployment、Service、ConfigMap。",
		"gin":       "Gin 是 Go 最流行的 Web 框架（使用率 48%），高性能，支持中间件、路由分组、参数验证。",
		"gorm":      "GORM 是 Go 最流行的 ORM 库，支持自动迁移、关联关系、Hook、软删除、Scope 查询。",
	}

	return Tool{
		Definition: ToolDefinition{
			Type: "function",
			Function: FunctionDef{
				Name:        "search_knowledge",
				Description: "搜索 Go 语言知识库，查找技术概念、框架、工具的相关信息",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "搜索关键词，如 'goroutine'、'gin 框架'",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		Execute: func(args map[string]interface{}) (string, error) {
			query, ok := args["query"].(string)
			if !ok {
				return "", fmt.Errorf("缺少 query 参数")
			}

			queryLower := strings.ToLower(query)
			var results []string
			for key, value := range searchDB {
				if strings.Contains(queryLower, key) || strings.Contains(key, queryLower) {
					results = append(results, value)
				}
			}

			if len(results) == 0 {
				return fmt.Sprintf("未找到与 '%s' 相关的知识，请尝试其他关键词", query), nil
			}

			return strings.Join(results, "\n"), nil
		},
	}
}

// ============================================================
// Agent 核心实现
// ============================================================

// AgentMessage Agent 对话消息（扩展支持工具调用）
type AgentMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`  // assistant 消息中的工具调用
	ToolCallID string     `json:"tool_call_id,omitempty"` // tool 消息中的调用 ID
}

// Agent AI Agent 核心结构
type Agent struct {
	Registry     *ToolRegistry
	Messages     []AgentMessage
	MaxIter      int    // 最大迭代次数（防止无限循环）
	BaseURL      string // LLM API 地址
	APIKey       string // API 密钥
	Model        string // 模型名称
}

// NewAgent 创建新的 Agent
func NewAgent(registry *ToolRegistry, baseURL, apiKey, model string) *Agent {
	return &Agent{
		Registry: registry,
		Messages: make([]AgentMessage, 0),
		MaxIter:  5, // 默认最多 5 轮工具调用
		BaseURL:  baseURL,
		APIKey:   apiKey,
		Model:    model,
	}
}

// ============================================================
// LLM API 调用结构体
// ============================================================

// LLMRequest LLM 请求体
type LLMRequest struct {
	Model       string           `json:"model"`
	Messages    []AgentMessage   `json:"messages"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
}

// LLMResponse LLM 响应体
type LLMResponse struct {
	Choices []struct {
		Message      AgentMessage `json:"message"`
		FinishReason string       `json:"finish_reason"`
	} `json:"choices"`
}

// callLLMWithTools 调用 LLM API（带工具定义）
func (a *Agent) callLLMWithTools() (*AgentMessage, error) {
	reqBody := LLMRequest{
		Model:       a.Model,
		Messages:    a.Messages,
		Tools:       a.Registry.GetDefinitions(),
		Temperature: 0.3,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := strings.TrimRight(a.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 错误 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var llmResp LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("API 返回空响应")
	}

	return &llmResp.Choices[0].Message, nil
}

// ============================================================
// Part A：模拟 Agent 执行（无需 API Key）
// ============================================================

// SimulatedStep 模拟的 Agent 执行步骤
type SimulatedStep struct {
	StepType string // "thought" / "action" / "observation" / "answer"
	Content  string
	ToolName string
	ToolArgs string
}

func partA() {
	fmt.Println("========================================")
	fmt.Println("Part A：AI Agent 模拟演示")
	fmt.Println("========================================")
	fmt.Println()

	// 创建工具注册表
	registry := NewToolRegistry()
	registry.Register(newCalculatorTool())
	registry.Register(newWeatherTool())
	registry.Register(newSearchTool())

	// 演示 1：展示工具定义
	fmt.Println("🔧 1. 已注册的工具：")
	for _, def := range registry.GetDefinitions() {
		fmt.Printf("   - %s: %s\n", def.Function.Name, def.Function.Description)
	}
	fmt.Println()

	// 演示 2：直接调用工具
	fmt.Println("⚡ 2. 工具直接调用测试：")

	// 计算器
	result, _ := registry.Execute("calculate", `{"expression": "15 * 3.14"}`)
	fmt.Printf("   calculate(\"15 * 3.14\") → %s\n", result)

	result, _ = registry.Execute("calculate", `{"expression": "sqrt(144)"}`)
	fmt.Printf("   calculate(\"sqrt(144)\") → %s\n", result)

	// 天气
	result, _ = registry.Execute("get_weather", `{"city": "北京"}`)
	fmt.Printf("   get_weather(\"北京\") → %s\n", result)

	// 搜索
	result, _ = registry.Execute("search_knowledge", `{"query": "goroutine"}`)
	fmt.Printf("   search_knowledge(\"goroutine\") → %s\n", result)
	fmt.Println()

	// 演示 3：模拟 Agent 执行循环（ReAct 模式）
	fmt.Println("🤖 3. 模拟 Agent 执行循环（ReAct 模式）")
	fmt.Println()

	// 场景 1：需要计算的问题
	fmt.Println("--- 场景 1：数学计算 ---")
	fmt.Println("用户: 一个圆的半径是 7cm，它的面积是多少？")
	fmt.Println()

	steps1 := []SimulatedStep{
		{StepType: "thought", Content: "用户问圆的面积，需要用公式 π * r²。半径 r=7，我需要计算 3.14159 * 7 * 7。"},
		{StepType: "action", Content: "调用计算器", ToolName: "calculate", ToolArgs: `{"expression": "3.14159 * 49"}`},
		{StepType: "observation", Content: ""},
		{StepType: "answer", Content: "圆的面积 = π × r² = 3.14159 × 7² = 3.14159 × 49 ≈ 153.94 平方厘米。"},
	}

	executeSimulatedSteps(registry, steps1)
	fmt.Println()

	// 场景 2：需要多个工具的问题
	fmt.Println("--- 场景 2：多工具协作 ---")
	fmt.Println("用户: 北京今天天气怎么样？适合户外运动吗？")
	fmt.Println()

	steps2 := []SimulatedStep{
		{StepType: "thought", Content: "用户问北京天气和是否适合户外运动，我需要先查询天气信息。"},
		{StepType: "action", Content: "查询天气", ToolName: "get_weather", ToolArgs: `{"city": "北京"}`},
		{StepType: "observation", Content: ""},
		{StepType: "answer", Content: "北京今天天气晴朗，气温 25°C，湿度 40%，北风3级。这样的天气非常适合户外运动，温度适宜，湿度低，风力不大。建议做好防晒措施。"},
	}

	executeSimulatedSteps(registry, steps2)
	fmt.Println()

	// 场景 3：需要搜索 + 计算的复合问题
	fmt.Println("--- 场景 3：搜索 + 推理 ---")
	fmt.Println("用户: goroutine 的初始栈大小是多少？如果创建 10 万个 goroutine，最少需要多少内存？")
	fmt.Println()

	steps3 := []SimulatedStep{
		{StepType: "thought", Content: "需要先查找 goroutine 的初始栈大小，然后计算 10 万个的总内存。"},
		{StepType: "action", Content: "搜索知识库", ToolName: "search_knowledge", ToolArgs: `{"query": "goroutine"}`},
		{StepType: "observation", Content: ""},
		{StepType: "thought", Content: "goroutine 初始栈 2KB，10 万个需要 2KB * 100000 = 200000KB。需要转换单位。"},
		{StepType: "action", Content: "计算内存", ToolName: "calculate", ToolArgs: `{"expression": "2 * 100000"}`},
		{StepType: "observation", Content: ""},
		{StepType: "answer", Content: "Goroutine 的初始栈大小为 2KB。创建 10 万个 goroutine 最少需要 2KB × 100,000 = 200,000KB ≈ 195MB 的栈内存。实际内存使用会更多，因为还包括 goroutine 的调度元数据和运行时开销。"},
	}

	executeSimulatedSteps(registry, steps3)
	fmt.Println()

	// 演示 4：工具定义的 JSON Schema
	fmt.Println("📋 4. 工具定义 JSON Schema（传给 LLM）：")
	toolDefs := registry.GetDefinitions()
	toolJSON, _ := json.MarshalIndent(toolDefs, "   ", "  ")
	fmt.Println("   " + string(toolJSON))
	fmt.Println()

	fmt.Println("✅ Part A 完成！Agent 模拟演示结束。")
	fmt.Println("💡 要连接真实 LLM API，请运行: go run ./agent-example/ real")
}

// executeSimulatedSteps 执行模拟的 Agent 步骤
func executeSimulatedSteps(registry *ToolRegistry, steps []SimulatedStep) {
	iteration := 0
	for _, step := range steps {
		switch step.StepType {
		case "thought":
			iteration++
			fmt.Printf("  💭 Thought #%d: %s\n", iteration, step.Content)
		case "action":
			fmt.Printf("  🔨 Action: %s(%s)\n", step.ToolName, step.ToolArgs)
			// 真正执行工具
			result, err := registry.Execute(step.ToolName, step.ToolArgs)
			if err != nil {
				fmt.Printf("  ❌ Error: %v\n", err)
			} else {
				fmt.Printf("  👁️ Observation: %s\n", result)
			}
		case "observation":
			// 已在 action 中处理
		case "answer":
			fmt.Printf("  ✅ Answer: %s\n", step.Content)
		}
	}
}

// ============================================================
// Part B：连接真实 LLM API
// ============================================================

func partB() {
	fmt.Println("========================================")
	fmt.Println("Part B：AI Agent + 真实 LLM API")
	fmt.Println("========================================")
	fmt.Println()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ 未设置 OPENAI_API_KEY 环境变量")
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

	// 创建工具注册表
	registry := NewToolRegistry()
	registry.Register(newCalculatorTool())
	registry.Register(newWeatherTool())
	registry.Register(newSearchTool())

	// 创建 Agent
	agent := NewAgent(registry, baseURL, apiKey, model)

	// 设置系统提示
	agent.Messages = append(agent.Messages, AgentMessage{
		Role:    "system",
		Content: "你是一个智能助手，可以使用工具来帮助用户。当需要计算、查询天气或搜索知识时，请调用相应的工具。使用中文回答。",
	})

	// 用户问题
	userQuery := "北京今天天气怎么样？如果气温超过 30 度就不适合跑步了，帮我判断一下。"
	fmt.Printf("用户: %s\n\n", userQuery)

	agent.Messages = append(agent.Messages, AgentMessage{
		Role:    "user",
		Content: userQuery,
	})

	// Agent 执行循环
	for i := 0; i < agent.MaxIter; i++ {
		fmt.Printf("--- 迭代 %d ---\n", i+1)

		// 调用 LLM
		response, err := agent.callLLMWithTools()
		if err != nil {
			fmt.Printf("❌ LLM 调用失败: %v\n", err)
			return
		}

		// 检查是否有工具调用
		if len(response.ToolCalls) > 0 {
			// 将 assistant 消息加入历史
			agent.Messages = append(agent.Messages, *response)

			for _, tc := range response.ToolCalls {
				fmt.Printf("🔨 调用工具: %s(%s)\n", tc.Function.Name, tc.Function.Arguments)

				// 执行工具
				result, err := registry.Execute(tc.Function.Name, tc.Function.Arguments)
				if err != nil {
					result = fmt.Sprintf("工具执行错误: %v", err)
				}
				fmt.Printf("👁️ 结果: %s\n", result)

				// 将工具结果加入对话历史
				agent.Messages = append(agent.Messages, AgentMessage{
					Role:       "tool",
					Content:    result,
					ToolCallID: tc.ID,
				})
			}
			fmt.Println()
			continue
		}

		// LLM 直接回答，结束循环
		fmt.Printf("\n✅ Agent 回答: %s\n", response.Content)
		break
	}

	fmt.Println()
	fmt.Println("✅ Part B 完成！")
}

// ============================================================
// 主函数
// ============================================================

func main() {
	fmt.Println("🤖 AI Agent Function Calling 示例")
	fmt.Println("ReAct 模式：Observe → Think → Act → Observe")
	fmt.Println()

	// Part A：模拟模式
	partA()

	// Part B：连接真实 API
	if len(os.Args) > 1 && os.Args[1] == "real" {
		fmt.Println()
		partB()
	}
}
