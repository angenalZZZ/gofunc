package f

import "gopkg.in/gomail.v2"

// MailSMTP 邮箱服务器信息
type MailSMTP struct {
	Port int    `json:"port" default:"465"`              // 163邮箱端口号
	Host string `json:"host" default:"smtp.163.com"`     // 163邮箱服务网址
	User string `json:"user" default:"angenals@163.com"` // 发送邮件的账号
	Pass string `json:"pass" default:""`                 // 密码或授权码
}

// MailMessage 收件人的电子邮件
type MailMessage struct {
	Recipient []string // 收件人
	Subject   string   // 邮件主题
	Body      string   // 邮件内容
}

// MailOptions 电子邮件信息
type MailOptions struct {
	MailSMTP    // 邮箱服务器信息
	MailMessage // 收件人的电子邮件
}

// MailSend 发送电子邮件
func MailSend(o *MailOptions) error {
	m := gomail.NewMessage()

	//设置发件人
	m.SetHeader("From", o.MailSMTP.User)

	//设置发送给多个用户
	m.SetHeader("To", o.MailMessage.Recipient...)

	//设置邮件主题
	m.SetHeader("Subject", o.MailMessage.Subject)

	//设置邮件正文
	m.SetBody("text/html", o.MailMessage.Body)

	d := gomail.NewDialer(o.MailSMTP.Host, o.MailSMTP.Port, o.MailSMTP.User, o.MailSMTP.Pass)
	return d.DialAndSend(m)
}
