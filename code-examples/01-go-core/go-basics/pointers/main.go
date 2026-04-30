// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
// 指针示例：值传递本质、取地址与解引用、new vs make、指针与结构体
package main

import "fmt"

// Config 配置结构体
type Config struct {
	Host    string
	Port    int
	Debug   bool
}

func main() {
	fmt.Println("========== Go 指针示例 ==========")

	// ========== 1. 指针基础 ==========
	fmt.Println("\n--- 1. 指针基础 ---")

	x := 42
	p := &x // 取地址
	fmt.Printf("x = %d, 地址 = %p\n", x, &x)
	fmt.Printf("p = %p, *p = %d\n", p, *p)

	// 通过指针修改原值
	*p = 100
	fmt.Printf("*p = 100 后, x = %d\n", x)

	// 指针的零值是 nil
	var nilPtr *int
	fmt.Printf("nil 指针: %v, 是否为 nil: %v\n", nilPtr, nilPtr == nil)

	// ========== 2. 值传递本质 ==========
	fmt.Println("\n--- 2. 值传递本质 ---")

	// 值传递：修改不影响原值
	a := 10
	modifyValue(a)
	fmt.Printf("值传递后 a = %d (未改变)\n", a)

	// 指针传递：可以修改原值
	b := 10
	modifyPointer(&b)
	fmt.Printf("指针传递后 b = %d (已改变)\n", b)

	// 结构体值传递
	cfg := Config{Host: "localhost", Port: 8080}
	modifyConfig(cfg)
	fmt.Printf("值传递后 cfg = %+v (未改变)\n", cfg)

	// 结构体指针传递
	modifyConfigPtr(&cfg)
	fmt.Printf("指针传递后 cfg = %+v (已改变)\n", cfg)

	// ========== 3. new vs make ==========
	fmt.Println("\n--- 3. new vs make ---")

	// new: 分配零值内存，返回指针
	intPtr := new(int)
	fmt.Printf("new(int): 值=%d, 类型=%T\n", *intPtr, intPtr)

	cfgPtr := new(Config)
	fmt.Printf("new(Config): %+v, 类型=%T\n", *cfgPtr, cfgPtr)

	// make: 初始化 slice/map/channel，返回值（不是指针）
	sl := make([]int, 0, 10)
	fmt.Printf("make([]int): %v, len=%d, cap=%d, 类型=%T\n", sl, len(sl), cap(sl), sl)

	m := make(map[string]int)
	fmt.Printf("make(map): %v, 类型=%T\n", m, m)

	ch := make(chan int, 5)
	fmt.Printf("make(chan): 类型=%T, cap=%d\n", ch, cap(ch))

	// 对比
	fmt.Println("\n--- new vs make 对比 ---")
	fmt.Println("new(T):  任意类型, 返回 *T, 零值初始化")
	fmt.Println("make(T): slice/map/chan, 返回 T, 初始化内部结构")

	// ========== 4. 指针与结构体 ==========
	fmt.Println("\n--- 4. 指针与结构体 ---")

	// 构造函数返回指针（Go 惯例）
	cfg2 := NewConfig("0.0.0.0", 3000, true)
	fmt.Printf("NewConfig: %+v\n", *cfg2)

	// Go 自动解引用：结构体指针可以直接访问字段
	cfg2.Port = 9090 // 等价于 (*cfg2).Port = 9090
	fmt.Printf("修改后: %+v\n", *cfg2)

	// ========== 5. 指针数组 vs 数组指针 ==========
	fmt.Println("\n--- 5. 指针数组 vs 数组指针 ---")

	// 指针数组：数组的元素是指针
	v1, v2, v3 := 1, 2, 3
	ptrArr := [3]*int{&v1, &v2, &v3}
	fmt.Printf("指针数组: [%d, %d, %d]\n", *ptrArr[0], *ptrArr[1], *ptrArr[2])

	// 数组指针：指向数组的指针
	arr := [3]int{10, 20, 30}
	arrPtr := &arr
	fmt.Printf("数组指针: %v\n", *arrPtr)

	// ========== 6. 函数返回局部变量的指针 ==========
	fmt.Println("\n--- 6. 返回局部变量指针（逃逸分析） ---")
	p2 := createInt(42)
	fmt.Printf("返回局部变量指针: *p = %d\n", *p2)
	fmt.Println("Go 编译器通过逃逸分析，自动将变量分配到堆上")
	fmt.Println("使用 go build -gcflags='-m' 可以查看逃逸分析结果")

	fmt.Println("\n========== 示例结束 ==========")
}

// 值传递：修改不影响原值
func modifyValue(x int) {
	x = 999
}

// 指针传递：可以修改原值
func modifyPointer(p *int) {
	*p = 999
}

// 结构体值传递
func modifyConfig(cfg Config) {
	cfg.Port = 9999
}

// 结构体指针传递
func modifyConfigPtr(cfg *Config) {
	cfg.Port = 9999
	cfg.Debug = true
}

// 构造函数（Go 惯例：NewXxx）
func NewConfig(host string, port int, debug bool) *Config {
	return &Config{
		Host:  host,
		Port:  port,
		Debug: debug,
	}
}

// 返回局部变量的指针（Go 允许，编译器会逃逸分析）
func createInt(val int) *int {
	x := val // 局部变量
	return &x // 编译器会将 x 分配到堆上
}
