package api

import (
	"encoding/json"
	"net/http"

	"server/rutor"

	"server/dlna"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	sets "server/settings"
	"server/torr"
)

// Action: get, set, def
type setsReqJS struct {
	requestI
	Sets *sets.BTSets `json:"sets,omitempty"`
}

// settings godoc
//
//	@Summary		Get / Set server settings
//	@Description	Allow to get or set server settings.
//
//	@Tags			API
//
//	@Param			request	body	setsReqJS	true	"Settings request. Available params for action: get, set, def"
//
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	sets.BTSets	"Settings JSON or nothing. Depends on what action has been asked."
//	@Router			/settings [post]
func settings(c *gin.Context) {
	var req setsReqJS
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.Action == "get" {
		resp := make(map[string]interface{})
		if buf, err := json.Marshal(sets.BTsets); err == nil {
			_ = json.Unmarshal(buf, &resp)
		}
		resp["PerUserData"] = sets.PerUserData
		resp["PerUserDataForced"] = sets.Args != nil && sets.Args.PerUserData
		resp["CurrentUser"] = currentUser(c)
		c.JSON(200, resp)
		return
	} else if req.Action == "set" {
		torr.SetSettings(req.Sets)
		dlna.Stop()
		if req.Sets.EnableDLNA {
			dlna.Start()
		}
		rutor.Stop()
		rutor.Start()
		c.Status(200)
		return
	} else if req.Action == "def" {
		torr.SetDefSettings()
		dlna.Stop()
		rutor.Stop()
		c.Status(200)
		return
	}
	c.AbortWithError(http.StatusBadRequest, errors.New("action is empty"))
}
