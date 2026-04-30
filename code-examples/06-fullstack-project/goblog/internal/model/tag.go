package model

import "time"

// Tag 标签模型
// 对应 tags 表，通过 article_tags 关联表实现与 Article 的多对多关系
type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Slug      string    `gorm:"uniqueIndex;size:50;not null" json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Articles  []Article `gorm:"many2many:article_tags" json:"articles,omitempty"`
}

// TableName 指定标签表名
func (Tag) TableName() string {
	return "tags"
}
