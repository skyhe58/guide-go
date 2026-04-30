package service

import (
	"testing"

	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// newTestCommentService 创建测试用的 CommentService 实例
func newTestCommentService(t *testing.T) (CommentService, *MockCommentRepo, *MockArticleRepo) {
	t.Helper()
	mockCommentRepo := new(MockCommentRepo)
	mockArticleRepo := new(MockArticleRepo)
	svc := NewCommentService(mockCommentRepo, mockArticleRepo)
	return svc, mockCommentRepo, mockArticleRepo
}

// ==================== 发表评论测试 ====================

func TestCommentService_Create_Success(t *testing.T) {
	// 发表评论成功：文章存在
	svc, mockCommentRepo, mockArticleRepo := newTestCommentService(t)

	article := &model.Article{ID: 1, Title: "测试文章"}
	mockArticleRepo.On("GetByID", uint(1)).Return(article, nil)
	mockCommentRepo.On("Create", mock.AnythingOfType("*model.Comment")).Return(nil)

	comment, appErr := svc.Create(1, 10, "这是一条评论")

	assert.Nil(t, appErr)
	assert.NotNil(t, comment)
	assert.Equal(t, uint(1), comment.ArticleID)
	assert.Equal(t, uint(10), comment.UserID)
	assert.Equal(t, "这是一条评论", comment.Content)
	mockCommentRepo.AssertExpectations(t)
	mockArticleRepo.AssertExpectations(t)
}

func TestCommentService_Create_ArticleNotFound(t *testing.T) {
	// 发表评论失败：文章不存在
	svc, _, mockArticleRepo := newTestCommentService(t)

	mockArticleRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	comment, appErr := svc.Create(999, 10, "评论内容")

	assert.Nil(t, comment)
	assert.Equal(t, errcode.ErrArticleNotFound, appErr)
	mockArticleRepo.AssertExpectations(t)
}

func TestCommentService_Create_DBError(t *testing.T) {
	// 发表评论失败：数据库错误
	svc, mockCommentRepo, mockArticleRepo := newTestCommentService(t)

	article := &model.Article{ID: 1}
	mockArticleRepo.On("GetByID", uint(1)).Return(article, nil)
	mockCommentRepo.On("Create", mock.AnythingOfType("*model.Comment")).Return(gorm.ErrInvalidDB)

	comment, appErr := svc.Create(1, 10, "评论内容")

	assert.Nil(t, comment)
	assert.Equal(t, errcode.ErrInternal, appErr)
	mockCommentRepo.AssertExpectations(t)
}

// ==================== 评论列表测试 ====================

func TestCommentService_ListByArticleID_Success(t *testing.T) {
	// 获取文章评论列表成功
	svc, mockCommentRepo, _ := newTestCommentService(t)

	comments := []model.Comment{
		{ID: 1, ArticleID: 1, UserID: 10, Content: "评论一"},
		{ID: 2, ArticleID: 1, UserID: 20, Content: "评论二"},
	}
	mockCommentRepo.On("ListByArticleID", uint(1), 0, 10).Return(comments, int64(2), nil)

	result, total, appErr := svc.ListByArticleID(1, 0, 10)

	assert.Nil(t, appErr)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	mockCommentRepo.AssertExpectations(t)
}

func TestCommentService_ListByArticleID_DBError(t *testing.T) {
	// 获取评论列表失败
	svc, mockCommentRepo, _ := newTestCommentService(t)

	mockCommentRepo.On("ListByArticleID", uint(1), 0, 10).Return([]model.Comment{}, int64(0), gorm.ErrInvalidDB)

	result, total, appErr := svc.ListByArticleID(1, 0, 10)

	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, errcode.ErrInternal, appErr)
	mockCommentRepo.AssertExpectations(t)
}

// ==================== 删除评论测试 ====================

func TestCommentService_Delete_ByAuthor(t *testing.T) {
	// 评论作者删除自己的评论
	svc, mockCommentRepo, _ := newTestCommentService(t)

	comment := &model.Comment{ID: 1, UserID: 10}
	mockCommentRepo.On("GetByID", uint(1)).Return(comment, nil)
	mockCommentRepo.On("Delete", uint(1)).Return(nil)

	appErr := svc.Delete(1, 10, "reader")

	assert.Nil(t, appErr)
	mockCommentRepo.AssertExpectations(t)
}

func TestCommentService_Delete_ByAdmin(t *testing.T) {
	// 管理员删除任意评论
	svc, mockCommentRepo, _ := newTestCommentService(t)

	comment := &model.Comment{ID: 1, UserID: 10}
	mockCommentRepo.On("GetByID", uint(1)).Return(comment, nil)
	mockCommentRepo.On("Delete", uint(1)).Return(nil)

	appErr := svc.Delete(1, 99, "admin")

	assert.Nil(t, appErr)
	mockCommentRepo.AssertExpectations(t)
}

func TestCommentService_Delete_NoPermission(t *testing.T) {
	// 非作者非管理员无权删除评论
	svc, mockCommentRepo, _ := newTestCommentService(t)

	comment := &model.Comment{ID: 1, UserID: 10}
	mockCommentRepo.On("GetByID", uint(1)).Return(comment, nil)

	appErr := svc.Delete(1, 20, "reader")

	assert.Equal(t, errcode.ErrCommentNoPermission, appErr)
	mockCommentRepo.AssertExpectations(t)
}

func TestCommentService_Delete_NotFound(t *testing.T) {
	// 删除不存在的评论
	svc, mockCommentRepo, _ := newTestCommentService(t)

	mockCommentRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	appErr := svc.Delete(999, 10, "reader")

	assert.Equal(t, errcode.ErrCommentNotFound, appErr)
	mockCommentRepo.AssertExpectations(t)
}

// ==================== 表驱动测试：评论删除权限 ====================

func TestCommentService_Delete_PermissionTable(t *testing.T) {
	tests := []struct {
		name         string
		commentOwner uint
		operatorID   uint
		operatorRole string
		expectErr    *errcode.AppError
	}{
		{
			name:         "评论作者删除",
			commentOwner: 10,
			operatorID:   10,
			operatorRole: "reader",
			expectErr:    nil,
		},
		{
			name:         "管理员删除",
			commentOwner: 10,
			operatorID:   99,
			operatorRole: "admin",
			expectErr:    nil,
		},
		{
			name:         "其他用户无权删除",
			commentOwner: 10,
			operatorID:   20,
			operatorRole: "reader",
			expectErr:    errcode.ErrCommentNoPermission,
		},
		{
			name:         "其他作者无权删除",
			commentOwner: 10,
			operatorID:   11,
			operatorRole: "author",
			expectErr:    errcode.ErrCommentNoPermission,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockCommentRepo, _ := newTestCommentService(t)

			comment := &model.Comment{ID: 1, UserID: tt.commentOwner}
			mockCommentRepo.On("GetByID", uint(1)).Return(comment, nil)
			if tt.expectErr == nil {
				mockCommentRepo.On("Delete", uint(1)).Return(nil)
			}

			appErr := svc.Delete(1, tt.operatorID, tt.operatorRole)

			if tt.expectErr == nil {
				assert.Nil(t, appErr)
			} else {
				assert.Equal(t, tt.expectErr, appErr)
			}
			mockCommentRepo.AssertExpectations(t)
		})
	}
}
