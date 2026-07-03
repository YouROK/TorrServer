//go:build !gst

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type gstreamerStubSettingsResponse struct {
	BuiltIn bool `json:"built_in"`
}

func GetGStreamerSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gstreamerStubSettingsResponse{BuiltIn: false})
}

func UpdateGStreamerSettings(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "gstreamer is not built in"})
}
