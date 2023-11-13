package pages

import (
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gin-gonic/gin"

	"server/settings"
	"server/torr"
	"server/web/pages/template"
)

func SetupRoute(route *gin.RouterGroup) {
	template.RouteWebPages(route)
	route.GET("/stat", statPage)
	route.GET("/magnets", getTorrents)
}

// stat godoc
//
//	@Summary		Stat server
//	@Description	Stat server.
//
//	@Tags			Pages
//
//	@Produce		text/plain
//	@Success		200	"Stats"
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
//	@Success		200	"Magnet links"
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
