// Kratos 微服务架构模式演示
// 本示例用纯 Go 标准库模拟 Kratos 框架的核心架构模式，无需安装 Kratos CLI。
// 演示内容：分层架构（Service/Biz/Data）、Wire 风格依赖注入、HTTP+gRPC 双协议 Transport、
// 中间件链（Logging/Recovery/Tracing）、Kratos 风格错误处理、配置管理。
//
// Go 1.22+
// 验证日期：2025-01-01
//
// 运行方式：go run ./kratos-example/

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ============================================================================
// 配置管理（模拟 Kratos Config 组件）
// Kratos 支持多种配置源（本地文件、etcd、Consul、Nacos），通过统一接口加载和热更新。
// ============================================================================

// AppConfig 应用配置结构，对应 Kratos 中 internal/conf/conf.proto 定义的配置
type AppConfig struct {
	Server ServerConfig `json:"server"`
	Data   DataConfig   `json:"data"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	HTTP HTTPConfig `json:"http"`
	GRPC GRPCConfig `json:"grpc"`
}

// HTTPConfig HTTP 服务配置
type HTTPConfig struct {
	Addr    string        `json:"addr"`
	Timeout time.Duration `json:"timeout"`
}

// GRPCConfig gRPC 服务配置
type GRPCConfig struct {
	Addr    string        `json:"addr"`
	Timeout time.Duration `json:"timeout"`
}

// DataConfig 数据层配置
type DataConfig struct {
	Database DatabaseConfig `json:"database"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver string `json:"driver"`
	Source string `json:"source"`
}

// loadConfig 加载配置（模拟 Kratos 的 config.New + config.Load）
// 实际 Kratos 项目中通过 config.New(config.WithSource(file.NewSource(path))) 加载
func loadConfig() *AppConfig {
	return &AppConfig{
		Server: ServerConfig{
			HTTP: HTTPConfig{Addr: ":8080", Timeout: 5 * time.Second},
			GRPC: GRPCConfig{Addr: ":9090", Timeout: 5 * time.Second},
		},
		Data: DataConfig{
			Database: DatabaseConfig{
				Driver: "mysql",
				Source: "root:root123@tcp(localhost:3306)/kratos_demo?parseTime=true",
			},
		},
	}
}

// ============================================================================
// 错误处理（模拟 Kratos errors 包）
// Kratos 定义了统一的错误模型，基于 gRPC Status 扩展，包含业务错误码和原因。
// ============================================================================

// AppError Kratos 风格的错误结构
// 实际 Kratos 中通过 Proto 定义错误枚举，自动生成错误构造函数
type AppError struct {
	HTTPCode int               // HTTP 状态码
	Code     int32             // 业务错误码
	Reason   string            // 错误原因（枚举值，如 USER_NOT_FOUND）
	Message  string            // 用户可读的错误信息
	Metadata map[string]string // 附加元数据
}

func (e *AppError) Error() string {
	return fmt.Sprintf("error: code=%d reason=%s message=%s", e.Code, e.Reason, e.Message)
}

// 预定义错误（模拟 Kratos 中 Proto 生成的错误枚举）
var (
	ErrUserNotFound = &AppError{
		HTTPCode: http.StatusNotFound,
		Code:     404,
		Reason:   "USER_NOT_FOUND",
		Message:  "用户不存在",
	}
	ErrInvalidParam = &AppError{
		HTTPCode: http.StatusBadRequest,
		Code:     400,
		Reason:   "INVALID_PARAM",
		Message:  "请求参数无效",
	}
	ErrInternalServer = &AppError{
		HTTPCode: http.StatusInternalServerError,
		Code:     500,
		Reason:   "INTERNAL_SERVER",
		Message:  "服务内部错误",
	}
)

// IsAppError 判断是否为 AppError 类型
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// ============================================================================
// Data 层（模拟 Kratos internal/data/）
// Data 层负责数据访问，实现 Biz 层定义的 Repository 接口。
// 实际项目中这里会连接 MySQL/PostgreSQL/Redis 等数据源。
// ============================================================================

// User 用户实体
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// UserRepo 用户数据仓库接口（定义在 Biz 层，实现在 Data 层 —— 依赖倒置）
type UserRepo interface {
	GetUser(ctx context.Context, id int64) (*User, error)
	ListUsers(ctx context.Context) ([]*User, error)
	CreateUser(ctx context.Context, user *User) (*User, error)
}

// userRepo UserRepo 的内存实现（模拟数据库访问）
type userRepo struct {
	mu    sync.RWMutex
	users map[int64]*User
	nextID int64
}

// NewUserRepo 构造函数（Wire Provider）
// 在 Kratos 中，这个函数会被注册到 Wire ProviderSet 中
func NewUserRepo() UserRepo {
	repo := &userRepo{
		users:  make(map[int64]*User),
		nextID: 1,
	}
	// 预置测试数据
	now := time.Now()
	repo.users[1] = &User{ID: 1, Username: "alice", Email: "alice@example.com", CreatedAt: now}
	repo.users[2] = &User{ID: 2, Username: "bob", Email: "bob@example.com", CreatedAt: now}
	repo.users[3] = &User{ID: 3, Username: "charlie", Email: "charlie@example.com", CreatedAt: now}
	repo.nextID = 4
	return repo
}

func (r *userRepo) GetUser(_ context.Context, id int64) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (r *userRepo) ListUsers(_ context.Context) ([]*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make([]*User, 0, len(r.users))
	for _, u := range r.users {
		users = append(users, u)
	}
	return users, nil
}

func (r *userRepo) CreateUser(_ context.Context, user *User) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user.ID = r.nextID
	r.nextID++
	user.CreatedAt = time.Now()
	r.users[user.ID] = user
	return user, nil
}

// ============================================================================
// Biz 层（模拟 Kratos internal/biz/）
// Biz 层包含核心业务逻辑，定义 Repository 接口（由 Data 层实现）。
// 这是 DDD 中的 Domain 层。
// ============================================================================

// UserUsecase 用户业务逻辑（Kratos 中称为 Usecase）
type UserUsecase struct {
	repo UserRepo
}

// NewUserUsecase 构造函数（Wire Provider）
func NewUserUsecase(repo UserRepo) *UserUsecase {
	return &UserUsecase{repo: repo}
}

// GetUser 获取用户（业务逻辑层，可以在这里添加缓存、权限校验等）
func (uc *UserUsecase) GetUser(ctx context.Context, id int64) (*User, error) {
	if id <= 0 {
		return nil, ErrInvalidParam
	}
	return uc.repo.GetUser(ctx, id)
}

// ListUsers 获取用户列表
func (uc *UserUsecase) ListUsers(ctx context.Context) ([]*User, error) {
	return uc.repo.ListUsers(ctx)
}

// CreateUser 创建用户（包含业务校验逻辑）
func (uc *UserUsecase) CreateUser(ctx context.Context, username, email string) (*User, error) {
	if username == "" || email == "" {
		return nil, ErrInvalidParam
	}
	user := &User{Username: username, Email: email}
	return uc.repo.CreateUser(ctx, user)
}

// ============================================================================
// Service 层（模拟 Kratos internal/service/）
// Service 层是 API 的实现层，负责参数转换和调用 Biz 层。
// 在 Kratos 中，Service 层实现 Proto 生成的接口。
// ============================================================================

// UserService 用户服务（实现 Proto 生成的 UserServiceServer 接口）
type UserService struct {
	uc *UserUsecase
}

// NewUserService 构造函数（Wire Provider）
func NewUserService(uc *UserUsecase) *UserService {
	return &UserService{uc: uc}
}

// ============================================================================
// Middleware 中间件（模拟 Kratos middleware 包）
// Kratos 中间件采用洋葱模型，签名为 func(Handler) Handler。
// ============================================================================

// Handler Kratos 风格的处理函数签名
type Handler func(ctx context.Context, req interface{}) (interface{}, error)

// Middleware Kratos 风格的中间件签名
type Middleware func(Handler) Handler

// 请求上下文键
type contextKey string

const (
	requestIDKey contextKey = "x-request-id"
	traceIDKey   contextKey = "x-trace-id"
)

// Recovery 恢复中间件（必须放在最外层）
// 捕获 panic，防止单个请求的 panic 导致整个服务崩溃
func Recovery() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req interface{}) (resp interface{}, err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[RECOVERY] panic recovered: %v", r)
					err = ErrInternalServer
				}
			}()
			return next(ctx, req)
		}
	}
}

// Logging 日志中间件
// 记录请求方法、路径、耗时、错误信息
func Logging() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			start := time.Now()
			reqID, _ := ctx.Value(requestIDKey).(string)

			resp, err := next(ctx, req)

			duration := time.Since(start)
			if err != nil {
				log.Printf("[LOG] request_id=%s duration=%v error=%v", reqID, duration, err)
			} else {
				log.Printf("[LOG] request_id=%s duration=%v status=ok", reqID, duration)
			}
			return resp, err
		}
	}
}

// Tracing 链路追踪中间件（模拟 OpenTelemetry 集成）
// 实际 Kratos 中通过 tracing.Server() 中间件集成 OpenTelemetry
func Tracing() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 生成或提取 TraceID
			traceID := fmt.Sprintf("trace-%d", rand.Int63n(1000000))
			ctx = context.WithValue(ctx, traceIDKey, traceID)
			log.Printf("[TRACE] span started: trace_id=%s", traceID)

			resp, err := next(ctx, req)

			log.Printf("[TRACE] span ended: trace_id=%s", traceID)
			return resp, err
		}
	}
}

// chainMiddleware 将多个中间件组合成链（洋葱模型）
// 执行顺序：第一个中间件最先执行（最外层），最后一个中间件最后执行（最内层）
func chainMiddleware(handler Handler, middlewares ...Middleware) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// ============================================================================
// Transport 层 — HTTP Server（模拟 Kratos transport/http）
// Kratos 的 HTTP Transport 基于 gorilla/mux，支持 Proto 生成的路由注册。
// ============================================================================

// HTTPServer 模拟 Kratos HTTP Transport
type HTTPServer struct {
	server      *http.Server
	userService *UserService
	middlewares []Middleware
}

// NewHTTPServer 创建 HTTP 服务器（Wire Provider）
func NewHTTPServer(cfg *HTTPConfig, svc *UserService) *HTTPServer {
	s := &HTTPServer{
		userService: svc,
		middlewares: []Middleware{
			Recovery(), // 最外层：panic 恢复
			Logging(),  // 日志记录
			Tracing(),  // 链路追踪
		},
	}

	mux := http.NewServeMux()

	// 注册路由（模拟 Kratos 的 Proto 生成路由注册）
	mux.HandleFunc("GET /api/v1/users", s.handleListUsers)
	mux.HandleFunc("GET /api/v1/users/{id}", s.handleGetUser)
	mux.HandleFunc("POST /api/v1/users", s.handleCreateUser)
	mux.HandleFunc("GET /healthz", s.handleHealth)

	s.server = &http.Server{
		Addr:         cfg.Addr,
		Handler:      s.requestIDMiddleware(mux),
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	}
	return s
}

// requestIDMiddleware 为每个请求生成唯一 ID
func (s *HTTPServer) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = fmt.Sprintf("req-%d", time.Now().UnixNano())
		}
		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		w.Header().Set("X-Request-ID", reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// writeJSON 统一 JSON 响应
func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// writeError 统一错误响应（Kratos 风格）
func writeError(w http.ResponseWriter, err error) {
	if appErr, ok := IsAppError(err); ok {
		writeJSON(w, appErr.HTTPCode, map[string]interface{}{
			"code":    appErr.Code,
			"reason":  appErr.Reason,
			"message": appErr.Message,
		})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
		"code":    500,
		"reason":  "INTERNAL_SERVER",
		"message": err.Error(),
	})
}

// handleListUsers 获取用户列表
func (s *HTTPServer) handleListUsers(w http.ResponseWriter, r *http.Request) {
	// 通过中间件链执行业务逻辑
	handler := chainMiddleware(
		func(ctx context.Context, _ interface{}) (interface{}, error) {
			return s.userService.uc.ListUsers(ctx)
		},
		s.middlewares...,
	)

	resp, err := handler(r.Context(), nil)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

// handleGetUser 获取单个用户
func (s *HTTPServer) handleGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var id int64
	fmt.Sscanf(idStr, "%d", &id)

	handler := chainMiddleware(
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.userService.uc.GetUser(ctx, req.(int64))
		},
		s.middlewares...,
	)

	resp, err := handler(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

// handleCreateUser 创建用户
func (s *HTTPServer) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, ErrInvalidParam)
		return
	}

	handler := chainMiddleware(
		func(ctx context.Context, input interface{}) (interface{}, error) {
			params := input.(map[string]string)
			return s.userService.uc.CreateUser(ctx, params["username"], params["email"])
		},
		s.middlewares...,
	)

	resp, err := handler(r.Context(), map[string]string{
		"username": req.Username,
		"email":    req.Email,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"code":    0,
		"message": "created",
		"data":    resp,
	})
}

// handleHealth 健康检查端点
func (s *HTTPServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Start 启动 HTTP 服务器
func (s *HTTPServer) Start() error {
	log.Printf("[HTTP] server listening on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop 优雅停止 HTTP 服务器
func (s *HTTPServer) Stop(ctx context.Context) error {
	log.Println("[HTTP] server shutting down...")
	return s.server.Shutdown(ctx)
}

// ============================================================================
// Transport 层 — gRPC Server（模拟 Kratos transport/grpc）
// 这里用简单的 TCP 服务器模拟 gRPC Transport 的概念。
// 实际 Kratos 中使用 google.golang.org/grpc 包。
// ============================================================================

// GRPCServer 模拟 Kratos gRPC Transport
type GRPCServer struct {
	addr        string
	listener    net.Listener
	userService *UserService
	done        chan struct{}
}

// NewGRPCServer 创建 gRPC 服务器（Wire Provider）
func NewGRPCServer(cfg *GRPCConfig, svc *UserService) *GRPCServer {
	return &GRPCServer{
		addr:        cfg.Addr,
		userService: svc,
		done:        make(chan struct{}),
	}
}

// Start 启动 gRPC 服务器（模拟）
func (s *GRPCServer) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("gRPC listen failed: %w", err)
	}
	log.Printf("[gRPC] server listening on %s (模拟 — 接受连接后立即关闭)", s.addr)

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.done:
					return // 正常关闭
				default:
					if !strings.Contains(err.Error(), "use of closed network connection") {
						log.Printf("[gRPC] accept error: %v", err)
					}
					return
				}
			}
			// 模拟 gRPC 连接处理：实际项目中这里是 gRPC Server 的 Serve 逻辑
			log.Printf("[gRPC] new connection from %s", conn.RemoteAddr())
			conn.Close()
		}
	}()
	return nil
}

// Stop 优雅停止 gRPC 服务器
func (s *GRPCServer) Stop() error {
	log.Println("[gRPC] server shutting down...")
	close(s.done)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// ============================================================================
// Wire 风格依赖注入（模拟 Google Wire）
// 实际 Kratos 项目中，以下代码由 wire generate 自动生成（wire_gen.go）。
// 这里手动模拟 Wire 的依赖注入过程，展示 Wire 的工作原理。
// ============================================================================

// App 应用实例（Kratos 中的 kratos.App）
type App struct {
	httpServer *HTTPServer
	grpcServer *GRPCServer
}

// initApp 模拟 Wire Injector 函数
// 在 Kratos 中，这个函数在 wire.go 中声明，由 wire generate 生成实现
//
// 实际 wire.go 中的声明：
//
//	func initApp(*conf.Server, *conf.Data) (*kratos.App, func(), error) {
//	    panic(wire.Build(
//	        data.ProviderSet,    // Data 层 Provider 集合
//	        biz.ProviderSet,     // Biz 层 Provider 集合
//	        service.ProviderSet, // Service 层 Provider 集合
//	        server.ProviderSet,  // Server 层 Provider 集合
//	        newApp,              // App 构造函数
//	    ))
//	}
func initApp(cfg *AppConfig) (*App, func(), error) {
	// Wire 按依赖关系自动排列构造顺序：
	// 1. Data 层（无依赖）
	userRepo := NewUserRepo()

	// 2. Biz 层（依赖 Data 层）
	userUsecase := NewUserUsecase(userRepo)

	// 3. Service 层（依赖 Biz 层）
	userService := NewUserService(userUsecase)

	// 4. Transport 层（依赖 Service 层 + Config）
	httpServer := NewHTTPServer(&cfg.Server.HTTP, userService)
	grpcServer := NewGRPCServer(&cfg.Server.GRPC, userService)

	// 5. 构造 App
	app := &App{
		httpServer: httpServer,
		grpcServer: grpcServer,
	}

	// cleanup 函数（Wire 自动生成的资源清理函数）
	cleanup := func() {
		log.Println("[Wire] cleaning up resources...")
	}

	return app, cleanup, nil
}

// ============================================================================
// 主函数（模拟 Kratos cmd/myservice/main.go）
// ============================================================================

func main() {
	fmt.Println("=== Kratos 微服务架构模式演示 ===")
	fmt.Println()
	fmt.Println("本示例演示 Kratos 框架的核心架构模式：")
	fmt.Println("  1. 分层架构：API → Service → Biz → Data")
	fmt.Println("  2. Wire 风格依赖注入（编译时 DI）")
	fmt.Println("  3. Transport 层：HTTP + gRPC 双协议")
	fmt.Println("  4. 中间件链：Recovery → Logging → Tracing")
	fmt.Println("  5. Kratos 风格错误处理（业务错误码 + Reason）")
	fmt.Println("  6. 配置管理（多源支持 + 热更新概念）")
	fmt.Println()

	// 1. 加载配置（模拟 Kratos config.Load）
	cfg := loadConfig()
	log.Printf("[CONFIG] loaded: http=%s grpc=%s db=%s",
		cfg.Server.HTTP.Addr, cfg.Server.GRPC.Addr, cfg.Data.Database.Driver)

	// 2. Wire 依赖注入（模拟 wire generate 生成的 initApp）
	app, cleanup, err := initApp(cfg)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}
	defer cleanup()

	// 3. 启动 gRPC 服务器
	if err := app.grpcServer.Start(); err != nil {
		log.Fatalf("failed to start gRPC server: %v", err)
	}

	// 4. 启动 HTTP 服务器（非阻塞）
	go func() {
		if err := app.httpServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start HTTP server: %v", err)
		}
	}()

	fmt.Println()
	fmt.Println("服务已启动，可以使用以下命令测试：")
	fmt.Println("  curl http://localhost:8080/healthz")
	fmt.Println("  curl http://localhost:8080/api/v1/users")
	fmt.Println("  curl http://localhost:8080/api/v1/users/1")
	fmt.Println("  curl -X POST http://localhost:8080/api/v1/users -d '{\"username\":\"dave\",\"email\":\"dave@example.com\"}'")
	fmt.Println("  curl http://localhost:8080/api/v1/users/999  # 测试错误处理")
	fmt.Println()
	fmt.Println("按 Ctrl+C 优雅停止服务...")

	// 5. 优雅停机（模拟 Kratos 的 signal 处理）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[APP] shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 按顺序停止服务：先停 HTTP，再停 gRPC
	if err := app.httpServer.Stop(ctx); err != nil {
		log.Printf("[APP] HTTP server shutdown error: %v", err)
	}
	if err := app.grpcServer.Stop(); err != nil {
		log.Printf("[APP] gRPC server shutdown error: %v", err)
	}

	log.Println("[APP] server exited gracefully")
}
