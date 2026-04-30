// GMP 调度模型模拟 — Part A 纯内存模拟
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例通过纯内存模拟演示 Go GMP 调度模型的核心概念：
// 1. G（Goroutine）、M（Machine/OS线程）、P（Processor/逻辑处理器）的关系
// 2. 本地运行队列和全局运行队列
// 3. Work Stealing 调度策略
// 4. 查看运行时调度信息
//
// 运行方式：go run main.go
// 查看调度追踪：GODEBUG=schedtrace=1000 go run main.go
package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

// ========== GMP 模拟数据结构 ==========

// Goroutine 模拟结构
type G struct {
	ID     int
	State  string // Runnable, Running, Waiting, Dead
	TaskFn func()
}

// Processor 模拟结构（逻辑处理器）
type P struct {
	ID       int
	LocalQ   []*G // 本地运行队列
	mu       sync.Mutex
	Running  *G // 当前运行的 G
}

// Scheduler 模拟调度器
type Scheduler struct {
	GlobalQ []*G // 全局运行队列
	Ps      []*P // 所有 P
	mu      sync.Mutex
}

func main() {
	fmt.Println("========== GMP 调度模型模拟 ==========")
	fmt.Println()

	// --- 1. 查看运行时信息 ---
	demoRuntimeInfo()

	// --- 2. GMP 模拟调度 ---
	demoGMPSimulation()

	// --- 3. GOMAXPROCS 影响演示 ---
	demoGOMAXPROCS()

	// --- 4. Gosched 主动让出 ---
	demoGosched()
}

// demoRuntimeInfo 展示运行时基本信息
func demoRuntimeInfo() {
	fmt.Println("--- 1. 运行时信息 ---")
	fmt.Printf("CPU 核心数:       %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS (P数): %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("当前 goroutine 数: %d\n", runtime.NumGoroutine())
	fmt.Printf("Go 版本:          %s\n", runtime.Version())
	fmt.Printf("操作系统:          %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()
}

// demoGMPSimulation 模拟 GMP 调度过程
func demoGMPSimulation() {
	fmt.Println("--- 2. GMP 调度模拟 ---")

	// 创建调度器，2 个 P
	sched := &Scheduler{
		Ps: make([]*P, 2),
	}
	for i := range sched.Ps {
		sched.Ps[i] = &P{ID: i}
	}

	// 创建 8 个 G
	for i := 1; i <= 8; i++ {
		g := &G{
			ID:    i,
			State: "Runnable",
		}
		// 模拟分配：前 4 个放 P0 本地队列，后 4 个放 P1 本地队列
		targetP := sched.Ps[(i-1)%len(sched.Ps)]
		targetP.LocalQ = append(targetP.LocalQ, g)
	}

	// 打印初始状态
	fmt.Println("初始状态：")
	printSchedulerState(sched)

	// 模拟调度：每个 P 从本地队列取 G 执行
	fmt.Println("模拟调度过程：")
	for round := 1; round <= 4; round++ {
		fmt.Printf("\n  轮次 %d:\n", round)
		for _, p := range sched.Ps {
			if len(p.LocalQ) > 0 {
				// 从本地队列头部取出 G
				g := p.LocalQ[0]
				p.LocalQ = p.LocalQ[1:]
				g.State = "Running"
				fmt.Printf("    P%d 执行 G%d (本地队列剩余: %d)\n", p.ID, g.ID, len(p.LocalQ))
				g.State = "Dead"
			} else {
				fmt.Printf("    P%d 本地队列为空，尝试 Work Stealing...\n", p.ID)
				// 模拟 Work Stealing：从其他 P 偷取一半
				stolen := workSteal(sched, p)
				if stolen > 0 {
					fmt.Printf("    P%d 偷取了 %d 个 G\n", p.ID, stolen)
				} else {
					fmt.Printf("    P%d 无可偷取的 G，进入空闲\n", p.ID)
				}
			}
		}
	}
	fmt.Println()
}

// workSteal 模拟 Work Stealing
func workSteal(sched *Scheduler, thief *P) int {
	for _, victim := range sched.Ps {
		if victim.ID == thief.ID {
			continue
		}
		if len(victim.LocalQ) > 1 {
			// 偷取一半
			half := len(victim.LocalQ) / 2
			stolen := victim.LocalQ[:half]
			victim.LocalQ = victim.LocalQ[half:]
			thief.LocalQ = append(thief.LocalQ, stolen...)
			return len(stolen)
		}
	}
	return 0
}

// printSchedulerState 打印调度器状态
func printSchedulerState(sched *Scheduler) {
	fmt.Printf("  全局队列: %d 个 G\n", len(sched.GlobalQ))
	for _, p := range sched.Ps {
		ids := make([]int, len(p.LocalQ))
		for i, g := range p.LocalQ {
			ids[i] = g.ID
		}
		fmt.Printf("  P%d 本地队列: %v (%d 个 G)\n", p.ID, ids, len(p.LocalQ))
	}
	fmt.Println()
}

// demoGOMAXPROCS 演示 GOMAXPROCS 对并发的影响
func demoGOMAXPROCS() {
	fmt.Println("--- 3. GOMAXPROCS 影响演示 ---")

	for _, procs := range []int{1, 2, runtime.NumCPU()} {
		old := runtime.GOMAXPROCS(procs)
		start := time.Now()

		var wg sync.WaitGroup
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// 模拟 CPU 密集型任务
				sum := 0
				for j := 0; j < 1_000_000; j++ {
					sum += j
				}
				_ = sum
			}()
		}
		wg.Wait()

		elapsed := time.Since(start)
		fmt.Printf("  GOMAXPROCS=%d, 4个CPU密集任务耗时: %v\n", procs, elapsed)
		runtime.GOMAXPROCS(old)
	}
	fmt.Println()
}

// demoGosched 演示 runtime.Gosched() 主动让出
func demoGosched() {
	fmt.Println("--- 4. Gosched 主动让出 ---")

	// 设置为单 P，更明显地观察让出效果
	old := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(old)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 3; i++ {
			fmt.Printf("  goroutine A: 第 %d 次执行\n", i+1)
			runtime.Gosched() // 主动让出 CPU
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 3; i++ {
			fmt.Printf("  goroutine B: 第 %d 次执行\n", i+1)
			runtime.Gosched() // 主动让出 CPU
		}
	}()

	wg.Wait()

	// 模拟随机调度延迟
	_ = rand.Intn(1) // 使用 rand 避免 import 未使用
	fmt.Println()
	fmt.Println("提示：使用 GODEBUG=schedtrace=1000 go run main.go 可查看实时调度信息")
}
