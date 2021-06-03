package pages

import (
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gin-gonic/gin"

	"server/settings"
	"server/torr"
	"server/web/pages/template"
)

func SetupRoute(route *gin.RouterGroup) {
	route.GET("/", mainPage)
	route.GET("/stat", statPage)
	route.GET("/magnets", getTorrents)
}

func mainPage(c *gin.Context) {
	c.Data(200, "text/html; charset=utf-8", template.IndexHtml)
}

func statPage(c *gin.Context) {
	torr.WriteStatus(c.Writer)
	c.Status(200)
}

func getTorrents(c *gin.Context) {
	list := settings.ListTorrent()
	http := "<div>"
	for _, db := range list {
		ts := db.TorrentSpec
		mi := metainfo.MetaInfo{
			AnnounceList: ts.Trackers,
		}
		mag := mi.Magnet(ts.DisplayName, ts.InfoHash)
		// mag := mi.Magnet(&ts.InfoHash, &metainfo.Info{Name: ts.DisplayName})
		http += "<p><a href='" + mag.String() + "'>magnet:?xt=urn:btih:" + mag.InfoHash.HexString() + "</a></p>"
	}
	http += "</div>"
	c.Data(200, "text/html; charset=utf-8", []byte(http))
}
