// MQTT 发布/订阅完整示例 — paho.mqtt.golang 客户端
// 演示：QoS 级别、遗嘱消息、保留消息、共享订阅、Topic 通配符
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解 MQTT 核心概念
// Part B：连接真实 EMQX Broker，需传入参数 'real'
//
// 运行方式：
//   go run ./mqtt/              # Part A：内存模拟
//   go run ./mqtt/ real         # Part B：连接真实 EMQX
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.mq.yml up -d emqx
//   连接地址：localhost:1883

package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ============================================================
// Part A：纯内存模拟 MQTT 核心概念
// ============================================================

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：MQTT 核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	// 1. 模拟 QoS 级别
	demoQoSLevels()

	// 2. 模拟遗嘱消息
	demoLastWill()

	// 3. 模拟保留消息
	demoRetainedMessage()

	// 4. 模拟共享订阅
	demoSharedSubscription()

	// 5. 模拟 Topic 通配符
	demoTopicWildcards()
}

// ============================================================
// 1. 模拟 QoS 级别
// ============================================================

func demoQoSLevels() {
	fmt.Println("\n--- 1. MQTT QoS 级别（内存模拟） ---")

	type QoSLevel struct {
		Level       int
		Name        string
		Guarantee   string
		Handshakes  int
		UseCase     string
	}

	levels := []QoSLevel{
		{0, "At Most Once", "最多一次，可能丢失", 1, "传感器周期上报（丢一条无所谓）"},
		{1, "At Least Once", "至少一次，可能重复", 2, "设备状态上报（重复可接受）"},
		{2, "Exactly Once", "精确一次，不丢不重", 4, "计费/支付指令（不能丢也不能重）"},
	}

	for _, l := range levels {
		fmt.Printf("\n  [QoS %d] %s\n", l.Level, l.Name)
		fmt.Printf("    保证: %s\n", l.Guarantee)
		fmt.Printf("    握手次数: %d\n", l.Handshakes)
		fmt.Printf("    适用场景: %s\n", l.UseCase)
	}

	fmt.Println("\n  QoS 消息传递流程：")
	fmt.Println("    QoS 0: Client → PUBLISH → Broker（发完即忘）")
	fmt.Println("    QoS 1: Client → PUBLISH → Broker → PUBACK → Client")
	fmt.Println("    QoS 2: Client → PUBLISH → Broker → PUBREC → Client → PUBREL → Broker → PUBCOMP → Client")

	fmt.Println("\n  ⚠️  最终 QoS = min(发布者 QoS, 订阅者 QoS)")
	fmt.Println("    发布者 QoS 2 + 订阅者 QoS 1 → 实际 QoS 1")
}

// ============================================================
// 2. 模拟遗嘱消息
// ============================================================

func demoLastWill() {
	fmt.Println("\n--- 2. 遗嘱消息 Last Will（内存模拟） ---")

	type Device struct {
		ClientID  string
		WillTopic string
		WillMsg   string
		Online    bool
	}

	device := Device{
		ClientID:  "sensor-001",
		WillTopic: "device/sensor-001/status",
		WillMsg:   `{"status":"offline","time":"` + time.Now().Format(time.RFC3339) + `"}`,
		Online:    true,
	}

	fmt.Println("\n  1. 设备连接时设置遗嘱消息：")
	fmt.Printf("     ClientID: %s\n", device.ClientID)
	fmt.Printf("     Will Topic: %s\n", device.WillTopic)
	fmt.Printf("     Will Payload: %s\n", device.WillMsg)
	fmt.Printf("     Will QoS: 1\n")
	fmt.Printf("     Will Retain: true\n")

	fmt.Println("\n  2. 设备正常工作中...")
	device.Online = true
	fmt.Printf("     设备状态: 在线 ✅\n")

	fmt.Println("\n  3. 设备异常断开（网络中断/断电）...")
	device.Online = false
	fmt.Printf("     设备状态: 离线 ❌\n")
	fmt.Printf("     Broker 自动发布遗嘱消息 → %s\n", device.WillTopic)
	fmt.Printf("     订阅者收到: %s\n", device.WillMsg)

	fmt.Println("\n  ⚠️  遗嘱消息触发条件：")
	fmt.Println("    ✅ 网络连接断开（TCP 超时）")
	fmt.Println("    ✅ 客户端崩溃")
	fmt.Println("    ❌ 客户端正常调用 Disconnect()（不触发遗嘱）")
}

// ============================================================
// 3. 模拟保留消息
// ============================================================

func demoRetainedMessage() {
	fmt.Println("\n--- 3. 保留消息 Retained Message（内存模拟） ---")

	// 模拟 Broker 的保留消息存储
	retainedStore := make(map[string]string)

	fmt.Println("\n  1. 设备发布保留消息：")
	retainedStore["device/sensor-001/status"] = `{"status":"online","temp":25.6}`
	fmt.Printf("     PUBLISH topic=device/sensor-001/status retain=true\n")
	fmt.Printf("     Payload: %s\n", retainedStore["device/sensor-001/status"])

	fmt.Println("\n  2. 新订阅者订阅该 Topic：")
	fmt.Printf("     SUBSCRIBE topic=device/sensor-001/status\n")
	if msg, ok := retainedStore["device/sensor-001/status"]; ok {
		fmt.Printf("     → 立即收到保留消息: %s\n", msg)
	}

	fmt.Println("\n  3. 更新保留消息：")
	retainedStore["device/sensor-001/status"] = `{"status":"online","temp":26.1}`
	fmt.Printf("     新的保留消息: %s\n", retainedStore["device/sensor-001/status"])

	fmt.Println("\n  4. 清除保留消息（发送空 Payload + retain=true）：")
	delete(retainedStore, "device/sensor-001/status")
	fmt.Println("     保留消息已清除")

	fmt.Println("\n  ⚠️  保留消息 vs 遗嘱消息：")
	fmt.Println("    保留消息: 主动发布，Broker 保存最后一条，新订阅者立即收到")
	fmt.Println("    遗嘱消息: 连接时预设，异常断开时 Broker 自动发布")
	fmt.Println("    两者可结合: 上线发保留消息 online，遗嘱设为 offline")
}

// ============================================================
// 4. 模拟共享订阅
// ============================================================

func demoSharedSubscription() {
	fmt.Println("\n--- 4. 共享订阅 Shared Subscription（内存模拟） ---")

	fmt.Println("\n  MQTT 5.0 共享订阅格式: $share/{group}/{topic}")
	fmt.Println()

	workers := []string{"Worker-1", "Worker-2", "Worker-3"}
	messages := []string{
		`{"sensor":"temp-001","value":25.6}`,
		`{"sensor":"temp-002","value":26.1}`,
		`{"sensor":"temp-003","value":24.8}`,
		`{"sensor":"humidity-001","value":65.2}`,
		`{"sensor":"humidity-002","value":70.1}`,
		`{"sensor":"temp-004","value":27.3}`,
	}

	fmt.Println("  订阅: $share/data-processors/sensor/+/data")
	fmt.Println("  每条消息只投递给组内一个 Worker（轮询）：")
	fmt.Println()

	for i, msg := range messages {
		worker := workers[i%len(workers)]
		fmt.Printf("    消息 %d → %s: %s\n", i+1, worker, msg)
	}

	fmt.Println("\n  ⚠️  共享订阅 vs 普通订阅：")
	fmt.Println("    普通订阅: 所有订阅者都收到（广播）")
	fmt.Println("    共享订阅: 组内只有一个收到（负载均衡）")
	fmt.Println("    类似 Kafka 消费组 / NATS Queue Group")
}

// ============================================================
// 5. 模拟 Topic 通配符
// ============================================================

func demoTopicWildcards() {
	fmt.Println("\n--- 5. Topic 通配符（内存模拟） ---")

	type MatchTest struct {
		Pattern string
		Topic   string
		Match   bool
	}

	tests := []MatchTest{
		{"sensor/+/temp", "sensor/room1/temp", true},
		{"sensor/+/temp", "sensor/room1/humidity", false},
		{"sensor/+/temp", "sensor/floor1/room1/temp", false},
		{"sensor/#", "sensor/room1/temp", true},
		{"sensor/#", "sensor/floor1/room1/temp", true},
		{"device/+/cmd", "device/switch-001/cmd", true},
	}

	fmt.Println()
	for _, t := range tests {
		result := "✅ 匹配"
		if !t.Match {
			result = "❌ 不匹配"
		}
		fmt.Printf("  Pattern: %-25s Topic: %-30s → %s\n", t.Pattern, t.Topic, result)
	}

	fmt.Println("\n  + — 匹配单层（不跨越 /）")
	fmt.Println("  # — 匹配多层（只能在末尾）")
	fmt.Println("  注意：MQTT 用 / 分隔，NATS 用 . 分隔")
}

// ============================================================
// Part B：连接真实 EMQX Broker（需要 Docker）
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：连接真实 EMQX（paho.mqtt.golang）")
	fmt.Println(strings.Repeat("=", 60))

	// 1. 基本发布/订阅
	demoRealPubSub()

	// 2. QoS 级别演示
	demoRealQoS()

	// 3. 遗嘱消息演示
	demoRealLastWill()
}

// createClient 创建 MQTT 客户端的辅助函数
func createClient(clientID string, opts ...func(*mqtt.ClientOptions)) mqtt.Client {
	options := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID(clientID).
		SetConnectTimeout(5 * time.Second)

	for _, opt := range opts {
		opt(options)
	}

	client := mqtt.NewClient(options)
	return client
}

// demoRealPubSub 演示真实的发布/订阅
func demoRealPubSub() {
	fmt.Println("\n--- 1. 发布/订阅 ---")

	// 创建订阅者
	subscriber := createClient("demo-subscriber-001")
	if token := subscriber.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("❌ 订阅者连接失败: %v\n", token.Error())
		fmt.Println("请先启动 EMQX: docker compose -f docker/docker-compose.mq.yml up -d emqx")
		return
	}
	defer subscriber.Disconnect(250)
	fmt.Println("✅ 订阅者连接成功")

	var wg sync.WaitGroup
	wg.Add(3)

	// 订阅 Topic（使用通配符）
	subscriber.Subscribe("demo/sensor/#", 1, func(c mqtt.Client, msg mqtt.Message) {
		fmt.Printf("  📨 收到: topic=%s qos=%d payload=%s\n",
			msg.Topic(), msg.Qos(), string(msg.Payload()))
		wg.Done()
	})

	time.Sleep(500 * time.Millisecond)

	// 创建发布者
	publisher := createClient("demo-publisher-001")
	if token := publisher.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("❌ 发布者连接失败: %v\n", token.Error())
		return
	}
	defer publisher.Disconnect(250)
	fmt.Println("✅ 发布者连接成功")

	// 发布消息
	messages := []struct {
		topic   string
		payload string
	}{
		{"demo/sensor/temp", `{"device":"temp-001","value":25.6}`},
		{"demo/sensor/humidity", `{"device":"hum-001","value":65.2}`},
		{"demo/sensor/pressure", `{"device":"press-001","value":1013.25}`},
	}

	for _, m := range messages {
		token := publisher.Publish(m.topic, 1, false, m.payload)
		token.Wait()
		fmt.Printf("  ✅ 发布: topic=%s\n", m.topic)
	}

	// 等待消息接收
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		fmt.Println("  ⚠️ 部分消息未收到（超时）")
	}
}

// demoRealQoS 演示不同 QoS 级别
func demoRealQoS() {
	fmt.Println("\n--- 2. QoS 级别演示 ---")

	client := createClient("demo-qos-client")
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("  连接失败: %v\n", token.Error())
		return
	}
	defer client.Disconnect(250)

	// 分别用不同 QoS 发布
	qosLevels := []byte{0, 1, 2}
	for _, qos := range qosLevels {
		topic := fmt.Sprintf("demo/qos/%d", qos)
		payload := fmt.Sprintf(`{"qos":%d,"msg":"QoS %d 测试消息"}`, qos, qos)
		token := client.Publish(topic, qos, false, payload)
		token.Wait()
		if token.Error() != nil {
			fmt.Printf("  ❌ QoS %d 发布失败: %v\n", qos, token.Error())
		} else {
			fmt.Printf("  ✅ QoS %d 发布成功: topic=%s\n", qos, topic)
		}
	}
}

// demoRealLastWill 演示遗嘱消息
func demoRealLastWill() {
	fmt.Println("\n--- 3. 遗嘱消息演示 ---")

	// 创建监听者（订阅设备状态）
	monitor := createClient("demo-monitor-001")
	if token := monitor.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("  监听者连接失败: %v\n", token.Error())
		return
	}
	defer monitor.Disconnect(250)

	var received sync.WaitGroup
	received.Add(1)

	monitor.Subscribe("demo/device/+/status", 1, func(c mqtt.Client, msg mqtt.Message) {
		fmt.Printf("  📨 设备状态变更: topic=%s payload=%s retain=%v\n",
			msg.Topic(), string(msg.Payload()), msg.Retained())
		received.Done()
	})

	time.Sleep(500 * time.Millisecond)

	// 创建设备客户端（设置遗嘱消息）
	device := createClient("demo-device-sensor-001", func(opts *mqtt.ClientOptions) {
		opts.SetWill(
			"demo/device/sensor-001/status",
			`{"status":"offline","reason":"unexpected"}`,
			1,    // QoS 1
			true, // Retain
		)
		opts.SetKeepAlive(2 * time.Second)
	})

	if token := device.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("  设备连接失败: %v\n", token.Error())
		return
	}
	fmt.Println("  ✅ 设备 sensor-001 已连接（遗嘱消息已设置）")

	// 发布上线状态（保留消息）
	device.Publish("demo/device/sensor-001/status",
		1, true, `{"status":"online"}`)
	fmt.Println("  ✅ 发布上线状态（保留消息）")

	// 等待状态消息
	done := make(chan struct{})
	go func() {
		received.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
	}

	// 正常断开（不触发遗嘱）
	device.Disconnect(250)
	fmt.Println("  设备正常断开（遗嘱消息不会触发）")
	fmt.Println("  提示：如果强制断开（不调用 Disconnect），Broker 会发布遗嘱消息")

	// 清理保留消息
	cleanClient := createClient("demo-cleanup")
	if token := cleanClient.Connect(); token.Wait() && token.Error() == nil {
		cleanClient.Publish("demo/device/sensor-001/status", 1, true, "")
		cleanClient.Disconnect(250)
	}
}

// ============================================================
// main 入口
// ============================================================

func main() {
	// Part A：纯内存模拟，直接运行理解原理
	partA()

	// Part B：连接真实 EMQX，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		partB()
	}
}
