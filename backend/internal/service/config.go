package service

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"blog/internal/dto"
	"blog/internal/model"
	"blog/pkg/response"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 站点配置缓存 key（内存缓存，配置更新/激活时失效）
const siteConfigCacheKey = "site_config:active"

// ConfigService 站点配置相关业务
type ConfigService struct {
	*Service
}

// NewConfigService 创建站点配置服务
func NewConfigService(base *Service) *ConfigService {
	return &ConfigService{Service: base}
}

// ActiveConfig 获取当前生效的站点配置（带内存缓存，5 分钟过期）
func (s *ConfigService) ActiveConfig() *model.SiteConfig {
	// 尝试读缓存（缓存内容为 JSON 序列化的配置）
	if cached, err := s.Cache().GetString(siteConfigCacheKey); err == nil && cached != "" {
		var cfg model.SiteConfig
		if json.Unmarshal([]byte(cached), &cfg) == nil {
			return &cfg
		}
	}

	// 缓存未命中，查询数据库（同一时间仅一条 is_active=1）
	var cfg model.SiteConfig
	if err := s.DB().Where("is_active = ?", 1).First(&cfg).Error; err != nil {
		return nil
	}

	// 写入缓存（TTL 5 分钟；配置变更时会主动失效）
	if data, err := json.Marshal(cfg); err == nil {
		_ = s.Cache().Set(siteConfigCacheKey, string(data), 5*time.Minute)
	}
	return &cfg
}

// invalidateCache 使站点配置缓存失效
func (s *ConfigService) invalidateCache() {
	_ = s.Cache().Delete(siteConfigCacheKey)
}

// normalizeSocialLinks 解析社交链接 JSON，兼容两种格式并统一返回数组：
//   - 新格式（推荐）：[{"name":"github","url":"https://..."}]，数组顺序即前台显示顺序
//   - 旧格式：{"github":"https://..."}，按键排序转换（与 Go map 序列化行为一致，顺序确定）
func normalizeSocialLinks(raw string) []dto.SocialLink {
	out := make([]dto.SocialLink, 0)
	if raw == "" {
		return out
	}
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return out
	}
	switch v := value.(type) {
	case []any:
		for _, item := range v {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			name, _ := m["name"].(string)
			url := ""
			switch u := m["url"].(type) {
			case string:
				url = u
			default:
				url = fmt.Sprint(u)
			}
			if name != "" || url != "" {
				out = append(out, dto.SocialLink{Name: name, URL: url})
			}
		}
	case map[string]any:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			out = append(out, dto.SocialLink{Name: k, URL: fmt.Sprint(v[k])})
		}
	}
	return out
}

// PublicConfig 前台生效配置；无生效配置时返回 nil（调用方按空对象处理）
func (s *ConfigService) PublicConfig() *dto.PublicConfig {
	cfg := s.ActiveConfig()
	if cfg == nil {
		return nil
	}

	// 解析 JSON 字段（home_images / social_links / extra）
	homeImages := make([]string, 0)
	if cfg.HomeImages != "" {
		_ = json.Unmarshal([]byte(cfg.HomeImages), &homeImages)
	}
	socialLinks := normalizeSocialLinks(cfg.SocialLinks)
	extra := make(map[string]any)
	if cfg.Extra != "" {
		_ = json.Unmarshal([]byte(cfg.Extra), &extra)
	}

	return &dto.PublicConfig{
		Title:              cfg.Title,
		Subtitle:           cfg.Subtitle,
		SideName:           cfg.SideName,
		Description:        cfg.Description,
		Avatar:             cfg.Avatar,
		ICP:                cfg.ICP,
		PoliceICP:          cfg.PoliceICP,
		Theme:              cfg.Theme,
		AboutContent:       cfg.AboutContent,
		HomeImages:         homeImages,
		SocialLinks:        socialLinks,
		CommentNeedReview:  cfg.CommentNeedReview,
		CommentNeedCaptcha: cfg.CommentNeedCaptcha,
		CustomCSSVars:      cfg.CustomCSSVars,
		CustomCSSVarsDark:  cfg.CustomCSSVarsDark,
		CustomCSSEnabled:   cfg.CustomCSSEnabled,
		Extra:              extra,
	}
}

// ListAdminConfigs 管理端配置列表（分页）
func (s *ConfigService) ListAdminConfigs(req *response.PageRequest) ([]model.SiteConfig, int64, *BizError) {
	var total int64
	if err := s.DB().Model(&model.SiteConfig{}).Count(&total).Error; err != nil {
		return nil, 0, NewServerError("查询配置失败")
	}

	var configs []model.SiteConfig
	if err := s.DB().Order("id DESC").
		Limit(req.PageSize).Offset(req.Offset()).
		Find(&configs).Error; err != nil {
		return nil, 0, NewServerError("查询配置失败")
	}
	return configs, total, nil
}

// GetAdminConfig 管理端单个配置详情
func (s *ConfigService) GetAdminConfig(id uint) (*model.SiteConfig, *BizError) {
	var cfg model.SiteConfig
	if err := s.DB().First(&cfg, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("配置不存在")
		}
		return nil, NewServerError("查询配置失败")
	}
	return &cfg, nil
}

// CreateConfig 新增配置：创建后自动置为激活并取消其它版本的激活状态
func (s *ConfigService) CreateConfig(req *dto.ConfigUpsertRequest) (*model.SiteConfig, *BizError) {
	cfg := model.SiteConfig{
		Version:            req.Version,
		Title:              derefStr(req.Title),
		Subtitle:           derefStr(req.Subtitle),
		SideName:           derefStr(req.SideName),
		Description:        derefStr(req.Description),
		Avatar:             derefStr(req.Avatar),
		ICP:                derefStr(req.ICP),
		PoliceICP:          derefStr(req.PoliceICP),
		Theme:              derefStr(req.Theme),
		AboutContent:       derefStr(req.AboutContent),
		CommentNeedReview:  1,
		CommentNeedCaptcha: 1,
		IsActive:           1,
		CreatedBy:          req.CreatedBy,
	}
	if req.CommentNeedReview != nil {
		cfg.CommentNeedReview = *req.CommentNeedReview
	}
	if req.CommentNeedCaptcha != nil {
		cfg.CommentNeedCaptcha = *req.CommentNeedCaptcha
	}
	if req.CustomCSSEnabled != nil {
		cfg.CustomCSSEnabled = *req.CustomCSSEnabled
	}
	cfg.CustomCSSVars = derefStr(req.CustomCSSVars)
	cfg.CustomCSSVarsDark = derefStr(req.CustomCSSVarsDark)
	// JSON 字段序列化存储
	if req.SocialLinks != nil {
		if data, err := json.Marshal(req.SocialLinks); err == nil {
			cfg.SocialLinks = string(data)
		}
	}
	if req.Extra != nil {
		if data, err := json.Marshal(req.Extra); err == nil {
			cfg.Extra = string(data)
		}
	}
	if req.HomeImages != nil {
		if data, err := json.Marshal(req.HomeImages); err == nil {
			cfg.HomeImages = string(data)
		}
	}

	// 事务：创建配置 + 取消其它版本的激活
	err := s.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.SiteConfig{}).Where("is_active = ?", 1).
			Update("is_active", 0).Error; err != nil {
			return err
		}
		return tx.Create(&cfg).Error
	})
	if err != nil {
		return nil, NewServerError("创建配置失败")
	}
	s.invalidateCache()
	s.Log().Info("创建配置", zap.Uint("id", cfg.ID), zap.String("version", cfg.Version))
	return &cfg, nil
}

// UpdateConfig 更新配置：仅更新提供的字段，不改变 is_active 状态
func (s *ConfigService) UpdateConfig(id uint, req *dto.ConfigUpsertRequest) (*model.SiteConfig, *BizError) {
	var cfg model.SiteConfig
	if err := s.DB().First(&cfg, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("配置不存在")
		}
		return nil, NewServerError("查询配置失败")
	}

	updates := map[string]any{}
	if req.Version != "" {
		updates["version"] = req.Version
	}
	// 文本字段：*string 支持清空（nil 表示不修改，空字符串表示清除）
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Subtitle != nil {
		updates["subtitle"] = *req.Subtitle
	}
	if req.SideName != nil {
		updates["side_name"] = *req.SideName
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if req.ICP != nil {
		updates["icp"] = *req.ICP
	}
	if req.PoliceICP != nil {
		updates["police_icp"] = *req.PoliceICP
	}
	// 主题：*string 支持清空（nil 表示不修改，空字符串表示跟随系统）
	if req.Theme != nil {
		updates["theme"] = *req.Theme
	}
	if req.AboutContent != nil {
		updates["about_content"] = *req.AboutContent
	}
	if req.CommentNeedReview != nil {
		updates["comment_need_review"] = *req.CommentNeedReview
	}
	if req.CommentNeedCaptcha != nil {
		updates["comment_need_captcha"] = *req.CommentNeedCaptcha
	}
	// 自定义 CSS 变量：*string 支持清空（nil 表示不修改，空字符串表示清除）
	if req.CustomCSSVars != nil {
		updates["custom_css_vars"] = *req.CustomCSSVars
	}
	if req.CustomCSSVarsDark != nil {
		updates["custom_css_vars_dark"] = *req.CustomCSSVarsDark
	}
	if req.CustomCSSEnabled != nil {
		updates["custom_css_enabled"] = *req.CustomCSSEnabled
	}
	if req.SocialLinks != nil {
		if data, err := json.Marshal(req.SocialLinks); err == nil {
			updates["social_links"] = string(data)
		}
	}
	if req.Extra != nil {
		if data, err := json.Marshal(req.Extra); err == nil {
			updates["extra"] = string(data)
		}
	}
	if req.HomeImages != nil {
		if data, err := json.Marshal(req.HomeImages); err == nil {
			updates["home_images"] = string(data)
		}
	}
	if req.CreatedBy != "" {
		updates["created_by"] = req.CreatedBy
	}

	if len(updates) > 0 {
		if err := s.DB().Model(&cfg).Updates(updates).Error; err != nil {
			return nil, NewServerError("更新配置失败")
		}
		s.invalidateCache()
		s.Log().Info("更新配置", zap.Uint("id", cfg.ID), zap.String("version", cfg.Version))
	}
	return &cfg, nil
}

// ActivateConfig 激活指定版本配置；同时取消其它版本的激活状态
func (s *ConfigService) ActivateConfig(id uint) *BizError {
	var cfg model.SiteConfig
	if err := s.DB().First(&cfg, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewNotFound("配置不存在")
		}
		return NewServerError("查询配置失败")
	}

	// 事务：取消其它激活 + 激活目标
	err := s.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.SiteConfig{}).
			Where("is_active = ? AND id <> ?", 1, cfg.ID).
			Update("is_active", 0).Error; err != nil {
			return err
		}
		return tx.Model(&cfg).Update("is_active", 1).Error
	})
	if err != nil {
		return NewServerError("激活配置失败")
	}
	s.invalidateCache()
	s.Log().Info("激活配置", zap.Uint("id", cfg.ID), zap.String("version", cfg.Version))
	return nil
}

// DeleteConfig 删除配置；激活中的配置禁止删除
func (s *ConfigService) DeleteConfig(id uint) *BizError {
	var cfg model.SiteConfig
	if err := s.DB().First(&cfg, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewNotFound("配置不存在")
		}
		return NewServerError("查询配置失败")
	}
	// 激活中的配置禁止删除
	if cfg.IsActive == 1 {
		return NewConflict("激活中的配置无法删除，请先激活其它版本")
	}

	if err := s.DB().Delete(&cfg).Error; err != nil {
		return NewServerError("删除配置失败")
	}
	s.invalidateCache()
	s.Log().Info("删除配置", zap.Uint("id", cfg.ID), zap.String("version", cfg.Version))
	return nil
}
