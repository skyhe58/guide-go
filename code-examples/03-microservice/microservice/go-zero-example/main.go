// Go-Zero 微服务架构模式演示
// 本示例用纯 Go 标准库模拟 Go-Zero 框架的核心架构模式，无需安装 goctl CLI。
// 演示内容：API 网关 + RPC 服务概念、Handler/Logic/Svc 分层、
// 内置中间件（Auth/Logging/Breaker）、服务治理（限流/熔断/负载均衡）。
//
// Go 1.22+
// 验证日期：2025-01-01
//
// 运行方式：go run ./go-zero-example/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ============================================================================
// 配置管理（模拟 Go-Zero etc/*.yaml 配置）
// Go-Zero 使用 YAML 配置文件，通过 conf.MustLoad 加载。
// ============================================================================

// RestConf 模拟 Go-Zero 的 rest.RestConf
type RestConf struct {
	Host    string
	Port    int
	Timeout int64 // 毫秒
}

// RpcClientConf 模拟 Go-Zero 的 zrpc.RpcClientConf
type RpcClientConf struct {
	Target string
}

// Config 模拟 Go-Zero 的 config.Config（API 网关配置）
type Config struct {
	Rest    RestConf
	UserRpc RpcClientConf
}

// ============================================================================
// 服务治理 — 令牌桶限流器（模拟 Go-Zero 内置限流）
// Go-Zero 使用自适应限流，基于 CPU 使用率动态调整。
// 这里用令牌桶算法演示限流概念。
// ============================================================================

// TokenBucketLimiter 令牌桶限流器
type TokenBucketLimiter struct {
	rate       float64   // 每秒生成令牌数
	capacity   float64   // 桶容量
	tokens     float64   // 当前令牌数
	lastRefill time.Time // 上次补充时间
	mu         sync.Mutex
}

// NewTokenBucketLimiter 创建令牌桶限流器
func NewTokenBucketLimiter(rate float64, capacity float64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

// Allow 判断是否允许请求通过
func (l *TokenBucketLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastRefill).Seconds()
	l.tokens = math.Min(l.capacity, l.tokens+elapsed*l.rate)
	l.lastRefill = now

	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

// ============================================================================
// 服务治理 — 熔断器（模拟 Go-Zero 内置熔断，基于 Google SRE 算法）
// Go-Zero 的熔断器不使用传统的状态机（Open/Closed/HalfOpen），
// 而是使用概率性丢弃，更加平滑。
// ============================================================================

// GoogleBreaker Google SRE 风格熔断器
type GoogleBreaker struct {
	requests int64 // 总请求数（滑动窗口）
	accepts  int64 // 成功请求数
	k        float64 // 倍率参数，默认 1.5
	mu       sync.Mutex
}

// NewGoogleBreaker 创建 Google SRE 熔断器
func NewGoogleBreaker() *GoogleBreaker {
	return &GoogleBreaker{k: 1.5}
}

// dropRatio 计算丢弃概率
// 公式：max(0, (requests - K * accepts) / (requests + 1))
func (b *GoogleBreaker) dropRatio() float64 {
	if b.requests == 0 {
		return 0
	}
	ratio := (float64(b.requests) - b.k*float64(b.accepts)) / (float64(b.requests) + 1)
	return math.Max(0, ratio)
}

// Allow 判断请求是否被熔断
func (b *GoogleBreaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	dr := b.dropRatio()
	if dr <= 0 {
		return true
	}
	// 概率性丢弃
	return rand.Float64() >= dr
}

// MarkSuccess 标记请求成功
func (b *GoogleBreaker) MarkSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	atomic.AddInt64(&b.requests, 1)
	atomic.AddInt64(&b.accepts, 1)
}

// MarkFailure 标记请求失败
func (b *GoogleBreaker) MarkFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	atomic.AddInt64(&b.requests, 1)
}

// Stats 获取熔断器统计信息
func (b *GoogleBreaker) Stats() map[string]interface{} {
	b.mu.Lock()
	defer b.mu.Unlock()
	return map[string]interface{}{
		"requests":   b.requests,
		"accepts":    b.accepts,
		"drop_ratio": fmt.Sprintf("%.2f%%", b.dropRatio()*100),
	}
}

// ============================================================================
// 服务治理 — P2C 负载均衡（模拟 Go-Zero 内置负载均衡）
// P2C（Power of Two Choices）：随机选两个节点，选延迟更低的那个。
// Go-Zero 使用 EWMA（指数加权移动平均）计算节点延迟。
// ============================================================================

// Node RPC 服务节点
type Node struct {
	Addr    string
	Latency time.Duration // EWMA 延迟
	Inflight int64        // 正在处理的请求数
}

// P2CBalancer P2C 负载均衡器
type P2CBalancer struct {
	nodes []*Node
	mu    sync.Mutex
}

// NewP2CBalancer 创建 P2C 负载均衡器
func NewP2CBalancer(addrs []string) *P2CBalancer {
	nodes := make([]*Node, len(addrs))
	for i, addr := range addrs {
		nodes[i] = &Node{
			Addr:    addr,
			Latency: time.Millisecond * time.Duration(10+rand.Intn(50)),
		}
	}
	return &P2CBalancer{nodes: nodes}
}

// Pick 选择一个节点（P2C 算法）
func (b *P2CBalancer) Pick() *Node {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.nodes) == 0 {
		return nil
	}
	if len(b.nodes) == 1 {
		return b.nodes[0]
	}

	// 随机选两个节点
	i := rand.Intn(len(b.nodes))
	j := rand.Intn(len(b.nodes))
	for j == i {
		j = rand.Intn(len(b.nodes))
	}

	// 选延迟更低的（考虑 inflight 请求数）
	nodeA := b.nodes[i]
	nodeB := b.nodes[j]
	loadA := float64(nodeA.Latency) * float64(nodeA.Inflight+1)
	loadB := float64(nodeB.Latency) * float64(nodeB.Inflight+1)

	if loadA <= loadB {
		atomic.AddInt64(&nodeA.Inflight, 1)
		return nodeA
	}
	atomic.AddInt64(&nodeB.Inflight, 1)
	return nodeB
}

// ============================================================================
// RPC 服务层（模拟 Go-Zero 的 user-rpc 服务）
// 在 Go-Zero 中，RPC 服务由 .proto 文件定义，goctl 生成代码骨架。
// ============================================================================

// UserModel 用户数据模型（模拟 goctl model 生成的代码）
type UserModel struct {
	mu     sync.RWMutex
	users  map[int64]*UserEntity
	nextID int64
}

// UserEntity 用户实体
type UserEntity struct {
	ID       int64     `json:"id"`
	Username string    `json:"username"`
	Nickname string    `json:"nickname"`
	Email    string    `json:"email"`
	Status   int       `json:"status"` // 0:禁用 1:正常
	CreateAt time.Time `json:"create_at"`
}

// NewUserModel 创建用户模型（模拟数据库 + 缓存层）
func NewUserModel() *UserModel {
	m := &UserModel{
		users:  make(map[int64]*UserEntity),
		nextID: 1,
	}
	// 预置测试数据
	now := time.Now()
	m.users[1] = &UserEntity{ID: 1, Username: "alice", Nickname: "Alice", Email: "alice@example.com", Status: 1, CreateAt: now}
	m.users[2] = &UserEntity{ID: 2, Username: "bob", Nickname: "Bob", Email: "bob@example.com", Status: 1, CreateAt: now}
	m.users[3] = &UserEntity{ID: 3, Username: "charlie", Nickname: "Charlie", Email: "charlie@example.com", Status: 0, CreateAt: now}
	m.nextID = 4
	return m
}

// FindOne 查询单个用户（模拟 goctl model 生成的带缓存查询）
func (m *UserModel) FindOne(id int64) (*UserEntity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	user, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found: id=%d", id)
	}
	log.Printf("[MODEL] cache miss, query db: user_id=%d", id)
	return user, nil
}

// FindAll 查询所有用户
func (m *UserModel) FindAll() ([]*UserEntity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*UserEntity, 0, len(m.users))
	for _, u := range m.users {
		result = append(result, u)
	}
	return result, nil
}

// Insert 插入用户
func (m *UserModel) Insert(user *UserEntity) (*UserEntity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	user.ID = m.nextID
	m.nextID++
	user.CreateAt = time.Now()
	user.Status = 1
	m.users[user.ID] = user
	log.Printf("[MODEL] inserted user: id=%d username=%s", user.ID, user.Username)
	return user, nil
}

// ============================================================================
// RPC Server（模拟 Go-Zero 的 user-rpc 服务实现）
// ============================================================================

// UserRpcServer 模拟 Go-Zero RPC 服务
type UserRpcServer struct {
	model *UserModel
}

// NewUserRpcServer 创建 RPC 服务
func NewUserRpcServer(model *UserModel) *UserRpcServer {
	return &UserRpcServer{model: model}
}

// GetUser RPC 方法：获取用户
func (s *UserRpcServer) GetUser(_ context.Context, id int64) (*UserEntity, error) {
	return s.model.FindOne(id)
}

// ListUsers RPC 方法：获取用户列表
func (s *UserRpcServer) ListUsers(_ context.Context) ([]*UserEntity, error) {
	return s.model.FindAll()
}

// CreateUser RPC 方法：创建用户
func (s *UserRpcServer) CreateUser(_ context.Context, user *UserEntity) (*UserEntity, error) {
	return s.model.Insert(user)
}

// ============================================================================
// ServiceContext（模拟 Go-Zero 的 svc.ServiceContext）
// ServiceContext 是 Go-Zero 的依赖注入容器，在服务启动时初始化所有依赖。
// ============================================================================

// ServiceContext Go-Zero 风格的服务上下文
type ServiceContext struct {
	Config     Config
	UserRpc    *UserRpcServer
	Limiter    *TokenBucketLimiter
	Breaker    *GoogleBreaker
	Balancer   *P2CBalancer
}

// NewServiceContext 创建服务上下文（Go-Zero 中在 svc/servicecontext.go 中定义）
func NewServiceContext(cfg Config) *ServiceContext {
	model := NewUserModel()
	return &ServiceContext{
		Config:   cfg,
		UserRpc:  NewUserRpcServer(model),
		Limiter:  NewTokenBucketLimiter(100, 200), // 每秒 100 个令牌，桶容量 200
		Breaker:  NewGoogleBreaker(),
		Balancer: NewP2CBalancer([]string{
			"user-rpc-1:9090",
			"user-rpc-2:9090",
			"user-rpc-3:9090",
		}),
	}
}

// ============================================================================
// Logic 层（模拟 Go-Zero 的 internal/logic/）
// Logic 层是业务逻辑的核心，goctl 生成骨架代码，开发者在这里填充业务逻辑。
// ============================================================================

// GetUserLogic 获取用户逻辑
type GetUserLogic struct {
	ctx    context.Context
	svcCtx *ServiceContext
}

// NewGetUserLogic 创建逻辑实例
func NewGetUserLogic(ctx context.Context, svcCtx *ServiceContext) *GetUserLogic {
	return &GetUserLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetUser 业务逻辑实现
func (l *GetUserLogic) GetUser(id int64) (*UserEntity, error) {
	// P2C 负载均衡选择 RPC 节点
	node := l.svcCtx.Balancer.Pick()
	log.Printf("[LOGIC] P2C selected node: %s (latency=%v)", node.Addr, node.Latency)

	// 调用 RPC 服务
	user, err := l.svcCtx.UserRpc.GetUser(l.ctx, id)
	if err != nil {
		l.svcCtx.Breaker.MarkFailure()
		return nil, err
	}
	l.svcCtx.Breaker.MarkSuccess()
	atomic.AddInt64(&node.Inflight, -1)
	return user, nil
}

// ListUsersLogic 获取用户列表逻辑
type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *ServiceContext
}

// NewListUsersLogic 创建逻辑实例
func NewListUsersLogic(ctx context.Context, svcCtx *ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListUsers 业务逻辑实现
func (l *ListUsersLogic) ListUsers() ([]*UserEntity, error) {
	return l.svcCtx.UserRpc.ListUsers(l.ctx)
}

// CreateUserLogic 创建用户逻辑
type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *ServiceContext
}

// NewCreateUserLogic 创建逻辑实例
func NewCreateUserLogic(ctx context.Context, svcCtx *ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{ctx: ctx, svcCtx: svcCtx}
}

// CreateUser 业务逻辑实现
func (l *CreateUserLogic) CreateUser(username, nickname, email string) (*UserEntity, error) {
	if username == "" || email == "" {
		return nil, fmt.Errorf("username and email are required")
	}
	user := &UserEntity{
		Username: username,
		Nickname: nickname,
		Email:    email,
	}
	return l.svcCtx.UserRpc.CreateUser(l.ctx, user)
}

// ============================================================================
// Handler 层（模拟 Go-Zero 的 internal/handler/）
// Handler 层由 goctl 自动生成，负责 HTTP 请求的参数解析和响应封装。
// ============================================================================

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// httpResult 统一响应函数（模拟 Go-Zero 的 httpx.OkJson / httpx.Error）
func httpResult(w http.ResponseWriter, code int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

// GetUserHandler 获取用户 Handler（goctl 自动生成）
func GetUserHandler(svcCtx *ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			httpResult(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid user id"})
			return
		}

		logic := NewGetUserLogic(r.Context(), svcCtx)
		user, err := logic.GetUser(id)
		if err != nil {
			httpResult(w, http.StatusNotFound, Response{Code: 404, Message: err.Error()})
			return
		}
		httpResult(w, http.StatusOK, Response{Code: 0, Message: "ok", Data: user})
	}
}

// ListUsersHandler 获取用户列表 Handler
func ListUsersHandler(svcCtx *ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logic := NewListUsersLogic(r.Context(), svcCtx)
		users, err := logic.ListUsers()
		if err != nil {
			httpResult(w, http.StatusInternalServerError, Response{Code: 500, Message: err.Error()})
			return
		}
		httpResult(w, http.StatusOK, Response{Code: 0, Message: "ok", Data: users})
	}
}

// CreateUserHandler 创建用户 Handler
func CreateUserHandler(svcCtx *ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Username string `json:"username"`
			Nickname string `json:"nickname"`
			Email    string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpResult(w, http.StatusBadRequest, Response{Code: 400, Message: "invalid request body"})
			return
		}

		logic := NewCreateUserLogic(r.Context(), svcCtx)
		user, err := logic.CreateUser(req.Username, req.Nickname, req.Email)
		if err != nil {
			httpResult(w, http.StatusBadRequest, Response{Code: 400, Message: err.Error()})
			return
		}
		httpResult(w, http.StatusCreated, Response{Code: 0, Message: "created", Data: user})
	}
}

// StatsHandler 服务治理统计信息
func StatsHandler(svcCtx *ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		stats := map[string]interface{}{
			"breaker":  svcCtx.Breaker.Stats(),
			"limiter":  map[string]interface{}{"rate": 100, "capacity": 200},
			"balancer": map[string]interface{}{"algorithm": "P2C", "nodes": len(svcCtx.Balancer.nodes)},
		}
		httpResult(w, http.StatusOK, Response{Code: 0, Message: "ok", Data: stats})
	}
}

// ============================================================================
// 中间件（模拟 Go-Zero 内置中间件）
// ============================================================================

// AuthMiddleware JWT 认证中间件（模拟 Go-Zero 的 rest.WithJwt）
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			// 演示模式：无 Token 也放行，但记录日志
			log.Printf("[AUTH] no token provided, path=%s (demo mode: allowed)", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}
		log.Printf("[AUTH] token verified: %s...", token[:min(20, len(token))])
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware 请求日志中间件
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[HTTP] --> %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("[HTTP] <-- %s %s (%v)", r.Method, r.URL.Path, time.Since(start))
	})
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(limiter *TokenBucketLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				log.Printf("[LIMITER] request rejected: %s %s", r.Method, r.URL.Path)
				httpResult(w, http.StatusServiceUnavailable, Response{
					Code:    503,
					Message: "service overloaded, please retry later",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// BreakerMiddleware 熔断中间件
func BreakerMiddleware(breaker *GoogleBreaker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !breaker.Allow() {
				log.Printf("[BREAKER] request circuit broken: %s %s", r.Method, r.URL.Path)
				httpResult(w, http.StatusServiceUnavailable, Response{
					Code:    503,
					Message: "service circuit broken, please retry later",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryMiddleware Panic 恢复中间件
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[RECOVERY] panic recovered: %v", err)
				httpResult(w, http.StatusInternalServerError, Response{
					Code:    500,
					Message: "internal server error",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// chainHTTPMiddleware 组合 HTTP 中间件
func chainHTTPMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// ============================================================================
// API 网关（模拟 Go-Zero 的 rest.MustNewServer）
// Go-Zero 的 API 网关负责路由分发、中间件执行、请求转发到 RPC 服务。
// ============================================================================

func main() {
	fmt.Println("=== Go-Zero 微服务架构模式演示 ===")
	fmt.Println()
	fmt.Println("本示例演示 Go-Zero 框架的核心架构模式：")
	fmt.Println("  1. API 网关 + RPC 服务分层架构")
	fmt.Println("  2. Handler → Logic → Model 代码分层（goctl 生成）")
	fmt.Println("  3. ServiceContext 依赖注入容器")
	fmt.Println("  4. 内置中间件：Auth / Logging / Recovery")
	fmt.Println("  5. 服务治理：令牌桶限流 / Google SRE 熔断 / P2C 负载均衡")
	fmt.Println()

	// 1. 加载配置（模拟 conf.MustLoad）
	cfg := Config{
		Rest:    RestConf{Host: "0.0.0.0", Port: 8888, Timeout: 5000},
		UserRpc: RpcClientConf{Target: "user-rpc:9090"},
	}

	// 2. 创建 ServiceContext（依赖注入）
	svcCtx := NewServiceContext(cfg)
	log.Printf("[CONFIG] API gateway: %s:%d, RPC target: %s",
		cfg.Rest.Host, cfg.Rest.Port, cfg.UserRpc.Target)

	// 3. 注册路由（模拟 goctl 生成的路由注册）
	mux := http.NewServeMux()

	// 公开路由（无需认证）
	mux.HandleFunc("GET /api/v1/users", ListUsersHandler(svcCtx))
	mux.HandleFunc("GET /api/v1/users/{id}", GetUserHandler(svcCtx))

	// 需要认证的路由
	mux.HandleFunc("POST /api/v1/users", CreateUserHandler(svcCtx))

	// 服务治理端点
	mux.HandleFunc("GET /api/v1/stats", StatsHandler(svcCtx))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpResult(w, http.StatusOK, Response{Code: 0, Message: "ok"})
	})

	// 4. 组装中间件链（模拟 Go-Zero 内置中间件）
	handler := chainHTTPMiddleware(
		mux,
		RecoveryMiddleware,                       // 最外层：Panic 恢复
		LoggingMiddleware,                        // 请求日志
		RateLimitMiddleware(svcCtx.Limiter),      // 限流
		BreakerMiddleware(svcCtx.Breaker),        // 熔断
		AuthMiddleware,                           // JWT 认证
	)

	// 5. 启动 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", cfg.Rest.Host, cfg.Rest.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.Rest.Timeout) * time.Millisecond,
		WriteTimeout: time.Duration(cfg.Rest.Timeout) * time.Millisecond,
	}

	go func() {
		log.Printf("[SERVER] API gateway listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	fmt.Println()
	fmt.Println("服务已启动，可以使用以下命令测试：")
	fmt.Println("  curl http://localhost:8888/healthz")
	fmt.Println("  curl http://localhost:8888/api/v1/users")
	fmt.Println("  curl http://localhost:8888/api/v1/users/1")
	fmt.Println("  curl -X POST http://localhost:8888/api/v1/users \\")
	fmt.Println("    -H 'Content-Type: application/json' \\")
	fmt.Println("    -d '{\"username\":\"dave\",\"nickname\":\"Dave\",\"email\":\"dave@example.com\"}'")
	fmt.Println("  curl http://localhost:8888/api/v1/stats  # 查看服务治理统计")
	fmt.Println("  curl http://localhost:8888/api/v1/users/999  # 测试错误处理")
	fmt.Println()
	fmt.Println("按 Ctrl+C 优雅停止服务...")

	// 6. 优雅停机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[SERVER] shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[SERVER] shutdown error: %v", err)
	}
	log.Println("[SERVER] exited gracefully")
}
