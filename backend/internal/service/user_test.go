package service

import (
	"testing"

	"blog/internal/dto"
	"blog/internal/model"
	"blog/pkg/response"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// errCode 提取业务错误码，非 BizError 返回 -1
func errCode(err error) int {
	if be, ok := err.(*BizError); ok {
		return be.Code
	}
	return -1
}

// createUser 直接落库一个测试用户（bcrypt 密码），返回模型
func createUser(t *testing.T, db *gorm.DB, username, password, email string, isMaster int) *model.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.Nil(t, err)
	user := model.User{
		Username: username,
		Password: string(hash),
		Email:    email,
		Status:   1,
		IsMaster: isMaster,
	}
	require.NoError(t, db.Create(&user).Error)
	return &user
}

// TestUserUpdateProfile 修改个人信息：正常修改、清空邮箱、用户名/邮箱唯一性冲突
func TestUserUpdateProfile(t *testing.T) {
	a := newTestApp(t)
	base := New(a)
	userSvc := NewUserService(base, NewCaptchaService(base))

	user := createUser(t, a.DB, "alice", "password123", "alice@example.com", 0)
	_ = createUser(t, a.DB, "bob", "password123", "bob@example.com", 0)

	// 正常修改：用户名/邮箱/手机号/头像
	updated, err := userSvc.UpdateProfile(user, &dto.ProfileUpdateRequest{
		Username: strPtr("alice_new"),
		Email:    strPtr("alice_new@example.com"),
		Phone:    strPtr("13800138000"),
		Avatar:   strPtr("https://example.com/a.png"),
	})
	require.Nil(t, err)
	require.Equal(t, "alice_new", updated.Username)
	require.Equal(t, "alice_new@example.com", updated.Email)
	require.Equal(t, "13800138000", updated.Phone)
	require.Equal(t, "https://example.com/a.png", updated.Avatar)

	// 清空邮箱与手机号（空串清除）
	updated, err = userSvc.UpdateProfile(user, &dto.ProfileUpdateRequest{
		Email: strPtr(""),
		Phone: strPtr(""),
	})
	require.Nil(t, err)
	require.Equal(t, "", updated.Email)
	require.Equal(t, "", updated.Phone)

	// 用户名被占用 → 冲突
	_, err = userSvc.UpdateProfile(user, &dto.ProfileUpdateRequest{Username: strPtr("bob")})
	require.Equal(t, response.CodeConflict, errCode(err))

	// 邮箱被占用 → 冲突
	_, err = userSvc.UpdateProfile(user, &dto.ProfileUpdateRequest{Email: strPtr("bob@example.com")})
	require.Equal(t, response.CodeConflict, errCode(err))

	// 空用户名 → 参数错误
	_, err = userSvc.UpdateProfile(user, &dto.ProfileUpdateRequest{Username: strPtr("  ")})
	require.Equal(t, response.CodeParamError, errCode(err))

	// 空请求体 → 参数错误
	_, err = userSvc.UpdateProfile(user, &dto.ProfileUpdateRequest{})
	require.Equal(t, response.CodeParamError, errCode(err))
}

// TestForgotPasswordFlow 忘记密码全流程：
// 发码 → 错误验证码拒绝 → 正确验证码重置 → 旧密码失效 / 新密码生效 / pwd_version 递增
func TestForgotPasswordFlow(t *testing.T) {
	a := newTestApp(t)
	base := New(a)
	userSvc := NewUserService(base, NewCaptchaService(base))

	createUser(t, a.DB, "alice", "old-password-1", "alice@example.com", 0)

	// 发码（dev 验证码固定 1234）
	token, expiresIn, err := userSvc.SendForgotPasswordCode("alice@example.com")
	require.Nil(t, err)
	require.NotEmpty(t, token)
	require.Greater(t, expiresIn, 0)

	// 错误验证码 → 参数错误
	err = userSvc.ResetForgotPassword(&dto.ForgotPasswordResetRequest{
		Email:       "alice@example.com",
		Token:       token,
		Code:        "0000",
		NewPassword: "new-password-1",
	})
	require.Equal(t, response.CodeParamError, errCode(err))

	// 正确验证码 → 重置成功
	err = userSvc.ResetForgotPassword(&dto.ForgotPasswordResetRequest{
		Email:       "alice@example.com",
		Token:       token,
		Code:        "1234",
		NewPassword: "new-password-1",
	})
	require.Nil(t, err)

	// 验证密码已更新且 pwd_version 递增（旧 Token 失效）
	var user model.User
	require.NoError(t, a.DB.Where("username = ?", "alice").First(&user).Error)
	require.Equal(t, uint(1), user.PwdVersion)
	require.Error(t, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("old-password-1")))
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("new-password-1")))

	// 验证码已一次性作废：同一 token 不可再次重置
	err = userSvc.ResetForgotPassword(&dto.ForgotPasswordResetRequest{
		Email:       "alice@example.com",
		Token:       token,
		Code:        "1234",
		NewPassword: "another-password",
	})
	require.Equal(t, response.CodeParamError, errCode(err))

	// 邮箱未注册：返回成功响应（假 token），不发码不报错（防用户枚举）
	token2, _, err := userSvc.SendForgotPasswordCode("nobody@example.com")
	require.Nil(t, err)
	require.Empty(t, token2)

	// 新密码过短 → 参数错误
	err = userSvc.ResetForgotPassword(&dto.ForgotPasswordResetRequest{
		Email:       "alice@example.com",
		Token:       token,
		Code:        "1234",
		NewPassword: "short",
	})
	require.Equal(t, response.CodeParamError, errCode(err))
}

// TestUserMasterProtection 保护规则：禁止删除/禁用/降级当前登录账号
func TestUserMasterProtection(t *testing.T) {
	a := newTestApp(t)
	base := New(a)
	userSvc := NewUserService(base, NewCaptchaService(base))

	master := createUser(t, a.DB, "master", "master-pass-1", "master@example.com", 1)
	normal := createUser(t, a.DB, "normal", "normal-pass-1", "normal@example.com", 0)

	// 删除自己 → 禁止
	err := userSvc.DeleteUser(master.ID, master)
	require.Equal(t, response.CodeForbidden, errCode(err))

	// 禁用自己 → 禁止
	zero := 0
	_, err = userSvc.UpdateUser(master.ID, &dto.UserUpdateRequest{Status: &zero}, master)
	require.Equal(t, response.CodeForbidden, errCode(err))

	// 取消自己的 master → 禁止
	_, err = userSvc.UpdateUser(master.ID, &dto.UserUpdateRequest{IsMaster: &zero}, master)
	require.Equal(t, response.CodeForbidden, errCode(err))

	// 删除其他用户 → 允许
	err = userSvc.DeleteUser(normal.ID, master)
	require.Nil(t, err)
	var count int64
	a.DB.Model(&model.User{}).Where("id = ?", normal.ID).Count(&count)
	require.Zero(t, count)
}

// TestUserCRUD 用户管理：创建 → 列表（关键字）→ 详情 → 更新 → 重置密码 → 删除
func TestUserCRUD(t *testing.T) {
	a := newTestApp(t)
	base := New(a)
	userSvc := NewUserService(base, NewCaptchaService(base))

	master := createUser(t, a.DB, "master", "master-pass-1", "master@example.com", 1)

	// 创建用户（默认非 master）
	created, err := userSvc.CreateUser(&dto.UserCreateRequest{
		Username: "editor",
		Password: "editor-pass-1",
		Email:    strPtr("editor@example.com"),
		Phone:    strPtr("13900139000"),
		Role:     strPtr("editor"),
	})
	require.Nil(t, err)
	require.Equal(t, 0, created.IsMaster)
	require.Equal(t, "editor", created.Role)
	require.Equal(t, 1, created.Status)

	// 重复用户名 → 冲突
	_, err = userSvc.CreateUser(&dto.UserCreateRequest{Username: "editor", Password: "editor-pass-2"})
	require.Equal(t, response.CodeConflict, errCode(err))

	// 密码过短 → 参数错误
	_, err = userSvc.CreateUser(&dto.UserCreateRequest{Username: "short", Password: "1234567"})
	require.Equal(t, response.CodeParamError, errCode(err))

	// 列表：关键字搜索
	kwReq := &dto.UserQuery{Keyword: "edito"}
	kwReq.Normalize()
	users, total, err := userSvc.ListUsers(kwReq)
	require.Nil(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, users, 1)
	require.Equal(t, "editor", users[0].Username)

	// 列表：空关键字返回全部
	allReq := &dto.UserQuery{}
	allReq.Normalize()
	users, total, err = userSvc.ListUsers(allReq)
	require.Nil(t, err)
	require.Equal(t, int64(2), total)

	// 详情
	got, err := userSvc.GetUser(created.ID)
	require.Nil(t, err)
	require.Equal(t, "editor", got.Username)

	// 更新：改名 + 提升为 master
	one := 1
	updated, err := userSvc.UpdateUser(created.ID, &dto.UserUpdateRequest{
		Username: strPtr("editor_v2"),
		IsMaster: &one,
	}, master)
	require.Nil(t, err)
	require.Equal(t, "editor_v2", updated.Username)
	require.Equal(t, 1, updated.IsMaster)

	// 更新不存在的用户 → 404
	_, err = userSvc.UpdateUser(99999, &dto.UserUpdateRequest{Username: strPtr("x")}, master)
	require.Equal(t, response.CodeNotFound, errCode(err))

	// 重置密码（无需原密码，pwd_version 递增）
	err = userSvc.ResetUserPassword(created.ID, "new-editor-pass")
	require.Nil(t, err)
	var after model.User
	require.NoError(t, a.DB.First(&after, created.ID).Error)
	require.Equal(t, uint(1), after.PwdVersion)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(after.Password), []byte("new-editor-pass")))

	// 删除
	err = userSvc.DeleteUser(created.ID, master)
	require.Nil(t, err)
	_, err = userSvc.GetUser(created.ID)
	require.Equal(t, response.CodeNotFound, errCode(err))
}
