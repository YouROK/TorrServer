package torr

import (
	"encoding/json"
	"slices"
	"sort"
	"sync"

	"server/settings"
	"server/torr/state"
	"server/torr/utils"

	"github.com/anacrolix/torrent/metainfo"
)

func init() {
	settings.OnTorrentsChanged = invalidateDBRows
	settings.OnTorrentChanged = patchDBRow
}

// Decoded settings rows, so a /torrents list poll does not unmarshal every row.
// The rows and the slice are read-only: a change builds a new slice, and callers
// get a new *Torrent from torrentFromDB each time, because GotInfo, Close and
// NewTorrent mutate the Torrent and its TorrentSpec.
var (
	dbRowsMu  sync.Mutex
	dbRows    []*settings.TorrentDB
	dbRowsGen uint64
)

func invalidateDBRows() {
	dbRowsMu.Lock()
	dbRows = nil
	dbRowsGen++
	dbRowsMu.Unlock()
}

// patchDBRow replaces, adds or (row nil) drops one cached row, so a single add
// or edit does not re-decode the whole library. It runs under settings.mu.
func patchDBRow(hash metainfo.Hash, row []byte) {
	var db *settings.TorrentDB
	if row != nil {
		if err := json.Unmarshal(row, &db); err != nil || db == nil || db.TorrentSpec == nil {
			invalidateDBRows()
			return
		}
	}
	dbRowsMu.Lock()
	defer dbRowsMu.Unlock()
	dbRowsGen++
	if dbRows == nil {
		return
	}
	rows := make([]*settings.TorrentDB, 0, len(dbRows)+1)
	for _, r := range dbRows {
		if r != nil && r.TorrentSpec != nil && r.InfoHash == hash {
			continue
		}
		rows = append(rows, r)
	}
	if db != nil {
		i := sort.Search(len(rows), func(i int) bool { return rows[i].Timestamp <= db.Timestamp })
		rows = slices.Insert(rows, i, db)
	}
	dbRows = rows
}

// listDBRows reads outside dbRowsMu: patchDBRow takes it under settings.mu, and
// ListTorrent takes settings.mu. A change during the read bumps dbRowsGen and the
// stale result is returned but not stored. nil (closed or empty DB) is not cached.
func listDBRows() []*settings.TorrentDB {
	dbRowsMu.Lock()
	rows, gen := dbRows, dbRowsGen
	dbRowsMu.Unlock()
	if rows != nil {
		return rows
	}
	rows = settings.ListTorrent()
	dbRowsMu.Lock()
	if dbRowsGen == gen {
		dbRows = rows
	}
	dbRowsMu.Unlock()
	return rows
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
