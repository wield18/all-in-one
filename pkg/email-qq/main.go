// Package emailqq 使用qq来发email
package emailqq

import (
	"fmt"

	"github.com/wield18/all-in-one/config"
	"gopkg.in/gomail.v2"
)

type QQEmail struct {
	From     string
	Password string
	Host     string
	Port     int
}

// 本地全局指针引用
var qqEmail = &QQEmail{}

// 初始化本地引用
func InitQQEmail(qq *config.QQ) *QQEmail {
	qqEmail.From = qq.From
	qqEmail.Password = qq.Password
	qqEmail.Host = qq.Host
	qqEmail.Port = qq.Port
	return qqEmail
}

// 获得本地引用
func GetQQEmail() *QQEmail {
	return qqEmail
}

func (qq *QQEmail) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", qq.From)
	m.SetHeader("To", to)
	// 如果有多个收件人，可以这样写：
	// m.SetHeader("To", "邮箱1@example.com", "邮箱2@example.com")
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	// 2. 创建并配置SMTP拨号器
	dialer := gomail.NewDialer(qq.Host, qq.Port, qq.From, qq.Password)
	// 如需启用SSL，可以用465端口：
	// dialer := gomail.NewDialer("smtp.qq.com", 465, from, password)
	// dialer.SSL = true

	// 3. 发送邮件
	if err := dialer.DialAndSend(m); err != nil {
		return err
	}
	fmt.Println("✅ 邮件发送成功！")
	return nil
}

func (qq *QQEmail) SendCaptcha(to, captcha string) error {
	return qq.SendEmail(to, "验证码测试邮件", fmt.Sprintf("验证码: %s", captcha))
}
