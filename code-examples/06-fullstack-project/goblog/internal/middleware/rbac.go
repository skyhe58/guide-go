package middleware

import (
	"guide-go/goblog/internal/auth"
	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/pkg/response"

	"github.com/gin-gonic/gin"
)

// RequireRole 返回 RBAC 角色权限中间件
// 从 Context 中读取用户角色，校验是否在允许的角色列表中
// 需要在 JWTAuth 中间件之后使用
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Context 获取用户角色
		userRole := GetRole(c)
		if userRole == "" {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		// 检查用户角色是否在允许列表中
		if !auth.HasRole(userRole, roles...) {
			response.Error(c, errcode.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}
