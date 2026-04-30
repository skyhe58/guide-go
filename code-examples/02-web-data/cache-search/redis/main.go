// go-redis 完整示例 — 数据结构 / Pipeline / 分布式锁
// 演示：Redis 五种数据结构操作、Pipeline 批量命令、事务、Pub/Sub、分布式锁
// Go 1.22+ | 验证日期 2025-01-01
//
// Part A：纯内存模拟，直接运行理解 Redis 核心概念
// Part B：连接真实 Redis，需传入参数 'real'
//
// 运行方式：
//   go run ./redis/              # Part A：内存模拟
//   go run ./redis/ real         # Part B：连接真实 Redis
//
// Part B 前置条件：
//   docker compose -f docker/docker-compose.yml up -d redis
//   连接地址：localhost:6379，无密码

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

	"github.com/redis/go-redis/v9"
)

// ============================================================
// Part A：纯内存模拟 Redis 核心概念
// ============================================================

func partA() {
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Part A：Redis 核心概念（内存模拟）")
	fmt.Println("=" + strings.Repeat("=", 59))

	// 1. 模拟 Redis 数据结构
	demoDataStructures()

	// 2. 模拟 Pipeline 概念
	demoPipelineConcept()

	// 3. 模拟分布式锁
	demoDistributedLockConcept()

	// 4. 模拟缓存穿透/击穿方案
	demoCacheProblems()
}

// ============================================================
// 1. 模拟 Redis 数据结构
// ============================================================

func demoDataStructures() {
	fmt.Println("\n--- 1. Redis 数据结构（内存模拟） ---")

	// --- String ---
	fmt.Println("\n[String] 最基本的数据类型，可存储字符串、整数、浮点数")
	stringStore := make(map[string]string)
	stringStore["user:1001:name"] = "张三"
	stringStore["user:1001:age"] = "28"
	stringStore["page:views"] = "10086"
	fmt.Printf("  SET user:1001:name = %s\n", stringStore["user:1001:name"])
	fmt.Printf("  SET page:views = %s\n", stringStore["page:views"])
	fmt.Println("  场景：缓存、计数器、分布式锁、Session 存储")

	// --- Hash ---
	fmt.Println("\n[Hash] 键值对集合，适合存储对象")
	hashStore := map[string]map[string]string{
		"user:1001": {
			"name":  "张三",
			"email": "zhangsan@example.com",
			"age":   "28",
		},
	}
	fmt.Printf("  HSET user:1001 name=%s email=%s\n",
		hashStore["user:1001"]["name"], hashStore["user:1001"]["email"])
	fmt.Println("  场景：用户信息、购物车、对象缓存")

	// --- List ---
	fmt.Println("\n[List] 有序列表，支持头尾插入")
	listStore := map[string][]string{
		"timeline:user1": {"消息3", "消息2", "消息1"},
	}
	// 模拟 LPUSH
	listStore["timeline:user1"] = append([]string{"消息4"}, listStore["timeline:user1"]...)
	fmt.Printf("  LPUSH timeline:user1 → %v\n", listStore["timeline:user1"])
	fmt.Println("  场景：消息队列、最新消息列表、时间线")

	// --- Set ---
	fmt.Println("\n[Set] 无序集合，自动去重")
	setStore := map[string]map[string]bool{
		"tags:article:1": {"Go": true, "后端": true, "入门": true},
		"tags:article:2": {"Go": true, "并发": true, "高级": true},
	}
	// 模拟 SINTER（交集）
	intersection := make([]string, 0)
	for tag := range setStore["tags:article:1"] {
		if setStore["tags:article:2"][tag] {
			intersection = append(intersection, tag)
		}
	}
	fmt.Printf("  SINTER tags:article:1 tags:article:2 → %v\n", intersection)
	fmt.Println("  场景：标签系统、共同好友、去重")

	// --- Sorted Set (ZSet) ---
	fmt.Println("\n[Sorted Set] 有序集合，每个元素关联一个分数")
	type ZSetMember struct {
		Member string
		Score  float64
	}
	zsetStore := []ZSetMember{
		{"player:alice", 2500},
		{"player:bob", 1800},
		{"player:charlie", 3200},
		{"player:dave", 2100},
	}
	// 按分数排序
	sort.Slice(zsetStore, func(i, j int) bool {
		return zsetStore[i].Score > zsetStore[j].Score
	})
	fmt.Println("  排行榜（ZREVRANGE）：")
	for i, m := range zsetStore {
		fmt.Printf("    第 %d 名: %s (分数: %.0f)\n", i+1, m.Member, m.Score)
	}
	fmt.Println("  场景：排行榜、延迟队列、带权重的队列")
}

// ============================================================
// 2. 模拟 Pipeline 概念
// ============================================================

func demoPipelineConcept() {
	fmt.Println("\n--- 2. Pipeline 概念（内存模拟） ---")

	fmt.Println("\n[无 Pipeline] 每条命令单独发送，N 条命令需要 N 次网络往返（RTT）")
	fmt.Println("  SET a 1  →  OK     (1 RTT)")
	fmt.Println("  SET b 2  →  OK     (1 RTT)")
	fmt.Println("  SET c 3  →  OK     (1 RTT)")
	fmt.Println("  总计：3 次 RTT")

	fmt.Println("\n[Pipeline] 多条命令打包发送，只需 1 次网络往返")
	fmt.Println("  PIPELINE {")
	fmt.Println("    SET a 1")
	fmt.Println("    SET b 2")
	fmt.Println("    SET c 3")
	fmt.Println("  } → [OK, OK, OK]   (1 RTT)")
	fmt.Println("  总计：1 次 RTT")

	fmt.Println("\n[事务 TxPipeline] MULTI/EXEC 包裹，保证原子执行")
	fmt.Println("  MULTI")
	fmt.Println("    DEBIT  account:A -100")
	fmt.Println("    CREDIT account:B +100")
	fmt.Println("  EXEC → [OK, OK]")
	fmt.Println("  注意：Redis 事务不支持回滚！")
}

// ============================================================
// 3. 模拟分布式锁
// ============================================================

// MemoryLock 内存模拟的分布式锁
type MemoryLock struct {
	mu    sync.Mutex
	locks map[string]string // key -> owner(uuid)
}

func NewMemoryLock() *MemoryLock {
	return &MemoryLock{locks: make(map[string]string)}
}

// TryLock 模拟 SET key uuid NX EX timeout
func (ml *MemoryLock) TryLock(key, owner string) bool {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	if _, exists := ml.locks[key]; exists {
		return false // 锁已被持有
	}
	ml.locks[key] = owner
	return true
}

// Unlock 模拟 Lua 脚本释放锁（判断 owner 后删除）
func (ml *MemoryLock) Unlock(key, owner string) bool {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	if ml.locks[key] == owner {
		delete(ml.locks, key)
		return true
	}
	return false // 不是锁的持有者，拒绝释放
}

func demoDistributedLockConcept() {
	fmt.Println("\n--- 3. 分布式锁（内存模拟） ---")

	lock := NewMemoryLock()
	lockKey := "order:lock:1001"

	// 模拟两个客户端竞争锁
	owner1 := "uuid-client-1"
	owner2 := "uuid-client-2"

	fmt.Printf("\n客户端 1 尝试加锁: SET %s %s NX EX 30\n", lockKey, owner1)
	if lock.TryLock(lockKey, owner1) {
		fmt.Println("  → 加锁成功 ✅")
	}

	fmt.Printf("客户端 2 尝试加锁: SET %s %s NX EX 30\n", lockKey, owner2)
	if !lock.TryLock(lockKey, owner2) {
		fmt.Println("  → 加锁失败 ❌（锁已被客户端 1 持有）")
	}

	fmt.Println("\n客户端 1 执行业务逻辑...")
	fmt.Println("客户端 1 释放锁（Lua 脚本：if GET == uuid then DEL）")
	if lock.Unlock(lockKey, owner1) {
		fmt.Println("  → 释放成功 ✅")
	}

	fmt.Printf("\n客户端 2 重试加锁: SET %s %s NX EX 30\n", lockKey, owner2)
	if lock.TryLock(lockKey, owner2) {
		fmt.Println("  → 加锁成功 ✅")
		lock.Unlock(lockKey, owner2)
	}

	fmt.Println("\n⚠️  分布式锁要点：")
	fmt.Println("  1. 原子加锁：SET key uuid NX EX timeout")
	fmt.Println("  2. 唯一标识：Value 用 UUID，防止误删他人的锁")
	fmt.Println("  3. Lua 脚本释放：判断和删除必须原子执行")
	fmt.Println("  4. 看门狗续期：防止业务超时导致锁过期")
}

// ============================================================
// 4. 模拟缓存穿透/击穿方案
// ============================================================

func demoCacheProblems() {
	fmt.Println("\n--- 4. 缓存穿透/击穿方案（内存模拟） ---")

	// 模拟缓存和数据库
	cache := make(map[string]string)
	db := map[string]string{
		"user:1": "张三",
		"user:2": "李四",
	}

	// --- 缓存穿透：查询不存在的数据 ---
	fmt.Println("\n[缓存穿透] 查询不存在的数据")
	key := "user:999"
	if val, ok := cache[key]; ok {
		fmt.Printf("  缓存命中: %s\n", val)
	} else {
		fmt.Println("  缓存未命中，查询数据库...")
		if val, ok := db[key]; ok {
			cache[key] = val
			fmt.Printf("  数据库命中: %s，写入缓存\n", val)
		} else {
			// 解决方案：缓存空值
			cache[key] = "<nil>"
			fmt.Println("  数据库也不存在！缓存空值（TTL=60s）防止穿透")
		}
	}

	// --- 缓存击穿：singleflight 模拟 ---
	fmt.Println("\n[缓存击穿] 热点 Key 过期，多个请求同时查数据库")
	fmt.Println("  解决方案：singleflight — 相同 Key 的并发请求只执行一次数据库查询")

	var (
		sfMu     sync.Mutex
		sfCalls  = make(map[string]chan string)
		dbCalled int
	)

	singleflightGet := func(key string) string {
		sfMu.Lock()
		if ch, ok := sfCalls[key]; ok {
			sfMu.Unlock()
			return <-ch // 等待第一个请求的结果
		}
		ch := make(chan string, 1)
		sfCalls[key] = ch
		sfMu.Unlock()

		// 只有第一个请求查数据库
		dbCalled++
		result := db["user:1"] // 模拟数据库查询
		ch <- result

		sfMu.Lock()
		delete(sfCalls, key)
		sfMu.Unlock()
		return result
	}

	result := singleflightGet("user:1")
	fmt.Printf("  singleflight 结果: %s（数据库查询次数: %d）\n", result, dbCalled)

	fmt.Println("\n[缓存雪崩] 大量 Key 同时过期")
	fmt.Println("  解决方案：过期时间加随机值")
	baseTTL := 300 // 5 分钟
	for i := 1; i <= 5; i++ {
		randomTTL := baseTTL + rand.Intn(60) // 加 0-60 秒随机值
		fmt.Printf("  Key user:%d TTL = %ds（基础 %ds + 随机 %ds）\n",
			i, randomTTL, baseTTL, randomTTL-baseTTL)
	}
}

// ============================================================
// Part B：连接真实 Redis（需要 Docker）
// ============================================================

func partB() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Part B：连接真实 Redis（go-redis）")
	fmt.Println(strings.Repeat("=", 60))

	ctx := context.Background()

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Printf("❌ 无法连接 Redis: %v\n", err)
		fmt.Println("请先启动 Redis: docker compose -f docker/docker-compose.yml up -d redis")
		return
	}
	fmt.Println("✅ Redis 连接成功")

	// 1. 数据结构操作
	demoRealDataStructures(ctx, rdb)

	// 2. Pipeline
	demoRealPipeline(ctx, rdb)

	// 3. 事务
	demoRealTransaction(ctx, rdb)

	// 4. 分布式锁
	demoRealDistributedLock(ctx, rdb)

	// 清理测试数据
	cleanupTestData(ctx, rdb)
}

// demoRealDataStructures 演示 go-redis 操作各种数据结构
func demoRealDataStructures(ctx context.Context, rdb *redis.Client) {
	fmt.Println("\n--- 1. 数据结构操作（go-redis） ---")

	// --- String ---
	fmt.Println("\n[String]")
	rdb.Set(ctx, "demo:name", "Go知识库", 5*time.Minute)
	val, _ := rdb.Get(ctx, "demo:name").Result()
	fmt.Printf("  GET demo:name = %s\n", val)

	// INCR 计数器
	rdb.Set(ctx, "demo:counter", 0, 5*time.Minute)
	rdb.Incr(ctx, "demo:counter")
	rdb.Incr(ctx, "demo:counter")
	rdb.IncrBy(ctx, "demo:counter", 10)
	counter, _ := rdb.Get(ctx, "demo:counter").Int64()
	fmt.Printf("  INCR demo:counter = %d\n", counter)

	// --- Hash ---
	fmt.Println("\n[Hash]")
	rdb.HSet(ctx, "demo:user:1001", map[string]interface{}{
		"name":  "张三",
		"email": "zhangsan@example.com",
		"age":   28,
	})
	rdb.Expire(ctx, "demo:user:1001", 5*time.Minute)
	userMap, _ := rdb.HGetAll(ctx, "demo:user:1001").Result()
	fmt.Printf("  HGETALL demo:user:1001 = %v\n", userMap)

	// --- List ---
	fmt.Println("\n[List]")
	rdb.Del(ctx, "demo:queue")
	rdb.RPush(ctx, "demo:queue", "任务1", "任务2", "任务3")
	rdb.Expire(ctx, "demo:queue", 5*time.Minute)
	task, _ := rdb.LPop(ctx, "demo:queue").Result()
	fmt.Printf("  LPOP demo:queue = %s\n", task)
	remaining, _ := rdb.LRange(ctx, "demo:queue", 0, -1).Result()
	fmt.Printf("  LRANGE demo:queue = %v\n", remaining)

	// --- Set ---
	fmt.Println("\n[Set]")
	rdb.SAdd(ctx, "demo:tags:1", "Go", "后端", "入门")
	rdb.SAdd(ctx, "demo:tags:2", "Go", "并发", "高级")
	rdb.Expire(ctx, "demo:tags:1", 5*time.Minute)
	rdb.Expire(ctx, "demo:tags:2", 5*time.Minute)
	inter, _ := rdb.SInter(ctx, "demo:tags:1", "demo:tags:2").Result()
	fmt.Printf("  SINTER demo:tags:1 demo:tags:2 = %v\n", inter)

	// --- Sorted Set ---
	fmt.Println("\n[Sorted Set]")
	rdb.Del(ctx, "demo:leaderboard")
	rdb.ZAdd(ctx, "demo:leaderboard",
		redis.Z{Score: 2500, Member: "alice"},
		redis.Z{Score: 1800, Member: "bob"},
		redis.Z{Score: 3200, Member: "charlie"},
		redis.Z{Score: 2100, Member: "dave"},
	)
	rdb.Expire(ctx, "demo:leaderboard", 5*time.Minute)
	topPlayers, _ := rdb.ZRevRangeWithScores(ctx, "demo:leaderboard", 0, 2).Result()
	fmt.Println("  排行榜 Top 3:")
	for i, z := range topPlayers {
		fmt.Printf("    第 %d 名: %s (分数: %.0f)\n", i+1, z.Member, z.Score)
	}
}

// demoRealPipeline 演示 Pipeline 批量操作
func demoRealPipeline(ctx context.Context, rdb *redis.Client) {
	fmt.Println("\n--- 2. Pipeline 批量操作 ---")

	pipe := rdb.Pipeline()
	for i := 0; i < 10; i++ {
		pipe.Set(ctx, fmt.Sprintf("demo:pipe:%d", i), fmt.Sprintf("value-%d", i), 5*time.Minute)
	}
	cmds, err := pipe.Exec(ctx)
	if err != nil {
		fmt.Printf("  Pipeline 执行失败: %v\n", err)
		return
	}
	fmt.Printf("  Pipeline 批量写入 %d 条数据成功\n", len(cmds))

	// Pipeline 批量读取
	pipe = rdb.Pipeline()
	getResults := make([]*redis.StringCmd, 10)
	for i := 0; i < 10; i++ {
		getResults[i] = pipe.Get(ctx, fmt.Sprintf("demo:pipe:%d", i))
	}
	pipe.Exec(ctx)
	fmt.Println("  Pipeline 批量读取结果:")
	for i, cmd := range getResults {
		fmt.Printf("    demo:pipe:%d = %s\n", i, cmd.Val())
	}
}

// demoRealTransaction 演示事务操作
func demoRealTransaction(ctx context.Context, rdb *redis.Client) {
	fmt.Println("\n--- 3. 事务（TxPipeline） ---")

	// 模拟转账：A 扣款 100，B 加款 100
	rdb.Set(ctx, "demo:account:A", 1000, 5*time.Minute)
	rdb.Set(ctx, "demo:account:B", 500, 5*time.Minute)

	txPipe := rdb.TxPipeline()
	txPipe.DecrBy(ctx, "demo:account:A", 100)
	txPipe.IncrBy(ctx, "demo:account:B", 100)
	_, err := txPipe.Exec(ctx)
	if err != nil {
		fmt.Printf("  事务执行失败: %v\n", err)
		return
	}

	balanceA, _ := rdb.Get(ctx, "demo:account:A").Int64()
	balanceB, _ := rdb.Get(ctx, "demo:account:B").Int64()
	fmt.Printf("  转账后 A 余额: %d, B 余额: %d\n", balanceA, balanceB)
}

// demoRealDistributedLock 演示分布式锁
func demoRealDistributedLock(ctx context.Context, rdb *redis.Client) {
	fmt.Println("\n--- 4. 分布式锁（go-redis） ---")

	lockKey := "demo:lock:order:1001"
	lockValue := fmt.Sprintf("owner-%d", time.Now().UnixNano())
	lockTTL := 30 * time.Second

	// 加锁：SET key value NX EX
	ok, err := rdb.SetNX(ctx, lockKey, lockValue, lockTTL).Result()
	if err != nil {
		fmt.Printf("  加锁失败: %v\n", err)
		return
	}
	if ok {
		fmt.Printf("  加锁成功 ✅ key=%s value=%s TTL=%v\n", lockKey, lockValue, lockTTL)
	} else {
		fmt.Println("  加锁失败 ❌ 锁已被其他客户端持有")
		return
	}

	// 模拟业务逻辑
	fmt.Println("  执行业务逻辑...")
	time.Sleep(100 * time.Millisecond)

	// 释放锁：Lua 脚本保证原子性
	luaScript := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	result, err := rdb.Eval(ctx, luaScript, []string{lockKey}, lockValue).Int64()
	if err != nil {
		fmt.Printf("  释放锁失败: %v\n", err)
		return
	}
	if result == 1 {
		fmt.Println("  释放锁成功 ✅（Lua 脚本原子释放）")
	} else {
		fmt.Println("  释放锁失败 ❌ 锁已被其他客户端持有")
	}
}

// cleanupTestData 清理测试数据
func cleanupTestData(ctx context.Context, rdb *redis.Client) {
	// 使用 SCAN 查找并删除 demo: 前缀的 Key
	var cursor uint64
	for {
		keys, nextCursor, err := rdb.Scan(ctx, cursor, "demo:*", 100).Result()
		if err != nil {
			break
		}
		if len(keys) > 0 {
			rdb.Del(ctx, keys...)
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
}

// ============================================================
// main 入口
// ============================================================

func main() {
	// Part A：纯内存模拟，直接运行理解原理
	partA()

	// Part B：连接真实 Redis，需传入参数 'real'
	if len(os.Args) > 1 && os.Args[1] == "real" {
		partB()
	}
}
