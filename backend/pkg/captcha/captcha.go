package captcha

import (
	"blog/pkg/cache"
	"blog/pkg/email"
	"errors"
	"strings"

	"go.uber.org/zap"
)

// # 验证码配置
type CaptchaConfig struct {
	Length int    `mapstructure:"length"`
	Expiry int    `mapstructure:"expiry"`
	Option string `mapstructure:"option"` // 验证码管理器类型，目前支持dev、email

	EmailSubject  string `mapstructure:"emailSubject"`
	EmailTemplate string `mapstructure:"emailTemplate"` // 验证码邮件模板文件路径

	// SMS相关配置， 没有实现
	// SmsSubject   string `mapstructure:"smsSubject"`
	// SmsTemplate string `mapstructure:"smsTemplate"`
}

type CaptchaOption struct {
	// EmailService 用于发送验证码的邮箱服务
	EmailService email.EmailService
	// Cache 用于存储验证码的缓存
	Cache cache.Cache
	// Logger 用于记录日志
	Logger *zap.Logger
}

func InitCaptcha(config CaptchaConfig, options CaptchaOption) (*CaptchaManager, error) {
	// 配置验证
	if config.Length < 4 {
		return nil, errors.New("验证码长度至少为4位")
	}
	if config.Expiry <= 0 {
		return nil, errors.New("验证码过期时间必须大于0")
	}
	if options.Logger == nil {
		options.Logger = zap.NewNop()
	}
	if options.Cache == nil {
		return nil, errors.New("Cache is nil")
	}

	captchaManagerType := strings.ToLower(config.Option)
	switch captchaManagerType {
	case "dev":
		provider := NewDevProvider()
		return NewCaptchaManager(provider, config.Expiry, ManagerOptions{
			Cache:  options.Cache,
			Logger: options.Logger,
		})
	case "email":
		if options.EmailService == nil {
			return nil, errors.New("EmailService is nil")
		}
		provider, err := NewEmailProvider(options.EmailService, config)
		if err != nil {
			return nil, err
		}
		return NewCaptchaManager(provider, config.Expiry, ManagerOptions{
			Cache:  options.Cache,
			Logger: options.Logger,
		})
	default:
		return nil, errors.New("Invalid captcha manager type: " + captchaManagerType)
	}
}
