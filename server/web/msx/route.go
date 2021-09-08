package msx

import "github.com/gin-gonic/gin"

func SetupRoute(route *gin.RouterGroup) {
	route.GET("/msx/start.json", msxStart)
	route.GET("/msx/torrents", msxTorrents)
	route.GET("/msx/playlist", msxPlaylist)
	route.GET("/msx/playlist/*fname", msxPlaylist)
}
