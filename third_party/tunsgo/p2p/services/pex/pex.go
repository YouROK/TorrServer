package pex

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/YouROK/tunsgo/opts"
	"github.com/YouROK/tunsgo/p2p/models"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type Pex struct {
	host host.Host
	opts *opts.Options
	ctx  context.Context

	lastSeeded   map[peer.ID]time.Time
	muLastSeeded sync.RWMutex
	sem          chan struct{}
}

func NewPex(c *models.SrvCtx) *Pex {
	return &Pex{
		host:       c.Host,
		opts:       c.Opts,
		ctx:        c.Ctx,
		lastSeeded: make(map[peer.ID]time.Time),
		sem:        make(chan struct{}, 5),
	}
}

func (p *Pex) Start() error {
	log.Println("[PEX] Service started")

	go p.subscribeToEvents()
	go p.gcLoop()

	return nil
}

func (p *Pex) Stop() {
	log.Println("[PEX] Service stoping...")
}

func (p *Pex) Name() string {
	return "Pex"
}

func (p *Pex) ProtocolID() protocol.ID {
	return "/tunsgo/pex/1.0.0"
}

func (p *Pex) HandleStream(stream network.Stream) {
	defer stream.Close()

	remotePeer := stream.Conn().RemotePeer()

	candidates := p.collectAddrInfos(remotePeer)

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	if len(candidates) > 30 {
		candidates = candidates[:30]
	}

	json.NewEncoder(stream).Encode(candidates)
}

func (p *Pex) collectAddrInfos(exclude peer.ID) []peer.AddrInfo {
	added := make(map[peer.ID]bool)
	added[p.host.ID()] = true
	added[exclude] = true

	res := make([]peer.AddrInfo, 0, 30)

	for _, pid := range p.host.Network().Peers() {
		if !added[pid] && p.isTunsGoPeer(pid) {
			info := p.host.Peerstore().PeerInfo(pid)
			if len(info.Addrs) > 0 {
				res = append(res, info)
				added[pid] = true
			}
		}
		if len(res) >= 60 {
			break
		}
	}

	if len(res) < 30 {
		for _, pid := range p.host.Peerstore().Peers() {
			if !added[pid] && p.isTunsGoPeer(pid) {
				info := p.host.Peerstore().PeerInfo(pid)
				if len(info.Addrs) > 0 {
					res = append(res, info)
					added[pid] = true
				}
			}
			if len(res) >= 60 {
				break
			}
		}
	}
	return res
}

func (p *Pex) subscribeToEvents() {
	sub, _ := p.host.EventBus().Subscribe(new(event.EvtPeerIdentificationCompleted))
	defer sub.Close()

	for {
		select {
		case <-p.ctx.Done():
			return
		case e := <-sub.Out():
			evt := e.(event.EvtPeerIdentificationCompleted)

			if len(p.host.Network().Peers()) < p.opts.P2P.LowConns {
				p.checkAndRequest(evt.Peer)
			}
		}
	}
}

func (p *Pex) checkAndRequest(id peer.ID) {
	protocols, _ := p.host.Peerstore().SupportsProtocols(id, p.ProtocolID())
	if len(protocols) == 0 {
		return
	}

	p.muLastSeeded.RLock()
	last, ok := p.lastSeeded[id]
	p.muLastSeeded.RUnlock()

	if ok && time.Since(last) < 5*time.Minute {
		return
	}

	select {
	case p.sem <- struct{}{}:
		go func() {
			defer func() { <-p.sem }()
			p.requestPeersFrom(id)
		}()
	default:
	}
}

func (p *Pex) requestPeersFrom(id peer.ID) {
	ctx, cancel := context.WithTimeout(p.ctx, 10*time.Second)
	defer cancel()

	stream, err := p.host.NewStream(ctx, id, p.ProtocolID())
	if err != nil {
		return
	}
	defer stream.Close()

	var discovered []peer.AddrInfo
	if err := json.NewDecoder(stream).Decode(&discovered); err != nil {
		return
	}

	p.muLastSeeded.Lock()
	p.lastSeeded[id] = time.Now()
	p.muLastSeeded.Unlock()

	for _, info := range discovered {
		if info.ID == p.host.ID() {
			continue
		}
		p.host.Peerstore().AddAddrs(info.ID, info.Addrs, time.Hour)
	}
}

func (p *Pex) isTunsGoPeer(id peer.ID) bool {
	protocols, err := p.host.Peerstore().SupportsProtocols(id, "/tunsgo")
	return err == nil && len(protocols) > 0
}

func (p *Pex) gcLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.muLastSeeded.Lock()
			now := time.Now()
			for id, t := range p.lastSeeded {
				if now.Sub(t) > 1*time.Hour {
					delete(p.lastSeeded, id)
				}
			}
			p.muLastSeeded.Unlock()
		}
	}
}
