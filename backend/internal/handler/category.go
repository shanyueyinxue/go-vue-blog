package handler

import (
	"blog/internal/dto"
	"blog/internal/service"
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListCategories 公开分类列表
// GET /api/categories
func (h *Handler) ListCategories(c *gin.Context) {
	categories, err := h.CategoryService.ListCategories()
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, categories)
}

// ListAdminCategories 管理端分类列表
// GET /api/admin/categories
func (h *Handler) ListAdminCategories(c *gin.Context) {
	var req response.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, "查询参数错误")
		return
	}
	req.Normalize()

	categories, total, err := h.CategoryService.ListAdminCategories(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.SuccessPage(c, categories, total, req.Page, req.PageSize)
}

// CreateCategory 创建分类
// POST /api/admin/categories
func (h *Handler) CreateCategory(c *gin.Context) {
	var req dto.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "分类名称不能为空")
		return
	}

	category, err := h.CategoryService.CreateCategory(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, category)
}

// UpdateCategory 更新分类
// PUT /api/admin/categories/:id
func (h *Handler) UpdateCategory(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "分类 ID 无效")
		return
	}

	var req dto.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请求体格式错误")
		return
	}

	category, err := h.CategoryService.UpdateCategory(id, &req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, category)
}

// DeleteCategory 删除分类；分类下存在文章时拒绝删除
// DELETE /api/admin/categories/:id
func (h *Handler) DeleteCategory(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "分类 ID 无效")
		return
	}

	if err := h.CategoryService.DeleteCategory(id); err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}
