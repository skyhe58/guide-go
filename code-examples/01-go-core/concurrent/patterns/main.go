// 并发编程 — 并发模式（fan-in/fan-out/pipeline/worker-pool）
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 Go 经典并发模式：
// 1. Pipeline（管道模式）—— 多阶段数据处理
// 2. Fan-Out / Fan-In（扇出/扇入）—— 并行处理
// 3. Worker Pool（工作池）—— 控制并发度
// 4. Rate Limiting（限流）—— 控制操作频率
// 5. Or-Done 模式 —— 安全的 channel 读取
//
// 运行方式：go run main.go
package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	fmt.Println("========== 并发模式示例 ==========")

	// --- 1. Pipeline ---
	demoPipeline()

	// --- 2. Fan-Out / Fan-In ---
	demoFanOutFanIn()

	// --- 3. Worker Pool ---
	demoWorkerPool()

	// --- 4. Rate Limiting ---
	demoRateLimiting()

	// --- 5. Or-Done ---
	demoOrDone()

	fmt.Println("\n========== 示例结束 ==========")
}

// ============================================================
// 1. Pipeline（管道模式）
// ============================================================

// generator 生成数据（Pipeline 第一阶段）
func generator(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// square 计算平方（Pipeline 中间阶段）
func square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// double 计算两倍（Pipeline 中间阶段）
func double(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * 2:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func demoPipeline() {
	fmt.Println("\n--- 1. Pipeline 管道模式 ---")
	fmt.Println("  流程: 生成 → 平方 → 两倍")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 构建 Pipeline: generator → square → double
	nums := generator(ctx, 1, 2, 3, 4, 5)
	squared := square(ctx, nums)
	doubled := double(ctx, squared)

	// 消费最终结果
	fmt.Print("  结果: ")
	for val := range doubled {
		fmt.Printf("%d ", val) // (1²×2, 2²×2, 3²×2, 4²×2, 5²×2) = 2, 8, 18, 32, 50
	}
	fmt.Println()
}

// ============================================================
// 2. Fan-Out / Fan-In（扇出/扇入）
// ============================================================

// heavyCompute 模拟 CPU 密集型计算
func heavyCompute(ctx context.Context, in <-chan int, id int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case <-ctx.Done():
				return
			default:
				// 模拟耗时计算
				time.Sleep(10 * time.Millisecond)
				result := n * n
				fmt.Printf("  Worker %d: %d² = %d\n", id, n, result)
				out <- result
			}
		}
	}()
	return out
}

// fanIn 将多个 channel 合并为一个
func fanIn(ctx context.Context, channels ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for val := range c {
				select {
				case out <- val:
				case <-ctx.Done():
					return
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func demoFanOutFanIn() {
	fmt.Println("\n--- 2. Fan-Out / Fan-In ---")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 生成数据
	nums := generator(ctx, 1, 2, 3, 4, 5, 6)

	// Fan-Out: 3 个 worker 并行处理
	w1 := heavyCompute(ctx, nums, 1)
	w2 := heavyCompute(ctx, nums, 2)
	w3 := heavyCompute(ctx, nums, 3)

	// Fan-In: 合并结果
	results := fanIn(ctx, w1, w2, w3)

	fmt.Print("  合并结果: ")
	for val := range results {
		fmt.Printf("%d ", val)
	}
	fmt.Println()
}

// ============================================================
// 3. Worker Pool（工作池）
// ============================================================

// Job 表示一个任务
type Job struct {
	ID    int
	Input int
}

// Result 表示任务结果
type Result struct {
	Job    Job
	Output int
}

func demoWorkerPool() {
	fmt.Println("\n--- 3. Worker Pool 工作池 ---")

	const numWorkers = 3
	const numJobs = 10

	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	// 启动固定数量的 worker
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				// 模拟处理
				time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
				result := Result{
					Job:    job,
					Output: job.Input * job.Input,
				}
				fmt.Printf("  Worker %d 处理 Job %d: %d² = %d\n",
					workerID, job.ID, job.Input, result.Output)
				results <- result
			}
		}(w)
	}

	// 发送任务
	for j := 1; j <= numJobs; j++ {
		jobs <- Job{ID: j, Input: j}
	}
	close(jobs) // 关闭 jobs channel，worker 的 for-range 自动退出

	// 等待所有 worker 完成后关闭 results
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	var total int
	for r := range results {
		total += r.Output
	}
	fmt.Printf("  所有结果之和: %d\n", total)
}

// ============================================================
// 4. Rate Limiting（限流）
// ============================================================

func demoRateLimiting() {
	fmt.Println("\n--- 4. Rate Limiting 限流 ---")

	// 简单限流：每 50ms 处理一个请求
	requests := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		requests <- i
	}
	close(requests)

	limiter := time.NewTicker(50 * time.Millisecond)
	defer limiter.Stop()

	start := time.Now()
	for req := range requests {
		<-limiter.C // 等待令牌
		fmt.Printf("  处理请求 %d（耗时 %v）\n", req, time.Since(start).Round(time.Millisecond))
	}

	// 突发限流：允许短时间内的突发请求
	fmt.Println("  突发限流:")
	burstyLimiter := make(chan time.Time, 3) // 允许突发 3 个

	// 预填充令牌
	for i := 0; i < 3; i++ {
		burstyLimiter <- time.Now()
	}

	// 持续补充令牌
	go func() {
		for t := range time.Tick(100 * time.Millisecond) {
			burstyLimiter <- t
		}
	}()

	burstyRequests := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		burstyRequests <- i
	}
	close(burstyRequests)

	start = time.Now()
	for req := range burstyRequests {
		<-burstyLimiter
		fmt.Printf("  突发请求 %d（耗时 %v）\n", req, time.Since(start).Round(time.Millisecond))
	}
}

// ============================================================
// 5. Or-Done 模式
// ============================================================

// orDone 封装 channel 读取，自动处理 done 信号
func orDone(done <-chan struct{}, c <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-done:
					return
				}
			}
		}
	}()
	return out
}

func demoOrDone() {
	fmt.Println("\n--- 5. Or-Done 模式 ---")

	done := make(chan struct{})
	dataCh := make(chan int)

	// 生产者：持续发送数据
	go func() {
		defer close(dataCh)
		for i := 1; i <= 100; i++ {
			dataCh <- i
		}
	}()

	// 消费者：使用 orDone 安全读取，随时可取消
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(done) // 取消信号
	}()

	count := 0
	for val := range orDone(done, dataCh) {
		count++
		_ = val
	}
	fmt.Printf("  在取消前处理了 %d 个数据\n", count)
	fmt.Println("  ✅ orDone 模式确保 goroutine 在收到 done 信号后安全退出")
}
