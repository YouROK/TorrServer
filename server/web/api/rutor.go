package api

import (
	"github.com/gin-gonic/gin"
	"server/rutor"
	"server/rutor/models"
)

func rutorSearch(c *gin.Context) {
	query := c.Query("query")
	list := rutor.Search(query)
	if list == nil {
		list = []*models.TorrentDetails{}
	}
	c.JSON(200, list)
}
