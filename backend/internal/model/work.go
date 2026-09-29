package model

import (
	"time"

	"gorm.io/gorm"
)

// Work 作品集模型（对应 docs/database.md 3.10 works 表）
type Work struct {
	ID          uint           `gorm:"primarykey" json:"id"`                                                   // 主键
	Name        string         `gorm:"size:100;not null" json:"name"`                                          // 作品名
	Slug        string         `gorm:"size:100;uniqueIndex" json:"slug"`                                       // SEO 友好别名（唯一）
	Description string         `gorm:"type:text" json:"description"`                                           // 作品描述
	Cover       string         `gorm:"size:500" json:"cover"`                                                  // 封面图 URL
	DemoURL     string         `gorm:"size:500" json:"demo_url"`                                               // 演示链接
	ArticleURL  string         `gorm:"size:500" json:"article_url"`                                            // 文章链接
	RepoURL     string         `gorm:"size:500" json:"repo_url"`                                               // 源码链接
	TechStack   string         `gorm:"type:text" json:"tech_stack"`                                            // 技术栈（JSON 字符串数组）
	Year        string         `gorm:"size:20" json:"year"`                                                    // 年份/时间（如 2025、2024-06）
	IsTop       int            `gorm:"default:0;index:idx_works_status_top_sort,priority:2" json:"is_top"`     // 是否置顶/推荐：1 是 / 0 否
	Status      int            `gorm:"index:idx_works_status_top_sort,priority:1" json:"status"`               // 状态：1 启用 / 0 禁用（服务层默认 1）
	SortOrder   int            `gorm:"default:0;index:idx_works_status_top_sort,priority:3" json:"sort_order"` // 排序权重
	Extra       string         `gorm:"type:text" json:"extra"`                                                 // 扩展字段（JSON 字符串）
	CreatedAt   time.Time      `json:"created_at"`                                                             // 创建时间
	UpdatedAt   time.Time      `json:"updated_at"`                                                             // 更新时间
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`                                                         // 软删除标记
}
