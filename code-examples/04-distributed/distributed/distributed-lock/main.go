// 分布式锁 — Redis + etcd 完整实现
// 演示：内存模拟分布式锁（TTL/Owner 验证/Lua 原子解锁）+ 真实 Redis SET NX EX
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解分布式锁核心原理
// Part B：连接真实 Redis，需传入参数 'real'
//
// 运行方式：
//   go run ./distributed-lock/              # Part A：内存模拟
//   go run ./distributed-lock/ real         # Part B：连接真实 Redis
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.yml up -d redis
//   连接地址：localhost:6379，无密码

package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// ============================================================
// Part A：纯内存模拟分布式锁
// 模拟 Redis SET NX EX + Lua 原子解锁的核心逻辑
// ============================================================

// ErrLockNotHeld 尝试释放未持有的锁
var ErrLockNotHeld = errors.New("lock not held by this owner")

// ErrLockAcquireFailed 获取锁失败
var ErrLockAcquireFailed = errors.New("failed to acquire lock")

// MemoryLockStore 内存锁存储（模拟 Redis）
// 模拟 Redis 的 SET NX EX 语义和 Lua 脚本原子解锁
type MemoryLockStore struct {
	mu    sync.Mutex
	locks map[string]*lockEntry
}

// lockEntry 锁条目
type lockEntry struct {
	owner    string    // 锁持有者标识（模拟 Redis value）
	expireAt time.Time // 过期时间（模拟 Redis EX）
}

// NewMemoryLockStore 创建内存锁存储
func NewMemoryLockStore() *MemoryLockStore {
	store := &MemoryLockStore{
		locks: make(map[string]*lockEntry),
	}
	// 启动过期清理 goroutine（模拟 Redis 过期机制）
	go store.cleanupExpired()
	return store
}

// cleanupExpired 定期清理过期锁
func (s *MemoryLockStore) cleanupExpired() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, entry := range s.locks {
			if now.After(entry.expireAt) {
				delete(s.locks, key)
			}
		}
		s.mu.Unlock()
	}
}

// SetNXEX 模拟 Redis SET key value NX EX seconds
// NX: 只在 key 不存在时设置（互斥性）
// EX: 设置过期时间（防死锁）
// 返回是否设置成功
func (s *MemoryLockStore) SetNXEX(key, value string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查 key 是否存在且未过期（模拟 NX 语义）
	if entry, exists := s.locks[key]; exists {
		if time.Now().Before(entry.expireAt) {
			return false // key 存在且未过期，设置失败
		}
		// key 已过期，删除后重新设置
		delete(s.locks, key)
	}

	// 设置锁（模拟 SET NX EX）
	s.locks[key] = &lockEntry{
		owner:    value,
		expireAt: time.Now().Add(ttl),
	}
	return true
}

// LuaUnlock 模拟 Redis Lua 脚本原子解锁
// Lua 脚本逻辑：
//
//	if redis.call("GET", KEYS[1]) == ARGV[1] then
//	    return redis.call("DEL", KEYS[1])
//	else
//	    return 0
//	end
//
// 关键：检查 owner 和删除是原子操作，防止误删其他客户端的锁
func (s *MemoryLockStore) LuaUnlock(key, owner string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.locks[key]
	if !exists {
		return false
	}

	// 原子性：检查 owner 并删除（模拟 Lua 脚本）
	if entry.owner == owner {
		delete(s.locks, key)
		return true
	}
	return false // owner 不匹配，拒绝解锁
}

// DistributedLock 分布式锁客户端
type DistributedLock struct {
	store *MemoryLockStore
	key   string
	owner string
	ttl   time.Duration
}

// NewDistributedLock 创建分布式锁
func NewDistributedLock(store *MemoryLockStore, key string, ttl time.Duration) *DistributedLock {
	return &DistributedLock{
		store: store,
		key:   key,
		owner: fmt.Sprintf("owner-%d-%d", rand.Intn(10000), time.Now().UnixNano()),
		ttl:   ttl,
	}
}

// Lock 获取锁（带重试）
func (dl *DistributedLock) Lock(ctx context.Context) error {
	// 重试策略：最多重试 10 次，每次间隔 50-100ms
	for i := 0; i < 10; i++ {
		if dl.store.SetNXEX(dl.key, dl.owner, dl.ttl) {
			return nil // 加锁成功
		}

		// 等待重试（带退避）
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(50+rand.Intn(50)) * time.Millisecond):
			// 继续重试
		}
	}
	return ErrLockAcquireFailed
}

// Unlock 释放锁（Lua 脚本原子操作）
func (dl *DistributedLock) Unlock() error {
	if !dl.store.LuaUnlock(dl.key, dl.owner) {
		return ErrLockNotHeld
	}
	return nil
}

// TryLock 尝试获取锁（不重试）
func (dl *DistributedLock) TryLock() bool {
	return dl.store.SetNXEX(dl.key, dl.owner, dl.ttl)
}

// ============================================================
// Part B：真实 Redis 分布式锁
// ============================================================

// RedisDistributedLock 基于 Redis 的分布式锁
type RedisDistributedLock struct {
	client *redis.Client
	key    string
	owner  string
	ttl    time.Duration
}

// NewRedisDistributedLock 创建 Redis 分布式锁
func NewRedisDistributedLock(client *redis.Client, key string, ttl time.Duration) *RedisDistributedLock {
	return &RedisDistributedLock{
		client: client,
		key:    key,
		owner:  fmt.Sprintf("owner-%d-%d", rand.Intn(10000), time.Now().UnixNano()),
		ttl:    ttl,
	}
}

// Lock 获取锁：SET key value NX EX seconds
func (rl *RedisDistributedLock) Lock(ctx context.Context) error {
	for i := 0; i < 10; i++ {
		// SET key value NX EX seconds
		ok, err := rl.client.SetNX(ctx, rl.key, rl.owner, rl.ttl).Result()
		if err != nil {
			return fmt.Errorf("redis SetNX failed: %w", err)
		}
		if ok {
			return nil // 加锁成功
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(50+rand.Intn(50)) * time.Millisecond):
		}
	}
	return ErrLockAcquireFailed
}

// Unlock 释放锁：Lua 脚本原子解锁
// 关键：必须用 Lua 脚本保证"检查 owner + 删除"的原子性
// 如果先 GET 再 DEL，两步之间可能有其他客户端获取了锁，导致误删
func (rl *RedisDistributedLock) Unlock(ctx context.Context) error {
	// Lua 脚本：检查 owner 匹配后才删除
	luaScript := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	result, err := rl.client.Eval(ctx, luaScript, []string{rl.key}, rl.owner).Int64()
	if err != nil {
		return fmt.Errorf("redis Eval failed: %w", err)
	}
	if result == 0 {
		return ErrLockNotHeld
	}
	return nil
}

// ============================================================
// 演示
// ============================================================

func main() {
	fmt.Println("=== 分布式锁演示 ===")
	fmt.Println()

	// Part A：内存模拟
	partA()

	// Part B：连接真实 Redis（需传入参数 'real'）
	if len(os.Args) > 1 && os.Args[1] == "real" {
		partB()
	}
}

// partA 内存模拟分布式锁
func partA() {
	fmt.Println("=== Part A：内存模拟分布式锁 ===")
	fmt.Println()

	store := NewMemoryLockStore()

	// 场景一：基本加锁/解锁
	demoBasicLock(store)

	// 场景二：互斥性验证
	demoMutualExclusion(store)

	// 场景三：TTL 过期自动释放
	demoTTLExpiry(store)

	// 场景四：并发竞争
	demoConcurrentCompetition(store)
}

// demoBasicLock 基本加锁/解锁流程
func demoBasicLock(store *MemoryLockStore) {
	fmt.Println("--- 场景一：基本加锁/解锁 ---")

	lock := NewDistributedLock(store, "resource:order:123", 30*time.Second)

	// 加锁
	ctx := context.Background()
	err := lock.Lock(ctx)
	fmt.Printf("  加锁: err=%v, owner=%s\n", err, lock.owner)

	// 执行业务逻辑
	fmt.Println("  执行业务逻辑...")

	// 解锁
	err = lock.Unlock()
	fmt.Printf("  解锁: err=%v\n", err)

	// 尝试重复解锁（应该失败）
	err = lock.Unlock()
	fmt.Printf("  重复解锁: err=%v（预期：lock not held）\n", err)
	fmt.Println()
}

// demoMutualExclusion 互斥性验证
func demoMutualExclusion(store *MemoryLockStore) {
	fmt.Println("--- 场景二：互斥性验证 ---")

	lockA := NewDistributedLock(store, "resource:stock:456", 30*time.Second)
	lockB := NewDistributedLock(store, "resource:stock:456", 30*time.Second)

	// A 先加锁
	resultA := lockA.TryLock()
	fmt.Printf("  Client A 加锁: %v\n", resultA)

	// B 尝试加锁（应该失败）
	resultB := lockB.TryLock()
	fmt.Printf("  Client B 加锁: %v（预期：false，锁被 A 持有）\n", resultB)

	// A 解锁
	_ = lockA.Unlock()
	fmt.Println("  Client A 解锁")

	// B 再次尝试（应该成功）
	resultB = lockB.TryLock()
	fmt.Printf("  Client B 再次加锁: %v（预期：true）\n", resultB)
	_ = lockB.Unlock()
	fmt.Println()
}

// demoTTLExpiry TTL 过期自动释放
func demoTTLExpiry(store *MemoryLockStore) {
	fmt.Println("--- 场景三：TTL 过期自动释放（防死锁）---")

	// 设置短 TTL（200ms）
	lockA := NewDistributedLock(store, "resource:payment:789", 200*time.Millisecond)
	lockB := NewDistributedLock(store, "resource:payment:789", 30*time.Second)

	// A 加锁
	resultA := lockA.TryLock()
	fmt.Printf("  Client A 加锁（TTL=200ms）: %v\n", resultA)

	// B 立即尝试（失败）
	resultB := lockB.TryLock()
	fmt.Printf("  Client B 立即尝试: %v（预期：false）\n", resultB)

	// 等待 A 的锁过期
	fmt.Println("  等待 300ms（A 的锁过期）...")
	time.Sleep(300 * time.Millisecond)

	// B 再次尝试（成功，A 的锁已过期）
	resultB = lockB.TryLock()
	fmt.Printf("  Client B 再次尝试: %v（预期：true，A 的锁已过期）\n", resultB)
	_ = lockB.Unlock()
	fmt.Println()
}

// demoConcurrentCompetition 并发竞争
func demoConcurrentCompetition(store *MemoryLockStore) {
	fmt.Println("--- 场景四：并发竞争（模拟库存扣减）---")
	fmt.Println("  10 个 goroutine 竞争同一把锁，模拟库存扣减")

	var (
		stock       int32 = 10 // 初始库存
		successOps  atomic.Int64
		failedLocks atomic.Int64
	)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			lock := NewDistributedLock(store, "lock:stock:product-1", 5*time.Second)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// 获取锁
			if err := lock.Lock(ctx); err != nil {
				failedLocks.Add(1)
				return
			}
			defer lock.Unlock()

			// 临界区：扣减库存
			current := atomic.LoadInt32(&stock)
			if current > 0 {
				// 模拟业务处理耗时
				time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
				atomic.AddInt32(&stock, -1)
				successOps.Add(1)
				fmt.Printf("    Worker %d: 扣减成功，剩余库存 %d\n", workerID, atomic.LoadInt32(&stock))
			} else {
				fmt.Printf("    Worker %d: 库存不足\n", workerID)
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("\n  结果: 成功扣减 %d 次, 获取锁失败 %d 次, 最终库存 %d\n",
		successOps.Load(), failedLocks.Load(), atomic.LoadInt32(&stock))
	fmt.Println()
}

// partB 连接真实 Redis
func partB() {
	fmt.Println("=== Part B：真实 Redis 分布式锁 ===")
	fmt.Println()

	// 连接 Redis
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	ctx := context.Background()

	// 检查连接
	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Printf("❌ Redis 连接失败: %v\n", err)
		fmt.Println("请先启动 Redis: docker compose -f docker/docker-compose.yml up -d redis")
		return
	}
	fmt.Println("✅ Redis 连接成功")
	fmt.Println()

	// 场景一：基本 Redis 锁
	fmt.Println("--- Redis 分布式锁基本操作 ---")

	lock := NewRedisDistributedLock(client, "distributed:lock:demo", 30*time.Second)

	// 加锁
	err := lock.Lock(ctx)
	fmt.Printf("  加锁: err=%v, owner=%s\n", err, lock.owner)

	// 验证锁存在
	val, _ := client.Get(ctx, "distributed:lock:demo").Result()
	ttl, _ := client.TTL(ctx, "distributed:lock:demo").Result()
	fmt.Printf("  Redis 中的值: %s, TTL: %v\n", val, ttl)

	// 解锁
	err = lock.Unlock(ctx)
	fmt.Printf("  解锁: err=%v\n", err)

	// 验证锁已删除
	exists, _ := client.Exists(ctx, "distributed:lock:demo").Result()
	fmt.Printf("  锁是否存在: %v（预期：0）\n", exists)
	fmt.Println()

	// 场景二：并发竞争
	fmt.Println("--- Redis 并发竞争 ---")
	fmt.Println("  5 个 goroutine 竞争 Redis 锁")

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			rLock := NewRedisDistributedLock(client, "distributed:lock:concurrent", 10*time.Second)
			lockCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			if err := rLock.Lock(lockCtx); err != nil {
				fmt.Printf("    Worker %d: 获取锁失败 - %v\n", id, err)
				return
			}

			fmt.Printf("    Worker %d: 获取锁成功，执行业务...\n", id)
			time.Sleep(200 * time.Millisecond)

			if err := rLock.Unlock(ctx); err != nil {
				fmt.Printf("    Worker %d: 释放锁失败 - %v\n", id, err)
			} else {
				fmt.Printf("    Worker %d: 释放锁成功\n", id)
			}
		}(i)
	}
	wg.Wait()
	fmt.Println()

	// 清理
	client.Del(ctx, "distributed:lock:demo", "distributed:lock:concurrent")
}
