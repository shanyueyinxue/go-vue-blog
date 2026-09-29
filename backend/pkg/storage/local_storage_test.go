package storage

import (
	"context"
	"testing"
)

func TestLocalStorageGetURL(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name    string
		baseURL string
		path    string
		want    string
	}{
		{"baseURL 为空时返回根路径", "", "img/a.png", "/img/a.png"},
		{"相对 baseURL 补前导斜杠（修复：避免子路由下解析成相对路径）", "uploads", "img/a.png", "/uploads/img/a.png"},
		{"带前导斜杠的 baseURL 保持不变", "/uploads", "img/a.png", "/uploads/img/a.png"},
		{"baseURL 为完整域名（CDN）", "https://cdn.example.com", "img/a.png", "https://cdn.example.com/img/a.png"},
		{"path 本身是完整 URL 直接返回", "", "https://cdn.example.com/x.png", "https://cdn.example.com/x.png"},
		{"path 前导斜杠被规范化", "uploads", "/img/a.png", "/uploads/img/a.png"},
		{"空 path 返回空", "uploads", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewLocalStorage("/tmp/uploads", tc.baseURL)
			if got := s.GetURL(ctx, tc.path); got != tc.want {
				t.Fatalf("GetURL(%q) = %q，期望 %q", tc.path, got, tc.want)
			}
		})
	}
}
