package dto

// CategoryRequest 分类创建/更新请求体
type CategoryRequest struct {
	Name        string  `json:"name" binding:"required"` // 分类名称（必填）
	Slug        string  `json:"slug"`                    // 别名（留空自动生成）
	Description *string `json:"description"`             // 描述（*string：nil 不修改，空串清除）
	Status      *int    `json:"status"`                  // 状态：1 启用 / 0 禁用
	SortOrder   *int    `json:"sort_order"`              // 排序权重
}
