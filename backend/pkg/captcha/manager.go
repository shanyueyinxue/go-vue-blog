package captcha

import (
	"time"

	"blog/pkg/cache"
	"blog/pkg/utils"

	"go.uber.org/zap"
)

// CaptchaEntry 缓存中存储的验证码条目
type CaptchaEntry struct {
	Code      string
	ExpiresAt int64 // Unix时间戳
}

// ManagerOptions 管理器选项
type ManagerOptions struct {
	// Cache 用于存储验证码的缓存
	Cache cache.Cache
	// Logger 用于记录日志
	Logger *zap.Logger
}

// CaptchaManager 通用验证码管理器
type CaptchaManager struct {
	provider CaptchaProvider
	cache    cache.Cache
	logger   *zap.Logger
	expiry   int // 过期时间（秒）
}

// NewCaptchaManager 创建通用验证码管理器
func NewCaptchaManager(provider CaptchaProvider, expiry int, options ManagerOptions) (*CaptchaManager, error) {
	if options.Logger == nil {
		options.Logger = zap.NewNop()
	}
	if options.Cache == nil {
		return nil, ErrStoreCode
	}

	return &CaptchaManager{
		provider: provider,
		cache:    options.Cache,
		logger:   options.Logger,
		expiry:   expiry,
	}, nil
}

// key 生成缓存键
func (m *CaptchaManager) key(target string) string {
	return "captcha:" + m.provider.GetType() + ":" + target
}

// GenerateCode 生成验证码（委托给提供者）
func (m *CaptchaManager) GenerateCode() string {
	return m.provider.GenerateCode()
}

// StoreCode 将验证码存入缓存，直接存储字符串，利用缓存 TTL 自动过期
func (m *CaptchaManager) StoreCode(code string, target string) CaptchaError {
	key := m.key(target)
	err := m.cache.Set(key, code, time.Duration(m.expiry)*time.Second)
	if err != nil {
		m.logger.Error("store code error - cache set failed", zap.Error(err))
		return ErrStoreCode
	}
	return nil
}

// VerifyCode 验证验证码，直接从缓存读取，成功则立即删除
func (m *CaptchaManager) VerifyCode(inputCode string, target string) (bool, CaptchaError) {
	key := m.key(target)
	storedCode, err := m.cache.GetString(key)
	if err != nil {
		// 键不存在或已过期，视为验证失败
		m.logger.Warn("verify code error - cache get failed", zap.Error(err))
		return false, ErrVerifyCode
	}

	if storedCode == inputCode {
		// 验证成功，立即删除验证码，防止重复使用
		_ = m.cache.Delete(key)
		return true, nil
	}
	return false, nil
}

// DeleteCode 主动删除验证码，若传入的 code 与缓存中一致则删除
func (m *CaptchaManager) DeleteCode(code string, target string) CaptchaError {
	key := m.key(target)
	storedCode, err := m.cache.GetString(key)
	if err != nil {
		// 缓存未命中或已过期，无需删除
		return nil
	}
	if storedCode == code {
		return m.cache.Delete(key)
	}
	return nil
}

// DeleteByTarget 根据目标删除验证码
func (m *CaptchaManager) DeleteByTarget(target string) CaptchaError {
	key := m.key(target)
	return m.cache.Delete(key)
}

// SendCode 发送验证码
func (m *CaptchaManager) SendCode(code, name, target string) CaptchaError {
	if !m.provider.ValidateFormat(target) {
		return ErrFormat
	}

	// 发送前先删除同一目标的旧验证码
	_ = m.DeleteByTarget(target)

	// 存储新验证码
	if err := m.StoreCode(code, target); err != nil {
		return err
	}

	// 发送验证码
	val := &CaptchaVal{
		Subject:      "验证码",
		Username:     name,
		Code:         code,
		Expiry:       m.expiry,
		ValidMinutes: utils.SecondsToMinutes(m.expiry),
		Note:         "请在有效时间内输入验证码以完成验证",
	}
	if err := m.provider.Send(target, val); err != nil {
		m.logger.Error("send code error", zap.Error(err))
		return ErrSendCode
	}
	return nil
}
