// zerolog + Gin 集成 — 请求日志/错误日志/链路 ID
// Go 1.22+ | 验证日期 2025-01-01
//
// 本示例演示 zerolog 与 Gin 框架的完整集成方案：
// - 结构化 JSON 请求日志中间件（method/path/status/latency/request_id）
// - 错误日志与堆栈追踪
// - 请求 ID（Request ID）链路追踪
// - 使用 httptest 验证完整流程
//
// 运行方式：go run ./zerolog-gin/
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ============================================================
// 中间件：请求 ID 生成
// 为每个请求生成唯一的 Request ID，用于链路追踪
// ============================================================

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从请求头获取（上游网关可能已生成）
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		// 存入 context 和响应头
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// ============================================================
// 中间件：zerolog 请求日志
// 记录每个 HTTP 请求的结构化日志（JSON 格式）
// ============================================================

func ZerologRequestLogger(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 请求完成后记录日志
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		requestID, _ := c.Get("request_id")

		// 根据状态码选择日志级别
		var event *zerolog.Event
		switch {
		case statusCode >= 500:
			event = logger.Error()
		case statusCode >= 400:
			event = logger.Warn()
		default:
			event = logger.Info()
		}

		event.
			Str("request_id", fmt.Sprintf("%v", requestID)).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", statusCode).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Int("body_size", c.Writer.Size()).
			Msg("HTTP 请求完成")
	}
}

// ============================================================
// 中间件：Panic Recovery（带 zerolog 错误日志）
// 捕获 panic 并记录错误日志，返回 500 响应
// ============================================================

func ZerologRecovery(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				requestID, _ := c.Get("request_id")
				logger.Error().
					Str("request_id", fmt.Sprintf("%v", requestID)).
					Interface("panic", r).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Msg("服务发生 panic，已恢复")

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":      "Internal Server Error",
					"request_id": requestID,
				})
			}
		}()
		c.Next()
	}
}

// ============================================================
// 业务 Handler
// ============================================================

// handleGetUser 获取用户信息
func handleGetUser(c *gin.Context) {
	userID := c.Param("id")
	requestID, _ := c.Get("request_id")

	// 使用带 request_id 的子 Logger 记录业务日志
	logger := log.With().
		Str("request_id", fmt.Sprintf("%v", requestID)).
		Str("handler", "GetUser").
		Logger()

	logger.Info().Str("user_id", userID).Msg("查询用户信息")

	// 模拟业务逻辑
	if userID == "0" {
		logger.Warn().Str("user_id", userID).Msg("用户不存在")
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":   userID,
		"name": "Alice",
	})
}

// handleCreateOrder 创建订单（演示错误日志）
func handleCreateOrder(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	logger := log.With().
		Str("request_id", fmt.Sprintf("%v", requestID)).
		Str("handler", "CreateOrder").
		Logger()

	logger.Info().Msg("开始创建订单")

	// 模拟业务错误
	err := fmt.Errorf("库存不足: 商品 SKU-001 剩余 0 件")
	logger.Error().
		Err(err).
		Str("sku", "SKU-001").
		Int("requested", 2).
		Int("available", 0).
		Msg("创建订单失败")

	c.JSON(http.StatusBadRequest, gin.H{
		"error":      err.Error(),
		"request_id": requestID,
	})
}

// handlePanic 演示 panic recovery
func handlePanic(c *gin.Context) {
	panic("模拟未预期的 panic 错误")
}

// ============================================================
// 路由配置
// ============================================================

func setupRouter(logger zerolog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 中间件链：RequestID → Recovery → Logger
	r.Use(RequestIDMiddleware())
	r.Use(ZerologRecovery(logger))
	r.Use(ZerologRequestLogger(logger))

	// 路由
	r.GET("/api/users/:id", handleGetUser)
	r.POST("/api/orders", handleCreateOrder)
	r.GET("/api/panic", handlePanic)

	return r
}

// ============================================================
// 主函数：使用 httptest 演示完整流程
// ============================================================

func main() {
	// 配置 zerolog：JSON 格式 + 时间戳 + 调用者信息
	zerolog.TimeFieldFormat = time.RFC3339Nano
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "user-api").
		Logger()

	// 设置全局 Logger
	log.Logger = logger

	fmt.Println("========== zerolog + Gin 集成演示 ==========")
	fmt.Println()

	router := setupRouter(logger)

	// 使用 httptest 模拟请求，演示日志输出
	fmt.Println("--- 1. 正常请求（200 OK）---")
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/users/42", nil)
	router.ServeHTTP(w, req)
	fmt.Printf("响应状态: %d, 响应体: %s\n\n", w.Code, w.Body.String())

	fmt.Println("--- 2. 用户不存在（404）---")
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/users/0", nil)
	router.ServeHTTP(w, req)
	fmt.Printf("响应状态: %d, 响应体: %s\n\n", w.Code, w.Body.String())

	fmt.Println("--- 3. 业务错误（400）---")
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/orders", nil)
	router.ServeHTTP(w, req)
	fmt.Printf("响应状态: %d, 响应体: %s\n\n", w.Code, w.Body.String())

	fmt.Println("--- 4. Panic Recovery（500）---")
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/panic", nil)
	router.ServeHTTP(w, req)
	fmt.Printf("响应状态: %d, 响应体: %s\n\n", w.Code, w.Body.String())

	fmt.Println("========== 演示完成 ==========")
	fmt.Println("所有日志均为结构化 JSON 格式，包含 request_id 用于链路追踪")
}
