// Package router 提供 GoBlog 的路由注册功能
// 按权限分组注册路由，挂载全局中间件和认证中间件
package router

import (
	"net/http"

	"guide-go/goblog/internal/auth"
	"guide-go/goblog/internal/config"
	"guide-go/goblog/internal/handler"
	"guide-go/goblog/internal/middleware"
	"guide-go/goblog/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

// Handlers 聚合所有模块的 Handler
type Handlers struct {
	User    *handler.UserHandler
	Article *handler.ArticleHandler
	Tag     *handler.TagHandler
	Comment *handler.CommentHandler
	Admin   *handler.AdminHandler
}

// SetupRouter 初始化 Gin 路由引擎
// 注册全局中间件、监控端点、API 路由分组
func SetupRouter(
	h *Handlers,
	jwtSvc *auth.JWTService,
	rdb *redis.Client,
	cfg *config.Config,
) *gin.Engine {
	// 根据配置设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()

	// ==================== 全局中间件 ====================
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.RateLimiter(&cfg.RateLimit))

	// ==================== 监控端点 ====================
	// Prometheus 指标端点
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 健康检查端点
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Response{
			Code:    0,
			Message: "ok",
			Data:    nil,
		})
	})

	// ==================== API v1 路由 ====================
	v1 := r.Group("/api/v1")
	{
		// ---------- 公开路由（无需认证） ----------
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", h.User.Register)
			authGroup.POST("/login", h.User.Login)
			authGroup.POST("/refresh", h.User.RefreshToken)
		}

		// 公开只读路由
		v1.GET("/articles", h.Article.List)
		v1.GET("/articles/search", h.Article.Search)
		v1.GET("/articles/:id", h.Article.GetByID)
		v1.GET("/tags", h.Tag.List)
		v1.GET("/tags/:id/articles", h.Tag.GetArticles)
		v1.GET("/articles/:id/comments", h.Comment.List)

		// ---------- 需要认证的路由 ----------
		authenticated := v1.Group("")
		authenticated.Use(middleware.JWTAuth(jwtSvc, rdb))
		{
			authenticated.POST("/auth/logout", h.User.Logout)
			authenticated.GET("/users/me", h.User.GetProfile)
			authenticated.PUT("/users/me", h.User.UpdateProfile)
			authenticated.POST("/articles/:id/comments", h.Comment.Create)
			authenticated.DELETE("/comments/:id", h.Comment.Delete)
		}

		// ---------- 需要 author 或 admin 角色的路由 ----------
		writer := v1.Group("")
		writer.Use(middleware.JWTAuth(jwtSvc, rdb), middleware.RequireRole("author", "admin"))
		{
			writer.POST("/articles", h.Article.Create)
			writer.PUT("/articles/:id", h.Article.Update)
			writer.DELETE("/articles/:id", h.Article.Delete)
			writer.POST("/tags", h.Tag.Create)
		}

		// ---------- 管理员路由 ----------
		admin := v1.Group("/admin")
		admin.Use(middleware.JWTAuth(jwtSvc, rdb), middleware.RequireRole("admin"))
		{
			admin.GET("/users", h.Admin.ListUsers)
			admin.PUT("/users/:id/role", h.Admin.UpdateRole)
			admin.PUT("/articles/:id/status", h.Admin.UpdateArticleStatus)
			admin.GET("/stats", h.Admin.GetStats)
		}
	}

	return r
}
