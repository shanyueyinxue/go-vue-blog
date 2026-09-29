package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders 全局安全响应头中间件。
//
// 当前输出：
//   - X-Content-Type-Options: nosniff —— 禁止浏览器对响应内容做 MIME 嗅探。
//     即使上传目录中残留了内容与扩展名不匹配的文件（如文本伪装成 .html），
//     浏览器也不会将其当作 text/html 解析执行，与文件上传扩展名白名单
//     （internal/service/file.go 的 textMimeExts）形成纵深防御。
//
// 注意：该中间件注册在全局中间件链（router.go），对包括 /uploads 静态文件
// 在内的所有响应生效。
// 可按需扩展其它安全头（Content-Security-Policy / X-Frame-Options /
// Referrer-Policy / Strict-Transport-Security 等）。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Next()
	}
}
