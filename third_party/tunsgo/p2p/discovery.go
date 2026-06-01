package p2p

import (
	"context"
	"log"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

func (s *P2PServer) startDiscovery() {
	time.Sleep(time.Second * 5)
	s.bootstrap()
	go s.announceDht()
	go s.discoveryPeers()
}

func (s *P2PServer) bootstrap() {
	successConnections := 0

	var wa sync.WaitGroup
	wa.Add(len(dht.DefaultBootstrapPeers))

	go func() {
		log.Println("[P2P] Bootstrap...")
		for _, addrStr := range dht.DefaultBootstrapPeers {
			addr, _ := multiaddr.NewMultiaddr(addrStr.String())
			pi, _ := peer.AddrInfoFromP2pAddr(addr)

			ctx, cancel := context.WithTimeout(s.ctx, 120*time.Second)
			if err := s.host.Connect(ctx, *pi); err == nil {
				successConnections++
			} else {
				log.Printf("[P2P] Error connect to bootstrap node %s: %v", pi.ID, err)
			}
			cancel()
			wa.Done()
		}
	}()

	wa.Wait()

	if successConnections > 0 {
		log.Printf("[P2P] Connected to %d bootstrap nodes", successConnections)
	} else {
		log.Println("[P2P] Warn: do not connect to any bootstrap nodes")
	}
}

func (s *P2PServer) announceDht() {

	if err := s.dht.Bootstrap(s.ctx); err != nil {
		log.Printf("[DHT] Error bootstrap DHT: %v", err)
	}

	ticker := time.NewTicker(3 * time.Hour)
	defer ticker.Stop()
	for {
		ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
		err := s.dht.Provide(ctx, s.cId, true)
		cancel()
		if err != nil {
			log.Printf("[DHT] Error announce to DHT: %v", err)
			time.Sleep(20 * time.Second)
			continue
		}

		select {
		case <-ticker.C:
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *P2PServer) discoveryPeers() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		ctx, cancel := context.WithTimeout(s.ctx, 1*time.Minute)
		peers := s.dht.FindProvidersAsync(ctx, s.cId, min(s.opts.P2P.LowConns, 20))

		for p := range peers {
			if p.ID == s.host.ID() || len(p.Addrs) == 0 {
				continue
			}

			if s.host.Network().Connectedness(p.ID) != network.Connected {
				err := s.host.Connect(s.ctx, p)
				if err == nil {
					s.host.ConnManager().UpsertTag(p.ID, "tuns-node", func(current int) int {
						return 100
					})
					s.srvctx.MuPeers.Lock()
					if info, ok := s.srvctx.Peers[p.ID]; ok {
						info.LastSeen = time.Now()
					}
					s.srvctx.MuPeers.Unlock()
				}
			}
		}
		cancel()

		select {
		case <-ticker.C:
		case <-s.ctx.Done():
			return
		}
	}
}
