package jwt

import "testing"

func TestValidateSecret(t *testing.T) {
	weak := []string{
		"",
		"secret",
		"change-me",
		"change_me",
		"password",
		"123456",
		"1234567890",
		"jwt-secret",
		"short", // 长度不足 16
	}
	for _, s := range weak {
		if err := ValidateSecret(s); err == nil {
			t.Errorf("ValidateSecret(%q) 应拒绝弱密钥", s)
		}
	}

	strong := []string{
		"0123456789abcdef",        // 恰好 16 位
		"a-very-long-random-42xyz", // 随机强密钥
	}
	for _, s := range strong {
		if err := ValidateSecret(s); err != nil {
			t.Errorf("ValidateSecret(%q) 不应拒绝强密钥: %v", s, err)
		}
	}
}
