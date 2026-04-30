package service

import (
	"testing"

	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// newTestAdminService 创建测试用的 AdminService 实例
func newTestAdminService(t *testing.T) (AdminService, *MockUserRepo, *MockArticleRepo, *MockCommentRepo, *MockTagRepo) {
	t.Helper()
	mockUserRepo := new(MockUserRepo)
	mockArticleRepo := new(MockArticleRepo)
	mockCommentRepo := new(MockCommentRepo)
	mockTagRepo := new(MockTagRepo)
	svc := NewAdminService(mockUserRepo, mockArticleRepo, mockCommentRepo, mockTagRepo)
	return svc, mockUserRepo, mockArticleRepo, mockCommentRepo, mockTagRepo
}

// ==================== 用户列表测试 ====================

func TestAdminService_ListUsers_Success(t *testing.T) {
	// 获取用户列表成功
	svc, mockUserRepo, _, _, _ := newTestAdminService(t)

	users := []model.User{
		{ID: 1, Username: "admin", Role: "admin"},
		{ID: 2, Username: "author1", Role: "author"},
	}
	mockUserRepo.On("List", 0, 10).Return(users, int64(2), nil)

	result, total, appErr := svc.ListUsers(0, 10)

	assert.Nil(t, appErr)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	mockUserRepo.AssertExpectations(t)
}

func TestAdminService_ListUsers_DBError(t *testing.T) {
	// 获取用户列表失败
	svc, mockUserRepo, _, _, _ := newTestAdminService(t)

	mockUserRepo.On("List", 0, 10).Return([]model.User{}, int64(0), gorm.ErrInvalidDB)

	result, total, appErr := svc.ListUsers(0, 10)

	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, errcode.ErrInternal, appErr)
	mockUserRepo.AssertExpectations(t)
}

// ==================== 修改用户角色测试 ====================

func TestAdminService_UpdateRole_Success(t *testing.T) {
	// 修改用户角色成功
	svc, mockUserRepo, _, _, _ := newTestAdminService(t)

	user := &model.User{ID: 2, Username: "author1", Role: "reader"}
	mockUserRepo.On("GetByID", uint(2)).Return(user, nil)
	mockUserRepo.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	appErr := svc.UpdateRole(2, 1, "author")

	assert.Nil(t, appErr)
	assert.Equal(t, "author", user.Role)
	mockUserRepo.AssertExpectations(t)
}

func TestAdminService_UpdateRole_CannotChangeSelf(t *testing.T) {
	// 不能修改自己的角色
	svc, _, _, _, _ := newTestAdminService(t)

	appErr := svc.UpdateRole(1, 1, "reader")

	assert.Equal(t, errcode.ErrCannotChangeOwnRole, appErr)
}

func TestAdminService_UpdateRole_InvalidRole(t *testing.T) {
	// 无效的角色值
	svc, _, _, _, _ := newTestAdminService(t)

	appErr := svc.UpdateRole(2, 1, "superadmin")

	assert.Equal(t, errcode.ErrInvalidRole, appErr)
}

func TestAdminService_UpdateRole_UserNotFound(t *testing.T) {
	// 用户不存在
	svc, mockUserRepo, _, _, _ := newTestAdminService(t)

	mockUserRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	appErr := svc.UpdateRole(999, 1, "author")

	assert.Equal(t, errcode.ErrUserNotFound, appErr)
	mockUserRepo.AssertExpectations(t)
}

// ==================== 表驱动测试：角色修改 ====================

func TestAdminService_UpdateRole_Table(t *testing.T) {
	tests := []struct {
		name       string
		userID     uint
		operatorID uint
		newRole    string
		setupRepo  func(*MockUserRepo)
		expectErr  *errcode.AppError
	}{
		{
			name:       "成功修改为 author",
			userID:     2,
			operatorID: 1,
			newRole:    "author",
			setupRepo: func(m *MockUserRepo) {
				user := &model.User{ID: 2, Role: "reader"}
				m.On("GetByID", uint(2)).Return(user, nil)
				m.On("Update", mock.AnythingOfType("*model.User")).Return(nil)
			},
			expectErr: nil,
		},
		{
			name:       "成功修改为 admin",
			userID:     2,
			operatorID: 1,
			newRole:    "admin",
			setupRepo: func(m *MockUserRepo) {
				user := &model.User{ID: 2, Role: "reader"}
				m.On("GetByID", uint(2)).Return(user, nil)
				m.On("Update", mock.AnythingOfType("*model.User")).Return(nil)
			},
			expectErr: nil,
		},
		{
			name:       "不能修改自己",
			userID:     1,
			operatorID: 1,
			newRole:    "reader",
			setupRepo:  func(m *MockUserRepo) {},
			expectErr:  errcode.ErrCannotChangeOwnRole,
		},
		{
			name:       "无效角色值",
			userID:     2,
			operatorID: 1,
			newRole:    "invalid",
			setupRepo:  func(m *MockUserRepo) {},
			expectErr:  errcode.ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockUserRepo, _, _, _ := newTestAdminService(t)
			tt.setupRepo(mockUserRepo)

			appErr := svc.UpdateRole(tt.userID, tt.operatorID, tt.newRole)

			if tt.expectErr == nil {
				assert.Nil(t, appErr)
			} else {
				assert.Equal(t, tt.expectErr, appErr)
			}
			mockUserRepo.AssertExpectations(t)
		})
	}
}

// ==================== 更新文章状态测试 ====================

func TestAdminService_UpdateArticleStatus_Success(t *testing.T) {
	// 更新文章状态成功
	svc, _, mockArticleRepo, _, _ := newTestAdminService(t)

	article := &model.Article{ID: 1, Status: "draft"}
	mockArticleRepo.On("GetByID", uint(1)).Return(article, nil)
	mockArticleRepo.On("Update", mock.AnythingOfType("*model.Article")).Return(nil)

	appErr := svc.UpdateArticleStatus(1, "published")

	assert.Nil(t, appErr)
	assert.Equal(t, "published", article.Status)
	mockArticleRepo.AssertExpectations(t)
}

func TestAdminService_UpdateArticleStatus_NotFound(t *testing.T) {
	// 文章不存在
	svc, _, mockArticleRepo, _, _ := newTestAdminService(t)

	mockArticleRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	appErr := svc.UpdateArticleStatus(999, "published")

	assert.Equal(t, errcode.ErrArticleNotFound, appErr)
	mockArticleRepo.AssertExpectations(t)
}

func TestAdminService_UpdateArticleStatus_DBError(t *testing.T) {
	// 更新文章状态失败：数据库错误
	svc, _, mockArticleRepo, _, _ := newTestAdminService(t)

	article := &model.Article{ID: 1, Status: "draft"}
	mockArticleRepo.On("GetByID", uint(1)).Return(article, nil)
	mockArticleRepo.On("Update", mock.AnythingOfType("*model.Article")).Return(gorm.ErrInvalidDB)

	appErr := svc.UpdateArticleStatus(1, "published")

	assert.Equal(t, errcode.ErrInternal, appErr)
	mockArticleRepo.AssertExpectations(t)
}

// ==================== 系统统计测试 ====================

func TestAdminService_GetStats_Success(t *testing.T) {
	// 获取系统统计成功
	svc, mockUserRepo, mockArticleRepo, _, mockTagRepo := newTestAdminService(t)

	// GetStats 调用 userRepo.List 两次（第一次结果被忽略，第二次获取总数）
	mockUserRepo.On("List", 0, 1).Return([]model.User{}, int64(100), nil)
	mockArticleRepo.On("List", 0, 1, mock.Anything).Return([]model.Article{}, int64(50), nil)
	tags := []model.Tag{
		{ID: 1, Name: "Go"},
		{ID: 2, Name: "Docker"},
		{ID: 3, Name: "K8s"},
	}
	mockTagRepo.On("List").Return(tags, nil)

	stats, appErr := svc.GetStats()

	assert.Nil(t, appErr)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(100), stats.UserCount)
	assert.Equal(t, int64(50), stats.ArticleCount)
	assert.Equal(t, int64(0), stats.CommentCount) // 简化处理，始终为 0
	assert.Equal(t, int64(3), stats.TagCount)
	mockUserRepo.AssertExpectations(t)
	mockArticleRepo.AssertExpectations(t)
	mockTagRepo.AssertExpectations(t)
}

func TestAdminService_GetStats_UserRepoError(t *testing.T) {
	// 获取统计失败：用户查询出错
	svc, mockUserRepo, _, _, _ := newTestAdminService(t)

	mockUserRepo.On("List", 0, 1).Return([]model.User{}, int64(0), gorm.ErrInvalidDB)

	stats, appErr := svc.GetStats()

	assert.Nil(t, stats)
	assert.Equal(t, errcode.ErrInternal, appErr)
	mockUserRepo.AssertExpectations(t)
}
