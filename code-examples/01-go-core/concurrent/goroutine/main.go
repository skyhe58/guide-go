// 并发编程 — goroutine（创建/泄漏检测）
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 goroutine 的核心用法：
// 1. goroutine 的创建和基本使用
// 2. 使用 WaitGroup 等待 goroutine 完成
// 3. goroutine 泄漏的常见原因和检测方法
// 4. 使用 context 预防 goroutine 泄漏
//
// 运行方式：go run main.go
// 竞争检测：go run -race main.go
package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	fmt.Println("========== goroutine 示例 ==========")

	// --- 1. 基本创建 ---
	demoBasicGoroutine()

	// --- 2. WaitGroup 等待 ---
	demoWaitGroup()

	// --- 3. 泄漏检测 ---
	demoLeakDetection()

	// --- 4. 使用 context 预防泄漏 ---
	demoContextPrevention()

	// --- 5. 常见错误演示 ---
	demoCommonMistakes()

	fmt.Println("\n========== 示例结束 ==========")
}

// ============================================================
// 1. goroutine 基本创建
// ============================================================

func demoBasicGoroutine() {
	fmt.Println("\n--- 1. goroutine 基本创建 ---")

	// 最简单的 goroutine
	go func() {
		fmt.Println("  Hello from goroutine!")
	}()

	// 带参数的 goroutine
	go func(name string) {
		fmt.Printf("  Hello, %s!\n", name)
	}("Go 并发")

	// 等待 goroutine 执行（生产环境应使用 WaitGroup）
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("  当前 goroutine 数量: %d\n", runtime.NumGoroutine())
}

// ============================================================
// 2. 使用 WaitGroup 等待 goroutine 完成
// ============================================================

func demoWaitGroup() {
	fmt.Println("\n--- 2. WaitGroup 等待 goroutine ---")

	var wg sync.WaitGroup

	// ✅ 正确：Add 在 go 语句之前调用
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(time.Duration(id*10) * time.Millisecond)
			fmt.Printf("  Worker %d 完成\n", id)
		}(i)
	}

	wg.Wait()
	fmt.Println("  所有 Worker 已完成")
}

// ============================================================
// 3. goroutine 泄漏检测
// ============================================================

func demoLeakDetection() {
	fmt.Println("\n--- 3. goroutine 泄漏检测 ---")

	before := runtime.NumGoroutine()
	fmt.Printf("  泄漏前 goroutine 数量: %d\n", before)

	// ⚠️ 演示泄漏：启动一个永远阻塞的 goroutine
	// 生产环境中这是一个 bug！
	leakyCh := make(chan int) // 无缓冲 channel，没有发送方
	go func() {
		// 这个 goroutine 会永远阻塞在这里，因为没有人向 leakyCh 发送数据
		val := <-leakyCh
		fmt.Println("  永远不会执行:", val)
	}()

	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()
	fmt.Printf("  泄漏后 goroutine 数量: %d（增加了 %d 个）\n", after, after-before)
	fmt.Println("  ⚠️ 泄漏的 goroutine 会一直占用内存，直到程序退出")

	// 注意：这里故意不清理泄漏的 goroutine，用于演示
	_ = leakyCh
}

// ============================================================
// 4. 使用 context 预防 goroutine 泄漏
// ============================================================

func demoContextPrevention() {
	fmt.Println("\n--- 4. 使用 context 预防泄漏 ---")

	before := runtime.NumGoroutine()

	// ✅ 正确：使用 context 控制 goroutine 生命周期
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	resultCh := make(chan string, 1)

	go func() {
		// 模拟耗时操作
		select {
		case <-time.After(100 * time.Millisecond):
			resultCh <- "任务完成"
		case <-ctx.Done():
			fmt.Println("  goroutine 收到取消信号，正常退出")
			return
		}
	}()

	select {
	case result := <-resultCh:
		fmt.Printf("  收到结果: %s\n", result)
	case <-ctx.Done():
		fmt.Println("  超时，取消任务")
	}

	// 等待 goroutine 退出
	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()
	fmt.Printf("  context 控制后 goroutine 数量: %d（变化: %d）\n", after, after-before)
}

// ============================================================
// 5. 常见错误演示
// ============================================================

func demoCommonMistakes() {
	fmt.Println("\n--- 5. 常见错误演示 ---")

	// ❌ 错误：Go 1.21 及之前版本的闭包变量捕获问题
	// Go 1.22 已修复 for 循环变量语义，每次迭代创建新变量
	fmt.Println("  闭包变量捕获（Go 1.22 已修复）:")
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Go 1.22+：每次迭代 i 是新变量，输出 0, 1, 2（顺序不定）
			// Go 1.21-：所有 goroutine 共享同一个 i，可能输出 3, 3, 3
			fmt.Printf("    i = %d\n", i)
		}()
	}
	wg.Wait()

	// ✅ 兼容旧版本的正确写法：通过参数传递
	fmt.Println("  通过参数传递（兼容所有版本）:")
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("    id = %d\n", id)
		}(i)
	}
	wg.Wait()

	// ❌ 错误：主 goroutine 提前退出
	// 如果 main 函数没有等待机制，子 goroutine 可能来不及执行
	fmt.Println("  ⚠️ 注意：main 返回时所有 goroutine 立即终止")
}
