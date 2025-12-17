package api

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/rutor/models"
	sets "server/settings"
	"server/torznab"
)

// torznabSearch godoc
//
//	@Summary		Makes a torznab search
//	@Description	Makes a torznab search.
//
//	@Tags			API
//
//	@Param			query	query	string	true	"Torznab query"
//
//	@Produce		json
//	@Success		200	{array}	models.TorrentDetails	"Torznab torrent search result(s)"
//	@Router			/torznab/search [get]
func torznabSearch(c *gin.Context) {
	if !sets.BTsets.EnableTorznabSearch {
		c.JSON(http.StatusBadRequest, []string{})
		return
	}
	query := c.Query("query")
	indexStr := c.DefaultQuery("index", "-1")
	index := -1
	if i, err := strconv.Atoi(indexStr); err == nil {
		index = i
	}

	query, _ = url.QueryUnescape(query)
	list := torznab.Search(query, index)
	if list == nil {
		list = []*models.TorrentDetails{}
	}
	c.JSON(200, list)
}

type torznabTestReq struct {
	Host string `json:"host"`
	Key  string `json:"key"`
}

func torznabTest(c *gin.Context) {
	var req torznabTestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if err := torznab.Test(req.Host, req.Key); err != nil {
		c.JSON(200, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true})
}
