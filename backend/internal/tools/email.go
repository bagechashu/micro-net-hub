package tools

import (
	"fmt"
	"math/rand"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"

	"github.com/patrickmn/go-cache"

	"github.com/go-mail/mail"
)

// 验证码放到缓存当中
var VerificationCodeCache = cache.New(24*time.Hour, 48*time.Hour)

func email(mailTo []string, subject string, body string) error {
	m := mail.NewMessage()
	m.SetHeader("From", config.Conf.Email.User)
	m.SetHeader("To", mailTo...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := mail.NewDialer(config.Conf.Email.Host, config.Conf.Email.Port, config.Conf.Email.User, config.Conf.Email.Pass)
	// Warning: can't longer than http server WriteTimeout
	d.Timeout = (config.Conf.System.WriteTimeout - 2) * time.Second

	if err := d.DialAndSend(m); err != nil {
		err := fmt.Errorf("send email error, maybe your email config is wrong: %s", err)
		global.Log.Errorf("%s", err)
		return err
	}
	return nil
}

// SendNewPass 邮件发送新密码
func SendNewPass(sendto []string, pass string) error {
	subject := "重置LDAP密码成功"

	// 邮件正文
	profileUrl := fmt.Sprintf("http://%s/#/profile/index", config.Conf.System.Domain)
	body := fmt.Sprintf(newPassBodyHtml, pass, profileUrl)
	return email(sendto, subject, body)
}

var newPassBodyHtml = `
<!DOCTYPE html>
<html>
  <head>
    <style>
      body {
        font-family: Arial, sans-serif;
      }
      .container {
        font-size: 16px;
        margin: 50px 50px;
        padding: 30px;
        border: 1px solid #ccc;
        border-radius: 8px;
        text-align: left;
      }
      .message {
        padding: 8px 32px 8px 32px;
      }
    </style>
  </head>
  <body>
    <div class="container">
      <div>尊敬的用户，您好！</div>
      <div class="message">
        密码重置为：<strong style="font-size: 22px;">%s</strong>
        <p/>
        修改请在 <a href="%s" style="color: #3498db; text-decoration: none">[个人中心]</a> 修改默认密码
      </div>
      <div>谢谢!</div>
    </div>
  </body>
</html>
`

// SendVerificationCode 邮件发送验证码
func SendVerificationCode(sendto []string) error {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	// 把验证码信息放到cache，以便于验证时拿到
	VerificationCodeCache.Set(sendto[0], vcode, time.Minute*5)
	subject := fmt.Sprintf("验证码-%s-重置LDAP密码", vcode)

	//发送的内容
	body := fmt.Sprintf(bodyHtml, vcode)
	return email(sendto, subject, body)
}

var bodyHtml = `
<!DOCTYPE html>
<html>
<head>
  <style>
    body {
      font-family: Arial, sans-serif;
    }
    .container {
      font-size: 16px;
      margin: 50px 50px;
      padding: 30px;
      border: 1px solid #ccc;
      border-radius: 8px;
      text-align: left;
    }
    .message {
      padding: 8px 32px 8px 32px;
    }
  </style>
</head>
<body>
  <div class="container">
    <div>
      尊敬的用户，您好！
    </div>
    <div class="message">
      你本次的验证码为 <strong style="font-size: 22px;">%s</strong> ，为了保证账号安全，验证码有效期为5分钟。请确认为本人操作，切勿向他人泄露，感谢您的理解与使用。
      <p/>
      此邮箱为系统邮箱，请勿回复。
    </div>
    <div>
      谢谢!
    </div>
  </div>
</body>
</html>
`
