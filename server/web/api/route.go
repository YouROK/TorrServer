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
	authorized.GET("/shutdown/*reason", shutdown)

	authorized.POST("/settings", settings)
	authorized.POST("/torznab/test", torznabTest)

	authorized.POST("/torrents", torrents)

	authorized.POST("/torrent/upload", torrentUpload)

	authorized.POST("/cache", cache)

	route.HEAD("/stream", stream)
	route.GET("/stream", stream)

	route.HEAD("/stream/*fname", stream)
	route.GET("/stream/*fname", stream)

	route.HEAD("/play/:hash/:id", play)
	route.GET("/play/:hash/:id", play)

	authorized.POST("/viewed", viewed)

	authorized.GET("/playlistall/all.m3u", allPlayList)

	route.GET("/playlist", playList)
	route.GET("/playlist/*fname", playList)

	authorized.GET("/download/:size", download)

	if config.SearchWA {
		route.GET("/search/*query", rutorSearch)
	} else {
		authorized.GET("/search/*query", rutorSearch)
	}

	if config.SearchWA {
		route.GET("/torznab/search/*query", torznabSearch)
	} else {
		authorized.GET("/torznab/search/*query", torznabSearch)
	}

	authorized.GET("/ffp/:hash/:id", ffp)
}
