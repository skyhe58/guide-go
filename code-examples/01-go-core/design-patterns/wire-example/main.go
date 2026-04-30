// Wire 依赖注入概念演示 — Google Wire 编译时 DI
// 本示例演示 Wire 的核心概念（Provider/Injector），不需要实际运行 wire generate
// 通过手动模拟 Wire 生成的代码来展示 DI 原理
//
// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
//
// 运行方式: go run ./wire-example/

package main

import "fmt"

// =============================================================================
// Part A: Wire 概念演示（手动模拟 Wire 生成的代码）
// =============================================================================

// --- 1. 定义业务类型 ---

// Config 应用配置
type Config struct {
	DBHost string
	DBPort int
	DBName string
}

// DB 数据库连接
type DB struct {
	config *Config
}

// UserRepository 用户数据访问层
type UserRepository struct {
	db *DB
}

// UserService 用户业务逻辑层
type UserService struct {
	repo *UserRepository
}

// App 应用入口
type App struct {
	userService *UserService
}

// --- 2. Provider 函数（Wire 的核心概念）---
// 每个 Provider 是一个普通的 Go 函数，接收依赖作为参数，返回一个值
// Wire 通过分析 Provider 的输入输出类型自动组装依赖链

// NewConfig 提供配置（Provider）
func NewConfig() *Config {
	return &Config{
		DBHost: "localhost",
		DBPort: 5432,
		DBName: "myapp",
	}
}

// NewDB 提供数据库连接（Provider），依赖 Config
func NewDB(cfg *Config) (*DB, error) {
	fmt.Printf("  [Provider] 创建 DB 连接: %s:%d/%s\n", cfg.DBHost, cfg.DBPort, cfg.DBName)
	return &DB{config: cfg}, nil
}

// NewUserRepository 提供用户仓库（Provider），依赖 DB
func NewUserRepository(db *DB) *UserRepository {
	fmt.Println("  [Provider] 创建 UserRepository")
	return &UserRepository{db: db}
}

// NewUserService 提供用户服务（Provider），依赖 UserRepository
func NewUserService(repo *UserRepository) *UserService {
	fmt.Println("  [Provider] 创建 UserService")
	return &UserService{repo: repo}
}

// NewApp 提供应用入口（Provider），依赖 UserService
func NewApp(userService *UserService) *App {
	fmt.Println("  [Provider] 创建 App")
	return &App{userService: userService}
}

// --- 3. Injector 函数（模拟 Wire 生成的代码）---

// InitializeApp 模拟 Wire 生成的注入器
// 在真实项目中，开发者只需编写函数签名：
//
//	//go:build wireinject
//	func InitializeApp() (*App, error) {
//	    wire.Build(NewConfig, NewDB, NewUserRepository, NewUserService, NewApp)
//	    return nil, nil
//	}
//
// Wire 会自动生成以下实现代码（wire_gen.go）：
func InitializeApp() (*App, error) {
	// Wire 分析依赖链后生成的代码：
	// 1. 先创建没有依赖的 Config
	config := NewConfig()

	// 2. 用 Config 创建 DB
	db, err := NewDB(config)
	if err != nil {
		return nil, err
	}

	// 3. 用 DB 创建 UserRepository
	userRepository := NewUserRepository(db)

	// 4. 用 UserRepository 创建 UserService
	userService := NewUserService(userRepository)

	// 5. 用 UserService 创建 App
	app := NewApp(userService)

	return app, nil
}

// --- 4. 接口绑定演示 ---

// Logger 日志接口
type Logger interface {
	Log(msg string)
}

// ConsoleLogger 控制台日志实现
type ConsoleLogger struct{}

func (l *ConsoleLogger) Log(msg string) {
	fmt.Printf("  [ConsoleLogger] %s\n", msg)
}

// FileLogger 文件日志实现
type FileLogger struct {
	path string
}

func (l *FileLogger) Log(msg string) {
	fmt.Printf("  [FileLogger -> %s] %s\n", l.path, msg)
}

// NewConsoleLogger Provider
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

// NewFileLogger Provider
func NewFileLogger() *FileLogger {
	return &FileLogger{path: "/var/log/app.log"}
}

// Service 依赖 Logger 接口
type Service struct {
	logger Logger
}

func NewService(logger Logger) *Service {
	return &Service{logger: logger}
}

func main() {
	fmt.Println("=== Google Wire 依赖注入概念演示 ===")
	fmt.Println()

	// --- 1. 依赖注入链 ---
	fmt.Println("--- 1. 依赖注入链（模拟 Wire 生成的代码）---")
	fmt.Println("依赖链: Config → DB → UserRepository → UserService → App")
	fmt.Println()

	app, err := InitializeApp()
	if err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return
	}
	fmt.Printf("\n应用初始化成功: %+v\n", app)
	fmt.Println()

	// --- 2. 接口绑定 ---
	fmt.Println("--- 2. 接口绑定（Wire 的 wire.Bind）---")
	fmt.Println("在 Wire 中，通过 wire.Bind(new(Logger), new(*ConsoleLogger))")
	fmt.Println("将 Logger 接口绑定到 ConsoleLogger 实现")
	fmt.Println()

	// 绑定到 ConsoleLogger
	consoleService := NewService(NewConsoleLogger())
	consoleService.logger.Log("使用控制台日志")

	// 绑定到 FileLogger（切换实现只需修改 wire.Bind）
	fileService := NewService(NewFileLogger())
	fileService.logger.Log("使用文件日志")
	fmt.Println()

	// --- 3. Wire vs 手动 DI ---
	fmt.Println("--- 3. Wire vs 手动 DI 对比 ---")
	fmt.Println()
	fmt.Println("手动 DI（当前演示的方式）:")
	fmt.Println("  - 需要手动编写依赖组装代码")
	fmt.Println("  - 依赖链变长时，代码冗长且容易出错")
	fmt.Println()
	fmt.Println("Wire 自动 DI:")
	fmt.Println("  - 只需编写 Provider 函数和 Injector 签名")
	fmt.Println("  - Wire 自动分析依赖关系并生成组装代码")
	fmt.Println("  - 编译时完成，零运行时开销")
	fmt.Println("  - 编译时发现依赖错误（如循环依赖、缺失依赖）")
	fmt.Println()
	fmt.Println("Wire 核心概念:")
	fmt.Println("  - Provider: 提供依赖的函数（如 NewDB、NewUserService）")
	fmt.Println("  - Provider Set: 将相关 Provider 组织在一起（wire.NewSet）")
	fmt.Println("  - Injector: 依赖注入入口函数（wire.Build）")
	fmt.Println("  - Bind: 将接口绑定到具体实现（wire.Bind）")
	fmt.Println()
	fmt.Println("实际应用: B 站 Kratos 框架、Google 内部 Go 项目")
}
