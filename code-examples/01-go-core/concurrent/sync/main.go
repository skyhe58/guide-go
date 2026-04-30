// 并发编程 — sync 包（Mutex/WaitGroup/Once/Pool）
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 sync 包的核心同步原语：
// 1. Mutex —— 互斥锁保护共享状态
// 2. RWMutex —— 读写锁（读多写少场景）
// 3. WaitGroup —— 等待一组 goroutine 完成
// 4. Once —— 确保操作只执行一次（单例模式）
// 5. Pool —— 临时对象池（减少 GC 压力）
// 6. 常见错误演示
//
// 运行方式：go run main.go
package main

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("========== sync 包示例 ==========")

	// --- 1. Mutex ---
	demoMutex()

	// --- 2. RWMutex ---
	demoRWMutex()

	// --- 3. WaitGroup ---
	demoWaitGroup()

	// --- 4. Once ---
	demoOnce()

	// --- 5. Pool ---
	demoPool()

	// --- 6. 常见错误 ---
	demoCommonMistakes()

	fmt.Println("\n========== 示例结束 ==========")
}

// ============================================================
// 1. Mutex —— 互斥锁
// ============================================================

// SafeCounter 使用 Mutex 保护的并发安全计数器
type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func demoMutex() {
	fmt.Println("\n--- 1. Mutex 互斥锁 ---")

	counter := &SafeCounter{}
	var wg sync.WaitGroup

	// 100 个 goroutine 并发递增
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}

	wg.Wait()
	fmt.Printf("  ✅ Mutex 保护的计数器: %d（期望 100）\n", counter.Value())

	// ❌ 对比：不使用锁的计数器（数据竞争）
	var unsafeCount int
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unsafeCount++ // ⚠️ 数据竞争！使用 go run -race 可检测
		}()
	}
	wg.Wait()
	fmt.Printf("  ❌ 无锁计数器: %d（可能小于 100，存在数据竞争）\n", unsafeCount)
}

// ============================================================
// 2. RWMutex —— 读写锁
// ============================================================

// ConfigCache 使用 RWMutex 的配置缓存（读多写少）
type ConfigCache struct {
	rw   sync.RWMutex
	data map[string]string
}

func NewConfigCache() *ConfigCache {
	return &ConfigCache{data: make(map[string]string)}
}

func (c *ConfigCache) Get(key string) (string, bool) {
	c.rw.RLock() // 读锁：多个 goroutine 可同时持有
	defer c.rw.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *ConfigCache) Set(key, value string) {
	c.rw.Lock() // 写锁：排他，阻塞所有读写
	defer c.rw.Unlock()
	c.data[key] = value
}

func demoRWMutex() {
	fmt.Println("\n--- 2. RWMutex 读写锁 ---")

	cache := NewConfigCache()
	cache.Set("db_host", "localhost")
	cache.Set("db_port", "5432")

	var wg sync.WaitGroup

	// 启动多个读 goroutine
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if val, ok := cache.Get("db_host"); ok {
				fmt.Printf("  读者 %d: db_host = %s\n", id, val)
			}
		}(i)
	}

	// 启动一个写 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		cache.Set("db_host", "192.168.1.100")
		fmt.Println("  写者: 更新 db_host")
	}()

	wg.Wait()
	val, _ := cache.Get("db_host")
	fmt.Printf("  最终值: db_host = %s\n", val)
}

// ============================================================
// 3. WaitGroup —— 等待一组 goroutine
// ============================================================

func demoWaitGroup() {
	fmt.Println("\n--- 3. WaitGroup ---")

	var wg sync.WaitGroup

	tasks := []string{"下载文件", "解析数据", "生成报告", "发送邮件"}

	for _, task := range tasks {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			time.Sleep(time.Duration(50) * time.Millisecond)
			fmt.Printf("  ✅ 任务完成: %s\n", t)
		}(task)
	}

	fmt.Println("  等待所有任务完成...")
	wg.Wait()
	fmt.Println("  所有任务已完成！")
}

// ============================================================
// 4. Once —— 单次执行（单例模式）
// ============================================================

// DBConnection 模拟数据库连接（单例）
type DBConnection struct {
	DSN string
}

var (
	dbInstance *DBConnection
	dbOnce     sync.Once
)

// GetDB 获取数据库连接单例
func GetDB() *DBConnection {
	dbOnce.Do(func() {
		fmt.Println("  初始化数据库连接...（只执行一次）")
		dbInstance = &DBConnection{DSN: "postgres://localhost:5432/mydb"}
	})
	return dbInstance
}

func demoOnce() {
	fmt.Println("\n--- 4. Once 单次执行 ---")

	var wg sync.WaitGroup

	// 多个 goroutine 同时获取单例
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			db := GetDB()
			fmt.Printf("  goroutine %d 获取到连接: %s\n", id, db.DSN)
		}(i)
	}

	wg.Wait()
	fmt.Println("  ✅ 初始化函数只执行了一次")
}

// ============================================================
// 5. Pool —— 临时对象池
// ============================================================

func demoPool() {
	fmt.Println("\n--- 5. Pool 对象池 ---")

	// 创建 Buffer 对象池
	bufPool := &sync.Pool{
		New: func() any {
			fmt.Println("  Pool.New: 创建新 Buffer")
			return new(bytes.Buffer)
		},
	}

	// 第一次 Get：池为空，调用 New
	buf1 := bufPool.Get().(*bytes.Buffer)
	buf1.WriteString("Hello, Pool!")
	fmt.Printf("  buf1 内容: %q\n", buf1.String())

	// 归还到池中（记得 Reset）
	buf1.Reset()
	bufPool.Put(buf1)

	// 第二次 Get：从池中复用
	buf2 := bufPool.Get().(*bytes.Buffer)
	fmt.Printf("  buf2 复用: 是同一个对象? %t\n", buf1 == buf2)

	// 性能对比：使用 Pool vs 每次 new
	var allocCount atomic.Int64
	poolWithCounter := &sync.Pool{
		New: func() any {
			allocCount.Add(1)
			return new(bytes.Buffer)
		},
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := poolWithCounter.Get().(*bytes.Buffer)
			buf.WriteString("test data")
			buf.Reset()
			poolWithCounter.Put(buf)
		}()
	}
	wg.Wait()
	fmt.Printf("  100 次操作，实际分配次数: %d（Pool 复用效果）\n", allocCount.Load())
}

// ============================================================
// 6. 常见错误演示
// ============================================================

func demoCommonMistakes() {
	fmt.Println("\n--- 6. 常见错误 ---")

	// ❌ 错误 1：Mutex 值拷贝
	fmt.Println("  ❌ Mutex 不可复制（值传递会导致锁失效）")
	fmt.Println("     应使用指针传递或嵌入结构体")

	// ❌ 错误 2：WaitGroup Add 位置错误
	fmt.Println("  ❌ wg.Add(1) 必须在 go 语句之前调用")
	fmt.Println("     否则 Wait 可能在 Add 之前返回")

	// ❌ 错误 3：忘记 Unlock
	fmt.Println("  ❌ 忘记 Unlock 导致死锁")
	fmt.Println("     建议: 始终使用 defer mu.Unlock()")

	// ✅ 正确模式：defer 确保 Unlock
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
	fmt.Println("  ✅ 使用 defer mu.Unlock() 确保锁释放")
}
