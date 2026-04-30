package middleware

import (
	"context"
	"strings"

	"guide-go/goblog/internal/auth"
	"guide-go/goblog/internal/cache"
	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Context 键名常量，用于在 Gin Context 中存储用户信息
const (
	ContextUserIDKey   = "user_id"
	ContextUsernameKey = "username"
	ContextRoleKey     = "role"
	ContextJTIKey      = "jti"
)

// JWTAuth 返回 JWT 认证中间件
// 从 Authorization Header 解析 Bearer Token，验证有效性
// 将用户信息（user_id、username、role）写入 Gin Context
// 同时检查 Token 是否在 Redis 黑名单中（已登出的 Token）
func JWTAuth(jwtSvc *auth.JWTService, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization Header 获取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		// 解析 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		// 验证 Token
		claims, err := jwtSvc.ParseToken(parts[1])
		if err != nil {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		// 验证是否为 Access Token
		if claims.Type != auth.AccessToken {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		// 检查 Token 是否在黑名单中（已登出）
		blacklistKey := cache.TokenBlacklistKey(claims.ID)
		exists, err := rdb.Exists(context.Background(), blacklistKey).Result()
		if err == nil && exists > 0 {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		// 将用户信息写入 Context
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUsernameKey, claims.Username)
		c.Set(ContextRoleKey, claims.Role)
		c.Set(ContextJTIKey, claims.ID)

		c.Next()
	}
}

// GetUserID 从 Gin Context 中获取当前用户 ID
func GetUserID(c *gin.Context) uint {
	id, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0
	}
	return id.(uint)
}

// GetUsername 从 Gin Context 中获取当前用户名
func GetUsername(c *gin.Context) string {
	username, exists := c.Get(ContextUsernameKey)
	if !exists {
		return ""
	}
	return username.(string)
}

// GetRole 从 Gin Context 中获取当前用户角色
func GetRole(c *gin.Context) string {
	role, exists := c.Get(ContextRoleKey)
	if !exists {
		return ""
	}
	return role.(string)
}
