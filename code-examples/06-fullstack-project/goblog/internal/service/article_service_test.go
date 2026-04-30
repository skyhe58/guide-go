package service

import (
	"testing"

	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/internal/model"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newTestArticleService 创建测试用的 ArticleService 实例
func newTestArticleService(t *testing.T) (ArticleService, *MockArticleRepo, *MockTagRepo, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	mockArticleRepo := new(MockArticleRepo)
	mockTagRepo := new(MockTagRepo)
	svc := NewArticleService(mockArticleRepo, mockTagRepo, rdb)

	t.Cleanup(func() {
		rdb.Close()
		mr.Close()
	})

	return svc, mockArticleRepo, mockTagRepo, mr
}

// ==================== 创建文章测试 ====================

func TestArticleService_Create_Success(t *testing.T) {
	// 创建文章成功
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	mockArticleRepo.On("Create", mock.AnythingOfType("*model.Article")).Return(nil)

	article, appErr := svc.Create(1, "测试标题", "测试内容", "test-slug", "draft", []uint{1, 2})

	assert.Nil(t, appErr)
	assert.NotNil(t, article)
	assert.Equal(t, uint(1), article.AuthorID)
	assert.Equal(t, "测试标题", article.Title)
	assert.Equal(t, "draft", article.Status)
	assert.Len(t, article.Tags, 2)
	mockArticleRepo.AssertExpectations(t)
}

func TestArticleService_Create_Published(t *testing.T) {
	// 创建已发布文章：应设置 PublishedAt
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	mockArticleRepo.On("Create", mock.AnythingOfType("*model.Article")).Return(nil)

	article, appErr := svc.Create(1, "已发布文章", "内容", "published-slug", "published", nil)

	assert.Nil(t, appErr)
	assert.NotNil(t, article)
	assert.Equal(t, "published", article.Status)
	assert.NotNil(t, article.PublishedAt, "已发布文章应设置 PublishedAt")
	mockArticleRepo.AssertExpectations(t)
}

func TestArticleService_Create_DBError(t *testing.T) {
	// 创建文章失败：数据库错误
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	mockArticleRepo.On("Create", mock.AnythingOfType("*model.Article")).Return(gorm.ErrInvalidDB)

	article, appErr := svc.Create(1, "标题", "内容", "slug", "draft", nil)

	assert.Nil(t, article)
	assert.Equal(t, errcode.ErrInternal, appErr)
	mockArticleRepo.AssertExpectations(t)
}

// ==================== 更新文章测试 ====================

func TestArticleService_Update_ByAuthor(t *testing.T) {
	// 文章作者更新自己的文章
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	existingArticle := &model.Article{
		ID:       1,
		AuthorID: 10,
		Title:    "旧标题",
		Content:  "旧内容",
		Status:   "draft",
	}
	mockArticleRepo.On("GetByID", uint(1)).Return(existingArticle, nil)
	mockArticleRepo.On("Update", mock.AnythingOfType("*model.Article")).Return(nil)

	article, appErr := svc.Update(1, 10, "author", "新标题", "新内容", "", "", nil)

	assert.Nil(t, appErr)
	assert.NotNil(t, article)
	assert.Equal(t, "新标题", article.Title)
	assert.Equal(t, "新内容", article.Content)
	mockArticleRepo.AssertExpectations(t)
}

func TestArticleService_Update_ByAdmin(t *testing.T) {
	// 管理员更新他人文章
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	existingArticle := &model.Article{
		ID:       1,
		AuthorID: 10,
		Title:    "旧标题",
		Status:   "draft",
	}
	mockArticleRepo.On("GetByID", uint(1)).Return(existingArticle, nil)
	mockArticleRepo.On("Update", mock.AnythingOfType("*model.Article")).Return(nil)

	article, appErr := svc.Update(1, 99, "admin", "管理员修改", "", "", "published", nil)

	assert.Nil(t, appErr)
	assert.NotNil(t, article)
	assert.Equal(t, "管理员修改", article.Title)
	assert.Equal(t, "published", article.Status)
	mockArticleRepo.AssertExpectations(t)
}

func TestArticleService_Update_NoPermission(t *testing.T) {
	// 非作者非管理员无权更新
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	existingArticle := &model.Article{
		ID:       1,
		AuthorID: 10,
	}
	mockArticleRepo.On("GetByID", uint(1)).Return(existingArticle, nil)

	article, appErr := svc.Update(1, 20, "reader", "新标题", "", "", "", nil)

	assert.Nil(t, article)
	assert.Equal(t, errcode.ErrArticleNoPermission, appErr)
	mockArticleRepo.AssertExpectations(t)
}

func TestArticleService_Update_NotFound(t *testing.T) {
	// 更新不存在的文章
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	mockArticleRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	article, appErr := svc.Update(999, 1, "admin", "标题", "", "", "", nil)

	assert.Nil(t, article)
	assert.Equal(t, errcode.ErrArticleNotFound, appErr)
	mockArticleRepo.AssertExpectations(t)
}

// ==================== 删除文章测试 ====================

func TestArticleService_Delete_ByAuthor(t *testing.T) {
	// 作者删除自己的文章
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	existingArticle := &model.Article{ID: 1, AuthorID: 10}
	mockArticleRepo.On("GetByID", uint(1)).Return(existingArticle, nil)
	mockArticleRepo.On("Delete", uint(1)).Return(nil)

	appErr := svc.Delete(1, 10, "author")

	assert.Nil(t, appErr)
	mockArticleRepo.AssertExpectations(t)
}

func TestArticleService_Delete_ByAdmin(t *testing.T) {
	// 管理员删除他人文章
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	existingArticle := &model.Article{ID: 1, AuthorID: 10}
	mockArticleRepo.On("GetByID", uint(1)).Return(existingArticle, nil)
	mockArticleRepo.On("Delete", uint(1)).Return(nil)

	appErr := svc.Delete(1, 99, "admin")

	assert.Nil(t, appErr)
	mockArticleRepo.AssertExpectations(t)
}

func TestArticleService_Delete_NoPermission(t *testing.T) {
	// 非作者非管理员无权删除
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	existingArticle := &model.Article{ID: 1, AuthorID: 10}
	mockArticleRepo.On("GetByID", uint(1)).Return(existingArticle, nil)

	appErr := svc.Delete(1, 20, "reader")

	assert.Equal(t, errcode.ErrArticleNoPermission, appErr)
	mockArticleRepo.AssertExpectations(t)
}

func TestArticleService_Delete_NotFound(t *testing.T) {
	// 删除不存在的文章
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	mockArticleRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	appErr := svc.Delete(999, 1, "admin")

	assert.Equal(t, errcode.ErrArticleNotFound, appErr)
	mockArticleRepo.AssertExpectations(t)
}

// ==================== 文章详情缓存测试 ====================

func TestArticleService_GetByID_CacheMiss(t *testing.T) {
	// 缓存未命中：从数据库查询并写入缓存
	svc, mockArticleRepo, _, mr := newTestArticleService(t)

	expectedArticle := &model.Article{
		ID:       1,
		AuthorID: 10,
		Title:    "测试文章",
		Content:  "文章内容",
		Status:   "published",
	}
	mockArticleRepo.On("GetByID", uint(1)).Return(expectedArticle, nil)
	mockArticleRepo.On("IncrViewCount", uint(1)).Return(nil)

	article, appErr := svc.GetByID(1)

	assert.Nil(t, appErr)
	assert.NotNil(t, article)
	assert.Equal(t, "测试文章", article.Title)
	// 验证缓存已写入
	assert.True(t, mr.Exists("article:1"), "文章详情应写入 Redis 缓存")
	mockArticleRepo.AssertExpectations(t)
}

func TestArticleService_GetByID_NotFound(t *testing.T) {
	// 文章不存在：应设置空值缓存防穿透
	svc, mockArticleRepo, _, mr := newTestArticleService(t)

	mockArticleRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	article, appErr := svc.GetByID(999)

	assert.Nil(t, article)
	assert.Equal(t, errcode.ErrArticleNotFound, appErr)
	// 验证空值缓存已设置（防穿透）
	assert.True(t, mr.Exists("article:999:null"), "应设置空值缓存防止缓存穿透")
	mockArticleRepo.AssertExpectations(t)
}

// ==================== 文章列表测试 ====================

func TestArticleService_List_Success(t *testing.T) {
	// 文章列表查询成功
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	articles := []model.Article{
		{ID: 1, Title: "文章一"},
		{ID: 2, Title: "文章二"},
	}
	mockArticleRepo.On("List", 0, 10, mock.Anything).Return(articles, int64(2), nil)

	result, total, appErr := svc.List(0, 10, "", 0)

	assert.Nil(t, appErr)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	mockArticleRepo.AssertExpectations(t)
}

func TestArticleService_List_WithFilters(t *testing.T) {
	// 带筛选条件的文章列表
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	articles := []model.Article{
		{ID: 1, Title: "已发布文章", Status: "published"},
	}
	mockArticleRepo.On("List", 0, 10, mock.Anything).Return(articles, int64(1), nil)

	result, total, appErr := svc.List(0, 10, "published", uint(1))

	assert.Nil(t, appErr)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	mockArticleRepo.AssertExpectations(t)
}

func TestArticleService_List_DBError(t *testing.T) {
	// 文章列表查询失败
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	mockArticleRepo.On("List", 0, 10, mock.Anything).Return([]model.Article{}, int64(0), gorm.ErrInvalidDB)

	result, total, appErr := svc.List(0, 10, "", 0)

	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, errcode.ErrInternal, appErr)
	mockArticleRepo.AssertExpectations(t)
}

// ==================== 文章搜索测试 ====================

func TestArticleService_Search_Success(t *testing.T) {
	// 搜索文章成功
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	articles := []model.Article{
		{ID: 1, Title: "Go 并发编程"},
	}
	mockArticleRepo.On("Search", "Go", 0, 10).Return(articles, int64(1), nil)

	result, total, appErr := svc.Search("Go", 0, 10)

	assert.Nil(t, appErr)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.Equal(t, "Go 并发编程", result[0].Title)
	mockArticleRepo.AssertExpectations(t)
}

func TestArticleService_Search_DBError(t *testing.T) {
	// 搜索文章失败
	svc, mockArticleRepo, _, _ := newTestArticleService(t)

	mockArticleRepo.On("Search", "Go", 0, 10).Return([]model.Article{}, int64(0), gorm.ErrInvalidDB)

	result, total, appErr := svc.Search("Go", 0, 10)

	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, errcode.ErrInternal, appErr)
	mockArticleRepo.AssertExpectations(t)
}

// ==================== 热门文章测试 ====================

func TestArticleService_GetHotArticles_Empty(t *testing.T) {
	// 热门文章排行榜为空
	svc, _, _, _ := newTestArticleService(t)

	articles, appErr := svc.GetHotArticles(10)

	assert.Nil(t, appErr)
	// Redis Sorted Set 为空时返回空列表
	assert.Empty(t, articles)
}

// ==================== 表驱动测试：文章删除权限 ====================

func TestArticleService_Delete_PermissionTable(t *testing.T) {
	// 表驱动测试：验证不同角色的删除权限
	tests := []struct {
		name         string
		authorID     uint
		operatorID   uint
		operatorRole string
		expectErr    *errcode.AppError
	}{
		{
			name:         "作者删除自己的文章",
			authorID:     10,
			operatorID:   10,
			operatorRole: "author",
			expectErr:    nil,
		},
		{
			name:         "管理员删除任意文章",
			authorID:     10,
			operatorID:   99,
			operatorRole: "admin",
			expectErr:    nil,
		},
		{
			name:         "读者无权删除",
			authorID:     10,
			operatorID:   20,
			operatorRole: "reader",
			expectErr:    errcode.ErrArticleNoPermission,
		},
		{
			name:         "其他作者无权删除",
			authorID:     10,
			operatorID:   11,
			operatorRole: "author",
			expectErr:    errcode.ErrArticleNoPermission,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockArticleRepo, _, _ := newTestArticleService(t)

			existingArticle := &model.Article{ID: 1, AuthorID: tt.authorID}
			mockArticleRepo.On("GetByID", uint(1)).Return(existingArticle, nil)
			if tt.expectErr == nil {
				mockArticleRepo.On("Delete", uint(1)).Return(nil)
			}

			appErr := svc.Delete(1, tt.operatorID, tt.operatorRole)

			if tt.expectErr == nil {
				assert.Nil(t, appErr)
			} else {
				assert.Equal(t, tt.expectErr, appErr)
			}
			mockArticleRepo.AssertExpectations(t)
		})
	}
}

// ==================== 表驱动测试：文章更新权限 ====================

func TestArticleService_Update_PermissionTable(t *testing.T) {
	tests := []struct {
		name         string
		authorID     uint
		operatorID   uint
		operatorRole string
		expectErr    *errcode.AppError
	}{
		{
			name:         "作者更新自己的文章",
			authorID:     10,
			operatorID:   10,
			operatorRole: "author",
			expectErr:    nil,
		},
		{
			name:         "管理员更新任意文章",
			authorID:     10,
			operatorID:   99,
			operatorRole: "admin",
			expectErr:    nil,
		},
		{
			name:         "读者无权更新",
			authorID:     10,
			operatorID:   20,
			operatorRole: "reader",
			expectErr:    errcode.ErrArticleNoPermission,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockArticleRepo, _, _ := newTestArticleService(t)

			existingArticle := &model.Article{ID: 1, AuthorID: tt.authorID, Title: "旧标题"}
			mockArticleRepo.On("GetByID", uint(1)).Return(existingArticle, nil)
			if tt.expectErr == nil {
				mockArticleRepo.On("Update", mock.AnythingOfType("*model.Article")).Return(nil)
			}

			_, appErr := svc.Update(1, tt.operatorID, tt.operatorRole, "新标题", "", "", "", nil)

			if tt.expectErr == nil {
				assert.Nil(t, appErr)
			} else {
				assert.Equal(t, tt.expectErr, appErr)
			}
			mockArticleRepo.AssertExpectations(t)
		})
	}
}
