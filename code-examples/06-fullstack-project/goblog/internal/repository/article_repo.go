package repository

import (
	"guide-go/goblog/internal/model"

	"gorm.io/gorm"
)

// ArticleScope 文章查询条件函数类型（Scope 模式）
// 用于灵活组合查询条件，如按标签筛选、按状态过滤等
type ArticleScope func(db *gorm.DB) *gorm.DB

// WithStatus 按文章状态筛选
func WithStatus(status string) ArticleScope {
	return func(db *gorm.DB) *gorm.DB {
		if status != "" {
			return db.Where("status = ?", status)
		}
		return db
	}
}

// WithAuthorID 按作者 ID 筛选
func WithAuthorID(authorID uint) ArticleScope {
	return func(db *gorm.DB) *gorm.DB {
		if authorID > 0 {
			return db.Where("author_id = ?", authorID)
		}
		return db
	}
}

// WithTagID 按标签 ID 筛选（通过关联表）
func WithTagID(tagID uint) ArticleScope {
	return func(db *gorm.DB) *gorm.DB {
		if tagID > 0 {
			return db.Where("id IN (SELECT article_id FROM article_tags WHERE tag_id = ?)", tagID)
		}
		return db
	}
}

// ArticleRepo 文章数据访问接口
type ArticleRepo interface {
	Create(article *model.Article) error
	GetByID(id uint) (*model.Article, error)
	GetBySlug(slug string) (*model.Article, error)
	Update(article *model.Article) error
	Delete(id uint) error
	List(offset, limit int, scopes ...ArticleScope) ([]model.Article, int64, error)
	Search(keyword string, offset, limit int) ([]model.Article, int64, error)
	IncrViewCount(id uint) error
}

// articleRepo 文章数据访问实现
type articleRepo struct {
	db *gorm.DB
}

// NewArticleRepo 创建文章 Repository 实例
func NewArticleRepo(db *gorm.DB) ArticleRepo {
	return &articleRepo{db: db}
}

// Create 创建文章
func (r *articleRepo) Create(article *model.Article) error {
	return r.db.Create(article).Error
}

// GetByID 根据 ID 查询文章（预加载作者和标签）
func (r *articleRepo) GetByID(id uint) (*model.Article, error) {
	var article model.Article
	if err := r.db.Preload("Author").Preload("Tags").First(&article, id).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

// GetBySlug 根据 Slug 查询文章
func (r *articleRepo) GetBySlug(slug string) (*model.Article, error) {
	var article model.Article
	if err := r.db.Preload("Author").Preload("Tags").Where("slug = ?", slug).First(&article).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

// Update 更新文章
func (r *articleRepo) Update(article *model.Article) error {
	return r.db.Save(article).Error
}

// Delete 软删除文章
func (r *articleRepo) Delete(id uint) error {
	return r.db.Delete(&model.Article{}, id).Error
}

// List 分页查询文章列表，支持 Scope 条件组合
func (r *articleRepo) List(offset, limit int, scopes ...ArticleScope) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	query := r.db.Model(&model.Article{})

	// 应用所有 Scope 条件
	for _, scope := range scopes {
		query = scope(query)
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，预加载作者和标签
	if err := query.Preload("Author").Preload("Tags").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// Search 按标题和内容模糊搜索文章
func (r *articleRepo) Search(keyword string, offset, limit int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	like := "%" + keyword + "%"
	query := r.db.Model(&model.Article{}).
		Where("title LIKE ? OR content LIKE ?", like, like)

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Preload("Author").Preload("Tags").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// IncrViewCount 增加文章浏览次数
func (r *articleRepo) IncrViewCount(id uint) error {
	return r.db.Model(&model.Article{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}
