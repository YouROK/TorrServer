package p2p

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/YouROK/tunsgo/opts"
	"github.com/YouROK/tunsgo/p2p/models"
	"github.com/YouROK/tunsgo/p2p/services"
	"github.com/YouROK/tunsgo/p2p/services/discover"
	"github.com/YouROK/tunsgo/p2p/services/hostpex"
	"github.com/YouROK/tunsgo/p2p/services/pex"
	"github.com/YouROK/tunsgo/p2p/services/urlproxy"
	"github.com/YouROK/tunsgo/version"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	tls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/multiformats/go-multihash"
)

const Rendezvous = "tunsgo-peers-0009"

type P2PServer struct {
	host host.Host
	dht  *dht.IpfsDHT
	ctx  context.Context
	cm   *connmgr.BasicConnMgr
	cId  cid.Cid

	slots chan struct{}

	opts *opts.Options

	srvc   *services.Manager
	srvctx *models.SrvCtx

	urlprx *urlproxy.UrlProxy
}

func NewP2PServer(opts *opts.Options) (*P2PServer, error) {
	log.Println("[P2P Server] Starting...")
	log.Println("[P2P Server] Version:", version.Version)
	log.Println("[P2P Server] Provide hosts:", opts.Hosts)

	key, err := LoadOrCreateIdentity()
	if err != nil {
		return nil, err
	}

	if opts.Server.Slots < 1 { //min 1 slot
		opts.Server.Slots = 1
	}
	if opts.Server.SlotSleep < 0 { // min sleep 0
		opts.Server.SlotSleep = 0
	}
	if opts.Server.SlotSleep > 300 { //max sleep 5 min
		opts.Server.SlotSleep = 0
	}

	ctx := context.Background()

	cm, err := connmgr.NewConnManager(
		opts.P2P.LowConns,
		opts.P2P.HiConns,
		connmgr.WithGracePeriod(time.Second*30),
	)
	if err != nil {
		return nil, err
	}

	relayResources := relay.DefaultResources()
	if relayResources.Limit != nil {
		relayResources.Limit.Duration = 2 * time.Minute
		relayResources.Limit.Data = 1 << 21 //2 mb
	}

	optsLp2p := []libp2p.Option{
		libp2p.Identity(key),
		libp2p.ChainOptions(libp2p.DefaultPrivateTransports),
		libp2p.Security(tls.ID, tls.New),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/0",
		),
		libp2p.ConnectionManager(cm),
		libp2p.NATPortMap(),

		libp2p.EnableRelay(),
		libp2p.EnableRelayService(relay.WithResources(relayResources)),
		libp2p.EnableAutoRelayWithStaticRelays(nil),
		libp2p.EnableNATService(),
		libp2p.EnableHolePunching(),
	}

	h, err := libp2p.New(optsLp2p...)
	if err != nil {
		cm.Close()
		return nil, err
	}
	log.Println("[P2P] ID", h.ID().String())

	idht, err := dht.New(ctx, h, dht.Mode(dht.ModeAuto))
	if err != nil {
		cm.Close()
		return nil, err
	}

	pref := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   multihash.SHA2_256,
		MhLength: -1,
	}
	c, _ := pref.Sum([]byte(Rendezvous))

	srv := &P2PServer{
		host:  h,
		dht:   idht,
		ctx:   ctx,
		cm:    cm,
		opts:  opts,
		cId:   c,
		slots: make(chan struct{}, opts.Server.Slots),
		srvc:  services.NewManager(h),
	}

	go srv.startDiscovery()

	srvctx := &models.SrvCtx{
		Host:    srv.host,
		Opts:    srv.opts,
		Ctx:     srv.ctx,
		Dht:     srv.dht,
		Slots:   srv.slots,
		Peers:   make(map[peer.ID]*models.PeerInfo),
		MuPeers: sync.RWMutex{},
	}

	srv.srvctx = srvctx

	srv.urlprx = urlproxy.NewUrlProxy(srvctx)
	srv.srvc.AddService(srv.urlprx)
	srv.srvc.AddService(hostpex.NewHostPex(srvctx))
	srv.srvc.AddService(pex.NewPex(srvctx))
	srv.srvc.AddService(discover.NewDiscover(srvctx))

	err = srv.srvc.Start()
	if err != nil {
		cm.Close()
		return nil, err
	}

	go srv.watchNetworkStatus()

	return srv, nil
}

func (s *P2PServer) Stop() {
	log.Println("[P2P Server] Stoping...")
	s.srvc.Stop()

	s.dht.Close()
	s.host.Close()
	s.cm.Close()
}

func (s *P2PServer) watchNetworkStatus() {
	sub, err := s.host.EventBus().Subscribe([]interface{}{
		new(event.EvtLocalReachabilityChanged),
	})
	if err != nil {
		log.Printf("[NET] Failed to subscribe to reachability events: %v", err)
		return
	}

	go func() {
		defer sub.Close()

		for {
			select {
			case e, ok := <-sub.Out():
				if !ok {
					log.Println("[NET] Event bus channel closed")
					return
				}

				switch evt := e.(type) {
				case event.EvtLocalReachabilityChanged:
					switch evt.Reachability {
					case network.ReachabilityPublic:
						log.Println("[NET] Reachability changed: PUBLIC. Node is now operating as a Relay Hop")
					case network.ReachabilityPrivate:
						log.Println("[NET] Reachability changed: PRIVATE. Node is operating behind NAT (client mode)")
					case network.ReachabilityUnknown:
						log.Println("[NET] Reachability changed: UNKNOWN. Determining network status...")
					}
				}
			case <-s.ctx.Done():
				log.Println("[NET] Stopping network reachability")
				return
			}
		}
	}()
}
