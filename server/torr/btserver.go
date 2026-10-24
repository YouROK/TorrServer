package torr

import (
	"context"
	"fmt"
	"log"
	"maps"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/dht/v2"
	"github.com/anacrolix/dht/v2/krpc"
	"github.com/anacrolix/publicip"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
	"server/torr/storage/torrstor"
	"server/torr/utils"
	"server/version"
)

type BTServer struct {
	config *torrent.ClientConfig
	client *torrent.Client

	storage *torrstor.Storage

	torrents map[metainfo.Hash]*Torrent

	mu        sync.RWMutex
	dhtCtx    context.Context
	dhtCancel context.CancelFunc
}

// dhtBootstrapHosts is resolved with a deadline in resolveDHTBootstrap.
// Tests replace the slice. GlobalBootstrapAddrs is not used: it looks up with
// context.Background and starts a refresher that never stops.
var dhtBootstrapHosts = append([]string(nil), dht.DefaultGlobalBootstrapHostPorts...)

const dhtBootstrapTimeout = 3 * time.Second

// lookupDHTHost resolves one bootstrap name. Tests replace it.
var lookupDHTHost = net.DefaultResolver.LookupIPAddr

var privateIPBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func NewBTS() *BTServer {
	bts := new(BTServer)
	bts.torrents = make(map[metainfo.Hash]*Torrent)
	return bts
}

func (bt *BTServer) Connect() error {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	bt.shutdownLocked()
	bt.dhtCtx, bt.dhtCancel = context.WithCancel(context.Background())
	utils.PrefetchTrackers()
	bt.configure()
	client, err := torrent.NewClient(bt.config)
	if err != nil {
		bt.shutdownLocked()
		return err
	}
	bt.client = client
	bt.torrents = make(map[metainfo.Hash]*Torrent)
	InitApiHelper(bt)
	return nil
}

func (bt *BTServer) Disconnect() {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	bt.shutdownLocked()
}

func (bt *BTServer) shutdownLocked() {
	utils.StopTrackers()
	if bt.dhtCancel != nil {
		bt.dhtCancel()
		bt.dhtCancel = nil
		bt.dhtCtx = nil
	}
	if bt.client == nil {
		return
	}
	for _, s := range bt.client.DhtServers() {
		if c, ok := s.(interface{ Close() }); ok {
			c.Close()
		}
	}
	bt.client.Close()
	bt.client = nil
	utils.FreeOSMemGC()
}

func (bt *BTServer) configure() {
	blocklist, _ := utils.ReadBlockedIP()
	bt.config = torrent.NewDefaultClientConfig()

	if settings.BTsets.EnableLPD {
		bt.config.LocalServiceDiscovery = &torrent.LocalServiceDiscoveryConfig{
			Ip6: settings.BTsets.LPDIPv6 && settings.BTsets.EnableIPv6,
		}
	} else {
		bt.config.LocalServiceDiscovery = nil
	}

	bt.storage = torrstor.NewStorage(settings.BTsets.CacheSize)
	bt.config.DefaultStorage = bt.storage

	userAgent := "qBittorrent/4.3.9"
	peerID := "-qB4390-"
	upnpID := "TorrServer/" + version.Version
	cliVers := userAgent

	bt.config.Debug = settings.BTsets.EnableDebug
	bt.config.DisableIPv6 = !settings.BTsets.EnableIPv6
	bt.config.DisableTCP = settings.BTsets.DisableTCP
	bt.config.DisableUTP = settings.BTsets.DisableUTP
	//	https://github.com/anacrolix/torrent/issues/703
	// bt.config.DisableWebtorrent = true //	NE
	// bt.config.DisableWebseeds = false  //	NE
	bt.config.NoDefaultPortForwarding = settings.BTsets.DisableUPNP
	bt.config.NoDHT = settings.BTsets.DisableDHT
	if !settings.BTsets.DisableDHT && bt.dhtCtx != nil {
		bt.config.ConfigureAnacrolixDhtServer = bt.configureDHT
	}
	bt.config.DisablePEX = settings.BTsets.DisablePEX
	bt.config.NoUpload = settings.BTsets.DisableUpload
	bt.config.IPBlocklist = blocklist
	bt.config.Bep20 = peerID
	bt.config.PeerID = utils.PeerIDRandom(peerID)
	bt.config.UpnpID = upnpID
	bt.config.HTTPUserAgent = userAgent
	bt.config.ExtendedHandshakeClientVersion = cliVers
	bt.config.EstablishedConnsPerTorrent = settings.BTsets.ConnectionsLimit
	bt.config.TotalHalfOpenConns = 500
	// Encryption/Obfuscation
	bt.config.EncryptionPolicy = torrent.EncryptionPolicy{ //	OE
		ForceEncryption: settings.BTsets.ForceEncrypt, //	OE
	} //	OE
	// bt.config.HeaderObfuscationPolicy = torrent.HeaderObfuscationPolicy{ //	NE
	// 	RequirePreferred: settings.BTsets.ForceEncrypt, //	NE
	// 	Preferred:        true,                         //	NE
	// } //	NE
	if settings.BTsets.DownloadRateLimit > 0 {
		bt.config.DownloadRateLimiter = utils.Limit(settings.BTsets.DownloadRateLimit * 1024)
	}
	if settings.BTsets.UploadRateLimit > 0 {
		bt.config.Seed = true
		bt.config.UploadRateLimiter = utils.Limit(settings.BTsets.UploadRateLimit * 1024)
	}
	if settings.TorAddr != "" {
		log.Println("Set listen addr", settings.TorAddr)
		bt.config.SetListenAddr(settings.TorAddr)
	} else {
		if settings.BTsets.PeersListenPort > 0 {
			log.Println("Set listen port", settings.BTsets.PeersListenPort)
			bt.config.ListenPort = settings.BTsets.PeersListenPort
		} else {
			log.Println("Set listen port to random autoselect (0)")
			bt.config.ListenPort = 0
		}
	}

	// Configure proxy if enabled
	if err := bt.configureProxy(); err != nil {
		log.Println("Proxy configuration error:", err)
	}

	log.Println("Client config:", settings.BTsets)

	var err error

	// set public IPv4
	if settings.PubIPv4 != "" {
		if ip4 := net.ParseIP(settings.PubIPv4); ip4.To4() != nil && !isPrivateIP(ip4) {
			bt.config.PublicIp4 = ip4
		}
	}
	if bt.config.PublicIp4 == nil {
		// Create a context with a 3-second timeout for the IP lookup
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		bt.config.PublicIp4, err = publicip.Get4(ctx)
		if err != nil {
			log.Printf("error getting public ipv4 address: %v", err)
		}
	}
	if bt.config.PublicIp4.To4() == nil { // possible IPv6 from publicip.Get4(ctx)
		bt.config.PublicIp4 = nil
	}
	if bt.config.PublicIp4 != nil {
		log.Println("PublicIp4:", bt.config.PublicIp4)
	}

	// set public IPv6
	if settings.PubIPv6 != "" {
		if ip6 := net.ParseIP(settings.PubIPv6); ip6.To16() != nil && ip6.To4() == nil && !isPrivateIP(ip6) {
			bt.config.PublicIp6 = ip6
		}
	}
	if bt.config.PublicIp6 == nil && settings.BTsets.EnableIPv6 {
		// Create a context with a 3-second timeout for the IP lookup
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		bt.config.PublicIp6, err = publicip.Get6(ctx)
		if err != nil {
			log.Printf("error getting public ipv6 address: %v", err)
		}
	}
	if bt.config.PublicIp6.To16() == nil { // just 4 sure it's valid IPv6
		bt.config.PublicIp6 = nil
	}
	if bt.config.PublicIp6 != nil {
		log.Println("PublicIp6:", bt.config.PublicIp6)
	}
}

func (bt *BTServer) configureProxy() error {
	proxyURL := settings.Args.ProxyURL

	if proxyURL == "" {
		return nil // No proxy configured
	}

	proxyMode := settings.Args.ProxyMode
	if proxyMode == "" {
		proxyMode = "tracker" // default
	}

	// Parse and validate proxy URL
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	scheme := parsedURL.Scheme
	// Validate proxy protocol
	switch scheme {
	case "socks5", "socks5h", "socks4", "socks4a", "http", "https":
		// Supported protocols
	default:
		return fmt.Errorf("unsupported proxy protocol: %s (supported: http, https, socks4, socks4a, socks5, socks5h)", scheme)
	}

	switch proxyMode {
	case "full":
		log.Printf("Configuring proxy for all BitTorrent traffic: %s://%s", scheme, parsedURL.Host)

		// Set ProxyURL - this will be used by anacrolix/torrent for all BitTorrent traffic
		bt.config.ProxyURL = proxyURL

		// Also set HTTPProxy explicitly for HTTP tracker requests
		bt.config.HTTPProxy = func(req *http.Request) (*url.URL, error) {
			return parsedURL, nil
		}

		log.Println("Proxy configured successfully for all BitTorrent connections (tracker, DHT, peers)")
	case "peers":
		log.Printf("Configuring proxy for peer connections only: %s://%s", scheme, parsedURL.Host)

		// Set ProxyURL for peer connections, but don't set HTTPProxy
		// This routes DHT and peer connections through proxy, but not HTTP tracker requests
		bt.config.ProxyURL = proxyURL

		log.Println("Proxy configured successfully for peer and DHT connections only")
	default:
		log.Printf("Configuring proxy for HTTP tracker requests only: %s://%s", scheme, parsedURL.Host)

		// Only set HTTPProxy for tracker requests, don't set ProxyURL
		bt.config.HTTPProxy = func(req *http.Request) (*url.URL, error) {
			return parsedURL, nil
		}

		log.Println("Proxy configured successfully for HTTP tracker connections only")
	}

	return nil
}

func (bt *BTServer) GetTorrent(hash torrent.InfoHash) *Torrent {
	bt.mu.RLock()
	torr := bt.torrents[hash]
	bt.mu.RUnlock()
	return torr
}

func (bt *BTServer) ListTorrents() map[metainfo.Hash]*Torrent {
	list := make(map[metainfo.Hash]*Torrent)
	bt.mu.RLock()
	maps.Copy(list, bt.torrents)
	bt.mu.RUnlock()
	return list
}

func (bt *BTServer) RemoveTorrent(hash torrent.InfoHash) bool {
	// Torrent.Close takes bt.mu, so the lock must be released before calling it
	if torr := bt.GetTorrent(hash); torr != nil {
		return torr.Close()
	}
	return false
}

func (bt *BTServer) configureDHT(cfg *dht.ServerConfig) {
	want4, want6 := dhtSocketFamily(cfg.Conn)
	if settings.BTsets == nil || !settings.BTsets.EnableIPv6 {
		want6 = false
	}
	if !want4 && !want6 {
		want4 = true
	}
	ctx := bt.dhtCtx
	cfg.StartingNodes = func() ([]dht.Addr, error) {
		addrs, err := resolveDHTBootstrap(ctx, dhtBootstrapHosts, want4, want6)
		if err != nil || len(addrs) == 0 {
			// Bootstrap returns before stopping its traversal when StartingNodes
			// errors, which leaks that goroutine. One blackhole node lets it finish.
			return []dht.Addr{dht.NewAddr(&net.UDPAddr{IP: net.IPv4(192, 0, 2, 1), Port: 1})}, nil
		}
		return addrs, nil
	}
	switch {
	case want4 && !want6:
		cfg.DefaultWant = []krpc.Want{krpc.WantNodes}
	case want6 && !want4:
		cfg.DefaultWant = []krpc.Want{krpc.WantNodes6}
	}
}

func dhtSocketFamily(c net.PacketConn) (want4, want6 bool) {
	if c == nil || c.LocalAddr() == nil {
		return true, false
	}
	host, _, err := net.SplitHostPort(c.LocalAddr().String())
	if err != nil {
		return true, false
	}
	ip := net.ParseIP(host)
	if ip == nil || ip.To4() != nil {
		return true, false
	}
	return false, true
}

// resolveDHTBootstrap resolves bootstrap host:port pairs under ctx and one
// extra deadline. want4/want6 drop the other family so an IPv4 socket does not
// walk unreachable IPv6 routers.
func resolveDHTBootstrap(ctx context.Context, hosts []string, want4, want6 bool) ([]dht.Addr, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, dhtBootstrapTimeout)
	defer cancel()
	// ponytail: sequential lookups share one deadline. Fan out if bootstrap latency matters.
	var addrs []dht.Addr
	for _, hp := range hosts {
		if ctx.Err() != nil {
			break
		}
		host, portStr, err := net.SplitHostPort(hp)
		if err != nil {
			continue
		}
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1 || port > 65535 {
			continue
		}
		if !want6 && strings.EqualFold(host, "router.silotis.us") {
			continue
		}
		// One hung name must not consume the whole deadline.
		hctx, hcancel := context.WithTimeout(ctx, time.Second)
		ips, err := lookupDHTHost(hctx, host)
		hcancel()
		if err != nil {
			continue
		}
		for _, ip := range ips {
			is4 := ip.IP.To4() != nil
			if (is4 && !want4) || (!is4 && !want6) {
				continue
			}
			addrs = append(addrs, dht.NewAddr(&net.UDPAddr{IP: ip.IP, Port: port}))
		}
	}
	if len(addrs) == 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("no dht bootstrap nodes")
	}
	return addrs, nil
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}
