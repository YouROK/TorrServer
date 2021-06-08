package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	sets "server/settings"
	"server/torr"
)

type requestI struct {
	Action string `json:"action,omitempty"`
}

func SetupRoute(route *gin.RouterGroup) {
	route.GET("/shutdown", shutdown)

	route.POST("/settings", settings)

	route.POST("/torrents", torrents)
	route.POST("/torrent/upload", torrentUpload)

	route.POST("/cache", cache)

	route.HEAD("/stream", stream)
	route.HEAD("/stream/*fname", stream)

	route.GET("/stream", stream)
	route.GET("/stream/*fname", stream)

	route.HEAD("/play/:hash/:id", play)
	route.GET("/play/:hash/:id", play)

	route.POST("/viewed", viewed)

	route.GET("/playlistall/all.m3u", allPlayList)
	route.GET("/playlist", playList)
	route.GET("/playlist/*fname", playList)
}

func shutdown(c *gin.Context) {
	if sets.ReadOnly {
		c.Status(http.StatusForbidden)
		return
	}
	c.Status(200)
	go func() {
		time.Sleep(1000)
		torr.Shutdown()
	}()
}
