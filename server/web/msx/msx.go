package msx

import (
	_ "embed"
	
	"server/version"

	"github.com/gin-gonic/gin"
)

var (
	//go:embed assets/tvx.js.gz
	tvx []byte
	//go:embed assets/tizen.html.gz
	tzn []byte
	//go:embed assets/torrents.min.html.gz
	trn []byte
	//go:embed assets/html5x.html.gz
	h5x []byte
	//go:embed assets/russian.json.gz
	rus []byte
)

func ass(b []byte, t string) func(*gin.Context) {
	return func(c *gin.Context) {
		c.Header("Content-Encoding", "gzip")
		c.Data(200, t+"; charset=UTF-8", b)
	}
}

func SetupRoute(r *gin.RouterGroup) {
	r.GET("/msx/start.json", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":      "TorrServer",
			"version":   "0.0.1",
			"parameter": "content:request:interaction:init@{PREFIX}{SERVER}/msx/torrents",
		})
	})
	r.GET("/msx/russian.json", ass(rus, "application/json"))
	r.GET("/msx/tvx.js", ass(tvx, "text/javascript"))
	r.GET("/msx/torrents", ass(trn, "text/html"))
	r.GET("/msx/tizen", ass(tzn, "text/html"))
	r.GET("/msx/html5x", ass(h5x, "text/html"))
}
