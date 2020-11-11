package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/anacrolix/missinggo/httptoo"
	sets "server/settings"
	"server/torr"
	"server/torr/state"
	"server/utils"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func allPlayList(c *gin.Context) {
	_, fromlast := c.GetQuery("fromlast")
	torrs := torr.ListTorrent()

	host := "http://" + c.Request.Host
	list := "#EXTM3U\n"
	hash := ""
	for _, tr := range torrs {
		list += getM3uList(tr.Status(), host, fromlast)
		hash += tr.Hash().HexString()
	}

	sendM3U(c, "all.m3u", hash, list)
}

func playList(c *gin.Context) {
	hash := c.Query("torrhash")
	_, fromlast := c.GetQuery("fromlast")
	if hash == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
		return
	}

	tors := torr.ListTorrent()
	var tor *torr.Torrent
	for _, tr := range tors {
		if tr.Hash().HexString() == hash {
			tor = tr
			break
		}
	}

	if tor == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// TODO проверить
	host := "http://" + c.Request.Host
	list := getM3uList(tor.Status(), host, fromlast)
	list = "#EXTM3U\n" + list

	sendM3U(c, tor.Name()+".m3u", tor.Hash().HexString(), list)
}

func sendM3U(c *gin.Context, name, hash string, m3u string) {
	c.Header("Content-Type", "audio/x-mpegurl")
	c.Header("Connection", "close")
	if hash != "" {
		c.Header("ETag", httptoo.EncodeQuotedString(fmt.Sprintf("%s/%s", hash, name)))
	}
	if name == "" {
		name = "playlist.m3u"
	}
	c.Header("Content-Disposition", `attachment; filename="`+name+`"`)
	http.ServeContent(c.Writer, c.Request, name, time.Now(), bytes.NewReader([]byte(m3u)))
	c.Status(200)
}

func getM3uList(tor *state.TorrentStatus, host string, fromLast bool) string {
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
				title := filepath.Base(f.Path)
				if tor.Title != "" {
					title = tor.Title
				} else if tor.Name != "" {
					title = tor.Name
				}
				m3u += host + "/stream/" + url.PathEscape(title) + "?link=" + tor.Hash + "&index=" + fmt.Sprint(f.Id) + "&play\n"
			}
		}
	}
	return m3u
}

func searchLastPlayed(tor *state.TorrentStatus) int {
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
