package model

import "time"

// SiteConfig 站点配置模型（对应 docs/database.md 3.7 site_configs 表）
// 采用「固定列 + extra JSON 扩展」结构；同一时间仅一条 is_active=1 的记录生效
type SiteConfig struct {
	ID                 uint      `gorm:"primarykey" json:"id"`                  // 主键
	Version            string    `gorm:"size:50;not null" json:"version"`       // 版本号（如 1.0.0）
	IsActive           int       `gorm:"default:0;index" json:"is_active"`      // 是否启用：1 生效 / 0 否
	Title              string    `gorm:"size:255" json:"title"`                 // 站点标题
	Subtitle           string    `gorm:"size:255" json:"subtitle"`              // 副标题
	SideName           string    `gorm:"size:255" json:"side_name"`             // 侧栏卡片名称（留空回退站点标题）
	Description        string    `gorm:"size:500" json:"description"`           // 站点描述
	Avatar             string    `gorm:"size:500" json:"avatar"`                // 博客头像 URL
	ICP                string    `gorm:"size:100" json:"icp"`                   // ICP 备案号
	PoliceICP          string    `gorm:"size:100" json:"police_icp"`            // 警察 ICP 备案号
	Theme              string    `gorm:"size:50" json:"theme"`                  // 主题标识
	AboutContent       string    `gorm:"type:text" json:"about_content"`        // About 页面内容
	HomeImages         string    `gorm:"type:text" json:"home_images"`          // 首页随机背景图（JSON 字符串数组）
	SocialLinks        string    `gorm:"type:text" json:"social_links"`         // 社交链接（JSON 字符串）
	CommentNeedReview  int       `json:"comment_need_review"`                   // 评论是否需要审核：1 是 / 0 否（服务层默认 1）
	CommentNeedCaptcha int       `json:"comment_need_captcha"`                  // 评论是否需要邮箱验证码：1 是 / 0 否（服务层默认 1）
	CustomCSSVars      string    `gorm:"type:text" json:"custom_css_vars"`      // 自定义 CSS 变量（每行一个 --name: value;，仅支持平坦变量）
	CustomCSSVarsDark  string    `gorm:"type:text" json:"custom_css_vars_dark"` // 黑暗主题专属的自定义 CSS 变量（仅 data-theme=dark 时生效）
	CustomCSSEnabled   int       `gorm:"default:0" json:"custom_css_enabled"`   // 是否启用自定义 CSS 变量：1 是 / 0 否
	Extra              string    `gorm:"type:text" json:"extra"`                // 扩展字段（JSON 字符串）
	CreatedBy          string    `gorm:"size:50" json:"created_by"`             // 创建者
	CreatedAt          time.Time `json:"created_at"`                            // 创建时间
	UpdatedAt          time.Time `json:"updated_at"`                            // 更新时间
}
