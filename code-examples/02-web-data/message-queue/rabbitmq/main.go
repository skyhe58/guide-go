// RabbitMQ 生产者/消费者完整示例 — amqp091-go 客户端
// 演示：Exchange 路由、Direct/Fanout/Topic 模式、死信队列、消息确认
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解 RabbitMQ 核心概念
// Part B：连接真实 RabbitMQ，需传入参数 'real'
//
// 运行方式：
//   go run ./rabbitmq/              # Part A：内存模拟
//   go run ./rabbitmq/ real         # Part B：连接真实 RabbitMQ
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.mq.yml up -d rabbitmq
//   连接地址：localhost:5672，用户名：guest，密码：guest

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ============================================================
// Part A：纯内存模拟 RabbitMQ 核心概念
// ============================================================

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：RabbitMQ 核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	// 1. 模拟 Exchange 路由
	demoExchangeRouting()

	// 2. 模拟死信队列
	demoDeadLetterQueue()

	// 3. 模拟延迟消息
	demoDelayedMessage()

	// 4. 模拟消息确认机制
	demoAckMechanism()
}

// ============================================================
// 1. 模拟 Exchange 路由
// ============================================================

// MemoryExchange 内存模拟的 Exchange
type MemoryExchange struct {
	Name     string
	Type     string // direct, fanout, topic
	Bindings map[string][]string // routingKey -> queueNames
}

// MemoryQueue 内存模拟的 Queue
type MemoryQueue struct {
	Name     string
	Messages []string
	Durable  bool
}

// MemoryRabbitMQ 内存模拟的 RabbitMQ Broker
type MemoryRabbitMQ struct {
	Exchanges map[string]*MemoryExchange
	Queues    map[string]*MemoryQueue
}

func NewMemoryRabbitMQ() *MemoryRabbitMQ {
	return &MemoryRabbitMQ{
		Exchanges: make(map[string]*MemoryExchange),
		Queues:    make(map[string]*MemoryQueue),
	}
}

func (r *MemoryRabbitMQ) DeclareExchange(name, typ string) {
	r.Exchanges[name] = &MemoryExchange{
		Name:     name,
		Type:     typ,
		Bindings: make(map[string][]string),
	}
}

func (r *MemoryRabbitMQ) DeclareQueue(name string, durable bool) {
	r.Queues[name] = &MemoryQueue{Name: name, Durable: durable}
}

func (r *MemoryRabbitMQ) Bind(exchange, routingKey, queue string) {
	ex := r.Exchanges[exchange]
	ex.Bindings[routingKey] = append(ex.Bindings[routingKey], queue)
}

func (r *MemoryRabbitMQ) Publish(exchange, routingKey, message string) []string {
	ex := r.Exchanges[exchange]
	var delivered []string

	switch ex.Type {
	case "direct":
		// 精确匹配 routing key
		for _, qName := range ex.Bindings[routingKey] {
			r.Queues[qName].Messages = append(r.Queues[qName].Messages, message)
			delivered = append(delivered, qName)
		}
	case "fanout":
		// 广播到所有绑定的队列
		seen := make(map[string]bool)
		for _, queues := range ex.Bindings {
			for _, qName := range queues {
				if !seen[qName] {
					r.Queues[qName].Messages = append(r.Queues[qName].Messages, message)
					delivered = append(delivered, qName)
					seen[qName] = true
				}
			}
		}
	case "topic":
		// 通配符匹配
		for bindKey, queues := range ex.Bindings {
			if topicMatch(bindKey, routingKey) {
				for _, qName := range queues {
					r.Queues[qName].Messages = append(r.Queues[qName].Messages, message)
					delivered = append(delivered, qName)
				}
			}
		}
	}
	return delivered
}

// topicMatch 模拟 RabbitMQ Topic 匹配（* 匹配一个词，# 匹配多个词）
func topicMatch(pattern, key string) bool {
	pParts := strings.Split(pattern, ".")
	kParts := strings.Split(key, ".")

	pi, ki := 0, 0
	for pi < len(pParts) && ki < len(kParts) {
		if pParts[pi] == "#" {
			return true
		}
		if pParts[pi] == "*" || pParts[pi] == kParts[ki] {
			pi++
			ki++
		} else {
			return false
		}
	}
	return pi == len(pParts) && ki == len(kParts)
}

func demoExchangeRouting() {
	fmt.Println("\n--- 1. Exchange 路由机制（内存模拟） ---")

	rmq := NewMemoryRabbitMQ()

	// --- Direct Exchange ---
	fmt.Println("\n[Direct Exchange] 精确匹配 Routing Key")
	rmq.DeclareExchange("logs.direct", "direct")
	rmq.DeclareQueue("error-queue", true)
	rmq.DeclareQueue("info-queue", true)
	rmq.Bind("logs.direct", "error", "error-queue")
	rmq.Bind("logs.direct", "info", "info-queue")

	delivered := rmq.Publish("logs.direct", "error", "数据库连接失败")
	fmt.Printf("  PUBLISH routing_key=error → 投递到: %v\n", delivered)
	delivered = rmq.Publish("logs.direct", "info", "用户登录成功")
	fmt.Printf("  PUBLISH routing_key=info → 投递到: %v\n", delivered)

	// --- Fanout Exchange ---
	fmt.Println("\n[Fanout Exchange] 广播到所有绑定队列")
	rmq.DeclareExchange("notifications", "fanout")
	rmq.DeclareQueue("email-queue", true)
	rmq.DeclareQueue("sms-queue", true)
	rmq.DeclareQueue("push-queue", true)
	rmq.Bind("notifications", "", "email-queue")
	rmq.Bind("notifications", "", "sms-queue")
	rmq.Bind("notifications", "", "push-queue")

	delivered = rmq.Publish("notifications", "", "新订单通知")
	fmt.Printf("  PUBLISH (fanout) → 投递到: %v\n", delivered)

	// --- Topic Exchange ---
	fmt.Println("\n[Topic Exchange] 通配符匹配")
	rmq.DeclareExchange("events", "topic")
	rmq.DeclareQueue("order-events", true)
	rmq.DeclareQueue("all-errors", true)
	rmq.DeclareQueue("all-events", true)
	rmq.Bind("events", "order.*", "order-events")
	rmq.Bind("events", "*.error", "all-errors")
	rmq.Bind("events", "#", "all-events")

	delivered = rmq.Publish("events", "order.created", "订单创建")
	fmt.Printf("  PUBLISH routing_key=order.created → 投递到: %v\n", delivered)
	delivered = rmq.Publish("events", "payment.error", "支付失败")
	fmt.Printf("  PUBLISH routing_key=payment.error → 投递到: %v\n", delivered)
}

// ============================================================
// 2. 模拟死信队列
// ============================================================

func demoDeadLetterQueue() {
	fmt.Println("\n--- 2. 死信队列 DLQ（内存模拟） ---")

	type QueueMessage struct {
		Body    string
		TTL     time.Duration
		Retries int
	}

	normalQueue := []QueueMessage{
		{"订单-1001 支付超时", 30 * time.Second, 0},
		{"订单-1002 处理失败", 0, 3},
		{"订单-1003 正常处理", 0, 0},
	}

	var deadLetterQueue []QueueMessage

	fmt.Println("\n  处理普通队列消息：")
	for _, msg := range normalQueue {
		if msg.TTL > 0 {
			fmt.Printf("    消息: %s → TTL 过期，转入死信队列\n", msg.Body)
			deadLetterQueue = append(deadLetterQueue, msg)
		} else if msg.Retries >= 3 {
			fmt.Printf("    消息: %s → 重试 %d 次失败，转入死信队列\n", msg.Body, msg.Retries)
			deadLetterQueue = append(deadLetterQueue, msg)
		} else {
			fmt.Printf("    消息: %s → ✅ 处理成功\n", msg.Body)
		}
	}

	fmt.Println("\n  死信队列中的消息：")
	for _, msg := range deadLetterQueue {
		fmt.Printf("    %s\n", msg.Body)
	}

	fmt.Println("\n  消息进入死信队列的三种情况：")
	fmt.Println("    1. 消费者拒绝消息（Nack 且 requeue=false）")
	fmt.Println("    2. 消息 TTL 过期")
	fmt.Println("    3. 队列达到最大长度")
}

// ============================================================
// 3. 模拟延迟消息
// ============================================================

func demoDelayedMessage() {
	fmt.Println("\n--- 3. 延迟消息（内存模拟） ---")

	fmt.Println("\n  方案一：TTL + 死信队列")
	fmt.Println("    1. 创建普通队列，设置 x-message-ttl=30000（30秒）")
	fmt.Println("    2. 设置 x-dead-letter-exchange 指向死信交换机")
	fmt.Println("    3. 消息过期后自动转入死信队列")
	fmt.Println("    4. 消费者从死信队列消费 → 实现延迟效果")
	fmt.Println("    ⚠️ 缺点：队列头部消息未过期会阻塞后续消息")

	fmt.Println("\n  方案二：延迟消息插件（推荐）")
	fmt.Println("    1. 安装 rabbitmq_delayed_message_exchange 插件")
	fmt.Println("    2. 声明 x-delayed-message 类型的 Exchange")
	fmt.Println("    3. 发送消息时设置 x-delay 头部（毫秒）")
	fmt.Println("    ✅ 优点：灵活设置任意延迟时间，无阻塞问题")

	fmt.Println("\n  [模拟] 订单超时取消场景：")
	type DelayedMsg struct {
		Body  string
		Delay time.Duration
	}
	msgs := []DelayedMsg{
		{"订单-1001 超时检查", 30 * time.Minute},
		{"订单-1002 超时检查", 30 * time.Minute},
		{"优惠券-2001 过期提醒", 24 * time.Hour},
	}
	for _, m := range msgs {
		fmt.Printf("    发送延迟消息: %s（延迟 %v）\n", m.Body, m.Delay)
	}
}

// ============================================================
// 4. 模拟消息确认机制
// ============================================================

func demoAckMechanism() {
	fmt.Println("\n--- 4. 消息确认机制（内存模拟） ---")

	fmt.Println("\n  [生产者确认 — Publisher Confirm]")
	fmt.Println("    1. 生产者开启 Confirm 模式")
	fmt.Println("    2. 发送消息后等待 Broker 确认")
	fmt.Println("    3. Broker 写入磁盘后返回 Ack")
	fmt.Println("    4. 超时未确认则重发")

	fmt.Println("\n  [消费者确认 — Consumer Ack]")
	type ConsumeResult struct {
		Message string
		Success bool
		Action  string
	}
	results := []ConsumeResult{
		{"订单-1001", true, "Ack（确认消费成功）"},
		{"订单-1002", false, "Nack + requeue=true（重新入队重试）"},
		{"订单-1003", false, "Nack + requeue=false（转入死信队列）"},
	}
	for _, r := range results {
		status := "✅"
		if !r.Success {
			status = "❌"
		}
		fmt.Printf("    消息: %s → %s %s\n", r.Message, status, r.Action)
	}

	fmt.Println("\n  ⚠️  关键点：")
	fmt.Println("    1. 关闭自动 Ack（autoAck=false），手动确认")
	fmt.Println("    2. 处理成功后再 Ack，避免消息丢失")
	fmt.Println("    3. 设置 Prefetch Count 控制预取数量")
}

// ============================================================
// Part B：连接真实 RabbitMQ（需要 Docker）
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：连接真实 RabbitMQ（amqp091-go）")
	fmt.Println(strings.Repeat("=", 60))

	// 建立连接
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Printf("❌ 无法连接 RabbitMQ: %v\n", err)
		fmt.Println("请先启动 RabbitMQ: docker compose -f docker/docker-compose.mq.yml up -d rabbitmq")
		return
	}
	defer conn.Close()
	fmt.Println("✅ RabbitMQ 连接成功")

	// 1. Direct Exchange 示例
	demoRealDirectExchange(conn)

	// 2. Fanout Exchange 示例
	demoRealFanoutExchange(conn)

	// 3. 简单队列生产/消费
	demoRealSimpleQueue(conn)

	// 清理
	cleanupRabbitMQ(conn)
}

// demoRealDirectExchange 演示 Direct Exchange
func demoRealDirectExchange(conn *amqp.Connection) {
	fmt.Println("\n--- 1. Direct Exchange ---")

	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("  创建 Channel 失败: %v\n", err)
		return
	}
	defer ch.Close()

	// 声明 Exchange
	exchangeName := "demo.logs.direct"
	err = ch.ExchangeDeclare(exchangeName, "direct", false, true, false, false, nil)
	if err != nil {
		fmt.Printf("  声明 Exchange 失败: %v\n", err)
		return
	}

	// 声明并绑定队列
	errorQ, _ := ch.QueueDeclare("demo-error-queue", false, true, false, false, nil)
	infoQ, _ := ch.QueueDeclare("demo-info-queue", false, true, false, false, nil)
	ch.QueueBind(errorQ.Name, "error", exchangeName, false, nil)
	ch.QueueBind(infoQ.Name, "info", exchangeName, false, nil)

	ctx := context.Background()

	// 发布消息
	ch.PublishWithContext(ctx, exchangeName, "error", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("数据库连接超时"),
	})
	ch.PublishWithContext(ctx, exchangeName, "info", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("用户登录成功"),
	})
	fmt.Println("  ✅ 发布 2 条消息到 Direct Exchange")

	// 消费
	errorMsgs, _ := ch.Consume(errorQ.Name, "", true, false, false, false, nil)
	infoMsgs, _ := ch.Consume(infoQ.Name, "", true, false, false, false, nil)

	timeout := time.After(2 * time.Second)
	for i := 0; i < 2; i++ {
		select {
		case msg := <-errorMsgs:
			fmt.Printf("  📨 error-queue: %s\n", string(msg.Body))
		case msg := <-infoMsgs:
			fmt.Printf("  📨 info-queue: %s\n", string(msg.Body))
		case <-timeout:
			return
		}
	}
}

// demoRealFanoutExchange 演示 Fanout Exchange
func demoRealFanoutExchange(conn *amqp.Connection) {
	fmt.Println("\n--- 2. Fanout Exchange（广播） ---")

	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("  创建 Channel 失败: %v\n", err)
		return
	}
	defer ch.Close()

	exchangeName := "demo.notifications"
	ch.ExchangeDeclare(exchangeName, "fanout", false, true, false, false, nil)

	// 声明 3 个队列并绑定
	queueNames := []string{"demo-email", "demo-sms", "demo-push"}
	consumers := make([]<-chan amqp.Delivery, 3)
	for i, name := range queueNames {
		q, _ := ch.QueueDeclare(name, false, true, false, false, nil)
		ch.QueueBind(q.Name, "", exchangeName, false, nil)
		consumers[i], _ = ch.Consume(q.Name, "", true, false, false, false, nil)
	}

	ctx := context.Background()
	ch.PublishWithContext(ctx, exchangeName, "", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("新订单通知：订单 1001 已创建"),
	})
	fmt.Println("  ✅ 发布 1 条广播消息")

	// 每个队列都应收到
	timeout := time.After(2 * time.Second)
	for i, consumer := range consumers {
		select {
		case msg := <-consumer:
			fmt.Printf("  📨 %s: %s\n", queueNames[i], string(msg.Body))
		case <-timeout:
		}
	}
}

// demoRealSimpleQueue 演示简单队列生产/消费
func demoRealSimpleQueue(conn *amqp.Connection) {
	fmt.Println("\n--- 3. 简单队列（生产/消费 + 手动 Ack） ---")

	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("  创建 Channel 失败: %v\n", err)
		return
	}
	defer ch.Close()

	queueName := "demo-orders"
	q, _ := ch.QueueDeclare(queueName, false, true, false, false, nil)

	// 设置 Prefetch
	ch.Qos(1, 0, false)

	ctx := context.Background()

	// 发布消息
	for i := 1; i <= 3; i++ {
		body := fmt.Sprintf(`{"id":"%d","user":"用户%d","amount":%d}`, i, i, i*100)
		ch.PublishWithContext(ctx, "", q.Name, false, false, amqp.Publishing{
			ContentType:  "application/json",
			Body:         []byte(body),
			DeliveryMode: amqp.Persistent, // 持久化消息
		})
	}
	fmt.Println("  ✅ 发布 3 条订单消息")

	// 消费消息（手动 Ack）
	msgs, _ := ch.Consume(q.Name, "", false, false, false, false, nil)
	timeout := time.After(2 * time.Second)
	for i := 0; i < 3; i++ {
		select {
		case msg := <-msgs:
			fmt.Printf("  📨 收到: %s\n", string(msg.Body))
			msg.Ack(false) // 手动确认
		case <-timeout:
			return
		}
	}
}

// cleanupRabbitMQ 清理测试资源
func cleanupRabbitMQ(conn *amqp.Connection) {
	ch, err := conn.Channel()
	if err != nil {
		return
	}
	defer ch.Close()

	// 删除测试队列和 Exchange
	queues := []string{"demo-error-queue", "demo-info-queue", "demo-email", "demo-sms", "demo-push", "demo-orders"}
	for _, q := range queues {
		ch.QueueDelete(q, false, false, false)
	}
	ch.ExchangeDelete("demo.logs.direct", false, false)
	ch.ExchangeDelete("demo.notifications", false, false)
}

// ============================================================
// main 入口
// ============================================================

func main() {
	// Part A：纯内存模拟，直接运行理解原理
	partA()

	// Part B：连接真实 RabbitMQ，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		partB()
	}
}
