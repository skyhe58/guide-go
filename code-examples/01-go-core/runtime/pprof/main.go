// pprof 性能分析示例 — Part A 纯内存模拟
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 Go pprof 性能分析工具的使用：
// 1. runtime/pprof 手动采集 CPU 和内存 profile
// 2. 模拟 CPU 密集型和内存密集型场景
// 3. 查看运行时内存统计
// 4. goroutine 信息采集
//
// 运行方式：go run main.go
// 分析 CPU profile：go tool pprof cpu.prof
// 分析内存 profile：go tool pprof mem.prof
package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
)

func main() {
	fmt.Println("========== pprof 性能分析示例 ==========")
	fmt.Println()

	// --- 1. CPU Profile 采集 ---
	demoCPUProfile()

	// --- 2. 内存 Profile 采集 ---
	demoMemProfile()

	// --- 3. Goroutine Profile ---
	demoGoroutineProfile()

	// --- 4. 内存统计信息 ---
	demoMemStats()

	fmt.Println("========== 分析命令 ==========")
	fmt.Println("  CPU 分析:    go tool pprof cpu.prof")
	fmt.Println("  内存分析:    go tool pprof mem.prof")
	fmt.Println("  交互命令:    top, list funcName, web, png")
	fmt.Println()
	fmt.Println("  HTTP 服务集成方式：")
	fmt.Println("    import _ \"net/http/pprof\"")
	fmt.Println("    go http.ListenAndServe(\":6060\", nil)")
	fmt.Println("    然后访问: http://localhost:6060/debug/pprof/")
}

// demoCPUProfile 采集 CPU profile
func demoCPUProfile() {
	fmt.Println("--- 1. CPU Profile 采集 ---")

	// 创建 CPU profile 文件
	f, err := os.Create("cpu.prof")
	if err != nil {
		fmt.Printf("  创建 cpu.prof 失败: %v\n", err)
		return
	}
	defer f.Close()

	// 开始 CPU 采集
	if err := pprof.StartCPUProfile(f); err != nil {
		fmt.Printf("  启动 CPU profile 失败: %v\n", err)
		return
	}

	// 模拟 CPU 密集型任务
	fmt.Println("  执行 CPU 密集型任务...")
	cpuIntensiveWork()

	// 停止采集
	pprof.StopCPUProfile()
	fmt.Println("  CPU profile 已保存到 cpu.prof")
	fmt.Println()
}

// cpuIntensiveWork 模拟 CPU 密集型工作
func cpuIntensiveWork() {
	// 计算大量浮点运算
	result := 0.0
	for i := 0; i < 1_000_000; i++ {
		result += math.Sqrt(float64(i))
		result += math.Sin(float64(i))
	}
	fmt.Printf("  计算结果: %.2f\n", result)
}

// demoMemProfile 采集内存 profile
func demoMemProfile() {
	fmt.Println("--- 2. 内存 Profile 采集 ---")

	// 模拟内存密集型任务
	fmt.Println("  执行内存密集型任务...")
	memIntensiveWork()

	// 强制 GC，确保内存统计准确
	runtime.GC()

	// 写入内存 profile
	f, err := os.Create("mem.prof")
	if err != nil {
		fmt.Printf("  创建 mem.prof 失败: %v\n", err)
		return
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		fmt.Printf("  写入内存 profile 失败: %v\n", err)
		return
	}
	fmt.Println("  内存 profile 已保存到 mem.prof")
	fmt.Println()
}

// memIntensiveWork 模拟内存密集型工作
func memIntensiveWork() {
	// 分配大量 slice
	data := make([][]byte, 100)
	for i := range data {
		data[i] = make([]byte, 10*1024) // 每个 10KB
	}

	// 模拟字符串拼接（低效方式，用于演示内存分配）
	s := ""
	for i := 0; i < 1000; i++ {
		s += "x"
	}
	_ = s

	fmt.Printf("  分配了 %d 个 slice，总计约 %d KB\n", len(data), len(data)*10)
}

// demoGoroutineProfile 展示 goroutine 信息
func demoGoroutineProfile() {
	fmt.Println("--- 3. Goroutine Profile ---")

	// 创建一些 goroutine
	var wg sync.WaitGroup
	done := make(chan struct{})

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-done // 阻塞等待
		}(i)
	}

	// 查看 goroutine 数量
	fmt.Printf("  当前 goroutine 数量: %d\n", runtime.NumGoroutine())

	// 获取 goroutine 栈信息
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false) // false = 只获取当前 goroutine
	fmt.Printf("  当前 goroutine 栈信息（前 200 字节）:\n")
	if n > 200 {
		n = 200
	}
	fmt.Printf("    %s...\n", string(buf[:n]))

	// 释放 goroutine
	close(done)
	wg.Wait()
	fmt.Printf("  释放后 goroutine 数量: %d\n", runtime.NumGoroutine())
	fmt.Println()
}

// demoMemStats 展示详细内存统计
func demoMemStats() {
	fmt.Println("--- 4. 内存统计信息 ---")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Println("  [堆内存]")
	fmt.Printf("    HeapAlloc (当前堆使用):  %d KB\n", m.HeapAlloc/1024)
	fmt.Printf("    HeapSys (堆系统分配):    %d KB\n", m.HeapSys/1024)
	fmt.Printf("    HeapIdle (堆空闲):       %d KB\n", m.HeapIdle/1024)
	fmt.Printf("    HeapInuse (堆使用中):    %d KB\n", m.HeapInuse/1024)
	fmt.Printf("    HeapObjects (堆对象数):  %d\n", m.HeapObjects)

	fmt.Println("  [GC 信息]")
	fmt.Printf("    NumGC (GC 次数):         %d\n", m.NumGC)
	fmt.Printf("    PauseTotalNs (累计暂停): %d μs\n", m.PauseTotalNs/1000)
	fmt.Printf("    NextGC (下次 GC 目标):   %d KB\n", m.NextGC/1024)

	fmt.Println("  [栈内存]")
	fmt.Printf("    StackInuse (栈使用):     %d KB\n", m.StackInuse/1024)
	fmt.Printf("    StackSys (栈系统分配):   %d KB\n", m.StackSys/1024)

	fmt.Println("  [其他]")
	fmt.Printf("    TotalAlloc (累计分配):   %d KB\n", m.TotalAlloc/1024)
	fmt.Printf("    Sys (系统总分配):        %d KB\n", m.Sys/1024)
	fmt.Printf("    Mallocs (分配次数):      %d\n", m.Mallocs)
	fmt.Printf("    Frees (释放次数):        %d\n", m.Frees)
	fmt.Println()
}
