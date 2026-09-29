package model

import (
	"time"

	"gorm.io/gorm"
)

// Friend 友链模型（对应 docs/database.md 3.8 friends 表）
type Friend struct {
	ID          uint           `gorm:"primarykey" json:"id"`          // 主键
	Name        string         `gorm:"size:100;not null" json:"name"` // 友链名称
	URL         string         `gorm:"size:500;not null" json:"url"`  // 友链地址
	Icon        string         `gorm:"size:500" json:"icon"`          // 友链图标（图片 URL 或 Font Awesome 类名）
	Description string         `gorm:"size:500" json:"description"`   // 友链描述
	Status      int            `json:"status"`                        // 状态：1 启用 / 0 禁用（服务层默认 1）
	SortOrder   int            `gorm:"default:0" json:"sort_order"`   // 排序权重
	CreatedAt   time.Time      `json:"created_at"`                    // 创建时间
	UpdatedAt   time.Time      `json:"updated_at"`                    // 更新时间
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`                // 软删除标记
}
