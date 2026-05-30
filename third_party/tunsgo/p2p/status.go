package p2p

import (
	"maps"
	"strings"
	"time"

	"github.com/YouROK/tunsgo/p2p/models"
	"github.com/libp2p/go-libp2p/core/peer"
)

type PeerDetail struct {
	ID        string    `json:"id"`
	Addrs     []string  `json:"addrs"`
	IsTuns    bool      `json:"is_tuns"`
	Latency   string    `json:"latency"`
	LastSeen  time.Time `json:"last_seen,omitempty"`
	Protocols []string  `json:"protocols"`
	Hosts     []string  `json:"hosts,omitempty"`
}

type P2PStatus struct {
	PeerID         string              `json:"peer_id"`
	ListenAddrs    []string            `json:"listen_addrs"`
	TotalConns     int                 `json:"total_conns"`
	TunsPeersCount int                 `json:"tuns_peers_count"`
	KnownDomains   map[string][]string `json:"known_domains"`
	Peers          []*PeerDetail       `json:"peers_list"`
	ConnectedPeers []*PeerDetail       `json:"connected_peers_list"`
	OurPeers       []*models.PeerInfo  `json:"our_peers_list"`
}

func (s *P2PServer) Status() *P2PStatus {
	status := &P2PStatus{
		PeerID:       s.host.ID().String(),
		TotalConns:   len(s.host.Network().Peers()),
		KnownDomains: make(map[string][]string),
		Peers:        []*PeerDetail{},
	}

	for _, addr := range s.host.Addrs() {
		status.ListenAddrs = append(status.ListenAddrs, addr.String())
	}

	peers := map[peer.ID]*models.PeerInfo{}
	s.srvctx.MuPeers.RLock()
	maps.Copy(peers, s.srvctx.Peers)
	s.srvctx.MuPeers.RUnlock()

	for _, info := range peers {
		status.OurPeers = append(status.OurPeers, info)
	}

	connectedPeers := s.host.Network().Peers()
	for _, pID := range connectedPeers {
		protocols, _ := s.host.Peerstore().GetProtocols(pID)

		var protocolsString []string
		isOurNode := false
		for _, proto := range protocols {
			protocolsString = append(protocolsString, string(proto))
			if strings.HasPrefix(string(proto), "/tunsgo/") {
				isOurNode = true
			}
		}

		if isOurNode {
			status.TunsPeersCount++
		}

		detail := &PeerDetail{
			ID:        pID.String(),
			IsTuns:    isOurNode,
			Protocols: protocolsString,
		}

		for _, a := range s.host.Peerstore().Addrs(pID) {
			detail.Addrs = append(detail.Addrs, a.String())
		}

		if info, ok := peers[pID]; ok {
			detail.LastSeen = info.LastSeen
			for _, domain := range info.Hosts {
				status.KnownDomains[domain] = append(status.KnownDomains[domain], pID.ShortString())
			}
		}

		status.Peers = append(status.Peers, detail)
		if isOurNode {
			status.ConnectedPeers = append(status.ConnectedPeers, detail)
		}
	}

	return status
}
