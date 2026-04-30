// Package middleware 提供 GoBlog 的 Gin 中间件
// 包含 CORS、请求 ID、日志、Panic 恢复、限流、JWT 认证、RBAC 权限等中间件
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS 返回跨域资源共享中间件
// 允许前端跨域请求，配置允许的 Origin、Method、Header
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		// 预检请求直接返回 204
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
