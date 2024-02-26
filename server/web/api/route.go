package api

import (
	config "server/settings"
	"server/web/auth"

	"github.com/gin-gonic/gin"
)

type requestI struct {
	Action string `json:"action,omitempty"`
}

func SetupRoute(route gin.IRouter) {
	authorized := route.Group("/", auth.CheckAuth())

	authorized.GET("/shutdown", shutdown)

	authorized.POST("/settings", settings)

	authorized.POST("/torrents", torrents)
	authorized.POST("/torrent/upload", torrentUpload)

	authorized.POST("/cache", cache)

	route.HEAD("/stream", stream)
	route.HEAD("/stream/*fname", stream)

	route.GET("/stream", stream)
	route.GET("/stream/*fname", stream)

	route.HEAD("/play/:hash/:id", play)
	route.GET("/play/:hash/:id", play)

	authorized.POST("/viewed", viewed)

	authorized.GET("/playlistall/all.m3u", allPlayList)
	route.GET("/playlist", playList)
	route.GET("/playlist/*fname", playList) // Is this endpoint still needed ? `fname` is never used in handler

	authorized.GET("/download/:size", download)

	if config.SearchWA {
		route.GET("/search/*query", rutorSearch)
	} else {
		authorized.GET("/search/*query", rutorSearch)
	}

	authorized.GET("/ffp/:hash/:id", ffp)
}
