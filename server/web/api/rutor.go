package api

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"server/rutor"
	"server/rutor/models"
	sets "server/settings"
)

func rutorSearch(c *gin.Context) {
	if !sets.BTsets.EnableRutorSearch {
		c.JSON(http.StatusBadRequest, []string{})
		return
	}
	query := c.Query("query")
	query, _ = url.QueryUnescape(query)
	list := rutor.Search(query)
	if list == nil {
		list = []*models.TorrentDetails{}
	}
	c.JSON(200, list)
}
