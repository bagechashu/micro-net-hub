package config

import (
	"path/filepath"
	"testing"
)

func TestResolvePath(t *testing.T) {
	old := configDir
	t.Cleanup(func() { configDir = old })
	configDir = filepath.FromSlash("/app/conf")

	cases := map[string]string{
		"":                  "",
		"auth-pub.pem":      filepath.FromSlash("/app/conf/auth-pub.pem"),
		"keys/auth-pub.pem": filepath.FromSlash("/app/conf/keys/auth-pub.pem"),
		" logo.svg ":        filepath.FromSlash("/app/conf/logo.svg"),
		"/etc/mnh/key.pem":  "/etc/mnh/key.pem",
	}
	for in, want := range cases {
		if got := ResolvePath(in); got != want {
			t.Errorf("ResolvePath(%q) = %q, want %q", in, got, want)
		}
	}

	configDir = ""
	if got := ResolvePath("auth-pub.pem"); got != "auth-pub.pem" {
		t.Errorf("without configDir: got %q", got)
	}
}
