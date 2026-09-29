package model

import (
	"time"

	"gorm.io/gorm"
)

// 评论状态常量
const (
	CommentStatusPending  = "pending"  // 待审核
	CommentStatusApproved = "approved" // 已通过
	CommentStatusRejected = "rejected" // 已拒绝
)

// Comment 评论模型（对应 docs/database.md 3.6 comments 表）
type Comment struct {
	ID         uint           `gorm:"primarykey" json:"id"`                  // 主键
	PostID     uint           `gorm:"index" json:"post_id"`                  // 所属文章 ID
	ParentID   uint           `gorm:"default:0;index" json:"parent_id"`      // 父评论 ID，0 表示顶层评论（支持楼中楼）
	Nickname   string         `gorm:"size:50;not null" json:"nickname"`      // 访客昵称（入库前已做 XSS 过滤）
	Email      string         `gorm:"size:100" json:"email"`                 // 邮箱（公开接口不回传）
	Phone      string         `gorm:"size:20" json:"phone"`                  // 手机号（公开接口不回传）
	Content    string         `gorm:"type:text;not null" json:"content"`     // 评论内容（入库前已做 XSS 过滤）
	Status     string         `gorm:"size:20;default:pending" json:"status"` // 审核状态
	IsAuthor   int            `gorm:"default:0" json:"is_author"`            // 是否博主评论：1 是 / 0 否
	ReplyCount int            `gorm:"default:0" json:"reply_count"`          // 回复数（冗余字段）
	Likes      int            `gorm:"default:0" json:"likes"`                // 点赞数
	IP         string         `gorm:"size:50" json:"ip"`                     // 评论者 IP（公开接口不回传）
	UserAgent  string         `gorm:"size:255" json:"user_agent"`            // 用户代理（公开接口不回传）
	ReviewedAt *time.Time     `json:"reviewed_at"`                           // 审核时间（可空）
	CreatedAt  time.Time      `json:"created_at"`                            // 创建时间
	UpdatedAt  time.Time      `json:"updated_at"`                            // 更新时间
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`                        // 软删除标记

	// 子评论列表（楼中楼，查询时按 ParentID 组装）
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`

	// 管理端聚合展示字段（非数据库字段）
	PostTitle      string `gorm:"-" json:"post_title,omitempty"`
	ParentNickname string `gorm:"-" json:"parent_nickname,omitempty"`
}
