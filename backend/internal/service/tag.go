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

// TagService 标签相关业务
type TagService struct {
	*Service
}

// NewTagService 创建标签服务
func NewTagService(base *Service) *TagService {
	return &TagService{Service: base}
}

// ListTags 公开标签列表：返回全部启用标签，附已发布文章数量
func (s *TagService) ListTags() ([]model.Tag, *BizError) {
	var tags []model.Tag
	if err := s.DB().Model(&model.Tag{}).Order("id ASC").Find(&tags).Error; err != nil {
		return nil, NewServerError("查询标签失败")
	}

	// 一次聚合查询统计每个标签关联的已发布文章数，避免 N+1
	type tagCountRow struct {
		TagID     uint
		PostCount int64
	}
	var counts []tagCountRow
	if err := s.DB().Model(&model.Post{}).
		Select("post_tags.tag_id AS tag_id, COUNT(DISTINCT posts.id) AS post_count").
		Joins("JOIN post_tags ON post_tags.post_id = posts.id").
		Where("posts.status = ?", PostStatusPublished).
		Group("post_tags.tag_id").
		Scan(&counts).Error; err != nil {
		return nil, NewServerError("查询标签文章数失败")
	}

	countMap := make(map[uint]int64, len(counts))
	for _, row := range counts {
		countMap[row.TagID] = row.PostCount
	}
	for i := range tags {
		tags[i].PostCount = countMap[tags[i].ID]
	}
	return tags, nil
}

// ListAdminTags 管理端标签列表（分页）
func (s *TagService) ListAdminTags(req *response.PageRequest) ([]model.Tag, int64, *BizError) {
	var total int64
	if err := s.DB().Model(&model.Tag{}).Count(&total).Error; err != nil {
		return nil, 0, NewServerError("查询标签失败")
	}

	var tags []model.Tag
	if err := s.DB().Order("id DESC").
		Limit(req.PageSize).Offset(req.Offset()).
		Find(&tags).Error; err != nil {
		return nil, 0, NewServerError("查询标签失败")
	}
	return tags, total, nil
}

// CreateTag 创建标签
func (s *TagService) CreateTag(req *dto.TagRequest) (*model.Tag, *BizError) {
	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = utils.GenerateSlug(req.Name)
	}
	slug = s.uniqueSlug(slug, 0)

	tag := model.Tag{Name: req.Name, Slug: slug}
	if err := s.DB().Create(&tag).Error; err != nil {
		return nil, NewServerError("创建标签失败")
	}
	s.Log().Info("创建标签", zap.Uint("id", tag.ID), zap.String("name", tag.Name))
	return &tag, nil
}

// UpdateTag 更新标签
func (s *TagService) UpdateTag(id uint, req *dto.TagRequest) (*model.Tag, *BizError) {
	var tag model.Tag
	if err := s.DB().First(&tag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("标签不存在")
		}
		return nil, NewServerError("查询标签失败")
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Slug != "" {
		updates["slug"] = s.uniqueSlug(req.Slug, tag.ID)
	}

	if len(updates) > 0 {
		if err := s.DB().Model(&tag).Updates(updates).Error; err != nil {
			return nil, NewServerError("更新标签失败")
		}
		s.DB().First(&tag, id)
	}
	s.Log().Info("更新标签", zap.Uint("id", tag.ID), zap.String("name", tag.Name))
	return &tag, nil
}

// DeleteTag 删除标签（物理删除）；删除时清理 post_tags 关联
func (s *TagService) DeleteTag(id uint) *BizError {
	var tag model.Tag
	if err := s.DB().First(&tag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewNotFound("标签不存在")
		}
		return NewServerError("查询标签失败")
	}

	err := s.DB().Transaction(func(tx *gorm.DB) error {
		// 清理文章-标签关联
		if err := tx.Where("tag_id = ?", id).Delete(&model.PostTag{}).Error; err != nil {
			return err
		}
		// 物理删除标签（tags 为轻量元数据表，无需软删除）
		return tx.Delete(&tag).Error
	})
	if err != nil {
		return NewServerError("删除标签失败")
	}
	s.Log().Info("删除标签", zap.Uint("id", tag.ID), zap.String("name", tag.Name))
	return nil
}

// uniqueSlug 保证标签 slug 唯一，冲突时追加数字后缀
func (s *TagService) uniqueSlug(slug string, excludeID uint) string {
	candidate := slug
	for i := 1; ; i++ {
		var count int64
		query := s.DB().Model(&model.Tag{}).Where("slug = ?", candidate)
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
