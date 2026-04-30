// SQS 消息队列 — AWS SDK for Go v2 完整示例
// 演示：标准队列 / FIFO 队列 / 可见性超时 / 死信队列 / 长轮询
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟消息队列，直接运行理解 SQS 核心概念
// Part B：连接 LocalStack SQS，需传入参数 'real'
//
// 运行方式：
//   go run ./sqs/              # Part A：内存模拟
//   go run ./sqs/ real         # Part B：连接 LocalStack SQS
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.localstack.yml up -d
//   端点地址：http://localhost:4566
//   凭证：test / test（LocalStack 接受任意凭证）

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// ============================================================
// Part A：纯内存模拟 SQS 消息队列
// ============================================================

// Message 消息结构
type Message struct {
	ID            string
	Body          string
	ReceiptHandle string
	ReceiveCount  int
	SentAt        time.Time
	VisibleAfter  time.Time // 可见性超时后的时间
	GroupID       string    // FIFO 队列的消息组 ID
	DedupID       string    // FIFO 队列的去重 ID
}

// QueueConfig 队列配置
type QueueConfig struct {
	Name              string
	IsFIFO            bool
	VisibilityTimeout time.Duration // 可见性超时（默认 30s）
	MaxReceiveCount   int           // 最大接收次数（超过后转入死信队列）
	DeadLetterQueue   string        // 死信队列名称
}

// InMemoryQueue 内存消息队列
type InMemoryQueue struct {
	Config   QueueConfig
	messages []*Message
	mu       sync.Mutex
	dedupMap map[string]time.Time // FIFO 去重映射
}

// InMemorySQS 内存 SQS 服务
type InMemorySQS struct {
	mu     sync.RWMutex
	queues map[string]*InMemoryQueue
}

// NewInMemorySQS 创建内存 SQS 实例
func NewInMemorySQS() *InMemorySQS {
	return &InMemorySQS{
		queues: make(map[string]*InMemoryQueue),
	}
}

// CreateQueue 创建队列
func (s *InMemorySQS) CreateQueue(cfg QueueConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.queues[cfg.Name]; exists {
		return fmt.Errorf("queue %q already exists", cfg.Name)
	}
	if cfg.VisibilityTimeout == 0 {
		cfg.VisibilityTimeout = 30 * time.Second
	}
	if cfg.MaxReceiveCount == 0 {
		cfg.MaxReceiveCount = 3
	}
	s.queues[cfg.Name] = &InMemoryQueue{
		Config:   cfg,
		messages: make([]*Message, 0),
		dedupMap: make(map[string]time.Time),
	}
	return nil
}

// SendMessage 发送消息
func (s *InMemorySQS) SendMessage(queueName, body string, groupID, dedupID string) (string, error) {
	s.mu.RLock()
	q, exists := s.queues[queueName]
	s.mu.RUnlock()
	if !exists {
		return "", fmt.Errorf("queue %q not found", queueName)
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// FIFO 去重检查（5 分钟窗口）
	if q.Config.IsFIFO && dedupID != "" {
		if lastSent, ok := q.dedupMap[dedupID]; ok {
			if time.Since(lastSent) < 5*time.Minute {
				return "", fmt.Errorf("duplicate message (dedupID=%s) within 5min window", dedupID)
			}
		}
		q.dedupMap[dedupID] = time.Now()
	}

	msgID := fmt.Sprintf("msg-%d-%d", time.Now().UnixNano(), rand.Intn(10000))
	msg := &Message{
		ID:           msgID,
		Body:         body,
		SentAt:       time.Now(),
		VisibleAfter: time.Now(), // 立即可见
		GroupID:      groupID,
		DedupID:      dedupID,
	}
	q.messages = append(q.messages, msg)
	return msgID, nil
}

// ReceiveMessage 接收消息（模拟可见性超时）
func (s *InMemorySQS) ReceiveMessage(queueName string, maxMessages int) []*Message {
	s.mu.RLock()
	q, exists := s.queues[queueName]
	s.mu.RUnlock()
	if !exists {
		return nil
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	var result []*Message

	for _, msg := range q.messages {
		if len(result) >= maxMessages {
			break
		}
		// 只返回当前可见的消息
		if now.After(msg.VisibleAfter) {
			msg.ReceiveCount++
			msg.ReceiptHandle = fmt.Sprintf("receipt-%d-%d", time.Now().UnixNano(), rand.Intn(10000))
			msg.VisibleAfter = now.Add(q.Config.VisibilityTimeout)

			// 检查是否超过最大接收次数 → 转入死信队列
			if msg.ReceiveCount > q.Config.MaxReceiveCount && q.Config.DeadLetterQueue != "" {
				s.moveToDLQ(q, msg)
				continue
			}
			result = append(result, msg)
		}
	}
	return result
}

// moveToDLQ 将消息转移到死信队列
func (s *InMemorySQS) moveToDLQ(sourceQueue *InMemoryQueue, msg *Message) {
	s.mu.RLock()
	dlq, exists := s.queues[sourceQueue.Config.DeadLetterQueue]
	s.mu.RUnlock()
	if !exists {
		return
	}

	// 从源队列移除
	for i, m := range sourceQueue.messages {
		if m.ID == msg.ID {
			sourceQueue.messages = append(sourceQueue.messages[:i], sourceQueue.messages[i+1:]...)
			break
		}
	}

	// 添加到死信队列
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	dlqMsg := &Message{
		ID:           msg.ID,
		Body:         msg.Body,
		SentAt:       time.Now(),
		VisibleAfter: time.Now(),
		ReceiveCount: 0,
	}
	dlq.messages = append(dlq.messages, dlqMsg)
}

// DeleteMessage 删除消息（确认处理成功）
func (s *InMemorySQS) DeleteMessage(queueName, receiptHandle string) error {
	s.mu.RLock()
	q, exists := s.queues[queueName]
	s.mu.RUnlock()
	if !exists {
		return fmt.Errorf("queue %q not found", queueName)
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	for i, msg := range q.messages {
		if msg.ReceiptHandle == receiptHandle {
			q.messages = append(q.messages[:i], q.messages[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("message with receipt %q not found", receiptHandle)
}

// GetQueueDepth 获取队列深度（可见消息数）
func (s *InMemorySQS) GetQueueDepth(queueName string) int {
	s.mu.RLock()
	q, exists := s.queues[queueName]
	s.mu.RUnlock()
	if !exists {
		return 0
	}
	q.mu.Lock()
	defer q.mu.Unlock()

	count := 0
	now := time.Now()
	for _, msg := range q.messages {
		if now.After(msg.VisibleAfter) {
			count++
		}
	}
	return count
}

// ListQueues 列出所有队列
func (s *InMemorySQS) ListQueues() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, 0, len(s.queues))
	for name := range s.queues {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ============================================================
// Part A 演示
// ============================================================

// OrderEvent 订单事件（模拟业务消息）
type OrderEvent struct {
	OrderID string  `json:"orderId"`
	UserID  string  `json:"userId"`
	Amount  float64 `json:"amount"`
	Action  string  `json:"action"`
}

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：SQS 消息队列核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	sqsSvc := NewInMemorySQS()

	// --- 1. 标准队列 ---
	demoStandardQueue(sqsSvc)

	// --- 2. 可见性超时 ---
	demoVisibilityTimeout(sqsSvc)

	// --- 3. 死信队列 ---
	demoDeadLetterQueue(sqsSvc)

	// --- 4. FIFO 队列 ---
	demoFIFOQueue(sqsSvc)
}

// demoStandardQueue 标准队列基本操作
func demoStandardQueue(sqsSvc *InMemorySQS) {
	fmt.Println("\n--- 1. 标准队列 ---")

	sqsSvc.CreateQueue(QueueConfig{Name: "order-queue", VisibilityTimeout: 30 * time.Second})
	fmt.Println("  创建标准队列: order-queue ✅")

	// 发送消息
	orders := []OrderEvent{
		{OrderID: "ORD-001", UserID: "U100", Amount: 99.9, Action: "created"},
		{OrderID: "ORD-002", UserID: "U101", Amount: 199.0, Action: "created"},
		{OrderID: "ORD-003", UserID: "U102", Amount: 49.5, Action: "created"},
	}

	for _, order := range orders {
		body, _ := json.Marshal(order)
		msgID, _ := sqsSvc.SendMessage("order-queue", string(body), "", "")
		fmt.Printf("  发送: %s (msgID=%s)\n", order.OrderID, msgID[:15]+"...")
	}

	fmt.Printf("  队列深度: %d\n", sqsSvc.GetQueueDepth("order-queue"))

	// 接收并处理消息
	fmt.Println("\n  接收消息（模拟消费者）:")
	messages := sqsSvc.ReceiveMessage("order-queue", 10)
	for _, msg := range messages {
		var order OrderEvent
		json.Unmarshal([]byte(msg.Body), &order)
		fmt.Printf("    处理订单: %s (金额: %.1f, 接收次数: %d)\n",
			order.OrderID, order.Amount, msg.ReceiveCount)

		// 处理成功，删除消息
		sqsSvc.DeleteMessage("order-queue", msg.ReceiptHandle)
	}
	fmt.Printf("  处理后队列深度: %d ✅\n", sqsSvc.GetQueueDepth("order-queue"))
}

// demoVisibilityTimeout 可见性超时演示
func demoVisibilityTimeout(sqsSvc *InMemorySQS) {
	fmt.Println("\n--- 2. 可见性超时 ---")

	// 创建短超时队列用于演示
	sqsSvc.CreateQueue(QueueConfig{
		Name:              "timeout-demo",
		VisibilityTimeout: 2 * time.Second, // 2 秒超时（演示用）
	})

	sqsSvc.SendMessage("timeout-demo", `{"task":"process-image"}`, "", "")
	fmt.Println("  发送消息到 timeout-demo 队列")

	// 消费者 A 接收消息
	msgs := sqsSvc.ReceiveMessage("timeout-demo", 1)
	if len(msgs) > 0 {
		fmt.Printf("  消费者 A 接收: %s (消息进入不可见状态)\n", msgs[0].Body)
	}

	// 消费者 B 立即尝试接收（应该为空）
	msgs2 := sqsSvc.ReceiveMessage("timeout-demo", 1)
	fmt.Printf("  消费者 B 立即接收: %d 条消息（消息不可见）\n", len(msgs2))

	// 等待可见性超时
	fmt.Println("  等待可见性超时（2s）...")
	time.Sleep(2 * time.Second)

	// 消费者 B 再次接收（消息重新可见）
	msgs3 := sqsSvc.ReceiveMessage("timeout-demo", 1)
	if len(msgs3) > 0 {
		fmt.Printf("  消费者 B 重新接收: %s (接收次数: %d) ✅\n", msgs3[0].Body, msgs3[0].ReceiveCount)
		sqsSvc.DeleteMessage("timeout-demo", msgs3[0].ReceiptHandle)
	}
}

// demoDeadLetterQueue 死信队列演示
func demoDeadLetterQueue(sqsSvc *InMemorySQS) {
	fmt.Println("\n--- 3. 死信队列 ---")

	// 创建死信队列
	sqsSvc.CreateQueue(QueueConfig{Name: "order-dlq"})
	fmt.Println("  创建死信队列: order-dlq ✅")

	// 创建主队列（关联死信队列，最大接收 3 次）
	sqsSvc.CreateQueue(QueueConfig{
		Name:              "order-main",
		VisibilityTimeout: 1 * time.Second, // 短超时便于演示
		MaxReceiveCount:   3,
		DeadLetterQueue:   "order-dlq",
	})
	fmt.Println("  创建主队列: order-main (maxReceiveCount=3) ✅")

	// 发送一条"有毒"消息（模拟处理总是失败）
	sqsSvc.SendMessage("order-main", `{"orderId":"POISON-001","error":"invalid data"}`, "", "")
	fmt.Println("  发送有毒消息: POISON-001")

	// 模拟消费者反复接收但不删除（处理失败）
	for i := 1; i <= 4; i++ {
		time.Sleep(1100 * time.Millisecond) // 等待可见性超时
		msgs := sqsSvc.ReceiveMessage("order-main", 1)
		if len(msgs) > 0 {
			fmt.Printf("  第 %d 次接收: %s (receiveCount=%d)\n", i, msgs[0].ID[:15]+"...", msgs[0].ReceiveCount)
		} else {
			fmt.Printf("  第 %d 次接收: 主队列为空（消息已转入死信队列）\n", i)
		}
	}

	// 检查死信队列
	dlqMsgs := sqsSvc.ReceiveMessage("order-dlq", 10)
	fmt.Printf("\n  死信队列消息数: %d\n", len(dlqMsgs))
	for _, msg := range dlqMsgs {
		fmt.Printf("    DLQ 消息: %s\n", msg.Body)
	}
	fmt.Println("  ✅ 有毒消息已自动转入死信队列，便于人工排查")
}

// demoFIFOQueue FIFO 队列演示
func demoFIFOQueue(sqsSvc *InMemorySQS) {
	fmt.Println("\n--- 4. FIFO 队列 ---")

	sqsSvc.CreateQueue(QueueConfig{
		Name:   "order-processing.fifo",
		IsFIFO: true,
	})
	fmt.Println("  创建 FIFO 队列: order-processing.fifo ✅")

	// 发送有序消息（同一消息组保证顺序）
	events := []struct {
		body    string
		groupID string
		dedupID string
	}{
		{`{"orderId":"F-001","step":"created"}`, "order-F-001", "F-001-created"},
		{`{"orderId":"F-001","step":"paid"}`, "order-F-001", "F-001-paid"},
		{`{"orderId":"F-001","step":"shipped"}`, "order-F-001", "F-001-shipped"},
		{`{"orderId":"F-002","step":"created"}`, "order-F-002", "F-002-created"},
	}

	for _, e := range events {
		msgID, err := sqsSvc.SendMessage("order-processing.fifo", e.body, e.groupID, e.dedupID)
		if err != nil {
			fmt.Printf("  发送失败: %v\n", err)
			continue
		}
		fmt.Printf("  发送: group=%s, dedup=%s (msgID=%s)\n", e.groupID, e.dedupID, msgID[:15]+"...")
	}

	// 测试去重：重复发送同一 dedupID
	_, err := sqsSvc.SendMessage("order-processing.fifo", `{"orderId":"F-001","step":"created"}`, "order-F-001", "F-001-created")
	if err != nil {
		fmt.Printf("\n  去重测试: %v ✅\n", err)
	}

	// 按顺序接收
	fmt.Println("\n  按顺序接收消息:")
	msgs := sqsSvc.ReceiveMessage("order-processing.fifo", 10)
	for i, msg := range msgs {
		fmt.Printf("    [%d] group=%s body=%s\n", i+1, msg.GroupID, msg.Body)
	}
	fmt.Println("  ✅ FIFO 队列保证同一 MessageGroupId 内的消息严格有序")
}

// ============================================================
// Part B：连接 LocalStack SQS（AWS SDK for Go v2）
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：连接 LocalStack SQS（AWS SDK v2）")
	fmt.Println(strings.Repeat("=", 60))

	ctx := context.Background()

	// 加载 AWS 配置
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		),
	)
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		return
	}

	// 创建 SQS 客户端
	client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})

	// 测试连接
	_, err = client.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		fmt.Printf("❌ 连接 LocalStack 失败: %v\n", err)
		fmt.Println("请先启动: docker compose -f docker/docker-compose.localstack.yml up -d")
		return
	}
	fmt.Println("✅ LocalStack SQS 连接成功")

	// --- 1. 标准队列 ---
	demoRealStandardQueue(ctx, client)

	// --- 2. FIFO 队列 ---
	demoRealFIFOQueue(ctx, client)

	// --- 3. 清理 ---
	cleanupSQS(ctx, client)
}

// demoRealStandardQueue 真实标准队列操作
func demoRealStandardQueue(ctx context.Context, client *sqs.Client) {
	fmt.Println("\n--- 1. 标准队列 ---")

	queueName := "demo-order-queue"

	// 创建队列
	createOutput, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
		Attributes: map[string]string{
			"VisibilityTimeout": "30",
			"MessageRetentionPeriod": "86400", // 1 天
		},
	})
	if err != nil {
		fmt.Printf("  创建队列失败: %v\n", err)
		return
	}
	queueURL := *createOutput.QueueUrl
	fmt.Printf("  创建队列: %s\n  URL: %s\n", queueName, queueURL)

	// 发送消息
	orders := []OrderEvent{
		{OrderID: "ORD-001", UserID: "U100", Amount: 99.9, Action: "created"},
		{OrderID: "ORD-002", UserID: "U101", Amount: 199.0, Action: "created"},
		{OrderID: "ORD-003", UserID: "U102", Amount: 49.5, Action: "created"},
	}

	for _, order := range orders {
		body, _ := json.Marshal(order)
		sendOutput, err := client.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:    aws.String(queueURL),
			MessageBody: aws.String(string(body)),
			MessageAttributes: map[string]sqstypes.MessageAttributeValue{
				"OrderID": {
					DataType:    aws.String("String"),
					StringValue: aws.String(order.OrderID),
				},
			},
		})
		if err != nil {
			fmt.Printf("  发送失败: %v\n", err)
			continue
		}
		fmt.Printf("  发送: %s (MessageId=%s)\n", order.OrderID, (*sendOutput.MessageId)[:15]+"...")
	}

	// 接收消息（长轮询）
	fmt.Println("\n  接收消息（长轮询 5s）:")
	recvOutput, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(queueURL),
		MaxNumberOfMessages:   10,
		WaitTimeSeconds:       5, // 长轮询
		MessageAttributeNames: []string{"All"},
	})
	if err != nil {
		fmt.Printf("  接收失败: %v\n", err)
		return
	}

	for _, msg := range recvOutput.Messages {
		var order OrderEvent
		json.Unmarshal([]byte(*msg.Body), &order)
		fmt.Printf("    处理: %s (金额: %.1f)\n", order.OrderID, order.Amount)

		// 删除消息（确认处理成功）
		client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(queueURL),
			ReceiptHandle: msg.ReceiptHandle,
		})
	}
	fmt.Printf("  处理完成: %d 条消息 ✅\n", len(recvOutput.Messages))
}

// demoRealFIFOQueue 真实 FIFO 队列操作
func demoRealFIFOQueue(ctx context.Context, client *sqs.Client) {
	fmt.Println("\n--- 2. FIFO 队列 ---")

	queueName := "demo-order-processing.fifo"

	// 创建 FIFO 队列
	createOutput, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
		Attributes: map[string]string{
			"FifoQueue":                 "true",
			"ContentBasedDeduplication": "true",
		},
	})
	if err != nil {
		fmt.Printf("  创建 FIFO 队列失败: %v\n", err)
		return
	}
	queueURL := *createOutput.QueueUrl
	fmt.Printf("  创建 FIFO 队列: %s ✅\n", queueName)

	// 发送有序消息
	steps := []string{"created", "paid", "shipped"}
	for _, step := range steps {
		body := fmt.Sprintf(`{"orderId":"FIFO-001","step":"%s"}`, step)
		_, err := client.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:       aws.String(queueURL),
			MessageBody:    aws.String(body),
			MessageGroupId: aws.String("order-FIFO-001"),
		})
		if err != nil {
			fmt.Printf("  发送失败: %v\n", err)
			continue
		}
		fmt.Printf("  发送: step=%s, group=order-FIFO-001\n", step)
	}

	// 按顺序接收
	fmt.Println("\n  按顺序接收:")
	recvOutput, _ := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     5,
	})
	for i, msg := range recvOutput.Messages {
		fmt.Printf("    [%d] %s\n", i+1, *msg.Body)
		client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(queueURL),
			ReceiptHandle: msg.ReceiptHandle,
		})
	}
	fmt.Println("  ✅ FIFO 队列保证消息严格有序")
}

// cleanupSQS 清理测试队列
func cleanupSQS(ctx context.Context, client *sqs.Client) {
	fmt.Println("\n--- 清理测试数据 ---")

	queues := []string{"demo-order-queue", "demo-order-processing.fifo"}
	for _, name := range queues {
		// 获取队列 URL
		getOutput, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
			QueueName: aws.String(name),
		})
		if err != nil {
			continue
		}
		client.DeleteQueue(ctx, &sqs.DeleteQueueInput{
			QueueUrl: getOutput.QueueUrl,
		})
	}
	fmt.Println("  测试队列已清理 ✅")
}

func main() {
	// Part A：纯内存模拟，直接运行理解原理
	partA()

	// Part B：连接 LocalStack SQS，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		partB()
	}
}
