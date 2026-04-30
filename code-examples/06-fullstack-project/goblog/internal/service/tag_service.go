package service

import (
	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/internal/model"
	"guide-go/goblog/internal/repository"

	"gorm.io/gorm"
)

// TagService 标签业务服务接口
type TagService interface {
	Create(name, slug string) (*model.Tag, *errcode.AppError)
	List() ([]model.Tag, *errcode.AppError)
	GetArticlesByTagID(tagID uint, offset, limit int) ([]model.Article, int64, *errcode.AppError)
}

// tagService 标签业务服务实现
type tagService struct {
	tagRepo repository.TagRepo
}

// NewTagService 创建标签服务实例
func NewTagService(tagRepo repository.TagRepo) TagService {
	return &tagService{tagRepo: tagRepo}
}

// Create 创建标签
// 检查标签名唯一性
func (s *tagService) Create(name, slug string) (*model.Tag, *errcode.AppError) {
	// 检查标签名是否已存在
	if _, err := s.tagRepo.FindByName(name); err == nil {
		return nil, errcode.ErrTagNameExists
	}

	tag := &model.Tag{
		Name: name,
		Slug: slug,
	}

	if err := s.tagRepo.Create(tag); err != nil {
		return nil, errcode.ErrInternal
	}

	return tag, nil
}

// List 获取所有标签列表
func (s *tagService) List() ([]model.Tag, *errcode.AppError) {
	tags, err := s.tagRepo.List()
	if err != nil {
		return nil, errcode.ErrInternal
	}
	return tags, nil
}

// GetArticlesByTagID 获取指定标签下的文章列表
func (s *tagService) GetArticlesByTagID(tagID uint, offset, limit int) ([]model.Article, int64, *errcode.AppError) {
	// 检查标签是否存在
	if _, err := s.tagRepo.GetByID(tagID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, errcode.ErrTagNotFound
		}
		return nil, 0, errcode.ErrInternal
	}

	articles, total, err := s.tagRepo.GetArticlesByTagID(tagID, offset, limit)
	if err != nil {
		return nil, 0, errcode.ErrInternal
	}

	return articles, total, nil
}
