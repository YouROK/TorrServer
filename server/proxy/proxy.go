package proxy

import (
	"server/log"
	"server/settings"
	"server/version"

	"github.com/yourok/tunsgo/opts"
	"github.com/yourok/tunsgo/p2p"
)

var (
	P2Proxy *p2p.P2PServer
)

func Start() {
	//TODO Сделать все настройки в btsets.go
	if settings.BTsets.EnableProxy {
		opts := opts.DefOptions()
		ProtocolID := "/tunsgo/" + version.Version
		RendezvousString := "tunsgo-peers-0008"
		srv, err := p2p.NewP2PServer(ProtocolID, RendezvousString, opts)
		if err != nil {
			log.TLogln("Error starting P2PServer:", err)
			return
		}
		P2Proxy = srv
	}
}

func Stop() {
	if P2Proxy != nil {
		P2Proxy.Stop()
		P2Proxy = nil
	}
}
