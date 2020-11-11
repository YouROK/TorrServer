package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	sets "server/settings"
)

/*
file index starts from 1
*/

// Action: set, rem, list
type viewedReqJS struct {
	requestI
	*sets.Viewed
}

func viewed(c *gin.Context) {
	var req viewedReqJS
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	switch req.Action {
	case "set":
		{
			setViewed(req, c)
		}
	case "rem":
		{
			remViewed(req, c)
		}
	case "list":
		{
			listViewed(req, c)
		}
	}
}

func setViewed(req viewedReqJS, c *gin.Context) {
	sets.SetViewed(req.Viewed)
	c.Status(200)
}

func remViewed(req viewedReqJS, c *gin.Context) {
	sets.RemViewed(req.Viewed)
	c.Status(200)
}

func listViewed(req viewedReqJS, c *gin.Context) {
	list := sets.ListViewed(req.Hash)
	c.JSON(200, list)
}
