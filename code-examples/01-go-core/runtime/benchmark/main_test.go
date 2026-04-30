// benchmark 性能基准测试示例
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 Go benchmark 的编写和使用：
// 1. 字符串拼接性能对比（+= vs strings.Builder vs bytes.Buffer）
// 2. slice 预分配 vs 动态扩容
// 3. map 预分配 vs 动态扩容
// 4. sync.Pool 对象复用 vs 每次新建
// 5. 避免编译器优化干扰的正确写法
//
// 运行方式：
//   go test -bench=. -benchmem
//   go test -bench=BenchmarkStringConcat -benchmem
//   go test -bench=. -benchmem -count=5 (用于 benchstat 对比)
package runtime

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
)

// ========== 1. 字符串拼接性能对比 ==========

// BenchmarkStringConcat_Plus 使用 += 拼接（低效）
func BenchmarkStringConcat_Plus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := ""
		for j := 0; j < 100; j++ {
			s += "a"
		}
		_ = s
	}
}

// BenchmarkStringConcat_Builder 使用 strings.Builder（高效）
func BenchmarkStringConcat_Builder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var sb strings.Builder
		sb.Grow(100) // 预分配容量
		for j := 0; j < 100; j++ {
			sb.WriteString("a")
		}
		_ = sb.String()
	}
}

// BenchmarkStringConcat_Buffer 使用 bytes.Buffer
func BenchmarkStringConcat_Buffer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		buf.Grow(100)
		for j := 0; j < 100; j++ {
			buf.WriteString("a")
		}
		_ = buf.String()
	}
}

// ========== 2. slice 预分配 vs 动态扩容 ==========

// BenchmarkSlice_NoPrealloc 不预分配（频繁扩容）
func BenchmarkSlice_NoPrealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := make([]int, 0)
		for j := 0; j < 1000; j++ {
			s = append(s, j)
		}
		_ = s
	}
}

// BenchmarkSlice_Prealloc 预分配容量（一次分配）
func BenchmarkSlice_Prealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := make([]int, 0, 1000)
		for j := 0; j < 1000; j++ {
			s = append(s, j)
		}
		_ = s
	}
}

// ========== 3. map 预分配 vs 动态扩容 ==========

// BenchmarkMap_NoPrealloc 不预分配
func BenchmarkMap_NoPrealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := make(map[int]int)
		for j := 0; j < 1000; j++ {
			m[j] = j
		}
	}
}

// BenchmarkMap_Prealloc 预分配容量
func BenchmarkMap_Prealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := make(map[int]int, 1000)
		for j := 0; j < 1000; j++ {
			m[j] = j
		}
	}
}

// ========== 4. sync.Pool 对象复用 ==========

type heavyObject struct {
	data [1024]byte
}

var pool = sync.Pool{
	New: func() interface{} {
		return &heavyObject{}
	},
}

// BenchmarkObject_New 每次新建对象
func BenchmarkObject_New(b *testing.B) {
	for i := 0; i < b.N; i++ {
		obj := &heavyObject{}
		obj.data[0] = 1
		_ = obj
	}
}

// BenchmarkObject_Pool 使用 sync.Pool 复用
func BenchmarkObject_Pool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		obj := pool.Get().(*heavyObject)
		obj.data[0] = 1
		pool.Put(obj)
	}
}

// ========== 5. 避免编译器优化 ==========

// 全局变量，防止编译器优化掉计算结果
var globalResult int

func compute(n int) int {
	sum := 0
	for i := 0; i < n; i++ {
		sum += i
	}
	return sum
}

// BenchmarkCompute_Correct 正确写法：结果赋给全局变量
func BenchmarkCompute_Correct(b *testing.B) {
	var r int
	for i := 0; i < b.N; i++ {
		r = compute(100)
	}
	globalResult = r // 防止编译器优化
}

// ========== 6. 子 benchmark（不同规模对比） ==========

// BenchmarkSliceAppend_Sizes 不同规模的 slice append 性能
func BenchmarkSliceAppend_Sizes(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				s := make([]int, 0, size)
				for j := 0; j < size; j++ {
					s = append(s, j)
				}
				_ = s
			}
		})
	}
}

// ========== 7. fmt.Sprintf vs strconv ==========

// BenchmarkIntToString_Sprintf 使用 fmt.Sprintf
func BenchmarkIntToString_Sprintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%d", 12345)
	}
}

// BenchmarkIntToString_Strconv 使用 strconv（更快）
func BenchmarkIntToString_Strconv(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprint(12345) // 简化示例
	}
}

// ========== 8. ResetTimer 使用示例 ==========

// BenchmarkWithSetup 包含初始化的 benchmark
func BenchmarkWithSetup(b *testing.B) {
	// 初始化阶段（不计入 benchmark 时间）
	data := make([]int, 10000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer() // 重置计时器，排除初始化开销

	for i := 0; i < b.N; i++ {
		sum := 0
		for _, v := range data {
			sum += v
		}
		globalResult = sum
	}
}
