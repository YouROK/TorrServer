package dlna

import (
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/anacrolix/dms/dlna"
	"github.com/anacrolix/dms/upnpav"

	"server/log"
	mt "server/mimetype"
	"server/settings"
	"server/torr"
	"server/torr/state"
)

func getRoot() (ret []interface{}) {
	// Torrents Object
	tObj := upnpav.Object{
		ID:         "%2FTR",
		ParentID:   "0",
		Restricted: 1,
		Title:      "Torrents",
		Class:      "object.container.storageFolder",
		Date:       upnpav.Timestamp{Time: time.Now()},
	}

	// add Torrents Object
	vol := len(torr.ListTorrent())
	cnt := upnpav.Container{Object: tObj, ChildCount: vol}
	ret = append(ret, cnt)

	return
}

func getTorrents() (ret []interface{}) {
	torrs := torr.ListTorrent()
	// sort by title as in cds SortCaps
	sort.Slice(torrs, func(i, j int) bool {
		return torrs[i].Title < torrs[j].Title
	})

	vol := 0
	for _, t := range torrs {
		vol++
		obj := upnpav.Object{
			ID:          "%2F" + t.TorrentSpec.InfoHash.HexString(),
			ParentID:    "%2FTR",
			Restricted:  1,
			Title:       strings.ReplaceAll(t.Title, "/", "|"),
			Class:       "object.container.storageFolder",
			Icon:        t.Poster,
			AlbumArtURI: t.Poster,
			Date:        upnpav.Timestamp{Time: time.Now()},
		}
		cnt := upnpav.Container{Object: obj, ChildCount: 1}
		ret = append(ret, cnt)
	}
	if vol == 0 {
		obj := upnpav.Object{
			ID:         "%2FNT",
			ParentID:   "%2FTR",
			Restricted: 1,
			Title:      "No Torrents",
			Class:      "object.container.storageFolder",
			Date:       upnpav.Timestamp{Time: time.Now()},
		}
		cnt := upnpav.Container{Object: obj, ChildCount: 0}
		ret = append(ret, cnt)
	}
	return
}

func getTorrent(path, host string) (ret []interface{}) {
	// find torrent without load
	torrs := torr.ListTorrent()
	var torr *torr.Torrent
	for _, t := range torrs {
		if strings.Contains(path, t.TorrentSpec.InfoHash.HexString()) {
			torr = t
			break
		}
	}
	if torr == nil {
		return nil
	}

	// get content from torrent
	parent := "%2F" + torr.TorrentSpec.InfoHash.HexString()
	// if torrent not loaded, get button for load
	if torr.Files() == nil {
		obj := upnpav.Object{
			ID:         parent + "%2FLD",
			ParentID:   parent,
			Restricted: 1,
			Title:      "Load Torrent",
			Class:      "object.container.storageFolder",
			Date:       upnpav.Timestamp{Time: time.Now()},
		}
		cnt := upnpav.Container{Object: obj, ChildCount: 1}
		ret = append(ret, cnt)
		return
	}

	ret = loadTorrent(path, host)
	return
}

func getTorrentMeta(path, host string) (ret interface{}) {
	// Meta object
	if path == "/" {
		// root object meta
		rootObj := upnpav.Object{
			ID:         "0",
			ParentID:   "-1",
			Restricted: 1,
			Searchable: 1,
			Title:      "TorrServer",
			Date:       upnpav.Timestamp{Time: time.Now()},
			Class:      "object.container.storageFolder",
		}
		meta := upnpav.Container{Object: rootObj, ChildCount: 1}
		return meta
	} else if filepath.Base(path) == "TR" {
		// TR Object Meta
		trObj := upnpav.Object{
			ID:         "%2FTR",
			ParentID:   "0",
			Restricted: 1,
			Searchable: 1,
			Title:      "Torrents",
			Date:       upnpav.Timestamp{Time: time.Now()},
			Class:      "object.container.storageFolder",
		}
		torrs := torr.ListTorrent()
		vol := len(torrs)
		meta := upnpav.Container{Object: trObj, ChildCount: vol}
		return meta
	} else if isHashPath(path) {
		// find torrent without load
		torrs := torr.ListTorrent()
		var torr *torr.Torrent
		for _, t := range torrs {
			if strings.Contains(path, t.TorrentSpec.InfoHash.HexString()) {
				torr = t
				break
			}
		}
		if torr == nil {
			return nil
		}
		// hash object meta
		obj := upnpav.Object{
			ID:         "%2F" + torr.TorrentSpec.InfoHash.HexString(),
			ParentID:   "%2FTR",
			Restricted: 1,
			Title:      torr.Title,
			Date:       upnpav.Timestamp{Time: time.Now()},
		}
		meta := upnpav.Container{Object: obj, ChildCount: 1}
		return meta
	} else if filepath.Base(path) == "LD" {
		parent := url.PathEscape(filepath.Dir(path))
		// LD object meta
		obj := upnpav.Object{
			ID:         parent + "%2FLD",
			ParentID:   parent,
			Restricted: 1,
			Searchable: 1,
			Title:      "Load Torrents",
			Date:       upnpav.Timestamp{Time: time.Now()},
		}
		meta := upnpav.Container{Object: obj, ChildCount: 1}
		return meta
	} else {
		file := filepath.Base(path)
		id := url.PathEscape(path)
		parent := url.PathEscape(filepath.Dir(path))
		// file object meta
		obj := upnpav.Object{
			ID:         id,
			ParentID:   parent,
			Restricted: 1,
			Searchable: 1,
			Title:      file,
			Date:       upnpav.Timestamp{Time: time.Now()},
		}
		meta := upnpav.Container{Object: obj, ChildCount: 1}
		return meta
	}
}

func loadTorrent(path, host string) (ret []interface{}) {
	hash := filepath.Base(filepath.Dir(path))
	if hash == "/" || hash == "\\" {
		hash = filepath.Base(path)
	}
	if len(hash) != 40 {
		return
	}

	tor := torr.GetTorrent(hash)
	if tor == nil {
		log.TLogln("Dlna error get info from torrent", hash)
		return
	}
	if len(tor.Files()) == 0 {
		time.Sleep(time.Millisecond * 200)
		timeout := time.Now().Add(time.Second * 60)
		for {
			tor = torr.GetTorrent(hash)
			if len(tor.Files()) > 0 {
				break
			}
			time.Sleep(time.Millisecond * 200)
			if time.Now().After(timeout) {
				return
			}
		}
	}
	parent := "%2F" + tor.TorrentSpec.InfoHash.HexString()
	files := tor.Status().FileStats
	for _, f := range files {
		obj := getObjFromTorrent(path, parent, host, tor, f)
		if obj != nil {
			ret = append(ret, obj)
		}
	}
	return
}

func getLink(host, path string) string {
	if !strings.HasPrefix(host, "http") {
		// if settings.Ssl {
		// 	host = "https://" + host
		// } else {
		host = "http://" + host
		// }
	}
	pos := strings.LastIndex(host, ":")
	if pos > 7 {
		host = host[:pos]
	}
	// if settings.Ssl {
	// 	return host + ":" + settings.SslPort + "/" + path
	// } else {
	return host + ":" + settings.Port + "/" + path
	// }
}

func getObjFromTorrent(path, parent, host string, torr *torr.Torrent, file *state.TorrentFileStat) (ret interface{}) {
	mime, err := mt.MimeTypeByPath(file.Path)
	if err != nil {
		if settings.BTsets.EnableDebug {
			log.TLogln("Can't detect mime type", err)
		}
		return
	}
	// TODO: handle subtitles for media
	if !mime.IsMedia() {
		return
	}
	if settings.BTsets.EnableDebug {
		log.TLogln("mime type", mime.String(), file.Path)
	}

	obj := upnpav.Object{
		ID:         parent + "%2F" + url.PathEscape(file.Path),
		ParentID:   parent,
		Restricted: 1,
		Title:      file.Path,
		Class:      "object.item." + mime.Type() + "Item",
		Date:       upnpav.Timestamp{Time: time.Now()},
	}

	item := upnpav.Item{
		Object: obj,
		Res:    make([]upnpav.Resource, 0, 1),
	}
	pathPlay := "stream/" + url.PathEscape(file.Path) + "?link=" + torr.TorrentSpec.InfoHash.HexString() + "&play&index=" + strconv.Itoa(file.Id)
	item.Res = append(item.Res, upnpav.Resource{
		URL: getLink(host, pathPlay),
		ProtocolInfo: fmt.Sprintf("http-get:*:%s:%s", mime, dlna.ContentFeatures{
			SupportRange:    true,
			SupportTimeSeek: true,
		}.String()),
		Size: uint64(file.Length),
	})
	return item
}
