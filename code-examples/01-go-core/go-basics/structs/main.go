// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
// 结构体与方法示例：定义、初始化、值接收者 vs 指针接收者、方法集、嵌入组合
package main

import "fmt"

// ========== 结构体定义 ==========

// User 用户结构体
type User struct {
	Name  string
	Age   int
	Email string
}

// Rect 矩形（演示值接收者 vs 指针接收者）
type Rect struct {
	Width  float64
	Height float64
}

// ========== 值接收者方法 ==========

// Area 计算面积（值接收者：不修改原值）
func (r Rect) Area() float64 {
	return r.Width * r.Height
}

// Perimeter 计算周长（值接收者）
func (r Rect) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// String 实现 fmt.Stringer 接口
func (r Rect) String() string {
	return fmt.Sprintf("Rect(%.1f x %.1f)", r.Width, r.Height)
}

// ========== 指针接收者方法 ==========

// Scale 缩放（指针接收者：修改原值）
func (r *Rect) Scale(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

// SetWidth 设置宽度（指针接收者）
func (r *Rect) SetWidth(w float64) {
	r.Width = w
}

// ========== 构造函数（Go 惯例） ==========

// NewUser 创建用户（Go 惯例：NewXxx 返回指针）
func NewUser(name string, age int, email string) *User {
	return &User{
		Name:  name,
		Age:   age,
		Email: email,
	}
}

func (u *User) String() string {
	return fmt.Sprintf("%s(%d岁, %s)", u.Name, u.Age, u.Email)
}

// ========== 结构体嵌入（组合） ==========

// Animal 基础动物
type Animal struct {
	Name   string
	Sound  string
}

func (a Animal) Speak() string {
	return fmt.Sprintf("%s: %s!", a.Name, a.Sound)
}

// Dog 狗（嵌入 Animal）
type Dog struct {
	Animal // 嵌入，Dog 自动获得 Animal 的字段和方法
	Breed  string
}

// Fetch Dog 特有方法
func (d Dog) Fetch(item string) string {
	return fmt.Sprintf("%s 捡回了 %s", d.Name, item)
}

// ========== 方法集规则演示 ==========

// Stringer 接口
type Stringer interface {
	String() string
}

// Counter 计数器（演示方法集）
type Counter struct {
	count int
}

// 值接收者方法
func (c Counter) Value() int {
	return c.count
}

// 指针接收者方法
func (c *Counter) Increment() {
	c.count++
}

func main() {
	fmt.Println("========== 结构体与方法示例 ==========")

	// ========== 1. 结构体初始化 ==========
	fmt.Println("\n--- 1. 结构体初始化方式 ---")

	// 方式1: 字段名初始化（推荐）
	u1 := User{Name: "Alice", Age: 25, Email: "alice@example.com"}
	fmt.Printf("字段名: %+v\n", u1)

	// 方式2: 按顺序初始化（不推荐，字段增减会出错）
	u2 := User{"Bob", 30, "bob@example.com"}
	fmt.Printf("按顺序: %+v\n", u2)

	// 方式3: new（返回指针，零值）
	u3 := new(User)
	fmt.Printf("new: %+v\n", *u3)

	// 方式4: 取地址（等价于 new + 初始化）
	u4 := &User{Name: "Charlie"}
	fmt.Printf("取地址: %+v\n", *u4)

	// 方式5: 构造函数
	u5 := NewUser("Diana", 28, "diana@example.com")
	fmt.Printf("构造函数: %s\n", u5)

	// ========== 2. 值接收者 vs 指针接收者 ==========
	fmt.Println("\n--- 2. 值接收者 vs 指针接收者 ---")

	rect := Rect{Width: 10, Height: 5}
	fmt.Printf("原始: %s\n", rect)
	fmt.Printf("面积: %.1f\n", rect.Area())
	fmt.Printf("周长: %.1f\n", rect.Perimeter())

	// 指针接收者方法可以修改原值
	rect.Scale(2)
	fmt.Printf("Scale(2) 后: %s\n", rect)
	fmt.Printf("新面积: %.1f\n", rect.Area())

	// 值接收者方法不能修改原值
	fmt.Println("\n--- 值接收者不修改原值 ---")
	r2 := Rect{Width: 3, Height: 4}
	demonstrateValueReceiver(r2)
	fmt.Printf("调用后 r2 不变: %s\n", r2)

	// ========== 3. 方法集规则 ==========
	fmt.Println("\n--- 3. 方法集规则 ---")
	fmt.Println("值类型 T 的方法集: 只包含值接收者方法")
	fmt.Println("指针类型 *T 的方法集: 包含值接收者 + 指针接收者方法")

	c := Counter{}
	c.Increment() // 值类型也能调用指针接收者方法（编译器自动取地址）
	c.Increment()
	fmt.Printf("Counter 值: %d\n", c.Value())

	// 但在接口赋值时，方法集规则严格执行
	// var s Stringer = rect  // ✅ Rect 有值接收者 String()
	// var s Stringer = &rect // ✅ *Rect 包含所有方法
	var s Stringer = rect
	fmt.Printf("接口调用: %s\n", s)

	// ========== 4. 结构体嵌入（组合） ==========
	fmt.Println("\n--- 4. 结构体嵌入（组合优于继承） ---")

	dog := Dog{
		Animal: Animal{Name: "旺财", Sound: "汪汪"},
		Breed:  "柴犬",
	}

	// 方法提升：Dog 可以直接调用 Animal 的方法
	fmt.Println(dog.Speak())
	fmt.Println(dog.Fetch("飞盘"))

	// 字段提升：Dog 可以直接访问 Animal 的字段
	fmt.Printf("名字: %s, 品种: %s\n", dog.Name, dog.Breed)

	// 也可以显式访问嵌入字段
	fmt.Printf("显式访问: %s\n", dog.Animal.Name)

	// ========== 5. 匿名结构体 ==========
	fmt.Println("\n--- 5. 匿名结构体 ---")
	point := struct {
		X, Y int
	}{X: 10, Y: 20}
	fmt.Printf("匿名结构体: %+v\n", point)

	// 常用于测试中的表驱动测试
	tests := []struct {
		input    int
		expected int
	}{
		{1, 1},
		{2, 4},
		{3, 9},
	}
	for _, tt := range tests {
		fmt.Printf("  square(%d) = %d, expected %d\n", tt.input, tt.input*tt.input, tt.expected)
	}

	// ========== 6. 结构体比较 ==========
	fmt.Println("\n--- 6. 结构体比较 ---")
	p1 := User{Name: "Alice", Age: 25}
	p2 := User{Name: "Alice", Age: 25}
	p3 := User{Name: "Bob", Age: 30}
	fmt.Printf("p1 == p2: %v (所有字段相等)\n", p1 == p2)
	fmt.Printf("p1 == p3: %v\n", p1 == p3)

	fmt.Println("\n========== 示例结束 ==========")
}

// 演示值接收者不修改原值
func demonstrateValueReceiver(r Rect) {
	r.Width = 999 // 修改的是副本
	fmt.Printf("函数内: %s\n", r)
}
