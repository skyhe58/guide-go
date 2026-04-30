// Kafka 生产者/消费者完整示例 — sarama 客户端
// 演示：Kafka 分区/消费组概念、同步生产者、消费组消费
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解 Kafka 核心概念
// Part B：连接真实 Kafka，需传入参数 'real'
//
// 运行方式：
//   go run ./kafka/              # Part A：内存模拟
//   go run ./kafka/ real         # Part B：连接真实 Kafka
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.mq.yml up -d kafka
//   连接地址：localhost:9092

package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// ============================================================
// Part A：纯内存模拟 Kafka 核心概念
// ============================================================

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：Kafka 核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	// 1. 模拟 Kafka 分区机制
	demoPartitionConcept()

	// 2. 模拟消费组
	demoConsumerGroupConcept()

	// 3. 模拟消息可靠性（acks 机制）
	demoAcksConcept()

	// 4. 模拟幂等生产者
	demoIdempotentProducer()
}

// ============================================================
// 1. 模拟 Kafka 分区机制
// ============================================================

// Partition 模拟 Kafka 分区
type Partition struct {
	ID       int
	Messages []Message
	Offset   int64
}

// Message 模拟 Kafka 消息
type Message struct {
	Key       string
	Value     string
	Partition int
	Offset    int64
}

// Topic 模拟 Kafka Topic
type Topic struct {
	Name       string
	Partitions []*Partition
}

// NewTopic 创建一个模拟 Topic
func NewTopic(name string, numPartitions int) *Topic {
	partitions := make([]*Partition, numPartitions)
	for i := 0; i < numPartitions; i++ {
		partitions[i] = &Partition{ID: i}
	}
	return &Topic{Name: name, Partitions: partitions}
}

// Produce 模拟生产消息（按 Key 哈希分区）
func (t *Topic) Produce(key, value string) Message {
	// 按 Key 哈希选择分区（模拟 Kafka 的分区策略）
	partIdx := 0
	if key != "" {
		hash := 0
		for _, c := range key {
			hash = hash*31 + int(c)
		}
		if hash < 0 {
			hash = -hash
		}
		partIdx = hash % len(t.Partitions)
	} else {
		partIdx = rand.Intn(len(t.Partitions))
	}

	p := t.Partitions[partIdx]
	msg := Message{
		Key:       key,
		Value:     value,
		Partition: partIdx,
		Offset:    int64(len(p.Messages)),
	}
	p.Messages = append(p.Messages, msg)
	return msg
}

func demoPartitionConcept() {
	fmt.Println("\n--- 1. Kafka 分区机制（内存模拟） ---")

	topic := NewTopic("orders", 3)

	// 模拟生产消息
	orders := []struct{ key, value string }{
		{"order-1001", `{"id":"1001","user":"张三","amount":99.9}`},
		{"order-1002", `{"id":"1002","user":"李四","amount":199.0}`},
		{"order-1001", `{"id":"1001","status":"paid"}`},
		{"order-1003", `{"id":"1003","user":"王五","amount":59.9}`},
		{"order-1002", `{"id":"1002","status":"shipped"}`},
	}

	fmt.Println("\n生产消息（按 Key 哈希分区）：")
	for _, o := range orders {
		msg := topic.Produce(o.key, o.value)
		fmt.Printf("  Key=%s → Partition %d, Offset %d\n", msg.Key, msg.Partition, msg.Offset)
	}

	fmt.Println("\n各分区消息分布：")
	for _, p := range topic.Partitions {
		fmt.Printf("  Partition %d: %d 条消息\n", p.ID, len(p.Messages))
		for _, msg := range p.Messages {
			fmt.Printf("    [Offset %d] Key=%s Value=%s\n", msg.Offset, msg.Key, msg.Value)
		}
	}

	fmt.Println("\n⚠️  关键点：")
	fmt.Println("  1. 相同 Key 的消息总是发送到同一分区（保证局部有序）")
	fmt.Println("  2. 不同 Key 的消息可能分布在不同分区（并行处理）")
	fmt.Println("  3. Kafka 只保证分区内有序，不保证跨分区有序")
}

// ============================================================
// 2. 模拟消费组
// ============================================================

func demoConsumerGroupConcept() {
	fmt.Println("\n--- 2. 消费组机制（内存模拟） ---")

	fmt.Println("\n[场景] Topic: orders（3 个分区），消费组 A（2 个消费者）")
	fmt.Println()

	// 模拟分区分配
	type Assignment struct {
		Consumer   string
		Partitions []int
	}

	// 2 个消费者消费 3 个分区
	assignments := []Assignment{
		{"Consumer-A1", []int{0, 1}},
		{"Consumer-A2", []int{2}},
	}

	for _, a := range assignments {
		fmt.Printf("  %s → 负责 Partition %v\n", a.Consumer, a.Partitions)
	}

	fmt.Println("\n[场景] 新增 Consumer-A3 加入消费组 → 触发 Rebalance")
	rebalanced := []Assignment{
		{"Consumer-A1", []int{0}},
		{"Consumer-A2", []int{1}},
		{"Consumer-A3", []int{2}},
	}
	for _, a := range rebalanced {
		fmt.Printf("  %s → 负责 Partition %v\n", a.Consumer, a.Partitions)
	}

	fmt.Println("\n[场景] 新增 Consumer-A4 → 消费者数 > 分区数")
	fmt.Println("  Consumer-A4 → 空闲（无分区可消费）")

	fmt.Println("\n⚠️  关键规则：")
	fmt.Println("  1. 一个分区在同一消费组内只能被一个消费者消费")
	fmt.Println("  2. 消费者数量 > 分区数量时，多余的消费者空闲")
	fmt.Println("  3. 不同消费组之间互不影响，各自维护 Offset")
}

// ============================================================
// 3. 模拟消息可靠性（acks 机制）
// ============================================================

func demoAcksConcept() {
	fmt.Println("\n--- 3. 消息可靠性 — acks 机制（内存模拟） ---")

	type AcksConfig struct {
		Acks        string
		Reliability string
		Performance string
		Description string
	}

	configs := []AcksConfig{
		{"acks=0", "最低", "最高", "不等待确认，发完即忘。可能丢消息"},
		{"acks=1", "中等", "中等", "Leader 写入即确认。Leader 宕机可能丢"},
		{"acks=all(-1)", "最高", "最低", "所有 ISR 副本写入确认。最安全"},
	}

	for _, c := range configs {
		fmt.Printf("\n  [%s] 可靠性: %s | 性能: %s\n", c.Acks, c.Reliability, c.Performance)
		fmt.Printf("    说明: %s\n", c.Description)
	}

	fmt.Println("\n  生产环境推荐配置：")
	fmt.Println("    acks=all + min.insync.replicas=2 + replication.factor=3")
	fmt.Println("    → 至少 2 个副本写入成功才确认，容忍 1 个 Broker 宕机")
}

// ============================================================
// 4. 模拟幂等生产者
// ============================================================

func demoIdempotentProducer() {
	fmt.Println("\n--- 4. 幂等生产者（内存模拟） ---")

	// 模拟 Broker 端去重
	type ProducerRecord struct {
		ProducerID int64
		SeqNum     int64
		Value      string
	}

	// Broker 记录每个 Producer 的最新 SeqNum
	brokerState := make(map[int64]int64) // ProducerID -> lastSeqNum

	records := []ProducerRecord{
		{1001, 1, "订单创建"},
		{1001, 2, "订单支付"},
		{1001, 2, "订单支付（重试）"}, // 重复消息
		{1001, 3, "订单发货"},
		{1001, 3, "订单发货（重试）"}, // 重复消息
	}

	fmt.Println("\n  Broker 端去重过程：")
	for _, r := range records {
		lastSeq, exists := brokerState[r.ProducerID]
		if exists && r.SeqNum <= lastSeq {
			fmt.Printf("    ProducerID=%d SeqNum=%d → ❌ 重复，丢弃: %s\n",
				r.ProducerID, r.SeqNum, r.Value)
		} else {
			brokerState[r.ProducerID] = r.SeqNum
			fmt.Printf("    ProducerID=%d SeqNum=%d → ✅ 写入: %s\n",
				r.ProducerID, r.SeqNum, r.Value)
		}
	}

	fmt.Println("\n  幂等生产者原理：")
	fmt.Println("    1. 每个 Producer 分配唯一 ProducerID")
	fmt.Println("    2. 每条消息携带递增 SequenceNumber")
	fmt.Println("    3. Broker 检测重复 SeqNum 并去重")
	fmt.Println("    4. 保证单分区内 Exactly-Once 语义")
}

// ============================================================
// Part B：连接真实 Kafka（需要 Docker）
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：连接真实 Kafka（sarama）")
	fmt.Println(strings.Repeat("=", 60))

	brokers := []string{"localhost:9092"}
	topic := "demo-orders"

	// 1. 同步生产者
	demoSyncProducer(brokers, topic)

	// 2. 消费组消费
	demoConsumerGroup(brokers, topic)
}

// demoSyncProducer 演示 sarama 同步生产者
func demoSyncProducer(brokers []string, topic string) {
	fmt.Println("\n--- 1. 同步生产者（sarama SyncProducer） ---")

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // acks=all
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		fmt.Printf("❌ 无法连接 Kafka: %v\n", err)
		fmt.Println("请先启动 Kafka: docker compose -f docker/docker-compose.mq.yml up -d kafka")
		return
	}
	defer producer.Close()
	fmt.Println("✅ Kafka 生产者连接成功")

	// 发送消息
	messages := []struct{ key, value string }{
		{"order-1001", `{"id":"1001","user":"张三","amount":99.9}`},
		{"order-1002", `{"id":"1002","user":"李四","amount":199.0}`},
		{"order-1003", `{"id":"1003","user":"王五","amount":59.9}`},
	}

	for _, m := range messages {
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(m.key),
			Value: sarama.StringEncoder(m.value),
		}
		partition, offset, err := producer.SendMessage(msg)
		if err != nil {
			fmt.Printf("  ❌ 发送失败: %v\n", err)
			continue
		}
		fmt.Printf("  ✅ 发送成功: key=%s partition=%d offset=%d\n", m.key, partition, offset)
	}
}

// ConsumerGroupHandler 实现 sarama.ConsumerGroupHandler 接口
type ConsumerGroupHandler struct {
	ready chan bool
}

func (h *ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("  📨 收到消息: topic=%s partition=%d offset=%d key=%s value=%s\n",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
		session.MarkMessage(msg, "") // 提交 Offset
	}
	return nil
}

// demoConsumerGroup 演示 sarama 消费组
func demoConsumerGroup(brokers []string, topic string) {
	fmt.Println("\n--- 2. 消费组消费（sarama ConsumerGroup） ---")

	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(brokers, "demo-group", config)
	if err != nil {
		fmt.Printf("❌ 创建消费组失败: %v\n", err)
		return
	}
	defer group.Close()
	fmt.Println("✅ 消费组创建成功，开始消费（5 秒后自动退出）...")

	handler := &ConsumerGroupHandler{ready: make(chan bool)}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := group.Consume(ctx, []string{topic}, handler); err != nil {
				if ctx.Err() != nil {
					return
				}
				fmt.Printf("  消费错误: %v\n", err)
				return
			}
			if ctx.Err() != nil {
				return
			}
			handler = &ConsumerGroupHandler{ready: make(chan bool)}
		}
	}()

	wg.Wait()
	fmt.Println("  消费组已退出")
}

// ============================================================
// main 入口
// ============================================================

func main() {
	// Part A：纯内存模拟，直接运行理解原理
	partA()

	// Part B：连接真实 Kafka，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		partB()
	}
}
