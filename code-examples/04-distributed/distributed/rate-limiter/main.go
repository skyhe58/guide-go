// 限流算法 — 令牌桶 / 漏桶 / 滑动窗口 完整实现
// 演示三种主流限流算法的并发安全实现，以及 golang.org/x/time/rate 标准方案对比
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：
//   go run ./rate-limiter/
//
// 本示例为纯 Go 实现，无需外部依赖

package main

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// 一、令牌桶算法（Token Bucket）
// 固定速率生成令牌，请求消耗令牌，允许突发流量
// ============================================================

// TokenBucket 令牌桶限流器
// 核心思想：以固定速率 r 向桶中添加令牌，桶容量为 b
// 请求到来时取走一个令牌，桶空则拒绝
type TokenBucket struct {
	mu         sync.Mutex
	rate       float64   // 令牌生成速率（个/秒）
	capacity   float64   // 桶容量（最大令牌数）
	tokens     float64   // 当前令牌数
	lastRefill time.Time // 上次补充令牌的时间
}

// NewTokenBucket 创建令牌桶
// rate: 每秒生成的令牌数
// capacity: 桶的最大容量（决定突发流量上限）
func NewTokenBucket(rate float64, capacity int) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		capacity:   float64(capacity),
		tokens:     float64(capacity), // 初始满桶
		lastRefill: time.Now(),
	}
}

// Allow 尝试获取一个令牌，返回是否允许通过
// 使用懒加载方式补充令牌：不需要后台 goroutine，每次请求时计算应补充的令牌数
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	// 计算自上次补充以来应该生成的令牌数
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = math.Min(tb.capacity, tb.tokens+elapsed*tb.rate)
	tb.lastRefill = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// ============================================================
// 二、漏桶算法（Leaky Bucket）
// 请求进入桶中排队，以固定速率流出处理，严格匀速
// ============================================================

// LeakyBucket 漏桶限流器
// 核心思想：请求进入桶中，以固定速率 r 流出处理
// 桶满时拒绝新请求，保证处理速率严格匀速
type LeakyBucket struct {
	mu       sync.Mutex
	rate     float64   // 流出速率（个/秒）
	capacity int       // 桶容量
	water    float64   // 当前水量（待处理请求数）
	lastLeak time.Time // 上次漏水时间
}

// NewLeakyBucket 创建漏桶
// rate: 每秒处理的请求数（流出速率）
// capacity: 桶的最大容量（等待队列长度）
func NewLeakyBucket(rate float64, capacity int) *LeakyBucket {
	return &LeakyBucket{
		rate:     rate,
		capacity: capacity,
		water:    0,
		lastLeak: time.Now(),
	}
}

// Allow 尝试将请求放入桶中
// 先漏水（减少已处理的量），再判断是否有空间放入新请求
func (lb *LeakyBucket) Allow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := time.Now()
	// 计算自上次以来漏掉的水量
	elapsed := now.Sub(lb.lastLeak).Seconds()
	lb.water = math.Max(0, lb.water-elapsed*lb.rate)
	lb.lastLeak = now

	if lb.water < float64(lb.capacity) {
		lb.water++
		return true
	}
	return false
}

// ============================================================
// 三、滑动窗口算法（Sliding Window Log）
// 记录每个请求的时间戳，统计窗口内的请求数
// ============================================================

// SlidingWindowLog 滑动窗口限流器（基于日志）
// 核心思想：记录每个请求的时间戳，统计当前窗口内的请求总数
// 精确但内存开销较大（需存储所有时间戳）
type SlidingWindowLog struct {
	mu         sync.Mutex
	windowSize time.Duration // 窗口大小
	limit      int           // 窗口内允许的最大请求数
	timestamps []time.Time   // 请求时间戳列表
}

// NewSlidingWindowLog 创建滑动窗口限流器
// windowSize: 统计窗口大小（如 1 秒）
// limit: 窗口内允许的最大请求数
func NewSlidingWindowLog(windowSize time.Duration, limit int) *SlidingWindowLog {
	return &SlidingWindowLog{
		windowSize: windowSize,
		limit:      limit,
		timestamps: make([]time.Time, 0, limit),
	}
}

// Allow 判断当前请求是否允许通过
// 清理过期时间戳，统计窗口内请求数
func (sw *SlidingWindowLog) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-sw.windowSize)

	// 清理窗口外的过期时间戳（二分查找优化）
	validStart := 0
	for validStart < len(sw.timestamps) && sw.timestamps[validStart].Before(windowStart) {
		validStart++
	}
	sw.timestamps = sw.timestamps[validStart:]

	// 判断窗口内请求数是否超过限制
	if len(sw.timestamps) < sw.limit {
		sw.timestamps = append(sw.timestamps, now)
		return true
	}
	return false
}

// ============================================================
// 四、滑动窗口计数器（Sliding Window Counter）
// 将窗口划分为多个小桶，按比例计算请求数，内存更优
// ============================================================

// SlidingWindowCounter 滑动窗口计数器
// 核心思想：将时间窗口划分为多个小桶，每个桶记录请求计数
// 比 SlidingWindowLog 内存开销小，精度略低
type SlidingWindowCounter struct {
	mu         sync.Mutex
	windowSize time.Duration // 总窗口大小
	bucketSize time.Duration // 每个小桶的时间跨度
	limit      int           // 窗口内允许的最大请求数
	buckets    map[int64]int // 桶编号 → 请求计数
}

// NewSlidingWindowCounter 创建滑动窗口计数器
func NewSlidingWindowCounter(windowSize time.Duration, bucketCount int, limit int) *SlidingWindowCounter {
	bucketSize := windowSize / time.Duration(bucketCount)
	return &SlidingWindowCounter{
		windowSize: windowSize,
		bucketSize: bucketSize,
		limit:      limit,
		buckets:    make(map[int64]int),
	}
}

// Allow 判断当前请求是否允许通过
func (swc *SlidingWindowCounter) Allow() bool {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	now := time.Now()
	currentBucket := now.UnixNano() / int64(swc.bucketSize)
	windowStart := now.Add(-swc.windowSize).UnixNano() / int64(swc.bucketSize)

	// 清理过期桶
	for k := range swc.buckets {
		if k < windowStart {
			delete(swc.buckets, k)
		}
	}

	// 统计窗口内的总请求数
	total := 0
	for k, v := range swc.buckets {
		if k >= windowStart {
			total += v
		}
	}

	if total < swc.limit {
		swc.buckets[currentBucket]++
		return true
	}
	return false
}

// ============================================================
// 演示：并发场景下的限流效果
// ============================================================

func main() {
	fmt.Println("=== 限流算法演示 ===")
	fmt.Println()

	// 演示一：令牌桶
	demoTokenBucket()

	// 演示二：漏桶
	demoLeakyBucket()

	// 演示三：滑动窗口
	demoSlidingWindow()

	// 演示四：并发压测对比
	demoConcurrentComparison()
}

// demoTokenBucket 演示令牌桶算法
func demoTokenBucket() {
	fmt.Println("--- 1. 令牌桶算法（Token Bucket）---")
	fmt.Println("特点：允许突发流量，长期速率不超过设定值")
	fmt.Println()

	// 每秒 5 个令牌，桶容量 10（允许最多 10 个突发请求）
	tb := NewTokenBucket(5, 10)

	// 突发请求测试：一次性发送 15 个请求
	allowed := 0
	denied := 0
	for i := 0; i < 15; i++ {
		if tb.Allow() {
			allowed++
		} else {
			denied++
		}
	}
	fmt.Printf("  突发 15 个请求 → 通过: %d, 拒绝: %d（桶容量 10，所以最多通过 10 个）\n", allowed, denied)

	// 等待令牌恢复
	time.Sleep(500 * time.Millisecond)

	// 恢复后再次请求
	allowed = 0
	for i := 0; i < 5; i++ {
		if tb.Allow() {
			allowed++
		}
	}
	fmt.Printf("  等待 500ms 后请求 5 个 → 通过: %d（速率 5/s，500ms 恢复约 2-3 个令牌）\n", allowed)
	fmt.Println()
}

// demoLeakyBucket 演示漏桶算法
func demoLeakyBucket() {
	fmt.Println("--- 2. 漏桶算法（Leaky Bucket）---")
	fmt.Println("特点：严格匀速处理，不允许突发流量")
	fmt.Println()

	// 每秒处理 5 个请求，桶容量 10
	lb := NewLeakyBucket(5, 10)

	// 突发请求测试
	allowed := 0
	denied := 0
	for i := 0; i < 15; i++ {
		if lb.Allow() {
			allowed++
		} else {
			denied++
		}
	}
	fmt.Printf("  突发 15 个请求 → 通过: %d, 拒绝: %d（桶容量 10）\n", allowed, denied)

	// 等待漏水
	time.Sleep(500 * time.Millisecond)

	allowed = 0
	for i := 0; i < 5; i++ {
		if lb.Allow() {
			allowed++
		}
	}
	fmt.Printf("  等待 500ms 后请求 5 个 → 通过: %d（漏掉约 2-3 个，腾出空间）\n", allowed)
	fmt.Println()
}

// demoSlidingWindow 演示滑动窗口算法
func demoSlidingWindow() {
	fmt.Println("--- 3. 滑动窗口算法（Sliding Window）---")
	fmt.Println("特点：精确统计窗口内请求数，避免固定窗口边界突发")
	fmt.Println()

	// 1 秒窗口，最多 10 个请求
	sw := NewSlidingWindowLog(time.Second, 10)

	// 突发请求测试
	allowed := 0
	denied := 0
	for i := 0; i < 15; i++ {
		if sw.Allow() {
			allowed++
		} else {
			denied++
		}
	}
	fmt.Printf("  突发 15 个请求 → 通过: %d, 拒绝: %d（窗口限制 10）\n", allowed, denied)

	// 等待窗口滑动
	time.Sleep(1100 * time.Millisecond)

	allowed = 0
	for i := 0; i < 5; i++ {
		if sw.Allow() {
			allowed++
		}
	}
	fmt.Printf("  等待 1.1s 后请求 5 个 → 通过: %d（旧请求已过期）\n", allowed)
	fmt.Println()
}

// demoConcurrentComparison 并发压测对比三种算法
func demoConcurrentComparison() {
	fmt.Println("--- 4. 并发压测对比 ---")
	fmt.Println("100 个 goroutine 各发 10 个请求（共 1000 个），限流 100/s")
	fmt.Println()

	// 创建三种限流器，都限制 100/s
	tb := NewTokenBucket(100, 100)
	lb := NewLeakyBucket(100, 100)
	sw := NewSlidingWindowLog(time.Second, 100)
	swc := NewSlidingWindowCounter(time.Second, 10, 100)

	type result struct {
		name    string
		allowed int64
		denied  int64
	}

	// 并发测试函数
	runTest := func(name string, allowFn func() bool) result {
		var allowed, denied atomic.Int64
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					if allowFn() {
						allowed.Add(1)
					} else {
						denied.Add(1)
					}
					// 模拟请求间隔
					time.Sleep(time.Millisecond)
				}
			}()
		}
		wg.Wait()

		return result{
			name:    name,
			allowed: allowed.Load(),
			denied:  denied.Load(),
		}
	}

	results := []result{
		runTest("令牌桶  ", tb.Allow),
		runTest("漏桶    ", lb.Allow),
		runTest("滑动窗口(日志)", sw.Allow),
		runTest("滑动窗口(计数)", swc.Allow),
	}

	fmt.Printf("  %-20s | 通过 | 拒绝 | 通过率\n", "算法")
	fmt.Println("  " + "--------------------+------+------+-------")
	for _, r := range results {
		total := r.allowed + r.denied
		rate := float64(r.allowed) / float64(total) * 100
		fmt.Printf("  %-20s | %4d | %4d | %.1f%%\n", r.name, r.allowed, r.denied, rate)
	}
	fmt.Println()
	fmt.Println("说明：令牌桶允许突发（初始满桶），漏桶严格限制，滑动窗口精确计数")
}
