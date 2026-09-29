package dto

// TagRequest 标签创建/更新请求体
type TagRequest struct {
	Name string `json:"name" binding:"required"` // 标签名称（必填）
	Slug string `json:"slug"`                    // 别名（留空自动生成）
}
