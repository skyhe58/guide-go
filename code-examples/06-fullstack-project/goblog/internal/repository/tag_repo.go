package repository

import (
	"guide-go/goblog/internal/model"

	"gorm.io/gorm"
)

// TagRepo 标签数据访问接口
type TagRepo interface {
	Create(tag *model.Tag) error
	GetByID(id uint) (*model.Tag, error)
	FindByName(name string) (*model.Tag, error)
	FindBySlug(slug string) (*model.Tag, error)
	Update(tag *model.Tag) error
	Delete(id uint) error
	List() ([]model.Tag, error)
	GetArticlesByTagID(tagID uint, offset, limit int) ([]model.Article, int64, error)
}

// tagRepo 标签数据访问实现
type tagRepo struct {
	db *gorm.DB
}

// NewTagRepo 创建标签 Repository 实例
func NewTagRepo(db *gorm.DB) TagRepo {
	return &tagRepo{db: db}
}

// Create 创建标签
func (r *tagRepo) Create(tag *model.Tag) error {
	return r.db.Create(tag).Error
}

// GetByID 根据 ID 查询标签
func (r *tagRepo) GetByID(id uint) (*model.Tag, error) {
	var tag model.Tag
	if err := r.db.First(&tag, id).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

// FindByName 根据名称查询标签
func (r *tagRepo) FindByName(name string) (*model.Tag, error) {
	var tag model.Tag
	if err := r.db.Where("name = ?", name).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

// FindBySlug 根据 Slug 查询标签
func (r *tagRepo) FindBySlug(slug string) (*model.Tag, error) {
	var tag model.Tag
	if err := r.db.Where("slug = ?", slug).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

// Update 更新标签
func (r *tagRepo) Update(tag *model.Tag) error {
	return r.db.Save(tag).Error
}

// Delete 删除标签
func (r *tagRepo) Delete(id uint) error {
	return r.db.Delete(&model.Tag{}, id).Error
}

// List 查询所有标签
func (r *tagRepo) List() ([]model.Tag, error) {
	var tags []model.Tag
	if err := r.db.Order("name ASC").Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

// GetArticlesByTagID 获取指定标签下的文章列表（分页）
func (r *tagRepo) GetArticlesByTagID(tagID uint, offset, limit int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	query := r.db.Model(&model.Article{}).
		Where("id IN (SELECT article_id FROM article_tags WHERE tag_id = ?)", tagID)

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
