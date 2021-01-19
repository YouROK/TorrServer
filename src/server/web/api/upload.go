package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"server/log"
	"server/torr"
	"server/torr/state"
	"server/web/api/utils"
)

func torrentUpload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer form.RemoveAll()

	save := len(form.Value["save"]) > 0
	var retList []*state.TorrentStatus
	title := ""
	if len(form.Value["title"]) > 0 {
		title = form.Value["title"][0]
	}
	poster := ""
	if len(form.Value["poster"]) > 0 {
		poster = form.Value["poster"][0]
	}
	data := ""
	if len(form.Value["data"]) > 0 {
		data = form.Value["data"][0]
	}

	for name, file := range form.File {
		log.TLogln("add torrent file", name)

		torrFile, err := file[0].Open()
		if err != nil {
			log.TLogln("error upload torrent:", err)
			continue
		}
		defer torrFile.Close()

		spec, err := utils.ParseFile(torrFile)
		if err != nil {
			log.TLogln("error upload torrent:", err)
			continue
		}

		tor, err := torr.AddTorrent(spec, title, poster, data)
		if err != nil {
			log.TLogln("error upload torrent:", err)
			continue
		}

		go func() {
			if !tor.GotInfo() {
				log.TLogln("error add torrent:", "timeout connection torrent")
				return
			}

			if save {
				torr.SaveTorrentToDB(tor)
			}
		}()

		retList = append(retList, tor.Status())
	}
	c.JSON(200, retList)
}
