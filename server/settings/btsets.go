package settings

import (
	"encoding/json"

	"server/log"
)

type BTSets struct {
	// Cache
	CacheSize       int64 // in byte, def 200 mb
	PreloadBuffer   bool
	ReaderReadAHead int // in percent, 5%-100%, [...S__X__E...] [S-E] not clean

	// Torrent
	ForceEncrypt             bool
	RetrackersMode           int  // 0 - don`t add, 1 - add retrackers (def), 2 - remove retrackers 3 - replace retrackers
	TorrentDisconnectTimeout int  // in seconds
	EnableDebug              bool // print logs

	// BT Config
	EnableIPv6         bool
	DisableTCP         bool
	DisableUTP         bool
	DisableUPNP        bool
	DisableDHT         bool
	DisablePEX         bool
	DisableUpload      bool
	DownloadRateLimit  int // in kb, 0 - inf
	UploadRateLimit    int // in kb, 0 - inf
	ConnectionsLimit   int
	DhtConnectionLimit int // 0 - inf
	PeersListenPort    int
	Strategy           int // 0 - RequestStrategyDuplicateRequestTimeout, 1 - RequestStrategyFuzzing, 2 - RequestStrategyFastest

	// Proxy
	ProxyURL                   string
	UseProxyForPeerConnections bool	//false - use proxy only for tracker connections, true - use proxy for tracker and peer connections
}

func (v *BTSets) String() string {
	buf, _ := json.Marshal(v)
	return string(buf)
}

var (
	BTsets *BTSets
)

func SetBTSets(sets *BTSets) {
	if ReadOnly {
		return
	}

	if sets.ReaderReadAHead < 5 {
		sets.ReaderReadAHead = 5
	}
	if sets.ReaderReadAHead > 100 {
		sets.ReaderReadAHead = 100
	}
	BTsets = sets
	buf, err := json.Marshal(BTsets)
	if err != nil {
		log.TLogln("Error marshal btsets", err)
		return
	}
	tdb.Set("Settings", "BitTorr", buf)
}

func loadBTSets() {
	buf := tdb.Get("Settings", "BitTorr")
	if len(buf) > 0 {
		err := json.Unmarshal(buf, &BTsets)
		if err == nil {
			if BTsets.ReaderReadAHead < 5 {
				BTsets.ReaderReadAHead = 5
			}
			return
		}
		log.TLogln("Error unmarshal btsets", err)
	}

	SetDefault()
}

func SetDefault() {
	sets := new(BTSets)
	sets.EnableDebug = false
	sets.DisableUTP = true
	sets.CacheSize = 100 * 1024 * 1024 // 100mb
	sets.PreloadBuffer = false
	sets.ConnectionsLimit = 25
	sets.DhtConnectionLimit = 500
	sets.RetrackersMode = 1
	sets.TorrentDisconnectTimeout = 30
	sets.ReaderReadAHead = 95 // 95% preload
	BTsets = sets
}
