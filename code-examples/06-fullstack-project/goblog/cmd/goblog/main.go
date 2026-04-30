// Package main 是 GoBlog 博客平台的程序入口
// 负责加载配置、初始化依赖、启动 HTTP 服务器并实现优雅启停
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"guide-go/goblog/internal/cache"
	"guide-go/goblog/internal/config"
	"guide-go/goblog/internal/database"
	"guide-go/goblog/internal/router"
	"guide-go/goblog/internal/wire"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Prometheus 指标定义
var (
	// httpRequestsTotal HTTP 请求总数计数器
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP 请求总数",
		},
		[]string{"method", "path", "status"},
	)

	// httpRequestDuration HTTP 请求延迟直方图
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP 请求延迟（秒）",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// httpActiveConnections 活跃连接数
	httpActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_connections",
			Help: "当前活跃的 HTTP 连接数",
		},
	)
)

func init() {
	// 注册 Prometheus 指标
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(httpActiveConnections)
}

func main() {
	// ==================== 1. 初始化日志 ====================
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	log.Info().Msg("GoBlog 博客平台启动中...")

	// ==================== 2. 加载配置 ====================
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatal().Err(err).Msg("加载配置失败")
	}
	log.Info().Int("port", cfg.Server.Port).Str("mode", cfg.Server.Mode).Msg("配置加载完成")

	// ==================== 3. 初始化数据库 ====================
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("初始化数据库失败")
	}
	defer func() {
		if err := database.CloseDB(db); err != nil {
			log.Error().Err(err).Msg("关闭数据库连接失败")
		}
		log.Info().Msg("数据库连接已关闭")
	}()
	log.Info().Msg("数据库连接成功")

	// ==================== 4. 初始化 Redis ====================
	rdb, err := cache.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("初始化 Redis 失败")
	}
	defer func() {
		if err := cache.CloseRedis(rdb); err != nil {
			log.Error().Err(err).Msg("关闭 Redis 连接失败")
		}
		log.Info().Msg("Redis 连接已关闭")
	}()
	log.Info().Msg("Redis 连接成功")

	// ==================== 5. Wire 依赖注入 ====================
	app := wire.InitializeApp(cfg, db, rdb)
	log.Info().Msg("依赖注入完成")

	// ==================== 6. 注册路由 ====================
	r := router.SetupRouter(app.Handlers, app.JWTSvc, rdb, cfg)
	log.Info().Msg("路由注册完成")

	// ==================== 7. 启动 HTTP 服务器 ====================
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 在 goroutine 中启动服务器
	go func() {
		log.Info().Str("addr", srv.Addr).Msg("HTTP 服务器启动")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP 服务器异常退出")
		}
	}()

	// ==================== 8. 优雅启停 ====================
	// 监听系统信号（SIGINT: Ctrl+C, SIGTERM: kill）
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 阻塞等待信号
	<-ctx.Done()
	log.Info().Msg("收到关闭信号，开始优雅关闭...")

	// 创建关闭超时上下文
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// 关闭 HTTP 服务器（等待进行中的请求完成）
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("HTTP 服务器关闭失败")
	}

	// 等待一小段时间确保所有资源释放
	time.Sleep(100 * time.Millisecond)

	log.Info().Msg("GoBlog 服务已安全关闭")
}
