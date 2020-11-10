package api

import (
	"github.com/gin-gonic/gin"
	"server/torr"
	"server/version"
)

type requestI struct {
	Action string `json:"action,omitempty"`
}

type responseI struct {
}

var bts *torr.BTServer

func SetupRouteApi(route *gin.Engine, serv *torr.BTServer) {
	bts = serv
	route.GET("/echo", echo)

	route.POST("/settings", settings)

	route.POST("/torrents", torrents)

	route.GET("/stream", stream)
	route.GET("/stream/*fname", stream)

	route.POST("/viewed", viewed)

	route.GET("/playlist/all.m3u", allPlayList)
	route.GET("/playlist", playList)
}

func echo(c *gin.Context) {
	c.String(200, "{\"version\": \"%v\"}", version.Version)
}
