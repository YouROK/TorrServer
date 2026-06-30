//go:build gst

package bridge

import (
	"server/gstreamer"

	"github.com/gin-gonic/gin"
)

func SetupRoute(route gin.IRouter) {
	gstreamer.SetupRoute(route)
}

func BuiltIn() bool {
	return true
}

func Stop() {
	gstreamer.Stop()
}

func Remove(hash string) bool {
	return gstreamer.Remove(hash)
}
