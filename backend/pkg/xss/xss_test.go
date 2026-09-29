package xss

import (
	"strings"
	"testing"
)

func TestSanitize(t *testing.T) {
	tests := []struct {
		name string
		in   string
		// want 期望输出；wantContains 非空时仅断言包含
		want         string
		wantContains string
	}{
		{
			name: "纯文本原样保留",
			in:   "hello world",
			want: "hello world",
		},
		{
			name: "白名单标签保留",
			in:   "<p>你好 <strong>世界</strong></p>",
			want: "<p>你好 <strong>世界</strong></p>",
		},
		{
			name: "非白名单标签被剥离但保留文本",
			in:   "<script>alert(1)</script>正文",
			want: "alert(1)正文",
		},
		{
			name: "事件属性被移除",
			in:   `<img src="/img/x.png" onerror="alert(1)">`,
			want: `<img src="/img/x.png"></img>`,
		},
		{
			name: "javascript 协议被移除",
			in:   `<a href="javascript:alert(1)">点我</a>`,
			want: "<a>点我</a>",
		},
		{
			name: "XSS 绕过：实体编码的标签不会被还原成标签",
			// 修复前：&lt;img ...&gt; 会被 html.Parse 解码成文本 <img ...> 再原样写出，
			// 最终输出可执行的标签；修复后文本节点重新转义，保持安全。
			in:           `&lt;img src=x onerror=alert(1)&gt;`,
			wantContains: "&lt;img",
		},
		{
			name: "实体编码的标签输出后不再包含原始标签",
			in:   `&lt;img src=x onerror=alert(1)&gt;`,
			want: `&lt;img src=x onerror=alert(1)&gt;`,
		},
		{
			name: "合法实体往返不丢内容",
			in:   "<p>a &amp; b</p>",
			want: "<p>a &amp; b</p>",
		},
		{
			name: "注释被移除",
			in:   "<!-- 注释 -->正文",
			want: "正文",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sanitize(tt.in)
			if tt.wantContains != "" {
				if !strings.Contains(got, tt.wantContains) {
					t.Fatalf("Sanitize(%q) = %q，期望包含 %q", tt.in, got, tt.wantContains)
				}
				return
			}
			if got != tt.want {
				t.Fatalf("Sanitize(%q) = %q，期望 %q", tt.in, got, tt.want)
			}
		})
	}
}
