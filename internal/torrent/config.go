package torrent

import (
	"silo/internal/torrent/storage/torrstor"
)

// Config содержит параметры P2P-клиента и кэша
type Config struct {
	ListenPort     int              `json:"listen_port"`      // 0 = случайный порт
	DownloadRateKB int              `json:"download_rate_kb"` // 0 = без ограничений
	UploadRateKB   int              `json:"upload_rate_kb"`   // 0 = без ограничений
	DisableDHT     bool             `json:"disable_dht"`      // Отключить DHT
	DisablePEX     bool             `json:"disable_pex"`      // Отключить PEX
	DisableUPNP    bool             `json:"disable_upnp"`     // Отключить UPnP
	DisableUTP     bool             `json:"disable_utp"`      // Отключить uTP
	DisableTCP     bool             `json:"disable_tcp"`      // Отключить TCP
	EnableIPv6     bool             `json:"enable_ipv6"`      // Включить IPv6
	Storage        *torrstor.Config `json:"storage"`          // Настройки RAM/дискового кэша
}

func DefaultConfig() *Config {
	return &Config{
		ListenPort:     0,
		DownloadRateKB: 0,
		UploadRateKB:   0,
		DisableDHT:     false,
		DisablePEX:     false,
		DisableUPNP:    false,
		DisableUTP:     false,
		DisableTCP:     false,
		EnableIPv6:     false,
		Storage:        torrstor.DefaultConfig(),
	}
}
