package helpers

import (
	"fmt"
	"net/url"

	"server/settings"
	"server/torr"
	"server/utils"
)

func MakeM3ULists(torrents []*settings.Torrent, host string) string {
	m3u := "#EXTM3U\n"

	for _, t := range torrents {
		m3u += "#EXTINF:0," + t.Name + "\n"
		m3u += host + "/torrent/play?link=" + url.QueryEscape(t.Magnet) + "&m3u=true&fname=" + utils.CleanFName(t.Name+".m3u") + "\n\n"
	}
	return m3u
}

func MakeM3UPlayList(tor torr.TorrentStats, magnet string, host string) string {
	m3u := "#EXTM3U\n"

	for _, f := range tor.FileStats {
		if GetMimeType(f.Path) != "*/*" {
			m3u += "#EXTINF:-1," + f.Path + "\n"
			mag := url.QueryEscape(magnet)
			m3u += host + "/torrent/play?link=" + mag + "&file=" + fmt.Sprint(f.Id) + "\n\n"
		}
	}
	return m3u
}
