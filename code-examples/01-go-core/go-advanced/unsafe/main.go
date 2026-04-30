// Go 进阶特性 — unsafe 包
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 Go unsafe 包的核心操作：
// 1. unsafe.Sizeof / Alignof / Offsetof —— 内存布局查询
// 2. unsafe.Pointer —— 任意指针转换
// 3. 结构体内存对齐优化
// 4. 指针运算访问结构体字段
//
// 适用场景：
//   - 底层库开发：需要极致性能的序列化/反序列化
//   - 内存布局优化：减少结构体的内存占用
//   - 与 C 代码交互（cgo）
//   - 标准库内部实现（math.Float64bits 等）
//
// 最佳实践：
//   - 日常开发中几乎不需要 unsafe，优先使用安全的 Go 代码
//   - 如果必须使用，封装在小函数中并充分测试
//   - 使用 go vet 检查 unsafe 使用是否符合规范
//   - 添加详细注释说明为什么需要 unsafe
//
// 常见陷阱：
//   - uintptr 不是指针，GC 不会追踪，对象可能被回收
//   - unsafe.Pointer 转换必须遵循 Go 规范定义的 6 种合法模式
//   - 内存布局可能随 Go 版本变化，unsafe 代码可能不兼容
//   - cgo 分配的内存必须手动释放，Go GC 不会回收
package main

import (
	"fmt"
	"unsafe"
)

// ============================================================
// 1. 内存布局查询
// ============================================================

// 演示基本类型的大小和对齐
func showBasicTypeSizes() {
	fmt.Println("  基本类型大小和对齐:")
	fmt.Printf("    bool:    大小=%d, 对齐=%d\n", unsafe.Sizeof(true), unsafe.Alignof(true))
	fmt.Printf("    int8:    大小=%d, 对齐=%d\n", unsafe.Sizeof(int8(0)), unsafe.Alignof(int8(0)))
	fmt.Printf("    int16:   大小=%d, 对齐=%d\n", unsafe.Sizeof(int16(0)), unsafe.Alignof(int16(0)))
	fmt.Printf("    int32:   大小=%d, 对齐=%d\n", unsafe.Sizeof(int32(0)), unsafe.Alignof(int32(0)))
	fmt.Printf("    int64:   大小=%d, 对齐=%d\n", unsafe.Sizeof(int64(0)), unsafe.Alignof(int64(0)))
	fmt.Printf("    float64: 大小=%d, 对齐=%d\n", unsafe.Sizeof(float64(0)), unsafe.Alignof(float64(0)))
	fmt.Printf("    string:  大小=%d, 对齐=%d\n", unsafe.Sizeof(""), unsafe.Alignof(""))
	fmt.Printf("    slice:   大小=%d, 对齐=%d\n", unsafe.Sizeof([]int{}), unsafe.Alignof([]int{}))
	fmt.Printf("    pointer: 大小=%d, 对齐=%d\n", unsafe.Sizeof((*int)(nil)), unsafe.Alignof((*int)(nil)))
}

// ============================================================
// 2. 结构体内存对齐
// ============================================================

// BadLayout 差的内存布局 —— 字段排列导致大量填充
type BadLayout struct {
	a bool   // 1 字节 + 7 字节填充（下一个字段需要 8 字节对齐）
	b int64  // 8 字节
	c bool   // 1 字节 + 7 字节填充（结构体大小需要是最大对齐的倍数）
}

// GoodLayout 优化后的内存布局 —— 按大小从大到小排列
type GoodLayout struct {
	b int64 // 8 字节
	a bool  // 1 字节
	c bool  // 1 字节 + 6 字节填充
}

// RealWorldExample 实际项目中的结构体优化示例
type RealWorldExample struct {
	// 优化前（如果乱序排列可能占用更多内存）
	ID        int64  // 8 字节
	Score     int64  // 8 字节
	Name      string // 16 字节（指针 + 长度）
	Email     string // 16 字节
	Age       int32  // 4 字节
	IsActive  bool   // 1 字节
	IsAdmin   bool   // 1 字节
	// + 2 字节填充 = 56 字节
}

// showStructLayout 展示结构体的内存布局
func showStructLayout() {
	fmt.Println("\n  结构体内存布局对比:")

	// BadLayout
	fmt.Printf("    BadLayout  大小: %d 字节\n", unsafe.Sizeof(BadLayout{}))
	fmt.Printf("      a(bool)  偏移: %d\n", unsafe.Offsetof(BadLayout{}.a))
	fmt.Printf("      b(int64) 偏移: %d\n", unsafe.Offsetof(BadLayout{}.b))
	fmt.Printf("      c(bool)  偏移: %d\n", unsafe.Offsetof(BadLayout{}.c))

	// GoodLayout
	fmt.Printf("    GoodLayout 大小: %d 字节\n", unsafe.Sizeof(GoodLayout{}))
	fmt.Printf("      b(int64) 偏移: %d\n", unsafe.Offsetof(GoodLayout{}.b))
	fmt.Printf("      a(bool)  偏移: %d\n", unsafe.Offsetof(GoodLayout{}.a))
	fmt.Printf("      c(bool)  偏移: %d\n", unsafe.Offsetof(GoodLayout{}.c))

	saved := unsafe.Sizeof(BadLayout{}) - unsafe.Sizeof(GoodLayout{})
	fmt.Printf("    节省: %d 字节（%.0f%%）\n", saved,
		float64(saved)/float64(unsafe.Sizeof(BadLayout{}))*100)

	// RealWorldExample
	fmt.Printf("\n    RealWorldExample 大小: %d 字节\n", unsafe.Sizeof(RealWorldExample{}))
}

// ============================================================
// 3. unsafe.Pointer 类型转换
// ============================================================

// float64ToBits 将 float64 转换为 uint64（查看 IEEE 754 位表示）
// 这是标准库 math.Float64bits 的实现方式
func float64ToBits(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

// bitsToFloat64 将 uint64 转换为 float64
func bitsToFloat64(b uint64) float64 {
	return *(*float64)(unsafe.Pointer(&b))
}

// stringToBytes 零拷贝将 string 转换为 []byte
// ⚠️ 返回的 []byte 不可修改！修改会导致未定义行为
func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// bytesToString 零拷贝将 []byte 转换为 string
func bytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// ============================================================
// 4. 指针运算
// ============================================================

// Point 用于演示指针运算的结构体
type Point struct {
	X int64
	Y int64
	Z int64
}

// accessFieldByOffset 通过指针运算访问结构体字段
func accessFieldByOffset() {
	p := Point{X: 10, Y: 20, Z: 30}
	fmt.Printf("    原始值: X=%d, Y=%d, Z=%d\n", p.X, p.Y, p.Z)

	// 获取结构体的起始地址
	ptr := unsafe.Pointer(&p)

	// 通过偏移量访问 Y 字段
	yOffset := unsafe.Offsetof(p.Y)
	yPtr := (*int64)(unsafe.Pointer(uintptr(ptr) + yOffset))
	fmt.Printf("    通过偏移量读取 Y: %d（偏移 %d 字节）\n", *yPtr, yOffset)

	// 通过偏移量修改 Z 字段
	zOffset := unsafe.Offsetof(p.Z)
	zPtr := (*int64)(unsafe.Pointer(uintptr(ptr) + zOffset))
	*zPtr = 100
	fmt.Printf("    修改 Z 后: X=%d, Y=%d, Z=%d\n", p.X, p.Y, p.Z)
}

// ============================================================
// 5. 数组元素访问
// ============================================================

// accessArrayElement 通过指针运算访问数组元素
func accessArrayElement() {
	arr := [5]int64{10, 20, 30, 40, 50}
	fmt.Printf("    原始数组: %v\n", arr)

	// 获取数组首元素地址
	ptr := unsafe.Pointer(&arr[0])
	elemSize := unsafe.Sizeof(arr[0])

	// 通过指针运算访问第 3 个元素（索引 2）
	elem2Ptr := (*int64)(unsafe.Pointer(uintptr(ptr) + 2*elemSize))
	fmt.Printf("    通过指针访问 arr[2]: %d\n", *elem2Ptr)

	// 修改第 4 个元素（索引 3）
	elem3Ptr := (*int64)(unsafe.Pointer(uintptr(ptr) + 3*elemSize))
	*elem3Ptr = 999
	fmt.Printf("    修改 arr[3] 后: %v\n", arr)
}

func main() {
	fmt.Println("========== Go unsafe 包示例 ==========")
	fmt.Println("⚠️ unsafe 操作绕过类型安全，仅用于底层库开发和性能优化")

	// --- 1. 基本类型大小 ---
	fmt.Println("\n--- 1. 基本类型大小和对齐 ---")
	showBasicTypeSizes()

	// --- 2. 结构体内存对齐 ---
	fmt.Println("\n--- 2. 结构体内存对齐优化 ---")
	showStructLayout()

	// --- 3. 类型转换 ---
	fmt.Println("\n--- 3. unsafe.Pointer 类型转换 ---")

	// float64 ↔ uint64
	f := 3.14
	bits := float64ToBits(f)
	restored := bitsToFloat64(bits)
	fmt.Printf("  float64 %.2f → uint64 %d → float64 %.2f\n", f, bits, restored)

	// string ↔ []byte 零拷贝
	s := "Hello, Go unsafe!"
	b := stringToBytes(s)
	s2 := bytesToString(b)
	fmt.Printf("  string → []byte: %v\n", b[:5])
	fmt.Printf("  []byte → string: %q\n", s2)

	// --- 4. 指针运算 ---
	fmt.Println("\n--- 4. 指针运算访问结构体字段 ---")
	accessFieldByOffset()

	// --- 5. 数组元素访问 ---
	fmt.Println("\n--- 5. 指针运算访问数组元素 ---")
	accessArrayElement()

	fmt.Println("\n========== 示例结束 ==========")
}
