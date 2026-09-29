package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"net/textproto"
	"strings"
	"sync"
)

// emailServiceImpl 邮件服务实现
type emailServiceImpl struct {
	config *EmailConfig
	client *smtp.Client
	header textproto.MIMEHeader
	mu     sync.RWMutex
	closed bool
}

// NewEmailService 创建新的邮件服务实例
func NewEmailService() EmailService {
	return &emailServiceImpl{
		header: make(textproto.MIMEHeader),
	}
}

// Init 初始化邮件服务
func (es *emailServiceImpl) Init(config *EmailConfig) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	// 验证配置
	if err := validateConfig(config); err != nil {
		return err
	}

	es.config = config
	es.closed = true
	es.client = nil
	es.header = make(textproto.MIMEHeader)

	return nil
}

// SetHeader 设置邮件头
func (es *emailServiceImpl) SetHeader(key, value string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	if es.header == nil {
		es.header = make(textproto.MIMEHeader)
	}

	// 不允许设置受保护的头部字段
	protectedHeaders := map[string]bool{
		"From":    true,
		"To":      true,
		"Subject": true,
		"Body":    true,
	}

	if !protectedHeaders[textproto.CanonicalMIMEHeaderKey(key)] {
		es.header.Set(key, value)
	}
}

// Dial 连接邮件服务
func (es *emailServiceImpl) Dial() error {
	es.mu.Lock()
	defer es.mu.Unlock()

	if es.config == nil {
		return ErrEmailServiceNotInit
	}

	// 建立SMTP连接
	addr := fmt.Sprintf("%s:%d", es.config.Host, es.config.Port)
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName:         es.config.Host,
		InsecureSkipVerify: es.config.InsecureSkipVerify,
	})
	if err != nil {
		return wrapError(ErrEmailDialFailed, "建立TLS连接失败", err)
	}

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, es.config.Host)
	if err != nil {
		conn.Close()
		return wrapError(ErrEmailDialFailed, "创建SMTP客户端失败", err)
	}

	// 身份认证
	auth := smtp.PlainAuth("", es.config.Sender, es.config.Password, es.config.Host)
	if err := client.Auth(auth); err != nil {
		client.Close()
		return wrapError(ErrEmailDialFailed, "SMTP身份认证失败", err)
	}

	es.client = client
	es.closed = false
	return nil
}

// Close 关闭邮件服务
func (es *emailServiceImpl) Close() error {
	es.mu.Lock()
	defer es.mu.Unlock()

	if es.closed {
		return nil
	}

	if es.client == nil {
		es.closed = true
		return nil
	}

	err := es.client.Quit()
	if err != nil {
		err = wrapError(ErrCloseEmailServiceFailed, "", err)
	}

	es.client = nil
	es.closed = true
	return err
}

// IsClosed 判断是否已关闭
func (es *emailServiceImpl) IsClosed() bool {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.closed
}

// Send 发送邮件
func (es *emailServiceImpl) Send(email *Email) error {
	es.mu.RLock()
	defer es.mu.RUnlock()

	if es.closed {
		return ErrEmailServiceClosed
	}

	if es.config == nil {
		return ErrEmailServiceNotInit
	}

	if es.client == nil {
		return ErrEmailServiceNotDial
	}

	if err := es.validateEmail(email); err != nil {
		return err
	}

	// 设置发件人
	if err := es.client.Mail(es.config.Sender); err != nil {
		return wrapError(ErrSendEmailFailed, "设置发件人失败", err)
	}

	// 设置收件人
	for _, to := range email.To {
		if err := es.client.Rcpt(to); err != nil {
			return wrapError(ErrSendEmailFailed, "设置收件人失败", err)
		}
	}

	// 发送邮件数据
	wc, err := es.client.Data()
	if err != nil {
		return wrapError(ErrSendEmailFailed, "准备邮件数据失败", err)
	}
	defer wc.Close()

	// 构建邮件内容
	message := es.buildMessage(email)
	if _, err := wc.Write([]byte(message)); err != nil {
		return wrapError(ErrSendEmailFailed, "写入邮件数据失败", err)
	}

	return nil
}

// DialAndSend 连接邮件服务并发送邮件
func (es *emailServiceImpl) DialAndSend(email *Email) error {
	// 连接邮件服务
	if err := es.Dial(); err != nil {
		return err
	}

	// 确保连接被关闭
	defer es.Close()

	// 发送邮件
	return es.Send(email)
}

// validateEmail 验证邮件内容
func (es *emailServiceImpl) validateEmail(email *Email) error {
	if email == nil {
		return ErrEmailContentEmpty
	}

	if len(email.To) == 0 {
		return ErrEmailToEmpty
	}

	for _, to := range email.To {
		if strings.TrimSpace(to) == "" {
			return ErrEmailToEmpty
		}
	}

	if strings.TrimSpace(email.Subject) == "" {
		return ErrEmailSubjectEmpty
	}

	if strings.TrimSpace(email.Body) == "" {
		return ErrEmailContentEmpty
	}

	return nil
}

// buildMessage 构建邮件消息
func (es *emailServiceImpl) buildMessage(email *Email) string {
	var sb strings.Builder

	// 基本头部
	sb.WriteString(fmt.Sprintf("From: %s <%s>\r\n", es.config.SenderName, es.config.Sender))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))

	// 自定义头部
	for key, values := range es.header {
		for _, value := range values {
			sb.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
		}
	}

	// 邮件内容类型
	sb.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	sb.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	sb.WriteString("\r\n")

	// 邮件正文
	sb.WriteString(email.Body)
	sb.WriteString("\r\n")

	return sb.String()
}
