package service

import (
	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/internal/model"
	"guide-go/goblog/internal/repository"

	"gorm.io/gorm"
)

// CommentService 评论业务服务接口
type CommentService interface {
	Create(articleID, userID uint, content string) (*model.Comment, *errcode.AppError)
	ListByArticleID(articleID uint, offset, limit int) ([]model.Comment, int64, *errcode.AppError)
	Delete(id, operatorID uint, operatorRole string) *errcode.AppError
}

// commentService 评论业务服务实现
type commentService struct {
	commentRepo repository.CommentRepo
	articleRepo repository.ArticleRepo
}

// NewCommentService 创建评论服务实例
func NewCommentService(commentRepo repository.CommentRepo, articleRepo repository.ArticleRepo) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		articleRepo: articleRepo,
	}
}

// Create 发表评论
// 检查文章是否存在
func (s *commentService) Create(articleID, userID uint, content string) (*model.Comment, *errcode.AppError) {
	// 检查文章是否存在
	if _, err := s.articleRepo.GetByID(articleID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrArticleNotFound
		}
		return nil, errcode.ErrInternal
	}

	comment := &model.Comment{
		ArticleID: articleID,
		UserID:    userID,
		Content:   content,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, errcode.ErrInternal
	}

	return comment, nil
}

// ListByArticleID 获取文章评论列表（分页）
func (s *commentService) ListByArticleID(articleID uint, offset, limit int) ([]model.Comment, int64, *errcode.AppError) {
	comments, total, err := s.commentRepo.ListByArticleID(articleID, offset, limit)
	if err != nil {
		return nil, 0, errcode.ErrInternal
	}
	return comments, total, nil
}

// Delete 删除评论
// 只有评论作者或管理员可以删除
func (s *commentService) Delete(id, operatorID uint, operatorRole string) *errcode.AppError {
	comment, err := s.commentRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errcode.ErrCommentNotFound
		}
		return errcode.ErrInternal
	}

	// 权限检查：只有评论作者或管理员可以删除
	if comment.UserID != operatorID && operatorRole != "admin" {
		return errcode.ErrCommentNoPermission
	}

	if err := s.commentRepo.Delete(id); err != nil {
		return errcode.ErrInternal
	}

	return nil
}
