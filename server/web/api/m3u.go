package api

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/anacrolix/missinggo/v2/httptoo"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"

	sets "server/settings"
	"server/torr"
	"server/torr/state"
	"server/utils"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// allPlayList godoc
//
//	@Summary		Get a M3U playlist with all torrents
//	@Description	Retrieve all torrents and generates a bundled M3U playlist.
//
//	@Tags			API
//
//	@Produce		audio/x-mpegurl
//	@Security		BasicAuth
//	@Success		200	{file}	file
//	@Router			/playlistall/all.m3u [get]
func allPlayList(c *gin.Context) {
	torrs := torr.ListTorrent()

	category := c.Query("category")
	search := c.Query("search")

	if category != "" || search != "" {
		var filtered []*torr.Torrent
		for _, tr := range torrs {
			st := tr.Status()

			if category == "uncategorized" {
				if st.Category != "" {
					continue
				}
			} else if category != "" && st.Category != category {
				continue
			}
			if search != "" &&
				!strings.Contains(strings.ToLower(st.Title), strings.ToLower(search)) {
				continue
			}
			filtered = append(filtered, tr)
		}
		torrs = filtered
	}

	host := utils.GetScheme(c) + "://" + utils.GetHost(c)
	list := "#EXTM3U\n"
	hash := ""
	// fn=file.m3u fix forkplayer bug with end .m3u in link
	for _, tr := range torrs {
		if sets.BTsets != nil && sets.BTsets.MergeAllM3U {
			if st := statusFromSpec(tr); st != nil {
				list += getM3uList(st, host, false, "")
			}
		} else {
			list += "#EXTINF:0"
			if tr.Poster != "" {
				list += " tvg-logo=\"" + tr.Poster + "\""
			}
			list += " type=\"playlist\"," + tr.Title + "\n"
			list += host + "/stream/" + url.PathEscape(tr.Title) + ".m3u?link=" + tr.TorrentSpec.InfoHash.HexString() + "&m3u&fn=file.m3u\n"
		}
		hash += tr.Hash().HexString()
	}

	sendM3U(c, "all.m3u", hash, list)
}

// statusFromSpec builds a minimal *state.TorrentStatus from locally-available
// metadata (TorrentSpec.InfoBytes), without starting/adding the torrent to the BT engine
func statusFromSpec(tr *torr.Torrent) *state.TorrentStatus {
	if tr == nil || tr.TorrentSpec == nil || len(tr.TorrentSpec.InfoBytes) == 0 {
		return nil
	}

	var info metainfo.Info
	if err := bencode.Unmarshal(tr.TorrentSpec.InfoBytes, &info); err != nil {
		return nil
	}

	st := new(state.TorrentStatus)
	st.Hash = tr.TorrentSpec.InfoHash.HexString()
	st.Title = tr.Title
	st.Name = info.Name

	files := info.UpvertedFiles()
	sort.Slice(files, func(i, j int) bool {
		return strings.Join(files[i].Path, "/") < strings.Join(files[j].Path, "/")
	})

	for i, f := range files {
		path := strings.Join(f.Path, "/")
		if path == "" {
			path = info.Name
		}
		st.FileStats = append(st.FileStats, &state.TorrentFileStat{
			Id:     i + 1,
			Path:   path,
			Length: f.Length,
		})
	}

	return st
}

// playList godoc
//
//	@Summary		Get HTTP link of torrent in M3U list
//	@Description	Get HTTP link of torrent in M3U list.
//
//	@Tags			API
//
//	@Param			hash		query	string	true	"Torrent hash"
//	@Param			fromlast	query	bool	false	"From last play file"
//
//	@Produce		audio/x-mpegurl
//	@Success		200	{file}	file
//	@Router			/playlist [get]
func playList(c *gin.Context) {
	hash, _ := c.GetQuery("hash")
	_, fromlast := c.GetQuery("fromlast")
	index := c.Query("index")
	if hash == "" {
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("hash is empty"))
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
			_ = c.AbortWithError(http.StatusInternalServerError, errors.New("error get torrent info"))
			return
		}
	}

	host := utils.GetScheme(c) + "://" + utils.GetHost(c)
	list := getM3uList(tor.Status(), host, fromlast, index)
	list = "#EXTM3U\n" + list
	name := strings.ReplaceAll(c.Param("fname"), `/`, "") // strip starting / from param
	if name == "" {
		name = tor.Name() + ".m3u"
	} else if !strings.HasSuffix(strings.ToLower(name), ".m3u") && !strings.HasSuffix(strings.ToLower(name), ".m3u8") {
		name += ".m3u"
	}

	sendM3U(c, name, tor.Hash().HexString(), list)
}

func sendM3U(c *gin.Context, name, hash string, m3u string) {
	c.Header("Content-Type", "audio/x-mpegurl")
	c.Header("Connection", "close")
	if hash != "" {
		etag := hex.EncodeToString([]byte(fmt.Sprintf("%s/%s", hash, name)))
		c.Header("ETag", httptoo.EncodeQuotedString(etag))
	}
	if name == "" {
		name = "playlist.m3u"
	}
	c.Header("Content-Disposition", `attachment; filename="`+name+`"`)
	http.ServeContent(c.Writer, c.Request, name, time.Now(), bytes.NewReader([]byte(m3u)))
}

func getM3uList(tor *state.TorrentStatus, host string, fromLast bool, startIndex string) string {
	m3u := ""
	from := 0
	if startIndex != "" {
		id, err := strconv.Atoi(startIndex)
		if err == nil {
			for i, f := range tor.FileStats {
				if f.Id == id {
					from = i
					break
				}
			}
		}
	} else if fromLast {
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
				fileNamesakes := findFileNamesakes(tor.FileStats, f) // find external media with same name (audio/subtiles tracks)
				if fileNamesakes != nil {
					m3u += "#EXTVLCOPT:input-slave="         // include VLC option for external media
					for _, namesake := range fileNamesakes { // include play-links to external media, with # splitter
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
	// find files with the same name in torrent
	name := filepath.Base(strings.TrimSuffix(file.Path, filepath.Ext(file.Path)))
	var namesakes []*state.TorrentFileStat
	for _, f := range files {
		if strings.Contains(f.Path, name) { // external tracks always include name of videofile
			if f != file { // exclude itself
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
