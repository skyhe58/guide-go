// RAG 检索增强生成 — Go 纯标准库实现
// 本示例演示完整的 RAG 流程：文档加载 → 文本分块 → 向量化（Embedding）→
// 内存向量存储 → 余弦相似度检索 → Prompt 构建 → LLM 生成回答。
// 使用纯 Go 实现，无需第三方向量数据库依赖。
//
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：
//   Part A（模拟模式，直接运行）：go run ./rag-example/
//   Part B（真实 API）：
//     export OPENAI_API_KEY="sk-your-key"
//     go run ./rag-example/ real
//   Ollama 本地模型（免费）：
//     ollama serve && ollama pull qwen2.5:7b
//     export OPENAI_BASE_URL="http://localhost:11434/v1"
//     export OPENAI_API_KEY="ollama"
//     export OPENAI_MODEL="qwen2.5:7b"
//     go run ./rag-example/ real

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// ============================================================
// 文档与向量存储数据结构
// ============================================================

// Document 表示一个文档片段
type Document struct {
	Content   string            // 文本内容
	Embedding []float64         // 向量表示
	Metadata  map[string]string // 元数据（来源、页码等）
}

// VectorStore 内存向量存储
type VectorStore struct {
	Documents []Document
}

// SearchResult 检索结果
type SearchResult struct {
	Document Document
	Score    float64 // 余弦相似度分数
}

// ============================================================
// 核心算法：余弦相似度
// ============================================================

// cosineSimilarity 计算两个向量的余弦相似度
// 公式：cos(θ) = (A·B) / (|A| × |B|)
// 值域：[-1, 1]，1 表示完全相同，0 表示正交，-1 表示完全相反
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i] // 点积
		normA += a[i] * a[i]      // A 的模的平方
		normB += b[i] * b[i]      // B 的模的平方
	}

	// 避免除以零
	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// ============================================================
// 文档加载器
// ============================================================

// loadDocuments 从字符串列表加载文档
func loadDocuments(texts []string, source string) []Document {
	docs := make([]Document, 0, len(texts))
	for i, text := range texts {
		docs = append(docs, Document{
			Content: text,
			Metadata: map[string]string{
				"source": source,
				"index":  fmt.Sprintf("%d", i),
			},
		})
	}
	return docs
}

// ============================================================
// 文本分块器
// ============================================================

// chunkText 将长文本按段落分块，支持最大块大小和重叠
// chunkSize: 每个块的最大字符数
// overlap: 相邻块之间的重叠字符数
func chunkText(text string, chunkSize, overlap int) []string {
	// 先按段落分割
	paragraphs := strings.Split(text, "\n\n")

	var chunks []string
	var currentChunk strings.Builder

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// 如果当前块加上新段落超过限制，保存当前块并开始新块
		if currentChunk.Len()+len(para) > chunkSize && currentChunk.Len() > 0 {
			chunk := currentChunk.String()
			chunks = append(chunks, chunk)

			// 重叠处理：保留上一个块的末尾部分
			currentChunk.Reset()
			if overlap > 0 && len(chunk) > overlap {
				currentChunk.WriteString(chunk[len(chunk)-overlap:])
				currentChunk.WriteString(" ")
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(para)
	}

	// 添加最后一个块
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

// ============================================================
// 模拟 Embedding（用于 Part A 演示）
// ============================================================

// simpleEmbedding 基于关键词频率的简单向量化
// 这是一个教学用的模拟实现，真实场景应使用 Embedding API
func simpleEmbedding(text string, vocabulary []string) []float64 {
	textLower := strings.ToLower(text)
	embedding := make([]float64, len(vocabulary))

	for i, word := range vocabulary {
		// 统计关键词出现次数，归一化为 [0, 1]
		count := float64(strings.Count(textLower, strings.ToLower(word)))
		// 使用 TF 近似：出现次数 / 文本长度
		if len(text) > 0 {
			embedding[i] = count / float64(len(text)) * 100
		}
	}

	// L2 归一化，使向量模长为 1
	var norm float64
	for _, v := range embedding {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range embedding {
			embedding[i] /= norm
		}
	}

	return embedding
}

// ============================================================
// 向量存储操作
// ============================================================

// NewVectorStore 创建新的向量存储
func NewVectorStore() *VectorStore {
	return &VectorStore{
		Documents: make([]Document, 0),
	}
}

// AddDocuments 批量添加文档到向量存储
func (vs *VectorStore) AddDocuments(docs []Document) {
	vs.Documents = append(vs.Documents, docs...)
}

// Search 检索与查询向量最相似的 Top-K 文档
func (vs *VectorStore) Search(queryEmbedding []float64, topK int) []SearchResult {
	results := make([]SearchResult, 0, len(vs.Documents))

	for _, doc := range vs.Documents {
		score := cosineSimilarity(queryEmbedding, doc.Embedding)
		results = append(results, SearchResult{
			Document: doc,
			Score:    score,
		})
	}

	// 按相似度降序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 返回 Top-K 结果
	if topK > len(results) {
		topK = len(results)
	}
	return results[:topK]
}

// ============================================================
// RAG Prompt 构建
// ============================================================

// buildRAGPrompt 构建 RAG 增强的 Prompt
func buildRAGPrompt(query string, retrievedDocs []SearchResult) string {
	var context strings.Builder
	context.WriteString("请根据以下参考文档回答用户的问题。如果参考文档中没有相关信息，请说明。\n\n")
	context.WriteString("## 参考文档\n\n")

	for i, result := range retrievedDocs {
		context.WriteString(fmt.Sprintf("### 文档 %d（相似度: %.4f）\n", i+1, result.Score))
		context.WriteString(result.Document.Content)
		context.WriteString("\n\n")
	}

	context.WriteString("## 用户问题\n\n")
	context.WriteString(query)
	context.WriteString("\n\n")
	context.WriteString("## 回答要求\n")
	context.WriteString("- 基于参考文档内容回答\n")
	context.WriteString("- 如果文档中没有相关信息，明确说明\n")
	context.WriteString("- 回答简洁准确，使用中文\n")

	return context.String()
}

// ============================================================
// LLM 调用（简化版，复用 OpenAI API 结构）
// ============================================================

// ChatMessage 对话消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest LLM 请求
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// ChatResponse LLM 响应
type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

// callLLM 调用 LLM API 生成回答
func callLLM(baseURL, apiKey, model, prompt string) (string, error) {
	reqBody := ChatRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "system", Content: "你是一个知识库助手，基于提供的参考文档回答问题。"},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3, // RAG 场景使用较低温度，减少幻觉
		MaxTokens:   500,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	url := strings.TrimRight(baseURL, "/") + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API 错误 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("API 返回空响应")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// ============================================================
// 知识库数据（Go 语言知识）
// ============================================================

// 模拟知识库文档
var knowledgeBase = `Go 语言的并发模型

Goroutine 是 Go 语言中的轻量级线程，由 Go 运行时调度器管理，而非操作系统。每个 goroutine 的初始栈大小仅为 2KB（可动态增长到 1GB），远小于操作系统线程的默认栈大小（通常 1-8MB）。这使得一个 Go 程序可以轻松创建数十万甚至上百万个 goroutine。

Channel 是 Go 语言中 goroutine 之间通信的管道机制。Channel 遵循 CSP（Communicating Sequential Processes）并发模型，通过通信来共享内存，而不是通过共享内存来通信。Channel 分为无缓冲 channel 和有缓冲 channel。无缓冲 channel 要求发送和接收同时就绪，有缓冲 channel 允许在缓冲区未满时异步发送。

Go 的 GMP 调度模型由三个核心组件组成：G（Goroutine）、M（Machine，操作系统线程）、P（Processor，逻辑处理器）。P 的数量默认等于 CPU 核心数（GOMAXPROCS），每个 P 维护一个本地运行队列。调度器通过工作窃取（work stealing）算法实现负载均衡。

Go 的垃圾回收器使用三色标记清除算法，配合写屏障（write barrier）实现并发 GC。GC 触发条件包括：堆内存增长到 GOGC 设定的比例（默认 100%，即翻倍时触发）、手动调用 runtime.GC()、超过 2 分钟未触发 GC。Go 1.19 引入了 GOMEMLIMIT 参数，允许设置内存上限。

sync 包提供了多种同步原语：Mutex（互斥锁）、RWMutex（读写锁）、WaitGroup（等待组）、Once（单次执行）、Pool（对象池）、Map（并发安全 Map）。其中 sync.Pool 常用于减少 GC 压力，sync.Once 常用于单例模式。

Context 包用于在 goroutine 之间传递取消信号、超时控制和请求范围的值。常用函数包括 context.WithCancel、context.WithTimeout、context.WithDeadline、context.WithValue。Context 应该作为函数的第一个参数传递，不应存储在结构体中。

Go 的错误处理采用显式返回 error 的方式，而非异常机制。Go 1.13 引入了 errors.Is 和 errors.As 用于错误链判断，fmt.Errorf 的 %w 动词用于包装错误。panic 和 recover 仅用于不可恢复的错误场景。`

// ============================================================
// Part A：模拟模式
// ============================================================

func partA() {
	fmt.Println("========================================")
	fmt.Println("Part A：RAG 完整流程演示（模拟模式）")
	fmt.Println("========================================")
	fmt.Println()

	// 词汇表（用于模拟 Embedding）
	vocabulary := []string{
		"goroutine", "channel", "并发", "线程", "调度",
		"gmp", "gc", "垃圾回收", "内存", "堆",
		"sync", "mutex", "锁", "waitgroup", "pool",
		"context", "取消", "超时", "传递", "信号",
		"error", "错误", "panic", "recover", "处理",
		"csp", "通信", "缓冲", "队列", "运行时",
	}

	// 步骤 1：文档加载
	fmt.Println("📄 步骤 1：文档加载")
	fmt.Printf("   原始文档长度: %d 字符\n", len(knowledgeBase))
	fmt.Println()

	// 步骤 2：文本分块
	fmt.Println("✂️  步骤 2：文本分块（chunkSize=300, overlap=50）")
	chunks := chunkText(knowledgeBase, 300, 50)
	fmt.Printf("   分块数量: %d\n", len(chunks))
	for i, chunk := range chunks {
		preview := chunk
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		fmt.Printf("   块 %d (%d 字符): %s\n", i+1, len(chunk), preview)
	}
	fmt.Println()

	// 步骤 3：向量化（Embedding）
	fmt.Println("🔢 步骤 3：向量化（模拟 Embedding）")
	fmt.Printf("   词汇表大小: %d\n", len(vocabulary))
	fmt.Printf("   向量维度: %d\n", len(vocabulary))

	store := NewVectorStore()
	for i, chunk := range chunks {
		embedding := simpleEmbedding(chunk, vocabulary)
		doc := Document{
			Content:   chunk,
			Embedding: embedding,
			Metadata: map[string]string{
				"source": "go-knowledge-base",
				"chunk":  fmt.Sprintf("%d", i),
			},
		}
		store.AddDocuments([]Document{doc})
		// 显示前 5 个维度
		fmt.Printf("   块 %d 向量（前5维）: [%.4f, %.4f, %.4f, %.4f, %.4f, ...]\n",
			i+1, embedding[0], embedding[1], embedding[2], embedding[3], embedding[4])
	}
	fmt.Println()

	// 步骤 4：向量检索
	fmt.Println("🔍 步骤 4：向量检索")
	queries := []string{
		"什么是 goroutine？它和线程有什么区别？",
		"Go 的垃圾回收是怎么工作的？",
		"context 包有什么用？",
	}

	for _, query := range queries {
		fmt.Printf("\n   查询: \"%s\"\n", query)

		// 将查询向量化
		queryEmbedding := simpleEmbedding(query, vocabulary)

		// 检索 Top-3 相关文档
		results := store.Search(queryEmbedding, 3)

		fmt.Println("   检索结果（Top-3）：")
		for i, result := range results {
			preview := result.Document.Content
			if len(preview) > 60 {
				preview = preview[:60] + "..."
			}
			fmt.Printf("     %d. [相似度: %.4f] %s\n", i+1, result.Score, preview)
		}

		// 步骤 5：构建 RAG Prompt
		ragPrompt := buildRAGPrompt(query, results)
		fmt.Printf("   RAG Prompt 长度: %d 字符\n", len(ragPrompt))
	}
	fmt.Println()

	// 步骤 5：模拟 LLM 生成
	fmt.Println("🤖 步骤 5：模拟 LLM 生成回答")
	fmt.Println()

	query := "什么是 goroutine？"
	queryEmbedding := simpleEmbedding(query, vocabulary)
	results := store.Search(queryEmbedding, 2)

	fmt.Printf("   问题: %s\n", query)
	fmt.Println("   检索到的上下文:")
	for i, r := range results {
		preview := r.Document.Content
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		fmt.Printf("     [%d] (%.4f) %s\n", i+1, r.Score, preview)
	}
	fmt.Println()

	// 模拟生成回答
	fmt.Println("   生成的回答（模拟）:")
	fmt.Println("   根据参考文档，Goroutine 是 Go 语言中的轻量级线程，由 Go 运行时")
	fmt.Println("   调度器管理。每个 goroutine 的初始栈大小仅为 2KB，可动态增长到 1GB，")
	fmt.Println("   远小于操作系统线程的默认栈大小。这使得 Go 程序可以轻松创建数十万")
	fmt.Println("   甚至上百万个 goroutine。")
	fmt.Println()

	// 演示余弦相似度计算
	fmt.Println("📐 附：余弦相似度计算演示")
	vecA := []float64{1.0, 0.0, 1.0, 0.0}
	vecB := []float64{1.0, 0.0, 1.0, 0.0}
	vecC := []float64{0.0, 1.0, 0.0, 1.0}
	vecD := []float64{0.5, 0.5, 0.5, 0.5}

	fmt.Printf("   A=[1,0,1,0] vs B=[1,0,1,0]: %.4f（完全相同）\n", cosineSimilarity(vecA, vecB))
	fmt.Printf("   A=[1,0,1,0] vs C=[0,1,0,1]: %.4f（完全正交）\n", cosineSimilarity(vecA, vecC))
	fmt.Printf("   A=[1,0,1,0] vs D=[.5,.5,.5,.5]: %.4f（部分相似）\n", cosineSimilarity(vecA, vecD))
	fmt.Println()

	fmt.Println("✅ Part A 完成！完整 RAG 流程演示结束。")
	fmt.Println("💡 要连接真实 LLM API，请运行: go run ./rag-example/ real")
}

// ============================================================
// Part B：连接真实 API
// ============================================================

func partB() {
	fmt.Println("========================================")
	fmt.Println("Part B：RAG + 真实 LLM API")
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

	// 词汇表
	vocabulary := []string{
		"goroutine", "channel", "并发", "线程", "调度",
		"gmp", "gc", "垃圾回收", "内存", "堆",
		"sync", "mutex", "锁", "waitgroup", "pool",
		"context", "取消", "超时", "传递", "信号",
		"error", "错误", "panic", "recover", "处理",
		"csp", "通信", "缓冲", "队列", "运行时",
	}

	// 构建向量存储
	chunks := chunkText(knowledgeBase, 300, 50)
	store := NewVectorStore()
	for i, chunk := range chunks {
		embedding := simpleEmbedding(chunk, vocabulary)
		store.AddDocuments([]Document{{
			Content:   chunk,
			Embedding: embedding,
			Metadata:  map[string]string{"chunk": fmt.Sprintf("%d", i)},
		}})
	}

	// RAG 查询
	query := "Go 的 GMP 调度模型是什么？"
	fmt.Printf("问题: %s\n\n", query)

	// 检索
	queryEmbedding := simpleEmbedding(query, vocabulary)
	results := store.Search(queryEmbedding, 3)

	fmt.Println("检索到的相关文档：")
	for i, r := range results {
		preview := r.Document.Content
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		fmt.Printf("  [%d] (相似度: %.4f) %s\n", i+1, r.Score, preview)
	}
	fmt.Println()

	// 构建 RAG Prompt 并调用 LLM
	ragPrompt := buildRAGPrompt(query, results)
	fmt.Println("正在调用 LLM 生成回答...")
	answer, err := callLLM(baseURL, apiKey, model, ragPrompt)
	if err != nil {
		fmt.Printf("❌ LLM 调用失败: %v\n", err)
		return
	}

	fmt.Println("\nLLM 回答：")
	fmt.Println(answer)
	fmt.Println()
	fmt.Println("✅ Part B 完成！")
}

// ============================================================
// 主函数
// ============================================================

func main() {
	// 固定随机种子以保证可重复性
	rand.New(rand.NewSource(42))

	fmt.Println("📚 RAG 检索增强生成示例")
	fmt.Println("完整流程：文档加载 → 文本分块 → 向量化 → 检索 → 生成")
	fmt.Println()

	// Part A：模拟模式
	partA()

	// Part B：连接真实 API
	if len(os.Args) > 1 && os.Args[1] == "real" {
		fmt.Println()
		partB()
	}
}
