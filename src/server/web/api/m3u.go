package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	sets "server/settings"
	"server/torr/state"
	"server/utils"
)

func allPlayList(c *gin.Context) {
	_, fromlast := c.GetQuery("fromlast")
	stats := listTorrents()

	host := "http://" + c.Request.Host
	list := "#EXTM3U\n"

	for _, stat := range stats {
		list += getM3uList(stat, host, fromlast)
	}

	c.Header("Content-Type", "audio/x-mpegurl")
	c.Header("Connection", "close")
	c.Header("Content-Disposition", `attachment; filename="all.m3u"`)
	http.ServeContent(c.Writer, c.Request, "all.m3u", time.Now(), bytes.NewReader([]byte(list)))

	c.Status(200)
}

func playList(c *gin.Context) {
	hash := c.Query("torrhash")
	_, fromlast := c.GetQuery("fromlast")
	if hash == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
		return
	}

	stats := listTorrents()
	var stat *state.TorrentStats
	for _, st := range stats {
		if st.Hash == hash {
			stat = st
			break
		}
	}

	if stat == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// TODO проверить
	host := "http://" + c.Request.Host
	list := getM3uList(stat, host, fromlast)
	list = "#EXTM3U\n" + list

	c.Header("Content-Type", "audio/x-mpegurl")
	c.Header("Connection", "close")
	c.Header("Content-Disposition", `attachment; filename="playlist.m3u"`)
	http.ServeContent(c.Writer, c.Request, "playlist.m3u", time.Now(), bytes.NewReader([]byte(list)))

	c.Status(200)
}

func getM3uList(tor *state.TorrentStats, host string, fromLast bool) string {
	m3u := ""
	from := 0
	if fromLast {
		pos := searchLastPlayed(tor)
		if pos != -1 {
			from = pos
		}
	}
	for i, f := range tor.FileStats {
		if i >= from {
			if utils.GetMimeType(f.Path) != "*/*" {
				fn := filepath.Base(f.Path)
				if fn == "" {
					fn = f.Path
				}
				m3u += "#EXTINF:0," + fn + "\n"
				// http://127.0.0.1:8090/stream/fname?link=...&index=0&play
				m3u += host + "/stream/" + url.QueryEscape(f.Path) + "?link=" + tor.Hash + "&file=" + fmt.Sprint(f.Id) + "\n"
			}
		}
	}
	return m3u
}

func searchLastPlayed(tor *state.TorrentStats) int {
	//TODO проверить
	viewed := sets.ListViewed(tor.Hash)
	for i := len(tor.FileStats); i > 0; i-- {
		stat := tor.FileStats[i]
		for _, v := range viewed {
			if stat.Id == v.FileIndex {
				return v.FileIndex
			}
		}
	}
	return -1
}
