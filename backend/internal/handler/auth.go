package handler

import (
	"blog/internal/dto"
	"blog/internal/middleware"
	"blog/internal/model"
	"blog/internal/service"
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// Login 管理员登录
// POST /api/admin/login
func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "用户名和密码不能为空")
		return
	}

	result, err := h.AuthService.Login(req.Username, req.Password, middleware.ClientIP(c))
	if err != nil {
		service.HandleError(c, err)
		return
	}

	// 登录成功，清零失败计数（防爆破计数器）
	middleware.ResetLoginFail(c)

	response.Success(c, gin.H{
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
		"expires_in":    result.ExpiresIn,
		"user":          result.User,
	})
}

// RefreshToken 刷新访问 Token
// POST /api/admin/refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "refresh_token 不能为空")
		return
	}

	result, err := h.AuthService.RefreshToken(req.RefreshToken)
	if err != nil {
		service.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
		"expires_in":    result.ExpiresIn,
	})
}

// Profile 获取管理员信息
// GET /api/admin/profile
func (h *Handler) Profile(c *gin.Context) {
	user, ok := c.Get(middleware.CtxUserKey)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	response.Success(c, user)
}

// UpdatePassword 修改管理员密码
// PUT /api/admin/profile/password
func (h *Handler) UpdatePassword(c *gin.Context) {
	userVal, ok := c.Get(middleware.CtxUserKey)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	user := userVal.(*model.User)

	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "原密码和新密码不能为空")
		return
	}

	if err := h.AuthService.UpdatePassword(user, req.OldPassword, req.NewPassword); err != nil {
		service.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "密码修改成功，请重新登录"})
}
