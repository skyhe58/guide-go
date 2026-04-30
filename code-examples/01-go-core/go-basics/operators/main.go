// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
// 运算符示例：算术、关系、逻辑、位运算、赋值、取地址与解引用
package main

import "fmt"

func main() {
	fmt.Println("========== Go 运算符示例 ==========")

	// ========== 1. 算术运算符 ==========
	fmt.Println("\n--- 1. 算术运算符 ---")
	a, b := 17, 5
	fmt.Printf("%d + %d = %d\n", a, b, a+b)
	fmt.Printf("%d - %d = %d\n", a, b, a-b)
	fmt.Printf("%d * %d = %d\n", a, b, a*b)
	fmt.Printf("%d / %d = %d (整数除法，截断小数)\n", a, b, a/b)
	fmt.Printf("%d %% %d = %d (取模)\n", a, b, a%b)

	// 自增自减（Go 中是语句，不是表达式）
	c := 10
	c++ // 只有后置，没有 ++c
	fmt.Printf("c++ = %d\n", c)
	c--
	fmt.Printf("c-- = %d\n", c)

	// ========== 2. 关系运算符 ==========
	fmt.Println("\n--- 2. 关系运算符 ---")
	x, y := 10, 20
	fmt.Printf("%d == %d → %v\n", x, y, x == y)
	fmt.Printf("%d != %d → %v\n", x, y, x != y)
	fmt.Printf("%d < %d  → %v\n", x, y, x < y)
	fmt.Printf("%d > %d  → %v\n", x, y, x > y)
	fmt.Printf("%d <= %d → %v\n", x, y, x <= y)
	fmt.Printf("%d >= %d → %v\n", x, y, x >= y)

	// ========== 3. 逻辑运算符（短路求值） ==========
	fmt.Println("\n--- 3. 逻辑运算符 ---")
	t, f := true, false
	fmt.Printf("true && false = %v\n", t && f)
	fmt.Printf("true || false = %v\n", t || f)
	fmt.Printf("!true = %v\n", !t)

	// 短路求值演示
	fmt.Println("短路求值: false && sideEffect() → sideEffect 不会执行")
	if false && sideEffect() {
		// 不会到达这里
	}

	// ========== 4. 位运算符 ==========
	fmt.Println("\n--- 4. 位运算符 ---")
	m, n := 0b1010, 0b1100 // 10, 12
	fmt.Printf("m = %04b (%d)\n", m, m)
	fmt.Printf("n = %04b (%d)\n", n, n)
	fmt.Printf("m & n  = %04b (%d) 按位与\n", m&n, m&n)
	fmt.Printf("m | n  = %04b (%d) 按位或\n", m|n, m|n)
	fmt.Printf("m ^ n  = %04b (%d) 按位异或\n", m^n, m^n)
	fmt.Printf("m << 2 = %04b (%d) 左移\n", m<<2, m<<2)
	fmt.Printf("m >> 1 = %04b (%d) 右移\n", m>>1, m>>1)

	// &^ 位清除运算符（Go 特有）
	fmt.Println("\n--- 位清除 &^ (Go 特有) ---")
	perm := 0b111 // 读+写+执行 = 7
	fmt.Printf("原始权限: %03b (%d)\n", perm, perm)
	perm = perm &^ 0b010 // 清除写权限
	fmt.Printf("清除写后: %03b (%d)\n", perm, perm)

	// 位运算实际应用：权限控制
	fmt.Println("\n--- 位运算应用：权限控制 ---")
	const (
		Read    = 1 << iota // 001
		Write               // 010
		Execute             // 100
	)
	userPerm := Read | Execute // 101
	fmt.Printf("用户权限: %03b\n", userPerm)
	fmt.Printf("有读权限? %v\n", userPerm&Read != 0)
	fmt.Printf("有写权限? %v\n", userPerm&Write != 0)
	fmt.Printf("有执行权限? %v\n", userPerm&Execute != 0)

	// ========== 5. 赋值运算符 ==========
	fmt.Println("\n--- 5. 赋值运算符 ---")
	v := 100
	v += 10
	fmt.Printf("v += 10 → %d\n", v)
	v -= 20
	fmt.Printf("v -= 20 → %d\n", v)
	v *= 2
	fmt.Printf("v *= 2  → %d\n", v)
	v /= 3
	fmt.Printf("v /= 3  → %d\n", v)
	v %= 7
	fmt.Printf("v %%= 7  → %d\n", v)

	// ========== 6. 取地址与解引用 ==========
	fmt.Println("\n--- 6. 取地址 & 与解引用 * ---")
	val := 42
	ptr := &val // 取地址
	fmt.Printf("val = %d, 地址 = %p\n", val, &val)
	fmt.Printf("ptr = %p, *ptr = %d\n", ptr, *ptr)

	*ptr = 100 // 通过指针修改原值
	fmt.Printf("*ptr = 100 后, val = %d\n", val)

	// ========== 7. Go 没有三元运算符 ==========
	fmt.Println("\n--- 7. Go 没有三元运算符 ---")
	score := 85
	// Go 不支持: result := score >= 60 ? "及格" : "不及格"
	// 必须用 if-else
	var result string
	if score >= 60 {
		result = "及格"
	} else {
		result = "不及格"
	}
	fmt.Printf("分数 %d: %s\n", score, result)

	fmt.Println("\n========== 示例结束 ==========")
}

func sideEffect() bool {
	fmt.Println("  [sideEffect 被调用了]")
	return true
}
