package msx

import "github.com/gin-gonic/gin"

func SetupRoute(route *gin.RouterGroup) {
	route.GET("/msx/start.json", msxStart)
	route.GET("/msx/torrents", msxTorrents)
	route.GET("/msx/playlist", msxPlaylist)
	route.GET("/msx/playlist/*fname", msxPlaylist)

	route.GET("/msx/html5x.html", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Msxhtml5xhtml)
	})

	route.GET("/msx/tizen.html", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Msxtizenhtml)
	})

	route.GET("/msx/tvx-plugin.min.js", func(c *gin.Context) {
		c.Data(200, "text/javascript; charset=utf-8", Msxtvxpluginminjs)
	})
}
