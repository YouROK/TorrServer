package utils

import (
	"time"

	"server/settings"
	"server/torr"
	"server/torr/state"

	"github.com/anacrolix/torrent/metainfo"
)

func AddTorrent(torr *torr.Torrent) {
	t := new(settings.TorrentDB)
	t.TorrentSpec = torr.TorrentSpec
	t.Title = torr.Title
	t.Poster = torr.Poster
	t.Timestamp = time.Now().Unix()
	t.Files = torr.Stats().FileStats
	settings.AddTorrent(t)
}

func GetTorrent(hash metainfo.Hash) *torr.Torrent {
	list := settings.ListTorrent()
	for _, db := range list {
		if hash == db.InfoHash {
			torr := new(torr.Torrent)
			torr.TorrentSpec = db.TorrentSpec
			torr.Title = db.Title
			torr.Poster = db.Poster
			torr.Status = state.TorrentInDB
			return torr
		}
	}
	return nil
}

func RemTorrent(hash metainfo.Hash) {
	settings.RemTorrent(hash)
}

func ListTorrents() []*torr.Torrent {
	var ret []*torr.Torrent
	list := settings.ListTorrent()
	for _, db := range list {
		torr := new(torr.Torrent)
		torr.TorrentSpec = db.TorrentSpec
		torr.Title = db.Title
		torr.Poster = db.Poster
		torr.Status = state.TorrentInDB
		ret = append(ret, torr)
	}
	return ret
}
