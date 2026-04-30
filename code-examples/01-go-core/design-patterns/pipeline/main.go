// Pipeline Pattern — Go 并发流水线模式
// 演示 Pipeline、Generator、Fan-Out/Fan-In 三种 Go 特有的并发模式
//
// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
//
// 运行方式: go run ./pipeline/

package main

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// =============================================================================
// Part A: Pipeline Pattern 完整演示
// =============================================================================

// --- 1. 基础 Pipeline ---

// generate 是 Pipeline 的第一个阶段：生成数据
func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// square 是 Pipeline 的中间阶段：计算平方
func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

// filterEven 是 Pipeline 的过滤阶段：只保留偶数
func filterEven(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if n%2 == 0 {
				out <- n
			}
		}
	}()
	return out
}

// sum 是 Pipeline 的最终阶段：求和
func sum(in <-chan int) int {
	total := 0
	for n := range in {
		total += n
	}
	return total
}

// --- 2. 带 Context 取消的 Pipeline ---

// generateWithCtx 支持 context 取消的生成器
func generateWithCtx(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				fmt.Println("  [生成器] 收到取消信号，停止生成")
				return
			}
		}
	}()
	return out
}

// squareWithCtx 支持 context 取消的平方计算
func squareWithCtx(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-ctx.Done():
				fmt.Println("  [平方] 收到取消信号，停止处理")
				return
			}
		}
	}()
	return out
}

// --- 3. Fan-Out / Fan-In ---

// isPrime 判断是否为素数（模拟耗时计算）
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i <= int(math.Sqrt(float64(n))); i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// primeWorker 素数检测 worker（Fan-Out 的每个 worker）
func primeWorker(id int, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if isPrime(n) {
				fmt.Printf("  [Worker %d] 发现素数: %d\n", id, n)
				out <- n
			}
		}
	}()
	return out
}

// fanIn 合并多个 channel 的结果（Fan-In）
func fanIn(channels ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	merged := make(chan int)

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				merged <- v
			}
		}(ch)
	}

	// 所有输入 channel 关闭后，关闭输出 channel
	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

func main() {
	fmt.Println("=== Go Pipeline Pattern 演示 ===")
	fmt.Println()

	// --- 1. 基础 Pipeline ---
	fmt.Println("--- 1. 基础 Pipeline: 生成 → 平方 → 过滤偶数 → 求和 ---")
	// Pipeline: generate → square → filterEven → sum
	// 输入: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10
	// 平方: 1, 4, 9, 16, 25, 36, 49, 64, 81, 100
	// 过滤偶数: 4, 16, 36, 64, 100
	// 求和: 220
	result := sum(filterEven(square(generate(1, 2, 3, 4, 5, 6, 7, 8, 9, 10))))
	fmt.Printf("结果: %d (预期: 220)\n", result)
	fmt.Println()

	// --- 2. 带 Context 取消的 Pipeline ---
	fmt.Println("--- 2. 带 Context 取消的 Pipeline ---")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 生成大量数据，但通过 context 提前取消
	nums := make([]int, 1000)
	for i := range nums {
		nums[i] = i + 1
	}

	count := 0
	for n := range squareWithCtx(ctx, generateWithCtx(ctx, nums...)) {
		count++
		_ = n
	}
	fmt.Printf("在超时前处理了 %d 个数据\n", count)
	fmt.Println()

	// --- 3. Fan-Out / Fan-In ---
	fmt.Println("--- 3. Fan-Out / Fan-In: 多 Worker 并行查找素数 ---")

	// 生成待检测的数字
	numbers := generate(2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30)

	// Fan-Out: 启动 3 个 worker 并行处理
	workerCount := 3
	workers := make([]<-chan int, workerCount)
	for i := 0; i < workerCount; i++ {
		workers[i] = primeWorker(i+1, numbers)
	}

	// Fan-In: 合并所有 worker 的结果
	primes := make([]int, 0)
	for p := range fanIn(workers...) {
		primes = append(primes, p)
	}
	fmt.Printf("\n找到 %d 个素数: %v\n", len(primes), primes)
	fmt.Println()

	// --- 总结 ---
	fmt.Println("=== Pipeline 模式要点 ===")
	fmt.Println("1. 每个阶段是一个 goroutine，通过 channel 连接")
	fmt.Println("2. 每个阶段负责关闭自己的输出 channel")
	fmt.Println("3. 使用 context 支持优雅取消")
	fmt.Println("4. Fan-Out 分发任务，Fan-In 合并结果")
	fmt.Println("5. 实际应用: Docker 镜像构建、K8s 准入控制、数据 ETL")
}
