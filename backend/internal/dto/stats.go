package dto

import "blog/internal/model"

// StatsSummary 仪表盘统计响应体
type StatsSummary struct {
	PostCount           int64        `json:"post_count"`            // 文章总数（含草稿）
	PublishedCount      int64        `json:"published_count"`       // 已发布文章数
	DraftCount          int64        `json:"draft_count"`           // 草稿数
	CategoryCount       int64        `json:"category_count"`        // 分类数
	TagCount            int64        `json:"tag_count"`             // 标签数
	CommentCount        int64        `json:"comment_count"`         // 评论总数
	PendingCommentCount int64        `json:"pending_comment_count"` // 待审核评论数
	TotalViews          int64        `json:"total_views"`           // 总浏览量
	RecentPosts         []model.Post `json:"recent_posts"`          // 最近 5 篇已发布文章
}
