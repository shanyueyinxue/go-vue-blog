package model

// PostTag 文章-标签关联模型（对应 docs/database.md 3.5 post_tags 表）
// 由 GORM 多对多关系自动维护，联合主键 (post_id, tag_id)
type PostTag struct {
	PostID uint `gorm:"primaryKey" json:"post_id"` // 外键 → posts.id
	TagID  uint `gorm:"primaryKey" json:"tag_id"`  // 外键 → tags.id
}
