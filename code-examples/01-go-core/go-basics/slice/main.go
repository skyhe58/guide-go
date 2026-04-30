// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
// 切片示例：创建、操作、扩容机制、底层数组共享、常见陷阱
package main

import "fmt"

func main() {
	fmt.Println("========== Go 切片示例 ==========")

	// ========== 1. 数组 vs 切片 ==========
	fmt.Println("\n--- 1. 数组 vs 切片 ---")

	// 数组：固定长度，值类型
	arr := [5]int{1, 2, 3, 4, 5}
	fmt.Printf("数组: %v, 类型: %T, len=%d\n", arr, arr, len(arr))

	// 切片：动态长度，引用类型
	sl := []int{1, 2, 3, 4, 5}
	fmt.Printf("切片: %v, 类型: %T, len=%d, cap=%d\n", sl, sl, len(sl), cap(sl))

	// ========== 2. 切片创建方式 ==========
	fmt.Println("\n--- 2. 切片创建方式 ---")

	// 字面量
	s1 := []int{1, 2, 3}
	fmt.Printf("字面量: %v\n", s1)

	// make（推荐：预分配容量）
	s2 := make([]int, 3, 10) // len=3, cap=10
	fmt.Printf("make: %v, len=%d, cap=%d\n", s2, len(s2), cap(s2))

	// 从数组切片
	array := [5]int{10, 20, 30, 40, 50}
	s3 := array[1:4] // [20, 30, 40]
	fmt.Printf("从数组: %v, len=%d, cap=%d\n", s3, len(s3), cap(s3))

	// nil 切片 vs 空切片
	var nilSlice []int
	emptySlice := []int{}
	fmt.Printf("nil 切片: %v, nil=%v, len=%d\n", nilSlice, nilSlice == nil, len(nilSlice))
	fmt.Printf("空切片: %v, nil=%v, len=%d\n", emptySlice, emptySlice == nil, len(emptySlice))

	// ========== 3. 切片操作 ==========
	fmt.Println("\n--- 3. 切片操作 ---")

	s := []int{1, 2, 3, 4, 5}

	// append
	s = append(s, 6)
	fmt.Printf("append(6): %v\n", s)

	s = append(s, 7, 8, 9)
	fmt.Printf("append(7,8,9): %v\n", s)

	// 追加另一个切片
	other := []int{10, 11}
	s = append(s, other...)
	fmt.Printf("append(other...): %v\n", s)

	// copy
	src := []int{1, 2, 3}
	dst := make([]int, len(src))
	n := copy(dst, src)
	fmt.Printf("copy: dst=%v, 复制了 %d 个元素\n", dst, n)

	// 删除元素（无内置方法）
	original := []int{1, 2, 3, 4, 5}
	idx := 2 // 删除索引 2 的元素
	original = append(original[:idx], original[idx+1:]...)
	fmt.Printf("删除索引2: %v\n", original)

	// 三索引切片（限制容量）
	base := []int{1, 2, 3, 4, 5}
	limited := base[1:3:3] // [2, 3], len=2, cap=2
	fmt.Printf("三索引切片: %v, len=%d, cap=%d\n", limited, len(limited), cap(limited))

	// ========== 4. 扩容机制 ==========
	fmt.Println("\n--- 4. 扩容机制 ---")
	demonstrateGrowth()

	// ========== 5. 底层数组共享 ==========
	fmt.Println("\n--- 5. 底层数组共享（重要陷阱！） ---")

	origin := []int{1, 2, 3, 4, 5}
	sub := origin[1:3] // [2, 3]，共享底层数组
	fmt.Printf("origin: %v\n", origin)
	fmt.Printf("sub: %v\n", sub)

	sub[0] = 999 // 修改 sub 会影响 origin！
	fmt.Printf("修改 sub[0]=999 后:\n")
	fmt.Printf("  origin: %v (被影响了！)\n", origin)
	fmt.Printf("  sub: %v\n", sub)

	// 安全拷贝避免共享
	fmt.Println("\n安全拷贝:")
	origin2 := []int{1, 2, 3, 4, 5}
	safeCopy := make([]int, 2)
	copy(safeCopy, origin2[1:3])
	safeCopy[0] = 999
	fmt.Printf("  origin2: %v (不受影响)\n", origin2)
	fmt.Printf("  safeCopy: %v\n", safeCopy)

	// ========== 6. 切片作为函数参数 ==========
	fmt.Println("\n--- 6. 切片作为函数参数 ---")

	data := []int{1, 2, 3}
	modifySlice(data)
	fmt.Printf("修改元素后: %v (被修改了，共享底层数组)\n", data)

	data2 := []int{1, 2, 3}
	appendToSlice(data2)
	fmt.Printf("append 后: %v (未改变，append 可能创建新数组)\n", data2)

	// ========== 7. 常见陷阱 ==========
	fmt.Println("\n--- 7. 内存泄漏陷阱 ---")
	fmt.Println("大切片的小子切片会阻止大切片被 GC:")
	fmt.Println("  bad:  sub := bigSlice[:10]  // 底层大数组不会被回收")
	fmt.Println("  good: sub := make([]int, 10); copy(sub, bigSlice[:10])")

	fmt.Println("\n========== 示例结束 ==========")
}

// 演示扩容过程
func demonstrateGrowth() {
	s := make([]int, 0)
	prevCap := cap(s)
	for i := 0; i < 20; i++ {
		s = append(s, i)
		if cap(s) != prevCap {
			fmt.Printf("  len=%2d → cap 从 %2d 扩容到 %2d\n", len(s), prevCap, cap(s))
			prevCap = cap(s)
		}
	}
}

// 切片参数：修改元素会影响原切片
func modifySlice(s []int) {
	if len(s) > 0 {
		s[0] = 999
	}
}

// 切片参数：append 不影响原切片
func appendToSlice(s []int) {
	s = append(s, 999)
	// 这里的 s 是局部变量，append 后可能指向新数组
	// 原切片的 header（len/cap）不会被修改
}

// Sum 计算切片元素之和（供测试使用）
func Sum(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// GrowSlice 演示切片扩容，返回每次扩容时的容量变化
func GrowSlice(n int) []int {
	s := make([]int, 0)
	caps := []int{0}
	for i := 0; i < n; i++ {
		prevCap := cap(s)
		s = append(s, i)
		if cap(s) != prevCap {
			caps = append(caps, cap(s))
		}
	}
	return caps
}
