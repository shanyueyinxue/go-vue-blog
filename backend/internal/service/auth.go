package service

import (
	"time"

	"blog/internal/dto"
	"blog/internal/model"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 管理员认证相关业务
type AuthService struct {
	*Service
}

// NewAuthService 创建认证服务
func NewAuthService(base *Service) *AuthService {
	return &AuthService{Service: base}
}

// Login 管理员登录：校验用户与密码，签发访问令牌与刷新令牌
func (s *AuthService) Login(username, password, ip string) (*dto.LoginResult, *BizError) {
	// 1. 查找用户
	var user model.User
	if err := s.DB().Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.Log().Warn("管理员登录失败", zap.String("username", username), zap.String("ip", ip), zap.String("reason", "用户不存在"))
			return nil, NewUnauthorized("用户名或密码错误")
		}
		s.Log().Error("管理员登录查询失败", zap.Error(err), zap.String("username", username), zap.String("ip", ip))
		return nil, NewServerError("查询用户失败")
	}
	if user.Status != 1 {
		s.Log().Warn("管理员登录失败", zap.String("username", username), zap.String("ip", ip), zap.String("reason", "账号已禁用"))
		return nil, NewForbidden("账号已被禁用")
	}

	// 2. 校验密码（bcrypt 比对）
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.Log().Warn("管理员登录失败", zap.String("username", username), zap.String("ip", ip), zap.String("reason", "密码错误"))
		return nil, NewUnauthorized("用户名或密码错误")
	}
	s.Log().Info("管理员登录成功", zap.String("username", username), zap.String("ip", ip))

	// 3. 更新登录信息（IP 与时间）
	now := time.Now()
	s.DB().Model(&user).Updates(map[string]any{
		"last_login_at": now,
		"last_ip":       ip,
	})
	user.LastLoginAt = &now
	user.LastIP = ip

	// 4. 生成访问令牌与刷新令牌
	accessToken, err := s.JWT().GenerateToken(user.ID, user.Username, user.Email, user.PwdVersion)
	if err != nil {
		return nil, NewServerError("生成 Token 失败")
	}
	refreshToken, err := s.JWT().GenerateRefreshToken(user.ID, user.Username, user.Email, user.PwdVersion)
	if err != nil {
		return nil, NewServerError("生成刷新 Token 失败")
	}

	return &dto.LoginResult{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.App.Config.JWT.Expiration,
		User:         &user,
	}, nil
}

// RefreshToken 刷新访问 Token（并轮换刷新令牌）
func (s *AuthService) RefreshToken(refreshToken string) (*dto.LoginResult, *BizError) {
	// 1. 解析刷新令牌
	claims, err := s.JWT().ParseToken(refreshToken)
	if err != nil {
		s.Log().Warn("刷新令牌失败", zap.String("reason", "解析失败"))
		return nil, NewUnauthorized("刷新令牌无效或已过期")
	}
	// 2. 校验 issuer 必须是 refresh token
	if claims.Issuer != s.JWT().GetRefreshIssuer() {
		s.Log().Warn("刷新令牌失败", zap.Uint("user_id", claims.UserID), zap.String("reason", "token 类型错误"))
		return nil, NewUnauthorized("Token 类型错误，请使用刷新令牌")
	}

	// 3. 确认用户仍有效
	var user model.User
	if err := s.DB().First(&user, claims.UserID).Error; err != nil || user.Status != 1 {
		return nil, NewUnauthorized("用户不存在或已被禁用")
	}
	if user.PwdVersion != claims.PwdVersion {
		return nil, NewUnauthorized("登录状态已失效，请重新登录")
	}

	// 4. 签发新访问令牌（并顺带刷新一个 refresh token 用于轮换）
	accessToken, err := s.JWT().GenerateToken(user.ID, user.Username, user.Email, user.PwdVersion)
	if err != nil {
		return nil, NewServerError("生成 Token 失败")
	}
	newRefreshToken, err := s.JWT().GenerateRefreshToken(user.ID, user.Username, user.Email, user.PwdVersion)
	if err != nil {
		return nil, NewServerError("生成刷新 Token 失败")
	}

	return &dto.LoginResult{
		Token:        accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    s.App.Config.JWT.Expiration,
	}, nil
}

// UpdatePassword 修改管理员密码
func (s *AuthService) UpdatePassword(user *model.User, oldPassword, newPassword string) *BizError {
	// 新密码长度校验（建议 ≥ 8 位）
	if len(newPassword) < 8 {
		return NewParamError("新密码长度不能少于 8 位")
	}

	// 校验原密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return NewUnauthorized("原密码错误")
	}

	// 生成新密码哈希并更新
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return NewServerError("密码加密失败")
	}
	// 同时递增密码版本号，让所有旧 Token 立即失效
	updates := map[string]any{
		"password":    string(hash),
		"pwd_version": gorm.Expr("pwd_version + 1"),
	}
	if err := s.DB().Model(user).Updates(updates).Error; err != nil {
		return NewServerError("修改密码失败")
	}
	s.Log().Info("管理员密码已修改", zap.Uint("user_id", user.ID), zap.String("username", user.Username))
	return nil
}
