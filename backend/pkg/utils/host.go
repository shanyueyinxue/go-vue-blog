package utils

import (
	"net"
	"regexp"
	"strings"
)

// IsValidHost 判断 host 是否合法。
// 合法情况包括：
// 1. 有效的 IPv4 或 IPv6 地址（如 "127.0.0.1"、"::1"）
// 2. 有效的域名（如 "example.com"、"sub.example.com"）
func IsValidHost(host string) bool {
	// 空字符串不合法
	if host == "" {
		return false
	}

	// 1. 检查是否为合法 IP 地址
	if net.ParseIP(host) != nil {
		return true
	}

	// 2. 检查是否为合法域名（RFC 1035 简化规则）
	// 总长度不能超过 253 个字符
	if len(host) > 253 {
		return false
	}

	// 域名正则：允许字母、数字、连字符，点分隔，不以连字符开头或结尾
	// 每个标签长度 1-63，总长度已在上面限制
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !domainRegex.MatchString(host) {
		return false
	}

	// 进一步检查每个标签的长度（正则已限制最大 63，这里再确保一下）
	labels := strings.Split(host, ".")
	for _, label := range labels {
		if len(label) > 63 {
			return false
		}
	}

	return true
}
