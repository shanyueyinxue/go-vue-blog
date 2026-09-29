package handler

import (
	"blog/internal/dto"
	"blog/internal/service"
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListMusics 公开音乐列表（仅启用，供前台 APlayer 使用）
// GET /api/musics
func (h *Handler) ListMusics(c *gin.Context) {
	musics, err := h.MusicService.ListMusics()
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, musics)
}

// ListAdminMusics 管理端音乐列表（分页）
// GET /api/admin/musics
func (h *Handler) ListAdminMusics(c *gin.Context) {
	var req dto.MusicQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, "查询参数错误")
		return
	}
	req.Normalize()

	musics, total, err := h.MusicService.ListAdminMusics(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.SuccessPage(c, musics, total, req.Page, req.PageSize)
}

// GetAdminMusic 管理端音乐详情
// GET /api/admin/musics/:id
func (h *Handler) GetAdminMusic(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "音乐 ID 无效")
		return
	}

	music, err := h.MusicService.GetMusic(id)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, music)
}

// CreateMusic 创建音乐
// POST /api/admin/musics
func (h *Handler) CreateMusic(c *gin.Context) {
	var req dto.MusicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "歌名不能为空")
		return
	}

	music, err := h.MusicService.CreateMusic(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, music)
}

// UpdateMusic 更新音乐
// PUT /api/admin/musics/:id
func (h *Handler) UpdateMusic(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "音乐 ID 无效")
		return
	}

	var req dto.MusicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请求体格式错误")
		return
	}

	music, err := h.MusicService.UpdateMusic(id, &req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, music)
}

// DeleteMusic 删除音乐（同步删除关联音频文件记录与存储对象）
// DELETE /api/admin/musics/:id
func (h *Handler) DeleteMusic(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "音乐 ID 无效")
		return
	}

	if err := h.MusicService.DeleteMusic(id); err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}
