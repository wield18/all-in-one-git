package emailqq

import (
	"fmt"
	"sync"

	"gopkg.in/gomail.v2"
)

// QQEmail
type QQEmail struct {
	From string `yaml:"From"`
	Code string `yaml:"Code"`
	Host string `yaml:"Host"`
	Port int    `yaml:"Port"`
}

// QQDialer 真正的拨号器
type QQDialer struct {
	dialer *gomail.Dialer
	from   string
	mu     sync.Mutex // 如果并发发送可能需要
}

// NewDialer 创建并配置SMTP拨号器
func NewDialer(conf QQEmail) *QQDialer {
	if conf.Host == "" {
		panic("conf.Host is empty string")
	}
	if conf.Port == 0 {
		panic("conf.Port is 0(int)")
	}
	if conf.From == "" {
		panic("conf.From is empty string")
	}
	if conf.Code == "" {
		panic("conf.Code is empty string")
	}
	dialer := gomail.NewDialer(conf.Host, conf.Port, conf.From, conf.Code)
	return &QQDialer{
		dialer: dialer,
		from:   conf.From,
	}
}

// 常用的验证
func (qq *QQDialer) SendCaptcha(to, captcha string) error {
	return qq.SendEmail(to, "验证码测试邮件", fmt.Sprintf("验证码: %s", captcha))
}

func (qq *QQDialer) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", qq.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	qq.mu.Lock()
	defer qq.mu.Unlock()
	return qq.dialer.DialAndSend(m)
}
