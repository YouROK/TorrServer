package torr

import (
	"io"
	"os"
	"sort"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/log"
	sets "server/settings"
)

var (
	bts *BTServer
)

func InitApiHelper(bt *BTServer) {
	bts = bt
}

func LoadTorrent(tor *Torrent) *Torrent {
	if tor.TorrentSpec == nil {
		return nil
	}
	tr, err := NewTorrent(tor.TorrentSpec, bts)
	if err != nil {
		return nil
	}
	if !tr.WaitInfo() {
		return nil
	}
	tr.Title = tor.Title
	tr.Poster = tor.Poster
	tr.Data = tor.Data
	return tr
}

func AddTorrent(spec *torrent.TorrentSpec, title, poster string, data string) (*Torrent, error) {
	torr, err := NewTorrent(spec, bts)
	if err != nil {
		log.TLogln("error add torrent:", err)
		return nil, err
	}

	torDB := GetTorrentDB(spec.InfoHash)

	if torr.Title == "" {
		torr.Title = title
		if title == "" && torDB != nil {
			torr.Title = torDB.Title
		}
	}
	if torr.Poster == "" {
		torr.Poster = poster
		if torr.Poster == "" && torDB != nil {
			torr.Poster = torDB.Poster
		}
	}
	if torr.Data == "" {
		torr.Data = data
		if torr.Data == "" && torDB != nil {
			torr.Data = torDB.Data
		}
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
	if tor != nil {
		tor.AddExpiredTime(time.Minute)
		return tor
	}

	tr := GetTorrentDB(hash)
	if tr != nil {
		tor = tr
		go func() {
			tr, _ := NewTorrent(tor.TorrentSpec, bts)
			if tr != nil {
				tr.Title = tor.Title
				tr.Poster = tor.Poster
				tr.Data = tor.Data
				tr.Size = tor.Size
				tr.Timestamp = tor.Timestamp
				tr.GotInfo()
			}
		}()
	}
	return tor
}

func SetTorrent(hashHex, title, poster, data string) *Torrent {
	hash := metainfo.NewHashFromHex(hashHex)
	tor := bts.GetTorrent(hash)
	if tor != nil {
		tor.Title = title
		tor.Poster = poster
		tor.Data = data
	}

	tor = GetTorrentDB(hash)
	if tor != nil {
		tor.Title = title
		tor.Poster = poster
		tor.Data = data
		AddTorrentDB(tor)
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
		if ret[i].Timestamp != ret[j].Timestamp {
			return ret[i].Timestamp > ret[j].Timestamp
		} else {
			return ret[i].Title > ret[j].Title
		}
	})

	return ret
}

func DropTorrent(hashHex string) {
	hash := metainfo.NewHashFromHex(hashHex)
	bts.RemoveTorrent(hash)
}

func SetSettings(set *sets.BTSets) {
	if sets.ReadOnly {
		return
	}
	bts.Disconnect()
	sets.SetBTSets(set)
	bts.Connect()
}

func SetDefSettings() {
	if sets.ReadOnly {
		return
	}
	bts.Disconnect()
	sets.SetDefault()
	bts.Connect()
}

func Shutdown() {
	bts.Disconnect()
	sets.CloseDB()
	os.Exit(0)
}

func WriteStatus(w io.Writer) {
	bts.client.WriteStatus(w)
}

func Preload(torr *Torrent, index int) {
	if !sets.BTsets.PreloadBuffer {
		size := int64(32 * 1024 * 1024)
		if size > sets.BTsets.CacheSize {
			size = sets.BTsets.CacheSize
		}
		torr.Preload(index, size)
	} else {
		size := int64(float32(sets.BTsets.ReaderReadAHead) / 100.0 * float32(sets.BTsets.CacheSize))
		torr.Preload(index, size)
	}
}
