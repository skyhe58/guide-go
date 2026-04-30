// Sentry + Gin 集成 — 错误上报/Panic Recovery/Breadcrumbs/上下文
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 Sentry Go SDK 与 Gin 框架的集成：
// - Sentry SDK 初始化（使用空 DSN 演示流程）
// - Gin 中间件实现 Panic Recovery 和错误上报
// - Breadcrumbs（面包屑）记录操作轨迹
// - Scope 上下文信息附加
// - 使用 httptest 验证完整流程
//
// 注意：本示例使用空 DSN，事件不会真正发送到 Sentry 服务器。
// 生产环境请替换为真实的 Sentry DSN。
//
// 运行方式：go run ./sentry-gin/
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

// ============================================================
// Sentry 初始化
// ============================================================

// initSentry 初始化 Sentry SDK
// 生产环境请通过环境变量注入 DSN：os.Getenv("SENTRY_DSN")
func initSentry() error {
	err := sentry.Init(sentry.ClientOptions{
		// DSN 为空时 SDK 正常运行但不发送事件，适合本地演示
		Dsn: "",
		// 环境标识
		Environment: "development",
		// Release 版本追踪：关联错误与代码版本
		Release: "user-api@1.0.0",
		// 性能监控采样率（0.0-1.0）
		TracesSampleRate: 1.0,
		// 发送前回调：可用于过滤或修改事件
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			fmt.Printf("[Sentry 事件捕获] Level=%s, Message=%s\n", event.Level, event.Message)
			if event.Exception != nil {
				for _, ex := range event.Exception {
					fmt.Printf("  Exception: Type=%s, Value=%s\n", ex.Type, ex.Value)
				}
			}
			return event
		},
	})
	return err
}

// ============================================================
// Sentry Gin 中间件
// 实现 Panic Recovery + 错误上报 + 请求上下文
// ============================================================

// SentryMiddleware Sentry 集成中间件
// 功能：设置请求上下文、捕获 panic、上报错误
func SentryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 为每个请求创建独立的 Sentry Hub（隔离 Scope）
		hub := sentry.CurrentHub().Clone()
		c.Set("sentry_hub", hub)

		// 设置请求上下文
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetRequest(c.Request)
			scope.SetTag("gin.route", c.FullPath())
			scope.SetTag("http.method", c.Request.Method)
		})

		// Panic Recovery
		defer func() {
			if r := recover(); r != nil {
				// 记录 panic 到 Sentry
				hub.RecoverWithContext(c.Request.Context(), r)
				// 确保事件发送完成
				hub.Flush(2 * time.Second)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
				})
			}
		}()

		c.Next()
	}
}

// getSentryHub 从 Gin context 获取 Sentry Hub
func getSentryHub(c *gin.Context) *sentry.Hub {
	if hub, exists := c.Get("sentry_hub"); exists {
		return hub.(*sentry.Hub)
	}
	return sentry.CurrentHub()
}

// ============================================================
// 业务 Handler
// ============================================================

// handleGetUser 获取用户信息（正常请求）
func handleGetUser(c *gin.Context) {
	hub := getSentryHub(c)
	userID := c.Param("id")

	// 添加 Breadcrumb：记录操作轨迹
	hub.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "handler",
		Message:  fmt.Sprintf("查询用户: %s", userID),
		Level:    sentry.LevelInfo,
		Data: map[string]interface{}{
			"user_id": userID,
		},
	}, nil)

	// 设置用户上下文
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:       userID,
			Username: "alice",
		})
	})

	c.JSON(http.StatusOK, gin.H{
		"id":   userID,
		"name": "Alice",
	})
}

// handleCreateOrder 创建订单（演示错误上报）
func handleCreateOrder(c *gin.Context) {
	hub := getSentryHub(c)

	// Breadcrumb：记录操作步骤
	hub.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "order",
		Message:  "开始创建订单",
		Level:    sentry.LevelInfo,
	}, nil)

	hub.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "order",
		Message:  "验证库存",
		Level:    sentry.LevelInfo,
		Data: map[string]interface{}{
			"sku":       "SKU-001",
			"requested": 2,
		},
	}, nil)

	// 模拟业务错误
	err := fmt.Errorf("库存不足: 商品 SKU-001 剩余 0 件")

	// 使用 Scope 附加上下文后上报错误
	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("module", "order")
		scope.SetExtra("sku", "SKU-001")
		scope.SetExtra("requested_quantity", 2)
		scope.SetExtra("available_quantity", 0)
		scope.SetLevel(sentry.LevelError)
		hub.CaptureException(err)
	})

	c.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
}

// handlePayment 支付接口（演示 Panic Recovery）
func handlePayment(c *gin.Context) {
	hub := getSentryHub(c)

	hub.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "payment",
		Message:  "开始处理支付",
		Level:    sentry.LevelInfo,
	}, nil)

	hub.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "payment",
		Message:  "调用支付网关",
		Level:    sentry.LevelInfo,
		Data: map[string]interface{}{
			"gateway": "stripe",
			"amount":  9900,
		},
	}, nil)

	// 模拟 panic（如空指针、数组越界等未预期错误）
	panic("支付网关返回异常: nil pointer dereference")
}

// handleHealth 健康检查
func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ============================================================
// 路由配置
// ============================================================

func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Sentry 中间件
	r.Use(SentryMiddleware())

	// 路由
	r.GET("/health", handleHealth)
	r.GET("/api/users/:id", handleGetUser)
	r.POST("/api/orders", handleCreateOrder)
	r.POST("/api/payment", handlePayment)

	return r
}

// ============================================================
// 主函数：使用 httptest 演示完整流程
// ============================================================

func main() {
	fmt.Println("========== Sentry + Gin 集成演示 ==========")
	fmt.Println("注意：使用空 DSN，事件通过 BeforeSend 回调打印到控制台")
	fmt.Println()

	// 初始化 Sentry
	if err := initSentry(); err != nil {
		fmt.Printf("Sentry 初始化失败: %v\n", err)
		return
	}
	defer sentry.Flush(2 * time.Second)

	router := setupRouter()

	// 1. 正常请求
	fmt.Println("--- 1. 正常请求（200 OK）---")
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/users/42", nil)
	router.ServeHTTP(w, req)
	fmt.Printf("响应状态: %d, 响应体: %s\n\n", w.Code, w.Body.String())

	// 2. 业务错误（触发 Sentry 错误上报）
	fmt.Println("--- 2. 业务错误 + Sentry 上报（400）---")
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/orders", nil)
	router.ServeHTTP(w, req)
	fmt.Printf("响应状态: %d, 响应体: %s\n\n", w.Code, w.Body.String())

	// 3. Panic Recovery（触发 Sentry panic 上报）
	fmt.Println("--- 3. Panic Recovery + Sentry 上报（500）---")
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/payment", nil)
	router.ServeHTTP(w, req)
	fmt.Printf("响应状态: %d, 响应体: %s\n\n", w.Code, w.Body.String())

	fmt.Println("========== 演示完成 ==========")
	fmt.Println("生产环境请设置真实的 Sentry DSN（通过环境变量 SENTRY_DSN）")
	fmt.Println("Sentry 会自动聚合相同错误、去重、并发送告警通知")
}
