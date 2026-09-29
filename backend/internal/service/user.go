package service

import (
	"strings"

	"blog/internal/dto"
	"blog/internal/model"
	"blog/pkg/utils"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户相关业务：个人信息、忘记密码、用户管理（仅超级管理员）
type UserService struct {
	*Service
	Captcha *CaptchaService // 邮箱验证码服务（忘记密码复用）
}

// NewUserService 创建用户服务
func NewUserService(base *Service, captchaSvc *CaptchaService) *UserService {
	return &UserService{Service: base, Captcha: captchaSvc}
}

// UpdateProfile 修改个人信息（本人，仅允许基础资料字段）
func (s *UserService) UpdateProfile(user *model.User, req *dto.ProfileUpdateRequest) (*model.User, *BizError) {
	updates := map[string]any{}

	if req.Username != nil {
		name := strings.TrimSpace(*req.Username)
		if name == "" {
			return nil, NewParamError("用户名不能为空")
		}
		if err := s.checkUsernameAvailable(name, user.ID); err != nil {
			return nil, err
		}
		updates["username"] = name
	}
	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if email != "" && !utils.IsEmail(email) {
			return nil, NewParamError("邮箱格式不正确")
		}
		if email != "" {
			if err := s.checkEmailAvailable(email, user.ID); err != nil {
				return nil, err
			}
		}
		updates["email"] = email
	}
	if req.Phone != nil {
		updates["phone"] = strings.TrimSpace(*req.Phone)
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}

	if len(updates) == 0 {
		return nil, NewParamError("没有需要修改的信息")
	}
	if err := s.DB().Model(user).Updates(updates).Error; err != nil {
		if isDuplicateKey(err) {
			return nil, NewConflict("用户名或邮箱已被占用")
		}
		return nil, NewServerError("修改个人信息失败")
	}
	s.DB().First(user, user.ID)
	s.Log().Info("用户修改个人信息", zap.Uint("user_id", user.ID), zap.String("username", user.Username))
	return user, nil
}

// SendForgotPasswordCode 忘记密码-发送邮箱验证码。
// 邮箱未注册时返回与成功一致的响应（防用户枚举），但不实际发送。
func (s *UserService) SendForgotPasswordCode(email string) (string, int, *BizError) {
	if !utils.IsEmail(email) {
		return "", 0, NewParamError("邮箱格式不正确")
	}

	var count int64
	if err := s.DB().Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return "", 0, NewServerError("查询用户失败")
	}
	if count == 0 {
		// 返回假的 token 与相同有效期，避免泄露邮箱是否注册
		s.Log().Warn("忘记密码：邮箱未注册，未发送验证码", zap.String("email", email))
		return "", s.App.Config.Captcha.Expiry, nil
	}

	token, expiresIn, err := s.Captcha.SendEmailCaptcha(&dto.EmailCaptchaRequest{Email: email, Scene: "forgot"})
	if err != nil {
		return "", 0, err
	}
	return token, expiresIn, nil
}

// ResetForgotPassword 忘记密码-校验邮箱验证码并重置密码（使旧 Token 全部失效）
func (s *UserService) ResetForgotPassword(req *dto.ForgotPasswordResetRequest) *BizError {
	if len(req.NewPassword) < 8 {
		return NewParamError("新密码长度不能少于 8 位")
	}
	// 校验验证码：token 绑定邮箱且一次性作废
	if err := s.Captcha.VerifyForEmail(req.Token, req.Code, req.Email); err != nil {
		return err
	}

	var user model.User
	if err := s.DB().Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewParamError("该邮箱未注册")
		}
		return NewServerError("查询用户失败")
	}
	if user.Status != 1 {
		return NewForbidden("账号已被禁用，无法重置密码")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return NewServerError("密码加密失败")
	}
	// 递增密码版本号，使该用户所有旧 Token 立即失效
	if err := s.DB().Model(&user).Updates(map[string]any{
		"password":    string(hash),
		"pwd_version": gorm.Expr("pwd_version + 1"),
	}).Error; err != nil {
		return NewServerError("重置密码失败")
	}
	s.Log().Info("用户通过邮箱验证码重置密码", zap.Uint("user_id", user.ID), zap.String("username", user.Username))
	return nil
}

// ListUsers 用户列表（分页 + 关键字搜索，仅超级管理员）
func (s *UserService) ListUsers(req *dto.UserQuery) ([]model.User, int64, *BizError) {
	q := s.DB().Model(&model.User{})
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		q = q.Where("username LIKE ? OR email LIKE ?", kw, kw)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, NewServerError("查询用户失败")
	}

	var users []model.User
	if err := q.Order("id ASC").Limit(req.PageSize).Offset(req.Offset()).Find(&users).Error; err != nil {
		return nil, 0, NewServerError("查询用户失败")
	}
	return users, total, nil
}

// GetUser 用户详情（仅超级管理员）
func (s *UserService) GetUser(id uint) (*model.User, *BizError) {
	var user model.User
	if err := s.DB().First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("用户不存在")
		}
		return nil, NewServerError("查询用户失败")
	}
	return &user, nil
}

// CreateUser 创建用户（仅超级管理员，默认非 master）
func (s *UserService) CreateUser(req *dto.UserCreateRequest) (*model.User, *BizError) {
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return nil, NewParamError("用户名不能为空")
	}
	if len(req.Password) < 8 {
		return nil, NewParamError("密码长度不能少于 8 位")
	}
	if err := s.checkUsernameAvailable(username, 0); err != nil {
		return nil, err
	}
	email := strings.TrimSpace(derefStr(req.Email))
	if email != "" && !utils.IsEmail(email) {
		return nil, NewParamError("邮箱格式不正确")
	}
	if email != "" {
		if err := s.checkEmailAvailable(email, 0); err != nil {
			return nil, err
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, NewServerError("密码加密失败")
	}

	user := model.User{
		Username: username,
		Password: string(hash),
		Email:    email,
		Phone:    strings.TrimSpace(derefStr(req.Phone)),
		Avatar:   derefStr(req.Avatar),
		Role:     "admin",
		Status:   1,
	}
	if req.Role != nil && *req.Role != "" {
		user.Role = *req.Role
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.IsMaster != nil && *req.IsMaster == 1 {
		user.IsMaster = 1
	}

	if err := s.DB().Create(&user).Error; err != nil {
		if isDuplicateKey(err) {
			return nil, NewConflict("用户名或邮箱已存在")
		}
		return nil, NewServerError("创建用户失败")
	}
	s.Log().Info("创建用户", zap.Uint("id", user.ID), zap.String("username", user.Username), zap.Int("is_master", user.IsMaster))
	return &user, nil
}

// UpdateUser 修改用户信息（仅超级管理员）。
// 保护规则：禁止禁用/降级当前登录账号（保证系统始终存在主管理员）。
func (s *UserService) UpdateUser(id uint, req *dto.UserUpdateRequest, operator *model.User) (*model.User, *BizError) {
	var user model.User
	if err := s.DB().First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewNotFound("用户不存在")
		}
		return nil, NewServerError("查询用户失败")
	}

	updates := map[string]any{}

	if req.Username != nil {
		name := strings.TrimSpace(*req.Username)
		if name == "" {
			return nil, NewParamError("用户名不能为空")
		}
		if err := s.checkUsernameAvailable(name, id); err != nil {
			return nil, err
		}
		updates["username"] = name
	}
	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if email != "" && !utils.IsEmail(email) {
			return nil, NewParamError("邮箱格式不正确")
		}
		if email != "" {
			if err := s.checkEmailAvailable(email, id); err != nil {
				return nil, err
			}
		}
		updates["email"] = email
	}
	if req.Phone != nil {
		updates["phone"] = strings.TrimSpace(*req.Phone)
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	// 自我保护：禁止操作者禁用自己
	if req.Status != nil && *req.Status != 1 && id == operator.ID {
		return nil, NewForbidden("不能禁用当前登录的账号")
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	// 自我保护：禁止操作者取消自己的超级管理员权限
	if req.IsMaster != nil && *req.IsMaster != 1 && user.IsMaster == 1 && id == operator.ID {
		return nil, NewForbidden("不能取消当前登录账号的超级管理员权限")
	}
	if req.IsMaster != nil {
		updates["is_master"] = *req.IsMaster
	}

	if len(updates) == 0 {
		return nil, NewParamError("没有需要修改的信息")
	}
	if err := s.DB().Model(&user).Updates(updates).Error; err != nil {
		if isDuplicateKey(err) {
			return nil, NewConflict("用户名或邮箱已被占用")
		}
		return nil, NewServerError("更新用户失败")
	}
	s.DB().First(&user, id)
	s.Log().Info("更新用户", zap.Uint("id", user.ID), zap.String("username", user.Username), zap.String("operator", operator.Username))
	return &user, nil
}

// ResetUserPassword 重置用户密码（仅超级管理员，无需原密码，改后旧 Token 失效）
func (s *UserService) ResetUserPassword(id uint, newPassword string) *BizError {
	if len(newPassword) < 8 {
		return NewParamError("新密码长度不能少于 8 位")
	}

	var user model.User
	if err := s.DB().First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewNotFound("用户不存在")
		}
		return NewServerError("查询用户失败")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return NewServerError("密码加密失败")
	}
	if err := s.DB().Model(&user).Updates(map[string]any{
		"password":    string(hash),
		"pwd_version": gorm.Expr("pwd_version + 1"),
	}).Error; err != nil {
		return NewServerError("重置密码失败")
	}
	s.Log().Info("重置用户密码", zap.Uint("user_id", user.ID), zap.String("username", user.Username))
	return nil
}

// DeleteUser 删除用户（软删除，仅超级管理员）。
// 保护规则：禁止删除当前登录账号（保证系统始终存在主管理员）。
func (s *UserService) DeleteUser(id uint, operator *model.User) *BizError {
	if id == operator.ID {
		return NewForbidden("不能删除当前登录的账号")
	}

	var user model.User
	if err := s.DB().First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return NewNotFound("用户不存在")
		}
		return NewServerError("查询用户失败")
	}
	if err := s.DB().Delete(&user).Error; err != nil {
		return NewServerError("删除用户失败")
	}
	s.Log().Info("删除用户", zap.Uint("id", user.ID), zap.String("username", user.Username), zap.String("operator", operator.Username))
	return nil
}

// checkUsernameAvailable 校验用户名未被其他用户占用
func (s *UserService) checkUsernameAvailable(username string, excludeID uint) *BizError {
	var count int64
	if err := s.DB().Model(&model.User{}).
		Where("username = ? AND id <> ?", username, excludeID).
		Count(&count).Error; err != nil {
		return NewServerError("查询用户失败")
	}
	if count > 0 {
		return NewConflict("用户名已被占用")
	}
	return nil
}

// checkEmailAvailable 校验邮箱未被其他用户占用
func (s *UserService) checkEmailAvailable(email string, excludeID uint) *BizError {
	var count int64
	if err := s.DB().Model(&model.User{}).
		Where("email = ? AND id <> ?", email, excludeID).
		Count(&count).Error; err != nil {
		return NewServerError("查询用户失败")
	}
	if count > 0 {
		return NewConflict("邮箱已被占用")
	}
	return nil
}
