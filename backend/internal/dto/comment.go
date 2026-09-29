package dto

import (
	"time"

	"blog/pkg/response"
)

// GetCommentsRequest 获取文章评论查询参数
type GetCommentsRequest struct {
	response.PageRequest
}

// PublicComment 公开评论对象（不含 email/phone/ip/user_agent 等隐私字段，见 docs/api.md 6.1）。
// 公开接口 GET /api/posts/:slug/comments 与 POST /api/comments 的响应经
// service 层转换为该 DTO 后再序列化，避免模型上的隐私字段泄露。
type PublicComment struct {
	ID         uint             `json:"id"`                // 主键
	PostID     uint             `json:"post_id"`           // 所属文章 ID
	ParentID   uint             `json:"parent_id"`         // 父评论 ID，0 表示顶层评论
	Nickname   string           `json:"nickname"`          // 访客昵称（入库前已做 XSS 过滤）
	Content    string           `json:"content"`           // 评论内容（入库前已做 XSS 过滤）
	Status     string           `json:"status"`            // 审核状态
	IsAuthor   int              `json:"is_author"`         // 是否博主评论：1 是 / 0 否
	ReplyCount int              `json:"reply_count"`       // 回复数（冗余字段）
	Likes      int              `json:"likes"`             // 点赞数
	CreatedAt  time.Time        `json:"created_at"`        // 创建时间
	Replies    []*PublicComment `json:"replies,omitempty"` // 子评论列表（楼中楼）
}

// CreateCommentRequest 提交评论请求体
type CreateCommentRequest struct {
	PostID       uint   `json:"post_id"`                     // 文章 ID（与 slug 二选一）
	Slug         string `json:"slug"`                        // 文章别名（与 post_id 二选一）
	ParentID     uint   `json:"parent_id"`                   // 父评论 ID，0 为顶层评论
	Nickname     string `json:"nickname" binding:"required"` // 昵称（必填）
	Email        string `json:"email"`                       // 邮箱（必填，用于接收验证码与通知）
	Phone        string `json:"phone"`                       // 手机号（可选）
	Content      string `json:"content" binding:"required"`  // 评论内容（必填）
	CaptchaToken string `json:"captcha_token"`               // 验证码 token（发送验证码时返回）
	CaptchaCode  string `json:"captcha_code"`                // 验证码内容
}

// AdminCommentQuery 管理端评论列表查询参数
type AdminCommentQuery struct {
	response.PageRequest
	Status  string `form:"status"`  // 审核状态筛选
	PostID  uint   `form:"post_id"` // 文章筛选
	Keyword string `form:"keyword"` // 昵称/内容关键字
}

// CommentAdminUpdateRequest 管理端修改评论请求体（仅允许修改 status / is_author，其余字段一律忽略）
type CommentAdminUpdateRequest struct {
	Status   *string `json:"status"`    // 审核状态：pending / approved / rejected（缺省不修改）
	IsAuthor *int    `json:"is_author"` // 是否博主评论：1 是 / 0 否（缺省不修改）
}
