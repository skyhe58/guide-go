// IoT Core 设备 MQTT 通信模拟 — 完整示例
// 演示：设备注册 / 证书管理 / MQTT 发布订阅 / 设备影子 / 规则引擎
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例为纯 Go 实现，无需 Docker 或 AWS 账号
// 通过内存模拟完整的 IoT Core 设备通信工作流程
//
// 运行方式：
//   go run ./iot-core/

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ============================================================
// IoT Core 核心数据结构
// ============================================================

// DeviceCertificate 设备证书（模拟 X.509）
type DeviceCertificate struct {
	CertificateID  string    // 证书 ID
	CertificateARN string    // 证书 ARN
	Status         string    // ACTIVE / INACTIVE / REVOKED
	CreatedAt      time.Time // 创建时间
	ThingName      string    // 关联的设备名称
	Fingerprint    string    // 证书指纹（SHA256）
}

// Thing IoT 设备实体
type Thing struct {
	ThingName  string            // 设备名称（唯一标识）
	ThingARN   string            // 设备 ARN
	Attributes map[string]string // 设备属性
	CertID     string            // 关联的证书 ID
	CreatedAt  time.Time
	Connected  bool // 连接状态
}

// DeviceShadow 设备影子（云端状态副本）
type DeviceShadow struct {
	ThingName string
	State     ShadowState
	Metadata  ShadowMetadata
	Version   int
	Timestamp time.Time
}

// ShadowState 影子状态
type ShadowState struct {
	Desired  map[string]interface{} `json:"desired"`  // 期望状态（应用设置）
	Reported map[string]interface{} `json:"reported"` // 实际状态（设备上报）
	Delta    map[string]interface{} `json:"delta"`    // 差异（desired - reported）
}

// ShadowMetadata 影子元数据
type ShadowMetadata struct {
	Desired  map[string]int64 `json:"desired"`
	Reported map[string]int64 `json:"reported"`
}

// MQTTMessage MQTT 消息
type MQTTMessage struct {
	Topic     string
	Payload   []byte
	QoS       int // 0: 最多一次, 1: 至少一次
	Retain    bool
	Timestamp time.Time
	ClientID  string
}

// RuleAction 规则引擎动作
type RuleAction struct {
	Type   string // "s3", "sqs", "lambda", "dynamodb"
	Target string // 目标资源
}

// IoTRule 规则引擎规则
type IoTRule struct {
	Name    string
	SQL     string       // SQL 过滤表达式
	Actions []RuleAction // 触发的动作
	Enabled bool
}

// ============================================================
// IoT Core 服务模拟
// ============================================================

// IoTCoreService 模拟 AWS IoT Core 服务
type IoTCoreService struct {
	mu           sync.RWMutex
	things       map[string]*Thing            // thingName -> thing
	certificates map[string]*DeviceCertificate // certID -> cert
	shadows      map[string]*DeviceShadow     // thingName -> shadow
	rules        map[string]*IoTRule           // ruleName -> rule
	subscribers  map[string][]chan MQTTMessage // topic -> subscriber channels
	ruleLog      []string                     // 规则引擎执行日志
}

// NewIoTCoreService 创建 IoT Core 服务实例
func NewIoTCoreService() *IoTCoreService {
	return &IoTCoreService{
		things:       make(map[string]*Thing),
		certificates: make(map[string]*DeviceCertificate),
		shadows:      make(map[string]*DeviceShadow),
		rules:        make(map[string]*IoTRule),
		subscribers:  make(map[string][]chan MQTTMessage),
		ruleLog:      make([]string, 0),
	}
}

// --- 设备注册 ---

// RegisterThing 注册设备
func (s *IoTCoreService) RegisterThing(name string, attrs map[string]string) (*Thing, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.things[name]; exists {
		return nil, fmt.Errorf("thing %q already exists", name)
	}

	thing := &Thing{
		ThingName:  name,
		ThingARN:   fmt.Sprintf("arn:aws:iot:us-east-1:123456789012:thing/%s", name),
		Attributes: attrs,
		CreatedAt:  time.Now(),
	}
	s.things[name] = thing

	// 自动创建设备影子
	s.shadows[name] = &DeviceShadow{
		ThingName: name,
		State: ShadowState{
			Desired:  make(map[string]interface{}),
			Reported: make(map[string]interface{}),
			Delta:    make(map[string]interface{}),
		},
		Metadata: ShadowMetadata{
			Desired:  make(map[string]int64),
			Reported: make(map[string]int64),
		},
		Version:   1,
		Timestamp: time.Now(),
	}

	return thing, nil
}

// --- 证书管理 ---

// CreateCertificate 创建设备证书
func (s *IoTCoreService) CreateCertificate(thingName string) (*DeviceCertificate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	thing, exists := s.things[thingName]
	if !exists {
		return nil, fmt.Errorf("thing %q not found", thingName)
	}

	// 生成证书 ID 和指纹
	certBytes := make([]byte, 32)
	rand.Read(certBytes)
	certID := hex.EncodeToString(certBytes[:16])
	fingerprint := sha256.Sum256(certBytes)

	cert := &DeviceCertificate{
		CertificateID:  certID,
		CertificateARN: fmt.Sprintf("arn:aws:iot:us-east-1:123456789012:cert/%s", certID),
		Status:         "ACTIVE",
		CreatedAt:      time.Now(),
		ThingName:      thingName,
		Fingerprint:    hex.EncodeToString(fingerprint[:]),
	}

	s.certificates[certID] = cert
	thing.CertID = certID

	return cert, nil
}

// RevokeCertificate 吊销证书
func (s *IoTCoreService) RevokeCertificate(certID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cert, exists := s.certificates[certID]
	if !exists {
		return fmt.Errorf("certificate %q not found", certID)
	}
	cert.Status = "REVOKED"
	return nil
}

// --- MQTT 通信 ---

// Connect 设备连接（模拟 MQTT Connect）
func (s *IoTCoreService) Connect(thingName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	thing, exists := s.things[thingName]
	if !exists {
		return fmt.Errorf("thing %q not found", thingName)
	}

	// 验证证书
	if thing.CertID == "" {
		return fmt.Errorf("thing %q has no certificate", thingName)
	}
	cert, exists := s.certificates[thing.CertID]
	if !exists || cert.Status != "ACTIVE" {
		return fmt.Errorf("certificate for %q is not active", thingName)
	}

	thing.Connected = true
	return nil
}

// Disconnect 设备断开连接
func (s *IoTCoreService) Disconnect(thingName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if thing, exists := s.things[thingName]; exists {
		thing.Connected = false
	}
}

// Subscribe 订阅 MQTT 主题
func (s *IoTCoreService) Subscribe(topic string) chan MQTTMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan MQTTMessage, 100)
	s.subscribers[topic] = append(s.subscribers[topic], ch)
	return ch
}

// Publish 发布 MQTT 消息
func (s *IoTCoreService) Publish(clientID, topic string, payload []byte, qos int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := MQTTMessage{
		Topic:     topic,
		Payload:   payload,
		QoS:       qos,
		Timestamp: time.Now(),
		ClientID:  clientID,
	}

	// 分发给订阅者
	for subTopic, channels := range s.subscribers {
		if matchTopic(subTopic, topic) {
			for _, ch := range channels {
				select {
				case ch <- msg:
				default:
					// 通道满，跳过
				}
			}
		}
	}

	// 触发规则引擎
	s.evaluateRules(msg)
}

// matchTopic 简单的 MQTT 主题匹配（支持 + 和 # 通配符）
func matchTopic(pattern, topic string) bool {
	if pattern == topic {
		return true
	}
	if pattern == "#" {
		return true
	}
	patternParts := strings.Split(pattern, "/")
	topicParts := strings.Split(topic, "/")

	for i, pp := range patternParts {
		if pp == "#" {
			return true
		}
		if i >= len(topicParts) {
			return false
		}
		if pp != "+" && pp != topicParts[i] {
			return false
		}
	}
	return len(patternParts) == len(topicParts)
}

// --- 设备影子 ---

// UpdateShadowDesired 更新设备影子期望状态（应用端调用）
func (s *IoTCoreService) UpdateShadowDesired(thingName string, desired map[string]interface{}) (*DeviceShadow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	shadow, exists := s.shadows[thingName]
	if !exists {
		return nil, fmt.Errorf("shadow for %q not found", thingName)
	}

	now := time.Now().Unix()
	for k, v := range desired {
		shadow.State.Desired[k] = v
		shadow.Metadata.Desired[k] = now
	}

	// 计算 delta（desired 中有但 reported 中没有或不同的字段）
	shadow.State.Delta = make(map[string]interface{})
	for k, dv := range shadow.State.Desired {
		rv, exists := shadow.State.Reported[k]
		if !exists || rv != dv {
			shadow.State.Delta[k] = dv
		}
	}

	shadow.Version++
	shadow.Timestamp = time.Now()

	// 如果设备在线，推送 delta
	if thing, ok := s.things[thingName]; ok && thing.Connected {
		deltaPayload, _ := json.Marshal(shadow.State.Delta)
		go s.Publish("shadow-service", fmt.Sprintf("$aws/things/%s/shadow/update/delta", thingName), deltaPayload, 1)
	}

	return shadow, nil
}

// UpdateShadowReported 更新设备影子实际状态（设备端调用）
func (s *IoTCoreService) UpdateShadowReported(thingName string, reported map[string]interface{}) (*DeviceShadow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	shadow, exists := s.shadows[thingName]
	if !exists {
		return nil, fmt.Errorf("shadow for %q not found", thingName)
	}

	now := time.Now().Unix()
	for k, v := range reported {
		shadow.State.Reported[k] = v
		shadow.Metadata.Reported[k] = now
	}

	// 重新计算 delta
	shadow.State.Delta = make(map[string]interface{})
	for k, dv := range shadow.State.Desired {
		rv, exists := shadow.State.Reported[k]
		if !exists || rv != dv {
			shadow.State.Delta[k] = dv
		}
	}

	shadow.Version++
	shadow.Timestamp = time.Now()
	return shadow, nil
}

// GetShadow 获取设备影子
func (s *IoTCoreService) GetShadow(thingName string) (*DeviceShadow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shadow, exists := s.shadows[thingName]
	if !exists {
		return nil, fmt.Errorf("shadow for %q not found", thingName)
	}
	return shadow, nil
}

// --- 规则引擎 ---

// AddRule 添加规则
func (s *IoTCoreService) AddRule(rule *IoTRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[rule.Name] = rule
}

// evaluateRules 评估规则（简化版，基于主题前缀匹配）
func (s *IoTCoreService) evaluateRules(msg MQTTMessage) {
	for _, rule := range s.rules {
		if !rule.Enabled {
			continue
		}
		// 简化：检查 SQL 中的 topic 是否匹配
		if strings.Contains(rule.SQL, "'"+msg.Topic+"'") || strings.Contains(rule.SQL, "'+/telemetry'") {
			for _, action := range rule.Actions {
				logEntry := fmt.Sprintf("[规则 %s] 消息从 %s → %s:%s",
					rule.Name, msg.Topic, action.Type, action.Target)
				s.ruleLog = append(s.ruleLog, logEntry)
			}
		}
	}
}

// ============================================================
// 演示函数
// ============================================================

func main() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("IoT Core 设备 MQTT 通信模拟（纯 Go）")
	fmt.Println("=" + strings.Repeat("=", 59))

	iotService := NewIoTCoreService()

	// --- 1. 设备注册 ---
	demoDeviceRegistration(iotService)

	// --- 2. 证书管理 ---
	demoCertificateManagement(iotService)

	// --- 3. MQTT 发布/订阅 ---
	demoMQTTCommunication(iotService)

	// --- 4. 设备影子 ---
	demoDeviceShadow(iotService)

	// --- 5. 规则引擎 ---
	demoRuleEngine(iotService)
}

// demoDeviceRegistration 设备注册
func demoDeviceRegistration(iotService *IoTCoreService) {
	fmt.Println("\n--- 1. 设备注册 ---")

	devices := []struct {
		name  string
		attrs map[string]string
	}{
		{"sensor-temp-001", map[string]string{"type": "temperature", "location": "warehouse-A"}},
		{"sensor-temp-002", map[string]string{"type": "temperature", "location": "warehouse-B"}},
		{"camera-001", map[string]string{"type": "camera", "location": "entrance"}},
		{"gateway-001", map[string]string{"type": "gateway", "location": "factory"}},
	}

	for _, d := range devices {
		thing, err := iotService.RegisterThing(d.name, d.attrs)
		if err != nil {
			fmt.Printf("  注册失败: %v\n", err)
			continue
		}
		fmt.Printf("  注册设备: %s\n    ARN: %s\n    属性: %v\n",
			thing.ThingName, thing.ThingARN, thing.Attributes)
	}
}

// demoCertificateManagement 证书管理
func demoCertificateManagement(iotService *IoTCoreService) {
	fmt.Println("\n--- 2. 证书管理 ---")

	// 为每个设备创建证书
	deviceNames := []string{"sensor-temp-001", "sensor-temp-002", "camera-001", "gateway-001"}
	for _, name := range deviceNames {
		cert, err := iotService.CreateCertificate(name)
		if err != nil {
			fmt.Printf("  创建证书失败: %v\n", err)
			continue
		}
		fmt.Printf("  设备 %s 证书:\n    ID: %s\n    状态: %s\n    指纹: %s...\n",
			name, cert.CertificateID[:16]+"...", cert.Status, cert.Fingerprint[:32]+"...")
	}

	// 吊销一个证书
	fmt.Println("\n  模拟证书吊销（设备被盗/退役）:")
	// 获取 camera-001 的证书 ID
	iotService.mu.RLock()
	cameraCertID := iotService.things["camera-001"].CertID
	iotService.mu.RUnlock()

	iotService.RevokeCertificate(cameraCertID)
	fmt.Printf("  吊销 camera-001 证书: %s ✅\n", cameraCertID[:16]+"...")

	// 尝试用吊销的证书连接
	err := iotService.Connect("camera-001")
	if err != nil {
		fmt.Printf("  camera-001 连接失败: %v ✅（证书已吊销）\n", err)
	}
}

// demoMQTTCommunication MQTT 发布/订阅
func demoMQTTCommunication(iotService *IoTCoreService) {
	fmt.Println("\n--- 3. MQTT 发布/订阅 ---")

	// 连接设备
	iotService.Connect("sensor-temp-001")
	iotService.Connect("sensor-temp-002")
	iotService.Connect("gateway-001")
	fmt.Println("  设备连接: sensor-temp-001, sensor-temp-002, gateway-001 ✅")

	// 网关订阅所有传感器数据（使用通配符）
	gatewayCh := iotService.Subscribe("devices/+/telemetry")
	fmt.Println("  gateway-001 订阅: devices/+/telemetry（通配符）")

	// 后端订阅特定设备
	backendCh := iotService.Subscribe("devices/sensor-temp-001/telemetry")
	fmt.Println("  后端订阅: devices/sensor-temp-001/telemetry")

	// 传感器发布遥测数据
	fmt.Println("\n  传感器发布遥测数据:")
	telemetryData := []struct {
		device string
		temp   float64
		humid  float64
	}{
		{"sensor-temp-001", 25.5, 60.2},
		{"sensor-temp-002", 28.3, 55.8},
		{"sensor-temp-001", 26.1, 59.5},
	}

	for _, td := range telemetryData {
		payload, _ := json.Marshal(map[string]interface{}{
			"temperature": td.temp,
			"humidity":    td.humid,
			"timestamp":   time.Now().Unix(),
		})
		topic := fmt.Sprintf("devices/%s/telemetry", td.device)
		iotService.Publish(td.device, topic, payload, 1)
		fmt.Printf("    %s → %s: temp=%.1f°C, humid=%.1f%%\n",
			td.device, topic, td.temp, td.humid)
	}

	// 读取网关收到的消息
	fmt.Println("\n  网关收到的消息:")
	timeout := time.After(100 * time.Millisecond)
	count := 0
	for {
		select {
		case msg := <-gatewayCh:
			count++
			var data map[string]interface{}
			json.Unmarshal(msg.Payload, &data)
			fmt.Printf("    [%d] topic=%s, temp=%.1f°C (from %s)\n",
				count, msg.Topic, data["temperature"].(float64), msg.ClientID)
		case <-timeout:
			goto doneGateway
		}
	}
doneGateway:
	fmt.Printf("  网关共收到 %d 条消息 ✅\n", count)

	// 读取后端收到的消息
	fmt.Println("\n  后端收到的消息（仅 sensor-temp-001）:")
	timeout = time.After(100 * time.Millisecond)
	count = 0
	for {
		select {
		case msg := <-backendCh:
			count++
			fmt.Printf("    [%d] %s\n", count, string(msg.Payload))
		case <-timeout:
			goto doneBackend
		}
	}
doneBackend:
	fmt.Printf("  后端共收到 %d 条消息 ✅\n", count)
}

// demoDeviceShadow 设备影子
func demoDeviceShadow(iotService *IoTCoreService) {
	fmt.Println("\n--- 4. 设备影子 ---")

	thingName := "sensor-temp-001"

	// 应用端设置期望状态
	fmt.Println("  应用端设置期望状态:")
	shadow, _ := iotService.UpdateShadowDesired(thingName, map[string]interface{}{
		"reportInterval": 30,
		"ledStatus":      "on",
		"threshold":      35.0,
	})
	printShadow(shadow)

	// 设备上报实际状态（部分）
	fmt.Println("\n  设备上报实际状态（部分字段）:")
	shadow, _ = iotService.UpdateShadowReported(thingName, map[string]interface{}{
		"reportInterval": 30,
		"ledStatus":      "on",
		"firmwareVersion": "1.2.0",
	})
	printShadow(shadow)

	// 应用端更新期望状态（设备离线场景）
	fmt.Println("\n  模拟设备离线后应用更新期望状态:")
	iotService.Disconnect(thingName)
	fmt.Printf("  设备 %s 已断开连接\n", thingName)

	shadow, _ = iotService.UpdateShadowDesired(thingName, map[string]interface{}{
		"ledStatus": "off",
		"threshold": 30.0,
	})
	fmt.Println("  应用设置: ledStatus=off, threshold=30.0")
	printShadow(shadow)

	// 设备重新上线，获取 delta
	fmt.Println("\n  设备重新上线，获取 delta:")
	iotService.Connect(thingName)
	shadow, _ = iotService.GetShadow(thingName)
	if len(shadow.State.Delta) > 0 {
		deltaJSON, _ := json.MarshalIndent(shadow.State.Delta, "    ", "  ")
		fmt.Printf("    Delta（需要执行的操作）:\n    %s\n", string(deltaJSON))
	}

	// 设备执行操作并上报
	fmt.Println("\n  设备执行 delta 操作并上报:")
	shadow, _ = iotService.UpdateShadowReported(thingName, map[string]interface{}{
		"ledStatus": "off",
		"threshold": 30.0,
	})
	if len(shadow.State.Delta) == 0 {
		fmt.Println("    Delta 为空 — desired 与 reported 已同步 ✅")
	}
}

// printShadow 打印设备影子状态
func printShadow(shadow *DeviceShadow) {
	desiredJSON, _ := json.Marshal(shadow.State.Desired)
	reportedJSON, _ := json.Marshal(shadow.State.Reported)
	deltaJSON, _ := json.Marshal(shadow.State.Delta)
	fmt.Printf("    Desired:  %s\n", string(desiredJSON))
	fmt.Printf("    Reported: %s\n", string(reportedJSON))
	fmt.Printf("    Delta:    %s\n", string(deltaJSON))
	fmt.Printf("    Version:  %d\n", shadow.Version)
}

// demoRuleEngine 规则引擎
func demoRuleEngine(iotService *IoTCoreService) {
	fmt.Println("\n--- 5. 规则引擎 ---")

	// 添加规则
	rules := []*IoTRule{
		{
			Name:    "store-telemetry",
			SQL:     "SELECT * FROM 'devices/+/telemetry' WHERE temperature > 25",
			Actions: []RuleAction{{Type: "dynamodb", Target: "telemetry-table"}},
			Enabled: true,
		},
		{
			Name:    "alert-high-temp",
			SQL:     "SELECT * FROM 'devices/+/telemetry' WHERE temperature > 35",
			Actions: []RuleAction{{Type: "sqs", Target: "alert-queue"}, {Type: "lambda", Target: "send-notification"}},
			Enabled: true,
		},
		{
			Name:    "archive-to-s3",
			SQL:     "SELECT * FROM 'devices/+/telemetry'",
			Actions: []RuleAction{{Type: "s3", Target: "iot-data-bucket"}},
			Enabled: true,
		},
	}

	for _, rule := range rules {
		iotService.AddRule(rule)
		actionDescs := make([]string, len(rule.Actions))
		for i, a := range rule.Actions {
			actionDescs[i] = fmt.Sprintf("%s→%s", a.Type, a.Target)
		}
		fmt.Printf("  添加规则: %s\n    SQL: %s\n    动作: %s\n",
			rule.Name, rule.SQL, strings.Join(actionDescs, ", "))
	}

	// 模拟设备发送数据触发规则
	fmt.Println("\n  模拟设备发送数据触发规则:")
	testData := []struct {
		device string
		temp   float64
	}{
		{"sensor-temp-001", 26.5},
		{"sensor-temp-002", 38.0}, // 高温告警
	}

	for _, td := range testData {
		payload, _ := json.Marshal(map[string]interface{}{
			"temperature": td.temp,
			"timestamp":   time.Now().Unix(),
		})
		topic := fmt.Sprintf("devices/%s/telemetry", td.device)
		iotService.Publish(td.device, topic, payload, 1)
		fmt.Printf("    %s 发送: temp=%.1f°C\n", td.device, td.temp)
	}

	// 显示规则引擎日志
	fmt.Println("\n  规则引擎执行日志:")
	iotService.mu.RLock()
	for _, log := range iotService.ruleLog {
		fmt.Printf("    %s\n", log)
	}
	iotService.mu.RUnlock()

	fmt.Println("\n  规则引擎工作原理:")
	fmt.Println("    1. 设备发布 MQTT 消息到 IoT Core")
	fmt.Println("    2. 规则引擎用 SQL 过滤消息")
	fmt.Println("    3. 匹配的消息触发动作（转发到 S3/SQS/Lambda/DynamoDB）")
	fmt.Println("    4. 实现设备数据的自动化处理管道")
}
