//go:build !gst

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetGStreamerSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"built_in": false})
}

func UpdateGStreamerSettings(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "gstreamer is not built in"})
}
