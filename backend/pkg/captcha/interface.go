package captcha

// CaptchaVal 用于存储验证码相关信息；插入tempalte中
type CaptchaVal struct {
	Subject      string
	Username     string
	Code         string
	Expiry       int    // 验证码过期时间，单位：秒
	ValidMinutes string // 验证码有效时间，单位：分钟
	Note         string // 验证码提示信息
}

// CaptchaProvider 验证码提供者接口
type CaptchaProvider interface {
	// GenerateCode 生成验证码
	GenerateCode() string
	// Send 发送验证码
	Send(target string, val *CaptchaVal) error
	// ValidateFormat 验证目标格式（邮箱/手机号）
	ValidateFormat(target string) bool
	// GetType 获取提供者类型
	GetType() string
}

// 自定义错误
type CaptchaError interface {
	error
}

type CaptchaErr string

func (e CaptchaErr) Error() string {
	return string(e)
}

var (
	ErrStoreCode  CaptchaErr = "存储验证码错误"
	ErrVerifyCode CaptchaErr = "验证码验证错误"
	ErrDeleteCode CaptchaErr = "删除验证码错误"
	ErrSendCode   CaptchaErr = "发送验证码错误"
	ErrFormat     CaptchaErr = "格式错误"
)
