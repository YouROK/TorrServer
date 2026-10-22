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

	// OfflineState is "", "downloading", "paused" or "ready" (see server/torr/state).
	OfflineState string `json:"offline_state,omitempty"`
	// OfflineDir is the directory the offline data of this torrent was stored in
	// (so changing the global setting does not orphan already downloaded torrents).
	OfflineDir string `json:"offline_dir,omitempty"`
}

type File struct {
	Name string `json:"name,omitempty"`
	Id   int    `json:"id,omitempty"`
	Size int64  `json:"size,omitempty"`
}

var mu sync.Mutex

// AddTorrent stores (creates or replaces) ONE torrent record.
//
// It used to read the whole list first (outside the lock) and then write EVERY torrent back
// from that snapshot. Two concurrent calls for different torrents therefore overwrote each
// other: the second call re-wrote the first torrent with the stale state it had read before
// the first call finished (for example an offline torrent that had just been paused came back
// as "downloading"). A removal racing with an add could resurrect the removed torrent in the
// same way. Only the record being added is written now.
func AddTorrent(torr *TorrentDB) {
	buf, err := json.Marshal(torr)
	if err != nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if tdb == nil {
		return
	}
	tdb.Set("Torrents", torr.InfoHash.HexString(), buf)
}

func ListTorrent() []*TorrentDB {
	// Use read lock to prevent migration during read
	dbMigrationLock.RLock()
	defer dbMigrationLock.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	var list []*TorrentDB
	if tdb == nil {
		// the DB is not open (e.g. the torrent client runs in tests without a
		// DB): nothing is stored, so there is nothing to list
		return nil
	}
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
	mu.Lock()
	tdb.Rem("Torrents", hash.HexString())
	mu.Unlock()
}
