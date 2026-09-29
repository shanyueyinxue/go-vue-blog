package service

import (
	"time"

	"blog/internal/dto"
	"blog/pkg/utils"
)

// 验证码相关缓存 key 前缀与冷却时间
const (
	captchaTokenKeyPrefix = "captcha:token:"    // token -> 邮箱 映射
	captchaCooldownKey    = "captcha:cooldown:" // 发送冷却标记
	captchaSendCooldown   = 60 * time.Second    // 同一邮箱 60 秒内不可重复发送
)

// CaptchaService 验证码发送与校验相关业务
type CaptchaService struct {
	*Service
}

// NewCaptchaService 创建验证码服务
func NewCaptchaService(base *Service) *CaptchaService {
	return &CaptchaService{Service: base}
}

// SendEmailCaptcha 发送邮箱验证码，返回一次性 token 与有效期（秒）
func (s *CaptchaService) SendEmailCaptcha(req *dto.EmailCaptchaRequest) (string, int, *BizError) {
	if !utils.IsEmail(req.Email) {
		return "", 0, NewParamError("邮箱格式不正确")
	}

	// 频率限制：同一邮箱 60 秒内不可重复发送
	cooldownKey := captchaCooldownKey + req.Email
	if s.Cache().Exists(cooldownKey) {
		return "", 0, NewTooManyRequests("发送过于频繁，请稍后再试")
	}

	// 生成并发送验证码
	code := s.App.Captcha.GenerateCode()
	name := "访客"
	switch req.Scene {
	case "login":
		name = "管理员"
	case "forgot":
		name = "用户"
	}
	if err := s.App.Captcha.SendCode(code, name, req.Email); err != nil {
		return "", 0, NewServerError("验证码发送失败")
	}

	// 生成一次性 token 用于提交时携带（映射到邮箱）
	token := utils.RandomStringMixed(32)
	_ = s.Cache().Set(captchaTokenKeyPrefix+token, req.Email, time.Duration(s.App.Config.Captcha.Expiry)*time.Second)
	// 记录发送冷却
	_ = s.Cache().Set(cooldownKey, "1", captchaSendCooldown)

	return token, s.App.Config.Captcha.Expiry, nil
}

// VerifyCaptcha 校验验证码（通过 token 解析邮箱，校验通过后即作废）
func (s *CaptchaService) VerifyCaptcha(req *dto.VerifyCaptchaRequest) (bool, *BizError) {
	// 通过 token 解析目标邮箱
	email, err := s.Cache().GetString(captchaTokenKeyPrefix + req.Token)
	if err != nil {
		return false, NewParamError("token 无效或已过期")
	}

	// 校验验证码（校验通过后即作废）
	valid, _ := s.App.Captcha.VerifyCode(req.Code, email)
	return valid, nil
}

// VerifyForEmail 校验验证码是否属于指定邮箱（用于评论等业务场景），校验通过后即作废。
// 同时校验 token 绑定的邮箱与提交的邮箱一致，防止验证码被挪用到其他邮箱。
func (s *CaptchaService) VerifyForEmail(token, code, email string) *BizError {
	if token == "" || code == "" {
		return NewParamError("请先获取并填写邮箱验证码")
	}
	if !utils.IsEmail(email) {
		return NewParamError("请填写有效邮箱")
	}

	bound, err := s.Cache().GetString(captchaTokenKeyPrefix + token)
	if err != nil {
		return NewParamError("验证码已失效，请重新获取")
	}
	if bound != email {
		return NewParamError("验证码与邮箱不一致")
	}

	valid, _ := s.App.Captcha.VerifyCode(code, bound)
	if !valid {
		return NewParamError("验证码错误，请重新输入")
	}
	return nil
}
