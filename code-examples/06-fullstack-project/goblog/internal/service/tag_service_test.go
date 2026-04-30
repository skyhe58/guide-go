package service

import (
	"testing"

	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// newTestTagService 创建测试用的 TagService 实例
func newTestTagService(t *testing.T) (TagService, *MockTagRepo) {
	t.Helper()
	mockRepo := new(MockTagRepo)
	svc := NewTagService(mockRepo)
	return svc, mockRepo
}

// ==================== 创建标签测试 ====================

func TestTagService_Create_Success(t *testing.T) {
	// 创建标签成功：标签名不存在
	svc, mockRepo := newTestTagService(t)

	mockRepo.On("FindByName", "Go").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", mock.AnythingOfType("*model.Tag")).Return(nil)

	tag, appErr := svc.Create("Go", "go")

	assert.Nil(t, appErr)
	assert.NotNil(t, tag)
	assert.Equal(t, "Go", tag.Name)
	assert.Equal(t, "go", tag.Slug)
	mockRepo.AssertExpectations(t)
}

func TestTagService_Create_NameExists(t *testing.T) {
	// 创建标签失败：标签名已存在
	svc, mockRepo := newTestTagService(t)

	existingTag := &model.Tag{ID: 1, Name: "Go"}
	mockRepo.On("FindByName", "Go").Return(existingTag, nil)

	tag, appErr := svc.Create("Go", "go")

	assert.Nil(t, tag)
	assert.Equal(t, errcode.ErrTagNameExists, appErr)
	mockRepo.AssertExpectations(t)
}

func TestTagService_Create_DBError(t *testing.T) {
	// 创建标签失败：数据库错误
	svc, mockRepo := newTestTagService(t)

	mockRepo.On("FindByName", "Go").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", mock.AnythingOfType("*model.Tag")).Return(gorm.ErrInvalidDB)

	tag, appErr := svc.Create("Go", "go")

	assert.Nil(t, tag)
	assert.Equal(t, errcode.ErrInternal, appErr)
	mockRepo.AssertExpectations(t)
}

// ==================== 标签列表测试 ====================

func TestTagService_List_Success(t *testing.T) {
	// 获取标签列表成功
	svc, mockRepo := newTestTagService(t)

	tags := []model.Tag{
		{ID: 1, Name: "Go", Slug: "go"},
		{ID: 2, Name: "Docker", Slug: "docker"},
	}
	mockRepo.On("List").Return(tags, nil)

	result, appErr := svc.List()

	assert.Nil(t, appErr)
	assert.Len(t, result, 2)
	assert.Equal(t, "Go", result[0].Name)
	mockRepo.AssertExpectations(t)
}

func TestTagService_List_DBError(t *testing.T) {
	// 获取标签列表失败
	svc, mockRepo := newTestTagService(t)

	mockRepo.On("List").Return(nil, gorm.ErrInvalidDB)

	result, appErr := svc.List()

	assert.Nil(t, result)
	assert.Equal(t, errcode.ErrInternal, appErr)
	mockRepo.AssertExpectations(t)
}

// ==================== 标签下文章列表测试 ====================

func TestTagService_GetArticlesByTagID_Success(t *testing.T) {
	// 获取标签下文章列表成功
	svc, mockRepo := newTestTagService(t)

	tag := &model.Tag{ID: 1, Name: "Go"}
	mockRepo.On("GetByID", uint(1)).Return(tag, nil)

	articles := []model.Article{
		{ID: 1, Title: "Go 入门"},
		{ID: 2, Title: "Go 进阶"},
	}
	mockRepo.On("GetArticlesByTagID", uint(1), 0, 10).Return(articles, int64(2), nil)

	result, total, appErr := svc.GetArticlesByTagID(1, 0, 10)

	assert.Nil(t, appErr)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestTagService_GetArticlesByTagID_TagNotFound(t *testing.T) {
	// 标签不存在
	svc, mockRepo := newTestTagService(t)

	mockRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	result, total, appErr := svc.GetArticlesByTagID(999, 0, 10)

	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, errcode.ErrTagNotFound, appErr)
	mockRepo.AssertExpectations(t)
}

func TestTagService_GetArticlesByTagID_DBError(t *testing.T) {
	// 查询标签下文章失败
	svc, mockRepo := newTestTagService(t)

	tag := &model.Tag{ID: 1, Name: "Go"}
	mockRepo.On("GetByID", uint(1)).Return(tag, nil)
	mockRepo.On("GetArticlesByTagID", uint(1), 0, 10).Return([]model.Article{}, int64(0), gorm.ErrInvalidDB)

	result, total, appErr := svc.GetArticlesByTagID(1, 0, 10)

	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, errcode.ErrInternal, appErr)
	mockRepo.AssertExpectations(t)
}
