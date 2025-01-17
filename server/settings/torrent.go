package settings

import (
	"encoding/json"
	"server/log"
	"server/settings/sqlite_models"
	"sort"
	"sync"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type TorrentDB struct {
	*torrent.TorrentSpec

	Title    string `json:"title,omitempty"`
	Category string `json:"category,omitempty"`
	Poster   string `json:"poster,omitempty"`
	Data     string `json:"data,omitempty"`

	Timestamp int64 `json:"timestamp,omitempty"`
	Size      int64 `json:"size,omitempty"`
}

type File struct {
	Name string `json:"name,omitempty"`
	Id   int    `json:"id,omitempty"`
	Size int64  `json:"size,omitempty"`
}

var mu sync.Mutex

func AddTorrent(torr *TorrentDB) {
	list := ListTorrent()
	mu.Lock()
	find := -1
	for i, db := range list {
		if db.InfoHash.HexString() == torr.InfoHash.HexString() {
			find = i
			break
		}
	}
	if find != -1 {
		list[find] = torr
	} else {
		list = append(list, torr)
	}
	for _, db := range list {
		buf, err := json.Marshal(db)
		if err == nil {
			tdb.Set("Torrents", db.InfoHash.HexString(), buf)
		}
	}
	mu.Unlock()
}

func ListTorrent() []*TorrentDB {
	mu.Lock()
	defer mu.Unlock()

	var list []*TorrentDB
	keys := tdb.List("Torrents")
	for _, key := range keys {
		buf := tdb.Get("Torrents", key)
		if len(buf) > 0 {
			var torr *TorrentDB
			err := json.Unmarshal(buf, &torr)
			if err == nil {
				list = append(list, torr)
			}
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Timestamp > list[j].Timestamp
	})
	return list
}

func RemTorrent(hash metainfo.Hash) {
	mu.Lock()
	tdb.Rem("Torrents", hash.HexString())
	mu.Unlock()
}

func (t TorrentDB) toSQLTorrent() *sqlite_models.SQLTorrent {
	value, err := json.Marshal(t.Trackers)
	if err != nil {
		return nil
	}

	return &sqlite_models.SQLTorrent{
		Hash:        t.InfoHash.HexString(),
		Category:    t.Category,
		Poster:      t.Poster,
		Size:        t.Size,
		Title:       t.Title,
		DisplayName: t.DisplayName,
		ChunkSize:   t.ChunkSize,
		Trackers:    value,
	}
}

func fromSQLTorrent(sqlTorrent sqlite_models.SQLTorrent) TorrentDB {
	var trackers [][]string
	err := json.Unmarshal(sqlTorrent.Trackers, &trackers)
	if err != nil {
		log.TLogln("cannot unmarshal trackers...")
		return TorrentDB{}
	}

	return TorrentDB{
		Title:     sqlTorrent.Title,
		Category:  sqlTorrent.Category,
		Poster:    sqlTorrent.Poster,
		Timestamp: sqlTorrent.UpdatedAt.Unix(),
		Size:      sqlTorrent.Size,
		TorrentSpec: &torrent.TorrentSpec{
			DisplayName: sqlTorrent.DisplayName,
			ChunkSize:   sqlTorrent.ChunkSize,
			Storage:     nil,
			InfoHash:    metainfo.NewHashFromHex(sqlTorrent.Hash),
			InfoBytes:   nil,
			Trackers:    trackers,
		},
	}
}
