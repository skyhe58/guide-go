package model

import (
	"time"

	"gorm.io/gorm"
)

// Comment 评论模型
// 对应 comments 表，支持软删除
// 关联用户和文章
type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ArticleID uint           `gorm:"not null;index" json:"article_id"`
	Article   Article        `gorm:"foreignKey:ArticleID" json:"-"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定评论表名
func (Comment) TableName() string {
	return "comments"
}
