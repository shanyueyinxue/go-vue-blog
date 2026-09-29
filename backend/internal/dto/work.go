package dto

// WorkRequest 作品创建/更新请求体
// 创建时 name 必填，slug 留空自动生成；更新时字段缺省表示不修改
// （*string 支持清空：nil 不修改，空字符串清除；*int 同理）
type WorkRequest struct {
	Name        string         `json:"name" binding:"required"` // 作品名（必填）
	Slug        string         `json:"slug"`                    // 别名（留空自动生成，重复时追加后缀）
	Description *string        `json:"description"`             // 作品描述
	Cover       *string        `json:"cover"`                   // 封面图 URL
	DemoURL     *string        `json:"demo_url"`                // 演示链接
	ArticleURL  *string        `json:"article_url"`             // 文章链接
	RepoURL     *string        `json:"repo_url"`                // 源码链接
	TechStack   []string       `json:"tech_stack"`              // 技术栈数组（入库序列化为 JSON 字符串）
	Year        *string        `json:"year"`                    // 年份/时间
	IsTop       *int           `json:"is_top"`                  // 是否置顶：1 / 0
	Status      *int           `json:"status"`                  // 状态：1 启用 / 0 禁用
	SortOrder   *int           `json:"sort_order"`              // 排序权重
	Extra       map[string]any `json:"extra"`                   // 扩展字段（JSON 对象）
}
