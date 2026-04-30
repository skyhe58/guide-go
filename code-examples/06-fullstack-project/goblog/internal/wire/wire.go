// Package wire 提供 GoBlog 的依赖注入配置
// 使用 Google Wire 风格定义 Provider Sets
// 管理 Handler → Service → Repository 三层依赖关系
package wire

import (
	"guide-go/goblog/internal/auth"
	"guide-go/goblog/internal/config"
	"guide-go/goblog/internal/handler"
	"guide-go/goblog/internal/repository"
	"guide-go/goblog/internal/router"
	"guide-go/goblog/internal/service"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// App 应用依赖聚合，包含路由所需的所有组件
type App struct {
	Handlers *router.Handlers
	JWTSvc   *auth.JWTService
	Config   *config.Config
}

// InitializeApp 手动依赖注入（模拟 Wire 生成代码）
// 按照依赖顺序创建各层实例：
// Config → Infrastructure → Repository → Service → Handler
func InitializeApp(cfg *config.Config, db *gorm.DB, rdb *redis.Client) *App {
	// ==================== JWT 服务 ====================
	jwtSvc := auth.NewJWTService(&cfg.JWT)

	// ==================== Repository 层 ====================
	userRepo := repository.NewUserRepo(db)
	articleRepo := repository.NewArticleRepo(db)
	tagRepo := repository.NewTagRepo(db)
	commentRepo := repository.NewCommentRepo(db)

	// ==================== Service 层 ====================
	userSvc := service.NewUserService(userRepo, jwtSvc, rdb)
	articleSvc := service.NewArticleService(articleRepo, tagRepo, rdb)
	tagSvc := service.NewTagService(tagRepo)
	commentSvc := service.NewCommentService(commentRepo, articleRepo)
	adminSvc := service.NewAdminService(userRepo, articleRepo, commentRepo, tagRepo)

	// ==================== Handler 层 ====================
	userHandler := handler.NewUserHandler(userSvc)
	articleHandler := handler.NewArticleHandler(articleSvc)
	tagHandler := handler.NewTagHandler(tagSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)
	adminHandler := handler.NewAdminHandler(adminSvc)

	// ==================== 聚合 Handlers ====================
	handlers := &router.Handlers{
		User:    userHandler,
		Article: articleHandler,
		Tag:     tagHandler,
		Comment: commentHandler,
		Admin:   adminHandler,
	}

	return &App{
		Handlers: handlers,
		JWTSvc:   jwtSvc,
		Config:   cfg,
	}
}
