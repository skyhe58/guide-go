package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"guide-go/goblog/internal/cache"
	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/internal/model"
	"guide-go/goblog/internal/repository"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

// 文章缓存相关常量
const (
	articleCacheTTL     = 30 * time.Minute // 文章缓存有效期
	articleNullCacheTTL = 5 * time.Minute  // 空值缓存有效期（防穿透）
)

// ArticleService 文章业务服务接口
type ArticleService interface {
	Create(authorID uint, title, content, slug, status string, tagIDs []uint) (*model.Article, *errcode.AppError)
	Update(id, operatorID uint, operatorRole, title, content, slug, status string, tagIDs []uint) (*model.Article, *errcode.AppError)
	Delete(id, operatorID uint, operatorRole string) *errcode.AppError
	GetByID(id uint) (*model.Article, *errcode.AppError)
	List(offset, limit int, status string, tagID uint) ([]model.Article, int64, *errcode.AppError)
	Search(keyword string, offset, limit int) ([]model.Article, int64, *errcode.AppError)
	GetHotArticles(limit int) ([]model.Article, *errcode.AppError)
}

// articleService 文章业务服务实现
type articleService struct {
	articleRepo repository.ArticleRepo
	tagRepo     repository.TagRepo
	rdb         *redis.Client
	sfGroup     singleflight.Group // singleflight 防止缓存击穿
}

// NewArticleService 创建文章服务实例
func NewArticleService(articleRepo repository.ArticleRepo, tagRepo repository.TagRepo, rdb *redis.Client) ArticleService {
	return &articleService{
		articleRepo: articleRepo,
		tagRepo:     tagRepo,
		rdb:         rdb,
	}
}

// Create 创建文章
func (s *articleService) Create(authorID uint, title, content, slug, status string, tagIDs []uint) (*model.Article, *errcode.AppError) {
	// 构建标签关联
	var tags []model.Tag
	for _, tagID := range tagIDs {
		tags = append(tags, model.Tag{ID: tagID})
	}

	article := &model.Article{
		AuthorID: authorID,
		Title:    title,
		Content:  content,
		Slug:     slug,
		Status:   status,
		Tags:     tags,
	}

	// 如果状态为已发布，设置发布时间
	if status == "published" {
		now := time.Now()
		article.PublishedAt = &now
	}

	if err := s.articleRepo.Create(article); err != nil {
		return nil, errcode.ErrInternal
	}

	return article, nil
}

// Update 更新文章
// 只有文章作者或管理员可以更新
func (s *articleService) Update(id, operatorID uint, operatorRole, title, content, slug, status string, tagIDs []uint) (*model.Article, *errcode.AppError) {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrArticleNotFound
		}
		return nil, errcode.ErrInternal
	}

	// 权限检查：只有作者本人或管理员可以编辑
	if article.AuthorID != operatorID && operatorRole != "admin" {
		return nil, errcode.ErrArticleNoPermission
	}

	// 更新字段
	if title != "" {
		article.Title = title
	}
	if content != "" {
		article.Content = content
	}
	if slug != "" {
		article.Slug = slug
	}
	if status != "" {
		article.Status = status
		if status == "published" && article.PublishedAt == nil {
			now := time.Now()
			article.PublishedAt = &now
		}
	}
	if tagIDs != nil {
		var tags []model.Tag
		for _, tagID := range tagIDs {
			tags = append(tags, model.Tag{ID: tagID})
		}
		article.Tags = tags
	}

	if err := s.articleRepo.Update(article); err != nil {
		return nil, errcode.ErrInternal
	}

	// 删除缓存
	s.deleteArticleCache(id)

	return article, nil
}

// Delete 软删除文章
// 只有文章作者或管理员可以删除
func (s *articleService) Delete(id, operatorID uint, operatorRole string) *errcode.AppError {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errcode.ErrArticleNotFound
		}
		return errcode.ErrInternal
	}

	// 权限检查
	if article.AuthorID != operatorID && operatorRole != "admin" {
		return errcode.ErrArticleNoPermission
	}

	if err := s.articleRepo.Delete(id); err != nil {
		return errcode.ErrInternal
	}

	// 删除缓存
	s.deleteArticleCache(id)

	return nil
}

// GetByID 获取文章详情
// 使用 Redis 缓存（Cache-Aside 模式）+ 空值缓存防穿透 + singleflight 防击穿
func (s *articleService) GetByID(id uint) (*model.Article, *errcode.AppError) {
	ctx := context.Background()
	cacheKey := cache.ArticleKey(id)

	// 1. 先查缓存
	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		// 缓存命中
		var article model.Article
		if err := json.Unmarshal([]byte(cached), &article); err == nil {
			return &article, nil
		}
	}

	// 检查空值缓存（防穿透）
	nullKey := cache.ArticleNullKey(id)
	if exists, _ := s.rdb.Exists(ctx, nullKey).Result(); exists > 0 {
		return nil, errcode.ErrArticleNotFound
	}

	// 2. 使用 singleflight 防止缓存击穿
	sfKey := fmt.Sprintf("article:%d", id)
	result, sfErr, _ := s.sfGroup.Do(sfKey, func() (interface{}, error) {
		// 查询数据库
		article, err := s.articleRepo.GetByID(id)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 设置空值缓存，防止缓存穿透
				s.rdb.Set(ctx, nullKey, "1", articleNullCacheTTL)
				return nil, err
			}
			return nil, err
		}

		// 写入缓存
		data, _ := json.Marshal(article)
		s.rdb.Set(ctx, cacheKey, string(data), articleCacheTTL)

		// 增加浏览次数
		_ = s.articleRepo.IncrViewCount(id)

		// 更新热门文章排行榜（Redis Sorted Set）
		s.rdb.ZIncrBy(ctx, cache.KeyHotArticles, 1, fmt.Sprintf("%d", id))

		return article, nil
	})

	if sfErr != nil {
		if sfErr == gorm.ErrRecordNotFound {
			return nil, errcode.ErrArticleNotFound
		}
		return nil, errcode.ErrInternal
	}

	return result.(*model.Article), nil
}

// List 分页查询文章列表
func (s *articleService) List(offset, limit int, status string, tagID uint) ([]model.Article, int64, *errcode.AppError) {
	var scopes []repository.ArticleScope
	if status != "" {
		scopes = append(scopes, repository.WithStatus(status))
	}
	if tagID > 0 {
		scopes = append(scopes, repository.WithTagID(tagID))
	}

	articles, total, err := s.articleRepo.List(offset, limit, scopes...)
	if err != nil {
		return nil, 0, errcode.ErrInternal
	}

	return articles, total, nil
}

// Search 搜索文章（按标题和内容模糊匹配）
func (s *articleService) Search(keyword string, offset, limit int) ([]model.Article, int64, *errcode.AppError) {
	articles, total, err := s.articleRepo.Search(keyword, offset, limit)
	if err != nil {
		return nil, 0, errcode.ErrInternal
	}
	return articles, total, nil
}

// GetHotArticles 获取热门文章排行榜（基于 Redis Sorted Set）
func (s *articleService) GetHotArticles(limit int) ([]model.Article, *errcode.AppError) {
	ctx := context.Background()

	// 从 Redis Sorted Set 获取热门文章 ID
	results, err := s.rdb.ZRevRangeWithScores(ctx, cache.KeyHotArticles, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, errcode.ErrInternal
	}

	var articles []model.Article
	for _, z := range results {
		var id uint
		if _, err := fmt.Sscanf(z.Member.(string), "%d", &id); err != nil {
			continue
		}
		article, dbErr := s.articleRepo.GetByID(id)
		if dbErr == nil {
			articles = append(articles, *article)
		}
	}

	return articles, nil
}

// deleteArticleCache 删除文章缓存
func (s *articleService) deleteArticleCache(id uint) {
	ctx := context.Background()
	s.rdb.Del(ctx, cache.ArticleKey(id))
	s.rdb.Del(ctx, cache.ArticleNullKey(id))
}
