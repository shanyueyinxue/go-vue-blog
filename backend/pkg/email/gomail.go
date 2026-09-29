package email

import (
	"crypto/tls"
	"strings"
	"sync"

	"gopkg.in/gomail.v2"
)

// goEmailServiceImpl 邮件服务实现
type goEmailServiceImpl struct {
	config *EmailConfig
	dialer *gomail.Dialer
	client gomail.SendCloser
	mutex  sync.RWMutex
	closed bool
	header map[string]string
}

// NewGomailService 创建新的邮件服务实例
func NewGomailService() EmailService {
	return &goEmailServiceImpl{
		header: make(map[string]string),
		closed: true,
	}
}

// Init 初始化邮件服务
func (es *goEmailServiceImpl) Init(config *EmailConfig) error {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	// 验证配置
	if err := validateConfig(config); err != nil {
		return err
	}

	es.config = config
	es.dialer = gomail.NewDialer(config.Host, config.Port, config.Sender, config.Password)

	// 配置TLS
	if config.InsecureSkipVerify {
		es.dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	es.closed = true
	return nil
}

// SetHeader 设置邮件头
func (es *goEmailServiceImpl) SetHeader(key, value string) {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	// 禁止设置的关键属性
	forbiddenHeaders := []string{"From", "To", "Subject", "Body"}
	normalizedKey := strings.TrimSpace(strings.ToLower(key))

	for _, forbidden := range forbiddenHeaders {
		if normalizedKey == strings.ToLower(forbidden) {
			// 记录日志或忽略，但不设置该头
			// 在实际项目中，可以考虑返回错误或记录警告日志
			return
		}
	}

	es.header[key] = value
}

// Dial 连接邮件服务
func (es *goEmailServiceImpl) Dial() error {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	if es.config == nil {
		return ErrEmailServiceNotInit
	}

	if es.dialer == nil {
		return ErrEmailServiceNotInit
	}

	client, err := es.dialer.Dial()
	if err != nil {
		return wrapError(ErrEmailDialFailed, "", err)
	}

	es.client = client
	es.closed = false
	return nil
}

// Close 关闭邮件服务
func (es *goEmailServiceImpl) Close() error {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	if es.closed || es.client == nil {
		return nil
	}

	if err := es.client.Close(); err != nil {
		return wrapError(ErrCloseEmailServiceFailed, "", err)
	}

	es.closed = true
	es.client = nil
	return nil
}

// IsClosed 判断是否已关闭
func (es *goEmailServiceImpl) IsClosed() bool {
	es.mutex.RLock()
	defer es.mutex.RUnlock()
	return es.closed
}

// Send 发送邮件
func (es *goEmailServiceImpl) Send(email *Email) error {
	if email == nil {
		return ErrEmailContentEmpty
	}

	es.mutex.RLock()
	defer es.mutex.RUnlock()

	if es.config == nil {
		return ErrEmailServiceNotInit
	}

	if es.closed || es.client == nil {
		return ErrEmailServiceClosed
	}

	// 创建邮件消息
	message, err := es.createMessage(email)
	if err != nil {
		return err
	}

	// 发送邮件
	if err := gomail.Send(es.client, message); err != nil {
		return wrapError(ErrSendEmailFailed, "通过gomail发送失败", err)
	}

	return nil
}

// DialAndSend 连接邮件服务并发送邮件
func (es *goEmailServiceImpl) DialAndSend(email *Email) error {
	// 连接邮件服务
	if err := es.Dial(); err != nil {
		return err
	}

	// 确保连接被关闭
	defer es.Close()

	// 发送邮件
	return es.Send(email)
}

// createMessage 创建邮件消息
func (es *goEmailServiceImpl) createMessage(email *Email) (*gomail.Message, error) {
	message := gomail.NewMessage()

	// 设置发件人
	if es.config.SenderName != "" {
		message.SetAddressHeader("From", es.config.Sender, es.config.SenderName)
	} else {
		message.SetHeader("From", es.config.Sender)
	}

	// 设置收件人
	if len(email.To) == 0 {
		return nil, ErrEmailToEmpty
	}

	message.SetHeader("To", email.To...)

	// 设置主题
	if email.Subject == "" {
		return nil, ErrEmailSubjectEmpty
	}
	message.SetHeader("Subject", email.Subject)

	// 设置正文
	message.SetBody("text/html", email.Body)

	// 设置自定义邮件头
	for key, value := range es.header {
		message.SetHeader(key, value)
	}

	return message, nil
}

// 验证接口实现
var _ EmailService = (*goEmailServiceImpl)(nil)
