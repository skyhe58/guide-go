package repository

import (
	"guide-go/goblog/internal/model"

	"gorm.io/gorm"
)

// CommentRepo 评论数据访问接口
type CommentRepo interface {
	Create(comment *model.Comment) error
	GetByID(id uint) (*model.Comment, error)
	Delete(id uint) error
	ListByArticleID(articleID uint, offset, limit int) ([]model.Comment, int64, error)
}

// commentRepo 评论数据访问实现
type commentRepo struct {
	db *gorm.DB
}

// NewCommentRepo 创建评论 Repository 实例
func NewCommentRepo(db *gorm.DB) CommentRepo {
	return &commentRepo{db: db}
}

// Create 创建评论
func (r *commentRepo) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

// GetByID 根据 ID 查询评论
func (r *commentRepo) GetByID(id uint) (*model.Comment, error) {
	var comment model.Comment
	if err := r.db.Preload("User").First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

// Delete 软删除评论
func (r *commentRepo) Delete(id uint) error {
	return r.db.Delete(&model.Comment{}, id).Error
}

// ListByArticleID 按文章 ID 分页查询评论列表
func (r *commentRepo) ListByArticleID(articleID uint, offset, limit int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64

	query := r.db.Model(&model.Comment{}).Where("article_id = ?", articleID)

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询，预加载用户信息
	if err := query.Preload("User").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}
