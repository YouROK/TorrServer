package models

import (
	"context"
	"sync"

	"github.com/YouROK/tunsgo/opts"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type SrvCtx struct {
	Host host.Host
	Opts *opts.Options
	Ctx  context.Context
	Dht  *dht.IpfsDHT

	Slots chan struct{}

	Peers   map[peer.ID]*PeerInfo
	MuPeers sync.RWMutex
}
