//go:build gst

package bridge

import (
	"server/gstreamer"
	"server/web/auth"

	"github.com/gin-gonic/gin"
)

func SetupRoute(route gin.IRouter) {
	gstreamer.SetupRoute(route.Group("/", auth.CheckAuth()))
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
