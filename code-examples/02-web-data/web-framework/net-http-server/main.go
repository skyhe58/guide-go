// net/http 标准库 HTTP 服务器示例
// 演示：路由（Go 1.22 增强）、中间件链、优雅关闭
// Go 1.22+ | 验证日期 2025-01-01
//
// 运行方式：go run main.go
// 测试：curl http://localhost:8080/api/users
//       curl http://localhost:8080/api/users/42
//       curl -X POST -H "Content-Type: application/json" -d '{"name":"Go开发者","email":"go@example.com"}' http://localhost:8080/api/users

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

// ============================================================
// 数据模型
// ============================================================

// User 用户模型
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ============================================================
// 中间件实现
// ============================================================

// Middleware 中间件类型定义
type Middleware func(http.Handler) http.Handler

// Chain 将多个中间件串联成链
// 执行顺序：第一个中间件最先执行（洋葱模型外层）
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	// 从后往前包装，确保第一个中间件在最外层
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// requestID 全局请求 ID 计数器
var requestID atomic.Int64

// RequestIDMiddleware 请求 ID 中间件
// 为每个请求分配唯一 ID，方便日志追踪
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestID.Add(1)
		// 将请求 ID 写入响应头
		w.Header().Set("X-Request-ID", fmt.Sprintf("%d", id))
		// 将请求 ID 存入 context，供后续中间件和 Handler 使用
		ctx := context.WithValue(r.Context(), "requestID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware 请求日志中间件
// 记录请求方法、路径、状态码和耗时
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装 ResponseWriter 以捕获状态码
		wrapped := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		reqID, _ := r.Context().Value("requestID").(int64)
		log.Printf("[请求日志] ID=%d | %s %s | 状态=%d | 耗时=%v",
			reqID, r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

// RecoveryMiddleware panic 恢复中间件
// 捕获 Handler 中的 panic，返回 500 错误而不是崩溃
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[Recovery] panic 已恢复: %v", err)
				writeJSON(w, http.StatusInternalServerError, Response{
					Code:    500,
					Message: "服务器内部错误",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware 跨域中间件
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// 预检请求直接返回
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// statusResponseWriter 包装 ResponseWriter 以捕获状态码
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// ============================================================
// 路由处理函数（Handler）
// ============================================================

// handleListUsers 获取用户列表
func handleListUsers(w http.ResponseWriter, r *http.Request) {
	// 模拟用户数据
	users := []User{
		{ID: 1, Name: "张三", Email: "zhangsan@example.com"},
		{ID: 2, Name: "李四", Email: "lisi@example.com"},
		{ID: 3, Name: "王五", Email: "wangwu@example.com"},
	}

	writeJSON(w, http.StatusOK, Response{
		Code:    200,
		Message: "获取用户列表成功",
		Data:    users,
	})
}

// handleGetUser 根据 ID 获取用户（Go 1.22 路径参数）
func handleGetUser(w http.ResponseWriter, r *http.Request) {
	// Go 1.22+ 使用 PathValue 获取路径参数
	id := r.PathValue("id")

	user := User{
		ID:    1,
		Name:  fmt.Sprintf("用户_%s", id),
		Email: fmt.Sprintf("user%s@example.com", id),
	}

	writeJSON(w, http.StatusOK, Response{
		Code:    200,
		Message: "获取用户成功",
		Data:    user,
	})
}

// handleCreateUser 创建用户
func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数格式错误: " + err.Error(),
		})
		return
	}

	// 简单验证
	if user.Name == "" {
		writeJSON(w, http.StatusBadRequest, Response{
			Code:    400,
			Message: "用户名不能为空",
		})
		return
	}

	user.ID = 100 // 模拟分配 ID

	writeJSON(w, http.StatusCreated, Response{
		Code:    201,
		Message: "创建用户成功",
		Data:    user,
	})
}

// handleHealthCheck 健康检查端点
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Response{
		Code:    200,
		Message: "服务运行正常",
		Data: map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		},
	})
}

// ============================================================
// 工具函数
// ============================================================

// writeJSON 写入 JSON 响应
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// ============================================================
// 主函数：启动服务器 + 优雅关闭
// ============================================================

func main() {
	// 创建路由
	mux := http.NewServeMux()

	// 注册路由（Go 1.22+ 增强路由：支持方法匹配和路径参数）
	mux.HandleFunc("GET /api/users", handleListUsers)
	mux.HandleFunc("GET /api/users/{id}", handleGetUser)
	mux.HandleFunc("POST /api/users", handleCreateUser)
	mux.HandleFunc("GET /health", handleHealthCheck)

	// 应用中间件链（执行顺序：CORS → Recovery → RequestID → Logging → Handler）
	handler := Chain(mux,
		CORSMiddleware,
		RecoveryMiddleware,
		RequestIDMiddleware,
		LoggingMiddleware,
	)

	// 配置 HTTP 服务器（生产环境必须设置超时）
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second, // 读取请求超时
		WriteTimeout: 15 * time.Second, // 写入响应超时
		IdleTimeout:  60 * time.Second, // 空闲连接超时
	}

	// 在 goroutine 中启动服务器
	go func() {
		log.Printf("🚀 HTTP 服务器启动在 http://localhost%s", srv.Addr)
		log.Println("📋 可用端点:")
		log.Println("   GET  /api/users      - 获取用户列表")
		log.Println("   GET  /api/users/{id}  - 获取指定用户")
		log.Println("   POST /api/users      - 创建用户")
		log.Println("   GET  /health         - 健康检查")
		log.Println("按 Ctrl+C 优雅关闭服务器")

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP 服务器异常退出: %v", err)
		}
	}()

	// ============================================================
	// 优雅关闭（Graceful Shutdown）
	// ============================================================

	// 创建信号通道，监听 SIGINT（Ctrl+C）和 SIGTERM（kill）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	sig := <-quit
	log.Printf("📡 收到信号 %v，开始优雅关闭...", sig)

	// 创建超时 context，给予 5 秒处理剩余请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown 会：
	// 1. 停止接受新连接
	// 2. 等待已有请求处理完成
	// 3. 超时后强制关闭
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("⚠️ 服务器关闭异常: %v", err)
	}

	log.Println("✅ 服务器已优雅关闭")
}
