package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	sets "server/settings"
	"server/torr"
)

//Action: get, set, def
type setsReqJS struct {
	requestI
	Sets *sets.BTSets `json:"sets,omitempty"`
}

func settings(c *gin.Context) {
	var req setsReqJS
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.Action == "get" {
		c.JSON(200, sets.BTsets)
		return
	} else if req.Action == "set" {
		torr.SetSettings(req.Sets)
		c.Status(200)
		return
	} else if req.Action == "def" {
		torr.SetDefSettings()
		c.Status(200)
		return
	}
	c.AbortWithError(http.StatusBadRequest, errors.New("action is empty"))
}
