// Go 进阶特性 — 泛型（Generics）
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 Go 泛型的核心特性（Go 1.18+）：
// 1. 泛型函数 —— 类型参数与类型推断
// 2. 类型约束 —— 内置约束与自定义约束
// 3. 泛型类型 —— 泛型栈、泛型集合
// 4. 实际应用 —— Map/Filter/Reduce 等工具函数
//
// 适用场景：
//   - 通用数据结构：Stack[T]、Set[T]、LinkedList[T]
//   - 通用工具函数：Map、Filter、Reduce、Contains
//   - 需要类型安全的容器：替代 interface{} 避免运行时断言
//
// 最佳实践：
//   - 能用接口解决的问题不必用泛型
//   - 泛型适合"操作相同、类型不同"的场景
//   - 优先使用标准库的 slices/maps 包（Go 1.21+）
//   - 约束越具体越好，避免 [T any] 无约束泛型
//
// 常见陷阱：
//   - 零值问题：泛型函数中不能用 nil，需要 var zero T
//   - 方法不支持类型参数：方法只能使用类型定义时的参数
//   - 不要为了泛型而泛型：简单场景用具体类型更清晰
package main

import (
	"cmp"
	"fmt"
	"strings"
)

// ============================================================
// 1. 泛型函数 —— 类型参数与类型推断
// ============================================================

// Contains 检查切片中是否包含目标元素
// comparable 是内置约束，表示支持 == 和 != 操作
func Contains[T comparable](slice []T, target T) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

// Index 返回元素在切片中的索引，未找到返回 -1
func Index[T comparable](slice []T, target T) int {
	for i, v := range slice {
		if v == target {
			return i
		}
	}
	return -1
}

// Max 返回两个值中的较大值
// cmp.Ordered 是 Go 1.21+ 标准库提供的约束
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Min 返回两个值中的较小值
func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// ============================================================
// 2. 自定义类型约束
// ============================================================

// Number 自定义数值类型约束
// ~ 表示底层类型（underlying type），支持自定义类型
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64
}

// Sum 计算切片元素之和
func Sum[T Number](numbers []T) T {
	var total T
	for _, n := range numbers {
		total += n
	}
	return total
}

// Average 计算平均值
func Average[T Number](numbers []T) float64 {
	if len(numbers) == 0 {
		return 0
	}
	var total float64
	for _, n := range numbers {
		total += float64(n)
	}
	return total / float64(len(numbers))
}

// 自定义类型也满足 Number 约束（因为使用了 ~）
type Score int

// ============================================================
// 3. 泛型类型 —— 泛型栈
// ============================================================

// Stack 泛型栈实现
type Stack[T any] struct {
	items []T
}

// Push 入栈
func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

// Pop 出栈，返回栈顶元素和是否成功
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T // 泛型中获取零值的方式
		return zero, false
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, true
}

// Peek 查看栈顶元素
func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// Size 返回栈大小
func (s *Stack[T]) Size() int {
	return len(s.items)
}

// IsEmpty 判断栈是否为空
func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

// ============================================================
// 4. 泛型集合（Set）
// ============================================================

// Set 泛型集合，基于 map 实现
type Set[T comparable] struct {
	items map[T]struct{}
}

// NewSet 创建新集合
func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{items: make(map[T]struct{})}
	for _, item := range items {
		s.Add(item)
	}
	return s
}

// Add 添加元素
func (s *Set[T]) Add(item T) {
	s.items[item] = struct{}{}
}

// Remove 删除元素
func (s *Set[T]) Remove(item T) {
	delete(s.items, item)
}

// Has 检查元素是否存在
func (s *Set[T]) Has(item T) bool {
	_, ok := s.items[item]
	return ok
}

// Size 返回集合大小
func (s *Set[T]) Size() int {
	return len(s.items)
}

// ToSlice 转换为切片
func (s *Set[T]) ToSlice() []T {
	result := make([]T, 0, len(s.items))
	for item := range s.items {
		result = append(result, item)
	}
	return result
}

// ============================================================
// 5. 工具函数 —— Map/Filter/Reduce
// ============================================================

// Map 对切片中的每个元素应用转换函数
func Map[T any, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// Filter 过滤切片中满足条件的元素
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce 将切片归约为单个值
func Reduce[T any, U any](slice []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}

// GroupBy 按键分组
func GroupBy[T any, K comparable](slice []T, keyFn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, v := range slice {
		key := keyFn(v)
		result[key] = append(result[key], v)
	}
	return result
}

func main() {
	fmt.Println("========== Go 泛型示例 ==========")

	// --- 1. 泛型函数 ---
	fmt.Println("\n--- 1. 泛型函数 ---")

	// Contains —— 类型推断，无需显式指定类型
	fmt.Printf("  Contains([]int{1,2,3}, 2) = %t\n",
		Contains([]int{1, 2, 3}, 2))
	fmt.Printf("  Contains([]string{\"a\",\"b\"}, \"c\") = %t\n",
		Contains([]string{"a", "b"}, "c"))

	// Max/Min
	fmt.Printf("  Max(10, 20) = %d\n", Max(10, 20))
	fmt.Printf("  Min(3.14, 2.71) = %.2f\n", Min(3.14, 2.71))
	fmt.Printf("  Max(\"apple\", \"banana\") = %s\n", Max("apple", "banana"))

	// --- 2. 自定义约束 ---
	fmt.Println("\n--- 2. 自定义类型约束 ---")
	ints := []int{1, 2, 3, 4, 5}
	floats := []float64{1.1, 2.2, 3.3}
	scores := []Score{85, 92, 78, 95}

	fmt.Printf("  Sum(ints) = %d\n", Sum(ints))
	fmt.Printf("  Sum(floats) = %.1f\n", Sum(floats))
	fmt.Printf("  Sum(scores) = %d（自定义类型 Score）\n", Sum(scores))
	fmt.Printf("  Average(ints) = %.2f\n", Average(ints))

	// --- 3. 泛型栈 ---
	fmt.Println("\n--- 3. 泛型栈 ---")

	// 整数栈
	intStack := &Stack[int]{}
	intStack.Push(1)
	intStack.Push(2)
	intStack.Push(3)
	fmt.Printf("  整数栈大小: %d\n", intStack.Size())
	if val, ok := intStack.Pop(); ok {
		fmt.Printf("  Pop: %d\n", val)
	}

	// 字符串栈
	strStack := &Stack[string]{}
	strStack.Push("Hello")
	strStack.Push("Go")
	strStack.Push("Generics")
	if val, ok := strStack.Peek(); ok {
		fmt.Printf("  字符串栈顶: %s\n", val)
	}

	// --- 4. 泛型集合 ---
	fmt.Println("\n--- 4. 泛型集合（Set）---")
	set := NewSet("Go", "Rust", "Python", "Go") // 重复元素自动去重
	fmt.Printf("  集合大小: %d\n", set.Size())
	fmt.Printf("  包含 Go: %t\n", set.Has("Go"))
	fmt.Printf("  包含 Java: %t\n", set.Has("Java"))
	set.Remove("Python")
	fmt.Printf("  删除 Python 后大小: %d\n", set.Size())

	// --- 5. Map/Filter/Reduce ---
	fmt.Println("\n--- 5. Map/Filter/Reduce ---")

	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Map: 数字 → 字符串
	strs := Map(numbers, func(n int) string {
		return fmt.Sprintf("第%d个", n)
	})
	fmt.Printf("  Map: %v\n", strs[:3])

	// Filter: 偶数
	evens := Filter(numbers, func(n int) bool {
		return n%2 == 0
	})
	fmt.Printf("  Filter(偶数): %v\n", evens)

	// Reduce: 求和
	sum := Reduce(numbers, 0, func(acc, n int) int {
		return acc + n
	})
	fmt.Printf("  Reduce(求和): %d\n", sum)

	// GroupBy: 按奇偶分组
	words := []string{"Go", "Rust", "Java", "Ruby", "C", "Python"}
	grouped := GroupBy(words, func(s string) int {
		return len(s)
	})
	fmt.Println("  GroupBy(按长度分组):")
	for length, group := range grouped {
		fmt.Printf("    长度 %d: %s\n", length, strings.Join(group, ", "))
	}

	fmt.Println("\n========== 示例结束 ==========")
}
