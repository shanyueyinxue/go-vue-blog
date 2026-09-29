package service

import (
	"encoding/json"
	"strings"

	"blog/internal/dto"
	"blog/internal/model"
	"blog/pkg/response"
	"blog/pkg/utils"

	"gorm.io/gorm"
)

// WorkService 作品集相关业务
type WorkService struct {
	*Service
}

// NewWorkService 创建作品服务
func NewWorkService(base *Service) *WorkService {
	return &WorkService{Service: base}
}

// ListWorks 公开作品列表（仅启用，置顶优先，再按 sort_order 升序）
func (s *WorkService) ListWorks() ([]model.Work, *BizError) {
	var works []model.Work
	if err := s.DB().Model(&model.Work{}).
		Where("status = ?", 1).
		Order("is_top DESC").Order("sort_order ASC").Order("id ASC").
		Find(&works).Error; err != nil {
		return nil, NewServerError("查询作品失败")
	}
	return works, nil
}

// ListAdminWorks 管理端作品列表（分页，含禁用）
func (s *WorkService) ListAdminWorks(req *response.PageRequest) ([]model.Work, int64, *BizError) {
	var total int64
	if err := s.DB().Model(&model.Work{}).Count(&total).Error; err != nil {
		return nil, 0, NewServerError("查询作品失败")
	}

	var works []model.Work
	if err := s.DB().Order("is_top DESC").Order("sort_order ASC").Order("id ASC").
		Limit(req.PageSize).Offset(req.Offset()).
		Find(&works).Error; err != nil {
		return nil, 0, NewServerError("查询作品失败")
	}
	return works, total, nil
}

// GetWork 管理端作品详情
func (s *WorkService) GetWork(id uint) (*model.Work, *BizError) {
	var work model.Work
	if err := s.DB().First(&work, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("作品不存在")
		}
		return nil, NewServerError("查询作品失败")
	}
	return &work, nil
}

// CreateWork 创建作品
func (s *WorkService) CreateWork(req *dto.WorkRequest) (*model.Work, *BizError) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, NewParamError("作品名不能为空")
	}

	// slug：留空按作品名自动生成，冲突时追加数字后缀
	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = utils.GenerateSlug(name)
	}
	slug = s.uniqueSlug(slug, 0)

	work := model.Work{
		Name:      name,
		Slug:      slug,
		Status:    1,
		SortOrder: 0,
	}
	// 可选字段（*string 支持清空）
	if req.Description != nil {
		work.Description = *req.Description
	}
	if req.Cover != nil {
		work.Cover = *req.Cover
	}
	if req.DemoURL != nil {
		work.DemoURL = *req.DemoURL
	}
	if req.ArticleURL != nil {
		work.ArticleURL = *req.ArticleURL
	}
	if req.RepoURL != nil {
		work.RepoURL = *req.RepoURL
	}
	if req.Year != nil {
		work.Year = *req.Year
	}
	if req.IsTop != nil {
		work.IsTop = *req.IsTop
	}
	if req.Status != nil {
		work.Status = *req.Status
	}
	if req.SortOrder != nil {
		work.SortOrder = *req.SortOrder
	}
	// JSON 字段序列化存储
	if len(req.TechStack) > 0 {
		if data, err := json.Marshal(req.TechStack); err == nil {
			work.TechStack = string(data)
		}
	}
	if req.Extra != nil {
		if data, err := json.Marshal(req.Extra); err == nil {
			work.Extra = string(data)
		}
	}

	err := withUniqueSlugRetry(maxSlugRetryAttempts,
		func() error { return s.DB().Create(&work).Error },
		// 唯一键冲突：软删除墓碑或并发占位，含软删除查重后重试
		func() { work.Slug = s.uniqueSlugUnscoped(work.Slug, 0) },
	)
	if err != nil {
		return nil, NewServerError("创建作品失败")
	}
	return &work, nil
}

// UpdateWork 更新作品（仅更新提供的字段）
func (s *WorkService) UpdateWork(id uint, req *dto.WorkRequest) (*model.Work, *BizError) {
	var work model.Work
	if err := s.DB().First(&work, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("作品不存在")
		}
		return nil, NewServerError("查询作品失败")
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	// slug 冲突时追加后缀
	if req.Slug != "" {
		updates["slug"] = s.uniqueSlug(strings.TrimSpace(req.Slug), id)
	}
	// 文本字段：*string 支持清空（nil 不修改，空串清除）
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Cover != nil {
		updates["cover"] = *req.Cover
	}
	if req.DemoURL != nil {
		updates["demo_url"] = *req.DemoURL
	}
	if req.ArticleURL != nil {
		updates["article_url"] = *req.ArticleURL
	}
	if req.RepoURL != nil {
		updates["repo_url"] = *req.RepoURL
	}
	if req.Year != nil {
		updates["year"] = *req.Year
	}
	if req.IsTop != nil {
		updates["is_top"] = *req.IsTop
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	// JSON 字段：tech_stack 传空数组表示清空
	if req.TechStack != nil {
		if data, err := json.Marshal(req.TechStack); err == nil {
			updates["tech_stack"] = string(data)
		}
	}
	if req.Extra != nil {
		if data, err := json.Marshal(req.Extra); err == nil {
			updates["extra"] = string(data)
		}
	}

	err := withUniqueSlugRetry(maxSlugRetryAttempts,
		func() error {
			if len(updates) == 0 {
				return nil
			}
			if err := s.DB().Model(&work).Updates(updates).Error; err != nil {
				return err
			}
			return s.DB().First(&work, id).Error
		},
		// 唯一键冲突：仅在本次更新涉及 slug 时有重试意义
		func() {
			if req.Slug != "" {
				newSlug := s.uniqueSlugUnscoped(req.Slug, work.ID)
				work.Slug = newSlug
				updates["slug"] = newSlug
			}
		},
	)
	if err != nil {
		return nil, NewServerError("更新作品失败")
	}
	return &work, nil
}

// DeleteWork 删除作品（软删除）
func (s *WorkService) DeleteWork(id uint) *BizError {
	var work model.Work
	if err := s.DB().First(&work, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewNotFound("作品不存在")
		}
		return NewServerError("查询作品失败")
	}

	if err := s.DB().Delete(&work).Error; err != nil {
		return NewServerError("删除作品失败")
	}
	return nil
}

// uniqueSlug 保证作品 slug 唯一，冲突时追加数字后缀。
// 按默认作用域查重（忽略软删除行），正常创建/更新路径使用。
func (s *WorkService) uniqueSlug(slug string, excludeID uint) string {
	return s.uniqueSlugWithScope(slug, excludeID, false)
}

// uniqueSlugUnscoped 含软删除记录查重：唯一键冲突重试时使用（见 withUniqueSlugRetry）
func (s *WorkService) uniqueSlugUnscoped(slug string, excludeID uint) string {
	return s.uniqueSlugWithScope(slug, excludeID, true)
}

// uniqueSlugWithScope 通用 slug 查重；includeDeleted=true 时用 Unscoped 含软删除行
func (s *WorkService) uniqueSlugWithScope(slug string, excludeID uint, includeDeleted bool) string {
	candidate := slug
	for i := 1; ; i++ {
		var count int64
		query := s.DB().Model(&model.Work{})
		if includeDeleted {
			query = query.Unscoped()
		}
		query = query.Where("slug = ?", candidate)
		if excludeID > 0 {
			query = query.Where("id <> ?", excludeID)
		}
		query.Count(&count)
		if count == 0 {
			return candidate
		}
		candidate = slug + "-" + utils.IntToString(i)
	}
}
