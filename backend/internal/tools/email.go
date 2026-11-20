package tools

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"

	"github.com/go-mail/mail"
	"github.com/patrickmn/go-cache"
)

// 验证码放到缓存当中
var VerificationCodeCache = cache.New(24*time.Hour, 48*time.Hour)
var lock sync.Mutex
var wg sync.WaitGroup

// inlineImage 以内嵌附件（Content-ID）方式随邮件发送的图片，正文用 cid:<Name> 引用。
// 很多邮件客户端（如 Gmail）不显示 data: URI 图片，内嵌附件兼容性最好。
type inlineImage struct {
	Name string
	Data []byte
}

func email(mailTo []string, subject string, body string, images ...inlineImage) error {
	if !config.Conf.Email.Enable {
		return nil
	}

	m := mail.NewMessage()
	m.SetHeader("From", config.Conf.Email.User)
	m.SetHeader("To", mailTo...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	for _, img := range images {
		m.EmbedReader(img.Name, bytes.NewReader(img.Data))
	}

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

// ---------------------------------------------------------------------------
// 模板数据
// ---------------------------------------------------------------------------

// mailPage 邮件外框：品牌栏 + 标题 + 称呼 + 正文 + 页脚
type mailPage struct {
	Project    string
	Initial    string
	Preheader  string // 收件箱列表里显示的摘要（正文中隐藏）
	Title      string
	Salutation string
	Header     template.HTML // config.notice.header-html（管理员配置，按受信任 HTML 处理）
	Content    template.HTML
	Footer     template.HTML // config.notice.footer-html
}

type kv struct {
	Label string
	Value string
	Mono  bool
	Link  string
}

type codeMail struct {
	Lead    string
	Code    string
	Notes   []string
	BtnText string
	BtnURL  string
}

type statusMail struct {
	Status   string
	Tone     string // success | danger
	Users    []string
	Explain  string
	Reminder string
}

type userInfoMail struct {
	WithVPN   bool
	Account   []kv
	QRSrc     template.URL
	PortalURL string
}

// ---------------------------------------------------------------------------
// 对外的发送函数（签名保持不变）
// ---------------------------------------------------------------------------

// SendNewPass 邮件发送新密码
func SendNewPass(sendto []string, password, username string) error {
	subject := fmt.Sprintf("[ %s ] LDAP Password reset successful.", projectName())

	content, err := render("code", codeMail{
		Lead:    "Your LDAP password has been reset. Your new password is:",
		Code:    password,
		Notes:   []string{"Please sign in and change this password as soon as possible."},
		BtnText: "Change password",
		BtnURL:  portalURL("#/profile/index"),
	})
	if err != nil {
		return err
	}
	body, err := page(mailPage{
		Preheader:  "Your LDAP password has been reset.",
		Title:      "Password reset",
		Salutation: username,
		Content:    content,
	})
	if err != nil {
		return err
	}
	return email(sendto, subject, body)
}

// SendVerificationCode 邮件发送验证码
func SendVerificationCode(sendto []string, username string) error {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	// 把验证码信息放到cache，以便于验证时拿到
	VerificationCodeCache.Set(sendto[0], vcode, time.Minute*5)
	subject := fmt.Sprintf("[ %s ] %s -- Verification code for resetting LDAP password.", projectName(), vcode)

	content, err := render("code", codeMail{
		Lead: "Use this verification code to reset your LDAP password:",
		Code: vcode,
		Notes: []string{
			"The code is valid for 5 minutes.",
			"If you did not request this, you can ignore this email. Never share the code with anyone.",
		},
	})
	if err != nil {
		return err
	}
	body, err := page(mailPage{
		Preheader:  fmt.Sprintf("Your verification code is %s (valid for 5 minutes).", vcode),
		Title:      "Verification code",
		Salutation: username,
		Content:    content,
	})
	if err != nil {
		return err
	}
	return email(sendto, subject, body)
}

// SendUserStatusNotifications 邮件发送用户删除、禁用、启用的通知
func SendUserStatusNotifications(sendto []string, users []string, status string) error {
	subject := fmt.Sprintf("[ %s ] Notice Of LDAP & VPN Account %s", projectName(), status)

	data := statusMail{Status: status, Users: users, Tone: "danger",
		Explain:  "These users can no longer sign in to the VPN or intranet systems with their LDAP accounts.",
		Reminder: "Please take note."}
	if status == "actived" {
		data.Tone = "success"
		data.Explain = "These users can now sign in to the VPN and intranet systems with their LDAP accounts."
	}
	content, err := render("status", data)
	if err != nil {
		return err
	}
	body, err := page(mailPage{
		Preheader:  fmt.Sprintf("%d account(s) %s.", len(users), status),
		Title:      fmt.Sprintf("Accounts %s", status),
		Salutation: "Administrator",
		Content:    content,
	})
	if err != nil {
		return err
	}

	global.Log.Debugf("User Status Notice to %s", sendto)
	return email(sendto, subject, body)
}

// SendUserInfo 邮件发送用户信息
func SendUserInfo(sendto []string, username string, password string, qrRawPngBase64 string) error {
	withVPN := config.Conf.Notice.VpnInfoSendSwitch
	subject := fmt.Sprintf("[ %s ] LDAP Account", projectName())
	if withVPN {
		subject = fmt.Sprintf("[ %s ] LDAP & VPN Account", projectName())
	}

	account := []kv{}
	if withVPN {
		account = append(account, kv{Label: "VPN server", Value: config.Conf.Notice.VPNServer, Mono: true})
	}
	account = append(account,
		kv{Label: "Intranet portal", Value: config.Conf.Notice.ServiceDomain, Link: portalURL("")},
		kv{Label: "Username", Value: username, Mono: true},
		kv{Label: "Password", Value: password, Mono: true},
	)

	// 发送的邮件用内嵌附件引用二维码；保存到本地的 HTML 用 data: URI，方便直接用浏览器打开
	build := func(qrSrc string) (string, error) {
		content, err := render("userInfo", userInfoMail{
			WithVPN:   withVPN,
			Account:   account,
			QRSrc:     template.URL(qrSrc), // 只会是 cid: 或本函数生成的 data:image/png
			PortalURL: portalURL(""),
		})
		if err != nil {
			return "", err
		}
		return page(mailPage{
			Preheader:  "Your intranet account is ready.",
			Title:      "Your account is ready",
			Salutation: username,
			Header:     template.HTML(config.Conf.Notice.HeaderHTML),
			Content:    content,
			Footer:     template.HTML(config.Conf.Notice.FooterHTML),
		})
	}

	var images []inlineImage
	mailQR, fileQR := "", ""
	if withVPN && qrRawPngBase64 != "" {
		png, err := base64.StdEncoding.DecodeString(qrRawPngBase64)
		if err != nil {
			return fmt.Errorf("decode totp qrcode: %w", err)
		}
		images = append(images, inlineImage{Name: "totp-qrcode.png", Data: png})
		mailQR = "cid:totp-qrcode.png"
		fileQR = "data:image/png;base64," + qrRawPngBase64
	}

	body, err := build(mailQR)
	if err != nil {
		return err
	}

	// 在没有可用 smtp 邮箱时使用,方便下发账户信息.
	if config.Conf.Notice.AccountCreatedNoticeSave {
		// 保存邮件内容到 config.Conf.Notice.AccountCreatedNoticeDir  目录下
		// 文件名为 username + 时间戳 + .html
		fileBody, err := build(fileQR)
		if err != nil {
			return err
		}
		// 用户名只保留安全字符，避免出现 "/" 等导致写到目录之外
		safeName := strings.Map(func(r rune) rune {
			if r == '-' || r == '_' || r == '.' || (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				return r
			}
			return '_'
		}, username)
		fileName := filepath.Join(config.Conf.Notice.AccountCreatedNoticeDir, fmt.Sprintf("%s_%d.html", safeName, time.Now().Unix()))
		if err := os.WriteFile(fileName, []byte(fileBody), 0644); err != nil {
			return err
		}
	}

	return email(sendto, subject, body, images...)
}

// ---------------------------------------------------------------------------
// 渲染
// ---------------------------------------------------------------------------

func projectName() string {
	if p := strings.TrimSpace(config.Conf.Notice.ProjectName); p != "" {
		return p
	}
	return "Micro Net Hub"
}

// portalURL 门户地址（service-domain 未配置时返回空，模板里不显示按钮/链接）
func portalURL(path string) string {
	d := strings.TrimSpace(config.Conf.Notice.ServiceDomain)
	if d == "" {
		return ""
	}
	if !strings.HasPrefix(d, "http://") && !strings.HasPrefix(d, "https://") {
		d = "http://" + d
	}
	return strings.TrimRight(d, "/") + "/" + path
}

func render(name string, data any) (template.HTML, error) {
	var buf bytes.Buffer
	if err := mailTemplates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("render email %s: %w", name, err)
	}
	return template.HTML(buf.String()), nil
}

func page(p mailPage) (string, error) {
	p.Project = projectName()
	p.Initial = strings.ToUpper(string([]rune(p.Project)[:1]))
	html, err := render("layout", p)
	return string(html), err
}

// 邮件样式：表格布局 + 内联样式（很多客户端会去掉 <style>），配色与管理后台一致（Tabler）。
// <style> 里只放增强项，以及 header-html / footer-html 里会用到的 .note、.key 等 class。
var mailTemplates = template.Must(template.New("mail").Funcs(template.FuncMap{
	"btn": func(text, url string) map[string]string { return map[string]string{"Text": text, "URL": url} },
}).Parse(`
{{define "layout"}}<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="color-scheme" content="light">
<title>{{.Title}}</title>
<style>
  body { margin: 0; padding: 0; background: #f6f8fb; }
  a { color: #066fd1; }
  .content h3 { margin: 20px 0 8px; font-size: 15px; color: #182433; }
  .content h5 { margin: 16px 0 6px; font-size: 13px; color: #182433; }
  .content p { margin: 0 0 8px; }
  .content ul, .content ol { margin: 0 0 12px; padding-left: 20px; }
  .content li { margin: 4px 0; }
  .note { margin: 12px 0; padding: 10px 14px; background: #e9f2fc; border-left: 3px solid #066fd1; border-radius: 4px; }
  .key { display: inline-block; padding: 1px 6px; font-family: SFMono-Regular, Menlo, Consolas, monospace; font-size: 12px; background: #f1f5f9; border: 1px solid #e6e9ee; border-radius: 4px; }
  .description { color: #667382; font-size: 12px; }
  @media only screen and (max-width: 620px) {
    .card { border-radius: 0 !important; }
    .pad { padding-left: 20px !important; padding-right: 20px !important; }
  }
</style>
</head>
<body style="margin:0;padding:0;background:#f6f8fb;">
<div style="display:none;max-height:0;overflow:hidden;opacity:0;">{{.Preheader}}</div>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="background:#f6f8fb;">
<tr><td align="center" style="padding:32px 12px;">
  <table role="presentation" width="600" cellpadding="0" cellspacing="0" border="0" style="width:100%;max-width:600px;">
    <!-- 品牌栏 -->
    <tr><td style="padding:0 4px 16px;">
      <table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr>
        <td width="32" height="32" align="center" valign="middle" style="width:32px;height:32px;background:#066fd1;border-radius:8px;color:#ffffff;font:700 16px/32px -apple-system,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">{{.Initial}}</td>
        <td style="padding-left:10px;font:600 16px/1.4 -apple-system,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;color:#182433;">{{.Project}}</td>
      </tr></table>
    </td></tr>
    <!-- 卡片 -->
    <tr><td class="card" style="background:#ffffff;border:1px solid #dce1e7;border-radius:8px;overflow:hidden;">
      <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
        <tr><td style="height:4px;background:#066fd1;font-size:0;line-height:0;">&nbsp;</td></tr>
        <tr><td class="pad content" style="padding:32px 40px;font:14px/1.6 -apple-system,'Segoe UI',Roboto,'PingFang SC','Microsoft YaHei',Helvetica,Arial,sans-serif;color:#182433;">
          <h1 style="margin:0 0 16px;font-size:20px;line-height:1.4;font-weight:600;color:#182433;">{{.Title}}</h1>
          <p style="margin:0 0 16px;">Hi <strong>{{.Salutation}}</strong>,</p>
          {{if .Header}}<div style="margin:0 0 20px;color:#3a4859;">{{.Header}}</div>{{end}}
          {{.Content}}
          {{if .Footer}}<div style="margin:28px 0 0;padding-top:20px;border-top:1px solid #e6e9ee;color:#3a4859;">{{.Footer}}</div>{{end}}
        </td></tr>
      </table>
    </td></tr>
    <!-- 页脚 -->
    <tr><td align="center" style="padding:20px 12px;font:12px/1.6 -apple-system,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;color:#8a97a8;">
      This is a system-generated email from {{.Project}}. Please do not reply.
    </td></tr>
  </table>
</td></tr>
</table>
</body>
</html>{{end}}

{{define "button"}}<table role="presentation" cellpadding="0" cellspacing="0" border="0" style="margin:20px 0 4px;"><tr>
  <td style="background:#066fd1;border-radius:6px;"><a href="{{.URL}}" target="_blank" style="display:inline-block;padding:10px 20px;font-weight:600;font-size:14px;color:#ffffff;text-decoration:none;border-radius:6px;">{{.Text}}</a></td>
</tr></table>{{end}}

{{define "code"}}<p style="margin:0 0 12px;">{{.Lead}}</p>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="margin:0 0 16px;">
  <tr><td align="center" style="padding:18px 12px;background:#f6f8fb;border:1px solid #e6e9ee;border-radius:8px;font:600 28px/1.2 SFMono-Regular,Menlo,Consolas,'Liberation Mono',monospace;letter-spacing:6px;color:#182433;">{{.Code}}</td></tr>
</table>
{{range .Notes}}<p style="margin:0 0 6px;font-size:13px;color:#667382;">{{.}}</p>{{end}}
{{if and .BtnURL .BtnText}}{{template "button" (btn .BtnText .BtnURL)}}{{end}}{{end}}

{{define "status"}}<p style="margin:0 0 12px;">The following account(s) have been
  <span style="display:inline-block;padding:1px 8px;border-radius:4px;font-weight:600;{{if eq .Tone "success"}}color:#2fb344;background:#eaf7ec;{{else}}color:#d63939;background:#fbebeb;{{end}}">{{.Status}}</span>:</p>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="margin:0 0 16px;border:1px solid #e6e9ee;border-radius:8px;">
  {{range $i, $u := .Users}}<tr><td style="padding:10px 14px;{{if $i}}border-top:1px solid #e6e9ee;{{end}}font:13px/1.4 SFMono-Regular,Menlo,Consolas,monospace;color:#182433;">{{$u}}</td></tr>{{end}}
</table>
<p style="margin:0 0 6px;font-size:13px;color:#3a4859;">{{.Explain}}</p>
<p style="margin:0;font-size:13px;color:#667382;">{{.Reminder}}</p>{{end}}

{{define "kv"}}<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="margin:0 0 20px;border:1px solid #e6e9ee;border-radius:8px;">
  {{range $i, $r := .}}<tr>
    <td width="130" style="width:130px;padding:10px 14px;{{if $i}}border-top:1px solid #e6e9ee;{{end}}font-size:13px;color:#667382;vertical-align:top;">{{$r.Label}}</td>
    <td style="padding:10px 14px;{{if $i}}border-top:1px solid #e6e9ee;{{end}}font-size:13px;color:#182433;word-break:break-all;{{if $r.Mono}}font-family:SFMono-Regular,Menlo,Consolas,monospace;{{end}}">{{if $r.Link}}<a href="{{$r.Link}}" target="_blank" style="color:#066fd1;text-decoration:none;">{{$r.Value}}</a>{{else}}{{$r.Value}}{{end}}</td>
  </tr>{{end}}
</table>{{end}}

{{define "userInfo"}}<h2 style="margin:8px 0 10px;font-size:15px;font-weight:600;color:#182433;">Account</h2>
{{template "kv" .Account}}
{{if .WithVPN}}{{if .QRSrc}}<h2 style="margin:8px 0 10px;font-size:15px;font-weight:600;color:#182433;">TOTP</h2>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="margin:0 0 20px;border:1px solid #e6e9ee;border-radius:8px;"><tr>
  <td width="168" style="width:168px;padding:12px;"><img src="{{.QRSrc}}" width="144" height="144" alt="TOTP QR code" style="display:block;width:144px;height:144px;border:0;"></td>
  <td style="padding:12px 16px 12px 0;font-size:13px;color:#3a4859;vertical-align:middle;">Scan this QR code with an authenticator app (e.g. Google Authenticator, andOTP). The app then shows a 6-digit <strong>TOTP code</strong> that changes every 30 seconds.</td>
</tr></table>{{end}}{{end}}
<h2 style="margin:8px 0 10px;font-size:15px;font-weight:600;color:#182433;">Getting started</h2>
<ol style="margin:0 0 16px;padding-left:20px;font-size:13px;color:#3a4859;">
  {{if .WithVPN}}<li style="margin:0 0 6px;">Scan the TOTP QR code with an authenticator app.</li>
  <li style="margin:0 0 6px;">Connect to the VPN with an OpenConnect client (e.g. Cisco Secure Client) using your username and the <strong>VPN password</strong> (see below).</li>{{end}}
  <li style="margin:0 0 6px;">Sign in to intranet systems (GitLab, Nexus, the intranet portal, etc.) with your LDAP username and password.</li>
  <li style="margin:0 0 6px;">Change your initial password in the intranet portal.</li>
  <li style="margin:0;">Forgot your password? Use <strong>Forgot password</strong> in the portal's sign-in dialog.</li>
</ol>
{{if .PortalURL}}{{template "button" (btn "Open intranet portal" .PortalURL)}}{{end}}
{{if .WithVPN}}<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="margin:20px 0 0;"><tr>
  <td style="padding:12px 16px;background:#e9f2fc;border-left:3px solid #066fd1;border-radius:4px;font-size:13px;color:#3a4859;">
    <strong style="color:#182433;">VPN password</strong> = LDAP password + current TOTP code<br>
    e.g. LDAP password <span style="font-family:SFMono-Regular,Menlo,Consolas,monospace;">ldappasswd</span> and TOTP code <span style="font-family:SFMono-Regular,Menlo,Consolas,monospace;">123456</span>
    &rarr; VPN password <span style="font-family:SFMono-Regular,Menlo,Consolas,monospace;font-weight:600;color:#182433;">ldappasswd123456</span>
  </td>
</tr></table>{{end}}{{end}}
`))
