package dto

import "blog/pkg/response"

// FileQuery 文件列表查询参数
type FileQuery struct {
	response.PageRequest
	MimeType string `form:"mime_type"` // 按 MIME 类型筛选（前缀匹配，如 image/）
	Keyword  string `form:"keyword"`   // 按原始文件名模糊搜索
}
