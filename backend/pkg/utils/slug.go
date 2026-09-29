package utils

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// GenerateSlug 根据标题生成 SEO 友好的 URL 别名（slug）。
// 规则：
//   - 保留字母与数字（含 Unicode，如中文）
//   - 连续的非字母数字字符替换为单个连字符 "-"
//   - 若生成结果为空，则回退为 "post-<时间戳>"，保证唯一可读
func GenerateSlug(title string) string {
	title = strings.TrimSpace(title)
	var sb strings.Builder
	lastDash := false // 标记上一个字符是否为连字符，用于折叠连续分隔符

	for _, r := range title {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(unicode.ToLower(r))
			lastDash = false
		} else {
			if !lastDash && sb.Len() > 0 {
				sb.WriteRune('-')
				lastDash = true
			}
		}
	}

	slug := strings.Trim(sb.String(), "-")
	if slug == "" {
		// 标题无可用于生成 slug 的字符时（如纯符号标题），追加时间戳保证唯一
		slug = fmt.Sprintf("post-%d", time.Now().UnixNano())
	}
	return slug
}
