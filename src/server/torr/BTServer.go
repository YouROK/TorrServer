package torr

import (
	"fmt"
	"io"
	"sync"

	"server/settings"
	"server/torr/storage/memcache"
	"server/torr/storage/state"
	"server/utils"

	"log"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type BTServer struct {
	config *torrent.ClientConfig
	client *torrent.Client

	storage *memcache.Storage

	torrents map[metainfo.Hash]*Torrent

	mu  sync.Mutex
	wmu sync.Mutex

	watching bool
}

func NewBTS() *BTServer {
	bts := new(BTServer)
	bts.torrents = make(map[metainfo.Hash]*Torrent)
	return bts
}

func (bt *BTServer) Connect() error {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	var err error
	bt.configure()
	bt.client, err = torrent.NewClient(bt.config)
	bt.torrents = make(map[metainfo.Hash]*Torrent)
	return err
}

func (bt *BTServer) Disconnect() {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	if bt.client != nil {
		bt.client.Close()
		bt.client = nil
		utils.FreeOSMemGC()
	}
}

func (bt *BTServer) Reconnect() error {
	bt.Disconnect()
	return bt.Connect()
}

func (bt *BTServer) configure() {
	bt.storage = memcache.NewStorage(settings.Get().CacheSize)

	//blocklist, _ := iplist.MMapPackedFile(filepath.Join(settings.Path, "blocklist"))
	blocklist, _ := utils.ReadBlockedIP()

	userAgent := "uTorrent/3.5.5"
	peerID := "-UT3550-"
	cliVers := "µTorrent 3.5.5"

	bt.config = torrent.NewDefaultClientConfig()

	bt.config.Debug = settings.Get().EnableDebug
	bt.config.DisableIPv6 = settings.Get().EnableIPv6 == false
	bt.config.DisableTCP = settings.Get().DisableTCP
	bt.config.DisableUTP = settings.Get().DisableUTP
	bt.config.NoDefaultPortForwarding = settings.Get().DisableUPNP
	bt.config.NoDHT = settings.Get().DisableDHT
	bt.config.NoUpload = settings.Get().DisableUpload
	bt.config.HeaderObfuscationPolicy = torrent.HeaderObfuscationPolicy {
		RequirePreferred: settings.Get().Encryption == 2, // Whether the value of Preferred is a strict requirement
		Preferred: settings.Get().Encryption != 1, // Whether header obfuscation is preferred
	}
	bt.config.IPBlocklist = blocklist
	bt.config.DefaultStorage = bt.storage
	bt.config.Bep20 = peerID
	bt.config.PeerID = utils.PeerIDRandom(peerID)
	bt.config.HTTPUserAgent = userAgent
	bt.config.ExtendedHandshakeClientVersion = cliVers
	bt.config.EstablishedConnsPerTorrent = settings.Get().ConnectionsLimit
	if settings.Get().DhtConnectionLimit > 0 {
		bt.config.ConnTracker.SetMaxEntries(settings.Get().DhtConnectionLimit)
	}
	if settings.Get().DownloadRateLimit > 0 {
		bt.config.DownloadRateLimiter = utils.Limit(settings.Get().DownloadRateLimit * 1024)
	}
	if settings.Get().UploadRateLimit > 0 {
		bt.config.UploadRateLimiter = utils.Limit(settings.Get().UploadRateLimit * 1024)
	}
	if settings.Get().PeersListenPort > 0 {
		bt.config.ListenPort = settings.Get().PeersListenPort
	}

	log.Println("Configure client:", settings.Get())
}

func (bt *BTServer) AddTorrent(magnet metainfo.Magnet, infobytes []byte, onAdd func(*Torrent)) (*Torrent, error) {
	torr, err := NewTorrent(magnet, infobytes, bt)
	if err != nil {
		return nil, err
	}

	if onAdd != nil {
		go func() {
			if torr.GotInfo() {
				onAdd(torr)
			}
		}()
	} else {
		go torr.GotInfo()
	}

	return torr, nil
}

func (bt *BTServer) List() []*Torrent {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	list := make([]*Torrent, 0)
	for _, t := range bt.torrents {
		list = append(list, t)
	}
	return list
}

func (bt *BTServer) GetTorrent(hash metainfo.Hash) *Torrent {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if t, ok := bt.torrents[hash]; ok {
		return t
	}

	return nil
}

func (bt *BTServer) RemoveTorrent(hash torrent.InfoHash) {
	if torr, ok := bt.torrents[hash]; ok {
		torr.Close()
	}
}

func (bt *BTServer) BTState() *BTState {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	btState := new(BTState)
	btState.LocalPort = bt.client.LocalPort()
	btState.PeerID = fmt.Sprintf("%x", bt.client.PeerID())
	btState.BannedIPs = len(bt.client.BadPeerIPs())
	for _, dht := range bt.client.DhtServers() {
		btState.DHTs = append(btState.DHTs, dht)
	}
	for _, t := range bt.torrents {
		btState.Torrents = append(btState.Torrents, t)
	}
	return btState
}

func (bt *BTServer) CacheState(hash metainfo.Hash) *state.CacheState {
	st := bt.GetTorrent(hash)
	if st == nil {
		return nil
	}

	cacheState := bt.storage.GetStats(hash)
	return cacheState
}

func (bt *BTServer) WriteState(w io.Writer) {
	bt.client.WriteStatus(w)
}
