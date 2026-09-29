package handler

import (
	"blog/internal/dto"
	"blog/internal/service"
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListFriends 公开友链列表
// GET /api/friends
func (h *Handler) ListFriends(c *gin.Context) {
	friends, err := h.FriendService.ListFriends()
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, friends)
}

// ListAdminFriends 管理端友链列表
// GET /api/admin/friends
func (h *Handler) ListAdminFriends(c *gin.Context) {
	var req response.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, "查询参数错误")
		return
	}
	req.Normalize()

	friends, total, err := h.FriendService.ListAdminFriends(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.SuccessPage(c, friends, total, req.Page, req.PageSize)
}

// CreateFriend 创建友链
// POST /api/admin/friends
func (h *Handler) CreateFriend(c *gin.Context) {
	var req dto.FriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "友链名称和地址不能为空")
		return
	}

	friend, err := h.FriendService.CreateFriend(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, friend)
}

// UpdateFriend 更新友链
// PUT /api/admin/friends/:id
func (h *Handler) UpdateFriend(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "友链 ID 无效")
		return
	}

	var req dto.FriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请求体格式错误")
		return
	}

	friend, err := h.FriendService.UpdateFriend(id, &req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, friend)
}

// DeleteFriend 删除友链（软删除）
// DELETE /api/admin/friends/:id
func (h *Handler) DeleteFriend(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "友链 ID 无效")
		return
	}

	if err := h.FriendService.DeleteFriend(id); err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}
