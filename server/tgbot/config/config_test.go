package config

import (
	"os"
	"path/filepath"
	"testing"

	"server/settings"
)

func TestOfficialBotAPI(t *testing.T) {
	cases := []struct {
		host string
		want bool
	}{
		{"", true},
		{"https://api.telegram.org", true},
		{"https://api.telegram.org/", true},
		{"http://api.telegram.org", true},
		{"https://bot.telegram.org", true},
		{"http://127.0.0.1:8081", false},
		{"http://localhost:8081", false},
		{"https://example.com/bot", false},
	}
	for _, tc := range cases {
		if got := OfficialBotAPI(tc.host); got != tc.want {
			t.Errorf("OfficialBotAPI(%q)=%v want %v", tc.host, got, tc.want)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	oldPath, oldCfg := settings.Path, Cfg
	settings.Path = t.TempDir()
	t.Cleanup(func() {
		settings.Path = oldPath
		Cfg = oldCfg
	})
	LoadConfig()
	if Cfg.HostTG != "https://api.telegram.org" || Cfg.WhiteIds == nil || Cfg.BlackIds == nil {
		t.Fatalf("created %#v", Cfg)
	}
	fn := filepath.Join(settings.Path, "tg.cfg")
	if err := os.WriteFile(fn, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	LoadConfig()
	if Cfg.WhiteIds == nil {
		t.Fatal("bad json")
	}
	if err := os.WriteFile(fn, []byte(`{"HostTG":"ftp://x","WhiteIds":null,"BlackIds":null}`), 0o600); err != nil {
		t.Fatal(err)
	}
	LoadConfig()
	if Cfg.HostTG != "https://api.telegram.org" || Cfg.WhiteIds == nil || Cfg.BlackIds == nil {
		t.Fatalf("repaired %#v", Cfg)
	}
}
