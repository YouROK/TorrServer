package helpers

import (
	"fmt"
	"net/url"
	"path/filepath"

	"server/settings"
	"server/torr"
)

func MakeM3ULists(torrents []*settings.Torrent, host string) string {
	m3u := "#EXTM3U\n"

	for _, t := range torrents {
		m3u += "#EXTINF:-1 type=\"playlist\", " + t.Name + "\n"

		magnet := t.Magnet
		mag, _, err := GetMagnet(magnet)
		if err == nil {
			mag.Trackers = []string{} //Remove retrackers for small link size
			magnet = mag.String()
		}
		magnet = url.QueryEscape(magnet)

		m3u += host + "/torrent/play?link=" + url.QueryEscape(magnet) + "&m3u=true\n\n"
	}
	return m3u
}

func MakeM3UPlayList(tor torr.TorrentStats, magnet string, host string) string {
	m3u := "#EXTM3U\n"

	mag, _, err := GetMagnet(magnet)
	if err == nil {
		mag.Trackers = []string{} //Remove retrackers for small link size
		magnet = mag.String()
	}
	magnet = url.QueryEscape(magnet)

	for _, f := range tor.FileStats {
		if GetMimeType(f.Path) != "*/*" {
			fn := filepath.Base(f.Path)
			if fn == "" {
				fn = f.Path
			}
			m3u += "#EXTINF:-1, " + fn + "\n"
			m3u += host + "/torrent/play?link=" + magnet + "&file=" + fmt.Sprint(f.Id) + "\n\n"
		}
	}
	return m3u
}
