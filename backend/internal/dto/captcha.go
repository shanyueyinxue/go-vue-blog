package dto

// EmailCaptchaRequest 发送邮箱验证码请求体
type EmailCaptchaRequest struct {
	Email string `json:"email" binding:"required"` // 目标邮箱
	Scene string `json:"scene"`                    // 场景：comment / login / forgot
}

// VerifyCaptchaRequest 校验验证码请求体
type VerifyCaptchaRequest struct {
	Token string `json:"token" binding:"required"` // 发送验证码时返回的 token
	Code  string `json:"code" binding:"required"`  // 验证码内容
}
