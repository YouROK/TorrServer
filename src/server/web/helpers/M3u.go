package helpers

import (
	"fmt"
	"net/url"
	"path/filepath"

	"server/settings"
	"server/torr"
	"server/utils"
)

func MakeM3ULists(torrents []*settings.Torrent, host string) string {
	m3u := "#EXTM3U\n"

	for _, t := range torrents {
		m3u += "#EXTINF:0 type=\"playlist\"," + t.Name + "\n"

		magnet := t.Magnet
		mag, _, err := GetMagnet(magnet)
		if err == nil {
			mag.Trackers = []string{} // remove retrackers for small link size
			mag.DisplayName = "" // clear dn from link - long query params may fail in QueryParam("link")
			magnet = mag.String()
		}
		m3u += host + "/torrent/play?link=" + url.QueryEscape(magnet) + "&m3u=true\n"
	}
	return m3u
}

func MakeM3UPlayList(tor torr.TorrentStats, magnet string, host string) string {
	m3u := "#EXTM3U\n"

	mag, _, err := GetMagnet(magnet)
	if err == nil {
		mag.Trackers = []string{} //Remove retrackers for small link size
		mag.DisplayName = "" //Remove dn from link (useless)
		magnet = mag.String()
	}
	magnet = url.QueryEscape(magnet)

	for _, f := range tor.FileStats {
		if GetMimeType(f.Path) != "*/*" {
			fn := filepath.Base(f.Path)
			if fn == "" {
				fn = f.Path
			}
			m3u += "#EXTINF:0," + fn + "\n"
			m3u += host + "/torrent/play/" + url.QueryEscape(utils.CleanFName(f.Path)) + "?link=" + magnet + "&file=" + fmt.Sprint(f.Id) + "\n"
		}
	}
	return m3u
}
