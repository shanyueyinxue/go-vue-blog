package handler

import (
	"blog/internal/dto"
	"blog/internal/middleware"
	"blog/internal/service"
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListPosts 公开文章列表
// GET /api/posts
func (h *Handler) ListPosts(c *gin.Context) {
	var req dto.PostListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, "查询参数错误")
		return
	}
	req.Normalize()

	posts, total, err := h.PostService.ListPosts(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.SuccessPage(c, posts, total, req.Page, req.PageSize)
}

// Archive 文章归档
// GET /api/posts/archive
func (h *Handler) Archive(c *gin.Context) {
	items, err := h.PostService.Archive()
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, items)
}

// GetPost 公开文章详情
// GET /api/posts/:slug
func (h *Handler) GetPost(c *gin.Context) {
	slug := c.Param("slug")

	post, err := h.PostService.GetPost(slug)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, post)
}

// ViewPost 文章浏览量 +1（按 IP 每小时去重）
// POST /api/posts/:slug/view
func (h *Handler) ViewPost(c *gin.Context) {
	slug := c.Param("slug")

	viewCount, err := h.PostService.ViewPost(slug, middleware.ClientIP(c))
	if err != nil {
		service.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"view_count": viewCount})
}

// ListAdminPosts 管理端文章列表
// GET /api/admin/posts
func (h *Handler) ListAdminPosts(c *gin.Context) {
	var req dto.PostListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, "查询参数错误")
		return
	}
	req.Normalize()

	posts, total, err := h.PostService.ListAdminPosts(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.SuccessPage(c, posts, total, req.Page, req.PageSize)
}

// CreatePost 创建文章
// POST /api/admin/posts
func (h *Handler) CreatePost(c *gin.Context) {
	var req dto.PostUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请求体格式错误")
		return
	}

	post, err := h.PostService.CreatePost(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, post)
}

// GetAdminPost 管理端文章详情
// GET /api/admin/posts/:id
func (h *Handler) GetAdminPost(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "文章 ID 无效")
		return
	}

	post, err := h.PostService.GetAdminPost(id)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, post)
}

// UpdatePost 更新文章
// PUT /api/admin/posts/:id
func (h *Handler) UpdatePost(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "文章 ID 无效")
		return
	}

	var req dto.PostUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请求体格式错误")
		return
	}

	post, err := h.PostService.UpdatePost(id, &req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, post)
}

// DeletePost 删除文章（软删除）
// DELETE /api/admin/posts/:id
func (h *Handler) DeletePost(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "文章 ID 无效")
		return
	}

	if err := h.PostService.DeletePost(id); err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}
