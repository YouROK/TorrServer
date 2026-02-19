package pages

import (
	"server/proxy"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/gin-gonic/gin"

	"server/settings"
	"server/torr"
	"server/web/auth"
	"server/web/pages/template"

	"golang.org/x/exp/slices"
)

func SetupRoute(route gin.IRouter) {
	authorized := route.Group("/", auth.CheckAuth())

	webPagesAuth := route.Group("/", func() gin.HandlerFunc {
		return func(c *gin.Context) {
			if slices.Contains([]string{"/site.webmanifest"}, c.FullPath()) {
				return
			}
			auth.CheckAuth()(c)
		}
	}())

	template.RouteWebPages(webPagesAuth)
	authorized.GET("/stat", statPage)
	authorized.GET("/magnets", getTorrents)
	authorized.GET("/proxy", proxy.P2Proxy.GinHandler)
}

// stat godoc
//
//	@Summary		TorrServer Statistics
//	@Description	Show server and torrents statistics.
//
//	@Tags			Pages
//
//	@Produce		text/plain
//	@Success		200	"TorrServer statistics"
//	@Router			/stat [get]
func statPage(c *gin.Context) {
	torr.WriteStatus(c.Writer)
	c.Status(200)
}

// getTorrents godoc
//
//	@Summary		Get HTML of magnet links
//	@Description	Get HTML of magnet links.
//
//	@Tags			Pages
//
//	@Produce		text/html
//	@Success		200	"HTML with Magnet links"
//	@Router			/magnets [get]
func getTorrents(c *gin.Context) {
	list := settings.ListTorrent()
	http := "<div>"
	for _, db := range list {
		ts := db.TorrentSpec
		mi := metainfo.MetaInfo{
			AnnounceList: ts.Trackers,
		}
		// mag := mi.Magnet(ts.DisplayName, ts.InfoHash)
		mag := mi.Magnet(&ts.InfoHash, &metainfo.Info{Name: ts.DisplayName})
		http += "<p><a href='" + mag.String() + "'>magnet:?xt=urn:btih:" + mag.InfoHash.HexString() + "</a></p>"
	}
	http += "</div>"
	c.Data(200, "text/html; charset=utf-8", []byte(http))
}
