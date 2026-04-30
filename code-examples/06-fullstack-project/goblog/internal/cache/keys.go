// Package cache 定义 GoBlog 的缓存键生成规则
// 所有 Redis 缓存键统一在此管理，避免键名冲突和硬编码
package cache

import "fmt"

// 缓存键前缀常量
const (
	// PrefixArticle 文章详情缓存键前缀
	PrefixArticle = "article:"

	// PrefixArticleNull 文章空值缓存键前缀（防止缓存穿透）
	PrefixArticleNull = "article:%d:null"

	// KeyHotArticles 热门文章排行榜缓存键（Sorted Set）
	KeyHotArticles = "articles:hot"

	// PrefixTokenBlacklist Token 黑名单缓存键前缀
	PrefixTokenBlacklist = "token:blacklist:"
)

// ArticleKey 生成文章详情缓存键
// 格式：article:{id}
func ArticleKey(id uint) string {
	return fmt.Sprintf("%s%d", PrefixArticle, id)
}

// ArticleNullKey 生成文章空值缓存键（防止缓存穿透）
// 格式：article:{id}:null
func ArticleNullKey(id uint) string {
	return fmt.Sprintf(PrefixArticleNull, id)
}

// TokenBlacklistKey 生成 Token 黑名单缓存键
// 格式：token:blacklist:{jti}
func TokenBlacklistKey(jti string) string {
	return PrefixTokenBlacklist + jti
}
