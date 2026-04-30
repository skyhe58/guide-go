// GC 三色标记模拟 — Part A 纯内存模拟
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例通过纯内存模拟演示 Go GC 的核心概念：
// 1. 三色标记法（白色/灰色/黑色）的完整标记过程
// 2. GC 统计信息查看
// 3. GOGC 和 GOMEMLIMIT 调优
// 4. 内存分配与 GC 触发观察
//
// 运行方式：go run main.go
// 查看 GC 日志：GODEBUG=gctrace=1 go run main.go
package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// ========== 三色标记模拟 ==========

// Color 表示对象的标记颜色
type Color int

const (
	White Color = iota // 白色：未扫描，GC 后回收
	Gray               // 灰色：已扫描，子对象未全部扫描
	Black              // 黑色：已完全扫描，确认存活
)

func (c Color) String() string {
	switch c {
	case White:
		return "⚪白色"
	case Gray:
		return "🔘灰色"
	case Black:
		return "⚫黑色"
	default:
		return "未知"
	}
}

// Object 模拟堆上的对象
type Object struct {
	Name     string
	Color    Color
	Children []*Object // 引用的子对象
}

// GCSimulator 三色标记 GC 模拟器
type GCSimulator struct {
	Objects []*Object // 所有堆对象
	Roots   []*Object // 根对象（栈/全局变量引用的对象）
	step    int
}

func main() {
	fmt.Println("========== GC 三色标记模拟 ==========")
	fmt.Println()

	// --- 1. 三色标记模拟 ---
	demoTriColorMarking()

	// --- 2. GC 统计信息 ---
	demoGCStats()

	// --- 3. GOGC 调优演示 ---
	demoGOGC()

	// --- 4. 内存分配与 GC 触发 ---
	demoGCTrigger()
}

// demoTriColorMarking 模拟三色标记过程
func demoTriColorMarking() {
	fmt.Println("--- 1. 三色标记法模拟 ---")

	// 构建对象图
	//   Root → A → B
	//          A → C
	//          D（无引用，垃圾）
	objA := &Object{Name: "A", Color: White}
	objB := &Object{Name: "B", Color: White}
	objC := &Object{Name: "C", Color: White}
	objD := &Object{Name: "D", Color: White} // 垃圾对象

	objA.Children = []*Object{objB, objC}

	sim := &GCSimulator{
		Objects: []*Object{objA, objB, objC, objD},
		Roots:   []*Object{objA}, // Root 引用 A
	}

	// 阶段 1：初始状态（所有对象为白色）
	sim.printState("初始状态：所有对象为白色")

	// 阶段 2：标记根对象为灰色
	fmt.Println("  [Mark Setup] 将根对象引用的对象标记为灰色...")
	for _, root := range sim.Roots {
		root.Color = Gray
	}
	sim.printState("根对象标灰后")

	// 阶段 3：扫描灰色对象
	fmt.Println("  [Concurrent Mark] 扫描灰色对象...")
	for {
		gray := sim.findGray()
		if gray == nil {
			break // 没有灰色对象，标记完成
		}
		fmt.Printf("    扫描 %s 的子对象: ", gray.Name)
		childNames := make([]string, 0)
		for _, child := range gray.Children {
			if child.Color == White {
				child.Color = Gray
				childNames = append(childNames, child.Name)
			}
		}
		if len(childNames) > 0 {
			fmt.Printf("将 [%s] 标灰\n", strings.Join(childNames, ", "))
		} else {
			fmt.Println("无白色子对象")
		}
		gray.Color = Black // 自身标黑
		fmt.Printf("    %s 标记为黑色 ✓\n", gray.Name)
	}
	sim.printState("标记完成后")

	// 阶段 4：清除白色对象
	fmt.Println("  [Concurrent Sweep] 清除白色对象（垃圾回收）...")
	for _, obj := range sim.Objects {
		if obj.Color == White {
			fmt.Printf("    回收对象 %s ♻️\n", obj.Name)
		} else {
			fmt.Printf("    保留对象 %s ✓\n", obj.Name)
		}
	}
	fmt.Println()
}

// findGray 查找第一个灰色对象
func (sim *GCSimulator) findGray() *Object {
	for _, obj := range sim.Objects {
		if obj.Color == Gray {
			return obj
		}
	}
	return nil
}

// printState 打印所有对象状态
func (sim *GCSimulator) printState(title string) {
	sim.step++
	fmt.Printf("\n  阶段 %d - %s\n", sim.step, title)
	for _, obj := range sim.Objects {
		refs := make([]string, len(obj.Children))
		for i, c := range obj.Children {
			refs[i] = c.Name
		}
		refStr := "无"
		if len(refs) > 0 {
			refStr = strings.Join(refs, ", ")
		}
		fmt.Printf("    %s: %s (引用: %s)\n", obj.Name, obj.Color, refStr)
	}
	fmt.Println()
}

// demoGCStats 展示 GC 统计信息
func demoGCStats() {
	fmt.Println("--- 2. GC 统计信息 ---")

	// 手动触发一次 GC
	runtime.GC()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("  堆分配量 (HeapAlloc):   %d KB\n", m.HeapAlloc/1024)
	fmt.Printf("  堆对象数 (HeapObjects):  %d\n", m.HeapObjects)
	fmt.Printf("  累计分配 (TotalAlloc):   %d KB\n", m.TotalAlloc/1024)
	fmt.Printf("  GC 次数 (NumGC):         %d\n", m.NumGC)
	fmt.Printf("  上次 GC 暂停:            %d μs\n", m.PauseNs[(m.NumGC+255)%256]/1000)
	fmt.Printf("  累计 GC 暂停:            %d μs\n", m.PauseTotalNs/1000)
	fmt.Printf("  下次 GC 目标 (NextGC):   %d KB\n", m.NextGC/1024)
	fmt.Println()
}

// demoGOGC 演示 GOGC 调优
func demoGOGC() {
	fmt.Println("--- 3. GOGC 调优演示 ---")

	// 获取当前 GOGC
	currentGOGC := debug.SetGCPercent(100)
	debug.SetGCPercent(currentGOGC) // 恢复
	fmt.Printf("  当前 GOGC: %d\n", currentGOGC)
	fmt.Println()

	fmt.Println("  GOGC 值说明：")
	fmt.Println("    GOGC=100（默认）: 堆增长 100%% 时触发 GC")
	fmt.Println("    GOGC=200:        堆增长 200%% 时触发（GC 频率低，内存占用高）")
	fmt.Println("    GOGC=50:         堆增长 50%% 时触发（GC 频率高，内存占用低）")
	fmt.Println("    GOGC=off:        关闭自动 GC（配合 GOMEMLIMIT 使用）")
	fmt.Println()

	fmt.Println("  推荐配置（容器环境）：")
	fmt.Println("    GOGC=off + GOMEMLIMIT=容器内存的70-80%%")
	fmt.Println("    让 GC 完全由内存上限驱动，避免不必要的 GC 开销")
	fmt.Println()
}

// demoGCTrigger 演示内存分配触发 GC
func demoGCTrigger() {
	fmt.Println("--- 4. 内存分配与 GC 触发 ---")

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	gcBefore := m.NumGC

	// 分配大量内存触发 GC
	fmt.Println("  分配大量内存以触发 GC...")
	for i := 0; i < 10; i++ {
		data := make([]byte, 1024*1024) // 1MB
		_ = data
	}

	runtime.ReadMemStats(&m)
	gcAfter := m.NumGC

	fmt.Printf("  分配前 GC 次数: %d\n", gcBefore)
	fmt.Printf("  分配后 GC 次数: %d\n", gcAfter)
	fmt.Printf("  触发了 %d 次 GC\n", gcAfter-gcBefore)
	fmt.Println()
	fmt.Println("提示：使用 GODEBUG=gctrace=1 go run main.go 可查看详细 GC 日志")
}
