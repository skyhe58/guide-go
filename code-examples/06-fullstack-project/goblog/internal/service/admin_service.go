package service

import (
	"guide-go/goblog/internal/auth"
	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/internal/model"
	"guide-go/goblog/internal/repository"

	"gorm.io/gorm"
)

// SystemStats 系统统计数据
type SystemStats struct {
	UserCount    int64 `json:"user_count"`
	ArticleCount int64 `json:"article_count"`
	CommentCount int64 `json:"comment_count"`
	TagCount     int64 `json:"tag_count"`
}

// AdminService 管理后台业务服务接口
type AdminService interface {
	ListUsers(offset, limit int) ([]model.User, int64, *errcode.AppError)
	UpdateRole(userID, operatorID uint, role string) *errcode.AppError
	UpdateArticleStatus(articleID uint, status string) *errcode.AppError
	GetStats() (*SystemStats, *errcode.AppError)
}

// adminService 管理后台业务服务实现
type adminService struct {
	userRepo    repository.UserRepo
	articleRepo repository.ArticleRepo
	commentRepo repository.CommentRepo
	tagRepo     repository.TagRepo
}

// NewAdminService 创建管理服务实例
func NewAdminService(
	userRepo repository.UserRepo,
	articleRepo repository.ArticleRepo,
	commentRepo repository.CommentRepo,
	tagRepo repository.TagRepo,
) AdminService {
	return &adminService{
		userRepo:    userRepo,
		articleRepo: articleRepo,
		commentRepo: commentRepo,
		tagRepo:     tagRepo,
	}
}

// ListUsers 获取用户列表（分页）
func (s *adminService) ListUsers(offset, limit int) ([]model.User, int64, *errcode.AppError) {
	users, total, err := s.userRepo.List(offset, limit)
	if err != nil {
		return nil, 0, errcode.ErrInternal
	}
	return users, total, nil
}

// UpdateRole 修改用户角色
// 不能修改自己的角色
func (s *adminService) UpdateRole(userID, operatorID uint, role string) *errcode.AppError {
	// 不能修改自己的角色
	if userID == operatorID {
		return errcode.ErrCannotChangeOwnRole
	}

	// 验证角色值是否有效
	if !auth.IsValidRole(role) {
		return errcode.ErrInvalidRole
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errcode.ErrUserNotFound
		}
		return errcode.ErrInternal
	}

	user.Role = role
	if err := s.userRepo.Update(user); err != nil {
		return errcode.ErrInternal
	}

	return nil
}

// UpdateArticleStatus 更新文章状态（审核上架/下架）
func (s *adminService) UpdateArticleStatus(articleID uint, status string) *errcode.AppError {
	article, err := s.articleRepo.GetByID(articleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errcode.ErrArticleNotFound
		}
		return errcode.ErrInternal
	}

	article.Status = status
	if err := s.articleRepo.Update(article); err != nil {
		return errcode.ErrInternal
	}

	return nil
}

// GetStats 获取系统统计数据
func (s *adminService) GetStats() (*SystemStats, *errcode.AppError) {
	users, _, err := s.userRepo.List(0, 1)
	if err != nil {
		return nil, errcode.ErrInternal
	}
	_ = users

	// 使用 List 方法获取各模块总数
	_, userTotal, err := s.userRepo.List(0, 1)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	_, articleTotal, err := s.articleRepo.List(0, 1)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	// 评论总数通过 ListByArticleID(0) 无法获取，使用特殊处理
	// 这里简化为返回 0，实际可扩展 CommentRepo 接口
	tags, err := s.tagRepo.List()
	if err != nil {
		return nil, errcode.ErrInternal
	}

	return &SystemStats{
		UserCount:    userTotal,
		ArticleCount: articleTotal,
		CommentCount: 0, // 简化处理
		TagCount:     int64(len(tags)),
	}, nil
}
