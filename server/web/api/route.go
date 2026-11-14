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
	authorized.GET("/waf", getWAF)
	authorized.GET("/mediabase", mediaBase)
	authorized.POST("/waf", updateWAF)
	authorized.POST("/torznab/test", torznabTest)

	authorized.POST("/torrents", torrents)

	authorized.POST("/torrent/upload", torrentUpload)

	authorized.POST("/cache", cache)

	authorized.HEAD("/stream", stream)
	authorized.GET("/stream", stream)

	authorized.HEAD("/stream/*fname", stream)
	authorized.GET("/stream/*fname", stream)

	authorized.HEAD("/play/:hash/:id", play)
	authorized.GET("/play/:hash/:id", play)

	authorized.GET("/subs/:hash/:id", subsTracks)

	authorized.POST("/viewed", viewed)

	authorized.GET("/playlistall/all.m3u", allPlayList)

	authorized.GET("/playlist", playList)
	authorized.GET("/playlist/*fname", playList)

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
	// Storage settings endpoints
	authorized.GET("/storage/settings", GetStorageSettings)
	authorized.POST("/storage/settings", UpdateStorageSettings)
	// TMDB settings endpoint
	authorized.GET("/tmdb/settings", tmdbSettings)
	// GStreamer
	authorized.GET("/gst/settings", GetGStreamerSettings)
	authorized.POST("/gst/settings", UpdateGStreamerSettings)
	// FFProbe
	authorized.GET("/ffp/status", ffprobeStatus)
	authorized.GET("/ffp/:hash/:id", ffp)
}
