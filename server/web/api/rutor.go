package api

import (
	"github.com/gin-gonic/gin"
	"server/rutor"
)

func rutorSearch(c *gin.Context) {
	query := c.Query("query")
	list := rutor.Search(query)
	c.JSON(200, list)
}
