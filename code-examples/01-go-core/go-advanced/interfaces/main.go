// Go 进阶特性 — 接口（Interfaces）
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 Go 接口的核心特性：
// 1. 隐式实现 —— 无需 implements 关键字
// 2. 类型断言与类型选择
// 3. 接口组合
// 4. 标准库常用接口（io.Reader/io.Writer/fmt.Stringer/sort.Interface）
//
// 适用场景：
//   - 定义行为契约，实现面向接口编程
//   - 解耦调用方和实现方，便于测试和替换
//   - 利用标准库接口融入 Go 生态
//
// 最佳实践：
//   - 接口应该小而精（1-3 个方法），Go 标准库的 io.Reader 只有一个方法
//   - "Accept interfaces, return structs" —— 函数参数用接口，返回值用具体类型
//   - 使用编译期断言确保类型实现了接口：var _ Interface = (*Type)(nil)
//
// 常见陷阱：
//   - 接口 nil 判断：持有 nil 指针的接口值 != nil
//   - 值接收者 vs 指针接收者：指针接收者实现的接口，值类型不满足
//   - 空接口 any 滥用：丢失类型安全性，应优先使用具体接口
package main

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
)

// ============================================================
// 1. 隐式实现 —— Go 接口的核心特性
// ============================================================

// Shape 定义了图形的行为契约
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Circle 隐式实现了 Shape 接口 —— 无需 implements 关键字
type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// Rectangle 也隐式实现了 Shape 接口
type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// 编译期断言：确保类型实现了接口
var _ Shape = Circle{}
var _ Shape = Rectangle{}

// printShapeInfo 接受接口参数 —— 面向接口编程
func printShapeInfo(s Shape) {
	fmt.Printf("  面积: %.2f, 周长: %.2f\n", s.Area(), s.Perimeter())
}

// ============================================================
// 2. 类型断言与类型选择
// ============================================================

// describeValue 使用类型选择（type switch）处理不同类型
func describeValue(v any) string {
	switch val := v.(type) {
	case int:
		return fmt.Sprintf("整数: %d", val)
	case string:
		return fmt.Sprintf("字符串: %q（长度 %d）", val, len(val))
	case float64:
		return fmt.Sprintf("浮点数: %.2f", val)
	case bool:
		return fmt.Sprintf("布尔值: %t", val)
	case Shape:
		return fmt.Sprintf("图形（面积: %.2f）", val.Area())
	default:
		return fmt.Sprintf("未知类型: %T", val)
	}
}

// ============================================================
// 3. 接口组合
// ============================================================

// Describer 接口 —— 描述自身
type Describer interface {
	Describe() string
}

// ShapeDescriber 组合了 Shape 和 Describer
type ShapeDescriber interface {
	Shape
	Describer
}

// NamedCircle 同时实现 Shape 和 Describer
type NamedCircle struct {
	Circle
	Name string
}

func (nc NamedCircle) Describe() string {
	return fmt.Sprintf("%s（半径: %.1f）", nc.Name, nc.Radius)
}

// 编译期断言：NamedCircle 满足组合接口
var _ ShapeDescriber = NamedCircle{}

// ============================================================
// 4. 标准库接口示例
// ============================================================

// --- fmt.Stringer ---

// User 实现 fmt.Stringer 接口，自定义打印格式
type User struct {
	Name string
	Age  int
}

func (u User) String() string {
	return fmt.Sprintf("%s（%d岁）", u.Name, u.Age)
}

// --- io.Reader / io.Writer ---

// CountingWriter 包装 io.Writer，统计写入字节数
type CountingWriter struct {
	Writer    io.Writer
	ByteCount int
}

func (cw *CountingWriter) Write(p []byte) (int, error) {
	n, err := cw.Writer.Write(p)
	cw.ByteCount += n
	return n, err
}

// --- sort.Interface ---

// ByAge 实现 sort.Interface，按年龄排序
type ByAge []User

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Less(i, j int) bool { return a[i].Age < a[j].Age }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// ============================================================
// 5. 接口 nil 判断陷阱演示
// ============================================================

type MyError struct {
	Message string
}

func (e *MyError) Error() string {
	return e.Message
}

// getError 演示接口 nil 陷阱
func getError(fail bool) error {
	var err *MyError = nil
	if fail {
		err = &MyError{Message: "发生错误"}
	}
	// ⚠️ 陷阱：即使 err == nil，返回的 error 接口值 != nil
	// 因为接口值 = (*MyError, nil)，type 部分不为 nil
	return err
}

// getErrorCorrect 正确的做法
func getErrorCorrect(fail bool) error {
	if fail {
		return &MyError{Message: "发生错误"}
	}
	return nil // 直接返回 nil，而不是将 nil 指针赋给接口
}

func main() {
	fmt.Println("========== Go 接口示例 ==========")

	// --- 1. 隐式实现 ---
	fmt.Println("\n--- 1. 隐式实现 ---")
	circle := Circle{Radius: 5}
	rect := Rectangle{Width: 3, Height: 4}

	fmt.Println("圆形:")
	printShapeInfo(circle)
	fmt.Println("矩形:")
	printShapeInfo(rect)

	// --- 2. 类型断言 ---
	fmt.Println("\n--- 2. 类型断言与类型选择 ---")
	values := []any{42, "Hello Go", 3.14, true, circle}
	for _, v := range values {
		fmt.Printf("  %s\n", describeValue(v))
	}

	// 带 ok 的类型断言（安全方式）
	var s Shape = circle
	if c, ok := s.(Circle); ok {
		fmt.Printf("  类型断言成功: 圆形半径 = %.1f\n", c.Radius)
	}

	// --- 3. 接口组合 ---
	fmt.Println("\n--- 3. 接口组合 ---")
	nc := NamedCircle{
		Circle: Circle{Radius: 10},
		Name:   "大圆",
	}
	fmt.Printf("  描述: %s\n", nc.Describe())
	fmt.Printf("  面积: %.2f\n", nc.Area()) // 方法提升

	// --- 4. 标准库接口 ---
	fmt.Println("\n--- 4. 标准库接口 ---")

	// fmt.Stringer
	user := User{Name: "张三", Age: 25}
	fmt.Printf("  Stringer: %s\n", user) // 自动调用 String()

	// io.Writer（CountingWriter）
	var buf strings.Builder
	cw := &CountingWriter{Writer: &buf}
	fmt.Fprint(cw, "Hello, ")
	fmt.Fprint(cw, "Go 接口!")
	fmt.Printf("  CountingWriter: 写入 %d 字节, 内容: %q\n", cw.ByteCount, buf.String())

	// sort.Interface
	users := []User{
		{Name: "张三", Age: 25},
		{Name: "李四", Age: 20},
		{Name: "王五", Age: 30},
	}
	sort.Sort(ByAge(users))
	fmt.Println("  按年龄排序:")
	for _, u := range users {
		fmt.Printf("    %s\n", u)
	}

	// --- 5. nil 判断陷阱 ---
	fmt.Println("\n--- 5. 接口 nil 判断陷阱 ---")
	err1 := getError(false)
	err2 := getErrorCorrect(false)
	fmt.Printf("  getError(false) == nil?        %t  ← ⚠️ 陷阱！\n", err1 == nil)
	fmt.Printf("  getErrorCorrect(false) == nil? %t  ← ✅ 正确\n", err2 == nil)

	fmt.Println("\n========== 示例结束 ==========")
}
