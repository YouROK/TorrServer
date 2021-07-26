package dlna

import (
	"github.com/anacrolix/dms/upnpav"
	"server/torr"
)

func getTorrentRoot() (ret []interface{}) {
	obj := upnpav.Object{
		ID:         "0",
		Restricted: 1,
		ParentID:   "-1",
		Class:      "object.container.storageFolder",
		Title:      "/",
	}
	cnt := upnpav.Container{Object: obj}
	ret = append(ret, cnt)
	return
}

func getTorrents() (ret []interface{}) {
	torrs := torr.ListTorrent()
	for _, t := range torrs {
		obj := upnpav.Object{
			ID:         "%2F" + t.TorrentSpec.InfoHash.HexString(),
			Restricted: 1,
			ParentID:   "0",
			Class:      "object.container.storageFolder",
			Title:      t.Title,
			//Icon:       t.Poster,
		}
		cnt := upnpav.Container{Object: obj}
		ret = append(ret, cnt)
	}
	return
}
