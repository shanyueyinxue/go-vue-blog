package model

import "time"

// Tag 标签模型（对应 docs/database.md 3.3 tags 表）
// 轻量元数据表，删除采用物理删除（无需软删除恢复，见 database.md 1.概述）
type Tag struct {
	ID        uint      `gorm:"primarykey" json:"id"`             // 主键
	Name      string    `gorm:"size:50;not null" json:"name"`     // 标签名称
	Slug      string    `gorm:"size:100;uniqueIndex" json:"slug"` // SEO 友好别名（唯一）
	CreatedAt time.Time `json:"created_at"`                       // 创建时间
	UpdatedAt time.Time `json:"updated_at"`                       // 更新时间

	PostCount int64 `gorm:"-" json:"post_count"` // 文章数量（聚合查询填充，非数据库字段）
}
