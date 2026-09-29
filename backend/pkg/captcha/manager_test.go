package captcha

import (
	"testing"
	"time"

	"blog/pkg/cache"
)

// mockCache 模拟缓存实现
type mockCache struct {
	data map[string]mockCacheEntry
}

type mockCacheEntry struct {
	value      any
	expiration time.Time
}

func newMockCache() *mockCache {
	return &mockCache{
		data: make(map[string]mockCacheEntry),
	}
}

func (m *mockCache) Set(key string, value any, expiration time.Duration) error {
	m.data[key] = mockCacheEntry{
		value:      value,
		expiration: time.Now().Add(expiration),
	}
	return nil
}

func (m *mockCache) Get(key string) (any, error) {
	entry, ok := m.data[key]
	if !ok {
		return nil, cache.ErrKeyNotFound
	}
	if time.Now().After(entry.expiration) {
		delete(m.data, key)
		return nil, cache.ErrKeyNotFound
	}
	return entry.value, nil
}

func (m *mockCache) GetString(key string) (string, error) {
	val, err := m.Get(key)
	if err != nil {
		return "", err
	}
	if s, ok := val.(string); ok {
		return s, nil
	}
	return "", cache.ErrInvalidType
}

func (m *mockCache) GetInt(key string) (int, error)       { return 0, nil }
func (m *mockCache) GetFloat(key string) (float64, error) { return 0, nil }
func (m *mockCache) GetBool(key string) (bool, error)     { return false, nil }
func (m *mockCache) Delete(key string) error              { delete(m.data, key); return nil }
func (m *mockCache) Exists(key string) bool               { _, ok := m.data[key]; return ok }
func (m *mockCache) Ping() error                          { return nil }
func (m *mockCache) Close() error                         { return nil }
func (m *mockCache) Self() any                            { return m }

func TestGenericCaptchaManager_DevProvider(t *testing.T) {
	expiry := 300 // 5分钟
	options := ManagerOptions{
		Cache:  newMockCache(),
		Logger: nil, // 使用默认
	}

	provider := NewDevProvider()
	manager, err := NewCaptchaManager(provider, expiry, options)
	if err != nil {
		t.Fatalf("创建管理器失败: %v", err)
	}

	// 测试生成验证码
	code := manager.GenerateCode()
	if code != "1234" {
		t.Errorf("期望验证码 '1234', 得到 '%s'", code)
	}

	target := "test@example.com"
	name := "Test User"

	// 测试发送验证码（开发环境不实际发送）
	err = manager.SendCode(code, name, target)
	if err != nil {
		t.Errorf("发送验证码失败: %v", err)
	}

	// 验证验证码
	valid, err := manager.VerifyCode(code, target)
	if !valid {
		t.Errorf("验证码验证失败: %v", err)
	}

	// 验证错误验证码
	valid, err = manager.VerifyCode("wrong", target)
	if valid {
		t.Error("错误验证码不应通过验证")
	}

	// 测试删除验证码
	err = manager.DeleteCode(code, target)
	if err != nil {
		t.Errorf("删除验证码失败: %v", err)
	}

	// 删除后验证应失败
	valid, err = manager.VerifyCode(code, target)
	if valid {
		t.Error("删除后验证码不应通过验证")
	}
}

func TestGenericCaptchaManager_EmailProvider(t *testing.T) {
	// 注意：此测试需要模拟邮件服务，暂时跳过
	t.Skip("需要模拟邮件服务，跳过")
}
