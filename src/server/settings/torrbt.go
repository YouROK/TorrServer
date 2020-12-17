package settings

import (
	"encoding/json"

	"server/log"
)

type BTSets struct {
	// Cache
	CacheSize         int64 // in byte, def 200 mb
	PreloadBufferSize int64 // in byte, buffer for preload

	// Reader
	ReaderPreload int // in percent, 32%-100%, [...S__X__E...] [S-E] not clean

	// Storage
	SaveOnDisk  bool   // save on disk?
	ContentPath string // path to save content

	// Torrent
	RetrackersMode           int  // 0 - don`t add, 1 - add retrackers (def), 2 - remove retrackers 3 - replace retrackers
	TorrentDisconnectTimeout int  // in seconds
	EnableDebug              bool // print logs

	// BT Config
	EnableIPv6         bool
	DisableTCP         bool
	DisableUTP         bool
	DisableUPNP        bool
	DisableDHT         bool
	DisableUpload      bool
	DownloadRateLimit  int // in kb, 0 - inf
	UploadRateLimit    int // in kb, 0 - inf
	ConnectionsLimit   int
	DhtConnectionLimit int // 0 - inf
	PeersListenPort    int
	Strategy           int // 0 - RequestStrategyDuplicateRequestTimeout, 1 - RequestStrategyFuzzing, 2 - RequestStrategyFastest
}

func (v *BTSets) String() string {
	buf, _ := json.Marshal(v)
	return string(buf)
}

var (
	BTsets *BTSets
)

func SetBTSets(sets *BTSets) {
	if tdb.ReadOnly {
		return
	}

	if sets.ReaderPreload < 32 {
		sets.ReaderPreload = 32
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
			if BTsets.ReaderPreload < 32 {
				BTsets.ReaderPreload = 32
			}
			return
		}
		log.TLogln("Error unmarshal btsets", err)
	}

	sets := new(BTSets)
	sets.EnableDebug = false
	sets.DisableUTP = true
	sets.CacheSize = 200 * 1024 * 1024 // 200mb
	sets.PreloadBufferSize = 20 * 1024 * 1024
	sets.ConnectionsLimit = 20
	sets.DhtConnectionLimit = 500
	sets.RetrackersMode = 1
	sets.TorrentDisconnectTimeout = 30
	sets.ReaderPreload = 70 // 70% preload
	BTsets = sets
}
