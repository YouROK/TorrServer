package pages

import (
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gin-gonic/gin"

	"server/settings"
	"server/torr"
	"server/web/auth"
	"server/web/pages/template"
)

func SetupRoute(route gin.IRouter) {
	authorized := route.Group("/", auth.CheckAuth())

	// SPA + static assets are public so the web UI can show a custom Basic login
	// form. API routes (and /stat, /magnets) stay behind CheckAuth.
	template.RouteWebPages(route)
	authorized.GET("/stat", statPage)
	authorized.GET("/magnets", getTorrents)
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
