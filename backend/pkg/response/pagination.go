package response

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// 分页参数默认值/上限
const (
	DefaultPage     = 1   // 默认页码
	DefaultPageSize = 10  // 默认每页条数
	MaxPageSize     = 100 // 每页条数上限
)

// PageRequest 分页查询参数（通过 query string 传递）
type PageRequest struct {
	Page     int `form:"page"`     // 页码，从 1 开始
	PageSize int `form:"pageSize"` // 每页条数
}

// Normalize 规范化分页参数：非法值回退到默认值，pageSize 不超过上限
func (p *PageRequest) Normalize() {
	if p.Page <= 0 {
		p.Page = DefaultPage
	}
	if p.PageSize <= 0 {
		p.PageSize = DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}
}

// Offset 计算 SQL 偏移量
func (p *PageRequest) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// PageResult 分页响应结构（docs/api.md 1.5）
type PageResult struct {
	List     any   `json:"list"`     // 当前页数据
	Total    int64 `json:"total"`    // 总条数
	Page     int   `json:"page"`     // 当前页码
	PageSize int   `json:"pageSize"` // 每页条数
}

// SuccessPage 返回分页数据
func SuccessPage(c *gin.Context, list any, total int64, page, pageSize int) {
	Success(c, PageResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetUintParam 解析路径或查询参数中的 uint 类型 ID，失败返回 0
func GetUintParam(c *gin.Context, key string) uint {
	val := c.Param(key)
	if val == "" {
		val = c.Query(key)
	}
	id, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0
	}
	return uint(id)
}
