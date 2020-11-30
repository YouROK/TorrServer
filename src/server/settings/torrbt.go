package settings

import (
	"encoding/json"

	"server/log"
)

type BTSets struct {
	CacheSize         int64 // in byte, def 200 mb
	PreloadBufferSize int64 // in byte, buffer for preload

	SaveOnDisk  bool   // save on disk?
	ContentPath string // path to save content

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
	BTsets = sets
}
