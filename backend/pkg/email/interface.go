package email

import (
	"blog/pkg/utils"
	"fmt"
	"strings"
)

type EmailConfig struct {
	Host               string `mapstructure:"host"`
	Port               int    `mapstructure:"port"`
	SenderName         string `mapstructure:"senderName"`
	Sender             string `mapstructure:"sender"`
	Password           string `mapstructure:"password"`
	InsecureSkipVerify bool   `mapstructure:"insecureSkipVerify"`
}

type Email struct {
	To      []string
	Subject string
	Body    string
}

// EmailService 邮件服务接口
//
//	SetHeader 设置邮件头，但是不能设置 From、To、Subject、Body 属性
type EmailService interface {
	// Init 初始化邮件服务
	Init(config *EmailConfig) error
	// SetHeader 设置邮件头
	SetHeader(key, value string)
	// Dial 连接邮件服务
	Dial() error
	// Close 关闭邮件服务
	Close() error
	// IsClosed 判断是否已关闭
	IsClosed() bool
	// Send 发送邮件
	Send(email *Email) error
	// DialAndSend 连接邮件服务并发送邮件
	DialAndSend(email *Email) error
}

type EmailError struct {
	Msg string
}

func (e *EmailError) Error() string {
	return e.Msg
}

// NewEmailError 创建 EmailError
func NewEmailError(msg string) *EmailError {
	return &EmailError{Msg: msg}
}

// wrapError 包装错误，添加上下文信息
func wrapError(baseErr *EmailError, context string, cause error) error {
	if cause == nil {
		return baseErr
	}
	if context == "" {
		return fmt.Errorf("%w: %v", baseErr, cause)
	}
	return fmt.Errorf("%w: %s: %v", baseErr, context, cause)
}

// validateConfig 验证邮件配置
func validateConfig(config *EmailConfig) error {
	if config == nil {
		return NewEmailError("配置不能为空")
	}

	if config.Host == "" {
		return NewEmailError("Host不能为空")
	}

	if config.Port <= 0 || config.Port > 65535 {
		return NewEmailError("无效的端口号")
	}

	// 检查常用SMTP端口
	if config.Port != 25 && config.Port != 465 && config.Port != 587 {
		// 非标准端口，记录警告但允许使用
		fmt.Println("警告：端口号不在常用SMTP端口范围内，可能导致连接失败")
	}

	if config.Sender == "" {
		return NewEmailError("发件人地址不能为空")
	}

	// 验证邮箱格式
	if err := validateEmailFormat(config.Sender); err != nil {
		return fmt.Errorf("发件人邮箱格式无效: %w", err)
	}

	if config.Password == "" {
		return NewEmailError("密码不能为空")
	}

	return nil
}

// validateEmailFormat 验证邮箱格式
func validateEmailFormat(email string) error {
	// 简单验证：包含@和.
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") || utils.IsEmail(email) == false {
		return NewEmailError("邮箱格式不正确")
	}

	// 进一步验证：使用net/mail包的ParseAddress
	// 注意：这里不实际导入net/mail，避免不必要的依赖
	// 更严格的验证可以由调用方实现

	return nil
}

var (
	// ErrEmailServiceNotInit 邮件服务未初始化
	ErrEmailServiceNotInit = NewEmailError("邮件服务未初始化; 请先调用 Init 方法初始化")
	// ErrEmailDialFailed 连接邮件服务失败
	ErrEmailDialFailed = NewEmailError("连接邮件服务失败")
	// ErrCloseEmailServiceFailed 关闭邮件服务失败
	ErrCloseEmailServiceFailed = NewEmailError("关闭邮件服务失败")
	// ErrEmailServiceNotDial 邮件服务未连接
	ErrEmailServiceNotDial = NewEmailError("邮件服务未连接; 请先调用 Dial 方法连接")
	// ErrSendEmailFailed 发送邮件失败
	ErrSendEmailFailed = NewEmailError("发送邮件失败")

	// ErrEmailServiceClosed 邮件服务已关闭
	ErrEmailServiceClosed = NewEmailError("邮件服务已关闭")
	// ErrEmailContentEmpty 邮件内容不能为空
	ErrEmailContentEmpty = NewEmailError("邮件内容不能为空")
	// ErrEmailToEmpty 收件人不能为空
	ErrEmailToEmpty = NewEmailError("收件人不能为空")
	// ErrEmailSubjectEmpty 邮件主题不能为空
	ErrEmailSubjectEmpty = NewEmailError("邮件主题不能为空")
)
