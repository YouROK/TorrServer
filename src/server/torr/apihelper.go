package torr

import (
	"errors"
	"sort"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"server/log"
)

var (
	bts *BTServer
)

func InitApiHelper(bt *BTServer) {
	bts = bt
}

func AddTorrent(spec *torrent.TorrentSpec, title, poster string) (*Torrent, error) {
	torr, err := NewTorrent(spec, bts)
	if err != nil {
		log.TLogln("error add torrent:", err)
		return nil, err
	}

	if !torr.GotInfo() {
		log.TLogln("error add torrent:", "timeout connection torrent")
		return nil, errors.New("timeout connection torrent")
	}

	torr.Title = title
	torr.Poster = poster

	if torr.Title == "" {
		torr.Title = torr.Name()
	}

	return torr, nil
}

func SaveTorrentToDB(torr *Torrent) {
	log.TLogln("save to db:", torr.Hash())
	AddTorrentDB(torr)
}

func GetTorrent(hashHex string) *Torrent {
	hash := metainfo.NewHashFromHex(hashHex)
	tor := bts.GetTorrent(hash)
	if tor == nil {
		tor = GetTorrentDB(hash)
	}

	return tor
}

func RemTorrent(hashHex string) {
	hash := metainfo.NewHashFromHex(hashHex)
	bts.RemoveTorrent(hash)
	RemTorrentDB(hash)
}

func ListTorrent() []*Torrent {
	btlist := bts.ListTorrents()
	dblist := ListTorrentsDB()

	for hash, t := range dblist {
		if _, ok := btlist[hash]; !ok {
			btlist[hash] = t
		}
	}
	var ret []*Torrent

	for _, t := range btlist {
		ret = append(ret, t)
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Timestamp > ret[j].Timestamp
	})

	return ret
}

func DropTorrent(hashHex string) {
	hash := metainfo.NewHashFromHex(hashHex)
	bts.RemoveTorrent(hash)
}
