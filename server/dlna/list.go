package dlna

import (
	"fmt"
	"net/url"
	"path/filepath"
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

func getTorrents() (ret []interface{}) {
	torrs := torr.ListTorrent()
	for _, t := range torrs {
		obj := upnpav.Object{
			ID:          "%2F" + t.TorrentSpec.InfoHash.HexString(),
			ParentID:    "0",
			Title:       t.Title,
			Class:       "object.container.storageFolder",
			Restricted:  1,
			Icon:        t.Poster,
			AlbumArtURI: t.Poster,
			Date: upnpav.Timestamp{Time: time.Now()},
		}
		cnt := upnpav.Container{Object: obj}
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
			ID:         parent + "%2FLoad Torrent",
			ParentID:   parent,
			Title:      "Load Torrent",
			Class:      "object.container.storageFolder",
			Restricted: 1,
			Date: upnpav.Timestamp{Time: time.Now()},
		}
		cnt := upnpav.Container{Object: obj}
		ret = append(ret, cnt)
		return
	}

	ret = loadTorrent(path, host)
	return
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
		log.TLogln("Can't detect mime type", err)	
		return
	}
	if !mime.IsMedia() {
		return
	}

	obj := upnpav.Object{
		ID:         parent + "%2F" + file.Path,
		ParentID:   parent,
		Title:      file.Path,
		Class:      "object.item." + mime.Type() + "Item",
		Restricted: 1,
		Date: upnpav.Timestamp{Time: time.Now()},
	}

	item := upnpav.Item{
		Object: obj,
		Res:    make([]upnpav.Resource, 0, 1),
	}
	pathPlay := "stream/" + url.PathEscape(file.Path) + "?link=" + torr.TorrentSpec.InfoHash.HexString() + "&play&index=" + strconv.Itoa(file.Id)
	item.Res = append(item.Res, upnpav.Resource{
		URL: getLink(host, pathPlay),
		ProtocolInfo: fmt.Sprintf("http-get:*:%s:%s", mime, dlna.ContentFeatures{
			SupportRange: true,
		}.String()),
		Size: uint64(file.Length),
	})
	return item
}
