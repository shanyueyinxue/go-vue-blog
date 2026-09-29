package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	hosts := []string{
		"127.0.0.1",          // IPv4 地址 → true
		"::1",                // IPv6 地址 → true
		"example.com",        // 域名 → true
		"sub-domain.example", // 含连字符的域名 → true
		"localhost",          // 单标签域名 → true
		"-abc.com",           // 以连字符开头 → false
		"abc-.com",           // 以连字符结尾 → false
		"a..b",               // 连续点 → false
		"verylonglabel" + strings.Repeat("a", 64) + ".com", // 标签超长 → false
		"", // 空字符串 → false
	}

	for _, h := range hosts {
		t.Logf("%q → %v\n", h, IsValidHost(h))
	}
	assert.Equal(t, true, IsValidHost("127.0.0.1"))
	assert.Equal(t, true, IsValidHost("::1"))
	assert.Equal(t, true, IsValidHost("example.com"))
	assert.Equal(t, true, IsValidHost("sub-domain.example"))
	assert.Equal(t, true, IsValidHost("localhost"))
	assert.Equal(t, false, IsValidHost("-abc.com"))
	assert.Equal(t, false, IsValidHost("abc-.com"))
	assert.Equal(t, false, IsValidHost("a..b"))
	assert.Equal(t, false, IsValidHost("verylonglabel"+strings.Repeat("a", 64)+".com"))
	assert.Equal(t, false, IsValidHost(""))
}
