package torr

import (
	"encoding/json"

	"server/settings"
	"server/torr/state"
	"server/torr/utils"

	"github.com/anacrolix/torrent/metainfo"
)

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
	list := settings.ListTorrent()
	for _, db := range list {
		if hash == db.InfoHash {
			torr := new(Torrent)
			torr.TorrentSpec = db.TorrentSpec
			torr.Title = db.Title
			torr.Poster = db.Poster
			torr.Category = db.Category
			torr.Timestamp = db.Timestamp
			torr.Size = db.Size
			torr.Data = db.Data
			torr.Stat = state.TorrentInDB
			return torr
		}
	}
	return nil
}

func RemTorrentDB(hash metainfo.Hash) {
	settings.RemTorrent(hash)
}

func ListTorrentsDB() map[metainfo.Hash]*Torrent {
	ret := make(map[metainfo.Hash]*Torrent)
	list := settings.ListTorrent()
	for _, db := range list {
		torr := new(Torrent)
		torr.TorrentSpec = db.TorrentSpec
		torr.Title = db.Title
		torr.Poster = db.Poster
		torr.Category = db.Category
		torr.Timestamp = db.Timestamp
		torr.Size = db.Size
		torr.Data = db.Data
		torr.Stat = state.TorrentInDB
		ret[torr.TorrentSpec.InfoHash] = torr
	}
	return ret
}
