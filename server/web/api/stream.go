package api

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"server/torr"
	"server/torr/state"
	"server/web/api/utils"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// get stat
// http://127.0.0.1:8090/stream/fname?link=...&stat
// get m3u
// http://127.0.0.1:8090/stream/fname?link=...&index=1&m3u
// http://127.0.0.1:8090/stream/fname?link=...&index=1&m3u&fromlast
// stream torrent
// http://127.0.0.1:8090/stream/fname?link=...&index=1&play
// http://127.0.0.1:8090/stream/fname?link=...&index=1&play&save
// http://127.0.0.1:8090/stream/fname?link=...&index=1&play&save&title=...&poster=...
// only save
// http://127.0.0.1:8090/stream/fname?link=...&save&title=...&poster=...

func stream(c *gin.Context) {
	link := c.Query("link")
	indexStr := c.Query("index")
	_, preload := c.GetQuery("preload")
	_, stat := c.GetQuery("stat")
	_, save := c.GetQuery("save")
	_, m3u := c.GetQuery("m3u")
	_, fromlast := c.GetQuery("fromlast")
	_, play := c.GetQuery("play")
	title := c.Query("title")
	poster := c.Query("poster")
	data := ""
	notAuth := c.GetBool("not_auth")

	if notAuth && play {
		streamNoAuth(c)
		return
	}
	if notAuth {
		c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if link == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("link should not be empty"))
		return
	}

	title, _ = url.QueryUnescape(title)
	link, _ = url.QueryUnescape(link)
	poster, _ = url.QueryUnescape(poster)

	spec, err := utils.ParseLink(link)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	tor := torr.GetTorrent(spec.InfoHash.HexString())
	if tor != nil {
		title = tor.Title
		poster = tor.Poster
		data = tor.Data
	}
	if tor == nil || tor.Stat == state.TorrentInDB {
		if title == "" {
			title = c.Param("fname")
			title, _ = url.PathUnescape(title)
			title = strings.TrimLeft(title, "/")
		}

		tor, err = torr.AddTorrent(spec, title, poster, data)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	if !tor.GotInfo() {
		c.AbortWithError(http.StatusInternalServerError, errors.New("timeout connection torrent"))
		return
	}

	// save to db
	if save {
		torr.SaveTorrentToDB(tor)
		c.Status(200) // only set status, not return
	}

	// find file
	index := -1
	if len(tor.Files()) == 1 {
		index = 1
	} else {
		ind, err := strconv.Atoi(indexStr)
		if err == nil {
			index = ind
		}
	}
	if index == -1 && play { // if file index not set and play file exec
		c.AbortWithError(http.StatusBadRequest, errors.New("\"index\" is empty or wrong"))
		return
	}
	// preload torrent
	if preload {
		torr.Preload(tor, index)
	}
	// return stat if query
	if stat {
		c.JSON(200, tor.Status())
		return
	} else
	// return m3u if query
	if m3u {
		m3ulist := "#EXTM3U\n" + getM3uList(tor.Status(), "http://"+c.Request.Host, fromlast)
		sendM3U(c, tor.Name()+".m3u", tor.Hash().HexString(), m3ulist)
		return
	} else
	// return play if query
	if play {
		tor.Stream(index, c.Request, c.Writer)
		return
	}
}

func streamNoAuth(c *gin.Context) {
	link := c.Query("link")
	indexStr := c.Query("index")
	_, preload := c.GetQuery("preload")
	title := c.Query("title")
	poster := c.Query("poster")
	data := ""

	if link == "" {
		c.AbortWithError(http.StatusBadRequest, errors.New("link should not be empty"))
		return
	}

	if title == "" {
		title = c.Param("fname")
		title, _ = url.PathUnescape(title)
		title = strings.TrimLeft(title, "/")
	} else {
		title, _ = url.QueryUnescape(title)
	}

	link, _ = url.QueryUnescape(link)
	poster, _ = url.QueryUnescape(poster)

	spec, err := utils.ParseLink(link)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	tor := torr.GetTorrent(spec.InfoHash.HexString())
	if tor != nil {
		title = tor.Title
		poster = tor.Poster
		data = tor.Data
	}
	if tor == nil || tor.Stat == state.TorrentInDB {
		if title == "" {
			title = c.Param("fname")
			title, _ = url.PathUnescape(title)
			title = strings.TrimLeft(title, "/")
		}

		tor, err = torr.AddTorrent(spec, title, poster, data)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	if !tor.GotInfo() {
		c.AbortWithError(http.StatusInternalServerError, errors.New("timeout connection torrent"))
		return
	}

	// find file
	index := -1
	if len(tor.Files()) == 1 {
		index = 1
	} else {
		ind, err := strconv.Atoi(indexStr)
		if err == nil {
			index = ind
		}
	}
	if index == -1 { // if file index not set and play file exec
		c.AbortWithError(http.StatusBadRequest, errors.New("\"index\" is empty or wrong"))
		return
	}
	// preload torrent
	if preload {
		torr.Preload(tor, index)
	}

	tor.Stream(index, c.Request, c.Writer)
}
