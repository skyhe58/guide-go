// NATS 消息系统完整示例 — Core NATS + JetStream
// 演示：发布/订阅、队列组、Request-Reply、JetStream 持久化
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解 NATS 核心概念
// Part B：连接真实 NATS，需传入参数 'real'
//
// 运行方式：
//   go run ./nats/              # Part A：内存模拟
//   go run ./nats/ real         # Part B：连接真实 NATS
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.mq.yml up -d nats
//   连接地址：localhost:4222

package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// ============================================================
// Part A：纯内存模拟 NATS 核心概念
// ============================================================

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：NATS 核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	// 1. 模拟发布/订阅
	demoPubSubConcept()

	// 2. 模拟队列组（负载均衡）
	demoQueueGroupConcept()

	// 3. 模拟 Request-Reply
	demoRequestReplyConcept()

	// 4. 模拟 Subject 通配符
	demoSubjectWildcards()

	// 5. 模拟 JetStream 概念
	demoJetStreamConcept()
}

// ============================================================
// 1. 模拟发布/订阅
// ============================================================

// MemoryBroker 内存模拟的消息代理
type MemoryBroker struct {
	mu          sync.RWMutex
	subscribers map[string][]chan string // subject -> channels
}

func NewMemoryBroker() *MemoryBroker {
	return &MemoryBroker{
		subscribers: make(map[string][]chan string),
	}
}

func (b *MemoryBroker) Subscribe(subject string) chan string {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan string, 10)
	b.subscribers[subject] = append(b.subscribers[subject], ch)
	return ch
}

func (b *MemoryBroker) Publish(subject, data string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	count := 0
	for sub, channels := range b.subscribers {
		if matchSubject(sub, subject) {
			for _, ch := range channels {
				select {
				case ch <- fmt.Sprintf("[%s] %s", subject, data):
					count++
				default:
				}
			}
		}
	}
	return count
}

// matchSubject 简单的 Subject 匹配（支持 * 和 > 通配符）
func matchSubject(pattern, subject string) bool {
	if pattern == subject {
		return true
	}
	pParts := strings.Split(pattern, ".")
	sParts := strings.Split(subject, ".")

	for i, pp := range pParts {
		if pp == ">" {
			return true // > 匹配剩余所有
		}
		if i >= len(sParts) {
			return false
		}
		if pp != "*" && pp != sParts[i] {
			return false
		}
	}
	return len(pParts) == len(sParts)
}

func demoPubSubConcept() {
	fmt.Println("\n--- 1. 发布/订阅模式（内存模拟） ---")

	broker := NewMemoryBroker()

	// 订阅者 A：订阅所有订单事件
	subA := broker.Subscribe("orders.>")
	// 订阅者 B：只订阅订单创建事件
	subB := broker.Subscribe("orders.created")

	// 发布消息
	fmt.Println("\n发布消息：")
	n := broker.Publish("orders.created", `{"id":"1001","user":"张三"}`)
	fmt.Printf("  PUBLISH orders.created → %d 个订阅者收到\n", n)

	n = broker.Publish("orders.paid", `{"id":"1001","status":"paid"}`)
	fmt.Printf("  PUBLISH orders.paid → %d 个订阅者收到\n", n)

	// 读取消息
	fmt.Println("\n订阅者 A（orders.>）收到：")
	for i := 0; i < 2; i++ {
		select {
		case msg := <-subA:
			fmt.Printf("  %s\n", msg)
		default:
		}
	}

	fmt.Println("订阅者 B（orders.created）收到：")
	select {
	case msg := <-subB:
		fmt.Printf("  %s\n", msg)
	default:
		fmt.Println("  （无消息）")
	}

	fmt.Println("\n⚠️  Pub/Sub 特点：")
	fmt.Println("  1. 所有匹配的订阅者都会收到消息（广播）")
	fmt.Println("  2. Core NATS 不持久化，订阅者离线时消息丢失")
}

// ============================================================
// 2. 模拟队列组（负载均衡）
// ============================================================

func demoQueueGroupConcept() {
	fmt.Println("\n--- 2. 队列组 Queue Group（内存模拟） ---")

	fmt.Println("\n[场景] 3 个 Worker 加入同一队列组 'workers'")
	fmt.Println("  每条消息只投递给其中一个 Worker（负载均衡）")

	workers := []string{"Worker-1", "Worker-2", "Worker-3"}
	tasks := []string{"任务A", "任务B", "任务C", "任务D", "任务E", "任务F"}

	fmt.Println("\n消息分发（随机轮询）：")
	for _, task := range tasks {
		worker := workers[rand.Intn(len(workers))]
		fmt.Printf("  %s → %s\n", task, worker)
	}

	fmt.Println("\n⚠️  Queue Group vs Kafka 消费组：")
	fmt.Println("  1. NATS Queue Group 不需要分区概念")
	fmt.Println("  2. 任意数量的消费者都能参与负载均衡")
	fmt.Println("  3. 更轻量，无 Rebalance 开销")
}

// ============================================================
// 3. 模拟 Request-Reply
// ============================================================

func demoRequestReplyConcept() {
	fmt.Println("\n--- 3. Request-Reply 模式（内存模拟） ---")

	fmt.Println("\n[场景] 用户服务请求订单服务查询订单详情")
	fmt.Println()

	// 模拟请求-响应
	type Request struct {
		Subject string
		Data    string
		ReplyTo string
	}

	req := Request{
		Subject: "api.order.get",
		Data:    `{"order_id":"1001"}`,
		ReplyTo: "_INBOX.abc123", // NATS 自动生成的回复地址
	}

	fmt.Printf("  1. 用户服务发送请求:\n")
	fmt.Printf("     Subject: %s\n", req.Subject)
	fmt.Printf("     Data: %s\n", req.Data)
	fmt.Printf("     ReplyTo: %s\n", req.ReplyTo)

	fmt.Printf("\n  2. 订单服务收到请求，处理后回复:\n")
	reply := `{"id":"1001","user":"张三","amount":99.9,"status":"paid"}`
	fmt.Printf("     Reply → %s\n", req.ReplyTo)
	fmt.Printf("     Data: %s\n", reply)

	fmt.Printf("\n  3. 用户服务收到响应:\n")
	fmt.Printf("     %s\n", reply)

	fmt.Println("\n⚠️  Request-Reply 特点：")
	fmt.Println("  1. 同步请求-响应模式，类似 HTTP 但基于消息")
	fmt.Println("  2. NATS 自动生成唯一的 _INBOX 回复地址")
	fmt.Println("  3. 支持超时设置，避免无限等待")
}

// ============================================================
// 4. 模拟 Subject 通配符
// ============================================================

func demoSubjectWildcards() {
	fmt.Println("\n--- 4. Subject 通配符（内存模拟） ---")

	type MatchTest struct {
		Pattern string
		Subject string
		Match   bool
	}

	tests := []MatchTest{
		{"orders.*", "orders.created", true},
		{"orders.*", "orders.us.created", false},
		{"orders.>", "orders.created", true},
		{"orders.>", "orders.us.created", true},
		{"sensor.*.temp", "sensor.room1.temp", true},
		{"sensor.*.temp", "sensor.room1.humidity", false},
	}

	fmt.Println()
	for _, t := range tests {
		result := "✅ 匹配"
		if !t.Match {
			result = "❌ 不匹配"
		}
		fmt.Printf("  Pattern: %-20s Subject: %-25s → %s\n", t.Pattern, t.Subject, result)
	}

	fmt.Println("\n  * — 匹配单个 token（不跨越 .）")
	fmt.Println("  > — 匹配一个或多个 token（只能在末尾）")
}

// ============================================================
// 5. 模拟 JetStream 概念
// ============================================================

func demoJetStreamConcept() {
	fmt.Println("\n--- 5. JetStream 持久化（内存模拟） ---")

	// 模拟 Stream 存储
	type StreamMessage struct {
		Seq     uint64
		Subject string
		Data    string
		Time    time.Time
	}

	stream := []StreamMessage{
		{1, "orders.created", `{"id":"1001"}`, time.Now().Add(-5 * time.Minute)},
		{2, "orders.paid", `{"id":"1001"}`, time.Now().Add(-4 * time.Minute)},
		{3, "orders.created", `{"id":"1002"}`, time.Now().Add(-3 * time.Minute)},
		{4, "orders.shipped", `{"id":"1001"}`, time.Now().Add(-2 * time.Minute)},
		{5, "orders.created", `{"id":"1003"}`, time.Now().Add(-1 * time.Minute)},
	}

	fmt.Println("\n  Stream: ORDERS（持久化存储）")
	fmt.Println("  Subjects: orders.>")
	fmt.Println("  Storage: File")
	fmt.Printf("  Messages: %d\n", len(stream))

	fmt.Println("\n  存储的消息：")
	for _, msg := range stream {
		fmt.Printf("    [Seq %d] %s → %s\n", msg.Seq, msg.Subject, msg.Data)
	}

	fmt.Println("\n  Core NATS vs JetStream：")
	fmt.Println("    Core NATS: At-Most-Once，不持久化，极致性能")
	fmt.Println("    JetStream: At-Least-Once，持久化到磁盘，支持回放")
	fmt.Println("    JetStream 类似 Kafka 的持久化能力，但更轻量")
}

// ============================================================
// Part B：连接真实 NATS（需要 Docker）
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：连接真实 NATS（nats.go）")
	fmt.Println(strings.Repeat("=", 60))

	// 连接 NATS
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		fmt.Printf("❌ 无法连接 NATS: %v\n", err)
		fmt.Println("请先启动 NATS: docker compose -f docker/docker-compose.mq.yml up -d nats")
		return
	}
	defer nc.Close()
	fmt.Println("✅ NATS 连接成功")

	// 1. 发布/订阅
	demoRealPubSub(nc)

	// 2. 队列组
	demoRealQueueGroup(nc)

	// 3. Request-Reply
	demoRealRequestReply(nc)

	// 4. JetStream
	demoRealJetStream(nc)
}

// demoRealPubSub 演示真实的发布/订阅
func demoRealPubSub(nc *nats.Conn) {
	fmt.Println("\n--- 1. 发布/订阅（Core NATS） ---")

	var wg sync.WaitGroup
	wg.Add(2)

	// 订阅者
	sub, err := nc.Subscribe("demo.orders.>", func(msg *nats.Msg) {
		fmt.Printf("  📨 订阅者收到: subject=%s data=%s\n", msg.Subject, string(msg.Data))
		wg.Done()
	})
	if err != nil {
		fmt.Printf("  订阅失败: %v\n", err)
		return
	}
	defer sub.Unsubscribe()

	time.Sleep(100 * time.Millisecond)

	// 发布消息
	nc.Publish("demo.orders.created", []byte(`{"id":"1001","user":"张三"}`))
	nc.Publish("demo.orders.paid", []byte(`{"id":"1001","status":"paid"}`))
	nc.Flush()

	// 等待消息处理
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
}

// demoRealQueueGroup 演示真实的队列组
func demoRealQueueGroup(nc *nats.Conn) {
	fmt.Println("\n--- 2. 队列组（Queue Group） ---")

	var wg sync.WaitGroup
	wg.Add(3)

	// 3 个 Worker 加入同一队列组
	for i := 1; i <= 3; i++ {
		workerID := i
		nc.QueueSubscribe("demo.tasks", "workers", func(msg *nats.Msg) {
			fmt.Printf("  📨 Worker-%d 收到: %s\n", workerID, string(msg.Data))
			wg.Done()
		})
	}

	time.Sleep(100 * time.Millisecond)

	// 发布 3 条任务
	for i := 1; i <= 3; i++ {
		nc.Publish("demo.tasks", []byte(fmt.Sprintf("任务-%d", i)))
	}
	nc.Flush()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
}

// demoRealRequestReply 演示真实的 Request-Reply
func demoRealRequestReply(nc *nats.Conn) {
	fmt.Println("\n--- 3. Request-Reply ---")

	// 响应者
	nc.Subscribe("demo.api.user.get", func(msg *nats.Msg) {
		reply := fmt.Sprintf(`{"id":"%s","name":"张三","email":"zhangsan@example.com"}`, string(msg.Data))
		msg.Respond([]byte(reply))
	})

	time.Sleep(100 * time.Millisecond)

	// 请求者
	resp, err := nc.Request("demo.api.user.get", []byte("1001"), 2*time.Second)
	if err != nil {
		fmt.Printf("  请求失败: %v\n", err)
		return
	}
	fmt.Printf("  📨 收到响应: %s\n", string(resp.Data))
}

// demoRealJetStream 演示真实的 JetStream
func demoRealJetStream(nc *nats.Conn) {
	fmt.Println("\n--- 4. JetStream 持久化 ---")

	js, err := nc.JetStream()
	if err != nil {
		fmt.Printf("  JetStream 初始化失败: %v\n", err)
		return
	}

	// 创建或更新 Stream
	streamName := "DEMO_ORDERS"
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{"demo.js.orders.>"},
		Storage:  nats.FileStorage,
		MaxAge:   1 * time.Hour,
	})
	if err != nil {
		fmt.Printf("  创建 Stream 失败: %v\n", err)
		return
	}
	fmt.Printf("  ✅ Stream '%s' 创建成功\n", streamName)

	// 发布持久化消息
	for i := 1; i <= 3; i++ {
		ack, err := js.Publish("demo.js.orders.created",
			[]byte(fmt.Sprintf(`{"id":"%d","user":"用户%d"}`, i, i)))
		if err != nil {
			fmt.Printf("  发布失败: %v\n", err)
			continue
		}
		fmt.Printf("  ✅ 发布成功: stream=%s seq=%d\n", ack.Stream, ack.Sequence)
	}

	// Pull 消费者
	sub, err := js.PullSubscribe("demo.js.orders.>", "demo-processor",
		nats.AckExplicit())
	if err != nil {
		fmt.Printf("  创建消费者失败: %v\n", err)
		return
	}

	msgs, err := sub.Fetch(10, nats.MaxWait(2*time.Second))
	if err != nil && err != nats.ErrTimeout {
		fmt.Printf("  拉取消息失败: %v\n", err)
		return
	}
	fmt.Printf("  拉取到 %d 条消息:\n", len(msgs))
	for _, msg := range msgs {
		fmt.Printf("    📨 subject=%s data=%s\n", msg.Subject, string(msg.Data))
		msg.Ack()
	}

	// 清理
	js.DeleteStream(streamName)
}

// ============================================================
// main 入口
// ============================================================

func main() {
	// Part A：纯内存模拟，直接运行理解原理
	partA()

	// Part B：连接真实 NATS，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		partB()
	}
}
