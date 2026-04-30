// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
// 变量声明、常量、iota 枚举、作用域与可见性规则示例
package main

import "fmt"

// ========== 包级变量 ==========
var packageVar = "我是包级变量（首字母小写，仅包内可见）"

// PackageExported 是导出变量（首字母大写，其他包可访问）
var PackageExported = "我是导出变量"

// ========== 常量 ==========
const Pi = 3.14159

// iota 枚举：基础用法
const (
	Sunday    = iota // 0
	Monday           // 1
	Tuesday          // 2
	Wednesday        // 3
	Thursday         // 4
	Friday           // 5
	Saturday         // 6
)

// iota 枚举：位运算（权限控制）
const (
	ReadPerm    = 1 << iota // 1  (001)
	WritePerm               // 2  (010)
	ExecutePerm             // 4  (100)
)

// iota 枚举：跳值与自定义表达式
const (
	_  = iota             // 0（跳过）
	KB = 1 << (10 * iota) // 1 << 10 = 1024
	MB                    // 1 << 20 = 1048576
	GB                    // 1 << 30
	TB                    // 1 << 40
)

func main() {
	fmt.Println("========== 变量、常量与作用域示例 ==========")

	// ========== 1. 变量声明方式 ==========
	fmt.Println("\n--- 1. 变量声明方式 ---")

	// 方式1: var 完整声明
	var name string = "Go"
	fmt.Printf("var 声明: name = %q\n", name)

	// 方式2: var 类型推断
	var age = 25
	fmt.Printf("类型推断: age = %d (类型: %T)\n", age, age)

	// 方式3: 短变量声明（最常用，仅函数内）
	count := 100
	fmt.Printf("短声明: count = %d\n", count)

	// 方式4: 批量声明
	var (
		host  = "localhost"
		port  = 8080
		debug = false
	)
	fmt.Printf("批量声明: host=%s, port=%d, debug=%v\n", host, port, debug)

	// 方式5: 多变量同时赋值
	x, y := 10, 20
	fmt.Printf("多变量: x=%d, y=%d\n", x, y)

	// 交换变量（Go 的优雅写法）
	x, y = y, x
	fmt.Printf("交换后: x=%d, y=%d\n", x, y)

	// ========== 2. 常量与 iota ==========
	fmt.Println("\n--- 2. 常量与 iota ---")
	fmt.Printf("Pi = %v\n", Pi)
	fmt.Printf("星期: Sunday=%d, Monday=%d, Saturday=%d\n", Sunday, Monday, Saturday)

	// 位运算权限
	perm := ReadPerm | WritePerm
	fmt.Printf("权限 Read|Write = %d (二进制: %03b)\n", perm, perm)
	fmt.Printf("有读权限? %v\n", perm&ReadPerm != 0)
	fmt.Printf("有执行权限? %v\n", perm&ExecutePerm != 0)

	// 存储单位
	fmt.Printf("KB=%d, MB=%d, GB=%d\n", KB, MB, GB)

	// ========== 3. 作用域 ==========
	fmt.Println("\n--- 3. 作用域与变量遮蔽 ---")
	fmt.Printf("包级变量: %s\n", packageVar)

	outer := "外层"
	fmt.Printf("外层变量: %s\n", outer)
	{
		// 内层作用域可以访问外层变量
		fmt.Printf("内层访问外层: %s\n", outer)

		// 变量遮蔽（shadowing）：内层同名变量遮蔽外层
		outer := "内层（遮蔽了外层）"
		fmt.Printf("遮蔽后: %s\n", outer)
	}
	// 外层变量不受影响
	fmt.Printf("回到外层: %s\n", outer)

	// if 初始化语句的作用域
	if v := compute(); v > 10 {
		fmt.Printf("if 内部: v=%d\n", v)
	}
	// v 在这里不可访问

	// ========== 4. 可见性规则 ==========
	fmt.Println("\n--- 4. 可见性规则 ---")
	fmt.Println("Go 的可见性规则极其简单：")
	fmt.Println("  - 首字母大写 → 导出（public）")
	fmt.Println("  - 首字母小写 → 未导出（private，仅包内可见）")
	fmt.Println("  - 适用于：变量、常量、函数、类型、结构体字段、方法")

	// ========== 5. 短变量声明的多赋值特性 ==========
	fmt.Println("\n--- 5. 短变量声明多赋值 ---")
	a, err := divide(10, 3)
	fmt.Printf("第一次: a=%v, err=%v\n", a, err)

	// 这里 err 不是新变量，而是复用上面的 err
	// 但 b 是新变量，所以 := 合法（至少一个新变量）
	b, err := divide(10, 0)
	fmt.Printf("第二次: b=%v, err=%v\n", b, err)

	fmt.Println("\n========== 示例结束 ==========")
}

func compute() int {
	return 42
}

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("除数不能为零")
	}
	return a / b, nil
}
