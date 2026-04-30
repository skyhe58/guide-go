package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Logger 返回 zerolog 请求日志中间件
// 记录每个请求的 method、path、status、latency、request_id
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算请求耗时
		latency := time.Since(start)

		// 获取请求 ID
		requestID, _ := c.Get(RequestIDKey)

		// 使用 zerolog 记录请求日志
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Interface("request_id", requestID).
			Msg("HTTP 请求")
	}
}
