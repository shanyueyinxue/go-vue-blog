package xss

import (
	"strings"

	"golang.org/x/net/html"
)

// 允许保留的 HTML 标签白名单（仅包含常见的富文本/评论安全标签）
var allowedTags = map[string]bool{
	"p": true, "br": true, "hr": true,
	"b": true, "strong": true, "i": true, "em": true, "u": true, "s": true,
	"code": true, "pre": true, "blockquote": true,
	"ul": true, "ol": true, "li": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"a": true, "img": true, "table": true, "thead": true, "tbody": true, "tr": true, "th": true, "td": true,
}

// 允许保留的属性白名单
var allowedAttrs = map[string]map[string]bool{
	"a":   {"href": true, "title": true, "target": true, "rel": true},
	"img": {"src": true, "alt": true, "title": true, "width": true, "height": true},
	"th":  {"colspan": true, "rowspan": true, "align": true},
	"td":  {"colspan": true, "rowspan": true, "align": true},
}

// 允许的协议白名单（用于 href/src 属性，防止 javascript: 等危险协议）
var allowedSchemes = []string{"http", "https", "mailto", "tel", "ftp"}

// Sanitize 过滤 HTML 字符串中的危险内容（XSS 防护）。
// 移除白名单之外的标签与属性，并转义危险协议链接。
// 该函数同时兼容纯文本输入（不做任何改动）。
func Sanitize(input string) string {
	if strings.TrimSpace(input) == "" {
		return input
	}

	// 解析 HTML，即使输入不是严格合法的 HTML 也能容错处理
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		// 解析失败时退化为纯文本（去除所有标签）
		return html.EscapeString(input)
	}

	var sb strings.Builder
	sanitizeNode(doc, &sb)
	return sb.String()
}

// sanitizeNode 递归遍历节点，仅保留白名单标签
func sanitizeNode(n *html.Node, sb *strings.Builder) {
	switch n.Type {
	case html.TextNode:
		// html.Parse 会把实体解码（如 &lt; 还原成 <），直接原样写出会把
		// "看似纯文本、实则可执行" 的内容重新变成标签，导致 XSS 绕过。
		// 因此文本节点必须重新转义后再输出。
		sb.WriteString(html.EscapeString(n.Data))
	case html.ElementNode:
		tag := n.Data
		if allowedTags[tag] {
			// 输出白名单内的标签
			sb.WriteString("<")
			sb.WriteString(tag)
			for _, attr := range n.Attr {
				if attrAllowed(tag, attr) {
					sb.WriteString(" ")
					sb.WriteString(attr.Key)
					sb.WriteString(`="`)
					sb.WriteString(html.EscapeString(attr.Val))
					sb.WriteString(`"`)
				}
			}
			sb.WriteString(">")
			// 递归处理子节点
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				sanitizeNode(child, sb)
			}
			// 输出闭合标签
			sb.WriteString("</")
			sb.WriteString(tag)
			sb.WriteString(">")
		} else {
			// 非白名单标签：丢弃标签本身，仅保留其文本内容
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				sanitizeNode(child, sb)
			}
		}
	default:
		// 注释、DOCTYPE 等其他节点直接丢弃
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			sanitizeNode(child, sb)
		}
	}
}

// attrAllowed 判断属性是否允许保留
func attrAllowed(tag string, attr html.Attribute) bool {
	allowed, ok := allowedAttrs[tag]
	if !ok {
		return false
	}
	if !allowed[attr.Key] {
		return false
	}
	// 检查危险协议（href / src 属性）
	if attr.Key == "href" || attr.Key == "src" {
		if !schemeAllowed(attr.Val) {
			return false
		}
	}
	return true
}

// schemeAllowed 检查 URL 的协议是否在安全协议白名单内
func schemeAllowed(url string) bool {
	lower := strings.ToLower(strings.TrimSpace(url))
	for _, scheme := range allowedSchemes {
		if strings.HasPrefix(lower, scheme+"://") {
			return true
		}
	}
	// 相对路径或锚点也是安全的
	if strings.HasPrefix(lower, "/") || strings.HasPrefix(lower, "#") || lower == "" {
		return true
	}
	return false
}
