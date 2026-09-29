package captcha

// DevProvider 开发环境验证码提供者
type DevProvider struct{}

func NewDevProvider() CaptchaProvider {
	return &DevProvider{}
}

var _ CaptchaProvider = (*DevProvider)(nil)

func (p *DevProvider) GenerateCode() string {
	return "1234"
}

func (p *DevProvider) Send(target string, val *CaptchaVal) error {
	// 开发环境不实际发送
	return nil
}

func (p *DevProvider) ValidateFormat(target string) bool {
	// 开发环境接受任何格式
	return true
}

func (p *DevProvider) GetType() string {
	return "dev"
}
