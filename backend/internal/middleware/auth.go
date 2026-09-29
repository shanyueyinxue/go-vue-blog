package middleware

import (
	"strings"

	"blog/internal/model"
	"blog/pkg/jwt"
	"blog/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 保存到上下文中的用户 key
const (
	CtxUserKey = "currentUser" // 当前登录用户对象
	CtxUserID  = "currentUserID"
)

// MasterOnly 超级管理员权限中间件：
// 仅放行 is_master=1 的用户，其余返回 403。必须注册在 JWTAuth 之后使用。
func MasterOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, ok := c.Get(CtxUserKey)
		if !ok {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		user := userVal.(*model.User)
		if user.IsMaster != 1 {
			response.Forbidden(c, "无权限：仅超级管理员可操作")
			c.Abort()
			return
		}
		c.Next()
	}
}

// JWTAuth JWT 认证中间件：
// 校验 Authorization: Bearer <token>，将登录用户信息注入上下文。
// 仅放行 access token（issuer 需等于配置的 issuer，refresh token 不可用于访问接口）。
func JWTAuth(jwtSvc *jwt.JWT, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从请求头提取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "未登录或 Token 已失效")
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "Authorization 头格式错误")
			c.Abort()
			return
		}
		tokenStr := parts[1]

		// 2. 解析 token
		claims, err := jwtSvc.ParseToken(tokenStr)
		if err != nil {
			response.Unauthorized(c, "Token 无效或已过期")
			c.Abort()
			return
		}
		// 3. 校验 issuer：拒绝 refresh token 用于访问接口
		if claims.Issuer != jwtSvc.Config.Issuer {
			response.Unauthorized(c, "Token 类型错误，请使用访问 Token")
			c.Abort()
			return
		}

		// 4. 从数据库加载用户，确认仍有效（未被禁用/删除）
		var user model.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			response.Unauthorized(c, "用户不存在")
			c.Abort()
			return
		}
		if user.Status != 1 {
			response.Forbidden(c, "账号已被禁用")
			c.Abort()
			return
		}
		if user.PwdVersion != claims.PwdVersion {
			response.Unauthorized(c, "登录状态已失效，请重新登录")
			c.Abort()
			return
		}

		// 5. 注入上下文供后续 handler 使用
		c.Set(CtxUserKey, &user)
		c.Set(CtxUserID, user.ID)
		c.Next()
	}
}
