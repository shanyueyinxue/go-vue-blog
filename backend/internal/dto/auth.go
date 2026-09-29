package dto

import "blog/internal/model"

// LoginRequest 登录请求体
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 管理员用户名
	Password string `json:"password" binding:"required"` // 密码
}

// RefreshTokenRequest 刷新 Token 请求体
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"` // 刷新令牌
}

// UpdatePasswordRequest 修改密码请求体
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"` // 原密码
	NewPassword string `json:"new_password" binding:"required"` // 新密码
}

// LoginResult 登录/刷新 Token 返回结果
type LoginResult struct {
	Token        string      // 访问令牌
	RefreshToken string      // 刷新令牌
	ExpiresIn    int         // 有效期（秒）
	User         *model.User // 用户信息（登录时返回，刷新时为空）
}
