package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSiteLogoFile(t *testing.T) {
	dir := t.TempDir()
	old := Conf.System
	t.Cleanup(func() { Conf.System = old })

	write := func(name string, size int) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, make([]byte, size), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	svg := write("logo.svg", 100)
	big := write("big.png", siteLogoMaxBytes+1)
	txt := write("logo.txt", 10)

	cases := []struct {
		name, path, wantType string
		ok                   bool
	}{
		{"empty", "", "", false},
		{"svg", svg, "image/svg+xml", true},
		{"upper-case ext", write("LOGO.PNG", 10), "image/png", true},
		{"too big", big, "", false},
		{"not an image", txt, "", false},
		{"missing", filepath.Join(dir, "nope.svg"), "", false},
		{"directory", dir + "/", "", false},
	}
	for _, c := range cases {
		Conf.System = &System{SiteLogo: c.path}
		_, ct, _, ok := SiteLogoFile()
		if ok != c.ok || ct != c.wantType {
			t.Errorf("%s: got ok=%v type=%q, want ok=%v type=%q", c.name, ok, ct, c.ok, c.wantType)
		}
	}

	// 相对路径按工作目录解析（与 rsa 密钥一致）
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	Conf.System = &System{SiteLogo: "logo.svg"}
	if _, _, _, ok := SiteLogoFile(); !ok {
		t.Error("relative path should resolve against the working directory")
	}
}

func TestSiteName(t *testing.T) {
	old := Conf.System
	t.Cleanup(func() { Conf.System = old })
	Conf.System = &System{SiteName: "  "}
	if SiteName() != DefaultSiteName {
		t.Errorf("blank site-name should fall back to %q", DefaultSiteName)
	}
	Conf.System = &System{SiteName: "MyProject"}
	if SiteName() != "MyProject" {
		t.Errorf("got %q", SiteName())
	}
}
