// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
// Map 示例：创建、操作、遍历无序性、用 map 实现 Set、并发不安全说明
package main

import (
	"fmt"
	"sort"
)

func main() {
	fmt.Println("========== Go Map 示例 ==========")

	// ========== 1. Map 创建 ==========
	fmt.Println("\n--- 1. Map 创建 ---")

	// 字面量
	m1 := map[string]int{
		"Go":     1,
		"Rust":   2,
		"Python": 3,
	}
	fmt.Printf("字面量: %v\n", m1)

	// make（推荐：预分配容量）
	m2 := make(map[string]int, 10) // 预分配 10 个元素的空间
	m2["Alice"] = 90
	m2["Bob"] = 85
	fmt.Printf("make: %v\n", m2)

	// nil map vs 空 map
	var nilMap map[string]int
	emptyMap := map[string]int{}
	fmt.Printf("nil map: %v, nil=%v\n", nilMap, nilMap == nil)
	fmt.Printf("空 map: %v, nil=%v\n", emptyMap, emptyMap == nil)

	// ========== 2. 基本操作 ==========
	fmt.Println("\n--- 2. 基本操作 ---")

	scores := make(map[string]int)

	// 写入
	scores["Alice"] = 90
	scores["Bob"] = 85
	scores["Charlie"] = 92
	fmt.Printf("写入后: %v\n", scores)

	// 读取（comma ok 模式）
	val, ok := scores["Alice"]
	fmt.Printf("Alice: %d, 存在=%v\n", val, ok)

	val, ok = scores["David"]
	fmt.Printf("David: %d, 存在=%v (不存在时返回零值)\n", val, ok)

	// 删除
	delete(scores, "Bob")
	fmt.Printf("删除 Bob 后: %v\n", scores)

	// 长度
	fmt.Printf("元素个数: %d\n", len(scores))

	// ========== 3. 遍历无序性 ==========
	fmt.Println("\n--- 3. 遍历无序性 ---")

	data := map[string]int{
		"a": 1, "b": 2, "c": 3, "d": 4, "e": 5,
	}

	fmt.Println("多次遍历 map（顺序可能不同）:")
	for round := 1; round <= 3; round++ {
		fmt.Printf("  第%d次: ", round)
		for k, v := range data {
			fmt.Printf("%s=%d ", k, v)
		}
		fmt.Println()
	}

	// 有序遍历：先排序 key
	fmt.Println("\n有序遍历（先排序 key）:")
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %s: %d\n", k, data[k])
	}

	// ========== 4. 用 Map 实现 Set ==========
	fmt.Println("\n--- 4. 用 Map 实现 Set ---")

	// Go 没有内置 Set，用 map[T]struct{} 实现（struct{} 不占内存）
	set := make(map[string]struct{})
	set["Go"] = struct{}{}
	set["Rust"] = struct{}{}
	set["Go"] = struct{}{} // 重复添加无效

	// 检查元素是否存在
	if _, exists := set["Go"]; exists {
		fmt.Println("Go 在集合中")
	}
	if _, exists := set["Java"]; !exists {
		fmt.Println("Java 不在集合中")
	}

	fmt.Printf("集合大小: %d\n", len(set))

	// ========== 5. Map 作为函数参数 ==========
	fmt.Println("\n--- 5. Map 作为函数参数（引用语义） ---")

	original := map[string]int{"a": 1, "b": 2}
	fmt.Printf("修改前: %v\n", original)
	modifyMap(original)
	fmt.Printf("修改后: %v (被修改了！map 是引用类型)\n", original)

	// ========== 6. 嵌套 Map ==========
	fmt.Println("\n--- 6. 嵌套 Map ---")

	// 学生成绩表
	grades := map[string]map[string]int{
		"Alice": {"数学": 90, "英语": 85},
		"Bob":   {"数学": 78, "英语": 92},
	}

	for student, subjects := range grades {
		fmt.Printf("  %s: ", student)
		for subject, score := range subjects {
			fmt.Printf("%s=%d ", subject, score)
		}
		fmt.Println()
	}

	// 嵌套 map 需要逐层初始化
	grades["Charlie"] = make(map[string]int)
	grades["Charlie"]["数学"] = 88
	fmt.Printf("  Charlie: 数学=%d\n", grades["Charlie"]["数学"])

	// ========== 7. 遍历中删除 ==========
	fmt.Println("\n--- 7. 遍历中删除（安全） ---")
	toClean := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
	fmt.Printf("清理前: %v\n", toClean)
	for k, v := range toClean {
		if v%2 == 0 {
			delete(toClean, k) // 在 for-range 中删除是安全的
		}
	}
	fmt.Printf("删除偶数值后: %v\n", toClean)

	// ========== 8. 并发不安全说明 ==========
	fmt.Println("\n--- 8. 并发安全说明 ---")
	fmt.Println("⚠️ Go 的 map 并发不安全！")
	fmt.Println("并发读写会触发: fatal error: concurrent map writes")
	fmt.Println("解决方案:")
	fmt.Println("  1. sync.Mutex / sync.RWMutex（通用方案）")
	fmt.Println("  2. sync.Map（读多写少场景）")
	fmt.Println("  3. 分片锁（高并发场景）")

	fmt.Println("\n========== 示例结束 ==========")
}

// Map 作为参数是引用语义
func modifyMap(m map[string]int) {
	m["c"] = 3
	m["a"] = 999
}
