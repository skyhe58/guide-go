package model

import (
	"time"

	"gorm.io/gorm"
)

// Article 文章模型
// 对应 articles 表，支持软删除
// 状态：draft（草稿）、published（已发布）、archived（已归档）
// 通过 article_tags 关联表实现与 Tag 的多对多关系
type Article struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	AuthorID    uint           `gorm:"not null;index" json:"author_id"`
	Author      User           `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Title       string         `gorm:"size:200;not null" json:"title"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	Slug        string         `gorm:"uniqueIndex;size:200;not null" json:"slug"`
	Status      string         `gorm:"size:20;default:draft" json:"status"`
	ViewCount   int            `gorm:"default:0" json:"view_count"`
	PublishedAt *time.Time     `json:"published_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Tags        []Tag          `gorm:"many2many:article_tags" json:"tags,omitempty"`
	Comments    []Comment      `gorm:"foreignKey:ArticleID" json:"comments,omitempty"`
}

// TableName 指定文章表名
func (Article) TableName() string {
	return "articles"
}
