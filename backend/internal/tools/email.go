package tools

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"

	"github.com/patrickmn/go-cache"

	"github.com/go-mail/mail"
)

// 验证码放到缓存当中
var VerificationCodeCache = cache.New(24*time.Hour, 48*time.Hour)
var lock sync.Mutex
var wg sync.WaitGroup

func email(mailTo []string, subject string, body string) error {
	if !config.Conf.Email.Enable {
		return nil
	}

	m := mail.NewMessage()
	m.SetHeader("From", config.Conf.Email.User)
	m.SetHeader("To", mailTo...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	lock.Lock()
	defer lock.Unlock()
	wg.Add(1)
	go func() {
		d := mail.NewDialer(config.Conf.Email.Host, config.Conf.Email.Port, config.Conf.Email.User, config.Conf.Email.Pass)
		// Warning: can't longer than http server WriteTimeout
		d.Timeout = (config.Conf.System.WriteTimeout - 2) * time.Second

		if err := d.DialAndSend(m); err != nil {
			err := fmt.Errorf("send email error, maybe your email config is wrong: %s", err)
			global.Log.Errorf("%s", err)
		}
		wg.Done()
	}()

	return nil
}

// SendNewPass 邮件发送新密码
func SendNewPass(sendto []string, password, username string) error {
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

	body := fmt.Sprintf(bodyHtml, username, header, main, footer)
	return email(sendto, subject, body)
}

// SendVerificationCode 邮件发送验证码
func SendVerificationCode(sendto []string, username string) error {
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

	body := fmt.Sprintf(bodyHtml, username, header, main, footer)
	return email(sendto, subject, body)
}

// SendUserDeleteDenyNotice 邮件发送用户删除和禁用的通知
func SendUserStatusNotifications(sendto []string, users []string, status string) error {
	subject := fmt.Sprintf("[ %s ] Notice Of LDAP & VPN Account %s", config.Conf.Notice.ProjectName, status)
	salutation := "Administrator"
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

	body := fmt.Sprintf(bodyHtml, salutation, header, main, footer)

	global.Log.Debugf("User Status Notice to %s", sendto)
	return email(sendto, subject, body)
}

// SendUserInfo 邮件发送用户信息
func SendUserInfo(sendto []string, username string, password string, qrRawPngBase64 string) error {
	var subject, main string
	if config.Conf.Notice.VpnInfoSendSwitch {
		subject = fmt.Sprintf("[ %s ] LDAP & VPN Account", config.Conf.Notice.ProjectName)
		main = fmt.Sprintf(`
      <h3> Intranet account and related information </h3>
      <h5> Portal </h5>
      <ul>
        <li> VPN Server address: <span class="key"> %s </span> </li>
        <li>
          Intranet Portal: <span class="key"> %s </span> 
          <div class="description"> (<b>Accessible after logging in to the VPN.</b> It includes an intranet website navigator and a personal account information manager.)</div>
        </li>
      </ul>
      <h5> LDAP Account </h5>
      <ul>
        <li> Username: <span class="key"> %s </span> </li>
        <li> Password: <span class="key"> %s </span> </li>
      </ul>
      <h5> TOTP QRcode </h5>
      <ul>
        <img src="data:image/png;base64,%s" alt="QR Code">
      </ul>
      <h5> How to use </h5>
      <ol>
        <li> Scan the QR code by <b>"MFA tools"(eg. Google authentication)</b>. Save it and you can get the <b>PIN Code</b>.</li>
        <li> Connect VPN Server with <b>"Ocserv VPN Client"(eg. Cisco Secure Client)</b> using LDAP username and Combined Password<b>[LDAP Password + OTP Code]</b>.</li>
        <li> Login <b>"Intranet systems"(eg. gitlab, nexus, Intranet Portal etc)</b> using LDAP username and LDAP password.</li>
        <li> Login <b>"Intranet Portal"</b> to change your default LDAP password.</li>
        <li> Click <span class="key">"Forget Password"</span> at <b>"Intranet Portal Login Dialog"</b> to reset your password, if forget.</li>
      </ol>
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
    `, config.Conf.Notice.VPNServer, config.Conf.Notice.ServiceDomain, username, password, qrRawPngBase64)

	} else {
		subject = fmt.Sprintf("[ %s ] LDAP Account", config.Conf.Notice.ProjectName)
		main = fmt.Sprintf(`
      <h3> LDAP Account and related information </h3>
      <h5> Portal </h5>
      <ul>
        <li>
          Intranet Portal: <span class="key"> %s </span> 
          <div class="description"> (<b>Accessible after logging in to the VPN.</b> It includes an intranet website navigator and a personal account information manager.)</div>
        </li>
      </ul>
      <h5> LDAP Account </h5>
      <ul>
        <li> Username: <span class="key"> %s </span> </li>
        <li> Password: <span class="key"> %s </span> </li>
      </ul>
      <h5> How to use </h5>
      <ol>
        <li> Login <b>"Intranet systems"(eg. gitlab, nexus, Intranet Portal etc)</b> using LDAP username and LDAP password.</li>
        <li> Login <span class="key"><b>"Intranet Portal"</b></span> to change your default LDAP password.</li>
        <li> Click <span class="key">"Forget Password"</span> at <b>"Intranet Portal Login Dialog"</b> to reset your password, if forget.</li>
      </ol>
    `, config.Conf.Notice.ServiceDomain, username, password)
	}

	header := config.Conf.Notice.HeaderHTML
	footer := config.Conf.Notice.FooterHTML

	body := fmt.Sprintf(bodyHtml, username, header, main, footer)

	// global.Log.Debugf("%s\n%s", subject, body)
	// return nil

	// 在没有可用 smtp 邮箱时使用,方便下发账户信息.
	if config.Conf.Notice.AccountCreatedNoticeSave {
		// 保存邮件内容到 config.Conf.Notice.AccountCreatedNoticeDir  目录下
		// 文件名为 username + 时间戳 + .html
		fileName := fmt.Sprintf("%s/%s_%d.html", config.Conf.Notice.AccountCreatedNoticeDir, username, time.Now().Unix())
		err := os.WriteFile(fileName, []byte(body), 0644)
		if err != nil {
			return err
		}
	}

	return email(sendto, subject, body)
}

var bodyHtml = `
<!DOCTYPE html>
<html>
<head>
  <style>
    body {
      margin: 0;
      padding: 0;
      font-size: small;
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
      padding: 2px 0 2px 10px;
      margin-top: 12px;
      background-color: #f0f8ff;
      border-left: 5px solid #85bedf;
    }
    .key {
      display: inline-flex;
      padding: 3px 8px;
      background-color: #f3fff3;
      border-radius: 3px;
      margin-right: 5px;
    }
    .description {
      color: #333;
      font-style: italic;
      font-size: x-small;
    }
  </style>
</head>
<body>
  <div class="container">
    <!-- salutation -->
    <h3> Dear %s, </h3>
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
      This is a system-generated email. Please do not reply.
    </div>
    <h3> Thanks </h3>
  </div>
</body>
</html>
`
