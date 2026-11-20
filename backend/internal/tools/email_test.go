package tools

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"micro-net-hub/internal/config"
)

// 1x1 透明 PNG
const testPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="

func setupMailConfig(t *testing.T, dir string) {
	t.Helper()
	old := *config.Conf
	t.Cleanup(func() { *config.Conf = old })
	config.Conf.System = &config.System{WriteTimeout: 10}
	config.Conf.Email = &config.Email{Enable: false}
	config.Conf.Notice = &config.Notice{
		ProjectName:              "micro-net-hub",
		ServiceDomain:            "portal.example.com",
		VpnInfoSendSwitch:        true,
		VPNServer:                "vpn.example.com",
		AccountCreatedNoticeSave: true,
		AccountCreatedNoticeDir:  dir,
		HeaderHTML:               `<p>Your intranet account has been approved.</p><p class="note">Please never share your account with anyone.</p>`,
		FooterHTML:               `<h3>Tools Recommended</h3><ul><li><span class="key">https://github.com/andOTP/andOTP</span></li></ul>`,
	}
}

func TestMailTemplatesEscapeAndEmbed(t *testing.T) {
	dir := t.TempDir()
	setupMailConfig(t, dir)

	evil := `<script>alert(1)</script>`
	if err := SendUserInfo([]string{"a@example.com"}, "bob"+evil, "p@ss<b>", testPNG); err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.html"))
	if len(files) != 1 {
		t.Fatalf("expected 1 saved notice, got %d", len(files))
	}
	saved, _ := os.ReadFile(files[0])
	html := string(saved)
	if strings.Contains(html, evil) || strings.Contains(html, "p@ss<b>") {
		t.Fatal("dynamic values must be HTML-escaped")
	}
	if !strings.Contains(html, "data:image/png;base64,"+testPNG) {
		t.Fatal("saved notice should inline the QR code as data URI")
	}
	if !strings.Contains(html, `class="note"`) {
		t.Fatal("header-html from config should be kept as-is")
	}

	// 发出去的正文用 cid: 引用二维码
	content, err := render("userInfo", userInfoMail{WithVPN: true, QRSrc: "cid:totp-qrcode.png"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), `src="cid:totp-qrcode.png"`) {
		t.Fatal("mail body should reference the embedded QR code")
	}
	if _, err := base64.StdEncoding.DecodeString(testPNG); err != nil {
		t.Fatal(err)
	}
}

// MAIL_PREVIEW_DIR=/some/dir go test ./internal/tools -run TestMailPreview 生成各邮件的预览 HTML
func TestMailPreview(t *testing.T) {
	out := os.Getenv("MAIL_PREVIEW_DIR")
	if out == "" {
		t.Skip("set MAIL_PREVIEW_DIR to write previews")
	}
	setupMailConfig(t, out)
	if err := SendUserInfo(nil, "alice", "Init#Pass2024", testPNG); err != nil {
		t.Fatal(err)
	}

	write := func(name string, p mailPage, tpl string, data any) {
		c, err := render(tpl, data)
		if err != nil {
			t.Fatal(err)
		}
		p.Content = c
		body, err := page(p)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("verification-code.html", mailPage{Title: "Verification code", Salutation: "alice"}, "code", codeMail{
		Lead:  "Use this verification code to reset your LDAP password:",
		Code:  "042317",
		Notes: []string{"The code is valid for 5 minutes.", "If you did not request this, you can ignore this email. Never share the code with anyone."},
	})
	write("new-password.html", mailPage{Title: "Password reset", Salutation: "alice"}, "code", codeMail{
		Lead: "Your LDAP password has been reset. Your new password is:", Code: "kP3#xQ9m",
		Notes:   []string{"Please sign in and change this password as soon as possible."},
		BtnText: "Change password", BtnURL: portalURL("#/profile/index"),
	})
	write("status-deactivated.html", mailPage{Title: "Accounts deactivated", Salutation: "Administrator"}, "status", statusMail{
		Status: "deactivated", Tone: "danger", Users: []string{"alice", "bob"},
		Explain:  "These users can no longer sign in to the VPN or intranet systems with their LDAP accounts.",
		Reminder: "Please take note.",
	})
}
