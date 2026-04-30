package middleware

import (
	"net/http"
	"runtime/debug"

	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Recovery 返回 Panic 恢复中间件
// 捕获 Handler 中的 panic，记录堆栈信息，返回 500 错误响应
// 防止单个请求的 panic 导致整个服务崩溃
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := debug.Stack()

				// 获取请求 ID
				requestID, _ := c.Get(RequestIDKey)

				// 记录 panic 日志（包含堆栈）
				log.Error().
					Interface("error", err).
					Str("stack", string(stack)).
					Interface("request_id", requestID).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Msg("捕获到 Panic")

				// 返回 500 错误响应
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Code:    errcode.ErrInternal.Code,
					Message: errcode.ErrInternal.Message,
					Data:    nil,
				})
			}
		}()

		c.Next()
	}
}
