package service

import (
	"blog/internal/dto"
	"blog/internal/model"
	"blog/pkg/response"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// FriendService 友链相关业务
type FriendService struct {
	*Service
}

// NewFriendService 创建友链服务
func NewFriendService(base *Service) *FriendService {
	return &FriendService{Service: base}
}

// ListFriends 公开友链列表（仅启用，按 sort_order 升序）
func (s *FriendService) ListFriends() ([]model.Friend, *BizError) {
	var friends []model.Friend
	if err := s.DB().Model(&model.Friend{}).
		Where("status = ?", 1).
		Order("sort_order ASC").Order("id ASC").
		Find(&friends).Error; err != nil {
		return nil, NewServerError("查询友链失败")
	}
	return friends, nil
}

// ListAdminFriends 管理端友链列表（分页，含禁用）
func (s *FriendService) ListAdminFriends(req *response.PageRequest) ([]model.Friend, int64, *BizError) {
	var total int64
	if err := s.DB().Model(&model.Friend{}).Count(&total).Error; err != nil {
		return nil, 0, NewServerError("查询友链失败")
	}

	var friends []model.Friend
	if err := s.DB().Order("sort_order ASC").Order("id ASC").
		Limit(req.PageSize).Offset(req.Offset()).
		Find(&friends).Error; err != nil {
		return nil, 0, NewServerError("查询友链失败")
	}
	return friends, total, nil
}

// CreateFriend 创建友链
func (s *FriendService) CreateFriend(req *dto.FriendRequest) (*model.Friend, *BizError) {
	friend := model.Friend{
		Name:        req.Name,
		URL:         req.URL,
		Icon:        derefStr(req.Icon),
		Description: derefStr(req.Description),
		Status:      1,
	}
	if req.Status != nil {
		friend.Status = *req.Status
	}
	if req.SortOrder != nil {
		friend.SortOrder = *req.SortOrder
	}

	if err := s.DB().Create(&friend).Error; err != nil {
		return nil, NewServerError("创建友链失败")
	}
	s.Log().Info("创建友链", zap.Uint("id", friend.ID), zap.String("name", friend.Name))
	return &friend, nil
}

// UpdateFriend 更新友链
func (s *FriendService) UpdateFriend(id uint, req *dto.FriendRequest) (*model.Friend, *BizError) {
	var friend model.Friend
	if err := s.DB().First(&friend, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("友链不存在")
		}
		return nil, NewServerError("查询友链失败")
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.URL != "" {
		updates["url"] = req.URL
	}
	// 图标/描述：*string 支持清空（nil 不修改，空串清除）
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	if len(updates) > 0 {
		if err := s.DB().Model(&friend).Updates(updates).Error; err != nil {
			return nil, NewServerError("更新友链失败")
		}
		s.DB().First(&friend, id)
	}
	s.Log().Info("更新友链", zap.Uint("id", friend.ID), zap.String("name", friend.Name))
	return &friend, nil
}

// DeleteFriend 删除友链（软删除）
func (s *FriendService) DeleteFriend(id uint) *BizError {
	var friend model.Friend
	if err := s.DB().First(&friend, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewNotFound("友链不存在")
		}
		return NewServerError("查询友链失败")
	}

	if err := s.DB().Delete(&friend).Error; err != nil {
		return NewServerError("删除友链失败")
	}
	s.Log().Info("删除友链", zap.Uint("id", friend.ID), zap.String("name", friend.Name))
	return nil
}
