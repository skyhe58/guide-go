// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
// 函数示例：多返回值、命名返回值、可变参数、闭包、defer、递归、匿名函数、函数作为值与类型
package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("========== Go 函数示例 ==========")

	// ========== 1. 多返回值 ==========
	fmt.Println("\n--- 1. 多返回值 ---")
	quotient, remainder := divmod(17, 5)
	fmt.Printf("17 ÷ 5 = %d 余 %d\n", quotient, remainder)

	// 忽略某个返回值
	_, r := divmod(10, 3)
	fmt.Printf("10 ÷ 3 余 %d\n", r)

	// ========== 2. 命名返回值 ==========
	fmt.Println("\n--- 2. 命名返回值 ---")
	a, b := swap(10, 20)
	fmt.Printf("swap(10, 20) = %d, %d\n", a, b)

	// ========== 3. 可变参数 ==========
	fmt.Println("\n--- 3. 可变参数 ---")
	fmt.Printf("sum() = %d\n", sum())
	fmt.Printf("sum(1,2,3) = %d\n", sum(1, 2, 3))
	fmt.Printf("sum(1..10) = %d\n", sum(1, 2, 3, 4, 5, 6, 7, 8, 9, 10))

	// 展开切片传递
	nums := []int{10, 20, 30}
	fmt.Printf("sum(切片...) = %d\n", sum(nums...))

	// ========== 4. 闭包 ==========
	fmt.Println("\n--- 4. 闭包 ---")

	// 计数器闭包
	counter := makeCounter()
	fmt.Printf("counter() = %d\n", counter())
	fmt.Printf("counter() = %d\n", counter())
	fmt.Printf("counter() = %d\n", counter())

	// 乘法器闭包
	double := multiplier(2)
	triple := multiplier(3)
	fmt.Printf("double(5) = %d\n", double(5))
	fmt.Printf("triple(5) = %d\n", triple(5))

	// ========== 5. defer ==========
	fmt.Println("\n--- 5. defer（LIFO 顺序） ---")
	deferDemo()

	// defer 与命名返回值
	fmt.Println("\n--- defer 修改命名返回值 ---")
	result := deferModifyReturn()
	fmt.Printf("deferModifyReturn() = %d (不是 0！)\n", result)

	// defer 参数立即求值
	fmt.Println("\n--- defer 参数立即求值 ---")
	deferArgEval()

	// ========== 6. 递归 ==========
	fmt.Println("\n--- 6. 递归 ---")
	fmt.Printf("factorial(5) = %d\n", factorial(5))
	fmt.Printf("fibonacci(10) = %d\n", fibonacci(10))

	// ========== 7. 匿名函数 ==========
	fmt.Println("\n--- 7. 匿名函数 ---")

	// 立即执行
	result2 := func(a, b int) int {
		return a + b
	}(3, 4)
	fmt.Printf("匿名函数立即执行: %d\n", result2)

	// 赋值给变量
	greet := func(name string) string {
		return "你好, " + name + "!"
	}
	fmt.Println(greet("Go"))

	// ========== 8. 函数作为值与类型 ==========
	fmt.Println("\n--- 8. 函数作为值与类型 ---")

	// 函数类型
	type MathOp func(int, int) int

	add := MathOp(func(a, b int) int { return a + b })
	sub := MathOp(func(a, b int) int { return a - b })

	fmt.Printf("apply(add, 10, 3) = %d\n", apply(add, 10, 3))
	fmt.Printf("apply(sub, 10, 3) = %d\n", apply(sub, 10, 3))

	// 高阶函数：map/filter 模式
	fmt.Println("\n--- 高阶函数示例 ---")
	words := []string{"hello", "world", "go", "language"}
	upper := mapStrings(words, strings.ToUpper)
	fmt.Printf("toUpper: %v\n", upper)

	long := filterStrings(words, func(s string) bool {
		return len(s) > 3
	})
	fmt.Printf("长度>3: %v\n", long)

	fmt.Println("\n========== 示例结束 ==========")
}

// 多返回值
func divmod(a, b int) (int, int) {
	return a / b, a % b
}

// 命名返回值 + 裸 return
func swap(a, b int) (x, y int) {
	x = b
	y = a
	return // 裸 return，返回命名返回值
}

// 可变参数
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// 闭包：计数器
func makeCounter() func() int {
	count := 0 // 被闭包捕获，生命周期延长
	return func() int {
		count++
		return count
	}
}

// 闭包：乘法器
func multiplier(factor int) func(int) int {
	return func(x int) int {
		return x * factor
	}
}

// defer 演示
func deferDemo() {
	fmt.Println("  开始")
	defer fmt.Println("  defer 1（最后执行）")
	defer fmt.Println("  defer 2")
	defer fmt.Println("  defer 3（最先执行）")
	fmt.Println("  结束")
	// 输出顺序: 开始 → 结束 → defer 3 → defer 2 → defer 1
}

// defer 修改命名返回值
func deferModifyReturn() (result int) {
	defer func() {
		result++ // defer 可以修改命名返回值！
	}()
	return 0 // 实际返回 1
}

// defer 参数立即求值
func deferArgEval() {
	x := 10
	defer fmt.Printf("  defer 中 x = %d (声明时的值)\n", x)
	x = 20
	fmt.Printf("  当前 x = %d\n", x)
}

// 递归：阶乘
func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

// 递归：斐波那契
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

// 函数作为参数
func apply(op func(int, int) int, a, b int) int {
	return op(a, b)
}

// 高阶函数：map
func mapStrings(ss []string, fn func(string) string) []string {
	result := make([]string, len(ss))
	for i, s := range ss {
		result[i] = fn(s)
	}
	return result
}

// 高阶函数：filter
func filterStrings(ss []string, fn func(string) bool) []string {
	var result []string
	for _, s := range ss {
		if fn(s) {
			result = append(result, s)
		}
	}
	return result
}
