package service

import (
	"guide-go/goblog/internal/model"
	"guide-go/goblog/internal/repository"

	"github.com/stretchr/testify/mock"
)

// ==================== MockUserRepo ====================

// MockUserRepo 用户 Repository Mock 实现
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepo) GetByID(id uint) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) FindByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) FindByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepo) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepo) List(offset, limit int) ([]model.User, int64, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

// ==================== MockArticleRepo ====================

// MockArticleRepo 文章 Repository Mock 实现
type MockArticleRepo struct {
	mock.Mock
}

func (m *MockArticleRepo) Create(article *model.Article) error {
	args := m.Called(article)
	return args.Error(0)
}

func (m *MockArticleRepo) GetByID(id uint) (*model.Article, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Article), args.Error(1)
}

func (m *MockArticleRepo) GetBySlug(slug string) (*model.Article, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Article), args.Error(1)
}

func (m *MockArticleRepo) Update(article *model.Article) error {
	args := m.Called(article)
	return args.Error(0)
}

func (m *MockArticleRepo) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockArticleRepo) List(offset, limit int, scopes ...repository.ArticleScope) ([]model.Article, int64, error) {
	args := m.Called(offset, limit, scopes)
	return args.Get(0).([]model.Article), args.Get(1).(int64), args.Error(2)
}

func (m *MockArticleRepo) Search(keyword string, offset, limit int) ([]model.Article, int64, error) {
	args := m.Called(keyword, offset, limit)
	return args.Get(0).([]model.Article), args.Get(1).(int64), args.Error(2)
}

func (m *MockArticleRepo) IncrViewCount(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// ==================== MockTagRepo ====================

// MockTagRepo 标签 Repository Mock 实现
type MockTagRepo struct {
	mock.Mock
}

func (m *MockTagRepo) Create(tag *model.Tag) error {
	args := m.Called(tag)
	return args.Error(0)
}

func (m *MockTagRepo) GetByID(id uint) (*model.Tag, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tag), args.Error(1)
}

func (m *MockTagRepo) FindByName(name string) (*model.Tag, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tag), args.Error(1)
}

func (m *MockTagRepo) FindBySlug(slug string) (*model.Tag, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Tag), args.Error(1)
}

func (m *MockTagRepo) Update(tag *model.Tag) error {
	args := m.Called(tag)
	return args.Error(0)
}

func (m *MockTagRepo) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTagRepo) List() ([]model.Tag, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Tag), args.Error(1)
}

func (m *MockTagRepo) GetArticlesByTagID(tagID uint, offset, limit int) ([]model.Article, int64, error) {
	args := m.Called(tagID, offset, limit)
	return args.Get(0).([]model.Article), args.Get(1).(int64), args.Error(2)
}

// ==================== MockCommentRepo ====================

// MockCommentRepo 评论 Repository Mock 实现
type MockCommentRepo struct {
	mock.Mock
}

func (m *MockCommentRepo) Create(comment *model.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockCommentRepo) GetByID(id uint) (*model.Comment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Comment), args.Error(1)
}

func (m *MockCommentRepo) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCommentRepo) ListByArticleID(articleID uint, offset, limit int) ([]model.Comment, int64, error) {
	args := m.Called(articleID, offset, limit)
	return args.Get(0).([]model.Comment), args.Get(1).(int64), args.Error(2)
}
