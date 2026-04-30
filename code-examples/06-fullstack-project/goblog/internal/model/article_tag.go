package model

// ArticleTag 文章-标签多对多关联表模型
// 使用 ArticleID 和 TagID 作为复合主键
type ArticleTag struct {
	ArticleID uint `gorm:"primaryKey" json:"article_id"`
	TagID     uint `gorm:"primaryKey" json:"tag_id"`
}

// TableName 指定关联表名
func (ArticleTag) TableName() string {
	return "article_tags"
}
