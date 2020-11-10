package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	settings2 "server/settings"
)

//Action: get, set
type setsReqJS struct {
	requestI
	Sets *settings2.BTSets `json:"sets,omitempty"`
}

func settings(c *gin.Context) {
	var req setsReqJS
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.Action == "get" {
		c.JSON(200, settings2.BTsets)
		return
	}
	if req.Action == "set" {
		settings2.SetBTSets(req.Sets)
		c.Status(200)
		return
	}
	c.AbortWithError(http.StatusBadRequest, errors.New("action is empty"))
}
