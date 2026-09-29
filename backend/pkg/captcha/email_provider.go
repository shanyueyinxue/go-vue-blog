package captcha

import (
	"blog/pkg/email"
	"blog/pkg/utils"
	"html/template"
	"strings"
)

type EmailProvider struct {
	e            email.EmailService
	templ        *template.Template
	length       int
	emailSubject string
}

var defaultEmailTemplate = `【{{.Username}}】您正在进行{{.Subject}}，
验证码：{{.Code}}
有效期{{.Expiry}} 秒。
有效期{{.ValidMinutes}}分钟。
{{.Note}}
如非本人操作请忽略。`

var _ CaptchaProvider = (*EmailProvider)(nil)

func NewEmailProvider(e email.EmailService, config CaptchaConfig) (*EmailProvider, error) {
	var templ *template.Template
	if config.EmailTemplate != "" {
		var err error
		templ, err = template.ParseFiles(config.EmailTemplate)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		templ, err = template.New("default").Parse(defaultEmailTemplate)
		if err != nil {
			return nil, err
		}
	}
	return &EmailProvider{
		e:            e,
		templ:        templ,
		length:       config.Length,
		emailSubject: config.EmailSubject,
	}, nil
}

func (p *EmailProvider) Send(target string, val *CaptchaVal) error {
	s := &strings.Builder{}
	err := p.templ.Execute(s, val)
	if err != nil {
		return err
	}

	subject := val.Subject
	if subject == "" {
		subject = p.emailSubject
	}
	if subject == "" {
		subject = "验证码"
	}
	msg := email.Email{
		To:      []string{target},
		Subject: subject,
		Body:    s.String(),
	}
	return p.e.DialAndSend(&msg)
}

func (p *EmailProvider) ValidateFormat(target string) bool {
	return utils.IsEmail(target)
}

func (p *EmailProvider) GetType() string {
	return "email"
}

func (p *EmailProvider) GenerateCode() string {
	return utils.RandomNumber(p.length)
}
