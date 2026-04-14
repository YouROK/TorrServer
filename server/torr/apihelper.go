package torr

import (
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/log"
	sets "server/settings"
)

var bts *BTServer

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
	tr.Users = tor.Users
	return tr
}

func AddTorrent(spec *torrent.TorrentSpec, title, poster string, data string, category string) (*Torrent, error) {
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
		if torr.Title == "" && torr.Torrent != nil && torr.Torrent.Info() != nil {
			torr.Title = torr.Info().Name
		}
	}

	if torr.Category == "" {
		torr.Category = category
		if torr.Category == "" && torDB != nil {
			torr.Category = torDB.Category
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

	if torDB != nil {
		torr.Users = append(torr.Users, torDB.Users...)
	}

	return torr, nil
}

func ensureTorrentUsers(torr *Torrent) bool {
	if !sets.PerUserData {
		return false
	}
	if torr == nil {
		return false
	}

	available := sets.ListUsers()
	availableSet := make(map[string]struct{}, len(available))
	for _, usr := range available {
		availableSet[usr] = struct{}{}
	}

	filtered := make([]string, 0, len(torr.Users))
	for _, usr := range torr.Users {
		if len(availableSet) > 0 {
			if _, ok := availableSet[usr]; !ok {
				continue
			}
		}
		filtered = append(filtered, usr)
	}

	if len(filtered) == 0 && len(available) > 0 {
		filtered = append(filtered, available...)
	}

	if slices.Equal(torr.Users, filtered) {
		return false
	}
	torr.Users = filtered
	return true
}

func addUserToTorrent(torr *Torrent, user string) bool {
	if !sets.PerUserData {
		return false
	}
	if torr == nil || user == "" {
		return false
	}
	if slices.Contains(torr.Users, user) {
		return false
	}
	torr.Users = append(torr.Users, user)
	return true
}

func removeUserFromTorrent(torr *Torrent, user string) bool {
	if !sets.PerUserData {
		return false
	}
	if torr == nil || user == "" {
		return false
	}
	idx := -1
	for i, u := range torr.Users {
		if u == user {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false
	}
	torr.Users = append(torr.Users[:idx], torr.Users[idx+1:]...)
	return true
}

func AddTorrentForUser(spec *torrent.TorrentSpec, title, poster, data, category, user string) (*Torrent, error) {
	torr, err := AddTorrent(spec, title, poster, data, category)
	if err != nil {
		return nil, err
	}
	addUserToTorrent(torr, user)
	SetTorrentUsersDB(torr.Hash(), torr.Users)
	return torr, nil
}

func SaveTorrentToDB(torr *Torrent) {
	log.TLogln("save to db:", torr.Hash())
	AddTorrentDB(torr)
}

func GetTorrent(hashHex string) *Torrent {
	hash := metainfo.NewHashFromHex(hashHex)
	timeout := time.Second * time.Duration(sets.BTsets.TorrentDisconnectTimeout)
	if timeout > time.Minute {
		timeout = time.Minute
	}
	tor := bts.GetTorrent(hash)
	if tor != nil {
		tor.AddExpiredTime(timeout)
		return tor
	}

	tr := GetTorrentDB(hash)
	if tr != nil {
		tor = tr
		go func() {
			log.TLogln("New torrent", tor.Hash())
			tr, _ := NewTorrent(tor.TorrentSpec, bts)
			if tr != nil {
				tr.Title = tor.Title
				tr.Poster = tor.Poster
				tr.Data = tor.Data
				tr.Size = tor.Size
				tr.Timestamp = tor.Timestamp
				tr.Category = tor.Category
				tr.Users = tor.Users
				tr.GotInfo()
			}
		}()
	}
	return tor
}

func SetTorrent(hashHex, title, poster, category string, data string) *Torrent {
	hash := metainfo.NewHashFromHex(hashHex)
	torr := bts.GetTorrent(hash)
	torrDb := GetTorrentDB(hash)

	if title == "" && torr == nil && torrDb != nil {
		torr = GetTorrent(hashHex)
		torr.GotInfo()
		if torr.Torrent != nil && torr.Torrent.Info() != nil {
			title = torr.Info().Name
		}
	}

	if torr != nil {
		if title == "" && torr.Torrent != nil && torr.Torrent.Info() != nil {
			title = torr.Info().Name
		}
		torr.Title = title
		torr.Poster = poster
		torr.Category = category
		if data != "" {
			torr.Data = data
		}
	}
	// update torrent data in DB
	if torrDb != nil {
		torrDb.Title = title
		torrDb.Poster = poster
		torrDb.Category = category
		if data != "" {
			torrDb.Data = data
		}
		AddTorrentDB(torrDb)
	}
	if torr != nil {
		return torr
	} else {
		return torrDb
	}
}

func RemTorrent(hashHex string) {
	if sets.ReadOnly {
		log.TLogln("API RemTorrent: Read-only DB mode!", hashHex)
		return
	}
	hash := metainfo.NewHashFromHex(hashHex)
	if bts.RemoveTorrent(hash) {
		if sets.BTsets.UseDisk && hashHex != "" && hashHex != "/" {
			name := filepath.Join(sets.BTsets.TorrentsSavePath, hashHex)
			ff, _ := os.ReadDir(name)
			for _, f := range ff {
				os.Remove(filepath.Join(name, f.Name()))
			}
			err := os.Remove(name)
			if err != nil {
				log.TLogln("Error remove cache:", err)
			}
		}
	}
	RemTorrentDB(hash)
}

func RemTorrentForUser(hashHex, user string) {
	if !sets.PerUserData || user == "" {
		RemTorrent(hashHex)
		return
	}

	hash := metainfo.NewHashFromHex(hashHex)
	torr := bts.GetTorrent(hash)
	if torr == nil {
		torr = GetTorrentDB(hash)
	}
	if torr == nil {
		return
	}

	ensureTorrentUsers(torr)
	if len(torr.Users) > 1 {
		removeUserFromTorrent(torr, user)
	} else {
		RemTorrent(hashHex)
		return
	}

	SetTorrentUsersDB(hash, torr.Users)
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

func ListTorrentForUser(user string) []*Torrent {
	list := ListTorrent()
	if !sets.PerUserData {
		return list
	}
	if user == "" {
		return list
	}

	var ret []*Torrent
	for _, tor := range list {
		if slices.Contains(tor.Users, user) {
			ret = append(ret, tor)
		}
	}

	return ret
}

func DropTorrent(hashHex string) {
	hash := metainfo.NewHashFromHex(hashHex)
	bts.RemoveTorrent(hash)
}

func SetSettings(set *sets.BTSets) {
	if sets.ReadOnly {
		log.TLogln("API SetSettings: Read-only DB mode!")
		return
	}
	sets.SetBTSets(set)
	log.TLogln("drop all torrents")
	dropAllTorrent()
	time.Sleep(time.Second * 1)
	log.TLogln("disconect")
	bts.Disconnect()
	log.TLogln("connect")
	bts.Connect()
	time.Sleep(time.Second * 1)
	log.TLogln("end set settings")
}

func SetDefSettings() {
	if sets.ReadOnly {
		log.TLogln("API SetDefSettings: Read-only DB mode!")
		return
	}
	sets.SetDefaultConfig()
	log.TLogln("drop all torrents")
	dropAllTorrent()
	time.Sleep(time.Second * 1)
	log.TLogln("disconect")
	bts.Disconnect()
	log.TLogln("connect")
	bts.Connect()
	time.Sleep(time.Second * 1)
	log.TLogln("end set default settings")
}

func dropAllTorrent() {
	for _, torr := range bts.torrents {
		torr.drop()
		<-torr.closed
	}
}

func Shutdown() {
	bts.Disconnect()
	sets.CloseDB()
	log.TLogln("Received shutdown. Quit")
	os.Exit(0)
}

func WriteStatus(w io.Writer) {
	bts.client.WriteStatus(w)
}

func Preload(torr *Torrent, index int) {
	cache := float32(sets.BTsets.CacheSize)
	preload := float32(sets.BTsets.PreloadCache)
	size := int64((cache / 100.0) * preload)
	if size <= 0 {
		return
	}
	if size > sets.BTsets.CacheSize {
		size = sets.BTsets.CacheSize
	}
	torr.Preload(index, size)
}
