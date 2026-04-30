// Go 1.22+ | 验证日期：2025-01-01
// Benchmark 测试示例
// 演示 Go 内置的性能基准测试框架
// 包含：基本 benchmark、b.ResetTimer、b.ReportAllocs、并行 benchmark、对比测试
// 运行方式：go test -bench=. -benchmem
package benchmark

import (
	"fmt"
	"strings"
	"testing"
)

// ============================================================
// 被测函数：字符串拼接的不同实现
// ============================================================

// concatPlus 使用 + 拼接字符串（性能最差）
func concatPlus(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "hello"
	}
	return s
}

// concatSprintf 使用 fmt.Sprintf 拼接字符串
func concatSprintf(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s = fmt.Sprintf("%s%s", s, "hello")
	}
	return s
}

// concatBuilder 使用 strings.Builder 拼接字符串（推荐方式）
func concatBuilder(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("hello")
	}
	return b.String()
}

// concatBuilderPrealloc 使用预分配的 strings.Builder（最优方式）
func concatBuilderPrealloc(n int) string {
	var b strings.Builder
	b.Grow(n * 5) // 预分配空间
	for i := 0; i < n; i++ {
		b.WriteString("hello")
	}
	return b.String()
}

// ============================================================
// 被测函数：切片操作
// ============================================================

// sliceAppend 不预分配的切片追加
func sliceAppend(n int) []int {
	var s []int
	for i := 0; i < n; i++ {
		s = append(s, i)
	}
	return s
}

// slicePrealloc 预分配的切片追加
func slicePrealloc(n int) []int {
	s := make([]int, 0, n)
	for i := 0; i < n; i++ {
		s = append(s, i)
	}
	return s
}

// ============================================================
// Benchmark 测试
// ============================================================

// 包级变量，防止编译器优化掉 benchmark 结果
var result string
var resultSlice []int

// BenchmarkConcatPlus 测试 + 拼接性能
func BenchmarkConcatPlus(b *testing.B) {
	var r string
	for i := 0; i < b.N; i++ {
		r = concatPlus(100)
	}
	result = r // 防止编译器优化
}

// BenchmarkConcatSprintf 测试 Sprintf 拼接性能
func BenchmarkConcatSprintf(b *testing.B) {
	var r string
	for i := 0; i < b.N; i++ {
		r = concatSprintf(100)
	}
	result = r
}

// BenchmarkConcatBuilder 测试 Builder 拼接性能
func BenchmarkConcatBuilder(b *testing.B) {
	var r string
	for i := 0; i < b.N; i++ {
		r = concatBuilder(100)
	}
	result = r
}

// BenchmarkConcatBuilderPrealloc 测试预分配 Builder 拼接性能
func BenchmarkConcatBuilderPrealloc(b *testing.B) {
	var r string
	for i := 0; i < b.N; i++ {
		r = concatBuilderPrealloc(100)
	}
	result = r
}

// BenchmarkWithResetTimer 演示 b.ResetTimer 排除初始化耗时
func BenchmarkWithResetTimer(b *testing.B) {
	// 模拟耗时的初始化操作
	data := make([]string, 1000)
	for i := range data {
		data[i] = fmt.Sprintf("item-%d", i)
	}

	b.ResetTimer() // 重置计时器，排除上面的初始化时间

	for i := 0; i < b.N; i++ {
		var builder strings.Builder
		for _, s := range data {
			builder.WriteString(s)
		}
		result = builder.String()
	}
}

// BenchmarkWithReportAllocs 演示 b.ReportAllocs 报告内存分配
func BenchmarkWithReportAllocs(b *testing.B) {
	b.ReportAllocs() // 显式报告内存分配（等价于 -benchmem 参数）
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 1024)
	}
}

// BenchmarkSliceAppend 测试不预分配的切片追加
func BenchmarkSliceAppend(b *testing.B) {
	var r []int
	for i := 0; i < b.N; i++ {
		r = sliceAppend(1000)
	}
	resultSlice = r
}

// BenchmarkSlicePrealloc 测试预分配的切片追加
func BenchmarkSlicePrealloc(b *testing.B) {
	var r []int
	for i := 0; i < b.N; i++ {
		r = slicePrealloc(1000)
	}
	resultSlice = r
}

// BenchmarkConcatParallel 并行 benchmark 示例
// 测试在多核并发场景下的性能
func BenchmarkConcatParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			concatBuilder(100)
		}
	})
}
