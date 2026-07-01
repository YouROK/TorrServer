package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	gstreamerbridge "server/gstreamer/bridge"
	"server/gstreamer"
)

type gstreamerSettingsResponse struct {
	Config   gstreamer.Config `json:"config"`
	Defaults gstreamer.Config `json:"defaults"`
}

type gstreamerSettingsRequest struct {
	Action string            `json:"action,omitempty"`
	Config *gstreamer.Config `json:"config,omitempty"`
}

// GetGStreamerSettings godoc
// @Summary Get GStreamer configuration
// @Description Retrieves current GStreamer settings and platform defaults
// @Tags API
// @Produce json
// @Success 200 {object} gstreamerSettingsResponse "GStreamer settings"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /gst/settings [get]
func GetGStreamerSettings(c *gin.Context) {
	if !gstreamerbridge.BuiltIn() {
		c.JSON(http.StatusOK, gin.H{"built_in": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"built_in": true,
		"config":   gstreamer.CurrentConfig(),
		"defaults": gstreamer.PlatformDefaults(),
	})
}

// UpdateGStreamerSettings godoc
// @Summary Update GStreamer configuration
// @Description Updates GStreamer settings in settings.json and applies them to the running server
// @Tags API
// @Accept json
// @Produce json
// @Param request body gstreamerSettingsRequest true "GStreamer settings request"
// @Success 200 {object} map[string]string "Update successful"
// @Failure 400 {object} map[string]string "Invalid input data"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Read-only mode"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /gst/settings [post]
func UpdateGStreamerSettings(c *gin.Context) {
	if !gstreamerbridge.BuiltIn() {
		c.JSON(http.StatusNotFound, gin.H{"error": "gstreamer is not built in"})
		return
	}
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
