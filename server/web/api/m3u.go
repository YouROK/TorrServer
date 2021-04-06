package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
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
	torrs := torr.ListTorrent()

	host := "http://" + c.Request.Host
	list := "#EXTM3U\n"
	hash := ""
	// fn=file.m3u fix forkplayer bug with end .m3u in link
	for _, tr := range torrs {
		list += "#EXTINF:0 type=\"playlist\"," + tr.Title + "\n"
		list += host + "/stream/" + url.PathEscape(tr.Title) + ".m3u?link=" + tr.TorrentSpec.InfoHash.HexString() + "&m3u&fn=file.m3u\n"
		hash += tr.Hash().HexString()
	}

	sendM3U(c, "all.m3u", hash, list)
}

// http://127.0.0.1:8090/playlist?hash=...
// http://127.0.0.1:8090/playlist?hash=...&fromlast
func playList(c *gin.Context) {
	hash, _ := c.GetQuery("hash")
	_, fromlast := c.GetQuery("fromlast")
	if hash == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
		return
	}

	tor := torr.GetTorrent(hash)
	if tor == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if tor.Stat == state.TorrentInDB {
		tor = torr.LoadTorrent(tor)
		if tor == nil {
			c.AbortWithError(http.StatusInternalServerError, errors.New("error get torrent info"))
			return
		}
	}

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
				fileNamesakes := findFileNamesakes(tor.FileStats, f) //find external media with same name (audio/subtiles tracks)
				if fileNamesakes != nil {
					m3u += "#EXTVLCOPT:input-slave="         //include VLC option for external media
					for _, namesake := range fileNamesakes { //include play-links to external media, with # splitter
						sname := filepath.Base(namesake.Path)
						m3u += host + "/stream/" + url.PathEscape(sname) + "?link=" + tor.Hash + "&index=" + fmt.Sprint(namesake.Id) + "&play#"
					}
					m3u += "\n"
				}
				name := filepath.Base(f.Path)
				m3u += host + "/stream/" + url.PathEscape(name) + "?link=" + tor.Hash + "&index=" + fmt.Sprint(f.Id) + "&play\n"
			}
		}
	}
	return m3u
}

func findFileNamesakes(files []*state.TorrentFileStat, file *state.TorrentFileStat) []*state.TorrentFileStat {
	//find files with the same name in torrent
	name := filepath.Base(strings.TrimSuffix(file.Path, filepath.Ext(file.Path)))
	var namesakes []*state.TorrentFileStat
	for _, f := range files {
		if strings.Contains(f.Path, name) { //external tracks always include name of videofile
			if f != file { //exclude itself
				namesakes = append(namesakes, f)
			}
		}
	}
	return namesakes
}

func searchLastPlayed(tor *state.TorrentStatus) int {
	viewed := sets.ListViewed(tor.Hash)
	if len(viewed) == 0 {
		return -1
	}
	sort.Slice(viewed, func(i, j int) bool {
		return viewed[i].FileIndex > viewed[j].FileIndex
	})

	lastViewedIndex := viewed[0].FileIndex

	for i, stat := range tor.FileStats {
		if stat.Id == lastViewedIndex {
			if i >= len(tor.FileStats) {
				return -1
			}
			return i
		}
	}

	return -1
}
