// 缓存一致性方案对比 — 架构设计场景
// 本示例实现并对比三种缓存策略：Cache-Aside、Write-Through、Write-Behind
// 演示各策略在并发场景下的行为差异和一致性问题
// 使用纯 Go 实现，无需外部依赖
//
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：go run ./04-distributed/architecture/cache-consistency/

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// Part A：缓存一致性方案对比（纯内存模拟）
// ============================================================

// --- 基础组件：模拟缓存和数据库 ---

// SimulatedDB 模拟数据库
// 使用 sync.Map 模拟数据库存储，加入随机延迟模拟真实 IO
type SimulatedDB struct {
	data       sync.Map
	readCount  int64 // 读次数统计
	writeCount int64 // 写次数统计
	readDelay  time.Duration
	writeDelay time.Duration
}

// NewSimulatedDB 创建模拟数据库
func NewSimulatedDB(readDelay, writeDelay time.Duration) *SimulatedDB {
	return &SimulatedDB{
		readDelay:  readDelay,
		writeDelay: writeDelay,
	}
}

// Get 从数据库读取（模拟 IO 延迟）
func (db *SimulatedDB) Get(key string) (string, bool) {
	atomic.AddInt64(&db.readCount, 1)
	time.Sleep(db.readDelay) // 模拟数据库读延迟
	val, ok := db.data.Load(key)
	if !ok {
		return "", false
	}
	return val.(string), true
}

// Set 写入数据库（模拟 IO 延迟）
func (db *SimulatedDB) Set(key, value string) {
	atomic.AddInt64(&db.writeCount, 1)
	time.Sleep(db.writeDelay) // 模拟数据库写延迟
	db.data.Store(key, value)
}

// Stats 返回统计信息
func (db *SimulatedDB) Stats() (reads, writes int64) {
	return atomic.LoadInt64(&db.readCount), atomic.LoadInt64(&db.writeCount)
}

// SimulatedCache 模拟缓存（Redis）
// 使用 sync.Map 模拟 Redis，延迟远低于数据库
type SimulatedCache struct {
	data       sync.Map
	hitCount   int64 // 缓存命中次数
	missCount  int64 // 缓存未命中次数
	writeCount int64 // 缓存写入次数
	delCount   int64 // 缓存删除次数
}

// NewSimulatedCache 创建模拟缓存
func NewSimulatedCache() *SimulatedCache {
	return &SimulatedCache{}
}

// Get 从缓存读取
func (c *SimulatedCache) Get(key string) (string, bool) {
	val, ok := c.data.Load(key)
	if ok {
		atomic.AddInt64(&c.hitCount, 1)
		return val.(string), true
	}
	atomic.AddInt64(&c.missCount, 1)
	return "", false
}

// Set 写入缓存
func (c *SimulatedCache) Set(key, value string) {
	atomic.AddInt64(&c.writeCount, 1)
	c.data.Store(key, value)
}

// Delete 删除缓存
func (c *SimulatedCache) Delete(key string) {
	atomic.AddInt64(&c.delCount, 1)
	c.data.Delete(key)
}

// Stats 返回统计信息
func (c *SimulatedCache) Stats() (hits, misses, writes, deletes int64) {
	return atomic.LoadInt64(&c.hitCount),
		atomic.LoadInt64(&c.missCount),
		atomic.LoadInt64(&c.writeCount),
		atomic.LoadInt64(&c.delCount)
}

// --- 策略 1：Cache-Aside（旁路缓存） ---

// CacheAsideStore Cache-Aside 策略实现
// 读：先查缓存 → 未命中则查 DB → 回填缓存
// 写：先更新 DB → 再删除缓存（推荐方案）
type CacheAsideStore struct {
	cache *SimulatedCache
	db    *SimulatedDB
}

// NewCacheAsideStore 创建 Cache-Aside 存储
func NewCacheAsideStore(cache *SimulatedCache, db *SimulatedDB) *CacheAsideStore {
	return &CacheAsideStore{cache: cache, db: db}
}

// Get 读取数据（Cache-Aside 读流程）
// 1. 先查缓存
// 2. 缓存未命中，查数据库
// 3. 将数据库结果回填到缓存
func (s *CacheAsideStore) Get(key string) (string, bool) {
	// 第一步：查缓存
	if val, ok := s.cache.Get(key); ok {
		return val, true // 缓存命中，直接返回
	}

	// 第二步：缓存未命中，查数据库
	val, ok := s.db.Get(key)
	if !ok {
		return "", false // 数据库也没有
	}

	// 第三步：回填缓存
	s.cache.Set(key, val)
	return val, true
}

// Set 写入数据（Cache-Aside 写流程）
// 推荐策略：先更新数据库，再删除缓存
// 为什么删除而不是更新缓存？因为并发写时更新缓存可能导致数据覆盖
func (s *CacheAsideStore) Set(key, value string) {
	// 第一步：先更新数据库
	s.db.Set(key, value)
	// 第二步：再删除缓存（而非更新缓存）
	s.cache.Delete(key)
}

// --- 策略 2：Write-Through（写穿透） ---

// WriteThroughStore Write-Through 策略实现
// 读：直接读缓存（缓存始终有最新数据）
// 写：先写缓存 → 缓存同步写 DB（保证强一致）
type WriteThroughStore struct {
	cache *SimulatedCache
	db    *SimulatedDB
}

// NewWriteThroughStore 创建 Write-Through 存储
func NewWriteThroughStore(cache *SimulatedCache, db *SimulatedDB) *WriteThroughStore {
	return &WriteThroughStore{cache: cache, db: db}
}

// Get 读取数据（Write-Through 读流程）
// 缓存中始终有最新数据，直接读缓存
// 缓存未命中时从 DB 加载
func (s *WriteThroughStore) Get(key string) (string, bool) {
	if val, ok := s.cache.Get(key); ok {
		return val, true
	}
	// 缓存未命中，从 DB 加载
	val, ok := s.db.Get(key)
	if !ok {
		return "", false
	}
	s.cache.Set(key, val)
	return val, true
}

// Set 写入数据（Write-Through 写流程）
// 先写缓存，再同步写数据库
// 优点：缓存始终是最新数据
// 缺点：写性能低（需要等待 DB 写入完成）
func (s *WriteThroughStore) Set(key, value string) {
	// 第一步：写缓存
	s.cache.Set(key, value)
	// 第二步：同步写数据库（阻塞等待）
	s.db.Set(key, value)
}

// --- 策略 3：Write-Behind（写回/异步写） ---

// WriteOp 写操作记录
type WriteOp struct {
	Key   string
	Value string
}

// WriteBehindStore Write-Behind 策略实现
// 读：直接读缓存
// 写：只写缓存，异步批量刷盘到 DB
// 优点：写性能极高
// 缺点：可能丢数据（缓存宕机时未刷盘的数据丢失）
type WriteBehindStore struct {
	cache         *SimulatedCache
	db            *SimulatedDB
	writeQueue    chan WriteOp
	flushInterval time.Duration
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

// NewWriteBehindStore 创建 Write-Behind 存储
func NewWriteBehindStore(cache *SimulatedCache, db *SimulatedDB, flushInterval time.Duration) *WriteBehindStore {
	s := &WriteBehindStore{
		cache:         cache,
		db:            db,
		writeQueue:    make(chan WriteOp, 1000),
		flushInterval: flushInterval,
		stopCh:        make(chan struct{}),
	}
	s.startFlusher()
	return s
}

// startFlusher 启动后台刷盘协程
func (s *WriteBehindStore) startFlusher() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.flushInterval)
		defer ticker.Stop()

		var batch []WriteOp
		for {
			select {
			case op := <-s.writeQueue:
				batch = append(batch, op)
				// 批量达到阈值时立即刷盘
				if len(batch) >= 10 {
					s.flushBatch(batch)
					batch = nil
				}
			case <-ticker.C:
				// 定时刷盘
				if len(batch) > 0 {
					s.flushBatch(batch)
					batch = nil
				}
			case <-s.stopCh:
				// 停止前刷盘剩余数据
				close(s.writeQueue)
				for op := range s.writeQueue {
					batch = append(batch, op)
				}
				if len(batch) > 0 {
					s.flushBatch(batch)
				}
				return
			}
		}
	}()
}

// flushBatch 批量写入数据库
func (s *WriteBehindStore) flushBatch(batch []WriteOp) {
	// 去重：相同 key 只保留最后一次写入
	latest := make(map[string]string)
	for _, op := range batch {
		latest[op.Key] = op.Value
	}
	for key, value := range latest {
		s.db.Set(key, value)
	}
}

// Get 读取数据（Write-Behind 读流程）
func (s *WriteBehindStore) Get(key string) (string, bool) {
	if val, ok := s.cache.Get(key); ok {
		return val, true
	}
	val, ok := s.db.Get(key)
	if !ok {
		return "", false
	}
	s.cache.Set(key, val)
	return val, true
}

// Set 写入数据（Write-Behind 写流程）
// 只写缓存，异步写数据库
// 写性能极高，但有数据丢失风险
func (s *WriteBehindStore) Set(key, value string) {
	// 第一步：写缓存（立即返回）
	s.cache.Set(key, value)
	// 第二步：加入异步写队列
	select {
	case s.writeQueue <- WriteOp{Key: key, Value: value}:
	default:
		// 队列满，直接同步写 DB（降级处理）
		s.db.Set(key, value)
	}
}

// Stop 停止 Write-Behind 刷盘协程
func (s *WriteBehindStore) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

// --- 演示：缓存失效竞态条件 ---

// demonstrateCacheRace 演示"先删缓存再更新 DB"的竞态问题
// 这是一个反面教材，展示为什么不推荐这种策略
func demonstrateCacheRace() {
	fmt.Println("--- 4. 缓存失效竞态条件演示（反面教材） ---")
	fmt.Println("  策略：先删缓存，再更新 DB（不推荐！）")
	fmt.Println()

	cache := NewSimulatedCache()
	db := NewSimulatedDB(5*time.Millisecond, 10*time.Millisecond)

	// 初始数据
	db.Set("price", "100")
	cache.Set("price", "100")

	var wg sync.WaitGroup
	results := make([]string, 2)

	// 线程 A：写操作（先删缓存，再更新 DB）
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 第一步：删除缓存
		cache.Delete("price")
		// 模拟网络延迟
		time.Sleep(15 * time.Millisecond)
		// 第二步：更新数据库
		db.Set("price", "200")
		results[0] = "写线程：缓存已删除，DB 已更新为 200"
	}()

	// 线程 B：读操作（在写操作的两步之间执行）
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Millisecond) // 在删缓存之后、更新 DB 之前
		// 缓存未命中
		if _, ok := cache.Get("price"); !ok {
			// 从 DB 读取（此时 DB 还是旧值 100）
			val, _ := db.Get("price")
			// 回填缓存（写入旧值！）
			cache.Set("price", val)
			results[1] = fmt.Sprintf("读线程：缓存未命中，从 DB 读到 %s，回填缓存", val)
		}
	}()

	wg.Wait()

	// 检查最终状态
	cacheVal, _ := cache.Get("price")
	dbVal, _ := db.Get("price")

	fmt.Printf("  %s\n", results[0])
	fmt.Printf("  %s\n", results[1])
	fmt.Printf("  最终状态 → 缓存: %s, DB: %s\n", cacheVal, dbVal)
	if cacheVal != dbVal {
		fmt.Println("  ⚠️  数据不一致！缓存中是旧值，DB 中是新值")
		fmt.Println("  这就是为什么推荐「先更新 DB，再删缓存」的原因")
	}
	fmt.Println()
}

func main() {
	fmt.Println("========== 缓存一致性方案对比演示 ==========")
	fmt.Println()

	// --- 演示 1：Cache-Aside 策略 ---
	fmt.Println("--- 1. Cache-Aside（旁路缓存）策略 ---")
	fmt.Println("  读：先查缓存 → 未命中查 DB → 回填缓存")
	fmt.Println("  写：先更新 DB → 再删除缓存")
	fmt.Println()

	caCache := NewSimulatedCache()
	caDB := NewSimulatedDB(2*time.Millisecond, 5*time.Millisecond)
	caStore := NewCacheAsideStore(caCache, caDB)

	// 写入数据
	caStore.Set("user:1", "Alice")
	caStore.Set("user:2", "Bob")

	// 第一次读取（缓存未命中，从 DB 加载）
	val, _ := caStore.Get("user:1")
	fmt.Printf("  第一次读 user:1: %s（缓存未命中，从 DB 加载）\n", val)

	// 第二次读取（缓存命中）
	val, _ = caStore.Get("user:1")
	fmt.Printf("  第二次读 user:1: %s（缓存命中）\n", val)

	// 更新数据
	caStore.Set("user:1", "Alice_Updated")
	val, _ = caStore.Get("user:1")
	fmt.Printf("  更新后读 user:1: %s（缓存已删除，重新从 DB 加载）\n", val)

	hits, misses, _, deletes := caCache.Stats()
	dbReads, dbWrites := caDB.Stats()
	fmt.Printf("  统计 → 缓存命中: %d, 未命中: %d, 删除: %d | DB 读: %d, 写: %d\n",
		hits, misses, deletes, dbReads, dbWrites)
	fmt.Println()

	// --- 演示 2：Write-Through 策略 ---
	fmt.Println("--- 2. Write-Through（写穿透）策略 ---")
	fmt.Println("  读：直接读缓存")
	fmt.Println("  写：先写缓存 → 同步写 DB")
	fmt.Println()

	wtCache := NewSimulatedCache()
	wtDB := NewSimulatedDB(2*time.Millisecond, 5*time.Millisecond)
	wtStore := NewWriteThroughStore(wtCache, wtDB)

	// 写入数据（同时写缓存和 DB）
	start := time.Now()
	wtStore.Set("product:1", "iPhone")
	wtStore.Set("product:2", "MacBook")
	writeTime := time.Since(start)

	// 读取（直接从缓存返回）
	val, _ = wtStore.Get("product:1")
	fmt.Printf("  读 product:1: %s（缓存命中）\n", val)

	hits, _, writes, _ := wtCache.Stats()
	_, dbWrites = wtDB.Stats()
	fmt.Printf("  统计 → 缓存命中: %d, 缓存写: %d, DB 写: %d\n", hits, writes, dbWrites)
	fmt.Printf("  写入耗时: %v（同步写 DB，较慢）\n", writeTime)
	fmt.Println()

	// --- 演示 3：Write-Behind 策略 ---
	fmt.Println("--- 3. Write-Behind（异步写回）策略 ---")
	fmt.Println("  读：直接读缓存")
	fmt.Println("  写：只写缓存 → 异步批量刷盘到 DB")
	fmt.Println()

	wbCache := NewSimulatedCache()
	wbDB := NewSimulatedDB(2*time.Millisecond, 5*time.Millisecond)
	wbStore := NewWriteBehindStore(wbCache, wbDB, 50*time.Millisecond)

	// 写入数据（只写缓存，立即返回）
	start = time.Now()
	for i := 0; i < 20; i++ {
		wbStore.Set(fmt.Sprintf("item:%d", i), fmt.Sprintf("value_%d", i))
	}
	writeTime = time.Since(start)
	fmt.Printf("  写入 20 条数据耗时: %v（只写缓存，极快）\n", writeTime)

	// 此时 DB 可能还没有数据（异步刷盘未完成）
	_, dbWrites = wbDB.Stats()
	fmt.Printf("  写入后立即检查 DB 写次数: %d（异步刷盘可能未完成）\n", dbWrites)

	// 等待异步刷盘完成
	time.Sleep(100 * time.Millisecond)
	_, dbWrites = wbDB.Stats()
	fmt.Printf("  等待 100ms 后 DB 写次数: %d（异步刷盘完成）\n", dbWrites)

	// 读取验证
	val, _ = wbStore.Get("item:5")
	fmt.Printf("  读 item:5: %s\n", val)

	wbStore.Stop() // 停止刷盘协程
	fmt.Println()

	// --- 演示 4：缓存失效竞态条件 ---
	demonstrateCacheRace()

	// --- 演示 5：并发场景下的 Cache-Aside ---
	fmt.Println("--- 5. 并发场景下的 Cache-Aside 性能测试 ---")
	concCache := NewSimulatedCache()
	concDB := NewSimulatedDB(1*time.Millisecond, 2*time.Millisecond)
	concStore := NewCacheAsideStore(concCache, concDB)

	// 预热数据
	for i := 0; i < 100; i++ {
		concDB.Set(fmt.Sprintf("key:%d", i), fmt.Sprintf("value_%d", i))
	}

	// 并发读写测试
	var wg sync.WaitGroup
	readOps := int64(0)
	writeOps := int64(0)

	start = time.Now()

	// 80% 读操作
	for i := 0; i < 80; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key:%d", rand.Intn(100))
				concStore.Get(key)
				atomic.AddInt64(&readOps, 1)
			}
		}(i)
	}

	// 20% 写操作
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key:%d", rand.Intn(100))
				value := fmt.Sprintf("updated_%d_%d", id, j)
				concStore.Set(key, value)
				atomic.AddInt64(&writeOps, 1)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	hits, misses, _, deletes = concCache.Stats()
	dbReads, dbWrites = concDB.Stats()
	hitRate := float64(hits) / float64(hits+misses) * 100

	fmt.Printf("  并发测试完成（耗时 %v）\n", elapsed)
	fmt.Printf("  读操作: %d, 写操作: %d\n", readOps, writeOps)
	fmt.Printf("  缓存命中: %d, 未命中: %d, 命中率: %.1f%%\n", hits, misses, hitRate)
	fmt.Printf("  缓存删除: %d（写操作触发）\n", deletes)
	fmt.Printf("  DB 读: %d, DB 写: %d\n", dbReads, dbWrites)
	fmt.Println()

	// --- 策略对比总结 ---
	fmt.Println("========== 策略对比总结 ==========")
	fmt.Println()
	fmt.Println("  | 策略           | 读性能 | 写性能 | 一致性   | 适用场景         |")
	fmt.Println("  |----------------|--------|--------|----------|------------------|")
	fmt.Println("  | Cache-Aside    | 高     | 中     | 最终一致 | 通用场景（推荐） |")
	fmt.Println("  | Write-Through  | 高     | 低     | 强一致   | 一致性要求高     |")
	fmt.Println("  | Write-Behind   | 高     | 极高   | 弱一致   | 写密集场景       |")
	fmt.Println()
	fmt.Println("  推荐：大多数场景使用 Cache-Aside + 先更新 DB 再删缓存")
}
