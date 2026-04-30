// Package cache 提供 GoBlog 的 Redis 缓存客户端初始化和管理功能
// 使用 go-redis v9 客户端，支持连接池配置
package cache

import (
	"context"
	"fmt"

	"guide-go/goblog/internal/config"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient 初始化 Redis 客户端
// 根据配置创建 go-redis 客户端实例，设置连接池参数
func NewRedisClient(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// 测试连接是否正常
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %w", err)
	}

	return client, nil
}

// CloseRedis 关闭 Redis 客户端连接
func CloseRedis(client *redis.Client) error {
	if client == nil {
		return nil
	}
	return client.Close()
}
