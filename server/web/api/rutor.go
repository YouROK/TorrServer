package api

import (
	"github.com/gin-gonic/gin"
	"net/url"
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
