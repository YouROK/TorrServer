package msx

import (
	_ "embed"

	"server/version"

	"github.com/gin-gonic/gin"
)

var (
	//go:embed assets/tvx.js.gz
	tvx []byte
	//go:embed assets/tizen.js.gz
	tzn []byte
	//go:embed assets/torrents.js.gz
	trs []byte
	//go:embed assets/torrent.js.gz
	trn []byte
	//go:embed assets/html5x.html.gz
	h5x []byte
	//go:embed assets/russian.json.gz
	rus []byte
)

func ass(c *gin.Context, b []byte, t string) {
	c.Header("Content-Encoding", "gzip")
	c.Data(200, t+"; charset=UTF-8", b)
}

func SetupRoute(r *gin.RouterGroup) {
	r.GET("/msx/:pth", func(c *gin.Context) {
		s := []string{"tvx", "tizen"}
		switch p := c.Param("pth"); p {
		case "start.json":
			c.JSON(200, gin.H{
				"name":      "TorrServer",
				"version":   version.Version,
				"parameter": "content:request:interaction:init@{PREFIX}{SERVER}/msx/torrents",
			})
		case "russian.json":
			ass(c, rus, "application.json")
		case "html5x":
			ass(c, h5x, "text/html")
		case "tvx.js":
			ass(c, tvx, "text/javascript")
		case "tizen.js":
			ass(c, tzn, "text/javascript")
		case "torrents.js":
			ass(c, trs, "text/javascript")
		case "torrent.js":
			ass(c, trn, "text/javascript")
		case "torrents":
			s = append(s, p)
			p = "torrent"
			fallthrough
		case "torrent":
			s = append(s, p)
			b := []byte("<!DOCTYPE html>\n<html>\n<head>\n<title>TorrServer Interaction Plugin</title>\n<meta charset='UTF-8' />\n")
			for _, j := range s {
				b = append(b, "<script type='text/javascript' src='"+j+".js'></script>\n"...)
			}
			c.Data(200, "text/html", append(b, "</head>\n<body></body>\n</html>"...))
		default:
			c.AbortWithStatus(404)
		}
	})
}
