package tools

import (
	"fmt"
	"math/rand"
	"strings"
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
func SendNewPass(sendto []string, password string) error {
	subject := fmt.Sprintf("[ %s ] LDAP Password reset successful.", config.Conf.Notice.ProjectName)

	header := config.Conf.Notice.HeaderHTML
	footer := ""

	profileUrl := fmt.Sprintf("http://%s/#/profile/index", config.Conf.Notice.ServiceDomain)

	main := fmt.Sprintf(`
    <p>
      The password has been reset to: 
      <div class="key">
        <strong style="font-size: 22px;"> %s </strong>
      </div>
    </p>
    <br>
    <p>Please modify your default password in the <a href="%s" style="color: #3498db; text-decoration: none">[Profile Index]</a> </p>
  `, password, profileUrl)

	body := fmt.Sprintf(bodyHtml, header, main, footer)
	return email(sendto, subject, body)
}

// SendVerificationCode 邮件发送验证码
func SendVerificationCode(sendto []string) error {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	// 把验证码信息放到cache，以便于验证时拿到
	VerificationCodeCache.Set(sendto[0], vcode, time.Minute*5)
	subject := fmt.Sprintf("[ %s ] %s -- Verification code for resetting LDAP password.", config.Conf.Notice.ProjectName, vcode)

	header := config.Conf.Notice.HeaderHTML
	footer := ""

	main := fmt.Sprintf(`
    <p>
      Your verification code for this session: 
      <div class="key">
        <strong style="font-size: 22px;"> %s </strong>
      </div>
    </p>
    <br>
    <p>To ensure the security of your account, the verification code is valid for 5 minutes. </p>
    <br>
    <p>Please confirm that you are the one operating and do not disclose it to others. Thank you for your understanding and use.</p>
  `, vcode)

	body := fmt.Sprintf(bodyHtml, header, main, footer)
	return email(sendto, subject, body)
}

// SendUserDeleteDenyNotice 邮件发送用户删除和禁用的通知
func SendUserStatusNotifications(sendto []string, users []string, status string) error {
	subject := fmt.Sprintf("[ %s ] Notice Of LDAP & VPN Account %s", config.Conf.Notice.ProjectName, status)

	header := config.Conf.Notice.HeaderHTML
	footer := ""

	userList := "<ul><li><span class=\"key\">" + strings.Join(users, "</span></li><li><span class=\"key\">") + "</span></li></ul>"
	main := fmt.Sprintf(`
    <p>
      The users below has been <span class="key"><strong style="font-size: 18px;"> %s </strong></span>: 
      %s 
    </p>
    <br>
    <p> Please take note.  </p>
    <br>
    <p> Deactivated or Deleted users cannot log in to the VPN or Intranet projects using their LDAP accounts. </p>
    <br>
    <p> Actived users can log in to the VPN and Intranet projects using their LDAP accounts. </p>
  `, status, userList)

	body := fmt.Sprintf(bodyHtml, header, main, footer)

	global.Log.Debugf("User Status Notice to %s", sendto)
	return email(sendto, subject, body)
}

// SendUserInfo 邮件发送用户信息
func SendUserInfo(sendto []string, username string, password string, qrRawPngBase64 string) error {
	subject := fmt.Sprintf("[ %s ] LDAP & VPN Account", config.Conf.Notice.ProjectName)

	header := config.Conf.Notice.HeaderHTML
	footer := config.Conf.Notice.FooterHTML

	main := fmt.Sprintf(`
    <h3> Intranet account and related information </h3>
    <div class="note">
      <p><b>LDAP account</b> is used to log into Intranet systems (like gitlab, nexus, etc). </p>
      <p><b>TOTP QRcode</b> is used to log into VPN systems (Ocserv).</p>
    </div>
    <h5> Portal </h5>
    <ul>
      <li>
        Intranet Portal: <span class="key"> %s </span> 
        <div class="description"> (<b>Visit after login VPN.</b> It includes an Intranet website navigator and a personal account information manager.) </div>
      </li>
      <li> VPN Server address: <span class="key"> %s </span> </li>
    </ul>
    <h5> LDAP Account </h5>
    <ul>
      <li> Username: <span class="key"> %s </span> </li>
      <li> Password: <span class="key"> %s </span> </li>
    </ul>
    <h5> TOTP QRcode </h5>
      <img src="data:image/png;base64,%s" alt="QR Code">
    <h5> VPN Password Notes </h5>
      <div style="padding-left: 30px;"> 
        <p>The <b>VPN password</b> is a combination of the <b>LDAP password</b> and the <b>TOTP dynamic code</b>. </p>
        <p>eg: <br></p>
        <div class="note">
          <p>
            if <br>
            <b>LDAP Password</b> is "ldappasswd" <br>
            <b>TOTP dynamic code</b> is "123456" 
          </p>
          <p>
            then <br>
            <b>VPN Password</b> is "ldappasswd123456" 
          </p>
        </div>
      </div>
  `, config.Conf.Notice.ServiceDomain, config.Conf.Notice.VPNServer, username, password, qrRawPngBase64)

	body := fmt.Sprintf(bodyHtml, header, main, footer)

	// global.Log.Debugf("%s\n%s", subject, body)
	// return nil
	return email(sendto, subject, body)
}

var bodyHtml = `
<!DOCTYPE html>
<html>
<head>
  <style>
    body {
      font-family: "Lucida Console", Courier, monospace;
      margin: 0;
      padding: 0;
    }
    hr {
      border: 0;
      height: 1px;
      background-image: linear-gradient(to right, rgba(0, 0, 0, 0), #333, rgba(0, 0, 0, 0));
      margin: 20px 0;
    }
    p {
      margin: 2px 0;
    }
    .container {
      max-width: 800px;
      margin: 20px auto;
      padding: 2px 20px;
      border: 1px solid #ccc;
      border-radius: 8px;
      background-color: #f9f9f9;
      text-align: left;
    }
    @media only screen and (max-width: 900px) {
      .container {
        margin: auto !important;
      }
    }
    .message {
      margin-bottom: 20px;
    }
    .note {
      padding: 5px;
      background-color: #f0f8ff;
      border-left: 5px solid #1897e1;
    }
    .key {
      display: inline-flex;
      padding: 3px 8px;
      background-color: #d7f9d6;
      border-radius: 3px;
      margin-right: 5px;
    }
    .description {
      color: #333;
      font-style: italic;
      font-size: xx-small;
    }
  </style>
</head>
<body>
  <div class="container">
    <h3> Dear user, </h3>
    <div class="message">
      <!-- header -->
      %s
    </div>
    
    <div class="message">
      <hr>
      <!-- main -->
      %s
    </div>
    <div class="message">
      <hr>
      <!-- footer -->
      %s
    </div>
    <div class="message">
      This email is a system mailbox. Please do not reply.
    </div>
    <h3> Thanks </h3>
  </div>
</body>
</html>
`
