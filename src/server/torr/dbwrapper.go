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
	t.Name = torr.Name()
	t.Title = torr.Title
	t.Poster = torr.Poster
	t.Timestamp = time.Now().Unix()
	t.Files = torr.Status().FileStats
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
		torr.Stat = state.TorrentInDB
		ret[torr.TorrentSpec.InfoHash] = torr
	}
	return ret
}
