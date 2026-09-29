package dto

// SocialLink 社交链接项（social_links 为 JSON 数组，顺序即前台显示顺序）
type SocialLink struct {
	Name string `json:"name"` // 名称（前台直接展示）
	URL  string `json:"url"`  // 链接地址
}

// ConfigUpsertRequest 站点配置创建/更新请求体
// 字段全部可选（除创建时 version 必填）；social_links 为 JSON 数组、extra 为 JSON 对象；
// 文本字段为 *string：nil 表示不修改，空字符串表示清除
type ConfigUpsertRequest struct {
	Version            string         `json:"version"`              // 版本号（创建必填）
	Title              *string        `json:"title"`                // 站点标题
	Subtitle           *string        `json:"subtitle"`             // 副标题
	SideName           *string        `json:"side_name"`            // 侧栏卡片名称（空串表示清除，回退站点标题）
	Description        *string        `json:"description"`          // 描述
	Avatar             *string        `json:"avatar"`               // 博客头像 URL
	ICP                *string        `json:"icp"`                  // ICP 备案号
	PoliceICP          *string        `json:"police_icp"`           // 警察 ICP 备案号
	Theme              *string        `json:"theme"`                // 主题（light/dark，空表示跟随系统）
	AboutContent       *string        `json:"about_content"`        // About 页面内容
	HomeImages         []string       `json:"home_images"`          // 首页随机背景图 URL 列表
	SocialLinks        []SocialLink     `json:"social_links"`         // 社交链接（JSON 数组，顺序即显示顺序）
	CommentNeedReview  *int           `json:"comment_need_review"`  // 评论是否需要审核
	CommentNeedCaptcha *int           `json:"comment_need_captcha"` // 评论是否需要邮箱验证码
	CustomCSSVars      *string        `json:"custom_css_vars"`      // 自定义 CSS 变量（平坦变量文本，空字符串表示清除）
	CustomCSSVarsDark  *string        `json:"custom_css_vars_dark"` // 黑暗主题专属的自定义 CSS 变量（平坦变量文本，空字符串表示清除）
	CustomCSSEnabled   *int           `json:"custom_css_enabled"`   // 是否启用自定义 CSS 变量：1 是 / 0 否
	Extra              map[string]any `json:"extra"`                // 扩展字段（JSON 对象）
	CreatedBy          string         `json:"created_by"`           // 创建者
}

// PublicConfig 前台生效配置响应体
type PublicConfig struct {
	Title              string         `json:"title"`
	Subtitle           string         `json:"subtitle"`
	SideName           string         `json:"side_name"`
	Description        string         `json:"description"`
	Avatar             string         `json:"avatar"`
	ICP                string         `json:"icp"`
	PoliceICP          string         `json:"police_icp"`
	Theme              string         `json:"theme"`
	AboutContent       string         `json:"about_content"`
	HomeImages         []string       `json:"home_images"`
	SocialLinks        []SocialLink     `json:"social_links"`
	CommentNeedReview  int            `json:"comment_need_review"`
	CommentNeedCaptcha int            `json:"comment_need_captcha"`
	CustomCSSVars      string         `json:"custom_css_vars"`
	CustomCSSVarsDark  string         `json:"custom_css_vars_dark"`
	CustomCSSEnabled   int            `json:"custom_css_enabled"`
	Extra              map[string]any `json:"extra"`
}
