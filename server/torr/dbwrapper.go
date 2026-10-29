package torr

import (
	"encoding/json"
	"sync"

	"server/settings"
	"server/torr/state"
	"server/torr/utils"

	"github.com/anacrolix/torrent/metainfo"
)

func init() {
	settings.OnTorrentsChanged = invalidateDBRows
}

// Decoded settings rows, so a /torrents list poll does not unmarshal every row.
// The rows are read-only: callers get a new *Torrent from torrentFromDB each time,
// because GotInfo, Close and NewTorrent mutate the Torrent and its TorrentSpec.
var (
	dbRowsMu sync.Mutex
	dbRows   []*settings.TorrentDB
)

func invalidateDBRows() {
	dbRowsMu.Lock()
	dbRows = nil
	dbRowsMu.Unlock()
}

// listDBRows holds dbRowsMu across settings.ListTorrent so a mutation hook cannot
// clear the cache between the read and the store. nil (closed or empty DB) is not cached.
func listDBRows() []*settings.TorrentDB {
	dbRowsMu.Lock()
	defer dbRowsMu.Unlock()
	if dbRows == nil {
		dbRows = settings.ListTorrent()
	}
	return dbRows
}

func torrentFromDB(db *settings.TorrentDB) *Torrent {
	spec := *db.TorrentSpec
	spec.Trackers = append([][]string(nil), spec.Trackers...)
	torr := new(Torrent)
	torr.TorrentSpec = &spec
	torr.Title = db.Title
	torr.Poster = db.Poster
	torr.Category = db.Category
	torr.Timestamp = db.Timestamp
	torr.Size = db.Size
	torr.Data = db.Data
	torr.Stat = state.TorrentInDB
	return torr
}

type tsFiles struct {
	TorrServer struct {
		Files []*state.TorrentFileStat `json:"Files"`
	} `json:"TorrServer"`
}

// FileStatsFromData parses the file list persisted in Torrent.Data by AddTorrentDB.
// Empty or invalid JSON returns nil.
func FileStatsFromData(data string) []*state.TorrentFileStat {
	if data == "" {
		return nil
	}
	var files tsFiles
	if err := json.Unmarshal([]byte(data), &files); err != nil {
		return nil
	}
	if len(files.TorrServer.Files) == 0 {
		return nil
	}
	return files.TorrServer.Files
}

func AddTorrentDB(torr *Torrent) {
	// Status() reads these fields under muTorrent. list can run while add's
	// background save is still filling Data.
	torr.muTorrent.Lock()
	spec := torr.TorrentSpec
	title := torr.Title
	category := torr.Category
	data := torr.Data
	poster := torr.Poster
	size := torr.Size
	timestamp := torr.Timestamp
	bt := torr.Torrent
	torr.muTorrent.Unlock()

	t := new(settings.TorrentDB)
	t.TorrentSpec = spec
	t.Title = title
	t.Category = category
	if data == "" {
		files := new(tsFiles)
		// Status locks muTorrent itself. Do not hold it across this call.
		files.TorrServer.Files = torr.Status().FileStats
		if buf, err := json.Marshal(files); err == nil {
			data = string(buf)
			torr.muTorrent.Lock()
			if torr.Data == "" {
				torr.Data = data
			} else {
				data = torr.Data
			}
			torr.muTorrent.Unlock()
		}
	}
	t.Data = data

	t.Poster = poster
	if t.Poster != "" {
		var existing string
		if db := GetTorrentDB(torr.Hash()); db != nil {
			existing = db.Poster
		}
		if existing != t.Poster {
			ok, verified := utils.CheckImgUrl(t.Poster)
			if !ok || (existing != "" && !verified) {
				t.Poster = existing
			}
		}
	}
	t.Size = size
	if t.Size == 0 && bt != nil {
		t.Size = bt.Length()
	}
	// don't override timestamp from DB on edit
	if existing := GetTorrentDB(torr.Hash()); existing != nil && existing.Timestamp != 0 {
		t.Timestamp = existing.Timestamp
	} else {
		t.Timestamp = timestamp
	}

	settings.AddTorrent(t)
}

func GetTorrentDB(hash metainfo.Hash) *Torrent {
	for _, db := range listDBRows() {
		if db != nil && db.TorrentSpec != nil && db.InfoHash == hash {
			return torrentFromDB(db)
		}
	}
	return nil
}

func RemTorrentDB(hash metainfo.Hash) {
	settings.RemTorrent(hash)
}

func ListTorrentsDB() map[metainfo.Hash]*Torrent {
	rows := listDBRows()
	ret := make(map[metainfo.Hash]*Torrent, len(rows))
	for _, db := range rows {
		if db != nil && db.TorrentSpec != nil {
			ret[db.InfoHash] = torrentFromDB(db)
		}
	}
	return ret
}
