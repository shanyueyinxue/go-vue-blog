package model

import (
	"time"

	"gorm.io/gorm"
)

// Category 分类模型（对应 docs/database.md 3.2 categories 表）
type Category struct {
	ID          uint           `gorm:"primarykey" json:"id"`             // 主键
	Name        string         `gorm:"size:100;not null" json:"name"`    // 分类名称
	Slug        string         `gorm:"size:100;uniqueIndex" json:"slug"` // SEO 友好别名（唯一）
	Description string         `gorm:"size:255" json:"description"`      // 分类描述
	Status      int            `json:"status"`                           // 状态：1 启用 / 0 禁用（服务层默认 1）
	SortOrder   int            `gorm:"default:0" json:"sort_order"`      // 排序权重
	CreatedAt   time.Time      `json:"created_at"`                       // 创建时间
	UpdatedAt   time.Time      `json:"updated_at"`                       // 更新时间
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`                   // 软删除标记

	PostCount int64 `gorm:"-" json:"post_count"` // 文章数量（聚合查询填充，非数据库字段）
}
