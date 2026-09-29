package dto

// FriendRequest 友链创建/更新请求体
type FriendRequest struct {
	Name        string  `json:"name" binding:"required"` // 友链名称（必填）
	URL         string  `json:"url" binding:"required"`  // 友链地址（必填）
	Icon        *string `json:"icon"`                    // 友链图标（*string：nil 不修改，空串清除）
	Description *string `json:"description"`             // 描述（*string：nil 不修改，空串清除）
	Status      *int    `json:"status"`                  // 状态：1 启用 / 0 禁用
	SortOrder   *int    `json:"sort_order"`              // 排序权重
}
