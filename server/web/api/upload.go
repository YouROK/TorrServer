package api

import (
	"mime/multipart"
	"net/http"

	"server/log"
	set "server/settings"
	"server/torr"
	"server/torr/state"
	"server/web/api/utils"

	"github.com/gin-gonic/gin"
)

// torrentUpload godoc
//
//	@Summary		Add .torrent files
//	@Description	Supports multiple files. Returns array of statuses.
//
//	@Tags			API
//
//	@Param			file	formData	file	true	"Torrent file(s) to insert"
//	@Param			save	formData	string	false	"Save to DB"
//	@Param			title	formData	string	false	"Torrent title (single file only)"
//	@Param			category	formData	string	false	"Torrent category"
//	@Param			poster	formData	string	false	"Torrent poster (single file only)"
//	@Param			data	formData	string	false	"Torrent data"
//
//	@Accept			multipart/form-data
//
//	@Produce		json
//	@Success		200	{array}		state.TorrentStatus	"Torrent statuses"
//	@Router			/torrent/upload [post]
func torrentUpload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer form.RemoveAll()

	save := len(form.Value["save"]) > 0
	title := ""
	if len(form.Value["title"]) > 0 {
		title = form.Value["title"][0]
	}
	category := ""
	if len(form.Value["category"]) > 0 {
		category = form.Value["category"][0]
	}
	poster := ""
	if len(form.Value["poster"]) > 0 {
		poster = form.Value["poster"][0]
	}
	data := ""
	if len(form.Value["data"]) > 0 {
		data = form.Value["data"][0]
	}

	var files []*multipart.FileHeader
	for _, fh := range form.File {
		files = append(files, fh...)
	}

	var stats []*state.TorrentStatus
	for _, fh := range files {
		log.TLogln("add .torrent", fh.Filename)

		torrFile, err := fh.Open()
		if err != nil {
			log.TLogln("error upload torrent:", err)
			continue
		}

		spec, err := utils.ParseFile(torrFile)
		torrFile.Close()
		if err != nil {
			log.TLogln("error upload torrent:", err)
			continue
		}

		tor, err := torr.AddTorrent(spec, title, poster, data, category)
		if err != nil {
			log.TLogln("error upload torrent:", err)
			continue
		}

		if tor.Data != "" && set.BTsets.EnableDebug {
			log.TLogln("torrent data:", tor.Data)
		}
		if tor.Category != "" && set.BTsets.EnableDebug {
			log.TLogln("torrent category:", tor.Category)
		}

		go func(t *torr.Torrent) {
			if !t.GotInfo() {
				log.TLogln("error add torrent:", "torrent connection timeout")
				return
			}

			if t.Title == "" {
				t.Title = t.Name()
			}

			if save {
				torr.SaveTorrentToDB(t)
			}
		}(tor)

		stats = append(stats, tor.Status())
	}
	c.JSON(200, stats)
}
