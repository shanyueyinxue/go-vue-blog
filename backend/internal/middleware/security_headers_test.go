package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// TestSecurityHeaders 校验全局安全响应头中间件输出 X-Content-Type-Options: nosniff，
// 对普通 API 路由与静态文件路由均生效。
func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 构造一个临时静态目录（模拟 /uploads），并放置一个内容为文本、扩展名为 .html 的文件，
	// 验证即使响应 Content-Type 是 text/html，也携带 nosniff 禁止浏览器嗅探/执行
	staticDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(staticDir, "legacy.html"), []byte("<script>alert(1)</script>"), 0o644))

	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.Static("/uploads", staticDir)

	t.Run("API 路由", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	})

	t.Run("静态文件路由", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/uploads/legacy.html", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		require.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
	})
}
