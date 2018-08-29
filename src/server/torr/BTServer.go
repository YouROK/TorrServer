package torr

import (
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"server/settings"
	"server/torr/storage"
	"server/torr/storage/memcache"
	"server/torr/storage/state"
	"server/utils"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/iplist"
	"github.com/anacrolix/torrent/metainfo"
)

type BTServer struct {
	config *torrent.ClientConfig
	client *torrent.Client

	storage storage.Storage

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

	blocklist, _ := iplist.MMapPackedFile(filepath.Join(settings.Path, "blocklist"))

	userAgent := "uTorrent/3.4.9"
	peerID := "-UT3490-"

	bt.config = torrent.NewDefaultClientConfig()

	bt.config.DisableIPv6 = true
	bt.config.DisableTCP = settings.Get().DisableTCP
	bt.config.DisableUTP = settings.Get().DisableUTP
	bt.config.NoDefaultPortForwarding = settings.Get().DisableUPNP
	bt.config.NoDHT = settings.Get().DisableDHT
	bt.config.NoUpload = settings.Get().DisableUpload
	bt.config.EncryptionPolicy = torrent.EncryptionPolicy{
		DisableEncryption: settings.Get().Encryption == 1,
		ForceEncryption:   settings.Get().Encryption == 2,
	}
	bt.config.IPBlocklist = blocklist
	bt.config.DefaultStorage = bt.storage
	bt.config.Bep20 = peerID
	bt.config.PeerID = utils.PeerIDRandom(peerID)
	bt.config.HTTPUserAgent = userAgent
	bt.config.EstablishedConnsPerTorrent = settings.Get().ConnectionsLimit

	bt.config.TorrentPeersHighWater = 3000
	bt.config.HalfOpenConnsPerTorrent = 50

	if settings.Get().DownloadRateLimit > 0 {
		bt.config.DownloadRateLimiter = utils.Limit(settings.Get().DownloadRateLimit * 1024)
	}
	if settings.Get().UploadRateLimit > 0 {
		bt.config.UploadRateLimiter = utils.Limit(settings.Get().UploadRateLimit * 1024)
	}

	//bt.config.Debug = true

	fmt.Println("Configure client:", settings.Get())
}

func (bt *BTServer) AddTorrent(magnet metainfo.Magnet, onAdd func(*Torrent)) (*Torrent, error) {
	torr, err := NewTorrent(magnet, bt)
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
