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

	t.Poster = torr.Poster
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
	t.Size = torr.Size
	if t.Size == 0 && torr.Torrent != nil {
		t.Size = torr.Torrent.Length()
	}
	// don't override timestamp from DB on edit
	t.Timestamp = torr.Timestamp // time.Now().Unix()
	t.OfflineState = torr.OfflineState()
	t.OfflineDir = torr.offlineDirRaw()
	t.OfflineUsedAt = torr.offlineUsedAt()
	if t.OfflineState != "" {
		// A live torrent object does not carry these: keep the stored values. The pin is
		// changed only by SetOfflinePinned, never by saving the torrent.
		if old := GetTorrentDB(torr.Hash()); old != nil {
			t.OfflinePinned = old.OfflinePinned()
			if t.OfflineUsedAt == 0 {
				t.OfflineUsedAt = old.offlineUsedAt()
			}
		}
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
			torr.setOfflineState(db.OfflineState)
			torr.setOfflineDir(db.OfflineDir)
			torr.setOfflineUsedAt(db.OfflineUsedAt)
			torr.setOfflinePinned(db.OfflinePinned)
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
		torr.setOfflineState(db.OfflineState)
		torr.setOfflineDir(db.OfflineDir)
		torr.setOfflineUsedAt(db.OfflineUsedAt)
		torr.setOfflinePinned(db.OfflinePinned)
		torr.Stat = state.TorrentInDB
		ret[torr.TorrentSpec.InfoHash] = torr
	}
	return ret
}
