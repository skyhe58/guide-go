// Package model 定义 GoBlog 的 GORM 数据模型
// 所有模型对应数据库表结构，使用 GORM 标签定义字段约束和关联关系
package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
// 对应 users 表，支持软删除
// 角色：admin（管理员）、author（作者）、reader（读者）
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string         `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Role         string         `gorm:"size:20;default:reader" json:"role"`
	Avatar       string         `gorm:"size:255" json:"avatar"`
	Bio          string         `gorm:"type:text" json:"bio"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Articles     []Article      `gorm:"foreignKey:AuthorID" json:"articles,omitempty"`
	Comments     []Comment      `gorm:"foreignKey:UserID" json:"comments,omitempty"`
}

// TableName 指定用户表名
func (User) TableName() string {
	return "users"
}
