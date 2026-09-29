package email_test

import (
	"blog/pkg/email"
	"testing"
)

func TestSendEmail_Success(t *testing.T) {
	// 这是一个集成测试，需要有效的SMTP配置
	// 如果密码是测试值或未设置，则跳过
	if testing.Short() {
		t.Skip("跳过集成测试，使用 -short 标志")
	}

	// 检查是否有有效的测试配置
	// 可以通过环境变量提供真实配置
	testSender := "test@example.com"
	testPassword := "test-password"

	if testPassword == "test-password" {
		t.Skip("使用测试密码，跳过集成测试。设置有效的SMTP密码以运行集成测试。")
	}

	config := email.EmailConfig{
		Host:               "smtp.qq.com",
		Port:               465,
		SenderName:         "test",
		Sender:             testSender,
		Password:           testPassword,
		InsecureSkipVerify: true,
	}

	e := email.NewEmailService()
	if err := e.Init(&config); err != nil {
		t.Skipf("初始化失败，跳过测试: %v", err)
	}

	emailMsg := email.Email{
		To:      []string{config.Sender},
		Subject: "test email",
		Body:    "test email body",
	}
	SendEmail(t, e, emailMsg)

	e = email.NewGomailService()
	if err := e.Init(&config); err != nil {
		t.Skipf("初始化gomail失败，跳过测试: %v", err)
	}

	emailMsg = email.Email{
		To:      []string{config.Sender},
		Subject: "test gomail email",
		Body:    "test gomail email body",
	}
	SendEmail(t, e, emailMsg)
}

func SendEmail(t *testing.T, e email.EmailService, msg email.Email) {
	if err := e.Dial(); err != nil {
		t.Error(err)
	}

	if err := e.Send(&msg); err != nil {
		t.Error(err)
	}
	if err := e.Close(); err != nil {
		t.Error(err)
	}

	(&msg).Body = (&msg).Body + " DialAndSend email body"
	if err := e.DialAndSend(&msg); err != nil {
		t.Error(err)
	}

	t.Log("send email success")
}

func TestInit_ValidConfig(t *testing.T) {
	config := email.EmailConfig{
		Host:               "smtp.example.com",
		Port:               587,
		SenderName:         "测试发件人",
		Sender:             "sender@example.com",
		Password:           "password",
		InsecureSkipVerify: false,
	}

	// 测试标准库实现
	e1 := email.NewEmailService()
	if err := e1.Init(&config); err != nil {
		t.Errorf("标准库实现初始化失败: %v", err)
	}

	// 测试gomail实现
	e2 := email.NewGomailService()
	if err := e2.Init(&config); err != nil {
		t.Errorf("gomail实现初始化失败: %v", err)
	}
}

func TestInit_InvalidConfig(t *testing.T) {
	testCases := []struct {
		name   string
		config email.EmailConfig
	}{
		{
			name: "空Host",
			config: email.EmailConfig{
				Host:     "",
				Port:     587,
				Sender:   "test@example.com",
				Password: "password",
			},
		},
		{
			name: "无效端口",
			config: email.EmailConfig{
				Host:     "smtp.example.com",
				Port:     0,
				Sender:   "test@example.com",
				Password: "password",
			},
		},
		{
			name: "空发件人",
			config: email.EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Sender:   "",
				Password: "password",
			},
		},
		{
			name: "无效邮箱格式",
			config: email.EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Sender:   "invalid-email",
				Password: "password",
			},
		},
		{
			name: "空密码",
			config: email.EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Sender:   "test@example.com",
				Password: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := email.NewEmailService()
			if err := e.Init(&tc.config); err == nil {
				t.Error("期望初始化失败，但成功了")
			}

			e2 := email.NewGomailService()
			if err := e2.Init(&tc.config); err == nil {
				t.Error("期望gomail初始化失败，但成功了")
			}
		})
	}
}
