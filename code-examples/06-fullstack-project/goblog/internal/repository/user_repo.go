// Package repository 提供 GoBlog 的数据访问层
// 使用 Repository 模式封装 GORM 数据库操作，接口与实现分离
package repository

import (
	"guide-go/goblog/internal/model"

	"gorm.io/gorm"
)

// UserRepo 用户数据访问接口
type UserRepo interface {
	Create(user *model.User) error
	GetByID(id uint) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(id uint) error
	List(offset, limit int) ([]model.User, int64, error)
}

// userRepo 用户数据访问实现
type userRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户 Repository 实例
func NewUserRepo(db *gorm.DB) UserRepo {
	return &userRepo{db: db}
}

// Create 创建用户
func (r *userRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// GetByID 根据 ID 查询用户
func (r *userRepo) GetByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查询用户
func (r *userRepo) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查询用户
func (r *userRepo) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户信息
func (r *userRepo) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Delete 软删除用户
func (r *userRepo) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

// List 分页查询用户列表
func (r *userRepo) List(offset, limit int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// 先查询总数
	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := r.db.Offset(offset).Limit(limit).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
