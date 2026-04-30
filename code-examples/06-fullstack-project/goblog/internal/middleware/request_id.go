package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 请求 ID 在 Context 中的键名
const RequestIDKey = "request_id"

// RequestID 返回请求 ID 中间件
// 为每个请求生成唯一的 UUID，写入 X-Request-ID Header 和 Gin Context
// 用于请求链路追踪和日志关联
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先使用客户端传入的 X-Request-ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 写入响应 Header 和 Context
		c.Header("X-Request-ID", requestID)
		c.Set(RequestIDKey, requestID)

		c.Next()
	}
}
