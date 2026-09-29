package utils

import (
	"regexp"
	"strings"
)

// IsEmail 检查字符串是否为有效的邮箱格式
func IsEmail(str string) bool {
	// 额外检查：不能以点号开头或结尾，不能有连续点号
	if strings.HasPrefix(str, ".") || strings.HasSuffix(str, ".") {
		return false
	}
	if strings.Contains(str, "..") {
		return false
	}
	// 正则表达式验证邮箱格式
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, str)
	return matched
}

// IsPhoneInChina 检查字符串是否为有效的手机号格式（中国手机号）
func IsPhoneInChina(str string) bool {
	// 中国手机号正则：1开头，第二位为3-9，总共11位数字
	phoneRegex := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(phoneRegex, str)
	return matched
}
