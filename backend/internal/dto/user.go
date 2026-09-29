package dto

import "blog/pkg/response"

// ProfileUpdateRequest 修改个人信息请求体（登录用户本人，仅允许修改基础资料，
// 不允许修改 is_master / status / role）
type ProfileUpdateRequest struct {
	Username *string `json:"username"` // 登录名（*string：nil 不修改，空串不合法）
	Email    *string `json:"email"`    // 邮箱（nil 不修改，空串清除）
	Phone    *string `json:"phone"`    // 手机号（nil 不修改，空串清除）
	Avatar   *string `json:"avatar"`   // 头像地址（nil 不修改，空串清除）
}

// ForgotPasswordSendRequest 忘记密码-发送邮箱验证码请求体（公开接口）
type ForgotPasswordSendRequest struct {
	Email string `json:"email" binding:"required"` // 注册邮箱
}

// ForgotPasswordResetRequest 忘记密码-重置密码请求体（公开接口）
type ForgotPasswordResetRequest struct {
	Email       string `json:"email" binding:"required"`        // 注册邮箱
	Token       string `json:"token" binding:"required"`        // 发送验证码时返回的 token
	Code        string `json:"code" binding:"required"`         // 邮箱验证码
	NewPassword string `json:"new_password" binding:"required"` // 新密码（≥ 8 位）
}

// UserCreateRequest 创建用户请求体（仅超级管理员）
type UserCreateRequest struct {
	Username string  `json:"username" binding:"required"` // 登录名（必填，唯一）
	Password string  `json:"password" binding:"required"` // 初始密码（必填，≥ 8 位）
	Email    *string `json:"email"`                       // 邮箱（可空）
	Phone    *string `json:"phone"`                       // 手机号（可空）
	Avatar   *string `json:"avatar"`                      // 头像地址（可空）
	Role     *string `json:"role"`                        // 角色，默认 admin
	Status   *int    `json:"status"`                      // 状态：1 启用 / 0 禁用，默认 1
	IsMaster *int    `json:"is_master"`                   // 是否超级管理员：1 / 0，默认 0
}

// UserUpdateRequest 更新用户请求体（仅超级管理员，nil 字段不修改）
type UserUpdateRequest struct {
	Username *string `json:"username"` // 登录名（nil 不修改，空串不合法）
	Email    *string `json:"email"`    // 邮箱（nil 不修改，空串清除）
	Phone    *string `json:"phone"`    // 手机号（nil 不修改，空串清除）
	Avatar   *string `json:"avatar"`   // 头像地址（nil 不修改，空串清除）
	Role     *string `json:"role"`     // 角色（nil 不修改）
	Status   *int    `json:"status"`   // 状态：1 启用 / 0 禁用（nil 不修改）
	IsMaster *int    `json:"is_master"` // 是否超级管理员：1 / 0（nil 不修改）
}

// UserPasswordResetRequest 重置用户密码请求体（仅超级管理员，无需原密码）
type UserPasswordResetRequest struct {
	NewPassword string `json:"new_password" binding:"required"` // 新密码（≥ 8 位）
}

// UserQuery 用户列表查询参数
type UserQuery struct {
	response.PageRequest
	Keyword string `form:"keyword"` // 用户名/邮箱关键字模糊搜索
}
