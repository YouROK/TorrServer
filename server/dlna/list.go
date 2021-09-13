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

	var vol = 0
	for _, t := range torrs {
		vol++
		obj := upnpav.Object{
			ID:          "%2F" + t.TorrentSpec.InfoHash.HexString(),
			ParentID:    "%2FTR",
			Restricted:  1,
			Title:       t.Title,
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
		cnt := upnpav.Container{Object: obj, ChildCount: 1}
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
	// https://github.com/1100101/minidlna/blob/ca6dbba18390ad6f8b8d7b7dbcf797dbfd95e2db/upnpsoap.c#L1237-L1243
	if path == "/" {
		rootObj := upnpav.Object{
			ID:         "0",
			ParentID:   "-1",
			Restricted: 1,
			Searchable: 1,
			Title:      "TorrServer",
			Date:       upnpav.Timestamp{Time: time.Now()},
			Class:      "object.container",
			SearchXML: `	<upnp:searchClass includeDerived="0">object.container.album.musicAlbum</upnp:searchClass>
	<upnp:searchClass includeDerived="0">object.container.genre.musicGenre</upnp:searchClass>
	<upnp:searchClass includeDerived="0">object.container.person.musicArtist</upnp:searchClass>
	<upnp:searchClass includeDerived="0">object.container.playlistContainer</upnp:searchClass>
	<upnp:searchClass includeDerived="0">object.container.storageFolder</upnp:searchClass>
	<upnp:searchClass includeDerived="0">object.item.audioItem.musicTrack</upnp:searchClass>
	<upnp:searchClass includeDerived="0">object.item.imageItem.photo</upnp:searchClass>
	<upnp:searchClass includeDerived="1">object.container.album</upnp:searchClass>
	<upnp:searchClass includeDerived="1">object.container.genre</upnp:searchClass>
	<upnp:searchClass includeDerived="1">object.container</upnp:searchClass>
	<upnp:searchClass includeDerived="1">object.item.audioItem</upnp:searchClass>
	<upnp:searchClass includeDerived="1">object.item.imageItem</upnp:searchClass>
	<upnp:searchClass includeDerived="1">object.item.videoItem</upnp:searchClass>
`,
		}
		// add Root Object
		meta := upnpav.Container{Object: rootObj}
		return meta
	} else if path == "/TR" {
		// TR Object Meta
		trObj := upnpav.Object{
			ID:         "%2FTR",
			ParentID:   "0",
			Restricted: 1,
			Searchable: 1,
			Title:      "Torrents",
			Date:       upnpav.Timestamp{Time: time.Now()},
			Class:      "object.container",
		}
		//vol := len(torr.ListTorrent())
		meta := upnpav.Container{Object: trObj}
		return meta
	}

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

	// Meta object
	if isHashPath(path) {
		// hash object meta
		obj := upnpav.Object{
			ID:         "%2F" + torr.TorrentSpec.InfoHash.HexString(),
			ParentID:   "%2FTR",
			Restricted: 1,
			Title:      torr.Title,
			Date:       upnpav.Timestamp{Time: time.Now()},
			Class:      "object.container",
		}
		meta := upnpav.Container{Object: obj}
		return meta
	} else if filepath.Base(path) == "LD" {
		parent := url.PathEscape(filepath.Dir(path))
		if settings.BTsets.EnableDebug {
			log.TLogln("getTorrentMeta parent for LD", parent)
		}
		// LD object meta
		obj := upnpav.Object{
			ID:         parent + "%2FLD",
			ParentID:   parent,
			Restricted: 1,
			Searchable: 1,
			Title:      "Load Torrents",
			Date:       upnpav.Timestamp{Time: time.Now()},
			Class:      "object.container",
		}
		meta := upnpav.Container{Object: obj}
		return meta
	} else {
		file := filepath.Base(path)
		id := url.PathEscape(path)
		parent := url.PathEscape(filepath.Dir(path))
		if settings.BTsets.EnableDebug {
			log.TLogln("getTorrentMeta id:", id, "parent:", parent)
		}
		// file object meta
		obj := upnpav.Object{
			ID:         id,
			ParentID:   parent,
			Restricted: 1,
			Searchable: 1,
			Title:      file,
			Date:       upnpav.Timestamp{Time: time.Now()},
			Class:      "object.container",
		}
		meta := upnpav.Container{Object: obj}
		return meta
	}
	// for error response
	return nil
}

func loadTorrent(path, host string) (ret []interface{}) {
	hash := filepath.Base(filepath.Dir(path))
	if hash == "/" {
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
		host = "http://" + host
	}
	pos := strings.LastIndex(host, ":")
	if pos > 7 {
		host = host[:pos]
	}
	return host + ":" + settings.Port + "/" + path
}

func getObjFromTorrent(path, parent, host string, torr *torr.Torrent, file *state.TorrentFileStat) (ret interface{}) {

	mime, err := MimeTypeByPath(file.Path)
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
