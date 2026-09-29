package service

import (
	"strings"

	"blog/internal/dto"
	"blog/internal/model"
	"blog/pkg/response"
	"blog/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CategoryService 分类相关业务
type CategoryService struct {
	*Service
}

// NewCategoryService 创建分类服务
func NewCategoryService(base *Service) *CategoryService {
	return &CategoryService{Service: base}
}

// ListCategories 公开分类列表：仅返回启用分类，按 sort_order 升序，附文章数量
func (s *CategoryService) ListCategories() ([]model.Category, *BizError) {
	var categories []model.Category
	err := s.DB().Model(&model.Category{}).
		Where("status = ?", 1).
		Order("sort_order ASC").Order("id ASC").
		Find(&categories).Error
	if err != nil {
		return nil, NewServerError("查询分类失败")
	}

	// 一次聚合查询统计每个分类下的已发布文章数，避免 N+1
	type categoryCountRow struct {
		CategoryID uint
		PostCount  int64
	}
	var counts []categoryCountRow
	if err := s.DB().Model(&model.Post{}).
		Select("category_id, COUNT(*) AS post_count").
		Where("status = ? AND category_id IS NOT NULL", PostStatusPublished).
		Group("category_id").
		Scan(&counts).Error; err != nil {
		return nil, NewServerError("查询分类文章数失败")
	}

	countMap := make(map[uint]int64, len(counts))
	for _, row := range counts {
		countMap[row.CategoryID] = row.PostCount
	}
	for i := range categories {
		categories[i].PostCount = countMap[categories[i].ID]
	}
	return categories, nil
}

// ListAdminCategories 管理端分类列表（含禁用，分页）
func (s *CategoryService) ListAdminCategories(req *response.PageRequest) ([]model.Category, int64, *BizError) {
	var total int64
	if err := s.DB().Model(&model.Category{}).Count(&total).Error; err != nil {
		return nil, 0, NewServerError("查询分类失败")
	}

	var categories []model.Category
	err := s.DB().Order("sort_order ASC").Order("id ASC").
		Limit(req.PageSize).Offset(req.Offset()).
		Find(&categories).Error
	if err != nil {
		return nil, 0, NewServerError("查询分类失败")
	}
	return categories, total, nil
}

// CreateCategory 创建分类
func (s *CategoryService) CreateCategory(req *dto.CategoryRequest) (*model.Category, *BizError) {
	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = utils.GenerateSlug(req.Name)
	}
	slug = s.uniqueSlug(slug, 0)

	category := model.Category{
		Name:        req.Name,
		Slug:        slug,
		Description: derefStr(req.Description),
		Status:      1,
	}
	if req.Status != nil {
		category.Status = *req.Status
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}

	err := withUniqueSlugRetry(maxSlugRetryAttempts,
		func() error { return s.DB().Create(&category).Error },
		// 唯一键冲突：软删除墓碑或并发占位，含软删除查重后重试
		func() { category.Slug = s.uniqueSlugUnscoped(category.Slug, 0) },
	)
	if err != nil {
		return nil, NewServerError("创建分类失败")
	}
	s.Log().Info("创建分类", zap.Uint("id", category.ID), zap.String("name", category.Name))
	return &category, nil
}

// UpdateCategory 更新分类
func (s *CategoryService) UpdateCategory(id uint, req *dto.CategoryRequest) (*model.Category, *BizError) {
	var category model.Category
	if err := s.DB().First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("分类不存在")
		}
		return nil, NewServerError("查询分类失败")
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Slug != "" {
		updates["slug"] = s.uniqueSlug(req.Slug, category.ID)
	}
	// 描述：*string 支持清空（nil 不修改，空串清除）
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	err := withUniqueSlugRetry(maxSlugRetryAttempts,
		func() error {
			if len(updates) == 0 {
				return nil
			}
			if err := s.DB().Model(&category).Updates(updates).Error; err != nil {
				return err
			}
			return s.DB().First(&category, id).Error
		},
		// 唯一键冲突：仅在本次更新涉及 slug 时有重试意义
		func() {
			if req.Slug != "" {
				newSlug := s.uniqueSlugUnscoped(req.Slug, category.ID)
				category.Slug = newSlug
				updates["slug"] = newSlug
			}
		},
	)
	if err != nil {
		return nil, NewServerError("更新分类失败")
	}
	s.Log().Info("更新分类", zap.Uint("id", category.ID), zap.String("name", category.Name))
	return &category, nil
}

// DeleteCategory 删除分类；分类下存在文章时拒绝删除
func (s *CategoryService) DeleteCategory(id uint) *BizError {
	var category model.Category
	if err := s.DB().First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewNotFound("分类不存在")
		}
		return NewServerError("查询分类失败")
	}

	// 检查分类下是否有文章
	var count int64
	s.DB().Model(&model.Post{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		return NewConflict("该分类下存在文章，无法删除")
	}

	if err := s.DB().Delete(&category).Error; err != nil {
		return NewServerError("删除分类失败")
	}
	s.Log().Info("删除分类", zap.Uint("id", category.ID), zap.String("name", category.Name))
	return nil
}

// uniqueSlug 保证分类 slug 唯一，冲突时追加数字后缀。
// 按默认作用域查重（忽略软删除行），正常创建/更新路径使用。
func (s *CategoryService) uniqueSlug(slug string, excludeID uint) string {
	return s.uniqueSlugWithScope(slug, excludeID, false)
}

// uniqueSlugUnscoped 含软删除记录查重：唯一键冲突重试时使用（见 withUniqueSlugRetry）
func (s *CategoryService) uniqueSlugUnscoped(slug string, excludeID uint) string {
	return s.uniqueSlugWithScope(slug, excludeID, true)
}

// uniqueSlugWithScope 通用 slug 查重；includeDeleted=true 时用 Unscoped 含软删除行
func (s *CategoryService) uniqueSlugWithScope(slug string, excludeID uint, includeDeleted bool) string {
	candidate := slug
	for i := 1; ; i++ {
		var count int64
		query := s.DB().Model(&model.Category{})
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
