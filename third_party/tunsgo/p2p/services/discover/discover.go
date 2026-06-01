package discover

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/YouROK/tunsgo/opts"
	"github.com/YouROK/tunsgo/p2p/models"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type Discover struct {
	host    host.Host
	opts    *opts.Options
	ctx     context.Context
	dht     *dht.IpfsDHT
	peers   map[peer.ID]*models.PeerInfo
	muPeers *sync.RWMutex
}

func NewDiscover(c *models.SrvCtx) *Discover {
	return &Discover{
		host:    c.Host,
		opts:    c.Opts,
		ctx:     c.Ctx,
		dht:     c.Dht,
		peers:   c.Peers,
		muPeers: &c.MuPeers,
	}
}

func (s *Discover) Start() error {
	log.Println("[DISCOVER] Service started")
	go s.discoveryLoop()
	return nil
}

func (s *Discover) Stop() {
	log.Println("[DISCOVER] Service stoping...")
}

func (s *Discover) Name() string {
	return "Discover"
}

func (s *Discover) ProtocolID() protocol.ID {
	return ""
}

func (s *Discover) HandleStream(stream network.Stream) {
	stream.Close()
}

func (s *Discover) discoveryLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.processPeers()
		}
	}
}

func (s *Discover) processPeers() {
	s.muPeers.RLock()
	peerIDs := make([]peer.ID, 0, len(s.peers))
	for pid := range s.peers {
		if pid == s.host.ID() {
			continue
		}
		peerIDs = append(peerIDs, pid)
	}
	s.muPeers.RUnlock()

	semaphore := make(chan struct{}, 5)

	for _, pid := range peerIDs {
		if s.host.Network().Connectedness(pid) == network.Connected {
			continue
		}

		semaphore <- struct{}{}
		go func(id peer.ID) {
			defer func() { <-semaphore }()

			if len(s.host.Peerstore().Addrs(id)) == 0 {
				s.findAndConnect(id)
			} else {
				s.connect(id)
			}
		}(pid)
	}
}

func (s *Discover) findAndConnect(pid peer.ID) {
	ctx, cancel := context.WithTimeout(s.ctx, 20*time.Second)
	defer cancel()

	_, err := s.dht.FindPeer(ctx, pid)
	if err != nil {
		return
	}

	s.connect(pid)
}

func (s *Discover) connect(pid peer.ID) {
	ctx, cancel := context.WithTimeout(s.ctx, 15*time.Second)
	defer cancel()

	if err := s.host.Connect(ctx, peer.AddrInfo{ID: pid}); err != nil {
		return
	}

	s.muPeers.Lock()
	_peer, ok := s.peers[pid]
	if !ok || _peer == nil {
		s.muPeers.Unlock()
		return
	}
	hosts := _peer.Hosts
	_peer.LastSeen = time.Now()
	s.muPeers.Unlock()
	log.Printf("[DISCOVER] Successfully connected to %s with hosts %v", pid, hosts)

	s.host.ConnManager().UpsertTag(pid, "tuns-node", func(current int) int {
		return 100
	})
}
