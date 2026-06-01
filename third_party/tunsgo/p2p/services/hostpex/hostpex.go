package hostpex

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/YouROK/tunsgo/opts"
	"github.com/YouROK/tunsgo/p2p/models"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type HostPex struct {
	host host.Host
	opts *opts.Options
	ctx  context.Context
	dht  *dht.IpfsDHT

	peers   map[peer.ID]*models.PeerInfo
	muPeers *sync.RWMutex

	lastSeeded   map[peer.ID]time.Time
	muLastSeeded sync.RWMutex

	maxPeers    int
	maxPerReply int
	sem         chan struct{}
}

func NewHostPex(c *models.SrvCtx) *HostPex {
	return &HostPex{
		host:        c.Host,
		opts:        c.Opts,
		ctx:         c.Ctx,
		dht:         c.Dht,
		peers:       c.Peers,
		muPeers:     &c.MuPeers,
		lastSeeded:  make(map[peer.ID]time.Time),
		maxPeers:    500,
		maxPerReply: 100,
		sem:         make(chan struct{}, 10),
	}
}

func (p *HostPex) Start() error {
	log.Println("[HOSTPEX] Service started")

	go p.subscribeToEvents()
	go p.gcLoop()
	go p.backgroundDiscovery()

	return nil
}

func (p *HostPex) Stop() {
	log.Println("[HOSTPEX] Service stoping...")
}

func (p *HostPex) Name() string {
	return "HostPex"
}

func (p *HostPex) ProtocolID() protocol.ID {
	return "/tunsgo/hostpex/1.0.0"
}

func (p *HostPex) HandleStream(stream network.Stream) {
	defer stream.Close()

	remotePeer := stream.Conn().RemotePeer()
	peers := p.collectPeersForReply(remotePeer)

	log.Printf("[HOSTPEX] Send peers to %s, count: %v", stream.Conn().RemotePeer().String(), len(peers))

	if err := json.NewEncoder(stream).Encode(peers); err != nil {
		log.Printf("[HOSTPEX] Encode error to %s: %v", remotePeer, err)
	}
}

func (p *HostPex) collectPeersForReply(remote peer.ID) []*models.PeerInfo {
	p.muPeers.RLock()
	defer p.muPeers.RUnlock()

	tmp := make([]*models.PeerInfo, 0, len(p.peers)+1)

	if len(p.opts.Hosts) > 0 {
		tmp = append(tmp, &models.PeerInfo{
			PeerID:   p.host.ID().String(),
			Hosts:    p.opts.Hosts,
			LastSeen: time.Now(),
		})
	}

	for pid, info := range p.peers {
		if pid == remote || len(info.Hosts) == 0 {
			continue
		}
		tmp = append(tmp, info)
	}

	if len(tmp) == 0 {
		return nil
	}

	rand.Shuffle(len(tmp), func(i, j int) {
		tmp[i], tmp[j] = tmp[j], tmp[i]
	})

	limit := p.maxPerReply
	if len(tmp) < limit {
		limit = len(tmp)
	}
	return tmp[:limit]
}

func (p *HostPex) subscribeToEvents() {
	sub, err := p.host.EventBus().Subscribe(new(event.EvtPeerIdentificationCompleted))
	if err != nil {
		log.Printf("[HOSTPEX] EventBus error: %v", err)
		return
	}
	defer sub.Close()

	for {
		select {
		case <-p.ctx.Done():
			return
		case e := <-sub.Out():
			evt := e.(event.EvtPeerIdentificationCompleted)
			p.checkAndRequest(evt.Peer)
		}
	}
}

func (p *HostPex) checkAndRequest(pid peer.ID) {
	protocols, err := p.host.Peerstore().SupportsProtocols(pid, p.ProtocolID())
	if err != nil || len(protocols) == 0 {
		return
	}

	p.muLastSeeded.RLock()
	last, ok := p.lastSeeded[pid]
	p.muLastSeeded.RUnlock()

	if ok && time.Since(last) < 5*time.Minute {
		return
	}

	select {
	case p.sem <- struct{}{}:
		go func() {
			defer func() { <-p.sem }()
			p.requestHostsFrom(pid)
		}()
	default:
	}
}

func (p *HostPex) requestHostsFrom(id peer.ID) {
	log.Println("[HOSTPEX] Requesting hosts from ", id)

	ctx, cancel := context.WithTimeout(p.ctx, 10*time.Second)
	defer cancel()

	stream, err := p.host.NewStream(ctx, id, p.ProtocolID())
	if err != nil {
		return
	}
	defer stream.Close()

	var discovered []*models.PeerInfo
	if err := json.NewDecoder(stream).Decode(&discovered); err != nil {
		return
	}

	p.muLastSeeded.Lock()
	p.lastSeeded[id] = time.Now()
	p.muLastSeeded.Unlock()

	for _, info := range discovered {
		p.addPeer(info)
	}
}

func (p *HostPex) addPeer(info *models.PeerInfo) {
	pid, err := peer.Decode(info.PeerID)
	if err != nil || pid == p.host.ID() {
		return
	}

	p.muPeers.Lock()
	if len(p.peers) >= p.maxPeers {
		p.remOldest(10)
	}
	info.LastSeen = time.Now()
	p.peers[pid] = info
	p.muPeers.Unlock()
}

func (p *HostPex) remOldest(count int) {
	for i := 0; i < count; i++ {
		var oldestID peer.ID
		var oldestTime time.Time
		found := false

		for pid, info := range p.peers {
			if !found || info.LastSeen.Before(oldestTime) {
				oldestTime = info.LastSeen
				oldestID = pid
				found = true
			}
		}

		if found {
			delete(p.peers, oldestID)
		} else {
			break
		}
	}
}

func (p *HostPex) backgroundDiscovery() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			for _, pid := range p.host.Network().Peers() {
				p.checkAndRequest(pid)
			}
		}
	}
}

func (p *HostPex) gcLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.cleanup()
		}
	}
}

func (p *HostPex) cleanup() {
	now := time.Now()

	p.muPeers.Lock()
	for pid, info := range p.peers {
		if now.Sub(info.LastSeen) > 30*time.Minute {
			delete(p.peers, pid)
		}
	}
	p.muPeers.Unlock()

	p.muLastSeeded.Lock()
	for pid, t := range p.lastSeeded {
		if now.Sub(t) > 1*time.Hour {
			delete(p.lastSeeded, pid)
		}
	}
	p.muLastSeeded.Unlock()
}
