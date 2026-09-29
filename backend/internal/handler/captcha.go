package handler

import (
	"blog/internal/dto"
	"blog/internal/service"
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// SendEmailCaptcha 发送邮箱验证码
// POST /api/captcha/email
func (h *Handler) SendEmailCaptcha(c *gin.Context) {
	var req dto.EmailCaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "email 不能为空")
		return
	}

	token, expiresIn, err := h.CaptchaService.SendEmailCaptcha(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{
		"token":      token,
		"expires_in": expiresIn,
	})
}

// VerifyCaptcha 校验验证码
// POST /api/captcha/verify
func (h *Handler) VerifyCaptcha(c *gin.Context) {
	var req dto.VerifyCaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "token 和 code 不能为空")
		return
	}

	valid, err := h.CaptchaService.VerifyCaptcha(&req)
	if err != nil {
		service.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"valid": valid})
}
