// 中间件模式 — Go Web 开发中最核心的设计模式
// 演示 HTTP 中间件链的实现原理（日志、认证、限流、恢复）
//
// Go 版本要求: Go 1.22+
// 最后验证日期: 2025-01-01
//
// 运行方式: go run ./middleware/

package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// =============================================================================
// Part A: 中间件模式完整演示（纯内存模拟，不启动真实 HTTP 服务器）
// =============================================================================

// Middleware 定义中间件类型：接收一个 Handler，返回一个新的 Handler
type Middleware func(http.Handler) http.Handler

// Chain 将多个中间件链式组合
// 执行顺序：第一个中间件最先执行（最外层）
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	// 从后往前包装，确保第一个中间件在最外层
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// =============================================================================
// 中间件实现
// =============================================================================

// LoggingMiddleware 日志中间件 — 记录请求方法、路径和耗时
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fmt.Printf("[LOG] --> %s %s\n", r.Method, r.URL.Path)

		// 调用下一个处理器
		next.ServeHTTP(w, r)

		fmt.Printf("[LOG] <-- %s %s (%v)\n", r.Method, r.URL.Path, time.Since(start))
	})
}

// AuthMiddleware 认证中间件 — 检查 Authorization 头
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			fmt.Println("[AUTH] 拒绝：缺少 Authorization 头")
			http.Error(w, "未授权", http.StatusUnauthorized)
			return
		}

		// 简单的 token 校验（实际项目中应验证 JWT）
		if !strings.HasPrefix(token, "Bearer ") {
			fmt.Println("[AUTH] 拒绝：无效的 Token 格式")
			http.Error(w, "无效的认证信息", http.StatusUnauthorized)
			return
		}

		fmt.Printf("[AUTH] 通过：Token = %s\n", token)
		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware 恢复中间件 — 捕获 panic，防止服务崩溃
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("[RECOVERY] 捕获 panic: %v\n", err)
				http.Error(w, "服务器内部错误", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware 限流中间件 — 简单的请求计数限流
func RateLimitMiddleware(maxRequests int) Middleware {
	requestCount := 0
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			if requestCount > maxRequests {
				fmt.Printf("[RATE] 拒绝：请求数 %d 超过限制 %d\n", requestCount, maxRequests)
				http.Error(w, "请求过于频繁", http.StatusTooManyRequests)
				return
			}
			fmt.Printf("[RATE] 通过：当前请求数 %d/%d\n", requestCount, maxRequests)
			next.ServeHTTP(w, r)
		})
	}
}

// =============================================================================
// 业务处理器
// =============================================================================

// helloHandler 业务处理器
func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[HANDLER] 处理业务逻辑")
	fmt.Fprintf(w, "Hello, Go 中间件模式!")
}

// =============================================================================
// 模拟 HTTP 请求（不启动真实服务器）
// =============================================================================

// mockRequest 模拟一个 HTTP 请求
type mockResponseWriter struct {
	statusCode int
	body       strings.Builder
}

func (m *mockResponseWriter) Header() http.Header        { return http.Header{} }
func (m *mockResponseWriter) WriteHeader(statusCode int)  { m.statusCode = statusCode }
func (m *mockResponseWriter) Write(b []byte) (int, error) { return m.body.Write(b) }

func simulateRequest(handler http.Handler, method, path, authToken string) {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		log.Fatal(err)
	}
	if authToken != "" {
		req.Header.Set("Authorization", authToken)
	}

	w := &mockResponseWriter{statusCode: 200}
	handler.ServeHTTP(w, req)

	fmt.Printf("[RESULT] 状态码: %d, 响应: %s\n", w.statusCode, w.body.String())
}

func main() {
	fmt.Println("=== Go 中间件模式演示 ===")
	fmt.Println()

	// 构建中间件链
	// 执行顺序: Recovery → Logging → RateLimit → Auth → Handler
	handler := Chain(
		http.HandlerFunc(helloHandler),
		RecoveryMiddleware,          // 最外层：捕获 panic
		LoggingMiddleware,           // 第二层：记录日志
		RateLimitMiddleware(3),      // 第三层：限流
		AuthMiddleware,              // 第四层：认证
	)

	// 模拟请求 1：正常请求（带 Token）
	fmt.Println("--- 请求 1: 正常请求 ---")
	simulateRequest(handler, "GET", "/api/hello", "Bearer my-jwt-token")
	fmt.Println()

	// 模拟请求 2：未认证请求（无 Token）
	fmt.Println("--- 请求 2: 未认证请求 ---")
	simulateRequest(handler, "GET", "/api/hello", "")
	fmt.Println()

	// 模拟请求 3：正常请求
	fmt.Println("--- 请求 3: 正常请求 ---")
	simulateRequest(handler, "POST", "/api/data", "Bearer another-token")
	fmt.Println()

	// 模拟请求 4：触发限流
	fmt.Println("--- 请求 4: 触发限流（第 4 次请求，限制 3 次）---")
	simulateRequest(handler, "GET", "/api/hello", "Bearer my-token")
	fmt.Println()

	fmt.Println("=== 中间件模式要点 ===")
	fmt.Println("1. 中间件签名: func(http.Handler) http.Handler")
	fmt.Println("2. 通过闭包捕获 next handler")
	fmt.Println("3. 在 next.ServeHTTP() 前后添加逻辑（洋葱模型）")
	fmt.Println("4. 实际应用: Gin c.Next()/c.Abort(), K8s 准入控制链")
}
