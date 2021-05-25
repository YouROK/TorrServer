package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/torr"
	"server/torr/state"
	"server/web/api/utils"
)

func play(c *gin.Context) {
	hash := c.Param("hash")
	indexStr := c.Param("id")
	notAuth := c.GetBool("not_auth")

	if hash == "" || indexStr == "" {
		c.AbortWithError(http.StatusNotFound, errors.New("link should not be empty"))
		return
	}

	spec, err := utils.ParseLink(hash)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	tor := torr.GetTorrent(spec.InfoHash.HexString())
	if tor == nil && notAuth {
		c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if tor == nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("error get torrent"))
		return
	}

	if tor.Stat == state.TorrentInDB {
		tor, err = torr.AddTorrent(spec, tor.Title, tor.Poster, tor.Data)
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
		c.AbortWithError(http.StatusBadRequest, errors.New("\"index\" is wrong"))
		return
	}

	tor.Stream(index, c.Request, c.Writer)
	return
}
