package handler

import (
	"blog/internal/dto"
	"blog/internal/service"
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListTags 公开标签列表
// GET /api/tags
func (h *Handler) ListTags(c *gin.Context) {
	tags, err := h.TagService.ListTags()
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, tags)
}

// ListAdminTags 管理端标签列表
// GET /api/admin/tags
func (h *Handler) ListAdminTags(c *gin.Context) {
	var req response.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, "查询参数错误")
		return
	}
	req.Normalize()

	tags, total, err := h.TagService.ListAdminTags(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.SuccessPage(c, tags, total, req.Page, req.PageSize)
}

// CreateTag 创建标签
// POST /api/admin/tags
func (h *Handler) CreateTag(c *gin.Context) {
	var req dto.TagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "标签名称不能为空")
		return
	}

	tag, err := h.TagService.CreateTag(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, tag)
}

// UpdateTag 更新标签
// PUT /api/admin/tags/:id
func (h *Handler) UpdateTag(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "标签 ID 无效")
		return
	}

	var req dto.TagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请求体格式错误")
		return
	}

	tag, err := h.TagService.UpdateTag(id, &req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, tag)
}

// DeleteTag 删除标签；删除时清理 post_tags 关联
// DELETE /api/admin/tags/:id
func (h *Handler) DeleteTag(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "标签 ID 无效")
		return
	}

	if err := h.TagService.DeleteTag(id); err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}
