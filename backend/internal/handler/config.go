package handler

import (
	"blog/internal/dto"
	"blog/internal/service"
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// PublicConfig 前台生效配置
// GET /api/configs/public
func (h *Handler) PublicConfig(c *gin.Context) {
	cfg := h.ConfigService.PublicConfig()
	if cfg == nil {
		// 无生效配置时返回空对象，避免前端报错
		response.Success(c, gin.H{})
		return
	}
	response.Success(c, cfg)
}

// ListAdminConfigs 管理端配置列表
// GET /api/admin/configs
func (h *Handler) ListAdminConfigs(c *gin.Context) {
	var req response.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, "查询参数错误")
		return
	}
	req.Normalize()

	configs, total, err := h.ConfigService.ListAdminConfigs(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.SuccessPage(c, configs, total, req.Page, req.PageSize)
}

// GetAdminConfig 管理端单个配置详情
// GET /api/admin/configs/:id
func (h *Handler) GetAdminConfig(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "配置 ID 无效")
		return
	}

	cfg, err := h.ConfigService.GetAdminConfig(id)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, cfg)
}

// CreateConfig 新增配置；创建后自动置为激活
// POST /api/admin/configs
func (h *Handler) CreateConfig(c *gin.Context) {
	var req dto.ConfigUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请求体格式错误")
		return
	}
	if req.Version == "" {
		response.ParamError(c, "version 不能为空")
		return
	}

	cfg, err := h.ConfigService.CreateConfig(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, cfg)
}

// UpdateConfig 更新配置；仅更新提供的字段
// PUT /api/admin/configs/:id
func (h *Handler) UpdateConfig(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "配置 ID 无效")
		return
	}

	var req dto.ConfigUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请求体格式错误")
		return
	}

	cfg, err := h.ConfigService.UpdateConfig(id, &req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, cfg)
}

// ActivateConfig 激活指定版本配置
// PUT /api/admin/configs/:id/activate
func (h *Handler) ActivateConfig(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "配置 ID 无效")
		return
	}

	if err := h.ConfigService.ActivateConfig(id); err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "配置已激活"})
}

// DeleteConfig 删除配置；激活中的配置禁止删除
// DELETE /api/admin/configs/:id
func (h *Handler) DeleteConfig(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "配置 ID 无效")
		return
	}

	if err := h.ConfigService.DeleteConfig(id); err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}
