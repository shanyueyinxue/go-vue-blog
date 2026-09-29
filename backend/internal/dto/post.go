package dto

import (
	"time"

	"blog/pkg/response"
)

// PostListRequest 文章列表查询参数（公开与管理接口共用）
type PostListRequest struct {
	response.PageRequest
	CategoryID uint   `form:"category_id"` // 按分类筛选
	TagID      uint   `form:"tag_id"`      // 按标签筛选
	Keyword    string `form:"keyword"`     // 标题/摘要关键字搜索
	Status     string `form:"status"`      // 管理接口：按状态筛选
	Sort       string `form:"sort"`        // 排序：latest / views / oldest
	TopFirst   string `form:"top_first"`   // 是否置顶优先（默认 true）
	IsTop      string `form:"is_top"`      // 管理接口：置顶筛选（1 仅置顶 / 0 仅非置顶，空为全部）
}

// PostUpsertRequest 文章创建/更新请求体
// 创建时 title、content 必填；更新时字段缺省表示不修改
type PostUpsertRequest struct {
	Title       string     `json:"title"`        // 标题（创建必填）
	Slug        string     `json:"slug"`         // 别名（留空自动生成）
	Content     string     `json:"content"`      // Markdown 正文
	Excerpt     *string    `json:"excerpt"`      // 摘要
	CoverImage  *string    `json:"cover_image"`  // 封面图
	CategoryID  *uint      `json:"category_id"`  // 分类 ID（0 表示清空分类）
	TagIDs      []uint     `json:"tag_ids"`      // 标签 ID 数组
	Status      *string    `json:"status"`       // draft / published
	IsTop       *int       `json:"is_top"`       // 是否置顶
	PublishedAt *time.Time `json:"published_at"` // 发布时间
}

// ArchivePost 归档文章摘要
type ArchivePost struct {
	ID          uint       `json:"id"`           // 文章 ID
	Title       string     `json:"title"`        // 标题
	Slug        string     `json:"slug"`         // 别名
	PublishedAt *time.Time `json:"published_at"` // 发布时间
}

// ArchiveItem 归档条目
type ArchiveItem struct {
	Year  int           `json:"year"`  // 年份
	Month int           `json:"month"` // 月份
	Count int64         `json:"count"` // 该月文章数
	Posts []ArchivePost `json:"posts"` // 文章列表
}
