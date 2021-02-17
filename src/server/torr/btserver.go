package torr

import (
	"log"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"server/settings"
	"server/torr/storage/torrstor"
	"server/torr/utils"
)

type BTServer struct {
	config *torrent.ClientConfig
	client *torrent.Client

	storage *torrstor.Storage

	torrents map[metainfo.Hash]*Torrent

	mu sync.Mutex
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
	InitApiHelper(bt)
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

func (bt *BTServer) configure() {
	blocklist, _ := utils.ReadBlockedIP()
	bt.config = torrent.NewDefaultClientConfig()

	if !settings.BTsets.SaveOnDisk {
		bt.storage = torrstor.NewStorage(settings.BTsets.CacheSize)
		bt.config.DefaultStorage = bt.storage
	} else {
		bt.config.DataDir = settings.BTsets.ContentPath
	}

	userAgent := "qBittorrent/4.3.2"
	peerID := "-qB4320-"
	cliVers := userAgent //"uTorrent/2210(25302)"

	bt.config.Debug = settings.BTsets.EnableDebug
	bt.config.DisableIPv6 = settings.BTsets.EnableIPv6 == false
	bt.config.DisableTCP = settings.BTsets.DisableTCP
	bt.config.DisableUTP = settings.BTsets.DisableUTP
	bt.config.NoDefaultPortForwarding = settings.BTsets.DisableUPNP
	bt.config.NoDHT = settings.BTsets.DisableDHT
	bt.config.NoUpload = settings.BTsets.DisableUpload
	bt.config.IPBlocklist = blocklist
	bt.config.Bep20 = peerID
	bt.config.PeerID = utils.PeerIDRandom(peerID)
	bt.config.HTTPUserAgent = userAgent
	bt.config.ExtendedHandshakeClientVersion = cliVers
	bt.config.EstablishedConnsPerTorrent = settings.BTsets.ConnectionsLimit
	bt.config.UpnpID = "YouROK/TorrServer"

	//bt.config.DropMutuallyCompletePeers = true
	//bt.config.DropDuplicatePeerIds = true

	// Encryption/Obfuscation
	bt.config.HeaderObfuscationPolicy = torrent.HeaderObfuscationPolicy{
		RequirePreferred: settings.BTsets.ForceEncrypt,
		Preferred:        true,
	}

	switch settings.BTsets.Strategy {
	case 1: // RequestStrategyFuzzing
		bt.config.DefaultRequestStrategy = torrent.RequestStrategyFuzzing()
	case 2: // RequestStrategyFastest
		bt.config.DefaultRequestStrategy = torrent.RequestStrategyFastest()
	default: // RequestStrategyDuplicateRequestTimeout
		bt.config.DefaultRequestStrategy = torrent.RequestStrategyDuplicateRequestTimeout(5 * time.Second)
	}

	if settings.BTsets.DhtConnectionLimit > 0 {
		bt.config.ConnTracker.SetMaxEntries(settings.BTsets.DhtConnectionLimit)
	}
	if settings.BTsets.DownloadRateLimit > 0 {
		bt.config.DownloadRateLimiter = utils.Limit(settings.BTsets.DownloadRateLimit * 1024)
	}
	if settings.BTsets.UploadRateLimit > 0 {
		bt.config.UploadRateLimiter = utils.Limit(settings.BTsets.UploadRateLimit * 1024)
	}
	if settings.BTsets.PeersListenPort > 0 {
		bt.config.ListenPort = settings.BTsets.PeersListenPort
	} else {
		bt.config.ListenPort = 50109
	}

	log.Println("Configure client:", settings.BTsets)
}

func (bt *BTServer) GetTorrent(hash torrent.InfoHash) *Torrent {
	if torr, ok := bt.torrents[hash]; ok {
		return torr
	}
	return nil
}

func (bt *BTServer) ListTorrents() map[metainfo.Hash]*Torrent {
	list := make(map[metainfo.Hash]*Torrent)
	for k, v := range bt.torrents {
		list[k] = v
	}
	return list
}

func (bt *BTServer) RemoveTorrent(hash torrent.InfoHash) {
	if torr, ok := bt.torrents[hash]; ok {
		torr.Close()
	}
}
