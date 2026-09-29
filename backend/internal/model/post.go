package model

import (
	"time"

	"gorm.io/gorm"
)

// Post 文章模型（对应 docs/database.md 3.4 posts 表）
type Post struct {
	ID           uint           `gorm:"primarykey" json:"id"`                                                           // 主键
	Title        string         `gorm:"size:255;not null" json:"title"`                                                 // 标题
	Slug         string         `gorm:"size:255;uniqueIndex" json:"slug"`                                               // SEO 友好 URL 别名（唯一）
	Content      string         `gorm:"type:text" json:"content"`                                                       // Markdown 正文
	Excerpt      string         `gorm:"size:500" json:"excerpt"`                                                        // 摘要（列表页展示）
	CoverImage   string         `gorm:"size:500" json:"cover_image"`                                                    // 封面图地址
	CategoryID   *uint          `gorm:"index" json:"category_id"`                                                       // 分类 ID（可空）
	Status       string         `gorm:"size:20;default:draft;index:idx_posts_status_top_time,priority:1" json:"status"` // 状态：draft 草稿 / published 发布
	IsTop        int            `gorm:"default:0;index:idx_posts_status_top_time,priority:2" json:"is_top"`             // 是否置顶：1 置顶 / 0 否
	ViewCount    int            `gorm:"default:0" json:"view_count"`                                                    // 浏览量
	CommentCount int            `gorm:"default:0" json:"comment_count"`                                                 // 评论数（冗余字段）
	PublishedAt  *time.Time     `gorm:"index:idx_posts_status_top_time,priority:3" json:"published_at"`                 // 发布时间（可空）
	CreatedAt    time.Time      `json:"created_at"`                                                                     // 创建时间
	UpdatedAt    time.Time      `json:"updated_at"`                                                                     // 更新时间
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`                                                                 // 软删除标记

	// 关联数据（Preload 填充）
	Category *Category `gorm:"foreignKey:CategoryID" json:"category"` // 所属分类
	Tags     []Tag     `gorm:"many2many:post_tags;" json:"tags"`      // 标签列表（多对多）
}
