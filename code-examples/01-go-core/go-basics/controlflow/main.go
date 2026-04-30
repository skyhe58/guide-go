// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
// 控制流示例：if/for/for-range/switch/label
package main

import "fmt"

func main() {
	fmt.Println("========== Go 控制流示例 ==========")

	// ========== 1. if 语句 ==========
	fmt.Println("\n--- 1. if 语句 ---")

	score := 85
	if score >= 90 {
		fmt.Println("优秀")
	} else if score >= 60 {
		fmt.Println("及格")
	} else {
		fmt.Println("不及格")
	}

	// if 带初始化语句（变量作用域限定在 if 块内）
	if n := computeScore(); n > 80 {
		fmt.Printf("if 初始化: 分数 %d > 80\n", n)
	}
	// n 在这里不可访问

	// ========== 2. for 循环 ==========
	fmt.Println("\n--- 2. for 循环（Go 唯一的循环关键字） ---")

	// 经典三段式
	fmt.Print("三段式: ")
	for i := 0; i < 5; i++ {
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// 类似 while
	fmt.Print("while 式: ")
	j := 0
	for j < 5 {
		fmt.Printf("%d ", j)
		j++
	}
	fmt.Println()

	// 无限循环 + break
	fmt.Print("无限循环: ")
	k := 0
	for {
		if k >= 5 {
			break
		}
		fmt.Printf("%d ", k)
		k++
	}
	fmt.Println()

	// continue 跳过
	fmt.Print("跳过偶数: ")
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			continue
		}
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// ========== 3. for-range 遍历 ==========
	fmt.Println("\n--- 3. for-range 遍历 ---")

	// 遍历切片
	fruits := []string{"苹果", "香蕉", "橙子"}
	fmt.Println("遍历切片:")
	for i, fruit := range fruits {
		fmt.Printf("  [%d] %s\n", i, fruit)
	}

	// 只要索引
	fmt.Print("只要索引: ")
	for i := range fruits {
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// 遍历字符串（按 rune）
	fmt.Println("遍历字符串 \"Go语言\":")
	for i, r := range "Go语言" {
		fmt.Printf("  字节偏移=%d, 字符=%c (U+%04X)\n", i, r, r)
	}

	// 遍历 map（顺序随机！）
	scores := map[string]int{"Alice": 90, "Bob": 85, "Charlie": 92}
	fmt.Println("遍历 map（顺序随机）:")
	for name, score := range scores {
		fmt.Printf("  %s: %d\n", name, score)
	}

	// Go 1.22+: for-range 整数
	fmt.Print("range 整数 (Go 1.22+): ")
	for i := range 5 {
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// ========== 4. switch 语句 ==========
	fmt.Println("\n--- 4. switch 语句 ---")

	// 基本 switch（默认不穿透，不需要 break）
	day := "Wednesday"
	switch day {
	case "Monday":
		fmt.Println("周一")
	case "Tuesday", "Wednesday": // 多值匹配
		fmt.Println("周二或周三")
	default:
		fmt.Println("其他")
	}

	// 无条件 switch（替代 if-else 链）
	temperature := 35
	switch {
	case temperature >= 40:
		fmt.Println("极热")
	case temperature >= 30:
		fmt.Println("炎热")
	case temperature >= 20:
		fmt.Println("舒适")
	default:
		fmt.Println("凉爽")
	}

	// switch 带初始化
	switch os := getOS(); os {
	case "linux":
		fmt.Println("Linux 系统")
	case "darwin":
		fmt.Println("macOS 系统")
	default:
		fmt.Printf("其他系统: %s\n", os)
	}

	// fallthrough 强制穿透
	fmt.Println("fallthrough 示例:")
	switch 1 {
	case 1:
		fmt.Println("  case 1")
		fallthrough // 强制执行下一个 case
	case 2:
		fmt.Println("  case 2 (通过 fallthrough 到达)")
	case 3:
		fmt.Println("  case 3 (不会到达)")
	}

	// 类型 switch
	fmt.Println("类型 switch:")
	printType(42)
	printType("hello")
	printType(3.14)
	printType(true)

	// ========== 5. label 与 break/continue ==========
	fmt.Println("\n--- 5. label 跳出多层循环 ---")
	fmt.Println("查找矩阵中第一个大于 5 的元素:")
	matrix := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
outer:
	for i, row := range matrix {
		for j, val := range row {
			if val > 5 {
				fmt.Printf("  找到: matrix[%d][%d] = %d\n", i, j, val)
				break outer // 跳出外层循环
			}
		}
	}

	fmt.Println("\n========== 示例结束 ==========")
}

func computeScore() int {
	return 85
}

func getOS() string {
	return "linux"
}

func printType(v interface{}) {
	switch v.(type) {
	case int:
		fmt.Printf("  %v 是 int\n", v)
	case string:
		fmt.Printf("  %v 是 string\n", v)
	case float64:
		fmt.Printf("  %v 是 float64\n", v)
	default:
		fmt.Printf("  %v 是未知类型\n", v)
	}
}
