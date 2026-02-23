package proxy

import (
	"server/log"
	"server/settings"

	"github.com/YouROK/tunsgo/opts"
	"github.com/YouROK/tunsgo/p2p"
)

var (
	P2Proxy *p2p.P2PServer
)

func Start() {
	if settings.BTsets.EnableProxy {
		cfg := opts.DefOptions()
		var err error

		cfg.Server.Port = settings.Args.Port
		cfg.Hosts = settings.BTsets.ProxyHosts

		P2Proxy, err = p2p.NewP2PServer(cfg)
		if err != nil {
			log.TLogln("Error starting P2PServer:", err)
			return
		}
	}
}

func Stop() {
	if P2Proxy != nil {
		P2Proxy.Stop()
		P2Proxy = nil
	}
}
