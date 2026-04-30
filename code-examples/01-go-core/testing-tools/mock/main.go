// Go 1.22+ | 验证日期：2025-01-01
// Mock 接口示例
// 演示 Go 中基于接口的 Mock 设计模式
// 定义接口抽象依赖，Service 层依赖接口而非具体实现
// 体现 Go "Accept Interfaces, Return Structs" 的设计哲学
package mock

import (
	"errors"
	"fmt"
)

// ============================================================
// 领域模型
// ============================================================

// User 用户模型
type User struct {
	ID    int
	Name  string
	Email string
}

// ============================================================
// 接口定义（抽象依赖）
// ============================================================

// UserRepository 用户数据访问接口
// Service 层依赖此接口而非具体实现，便于测试时替换为 Mock
type UserRepository interface {
	GetByID(id int) (*User, error)
	Create(user *User) error
	Update(user *User) error
}

// ============================================================
// 错误定义
// ============================================================

// ErrNotFound 用户不存在错误
var ErrNotFound = errors.New("用户不存在")

// ErrDuplicateEmail 邮箱重复错误
var ErrDuplicateEmail = errors.New("邮箱已被注册")

// ============================================================
// Service 层（业务逻辑）
// ============================================================

// UserService 用户业务逻辑层
type UserService struct {
	repo UserRepository // 依赖接口而非具体实现
}

// NewUserService 创建 UserService 实例
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetUser 获取用户信息
func (s *UserService) GetUser(id int) (*User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("无效的用户 ID: %d", id)
	}
	return s.repo.GetByID(id)
}

// CreateUser 创建新用户
func (s *UserService) CreateUser(name, email string) (*User, error) {
	if name == "" {
		return nil, fmt.Errorf("用户名不能为空")
	}
	if email == "" {
		return nil, fmt.Errorf("邮箱不能为空")
	}

	user := &User{
		Name:  name,
		Email: email,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

// UpdateUserName 更新用户名称
func (s *UserService) UpdateUserName(id int, newName string) error {
	if newName == "" {
		return fmt.Errorf("用户名不能为空")
	}

	user, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}

	user.Name = newName
	return s.repo.Update(user)
}
