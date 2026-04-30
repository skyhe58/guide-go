// etcd 服务注册与发现 + 配置中心 — 完整示例
// 演示：Lease 租约 / Watch 机制 / KV 配置管理 / 服务注册与发现
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解 etcd 核心概念
// Part B：连接真实 etcd，需传入参数 'real'
//
// 运行方式：
//   go run ./etcd/              # Part A：内存模拟
//   go run ./etcd/ real         # Part B：连接真实 etcd
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.etcd.yml up -d
//   连接地址：localhost:2379，无需认证

package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// ============================================================
// Part A：纯内存模拟 etcd 核心概念
// ============================================================

// Lease 表示一个租约，包含 TTL 和过期时间
type Lease struct {
	ID        int64
	TTL       int64 // 秒
	ExpiresAt time.Time
	Keys      []string // 绑定的 key 列表
}

// WatchEvent 表示一个 Watch 事件
type WatchEvent struct {
	Type  string // "PUT" 或 "DELETE"
	Key   string
	Value string
}

// InMemoryEtcd 模拟 etcd 的核心功能：KV 存储 + Lease + Watch
type InMemoryEtcd struct {
	mu       sync.RWMutex
	store    map[string]string       // KV 存储
	leases   map[int64]*Lease        // 租约表
	watchers map[string][]chan WatchEvent // key 前缀 -> 事件通道
	revision int64                   // 全局递增版本号（模拟 MVCC）
	nextLease int64                  // 下一个 Lease ID
}

// NewInMemoryEtcd 创建内存模拟的 etcd 实例
func NewInMemoryEtcd() *InMemoryEtcd {
	e := &InMemoryEtcd{
		store:    make(map[string]string),
		leases:   make(map[int64]*Lease),
		watchers: make(map[string][]chan WatchEvent),
		nextLease: 1000,
	}
	// 启动 Lease 过期检查协程
	go e.leaseExpireLoop()
	return e
}

// Put 写入键值对，可选绑定 Lease
func (e *InMemoryEtcd) Put(key, value string, leaseID int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.store[key] = value
	e.revision++

	// 绑定 Lease
	if leaseID > 0 {
		if lease, ok := e.leases[leaseID]; ok {
			lease.Keys = append(lease.Keys, key)
		}
	}

	// 通知所有匹配的 Watcher
	e.notifyWatchers(WatchEvent{Type: "PUT", Key: key, Value: value})
	return nil
}

// Get 读取键值对，支持前缀查询
func (e *InMemoryEtcd) Get(key string, prefix bool) map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make(map[string]string)
	if prefix {
		for k, v := range e.store {
			if strings.HasPrefix(k, key) {
				result[k] = v
			}
		}
	} else {
		if v, ok := e.store[key]; ok {
			result[key] = v
		}
	}
	return result
}

// Delete 删除键值对
func (e *InMemoryEtcd) Delete(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.store[key]; ok {
		delete(e.store, key)
		e.revision++
		e.notifyWatchers(WatchEvent{Type: "DELETE", Key: key})
	}
}

// GrantLease 创建租约
func (e *InMemoryEtcd) GrantLease(ttlSeconds int64) int64 {
	e.mu.Lock()
	defer e.mu.Unlock()

	id := e.nextLease
	e.nextLease++
	e.leases[id] = &Lease{
		ID:        id,
		TTL:       ttlSeconds,
		ExpiresAt: time.Now().Add(time.Duration(ttlSeconds) * time.Second),
		Keys:      []string{},
	}
	return id
}

// KeepAlive 续约租约
func (e *InMemoryEtcd) KeepAlive(leaseID int64) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if lease, ok := e.leases[leaseID]; ok {
		lease.ExpiresAt = time.Now().Add(time.Duration(lease.TTL) * time.Second)
		return true
	}
	return false
}

// Watch 监听 key 前缀的变更事件，返回事件通道
func (e *InMemoryEtcd) Watch(prefix string) chan WatchEvent {
	e.mu.Lock()
	defer e.mu.Unlock()

	ch := make(chan WatchEvent, 100)
	e.watchers[prefix] = append(e.watchers[prefix], ch)
	return ch
}

// notifyWatchers 通知匹配的 Watcher（调用方需持有写锁）
func (e *InMemoryEtcd) notifyWatchers(event WatchEvent) {
	for prefix, channels := range e.watchers {
		if strings.HasPrefix(event.Key, prefix) {
			for _, ch := range channels {
				select {
				case ch <- event:
				default:
					// 通道满了，跳过（生产环境应处理背压）
				}
			}
		}
	}
}

// leaseExpireLoop 定期检查并清理过期的 Lease
func (e *InMemoryEtcd) leaseExpireLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		e.mu.Lock()
		now := time.Now()
		for id, lease := range e.leases {
			if now.After(lease.ExpiresAt) {
				// Lease 过期，删除关联的所有 key
				for _, key := range lease.Keys {
					if _, ok := e.store[key]; ok {
						delete(e.store, key)
						e.revision++
						e.notifyWatchers(WatchEvent{Type: "DELETE", Key: key})
					}
				}
				delete(e.leases, id)
			}
		}
		e.mu.Unlock()
	}
}

// ServiceInstance 表示一个服务实例
type ServiceInstance struct {
	Name    string
	Address string
	Port    int
	LeaseID int64
}

// ServiceRegistry 基于内存 etcd 的服务注册中心
type ServiceRegistry struct {
	etcd *InMemoryEtcd
}

// NewServiceRegistry 创建服务注册中心
func NewServiceRegistry(etcd *InMemoryEtcd) *ServiceRegistry {
	return &ServiceRegistry{etcd: etcd}
}

// Register 注册服务实例（绑定 Lease 实现自动摘除）
func (r *ServiceRegistry) Register(instance ServiceInstance) int64 {
	leaseID := r.etcd.GrantLease(3) // TTL=3s（演示用，生产环境建议 10-30s）
	key := fmt.Sprintf("/services/%s/%s:%d", instance.Name, instance.Address, instance.Port)
	value := fmt.Sprintf("%s:%d", instance.Address, instance.Port)
	r.etcd.Put(key, value, leaseID)
	return leaseID
}

// Discover 发现指定服务的所有实例
func (r *ServiceRegistry) Discover(serviceName string) []string {
	prefix := fmt.Sprintf("/services/%s/", serviceName)
	result := r.etcd.Get(prefix, true)

	instances := make([]string, 0, len(result))
	for _, v := range result {
		instances = append(instances, v)
	}
	sort.Strings(instances)
	return instances
}

// WatchService 监听服务实例变更
func (r *ServiceRegistry) WatchService(serviceName string) chan WatchEvent {
	prefix := fmt.Sprintf("/services/%s/", serviceName)
	return r.etcd.Watch(prefix)
}

// ConfigCenter 基于内存 etcd 的配置中心
type ConfigCenter struct {
	etcd  *InMemoryEtcd
	mu    sync.RWMutex
	cache map[string]string // 本地配置缓存
}

// NewConfigCenter 创建配置中心
func NewConfigCenter(etcd *InMemoryEtcd, appName string) *ConfigCenter {
	cc := &ConfigCenter{
		etcd:  etcd,
		cache: make(map[string]string),
	}

	// 加载初始配置
	prefix := fmt.Sprintf("/config/%s/", appName)
	configs := etcd.Get(prefix, true)
	for k, v := range configs {
		cc.cache[k] = v
	}

	// 启动 Watch 监听配置变更
	watchCh := etcd.Watch(prefix)
	go func() {
		for event := range watchCh {
			cc.mu.Lock()
			switch event.Type {
			case "PUT":
				cc.cache[event.Key] = event.Value
				fmt.Printf("  [配置中心] 配置更新: %s = %s\n", event.Key, event.Value)
			case "DELETE":
				delete(cc.cache, event.Key)
				fmt.Printf("  [配置中心] 配置删除: %s\n", event.Key)
			}
			cc.mu.Unlock()
		}
	}()

	return cc
}

// Get 读取配置（从本地缓存）
func (cc *ConfigCenter) Get(key string) (string, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	v, ok := cc.cache[key]
	return v, ok
}

// GetAll 获取所有配置
func (cc *ConfigCenter) GetAll() map[string]string {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	result := make(map[string]string, len(cc.cache))
	for k, v := range cc.cache {
		result[k] = v
	}
	return result
}

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：etcd 核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	etcd := NewInMemoryEtcd()

	// --- 1. 服务注册与发现 ---
	demoServiceDiscovery(etcd)

	// --- 2. Lease 租约与自动摘除 ---
	demoLeaseExpiry(etcd)

	// --- 3. 配置中心 ---
	demoConfigCenter(etcd)
}

// demoServiceDiscovery 演示服务注册与发现
func demoServiceDiscovery(etcd *InMemoryEtcd) {
	fmt.Println("\n--- 1. 服务注册与发现 ---")

	registry := NewServiceRegistry(etcd)

	// 启动 Watch 监听服务变更
	watchCh := registry.WatchService("api-gateway")
	go func() {
		for event := range watchCh {
			fmt.Printf("  [Watch] 事件: %s, Key: %s\n", event.Type, event.Key)
		}
	}()

	// 注册三个服务实例
	instances := []ServiceInstance{
		{Name: "api-gateway", Address: "192.168.1.10", Port: 8080},
		{Name: "api-gateway", Address: "192.168.1.11", Port: 8080},
		{Name: "api-gateway", Address: "192.168.1.12", Port: 8080},
	}

	for _, inst := range instances {
		leaseID := registry.Register(inst)
		fmt.Printf("  注册服务: %s -> %s:%d (LeaseID=%d)\n",
			inst.Name, inst.Address, inst.Port, leaseID)
	}

	time.Sleep(100 * time.Millisecond) // 等待 Watch 事件处理

	// 发现服务
	discovered := registry.Discover("api-gateway")
	fmt.Printf("  发现 %d 个实例: %v\n", len(discovered), discovered)
}

// demoLeaseExpiry 演示 Lease 过期自动摘除
func demoLeaseExpiry(etcd *InMemoryEtcd) {
	fmt.Println("\n--- 2. Lease 租约与自动摘除 ---")

	registry := NewServiceRegistry(etcd)

	// 注册一个短 TTL 的服务（模拟服务宕机）
	leaseID := registry.Register(ServiceInstance{
		Name: "user-service", Address: "10.0.0.1", Port: 9090,
	})
	fmt.Printf("  注册 user-service (LeaseID=%d, TTL=3s)\n", leaseID)

	// 注册另一个服务并持续续约（模拟正常服务）
	leaseID2 := registry.Register(ServiceInstance{
		Name: "user-service", Address: "10.0.0.2", Port: 9090,
	})
	fmt.Printf("  注册 user-service (LeaseID=%d, TTL=3s, 持续续约)\n", leaseID2)

	// 模拟 10.0.0.2 持续续约
	go func() {
		for i := 0; i < 5; i++ {
			time.Sleep(1 * time.Second)
			etcd.KeepAlive(leaseID2)
		}
	}()

	// 查看初始状态
	discovered := registry.Discover("user-service")
	fmt.Printf("  初始实例: %v\n", discovered)

	// 等待 10.0.0.1 的 Lease 过期（不续约，模拟宕机）
	fmt.Println("  等待 10.0.0.1 的 Lease 过期（3.5s）...")
	time.Sleep(3500 * time.Millisecond)

	// 再次查看
	discovered = registry.Discover("user-service")
	fmt.Printf("  Lease 过期后实例: %v\n", discovered)
	fmt.Println("  ✅ 10.0.0.1 已被自动摘除（Lease 过期），10.0.0.2 仍然存活（持续续约）")
}

// demoConfigCenter 演示配置中心功能
func demoConfigCenter(etcd *InMemoryEtcd) {
	fmt.Println("\n--- 3. 配置中心（KV 存储 + Watch 变更） ---")

	// 预置一些配置
	etcd.Put("/config/myapp/database/host", "localhost", 0)
	etcd.Put("/config/myapp/database/port", "5432", 0)
	etcd.Put("/config/myapp/feature/new-ui", "false", 0)
	etcd.Put("/config/myapp/rate-limit", "1000", 0)

	// 创建配置中心（会自动加载初始配置并启动 Watch）
	cc := NewConfigCenter(etcd, "myapp")

	// 读取配置
	if host, ok := cc.Get("/config/myapp/database/host"); ok {
		fmt.Printf("  数据库地址: %s\n", host)
	}
	if port, ok := cc.Get("/config/myapp/database/port"); ok {
		fmt.Printf("  数据库端口: %s\n", port)
	}

	// 模拟运维人员修改配置
	fmt.Println("\n  模拟运维人员修改配置...")
	etcd.Put("/config/myapp/feature/new-ui", "true", 0)
	etcd.Put("/config/myapp/rate-limit", "2000", 0)

	time.Sleep(100 * time.Millisecond) // 等待 Watch 事件处理

	// 验证配置已更新
	if val, ok := cc.Get("/config/myapp/feature/new-ui"); ok {
		fmt.Printf("  new-ui 功能开关: %s ✅\n", val)
	}
	if val, ok := cc.Get("/config/myapp/rate-limit"); ok {
		fmt.Printf("  限流阈值: %s ✅\n", val)
	}
}

// ============================================================
// Part B：连接真实 etcd
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：连接真实 etcd（localhost:2379）")
	fmt.Println(strings.Repeat("=", 60))

	// 创建 etcd 客户端
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Printf("❌ 连接 etcd 失败: %v\n", err)
		fmt.Println("请先启动 etcd: docker compose -f docker/docker-compose.etcd.yml up -d")
		return
	}
	defer cli.Close()
	fmt.Println("✅ 成功连接 etcd")

	ctx := context.Background()

	// --- 1. 基本 KV 操作 ---
	demoRealKV(ctx, cli)

	// --- 2. Lease 租约 ---
	demoRealLease(ctx, cli)

	// --- 3. Watch 监听 ---
	demoRealWatch(ctx, cli)

	// --- 4. 服务注册与发现 ---
	demoRealServiceDiscovery(ctx, cli)

	// 清理测试数据
	cli.Delete(ctx, "/demo/", clientv3.WithPrefix())
	cli.Delete(ctx, "/services/demo/", clientv3.WithPrefix())
	cli.Delete(ctx, "/config/demo/", clientv3.WithPrefix())
	fmt.Println("\n✅ 测试数据已清理")
}

// demoRealKV 演示真实 etcd KV 操作
func demoRealKV(ctx context.Context, cli *clientv3.Client) {
	fmt.Println("\n--- 1. KV 基本操作 ---")

	// Put 写入
	_, err := cli.Put(ctx, "/demo/greeting", "你好，etcd！")
	if err != nil {
		fmt.Printf("  Put 失败: %v\n", err)
		return
	}
	fmt.Println("  Put /demo/greeting = 你好，etcd！")

	// Get 读取
	resp, err := cli.Get(ctx, "/demo/greeting")
	if err != nil {
		fmt.Printf("  Get 失败: %v\n", err)
		return
	}
	for _, kv := range resp.Kvs {
		fmt.Printf("  Get %s = %s (Revision=%d)\n",
			string(kv.Key), string(kv.Value), kv.ModRevision)
	}

	// 前缀查询
	cli.Put(ctx, "/demo/config/host", "localhost")
	cli.Put(ctx, "/demo/config/port", "8080")
	cli.Put(ctx, "/demo/config/debug", "true")

	resp, _ = cli.Get(ctx, "/demo/config/", clientv3.WithPrefix())
	fmt.Printf("  前缀查询 /demo/config/ 返回 %d 个 key:\n", resp.Count)
	for _, kv := range resp.Kvs {
		fmt.Printf("    %s = %s\n", string(kv.Key), string(kv.Value))
	}
}

// demoRealLease 演示真实 etcd Lease 操作
func demoRealLease(ctx context.Context, cli *clientv3.Client) {
	fmt.Println("\n--- 2. Lease 租约 ---")

	// 创建 Lease（TTL=5s）
	leaseResp, err := cli.Grant(ctx, 5)
	if err != nil {
		fmt.Printf("  Grant 失败: %v\n", err)
		return
	}
	fmt.Printf("  创建 Lease: ID=%d, TTL=%ds\n", leaseResp.ID, leaseResp.TTL)

	// 绑定 Lease 写入 KV
	_, err = cli.Put(ctx, "/demo/ephemeral", "临时数据", clientv3.WithLease(leaseResp.ID))
	if err != nil {
		fmt.Printf("  Put with Lease 失败: %v\n", err)
		return
	}
	fmt.Println("  写入临时 KV（绑定 Lease）")

	// 验证 key 存在
	resp, _ := cli.Get(ctx, "/demo/ephemeral")
	fmt.Printf("  Lease 有效时: key 存在=%v\n", resp.Count > 0)

	// KeepAlive 续约演示
	keepAliveCh, err := cli.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		fmt.Printf("  KeepAlive 失败: %v\n", err)
		return
	}

	// 读取一次续约响应
	select {
	case ka := <-keepAliveCh:
		if ka != nil {
			fmt.Printf("  KeepAlive 续约成功: TTL=%ds\n", ka.TTL)
		}
	case <-time.After(2 * time.Second):
		fmt.Println("  KeepAlive 超时")
	}

	// 主动撤销 Lease
	cli.Revoke(ctx, leaseResp.ID)
	resp, _ = cli.Get(ctx, "/demo/ephemeral")
	fmt.Printf("  Lease 撤销后: key 存在=%v ✅\n", resp.Count > 0)
}

// demoRealWatch 演示真实 etcd Watch 操作
func demoRealWatch(ctx context.Context, cli *clientv3.Client) {
	fmt.Println("\n--- 3. Watch 监听 ---")

	watchCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 启动 Watch
	watchCh := cli.Watch(watchCtx, "/demo/watch/", clientv3.WithPrefix())

	// 在另一个 goroutine 中写入数据
	go func() {
		time.Sleep(200 * time.Millisecond)
		cli.Put(ctx, "/demo/watch/key1", "value1")
		time.Sleep(200 * time.Millisecond)
		cli.Put(ctx, "/demo/watch/key2", "value2")
		time.Sleep(200 * time.Millisecond)
		cli.Put(ctx, "/demo/watch/key1", "value1-updated")
		time.Sleep(200 * time.Millisecond)
		cli.Delete(ctx, "/demo/watch/key2")
	}()

	// 接收 Watch 事件
	eventCount := 0
	for resp := range watchCh {
		for _, ev := range resp.Events {
			eventType := "PUT"
			if ev.Type == clientv3.EventTypeDelete {
				eventType = "DELETE"
			}
			fmt.Printf("  [Watch] %s %s", eventType, string(ev.Kv.Key))
			if ev.Type != clientv3.EventTypeDelete {
				fmt.Printf(" = %s", string(ev.Kv.Value))
			}
			fmt.Println()
			eventCount++
		}
	}
	fmt.Printf("  共收到 %d 个 Watch 事件 ✅\n", eventCount)

	// 清理
	cli.Delete(ctx, "/demo/watch/", clientv3.WithPrefix())
}

// demoRealServiceDiscovery 演示真实 etcd 服务注册与发现
func demoRealServiceDiscovery(ctx context.Context, cli *clientv3.Client) {
	fmt.Println("\n--- 4. 服务注册与发现 ---")

	serviceName := "demo-api"
	servicePrefix := fmt.Sprintf("/services/demo/%s/", serviceName)

	// 注册多个服务实例
	instances := []struct {
		addr string
		port int
	}{
		{"192.168.1.10", 8080},
		{"192.168.1.11", 8080},
		{"192.168.1.12", 8080},
	}

	var leaseIDs []clientv3.LeaseID
	for _, inst := range instances {
		// 创建 Lease
		lease, _ := cli.Grant(ctx, 10)
		leaseIDs = append(leaseIDs, lease.ID)

		// 注册服务
		key := fmt.Sprintf("%s%s:%d", servicePrefix, inst.addr, inst.port)
		value := fmt.Sprintf("%s:%d", inst.addr, inst.port)
		cli.Put(ctx, key, value, clientv3.WithLease(lease.ID))
		fmt.Printf("  注册: %s -> %s (LeaseID=%d)\n", key, value, lease.ID)
	}

	// 发现服务
	resp, _ := cli.Get(ctx, servicePrefix, clientv3.WithPrefix())
	fmt.Printf("  发现 %d 个实例:\n", resp.Count)
	for _, kv := range resp.Kvs {
		fmt.Printf("    %s\n", string(kv.Value))
	}

	// 模拟负载均衡（随机选择）
	if resp.Count > 0 {
		idx := rand.Intn(int(resp.Count))
		fmt.Printf("  随机负载均衡选择: %s\n", string(resp.Kvs[idx].Value))
	}

	// 清理：撤销所有 Lease
	for _, id := range leaseIDs {
		cli.Revoke(ctx, id)
	}
}

func main() {
	partA()

	// Part B：连接真实 etcd，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		partB()
	}
}
