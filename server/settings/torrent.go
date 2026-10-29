package settings

import (
	"encoding/json"
	"sort"
	"sync"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type TorrentDB struct {
	*torrent.TorrentSpec

	Title    string `json:"title,omitempty"`
	Category string `json:"category,omitempty"`
	Poster   string `json:"poster,omitempty"`
	Data     string `json:"data,omitempty"`

	Timestamp int64 `json:"timestamp,omitempty"`
	Size      int64 `json:"size,omitempty"`
}

type File struct {
	Name string `json:"name,omitempty"`
	Id   int    `json:"id,omitempty"`
	Size int64  `json:"size,omitempty"`
}

var mu sync.Mutex

// OnTorrentChanged runs under mu after one Torrents row is written (row is its
// JSON) or removed (row is nil), so hooks see writes in DB order. It must not
// call back into settings. Package torr patches its decoded-row cache.
var OnTorrentChanged = func(metainfo.Hash, []byte) {}

// AddTorrent writes one row. It used to rewrite every row, one bbolt write
// transaction and a full JSON re-encode of InfoBytes each, on every add or edit.
func AddTorrent(torr *TorrentDB) {
	if tdb == nil {
		return
	}
	buf, err := json.Marshal(torr)
	if err != nil {
		return
	}
	dbMigrationLock.RLock()
	defer dbMigrationLock.RUnlock()
	mu.Lock()
	defer mu.Unlock()
	tdb.Set("Torrents", torr.InfoHash.HexString(), buf)
	OnTorrentChanged(torr.InfoHash, buf)
}

func ListTorrent() []*TorrentDB {
	if tdb == nil {
		return nil
	}
	// Use read lock to prevent migration during read
	dbMigrationLock.RLock()
	defer dbMigrationLock.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	var list []*TorrentDB
	keys := tdb.List("Torrents")
	for _, key := range keys {
		buf := tdb.Get("Torrents", key)
		if len(buf) > 0 {
			var torr *TorrentDB
			err := json.Unmarshal(buf, &torr)
			if err == nil {
				list = append(list, torr)
			}
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Timestamp > list[j].Timestamp
	})
	return list
}

func RemTorrent(hash metainfo.Hash) {
	if tdb == nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	tdb.Rem("Torrents", hash.HexString())
	OnTorrentChanged(hash, nil)
}
