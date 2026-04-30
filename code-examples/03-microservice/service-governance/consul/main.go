// Consul 服务注册与发现 — 完整示例
// 演示：服务注册 / 健康检查模拟 / KV 存储 / 服务发现
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解 Consul 核心概念
// Part B：连接真实 Consul，需传入参数 'real'
//
// 运行方式：
//   go run ./consul/              # Part A：内存模拟
//   go run ./consul/ real         # Part B：连接真实 Consul
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.consul.yml up -d
//   连接地址：localhost:8500，无需认证

package main

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

// ============================================================
// Part A：纯内存模拟 Consul 核心概念
// ============================================================

// HealthStatus 健康状态枚举
type HealthStatus string

const (
	StatusPassing  HealthStatus = "passing"
	StatusWarning  HealthStatus = "warning"
	StatusCritical HealthStatus = "critical"
)

// ServiceEntry 服务注册条目
type ServiceEntry struct {
	ID       string
	Name     string
	Address  string
	Port     int
	Tags     []string
	Health   HealthStatus
	CheckFn  func() HealthStatus // 健康检查函数
}

// KVEntry KV 存储条目
type KVEntry struct {
	Key         string
	Value       []byte
	ModifyIndex uint64
}

// InMemoryConsul 模拟 Consul 的核心功能
type InMemoryConsul struct {
	mu          sync.RWMutex
	services    map[string]*ServiceEntry // serviceID -> entry
	kvStore     map[string]*KVEntry      // key -> entry
	modifyIndex uint64
	stopCh      chan struct{}
}

// NewInMemoryConsul 创建内存模拟的 Consul 实例
func NewInMemoryConsul() *InMemoryConsul {
	c := &InMemoryConsul{
		services: make(map[string]*ServiceEntry),
		kvStore:  make(map[string]*KVEntry),
		stopCh:   make(chan struct{}),
	}
	// 启动健康检查循环
	go c.healthCheckLoop()
	return c
}

// RegisterService 注册服务
func (c *InMemoryConsul) RegisterService(entry *ServiceEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry.Health = StatusPassing
	c.services[entry.ID] = entry
}

// DeregisterService 注销服务
func (c *InMemoryConsul) DeregisterService(serviceID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.services, serviceID)
}

// DiscoverHealthy 发现健康的服务实例
func (c *InMemoryConsul) DiscoverHealthy(serviceName string) []*ServiceEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*ServiceEntry
	for _, svc := range c.services {
		if svc.Name == serviceName && svc.Health == StatusPassing {
			result = append(result, svc)
		}
	}
	// 按 ID 排序保证输出稳定
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// DiscoverAll 发现所有服务实例（含不健康的）
func (c *InMemoryConsul) DiscoverAll(serviceName string) []*ServiceEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*ServiceEntry
	for _, svc := range c.services {
		if svc.Name == serviceName {
			result = append(result, svc)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// KVPut 写入 KV
func (c *InMemoryConsul) KVPut(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.modifyIndex++
	c.kvStore[key] = &KVEntry{
		Key:         key,
		Value:       value,
		ModifyIndex: c.modifyIndex,
	}
}

// KVGet 读取 KV
func (c *InMemoryConsul) KVGet(key string) (*KVEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.kvStore[key]
	return entry, ok
}

// KVList 列出指定前缀的所有 KV
func (c *InMemoryConsul) KVList(prefix string) []*KVEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*KVEntry
	for k, v := range c.kvStore {
		if strings.HasPrefix(k, prefix) {
			result = append(result, v)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}

// KVDelete 删除 KV
func (c *InMemoryConsul) KVDelete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.kvStore, key)
}

// healthCheckLoop 定期执行健康检查
func (c *InMemoryConsul) healthCheckLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			for _, svc := range c.services {
				if svc.CheckFn != nil {
					svc.Health = svc.CheckFn()
				}
			}
			c.mu.Unlock()
		case <-c.stopCh:
			return
		}
	}
}

// Stop 停止 Consul 模拟
func (c *InMemoryConsul) Stop() {
	close(c.stopCh)
}

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：Consul 核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	consul := NewInMemoryConsul()
	defer consul.Stop()

	// --- 1. 服务注册与发现 ---
	demoServiceRegistration(consul)

	// --- 2. 健康检查模拟 ---
	demoHealthCheck(consul)

	// --- 3. KV 存储 ---
	demoKVStore(consul)
}

// demoServiceRegistration 演示服务注册与发现
func demoServiceRegistration(consul *InMemoryConsul) {
	fmt.Println("\n--- 1. 服务注册与发现 ---")

	// 注册多个服务实例
	services := []*ServiceEntry{
		{
			ID: "api-1", Name: "api-service",
			Address: "192.168.1.10", Port: 8080,
			Tags: []string{"v1", "primary"},
		},
		{
			ID: "api-2", Name: "api-service",
			Address: "192.168.1.11", Port: 8080,
			Tags: []string{"v1", "secondary"},
		},
		{
			ID: "api-3", Name: "api-service",
			Address: "192.168.1.12", Port: 8080,
			Tags: []string{"v2", "canary"},
		},
		{
			ID: "db-1", Name: "db-service",
			Address: "192.168.1.20", Port: 5432,
			Tags: []string{"primary"},
		},
	}

	for _, svc := range services {
		consul.RegisterService(svc)
		fmt.Printf("  注册服务: %s (%s:%d) tags=%v\n",
			svc.ID, svc.Address, svc.Port, svc.Tags)
	}

	// 发现 api-service
	healthy := consul.DiscoverHealthy("api-service")
	fmt.Printf("\n  发现 api-service 健康实例 (%d 个):\n", len(healthy))
	for _, svc := range healthy {
		fmt.Printf("    %s -> %s:%d [%s]\n",
			svc.ID, svc.Address, svc.Port, svc.Health)
	}

	// 模拟负载均衡（随机选择）
	if len(healthy) > 0 {
		chosen := healthy[rand.Intn(len(healthy))]
		fmt.Printf("  随机负载均衡选择: %s:%d\n", chosen.Address, chosen.Port)
	}

	// 注销一个服务
	consul.DeregisterService("api-3")
	fmt.Println("\n  注销 api-3（金丝雀实例）")
	healthy = consul.DiscoverHealthy("api-service")
	fmt.Printf("  注销后剩余 %d 个健康实例\n", len(healthy))
}

// demoHealthCheck 演示健康检查机制
func demoHealthCheck(consul *InMemoryConsul) {
	fmt.Println("\n--- 2. 健康检查模拟 ---")

	// 模拟一个会变为不健康的服务
	checkCount := 0
	consul.RegisterService(&ServiceEntry{
		ID: "flaky-1", Name: "flaky-service",
		Address: "10.0.0.1", Port: 9090,
		CheckFn: func() HealthStatus {
			checkCount++
			// 前 2 次检查通过，之后变为不健康（模拟服务故障）
			if checkCount <= 2 {
				return StatusPassing
			}
			return StatusCritical
		},
	})

	// 注册一个始终健康的服务
	consul.RegisterService(&ServiceEntry{
		ID: "stable-1", Name: "flaky-service",
		Address: "10.0.0.2", Port: 9090,
		CheckFn: func() HealthStatus {
			return StatusPassing
		},
	})

	fmt.Println("  注册 flaky-service: flaky-1（会变不健康）和 stable-1（始终健康）")

	// 初始状态
	all := consul.DiscoverAll("flaky-service")
	healthy := consul.DiscoverHealthy("flaky-service")
	fmt.Printf("  初始: 总实例=%d, 健康实例=%d\n", len(all), len(healthy))

	// 等待健康检查执行
	fmt.Println("  等待健康检查执行（3s）...")
	time.Sleep(3 * time.Second)

	// 检查后状态
	all = consul.DiscoverAll("flaky-service")
	healthy = consul.DiscoverHealthy("flaky-service")
	fmt.Printf("  健康检查后: 总实例=%d, 健康实例=%d\n", len(all), len(healthy))
	for _, svc := range all {
		fmt.Printf("    %s (%s:%d) -> 状态: %s\n",
			svc.ID, svc.Address, svc.Port, svc.Health)
	}
	fmt.Println("  ✅ flaky-1 被标记为 critical，不再出现在健康实例列表中")
}

// demoKVStore 演示 KV 存储
func demoKVStore(consul *InMemoryConsul) {
	fmt.Println("\n--- 3. KV 存储 ---")

	// 写入配置
	configs := map[string]string{
		"config/app/database/host":     "localhost",
		"config/app/database/port":     "5432",
		"config/app/database/name":     "mydb",
		"config/app/redis/host":        "localhost",
		"config/app/redis/port":        "6379",
		"config/app/feature/dark-mode": "true",
	}

	for k, v := range configs {
		consul.KVPut(k, []byte(v))
	}
	fmt.Printf("  写入 %d 个配置项\n", len(configs))

	// 读取单个配置
	if entry, ok := consul.KVGet("config/app/database/host"); ok {
		fmt.Printf("  读取 database/host = %s (ModifyIndex=%d)\n",
			string(entry.Value), entry.ModifyIndex)
	}

	// 列出前缀下所有配置
	dbConfigs := consul.KVList("config/app/database/")
	fmt.Printf("\n  数据库配置（前缀查询）:\n")
	for _, entry := range dbConfigs {
		fmt.Printf("    %s = %s\n", entry.Key, string(entry.Value))
	}

	// 更新配置
	consul.KVPut("config/app/feature/dark-mode", []byte("false"))
	if entry, ok := consul.KVGet("config/app/feature/dark-mode"); ok {
		fmt.Printf("\n  更新后 dark-mode = %s (ModifyIndex=%d)\n",
			string(entry.Value), entry.ModifyIndex)
	}

	// 删除配置
	consul.KVDelete("config/app/feature/dark-mode")
	if _, ok := consul.KVGet("config/app/feature/dark-mode"); !ok {
		fmt.Println("  删除 dark-mode 配置 ✅")
	}
}

// ============================================================
// Part B：连接真实 Consul
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：连接真实 Consul（localhost:8500）")
	fmt.Println(strings.Repeat("=", 60))

	// 创建 Consul 客户端
	client, err := api.NewClient(&api.Config{
		Address: "localhost:8500",
	})
	if err != nil {
		fmt.Printf("❌ 创建 Consul 客户端失败: %v\n", err)
		return
	}

	// 检查连接
	leader, err := client.Status().Leader()
	if err != nil {
		fmt.Printf("❌ 连接 Consul 失败: %v\n", err)
		fmt.Println("请先启动 Consul: docker compose -f docker/docker-compose.consul.yml up -d")
		return
	}
	fmt.Printf("✅ 成功连接 Consul (Leader: %s)\n", leader)

	// --- 1. 服务注册 ---
	demoRealServiceRegister(client)

	// --- 2. 服务发现 ---
	demoRealServiceDiscover(client)

	// --- 3. KV 操作 ---
	demoRealKV(client)

	// --- 4. 清理 ---
	demoRealCleanup(client)
}

// demoRealServiceRegister 真实 Consul 服务注册
func demoRealServiceRegister(client *api.Client) {
	fmt.Println("\n--- 1. 服务注册 ---")

	agent := client.Agent()

	// 注册多个服务实例
	registrations := []*api.AgentServiceRegistration{
		{
			ID:      "demo-api-1",
			Name:    "demo-api",
			Address: "192.168.1.10",
			Port:    8080,
			Tags:    []string{"v1", "primary"},
			Check: &api.AgentServiceCheck{
				TTL:                            "15s",
				DeregisterCriticalServiceAfter: "1m",
			},
		},
		{
			ID:      "demo-api-2",
			Name:    "demo-api",
			Address: "192.168.1.11",
			Port:    8080,
			Tags:    []string{"v1", "secondary"},
			Check: &api.AgentServiceCheck{
				TTL:                            "15s",
				DeregisterCriticalServiceAfter: "1m",
			},
		},
		{
			ID:      "demo-db-1",
			Name:    "demo-db",
			Address: "192.168.1.20",
			Port:    5432,
			Tags:    []string{"primary"},
			Check: &api.AgentServiceCheck{
				TTL:                            "15s",
				DeregisterCriticalServiceAfter: "1m",
			},
		},
	}

	for _, reg := range registrations {
		err := agent.ServiceRegister(reg)
		if err != nil {
			fmt.Printf("  注册 %s 失败: %v\n", reg.ID, err)
			continue
		}
		fmt.Printf("  注册: %s (%s:%d) tags=%v\n",
			reg.ID, reg.Address, reg.Port, reg.Tags)

		// TTL 健康检查：主动上报健康状态
		checkID := "service:" + reg.ID
		err = agent.PassTTL(checkID, "服务正常运行")
		if err != nil {
			fmt.Printf("  TTL 上报失败 (%s): %v\n", checkID, err)
		}
	}
}

// demoRealServiceDiscover 真实 Consul 服务发现
func demoRealServiceDiscover(client *api.Client) {
	fmt.Println("\n--- 2. 服务发现 ---")

	health := client.Health()

	// 发现健康的 demo-api 实例
	services, _, err := health.Service("demo-api", "", true, nil)
	if err != nil {
		fmt.Printf("  发现服务失败: %v\n", err)
		return
	}

	fmt.Printf("  发现 demo-api 健康实例 (%d 个):\n", len(services))
	for _, svc := range services {
		fmt.Printf("    %s -> %s:%d tags=%v\n",
			svc.Service.ID, svc.Service.Address, svc.Service.Port, svc.Service.Tags)
	}

	// 按 tag 过滤
	v1Services, _, _ := health.Service("demo-api", "v1", true, nil)
	fmt.Printf("  按 tag=v1 过滤: %d 个实例\n", len(v1Services))

	// 列出所有已注册的服务
	catalog := client.Catalog()
	serviceMap, _, _ := catalog.Services(nil)
	fmt.Println("\n  所有已注册的服务:")
	for name, tags := range serviceMap {
		fmt.Printf("    %s (tags: %v)\n", name, tags)
	}
}

// demoRealKV 真实 Consul KV 操作
func demoRealKV(client *api.Client) {
	fmt.Println("\n--- 3. KV 操作 ---")

	kv := client.KV()

	// 写入配置
	configs := map[string]string{
		"demo/config/database/host": "localhost",
		"demo/config/database/port": "5432",
		"demo/config/app/name":      "my-service",
		"demo/config/app/version":   "1.0.0",
	}

	for k, v := range configs {
		_, err := kv.Put(&api.KVPair{Key: k, Value: []byte(v)}, nil)
		if err != nil {
			fmt.Printf("  KV Put 失败: %v\n", err)
			continue
		}
	}
	fmt.Printf("  写入 %d 个配置项\n", len(configs))

	// 读取单个配置
	pair, _, err := kv.Get("demo/config/database/host", nil)
	if err != nil {
		fmt.Printf("  KV Get 失败: %v\n", err)
	} else if pair != nil {
		fmt.Printf("  读取 database/host = %s (ModifyIndex=%d)\n",
			string(pair.Value), pair.ModifyIndex)
	}

	// 列出前缀下所有配置
	pairs, _, err := kv.List("demo/config/database/", nil)
	if err != nil {
		fmt.Printf("  KV List 失败: %v\n", err)
	} else {
		fmt.Printf("  数据库配置（前缀查询，%d 项）:\n", len(pairs))
		for _, p := range pairs {
			fmt.Printf("    %s = %s\n", p.Key, string(p.Value))
		}
	}

	// CAS（Compare-And-Swap）原子更新
	pair, _, _ = kv.Get("demo/config/app/version", nil)
	if pair != nil {
		pair.Value = []byte("2.0.0")
		success, _, err := kv.CAS(pair, nil)
		if err != nil {
			fmt.Printf("  CAS 失败: %v\n", err)
		} else {
			fmt.Printf("  CAS 更新 app/version: success=%v\n", success)
		}
	}
}

// demoRealCleanup 清理测试数据
func demoRealCleanup(client *api.Client) {
	fmt.Println("\n--- 4. 清理测试数据 ---")

	agent := client.Agent()
	kv := client.KV()

	// 注销服务
	for _, id := range []string{"demo-api-1", "demo-api-2", "demo-db-1"} {
		agent.ServiceDeregister(id)
	}
	fmt.Println("  已注销所有 demo 服务")

	// 删除 KV
	kv.DeleteTree("demo/", nil)
	fmt.Println("  已删除所有 demo KV 数据")
	fmt.Println("  ✅ 清理完成")
}

func main() {
	partA()

	// Part B：连接真实 Consul，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		partB()
	}
}
