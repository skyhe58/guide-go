package service

import (
	"testing"
	"time"

	"guide-go/goblog/internal/auth"
	"guide-go/goblog/internal/config"
	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/internal/model"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newTestUserService 创建测试用的 UserService 实例
// 使用 miniredis 模拟 Redis，避免依赖真实 Redis 服务
func newTestUserService(t *testing.T) (UserService, *MockUserRepo, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	jwtCfg := &config.JWTConfig{
		Secret:          "test-secret-key-for-unit-tests",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "goblog-test",
	}
	jwtSvc := auth.NewJWTService(jwtCfg)
	mockRepo := new(MockUserRepo)
	svc := NewUserService(mockRepo, jwtSvc, rdb)

	t.Cleanup(func() {
		rdb.Close()
		mr.Close()
	})

	return svc, mockRepo, mr
}

// ==================== 用户注册测试 ====================

func TestUserService_Register_Success(t *testing.T) {
	// 注册成功：用户名和邮箱均不存在
	svc, mockRepo, _ := newTestUserService(t)

	mockRepo.On("FindByUsername", "testuser").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("FindByEmail", "test@example.com").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(nil)

	user, appErr := svc.Register("testuser", "test@example.com", "password123")

	assert.Nil(t, appErr)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, auth.RoleReader, user.Role)
	// 密码应已加密，不应为明文
	assert.NotEqual(t, "password123", user.PasswordHash)
	assert.NotEmpty(t, user.PasswordHash)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_UsernameExists(t *testing.T) {
	// 注册失败：用户名已存在
	svc, mockRepo, _ := newTestUserService(t)

	existingUser := &model.User{ID: 1, Username: "testuser"}
	mockRepo.On("FindByUsername", "testuser").Return(existingUser, nil)

	user, appErr := svc.Register("testuser", "test@example.com", "password123")

	assert.Nil(t, user)
	assert.Equal(t, errcode.ErrUsernameExists, appErr)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_EmailExists(t *testing.T) {
	// 注册失败：邮箱已注册
	svc, mockRepo, _ := newTestUserService(t)

	mockRepo.On("FindByUsername", "newuser").Return(nil, gorm.ErrRecordNotFound)
	existingUser := &model.User{ID: 1, Email: "test@example.com"}
	mockRepo.On("FindByEmail", "test@example.com").Return(existingUser, nil)

	user, appErr := svc.Register("newuser", "test@example.com", "password123")

	assert.Nil(t, user)
	assert.Equal(t, errcode.ErrEmailExists, appErr)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_CreateFailed(t *testing.T) {
	// 注册失败：数据库创建用户出错
	svc, mockRepo, _ := newTestUserService(t)

	mockRepo.On("FindByUsername", "testuser").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("FindByEmail", "test@example.com").Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(gorm.ErrInvalidDB)

	user, appErr := svc.Register("testuser", "test@example.com", "password123")

	assert.Nil(t, user)
	assert.Equal(t, errcode.ErrInternal, appErr)
	mockRepo.AssertExpectations(t)
}

// ==================== 用户登录测试 ====================

func TestUserService_Login_Success(t *testing.T) {
	// 登录成功：用户名和密码正确
	svc, mockRepo, _ := newTestUserService(t)

	hash, _ := auth.HashPassword("password123")
	existingUser := &model.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: hash,
		Role:         auth.RoleReader,
	}
	mockRepo.On("FindByUsername", "testuser").Return(existingUser, nil)

	tokenPair, appErr := svc.Login("testuser", "password123")

	assert.Nil(t, appErr)
	assert.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.True(t, tokenPair.ExpiresIn > 0)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	// 登录失败：用户不存在
	svc, mockRepo, _ := newTestUserService(t)

	mockRepo.On("FindByUsername", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	tokenPair, appErr := svc.Login("nonexistent", "password123")

	assert.Nil(t, tokenPair)
	assert.Equal(t, errcode.ErrUserNotFound, appErr)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	// 登录失败：密码错误
	svc, mockRepo, _ := newTestUserService(t)

	hash, _ := auth.HashPassword("correct-password")
	existingUser := &model.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: hash,
		Role:         auth.RoleReader,
	}
	mockRepo.On("FindByUsername", "testuser").Return(existingUser, nil)

	tokenPair, appErr := svc.Login("testuser", "wrong-password")

	assert.Nil(t, tokenPair)
	assert.Equal(t, errcode.ErrPasswordWrong, appErr)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Login_DBError(t *testing.T) {
	// 登录失败：数据库查询出错
	svc, mockRepo, _ := newTestUserService(t)

	mockRepo.On("FindByUsername", "testuser").Return(nil, gorm.ErrInvalidDB)

	tokenPair, appErr := svc.Login("testuser", "password123")

	assert.Nil(t, tokenPair)
	assert.Equal(t, errcode.ErrInternal, appErr)
	mockRepo.AssertExpectations(t)
}

// ==================== Token 刷新测试 ====================

func TestUserService_RefreshToken_Success(t *testing.T) {
	// Token 刷新成功：使用有效的 Refresh Token
	svc, mockRepo, _ := newTestUserService(t)
	_ = mockRepo // 刷新 Token 不需要 repo 调用

	// 先通过登录获取 Refresh Token
	hash, _ := auth.HashPassword("password123")
	existingUser := &model.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: hash,
		Role:         auth.RoleReader,
	}
	mockRepo.On("FindByUsername", "testuser").Return(existingUser, nil)

	loginPair, _ := svc.Login("testuser", "password123")
	require.NotNil(t, loginPair)

	// 使用 Refresh Token 刷新
	newPair, appErr := svc.RefreshToken(loginPair.RefreshToken)

	assert.Nil(t, appErr)
	assert.NotNil(t, newPair)
	assert.NotEmpty(t, newPair.AccessToken)
	assert.NotEmpty(t, newPair.RefreshToken)
}

func TestUserService_RefreshToken_InvalidToken(t *testing.T) {
	// Token 刷新失败：无效的 Token
	svc, _, _ := newTestUserService(t)

	newPair, appErr := svc.RefreshToken("invalid-token-string")

	assert.Nil(t, newPair)
	assert.Equal(t, errcode.ErrRefreshTokenInvalid, appErr)
}

// ==================== 用户登出测试 ====================

func TestUserService_Logout_Success(t *testing.T) {
	// 登出成功：Token JTI 加入 Redis 黑名单
	svc, _, mr := newTestUserService(t)

	appErr := svc.Logout("test-jti-12345", 15*time.Minute)

	assert.Nil(t, appErr)
	// 验证 Redis 中存在黑名单记录
	val, err := mr.Get("token:blacklist:test-jti-12345")
	assert.NoError(t, err)
	assert.Equal(t, "1", val)
}

// ==================== 获取用户资料测试 ====================

func TestUserService_GetProfile_Success(t *testing.T) {
	// 获取资料成功
	svc, mockRepo, _ := newTestUserService(t)

	expectedUser := &model.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     auth.RoleReader,
		Avatar:   "avatar.png",
		Bio:      "Hello World",
	}
	mockRepo.On("GetByID", uint(1)).Return(expectedUser, nil)

	user, appErr := svc.GetProfile(1)

	assert.Nil(t, appErr)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetProfile_NotFound(t *testing.T) {
	// 获取资料失败：用户不存在
	svc, mockRepo, _ := newTestUserService(t)

	mockRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	user, appErr := svc.GetProfile(999)

	assert.Nil(t, user)
	assert.Equal(t, errcode.ErrUserNotFound, appErr)
	mockRepo.AssertExpectations(t)
}

// ==================== 更新用户资料测试 ====================

func TestUserService_UpdateProfile_Success(t *testing.T) {
	// 更新资料成功
	svc, mockRepo, _ := newTestUserService(t)

	existingUser := &model.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     auth.RoleReader,
	}
	mockRepo.On("GetByID", uint(1)).Return(existingUser, nil)
	mockRepo.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	user, appErr := svc.UpdateProfile(1, "new-avatar.png", "新的简介")

	assert.Nil(t, appErr)
	assert.NotNil(t, user)
	assert.Equal(t, "new-avatar.png", user.Avatar)
	assert.Equal(t, "新的简介", user.Bio)
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateProfile_UserNotFound(t *testing.T) {
	// 更新资料失败：用户不存在
	svc, mockRepo, _ := newTestUserService(t)

	mockRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

	user, appErr := svc.UpdateProfile(999, "avatar.png", "bio")

	assert.Nil(t, user)
	assert.Equal(t, errcode.ErrUserNotFound, appErr)
	mockRepo.AssertExpectations(t)
}
