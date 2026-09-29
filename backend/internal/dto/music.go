package dto

import "blog/pkg/response"

// MusicRequest 音乐创建/更新请求体
// 创建时 name、file_id 必填；更新时字段缺省表示不修改（*string 支持清空）
type MusicRequest struct {
	Name      string  `json:"name" binding:"required"` // 歌名（必填）
	Artist    *string `json:"artist"`                  // 歌手（*string 支持清空）
	Cover     *string `json:"cover"`                   // 封面图 URL（*string 支持清空）
	FileID    *uint   `json:"file_id"`                 // 关联文件 ID（创建必填；更新时提供且变更会同步清理旧音频文件）
	SortOrder *int    `json:"sort_order"`              // 排序权重
	Status    *int    `json:"status"`                  // 状态：1 启用 / 0 禁用
	Lrc       *string `json:"lrc"`                     // 歌词文本（*string 支持清空）
}

// MusicQuery 管理端音乐列表查询参数
type MusicQuery struct {
	response.PageRequest
	Keyword string `form:"keyword"` // 按歌名/歌手模糊搜索
}

// PublicMusic 公开音乐对象（不含 file_id 等内部字段）
type PublicMusic struct {
	ID     uint   `json:"id"`     // 主键
	Name   string `json:"name"`   // 歌名
	Artist string `json:"artist"` // 歌手
	Cover  string `json:"cover"`  // 封面图 URL
	URL    string `json:"url"`    // 音频完整访问 URL（浏览器直连存储）
	Lrc    string `json:"lrc"`    // 歌词文本
}
