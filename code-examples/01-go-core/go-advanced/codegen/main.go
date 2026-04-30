// Go 进阶特性 — 代码生成（Code Generation）
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 Go 代码生成的核心概念：
// 1. go generate 指令的使用方式
// 2. 手动实现 stringer 模式（枚举字符串化）
// 3. 构建标签（Build Tags）的使用
// 4. 交叉编译示例
//
// 适用场景：
//   - 枚举类型的 String() 方法生成（stringer）
//   - Mock 接口实现生成（mockgen）
//   - Protocol Buffers 代码生成（protoc）
//   - 序列化/反序列化代码生成（easyjson）
//   - 依赖注入代码生成（Wire）
//
// 最佳实践：
//   - 生成的代码应提交到版本控制，确保 go build 不依赖生成工具
//   - 在 Makefile 中添加 generate target，统一管理代码生成
//   - 生成的文件头部添加 "Code generated ... DO NOT EDIT." 注释
//   - 使用 //go:generate 指令而非手动运行命令
//
// 常见陷阱：
//   - 忘记运行 go generate 导致生成代码过时
//   - //go:generate 中 // 和 go 之间不能有空格
//   - 生成工具未安装导致 go generate 失败
//   - 构建标签 //go:build 必须在 package 声明之前
package main

import (
	"fmt"
	"runtime"
	"strings"
)

// ============================================================
// 1. 手动实现 stringer 模式
// ============================================================

// 在实际项目中，通常使用 go generate + stringer 工具自动生成
// 这里手动实现以展示原理

// Color 颜色枚举
type Color int

const (
	Red Color = iota
	Green
	Blue
	Yellow
	Purple
)

// colorNames 枚举值到字符串的映射
var colorNames = map[Color]string{
	Red:    "Red",
	Green:  "Green",
	Blue:   "Blue",
	Yellow: "Yellow",
	Purple: "Purple",
}

// String 实现 fmt.Stringer 接口
// 在实际项目中，这个方法由 stringer 工具自动生成：
//
//go:generate stringer -type=Color
func (c Color) String() string {
	if name, ok := colorNames[c]; ok {
		return name
	}
	return fmt.Sprintf("Color(%d)", int(c))
}

// IsValid 检查枚举值是否有效
func (c Color) IsValid() bool {
	_, ok := colorNames[c]
	return ok
}

// ParseColor 从字符串解析颜色
func ParseColor(s string) (Color, error) {
	s = strings.ToLower(s)
	for c, name := range colorNames {
		if strings.ToLower(name) == s {
			return c, nil
		}
	}
	return 0, fmt.Errorf("未知颜色: %s", s)
}

// ============================================================
// 2. 更复杂的枚举示例：HTTP 方法
// ============================================================

// HTTPMethod HTTP 方法枚举
type HTTPMethod int

const (
	GET HTTPMethod = iota
	POST
	PUT
	DELETE
	PATCH
)

var httpMethodNames = [...]string{
	GET:    "GET",
	POST:   "POST",
	PUT:    "PUT",
	DELETE: "DELETE",
	PATCH:  "PATCH",
}

func (m HTTPMethod) String() string {
	if int(m) < len(httpMethodNames) {
		return httpMethodNames[m]
	}
	return fmt.Sprintf("HTTPMethod(%d)", int(m))
}

// ============================================================
// 3. 代码生成模式：接口 Mock
// ============================================================

// UserRepository 用户仓库接口
// 在实际项目中使用 mockgen 生成 Mock 实现：
//
//go:generate mockgen -source=main.go -destination=mock_main.go -package=main
type UserRepository interface {
	GetByID(id int) (*UserRecord, error)
	Create(name string) (*UserRecord, error)
	Delete(id int) error
}

// UserRecord 用户记录
type UserRecord struct {
	ID   int
	Name string
}

// MockUserRepository 手动实现的 Mock（演示 mockgen 生成的代码结构）
type MockUserRepository struct {
	GetByIDFunc func(id int) (*UserRecord, error)
	CreateFunc  func(name string) (*UserRecord, error)
	DeleteFunc  func(id int) error
}

func (m *MockUserRepository) GetByID(id int) (*UserRecord, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, fmt.Errorf("GetByID not mocked")
}

func (m *MockUserRepository) Create(name string) (*UserRecord, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(name)
	}
	return nil, fmt.Errorf("Create not mocked")
}

func (m *MockUserRepository) Delete(id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return fmt.Errorf("Delete not mocked")
}

// ============================================================
// 4. 构建标签与平台信息
// ============================================================

// getPlatformInfo 获取当前平台信息
// 在实际项目中，可以通过构建标签为不同平台提供不同实现
func getPlatformInfo() string {
	return fmt.Sprintf("OS: %s, Arch: %s, Go: %s, CPUs: %d",
		runtime.GOOS, runtime.GOARCH, runtime.Version(), runtime.NumCPU())
}

// getCrossCompileCommands 返回常用的交叉编译命令
func getCrossCompileCommands() []string {
	return []string{
		"GOOS=linux   GOARCH=amd64 go build -o app-linux-amd64",
		"GOOS=linux   GOARCH=arm64 go build -o app-linux-arm64",
		"GOOS=darwin  GOARCH=arm64 go build -o app-darwin-arm64",
		"GOOS=windows GOARCH=amd64 go build -o app-windows.exe",
		"CGO_ENABLED=0 GOOS=linux go build -o app-static  # 纯静态二进制",
	}
}

// ============================================================
// 5. Makefile 中的 generate target 示例
// ============================================================

func showMakefileExample() {
	makefile := `
# Makefile 中的代码生成 target 示例
.PHONY: generate mock proto

# 运行所有代码生成
generate:
    go generate ./...

# 生成 Mock
mock:
    mockgen -source=internal/service/user.go \
            -destination=internal/service/mock_user.go \
            -package=service

# 生成 Protobuf 代码
proto:
    protoc --go_out=. --go-grpc_out=. api/proto/*.proto

# 生成枚举字符串方法
stringer:
    stringer -type=Color,HTTPMethod ./...

# 检查生成的代码是否最新
check-generate:
    go generate ./...
    git diff --exit-code || (echo "生成的代码不是最新的" && exit 1)`

	fmt.Println(makefile)
}

func main() {
	fmt.Println("========== Go 代码生成示例 ==========")

	// --- 1. stringer 模式 ---
	fmt.Println("\n--- 1. 枚举字符串化（stringer 模式）---")
	colors := []Color{Red, Green, Blue, Yellow, Purple}
	for _, c := range colors {
		fmt.Printf("  Color(%d) = %s, 有效: %t\n", int(c), c, c.IsValid())
	}

	// 无效枚举值
	invalid := Color(99)
	fmt.Printf("  Color(99) = %s, 有效: %t\n", invalid, invalid.IsValid())

	// 字符串解析
	if c, err := ParseColor("blue"); err == nil {
		fmt.Printf("  ParseColor(\"blue\") = %s\n", c)
	}
	if _, err := ParseColor("unknown"); err != nil {
		fmt.Printf("  ParseColor(\"unknown\") 错误: %v\n", err)
	}

	// --- 2. HTTP 方法枚举 ---
	fmt.Println("\n--- 2. HTTP 方法枚举 ---")
	methods := []HTTPMethod{GET, POST, PUT, DELETE, PATCH}
	for _, m := range methods {
		fmt.Printf("  HTTPMethod(%d) = %s\n", int(m), m)
	}

	// --- 3. Mock 示例 ---
	fmt.Println("\n--- 3. Mock 接口示例 ---")
	mock := &MockUserRepository{
		GetByIDFunc: func(id int) (*UserRecord, error) {
			return &UserRecord{ID: id, Name: "张三"}, nil
		},
		CreateFunc: func(name string) (*UserRecord, error) {
			return &UserRecord{ID: 1, Name: name}, nil
		},
	}

	user, _ := mock.GetByID(1)
	fmt.Printf("  mock.GetByID(1) = %+v\n", user)

	newUser, _ := mock.Create("李四")
	fmt.Printf("  mock.Create(\"李四\") = %+v\n", newUser)

	// --- 4. 平台信息与交叉编译 ---
	fmt.Println("\n--- 4. 平台信息与交叉编译 ---")
	fmt.Printf("  当前平台: %s\n", getPlatformInfo())
	fmt.Println("  常用交叉编译命令:")
	for _, cmd := range getCrossCompileCommands() {
		fmt.Printf("    $ %s\n", cmd)
	}

	// --- 5. Makefile 示例 ---
	fmt.Println("\n--- 5. Makefile generate target 示例 ---")
	showMakefileExample()

	fmt.Println("\n========== 示例结束 ==========")
}
