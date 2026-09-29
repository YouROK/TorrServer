package api

import (
	"fmt"
	"net/http"

	"server/ffprobe"
	sets "server/settings"

	"github.com/gin-gonic/gin"
)

type ffprobeStatusResponse struct {
	Available bool `json:"available"`
}

// ffprobeStatus godoc
//
//	@Summary		ffprobe availability
//	@Description	Reports whether the ffprobe binary is available. Features that need real
//	@Description	media duration (e.g. saving playback position) are only usable when true.
//
//	@Tags			API
//
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200	{object}	ffprobeStatusResponse
//	@Router			/ffp/status [get]
func ffprobeStatus(c *gin.Context) {
	c.JSON(http.StatusOK, ffprobeStatusResponse{Available: ffprobe.Exists()})
}

// ffp godoc
//
//	@Summary		Gather informations using ffprobe
//	@Description	Gather informations using ffprobe.
//
//	@Tags			API
//
//	@Param			hash	path	string	true	"Torrent hash"
//	@Param			id		path	string	true	"File index in torrent"
//
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200	"Data returned from ffprobe"
//	@Router			/ffp/{hash}/{id} [get]
func ffp(c *gin.Context) {
	hash := c.Param("hash")
	indexStr := c.Param("id")

	if hash == "" || indexStr == "" {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "link should not be empty"})
		return
	}

	if !ffprobe.Exists() {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "ffprobe binary not found"})
		return
	}

	link := "http://127.0.0.1:" + sets.Port + "/play/" + hash + "/" + indexStr

	data, err := ffprobe.ProbeUrl(link)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("error getting data: %v", err).Error()})
		return
	}

	c.JSON(200, data)
}
