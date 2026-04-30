// Package service 提供 GoBlog 的业务逻辑层
// Service 层调用 Repository 层进行数据操作，处理业务规则和错误
package service

import (
	"context"
	"time"

	"guide-go/goblog/internal/auth"
	"guide-go/goblog/internal/cache"
	"guide-go/goblog/internal/errcode"
	"guide-go/goblog/internal/model"
	"guide-go/goblog/internal/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// UserService 用户业务服务接口
type UserService interface {
	Register(username, email, password string) (*model.User, *errcode.AppError)
	Login(username, password string) (*auth.TokenPair, *errcode.AppError)
	RefreshToken(refreshToken string) (*auth.TokenPair, *errcode.AppError)
	Logout(jti string, expiration time.Duration) *errcode.AppError
	GetProfile(userID uint) (*model.User, *errcode.AppError)
	UpdateProfile(userID uint, avatar, bio string) (*model.User, *errcode.AppError)
}

// userService 用户业务服务实现
type userService struct {
	userRepo repository.UserRepo
	jwtSvc   *auth.JWTService
	rdb      *redis.Client
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepo, jwtSvc *auth.JWTService, rdb *redis.Client) UserService {
	return &userService{
		userRepo: userRepo,
		jwtSvc:   jwtSvc,
		rdb:      rdb,
	}
}

// Register 用户注册
// 检查用户名和邮箱唯一性，使用 bcrypt 加密密码
func (s *userService) Register(username, email, password string) (*model.User, *errcode.AppError) {
	// 检查用户名是否已存在
	if _, err := s.userRepo.FindByUsername(username); err == nil {
		return nil, errcode.ErrUsernameExists
	}

	// 检查邮箱是否已注册
	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return nil, errcode.ErrEmailExists
	}

	// 密码加密
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	// 创建用户
	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		Role:         auth.RoleReader, // 默认角色为读者
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errcode.ErrInternal
	}

	return user, nil
}

// Login 用户登录
// 验证密码，返回 JWT 双令牌
func (s *userService) Login(username, password string) (*auth.TokenPair, *errcode.AppError) {
	// 查找用户
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrInternal
	}

	// 验证密码
	if !auth.CheckPassword(password, user.PasswordHash) {
		return nil, errcode.ErrPasswordWrong
	}

	// 签发 JWT 双令牌
	tokenPair, err := s.jwtSvc.GenerateTokenPair(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	return tokenPair, nil
}

// RefreshToken 刷新 Access Token
func (s *userService) RefreshToken(refreshToken string) (*auth.TokenPair, *errcode.AppError) {
	tokenPair, err := s.jwtSvc.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, errcode.ErrRefreshTokenInvalid
	}
	return tokenPair, nil
}

// Logout 用户登出
// 将当前 Token 的 JTI 加入 Redis 黑名单
func (s *userService) Logout(jti string, expiration time.Duration) *errcode.AppError {
	key := cache.TokenBlacklistKey(jti)
	if err := s.rdb.Set(context.Background(), key, "1", expiration).Err(); err != nil {
		return errcode.ErrInternal
	}
	return nil
}

// GetProfile 获取用户资料
func (s *userService) GetProfile(userID uint) (*model.User, *errcode.AppError) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrInternal
	}
	return user, nil
}

// UpdateProfile 更新用户资料（头像和简介）
func (s *userService) UpdateProfile(userID uint, avatar, bio string) (*model.User, *errcode.AppError) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errcode.ErrUserNotFound
		}
		return nil, errcode.ErrInternal
	}

	// 更新字段
	if avatar != "" {
		user.Avatar = avatar
	}
	if bio != "" {
		user.Bio = bio
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, errcode.ErrInternal
	}

	return user, nil
}
