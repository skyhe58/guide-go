// 秒杀系统核心链路 — 架构设计场景
// 本示例演示秒杀系统的三大核心组件：令牌桶限流 → 原子库存扣减 → 异步下单队列
// 使用纯 Go 实现，无需外部依赖，通过并发 goroutine 模拟真实秒杀场景
//
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：go run ./04-distributed/architecture/seckill/

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// Part A：秒杀系统核心链路（纯内存模拟）
// ============================================================

// --- 组件 1：令牌桶限流器 ---

// TokenBucket 令牌桶限流器
// 以固定速率生成令牌，请求到来时消耗令牌
// 桶满时丢弃新令牌，桶空时拒绝请求
type TokenBucket struct {
	rate       float64   // 令牌生成速率（个/秒）
	capacity   float64   // 桶容量（最大令牌数）
	tokens     float64   // 当前令牌数
	lastRefill time.Time // 上次补充令牌的时间
	mu         sync.Mutex
}

// NewTokenBucket 创建令牌桶
// rate: 每秒生成的令牌数
// capacity: 桶的最大容量（允许的突发量）
func NewTokenBucket(rate float64, capacity int) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		capacity:   float64(capacity),
		tokens:     float64(capacity), // 初始满桶
		lastRefill: time.Now(),
	}
}

// Allow 尝试获取一个令牌
// 返回 true 表示放行，false 表示限流拒绝
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 计算自上次补充以来应生成的令牌数
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.rate
	// 令牌数不超过桶容量
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now

	// 尝试消耗一个令牌
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// --- 组件 2：原子库存扣减器 ---

// StockManager 库存管理器
// 使用 atomic 原子操作模拟 Redis DECR 命令
// 保证在高并发下库存扣减的正确性，防止超卖
type StockManager struct {
	stock int64 // 当前库存（原子操作）
}

// NewStockManager 创建库存管理器
func NewStockManager(initialStock int64) *StockManager {
	return &StockManager{stock: initialStock}
}

// Deduct 原子扣减库存
// 模拟 Redis DECR 命令：先扣减，如果结果为负则回滚
// 返回值：(是否成功, 扣减后的库存)
func (sm *StockManager) Deduct() (bool, int64) {
	// 原子减 1（等价于 Redis DECR）
	remaining := atomic.AddInt64(&sm.stock, -1)

	if remaining < 0 {
		// 库存不足，回滚（等价于 Redis INCR）
		atomic.AddInt64(&sm.stock, 1)
		return false, 0
	}
	return true, remaining
}

// GetStock 获取当前库存
func (sm *StockManager) GetStock() int64 {
	return atomic.LoadInt64(&sm.stock)
}

// --- 组件 3：异步下单队列 ---

// OrderRequest 下单请求
type OrderRequest struct {
	UserID    int
	ProductID string
	Timestamp time.Time
}

// OrderResult 下单结果
type OrderResult struct {
	OrderID string
	UserID  int
	Success bool
	Message string
}

// AsyncOrderQueue 异步下单队列
// 使用 buffered channel 模拟消息队列（如 Kafka/RabbitMQ）
// 将瞬时的写压力分散到一段时间内处理
type AsyncOrderQueue struct {
	ch       chan OrderRequest
	results  sync.Map // userID → OrderResult
	orderSeq int64    // 订单序号（原子递增）
}

// NewAsyncOrderQueue 创建异步下单队列
// bufferSize: 队列缓冲区大小（模拟 MQ 的容量）
func NewAsyncOrderQueue(bufferSize int) *AsyncOrderQueue {
	return &AsyncOrderQueue{
		ch: make(chan OrderRequest, bufferSize),
	}
}

// Enqueue 将下单请求放入队列
// 返回 false 表示队列已满（系统过载）
func (q *AsyncOrderQueue) Enqueue(req OrderRequest) bool {
	select {
	case q.ch <- req:
		return true
	default:
		return false // 队列满，拒绝请求
	}
}

// StartConsumers 启动消费者协程
// 模拟订单服务异步消费消息，创建订单
func (q *AsyncOrderQueue) StartConsumers(workerCount int, wg *sync.WaitGroup) {
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for req := range q.ch {
				// 模拟订单创建（数据库写入）
				time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)

				orderNum := atomic.AddInt64(&q.orderSeq, 1)
				orderID := fmt.Sprintf("ORD-%06d", orderNum)

				result := OrderResult{
					OrderID: orderID,
					UserID:  req.UserID,
					Success: true,
					Message: "下单成功",
				}
				q.results.Store(req.UserID, result)
			}
		}(i)
	}
}

// GetResult 查询下单结果
func (q *AsyncOrderQueue) GetResult(userID int) (OrderResult, bool) {
	val, ok := q.results.Load(userID)
	if !ok {
		return OrderResult{}, false
	}
	return val.(OrderResult), true
}

// Close 关闭队列（停止接收新请求）
func (q *AsyncOrderQueue) Close() {
	close(q.ch)
}

// --- 秒杀系统主流程 ---

// SeckillSystem 秒杀系统
// 整合限流器、库存管理器和异步下单队列
type SeckillSystem struct {
	limiter  *TokenBucket
	stock    *StockManager
	queue    *AsyncOrderQueue
	stats    SeckillStats
	userLock sync.Map // 用户维度去重（防止同一用户重复购买）
}

// SeckillStats 秒杀统计数据
type SeckillStats struct {
	totalRequests   int64 // 总请求数
	rateLimited     int64 // 被限流的请求数
	stockEmpty      int64 // 库存不足的请求数
	duplicateUser   int64 // 重复购买的请求数
	enqueueSuccess  int64 // 成功进入队列的请求数
	enqueueFailed   int64 // 队列满被拒绝的请求数
}

// NewSeckillSystem 创建秒杀系统
func NewSeckillSystem(stock int64, rateLimit float64, queueSize int) *SeckillSystem {
	return &SeckillSystem{
		limiter: NewTokenBucket(rateLimit, int(rateLimit)*2), // 桶容量为速率的 2 倍
		stock:   NewStockManager(stock),
		queue:   NewAsyncOrderQueue(queueSize),
	}
}

// HandleRequest 处理秒杀请求
// 完整链路：限流 → 去重 → 库存扣减 → 异步下单
func (s *SeckillSystem) HandleRequest(userID int, productID string) string {
	atomic.AddInt64(&s.stats.totalRequests, 1)

	// 第一关：令牌桶限流
	if !s.limiter.Allow() {
		atomic.AddInt64(&s.stats.rateLimited, 1)
		return "系统繁忙，请稍后重试"
	}

	// 第二关：用户去重（防止同一用户重复购买）
	if _, loaded := s.userLock.LoadOrStore(userID, true); loaded {
		atomic.AddInt64(&s.stats.duplicateUser, 1)
		return "您已参与过本次秒杀"
	}

	// 第三关：原子库存扣减
	ok, remaining := s.stock.Deduct()
	if !ok {
		s.userLock.Delete(userID) // 扣减失败，释放用户锁
		atomic.AddInt64(&s.stats.stockEmpty, 1)
		return "商品已售罄"
	}

	// 第四关：异步下单（发送到消息队列）
	req := OrderRequest{
		UserID:    userID,
		ProductID: productID,
		Timestamp: time.Now(),
	}
	if !s.queue.Enqueue(req) {
		// 队列满，回滚库存和用户锁
		atomic.AddInt64(&s.stock.stock, 1)
		s.userLock.Delete(userID)
		atomic.AddInt64(&s.stats.enqueueFailed, 1)
		return "系统繁忙，请稍后重试"
	}

	atomic.AddInt64(&s.stats.enqueueSuccess, 1)
	_ = remaining
	return fmt.Sprintf("秒杀成功！排队中，剩余库存: %d", remaining)
}

// PrintStats 打印秒杀统计数据
func (s *SeckillSystem) PrintStats() {
	fmt.Println("\n========== 秒杀统计 ==========")
	fmt.Printf("总请求数:       %d\n", atomic.LoadInt64(&s.stats.totalRequests))
	fmt.Printf("被限流:         %d\n", atomic.LoadInt64(&s.stats.rateLimited))
	fmt.Printf("库存不足:       %d\n", atomic.LoadInt64(&s.stats.stockEmpty))
	fmt.Printf("重复购买:       %d\n", atomic.LoadInt64(&s.stats.duplicateUser))
	fmt.Printf("成功下单:       %d\n", atomic.LoadInt64(&s.stats.enqueueSuccess))
	fmt.Printf("队列满拒绝:     %d\n", atomic.LoadInt64(&s.stats.enqueueFailed))
	fmt.Printf("最终剩余库存:   %d\n", s.stock.GetStock())
	fmt.Println("================================")
}

func main() {
	fmt.Println("========== 秒杀系统核心链路演示 ==========")
	fmt.Println()

	// --- 演示 1：令牌桶限流器 ---
	fmt.Println("--- 1. 令牌桶限流器演示 ---")
	bucket := NewTokenBucket(10, 5) // 每秒 10 个令牌，桶容量 5
	allowed, denied := 0, 0
	for i := 0; i < 20; i++ {
		if bucket.Allow() {
			allowed++
		} else {
			denied++
		}
	}
	fmt.Printf("20 个请求：放行 %d，拒绝 %d（桶容量 5，初始满桶）\n", allowed, denied)
	fmt.Println()

	// --- 演示 2：原子库存扣减 ---
	fmt.Println("--- 2. 原子库存扣减演示（并发安全） ---")
	stockMgr := NewStockManager(10) // 初始库存 10
	var wg sync.WaitGroup
	successCount := int64(0)

	// 50 个 goroutine 并发扣减库存
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ok, _ := stockMgr.Deduct(); ok {
				atomic.AddInt64(&successCount, 1)
			}
		}()
	}
	wg.Wait()
	fmt.Printf("50 个并发请求抢 10 件库存：成功 %d，剩余库存 %d（无超卖）\n",
		successCount, stockMgr.GetStock())
	fmt.Println()

	// --- 演示 3：完整秒杀流程 ---
	fmt.Println("--- 3. 完整秒杀流程演示 ---")
	fmt.Println("配置：库存 100，限流 500/秒，队列容量 200")
	fmt.Println("模拟：1000 个用户并发抢购")
	fmt.Println()

	system := NewSeckillSystem(
		100,  // 库存 100 件
		500,  // 限流 500 QPS
		200,  // 队列容量 200
	)

	// 启动异步订单消费者（3 个 worker）
	var consumerWg sync.WaitGroup
	system.queue.StartConsumers(3, &consumerWg)

	// 模拟 1000 个用户并发秒杀
	var userWg sync.WaitGroup
	for i := 1; i <= 1000; i++ {
		userWg.Add(1)
		go func(userID int) {
			defer userWg.Done()
			// 模拟用户随机到达
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			result := system.HandleRequest(userID, "PRODUCT-001")
			// 打印部分结果（前 5 个成功的）
			if userID <= 5 {
				fmt.Printf("  用户 %d: %s\n", userID, result)
			}
		}(i)
	}

	// 等待所有用户请求完成
	userWg.Wait()

	// 关闭队列并等待消费者处理完成
	system.queue.Close()
	consumerWg.Wait()

	// 打印统计数据
	system.PrintStats()

	// 验证：成功下单数不超过库存
	successOrders := atomic.LoadInt64(&system.stats.enqueueSuccess)
	if successOrders <= 100 {
		fmt.Println("✅ 验证通过：未发生超卖")
	} else {
		fmt.Println("❌ 验证失败：发生超卖！")
	}

	// 查询部分订单结果
	fmt.Println("\n--- 部分订单结果 ---")
	for i := 1; i <= 5; i++ {
		if result, ok := system.queue.GetResult(i); ok {
			fmt.Printf("  用户 %d → 订单号: %s, 状态: %s\n",
				result.UserID, result.OrderID, result.Message)
		}
	}
}
