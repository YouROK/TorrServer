package torr

import (
	"time"

	"server/settings"
	"server/torr/state"

	"github.com/anacrolix/torrent/metainfo"
)

func AddTorrentDB(torr *Torrent) {
	t := new(settings.TorrentDB)
	t.TorrentSpec = torr.TorrentSpec
	t.Title = torr.Title
	t.Data = torr.Data
	t.Poster = torr.Poster
	t.Size = torr.Size
	if t.Size == 0 && torr.Torrent != nil {
		t.Size = torr.Torrent.Length()
	}
	t.Timestamp = time.Now().Unix()
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
		torr.Timestamp = db.Timestamp
		torr.Size = db.Size
		torr.Data = db.Data
		torr.Stat = state.TorrentInDB
		ret[torr.TorrentSpec.InfoHash] = torr
	}
	return ret
}
