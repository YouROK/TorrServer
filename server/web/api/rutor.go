package api

import (
	"net/url"

	"github.com/gin-gonic/gin"

	"server/rutor"
	"server/rutor/models"
)

func rutorSearch(c *gin.Context) {
	query := c.Query("query")
	query, _ = url.QueryUnescape(query)
	list := rutor.Search(query)
	if list == nil {
		list = []*models.TorrentDetails{}
	}
	c.JSON(200, list)
}
