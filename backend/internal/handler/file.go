package handler

import (
	"context"

	"blog/internal/dto"
	"blog/internal/middleware"
	"blog/internal/service"
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// UploadFile 上传文件（multipart/form-data: file 必填, dir 可选）
// POST /api/admin/files
func (h *Handler) UploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.ParamError(c, "请选择要上传的文件")
		return
	}

	dir := c.PostForm("dir")
	userID := c.GetUint(middleware.CtxUserID)

	file, err := h.FileService.Upload(fileHeader, dir, userID)
	if err != nil {
		service.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{
		"id":        file.ID,
		"filename":  file.Filename,
		"url":       h.FileService.GetURL(context.Background(), file.Path),
		"size":      file.Size,
		"mime_type": file.MimeType,
	})
}

// ListFiles 文件列表
// GET /api/admin/files
func (h *Handler) ListFiles(c *gin.Context) {
	var req dto.FileQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, "查询参数错误")
		return
	}
	req.Normalize()

	files, total, err := h.FileService.ListFiles(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.SuccessPage(c, files, total, req.Page, req.PageSize)
}

// DeleteFile 删除文件（清理存储并删除数据库记录）
// DELETE /api/admin/files/:id
func (h *Handler) DeleteFile(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "文件 ID 无效")
		return
	}

	if err := h.FileService.DeleteFile(id); err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}
