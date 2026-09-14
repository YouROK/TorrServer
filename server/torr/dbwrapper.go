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
	t := new(settings.TorrentDB)
	t.TorrentSpec = torr.TorrentSpec
	t.Title = torr.Title
	t.Category = torr.Category
	if torr.Data == "" {
		files := new(tsFiles)
		files.TorrServer.Files = torr.Status().FileStats
		buf, err := json.Marshal(files)
		if err == nil {
			t.Data = string(buf)
			torr.Data = t.Data
		}
	} else {
		t.Data = torr.Data
	}

	if torr.Poster != "" && utils.CheckImgUrl(torr.Poster) {
		t.Poster = torr.Poster
	}
	t.Size = torr.Size
	if t.Size == 0 && torr.Torrent != nil {
		t.Size = torr.Torrent.Length()
	}
	// don't override timestamp from DB on edit
	if existing := GetTorrentDB(torr.Hash()); existing != nil && existing.Timestamp != 0 {
		t.Timestamp = existing.Timestamp
	} else {
		t.Timestamp = torr.Timestamp
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
