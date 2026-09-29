package handler

import (
	"blog/internal/dto"
	"blog/internal/middleware"
	"blog/internal/model"
	"blog/internal/service"
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListComments 公开获取某文章已通过评论
// GET /api/posts/:slug/comments
func (h *Handler) ListComments(c *gin.Context) {
	slug := c.Param("slug")

	var req dto.GetCommentsRequest
	_ = c.ShouldBindQuery(&req)
	req.Normalize()

	comments, total, err := h.CommentService.ListComments(slug, &req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.SuccessPage(c, comments, total, req.Page, req.PageSize)
}

// CreateComment 提交评论
// POST /api/comments
func (h *Handler) CreateComment(c *gin.Context) {
	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "昵称和评论内容不能为空")
		return
	}

	comment, status, err := h.CommentService.CreateComment(&req, middleware.ClientIP(c), c.Request.UserAgent())
	if err != nil {
		service.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{
		"comment": comment,
		"status":  status,
	})
}

// ListAdminComments 管理端评论列表
// GET /api/admin/comments
func (h *Handler) ListAdminComments(c *gin.Context) {
	var req dto.AdminCommentQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, "查询参数错误")
		return
	}
	req.Normalize()

	comments, total, err := h.CommentService.ListAdminComments(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.SuccessPage(c, comments, total, req.Page, req.PageSize)
}

// UpdateComment 修改评论（仅允许修改 status / is_author，其余字段忽略）
// PUT /api/admin/comments/:id
func (h *Handler) UpdateComment(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "评论 ID 无效")
		return
	}

	var req dto.CommentAdminUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请求体格式错误")
		return
	}
	if req.Status != nil && !isValidCommentStatus(*req.Status) {
		response.ParamError(c, "status 仅支持 pending / approved / rejected")
		return
	}
	if req.IsAuthor != nil && *req.IsAuthor != 0 && *req.IsAuthor != 1 {
		response.ParamError(c, "is_author 仅支持 0 或 1")
		return
	}

	comment, err := h.CommentService.UpdateComment(id, &req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, comment)
}

// isValidCommentStatus 校验评论状态取值
func isValidCommentStatus(status string) bool {
	return status == model.CommentStatusPending ||
		status == model.CommentStatusApproved ||
		status == model.CommentStatusRejected
}

// DeleteComment 删除评论（软删除，含级联子评论）
// DELETE /api/admin/comments/:id
func (h *Handler) DeleteComment(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "评论 ID 无效")
		return
	}

	if err := h.CommentService.DeleteComment(id); err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}
