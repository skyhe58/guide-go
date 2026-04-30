// Package main 提供 GoBlog 数据库种子数据初始化脚本
// 用于创建管理员账号、示例文章和示例标签
//
// 使用方式：
//   go run scripts/seed.go
//
// 前置条件：
//   1. PostgreSQL 已启动（docker compose up -d postgres）
//   2. 数据库迁移已执行（make migrate-up）
package main

import (
	"fmt"
	"log"
	"time"

	"guide-go/goblog/internal/auth"
	"guide-go/goblog/internal/config"
	"guide-go/goblog/internal/database"
	"guide-go/goblog/internal/model"

	"gorm.io/gorm"
)

func main() {
	fmt.Println("==================== GoBlog 种子数据初始化 ====================")

	// 1. 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 连接数据库
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer func() {
		_ = database.CloseDB(db)
	}()
	fmt.Println("✅ 数据库连接成功")

	// 3. 自动迁移表结构（确保表存在）
	if err := db.AutoMigrate(
		&model.User{},
		&model.Article{},
		&model.Tag{},
		&model.Comment{},
	); err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}
	fmt.Println("✅ 表结构迁移完成")

	// 4. 创建种子数据
	seedUsers(db)
	seedTags(db)
	seedArticles(db)

	fmt.Println("==================== 种子数据初始化完成 ====================")
}

// seedUsers 创建管理员和示例用户
func seedUsers(db *gorm.DB) {
	fmt.Println("\n--- 创建用户 ---")

	users := []struct {
		Username string
		Email    string
		Password string
		Role     string
		Bio      string
	}{
		{
			Username: "admin",
			Email:    "admin@goblog.dev",
			Password: "admin123",
			Role:     "admin",
			Bio:      "GoBlog 系统管理员",
		},
		{
			Username: "zhangsan",
			Email:    "zhangsan@goblog.dev",
			Password: "author123",
			Role:     "author",
			Bio:      "Go 语言爱好者，专注后端开发",
		},
		{
			Username: "lisi",
			Email:    "lisi@goblog.dev",
			Password: "author123",
			Role:     "author",
			Bio:      "云原生工程师，Kubernetes 实践者",
		},
		{
			Username: "reader01",
			Email:    "reader01@goblog.dev",
			Password: "reader123",
			Role:     "reader",
			Bio:      "Go 学习者",
		},
	}

	for _, u := range users {
		// 检查用户是否已存在
		var count int64
		db.Model(&model.User{}).Where("username = ?", u.Username).Count(&count)
		if count > 0 {
			fmt.Printf("  ⏭️  用户 %s 已存在，跳过\n", u.Username)
			continue
		}

		// 加密密码
		hash, err := auth.HashPassword(u.Password)
		if err != nil {
			log.Printf("  ❌ 加密密码失败（%s）: %v", u.Username, err)
			continue
		}

		user := model.User{
			Username:     u.Username,
			Email:        u.Email,
			PasswordHash: hash,
			Role:         u.Role,
			Bio:          u.Bio,
		}

		if err := db.Create(&user).Error; err != nil {
			log.Printf("  ❌ 创建用户失败（%s）: %v", u.Username, err)
			continue
		}
		fmt.Printf("  ✅ 创建用户: %s（角色: %s，密码: %s）\n", u.Username, u.Role, u.Password)
	}
}

// seedTags 创建示例标签
func seedTags(db *gorm.DB) {
	fmt.Println("\n--- 创建标签 ---")

	tags := []struct {
		Name string
		Slug string
	}{
		{Name: "Go 基础", Slug: "go-basics"},
		{Name: "并发编程", Slug: "concurrency"},
		{Name: "Web 开发", Slug: "web-dev"},
		{Name: "数据库", Slug: "database"},
		{Name: "Redis", Slug: "redis"},
		{Name: "微服务", Slug: "microservice"},
		{Name: "Docker", Slug: "docker"},
		{Name: "Kubernetes", Slug: "kubernetes"},
		{Name: "设计模式", Slug: "design-patterns"},
		{Name: "面试", Slug: "interview"},
	}

	for _, t := range tags {
		var count int64
		db.Model(&model.Tag{}).Where("slug = ?", t.Slug).Count(&count)
		if count > 0 {
			fmt.Printf("  ⏭️  标签 %s 已存在，跳过\n", t.Name)
			continue
		}

		tag := model.Tag{
			Name: t.Name,
			Slug: t.Slug,
		}
		if err := db.Create(&tag).Error; err != nil {
			log.Printf("  ❌ 创建标签失败（%s）: %v", t.Name, err)
			continue
		}
		fmt.Printf("  ✅ 创建标签: %s\n", t.Name)
	}
}

// seedArticles 创建示例文章
func seedArticles(db *gorm.DB) {
	fmt.Println("\n--- 创建文章 ---")

	// 获取作者用户
	var zhangsan model.User
	if err := db.Where("username = ?", "zhangsan").First(&zhangsan).Error; err != nil {
		log.Printf("  ❌ 未找到作者 zhangsan: %v", err)
		return
	}

	var lisi model.User
	if err := db.Where("username = ?", "lisi").First(&lisi).Error; err != nil {
		log.Printf("  ❌ 未找到作者 lisi: %v", err)
		return
	}

	// 获取标签
	var goBasicsTag, concurrencyTag, webDevTag, dockerTag, redisTag model.Tag
	db.Where("slug = ?", "go-basics").First(&goBasicsTag)
	db.Where("slug = ?", "concurrency").First(&concurrencyTag)
	db.Where("slug = ?", "web-dev").First(&webDevTag)
	db.Where("slug = ?", "docker").First(&dockerTag)
	db.Where("slug = ?", "redis").First(&redisTag)

	now := time.Now()

	articles := []struct {
		AuthorID uint
		Title    string
		Slug     string
		Content  string
		Status   string
		Tags     []model.Tag
	}{
		{
			AuthorID: zhangsan.ID,
			Title:    "Go 语言 Slice 扩容机制详解",
			Slug:     "go-slice-grow-mechanism",
			Content: `# Go 语言 Slice 扩容机制详解

## 前言

Slice（切片）是 Go 语言中最常用的数据结构之一，理解其扩容机制对于编写高性能代码至关重要。

## 底层结构

Slice 在运行时由三个字段组成：
- **Data**：指向底层数组的指针
- **Len**：切片的长度（当前元素个数）
- **Cap**：切片的容量（底层数组的长度）

## 扩容规则

Go 1.18+ 的扩容策略：
1. 如果新容量大于旧容量的两倍，直接使用新容量
2. 如果旧容量小于 256，新容量翻倍
3. 否则按 newcap += (newcap + 3*threshold) / 4 增长

## 面试要点

这是 Go 面试中的高频题目，需要掌握扩容的具体规则和内存分配策略。`,
			Status: "published",
			Tags:   []model.Tag{goBasicsTag},
		},
		{
			AuthorID: zhangsan.ID,
			Title:    "Goroutine 与 Channel 并发模式实战",
			Slug:     "goroutine-channel-patterns",
			Content: `# Goroutine 与 Channel 并发模式实战

## 前言

Go 语言的并发模型基于 CSP（Communicating Sequential Processes），通过 goroutine 和 channel 实现高效的并发编程。

## 核心模式

### Fan-Out / Fan-In
将任务分发给多个 goroutine 并行处理，再将结果汇总。

### Pipeline
将处理流程分为多个阶段，每个阶段由独立的 goroutine 处理。

### Worker Pool
创建固定数量的 worker goroutine，从共享的任务队列中获取任务执行。

## 最佳实践

1. 始终确保 goroutine 能够正常退出
2. 使用 context 控制 goroutine 生命周期
3. 避免 goroutine 泄漏`,
			Status: "published",
			Tags:   []model.Tag{goBasicsTag, concurrencyTag},
		},
		{
			AuthorID: lisi.ID,
			Title:    "使用 Gin 框架构建 RESTful API",
			Slug:     "gin-restful-api-guide",
			Content: `# 使用 Gin 框架构建 RESTful API

## 前言

Gin 是 Go 语言中最流行的 Web 框架，以高性能和简洁的 API 著称。

## 核心特性

- 快速路由（基于 httprouter）
- 中间件支持
- JSON 验证
- 路由分组
- 错误管理

## 项目结构

推荐采用分层架构：Handler → Service → Repository，职责清晰，便于测试。

## 中间件链

Gin 的中间件机制是其核心设计之一，通过 Use() 方法注册全局或分组中间件。`,
			Status: "published",
			Tags:   []model.Tag{webDevTag, goBasicsTag},
		},
		{
			AuthorID: lisi.ID,
			Title:    "Docker 多阶段构建优化 Go 应用镜像",
			Slug:     "docker-multistage-build-go",
			Content: `# Docker 多阶段构建优化 Go 应用镜像

## 前言

Go 语言编译为静态链接的单一二进制文件，天然适合容器化部署。通过多阶段构建，可以将镜像大小控制在 20MB 以内。

## 多阶段构建

第一阶段使用 golang:alpine 编译，第二阶段使用 scratch 空镜像运行。

## 关键配置

- CGO_ENABLED=0：禁用 CGO，生成纯静态二进制
- -ldflags="-s -w"：去除调试信息，减小体积
- scratch 基础镜像：零依赖，最小攻击面

## 安全建议

1. 使用非 root 用户运行
2. 只复制必要的文件
3. 定期更新基础镜像`,
			Status: "published",
			Tags:   []model.Tag{dockerTag},
		},
		{
			AuthorID: zhangsan.ID,
			Title:    "Redis 缓存策略与常见问题解决方案",
			Slug:     "redis-cache-strategy",
			Content: `# Redis 缓存策略与常见问题解决方案

## 前言

Redis 作为高性能缓存中间件，在后端开发中扮演着重要角色。本文介绍常见的缓存策略和问题解决方案。

## 缓存模式

### Cache-Aside（旁路缓存）
读：先查缓存，未命中则查数据库并写入缓存。
写：先更新数据库，再删除缓存。

## 常见问题

### 缓存穿透
查询不存在的数据，请求直接打到数据库。
解决方案：布隆过滤器、空值缓存。

### 缓存击穿
热点 Key 过期，大量请求同时打到数据库。
解决方案：singleflight、互斥锁。

### 缓存雪崩
大量 Key 同时过期。
解决方案：随机过期时间、多级缓存。`,
			Status: "published",
			Tags:   []model.Tag{redisTag},
		},
	}

	for _, a := range articles {
		var count int64
		db.Model(&model.Article{}).Where("slug = ?", a.Slug).Count(&count)
		if count > 0 {
			fmt.Printf("  ⏭️  文章「%s」已存在，跳过\n", a.Title)
			continue
		}

		article := model.Article{
			AuthorID:    a.AuthorID,
			Title:       a.Title,
			Slug:        a.Slug,
			Content:     a.Content,
			Status:      a.Status,
			PublishedAt: &now,
			Tags:        a.Tags,
		}

		if err := db.Create(&article).Error; err != nil {
			log.Printf("  ❌ 创建文章失败（%s）: %v", a.Title, err)
			continue
		}
		fmt.Printf("  ✅ 创建文章: %s\n", a.Title)
	}
}
