package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	sets "server/settings"
	"server/torr"
	"server/version"
)

type requestI struct {
	Action string `json:"action,omitempty"`
}

func SetupRoute(route *gin.Engine) {
	route.GET("/echo", echo)
	route.GET("/shutdown", shutdown)

	route.POST("/settings", settings)

	route.POST("/torrents", torrents)
	route.POST("/torrent/upload", torrentUpload)

	route.GET("/stream", stream)
	route.GET("/stream/*fname", stream)

	route.POST("/viewed", viewed)

	route.GET("/playlistall/all.m3u", allPlayList)
	route.GET("/playlist", playList)
	route.GET("/playlist/*fname", playList)
}

func echo(c *gin.Context) {
	c.String(200, "%v", version.Version)
}

func shutdown(c *gin.Context) {
	if sets.IsReadOnly() {
		c.Status(http.StatusForbidden)
		return
	}
	c.Status(200)
	go func() {
		time.Sleep(1000)
		torr.Shutdown()
	}()
}
