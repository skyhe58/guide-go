// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
// 数据类型示例：基本类型、复合类型、零值机制、类型转换
package main

import (
	"fmt"
	"math"
	"strconv"
	"unsafe"
)

func main() {
	fmt.Println("========== Go 数据类型示例 ==========")

	// ========== 1. 基本类型与零值 ==========
	fmt.Println("\n--- 1. 零值机制 ---")
	var (
		b    bool       // false
		i    int        // 0
		f    float64    // 0.0
		s    string     // ""
		p    *int       // nil
		sl   []int      // nil
		m    map[string]int // nil
		ch   chan int    // nil
		fn   func()     // nil
		ifc  error      // nil
	)
	fmt.Printf("bool: %v\n", b)
	fmt.Printf("int: %v\n", i)
	fmt.Printf("float64: %v\n", f)
	fmt.Printf("string: %q\n", s)
	fmt.Printf("*int: %v\n", p)
	fmt.Printf("[]int: %v (nil=%v)\n", sl, sl == nil)
	fmt.Printf("map: %v (nil=%v)\n", m, m == nil)
	fmt.Printf("chan: %v\n", ch)
	fmt.Printf("func: %v\n", fn != nil)
	fmt.Printf("error: %v\n", ifc)

	// ========== 2. 整数类型与大小 ==========
	fmt.Println("\n--- 2. 整数类型大小 ---")
	fmt.Printf("int    大小: %d 字节\n", unsafe.Sizeof(int(0)))
	fmt.Printf("int8   大小: %d 字节, 范围: %d ~ %d\n", unsafe.Sizeof(int8(0)), math.MinInt8, math.MaxInt8)
	fmt.Printf("int16  大小: %d 字节, 范围: %d ~ %d\n", unsafe.Sizeof(int16(0)), math.MinInt16, math.MaxInt16)
	fmt.Printf("int32  大小: %d 字节, 范围: %d ~ %d\n", unsafe.Sizeof(int32(0)), math.MinInt32, math.MaxInt32)
	fmt.Printf("int64  大小: %d 字节\n", unsafe.Sizeof(int64(0)))

	// ========== 3. 浮点类型 ==========
	fmt.Println("\n--- 3. 浮点类型 ---")
	var f32 float32 = 3.14
	var f64 float64 = 3.141592653589793
	fmt.Printf("float32: %v (精度约 7 位)\n", f32)
	fmt.Printf("float64: %v (精度约 15 位)\n", f64)

	// 浮点精度陷阱
	fmt.Printf("0.1 + 0.2 == 0.3? %v (实际值: %.20f)\n", 0.1+0.2 == 0.3, 0.1+0.2)

	// ========== 4. byte 和 rune ==========
	fmt.Println("\n--- 4. byte vs rune ---")
	var byteVal byte = 'A'   // uint8
	var runeVal rune = '中'   // int32
	fmt.Printf("byte 'A': %d, 大小: %d 字节\n", byteVal, unsafe.Sizeof(byteVal))
	fmt.Printf("rune '中': %d (U+%04X), 大小: %d 字节\n", runeVal, runeVal, unsafe.Sizeof(runeVal))

	// 中文字符串的 byte 和 rune 长度
	str := "Hello你好"
	fmt.Printf("字符串 %q: len=%d (字节), rune 数=%d\n", str, len(str), len([]rune(str)))

	// ========== 5. 类型转换 ==========
	fmt.Println("\n--- 5. 类型转换（必须显式） ---")

	// 数值类型之间
	var intVal int = 42
	var floatVal float64 = float64(intVal)
	var uintVal uint = uint(floatVal)
	fmt.Printf("int(%d) → float64(%v) → uint(%d)\n", intVal, floatVal, uintVal)

	// 字符串与数字（strconv 包）
	numStr := strconv.Itoa(42)
	fmt.Printf("Itoa(42) = %q\n", numStr)

	num, err := strconv.Atoi("123")
	fmt.Printf("Atoi(\"123\") = %d, err=%v\n", num, err)

	_, err = strconv.Atoi("abc")
	fmt.Printf("Atoi(\"abc\") err=%v\n", err)

	// float 与 string
	fStr := strconv.FormatFloat(3.14, 'f', 2, 64)
	fmt.Printf("FormatFloat(3.14) = %q\n", fStr)

	fParsed, _ := strconv.ParseFloat("3.14", 64)
	fmt.Printf("ParseFloat(\"3.14\") = %v\n", fParsed)

	// ========== 6. 复合类型预览 ==========
	fmt.Println("\n--- 6. 复合类型预览 ---")

	// 数组（固定长度，值类型）
	arr := [3]int{1, 2, 3}
	fmt.Printf("数组: %v, 类型: %T\n", arr, arr)

	// 切片（动态长度，引用类型）
	slice := []int{1, 2, 3}
	fmt.Printf("切片: %v, len=%d, cap=%d\n", slice, len(slice), cap(slice))

	// Map
	mp := map[string]int{"Go": 1, "Rust": 2}
	fmt.Printf("Map: %v\n", mp)

	// 结构体
	type Point struct{ X, Y int }
	pt := Point{X: 10, Y: 20}
	fmt.Printf("结构体: %+v\n", pt)

	fmt.Println("\n========== 示例结束 ==========")
}
