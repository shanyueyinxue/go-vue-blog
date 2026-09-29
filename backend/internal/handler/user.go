package handler

import (
	"blog/internal/dto"
	"blog/internal/middleware"
	"blog/internal/model"
	"blog/internal/service"
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// UpdateProfile 修改个人信息（本人）
// PUT /api/admin/profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	userVal, ok := c.Get(middleware.CtxUserKey)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	user := userVal.(*model.User)

	var req dto.ProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请求体格式错误")
		return
	}

	updated, err := h.UserService.UpdateProfile(user, &req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, updated)
}

// SendForgotPasswordCode 忘记密码：发送邮箱验证码（公开）
// POST /api/forgot-password/send-code
func (h *Handler) SendForgotPasswordCode(c *gin.Context) {
	var req dto.ForgotPasswordSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "email 不能为空")
		return
	}

	token, expiresIn, err := h.UserService.SendForgotPasswordCode(req.Email)
	if err != nil {
		service.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{
		"token":      token,
		"expires_in": expiresIn,
	})
}

// ResetForgotPassword 忘记密码：校验验证码并重置密码（公开）
// POST /api/forgot-password/reset
func (h *Handler) ResetForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "email、验证码和新密码不能为空")
		return
	}

	if err := h.UserService.ResetForgotPassword(&req); err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "密码重置成功，请使用新密码登录"})
}

// ListUsers 用户列表（仅超级管理员）
// GET /api/admin/users
func (h *Handler) ListUsers(c *gin.Context) {
	var req dto.UserQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ParamError(c, "查询参数错误")
		return
	}
	req.Normalize()

	users, total, err := h.UserService.ListUsers(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.SuccessPage(c, users, total, req.Page, req.PageSize)
}

// CreateUser 创建用户（仅超级管理员）
// POST /api/admin/users
func (h *Handler) CreateUser(c *gin.Context) {
	var req dto.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "用户名和密码不能为空")
		return
	}

	user, err := h.UserService.CreateUser(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, user)
}

// GetUser 用户详情（仅超级管理员）
// GET /api/admin/users/:id
func (h *Handler) GetUser(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "用户 ID 无效")
		return
	}

	user, err := h.UserService.GetUser(id)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, user)
}

// UpdateUser 修改用户信息（仅超级管理员）
// PUT /api/admin/users/:id
func (h *Handler) UpdateUser(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "用户 ID 无效")
		return
	}

	var req dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请求体格式错误")
		return
	}

	userVal, _ := c.Get(middleware.CtxUserKey)
	operator := userVal.(*model.User)

	user, err := h.UserService.UpdateUser(id, &req, operator)
	if err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, user)
}

// ResetUserPassword 重置用户密码（仅超级管理员，无需原密码）
// PUT /api/admin/users/:id/password
func (h *Handler) ResetUserPassword(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "用户 ID 无效")
		return
	}

	var req dto.UserPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "new_password 不能为空")
		return
	}

	if err := h.UserService.ResetUserPassword(id, req.NewPassword); err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "密码重置成功"})
}

// DeleteUser 删除用户（仅超级管理员，软删除）
// DELETE /api/admin/users/:id
func (h *Handler) DeleteUser(c *gin.Context) {
	id := response.GetUintParam(c, "id")
	if id == 0 {
		response.ParamError(c, "用户 ID 无效")
		return
	}

	userVal, _ := c.Get(middleware.CtxUserKey)
	operator := userVal.(*model.User)

	if err := h.UserService.DeleteUser(id, operator); err != nil {
		service.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"message": "删除成功"})
}
