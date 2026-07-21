package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	sets "server/settings"
	"server/torr"
	"server/torr/state"
	"server/web/api/utils"
)

// play godoc
//
//	@Summary		Play given torrent by infohash
//	@Description	Play given torrent referenced by infohash and file id.
//
//	@Tags			API
//
//	@Param			hash		path	string	true	"Torrent infohash"
//	@Param			id			path	string	true	"File index in torrent"
//
//	@Produce		application/octet-stream
//	@Success		200	"Torrent data"
//	@Router			/play/{hash}/{id} [get]
func play(c *gin.Context) {
	hash := c.Param("hash")
	indexStr := c.Param("id")
	notAuth := c.GetBool("auth_required") && c.GetString(gin.AuthUserKey) == ""

	if hash == "" || indexStr == "" {
		_ = c.AbortWithError(http.StatusNotFound, errors.New("no infohash or file index in link"))
		return
	}

	spec, err := utils.ParseLink(hash)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	tor := torr.GetTorrent(spec.InfoHash.HexString())
	if tor == nil && notAuth {
		if !sets.StreamWA {
			c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tor, err = torr.AddTorrent(spec, "", "", "", "")
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	if tor == nil {
		_ = c.AbortWithError(http.StatusInternalServerError, errors.New("error get torrent"))
		return
	}

	if tor.Stat == state.TorrentInDB {
		tor, err = torr.AddTorrent(spec, tor.Title, tor.Poster, tor.Data, tor.Category)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	if !tor.GotInfo() {
		_ = c.AbortWithError(http.StatusInternalServerError, errors.New("torrent connection timeout"))
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
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("file \"index\" is wrong"))
		return
	}

	_ = tor.Stream(index, c.Request, c.Writer)
}
