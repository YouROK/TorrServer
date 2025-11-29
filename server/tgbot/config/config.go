package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"server/log"
	"server/settings"
)

type Config struct {
	HostTG   string
	HostWeb  string
	WhiteIds []int64
	BlackIds []int64
}

var Cfg *Config

func LoadConfig() {
	Cfg = &Config{}
	fn := filepath.Join(settings.Path, "tg.cfg")
	buf, err := os.ReadFile(fn)
	if err != nil {
		Cfg.WhiteIds = []int64{}
		Cfg.BlackIds = []int64{}
		Cfg.HostTG = "https://api.telegram.org"
		buf, _ = json.MarshalIndent(Cfg, "", " ")
		if buf != nil {
			os.WriteFile(fn, buf, 0o666)
		}
		return
	}
	err = json.Unmarshal(buf, &Cfg)
	if err != nil {
		log.TLogln("Error read tg config:", err)
	}
	if Cfg.HostTG == "" {
		Cfg.HostTG = "https://api.telegram.org"
	}
}
