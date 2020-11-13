package api

import (
	"github.com/gin-gonic/gin"
	"server/version"
)

type requestI struct {
	Action string `json:"action,omitempty"`
}

func SetupRoute(route *gin.Engine) {
	route.GET("/echo", echo)

	route.POST("/settings", settings)

	route.POST("/torrents", torrents)
	route.POST("/torrent/upload", torrentUpload)

	route.GET("/stream", stream)
	route.GET("/stream/*fname", stream)

	route.POST("/viewed", viewed)

	route.GET("/playlist/all.m3u", allPlayList)
	route.GET("/playlist", playList)
}

func echo(c *gin.Context) {
	c.String(200, "%v", version.Version)
}
