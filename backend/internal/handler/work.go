package handler

import (
	"blog/internal/dto"
	"blog/internal/service"
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListWorks 公开作品列表（仅启用）
// GET /api/works
func (h *Handler) ListWorks(c *gin.Context) {
	works, err := h.WorkService.ListWorks()
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, works)
}

// ListAdminWorks 管理端作品列表
// GET /api/admin/works
func (h *Handler) ListAdminWorks(c *gin.Context) {
	var req response.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, "查询参数错误")
		return
	}
	req.Normalize()

	works, total, err := h.WorkService.ListAdminWorks(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.SuccessPage(c, works, total, req.Page, req.PageSize)
}

// GetAdminWork 管理端作品详情
// GET /api/admin/works/:id
func (h *Handler) GetAdminWork(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "作品 ID 无效")
		return
	}

	work, err := h.WorkService.GetWork(id)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, work)
}

// CreateWork 创建作品
// POST /api/admin/works
func (h *Handler) CreateWork(c *gin.Context) {
	var req dto.WorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "作品名不能为空")
		return
	}

	work, err := h.WorkService.CreateWork(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, work)
}

// UpdateWork 更新作品
// PUT /api/admin/works/:id
func (h *Handler) UpdateWork(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "作品 ID 无效")
		return
	}

	var req dto.WorkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请求体格式错误")
		return
	}

	work, err := h.WorkService.UpdateWork(id, &req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, work)
}

// DeleteWork 删除作品（软删除）
// DELETE /api/admin/works/:id
func (h *Handler) DeleteWork(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "作品 ID 无效")
		return
	}

	if err := h.WorkService.DeleteWork(id); err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}
