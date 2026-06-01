package opts

type Options struct {
	Server struct {
		Port      string `yaml:"port"`
		Slots     int    `yaml:"slots"`
		SlotSleep int    `yaml:"slot_sleep"`
	} `yaml:"server"`

	P2P struct {
		LowConns int `yaml:"low_conns"`
		HiConns  int `yaml:"hi_conns"`
	} `yaml:"p2p"`

	Hosts []string `yaml:"provided_hosts"`
}

func DefOptions() *Options {
	cfg := &Options{}

	cfg.Server.Port = "8080"
	cfg.Server.Slots = 5
	cfg.Server.SlotSleep = 1

	cfg.P2P.LowConns = 50
	cfg.P2P.HiConns = 200

	cfg.Hosts = []string{"*themoviedb.org", "*tmdb.org"}

	return cfg
}
