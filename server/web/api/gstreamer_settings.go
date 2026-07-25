//go:build gst

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"server/gstreamer"
)

type gstreamerSettingsResponse struct {
	BuiltIn  bool             `json:"built_in"`
	Config   gstreamer.Config `json:"config"`
	Defaults gstreamer.Config `json:"defaults"`
}

type gstreamerSettingsRequest struct {
	Action string            `json:"action,omitempty"`
	Config *gstreamer.Config `json:"config,omitempty"`
}

func GetGStreamerSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gstreamerSettingsResponse{
		BuiltIn:  true,
		Config:   gstreamer.CurrentConfig(),
		Defaults: gstreamer.PlatformDefaults(),
	})
}

func UpdateGStreamerSettings(c *gin.Context) {
	var req gstreamerSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch req.Action {
	case "def":
		if err := gstreamer.ResetConfig(); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	case "set", "":
		if req.Config == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "config is required"})
			return
		}
		if err := gstreamer.UpdateConfig(*req.Config); err != nil {
			if err.Error() == "read-only mode" {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown action"})
	}
}
