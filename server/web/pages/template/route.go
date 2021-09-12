package template

import (
	"github.com/gin-gonic/gin"
)

func RouteWebPages(route *gin.RouterGroup) {
	route.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Indexhtml)
	})

	route.GET("/android-chrome-192x192.png", func(c *gin.Context) {
		c.Data(200, "image/png", Androidchrome192x192png)
	})

	route.GET("/android-chrome-512x512.png", func(c *gin.Context) {
		c.Data(200, "image/png", Androidchrome512x512png)
	})

	route.GET("/apple-touch-icon.png", func(c *gin.Context) {
		c.Data(200, "image/png", Appletouchiconpng)
	})

	route.GET("/asset-manifest.json", func(c *gin.Context) {
		c.Data(200, "application/json", Assetmanifestjson)
	})

	route.GET("/browserconfig.xml", func(c *gin.Context) {
		c.Data(200, "application/xml", Browserconfigxml)
	})

	route.GET("/dlnaicon-120.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Dlnaicon120jpg)
	})

	route.GET("/dlnaicon-120.png", func(c *gin.Context) {
		c.Data(200, "image/png", Dlnaicon120png)
	})

	route.GET("/dlnaicon-48.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Dlnaicon48jpg)
	})

	route.GET("/dlnaicon-48.png", func(c *gin.Context) {
		c.Data(200, "image/png", Dlnaicon48png)
	})

	route.GET("/favicon-16x16.png", func(c *gin.Context) {
		c.Data(200, "image/png", Favicon16x16png)
	})

	route.GET("/favicon-32x32.png", func(c *gin.Context) {
		c.Data(200, "image/png", Favicon32x32png)
	})

	route.GET("/favicon.ico", func(c *gin.Context) {
		c.Data(200, "image/x-icon", Faviconico)
	})

	route.GET("/index.html", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Indexhtml)
	})

	route.GET("/mstile-150x150.png", func(c *gin.Context) {
		c.Data(200, "image/png", Mstile150x150png)
	})

	route.GET("/msx/tizen.html", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Msxtizenhtml)
	})

	route.GET("/msx/tizen.js", func(c *gin.Context) {
		c.Data(200, "application/javascript", Msxtizenjs)
	})

	route.GET("/msx/tvx-plugin.min.js", func(c *gin.Context) {
		c.Data(200, "application/javascript", Msxtvxpluginminjs)
	})

	route.GET("/msx/webOSTV-1.2.4.js", func(c *gin.Context) {
		c.Data(200, "application/javascript", MsxwebOSTV124js)
	})

	route.GET("/msx/webos.html", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Msxweboshtml)
	})

	route.GET("/site.webmanifest", func(c *gin.Context) {
		c.Data(200, "application/manifest+json", Sitewebmanifest)
	})

	route.GET("/static/js/2.74a4e2cb.chunk.js", func(c *gin.Context) {
		c.Data(200, "application/javascript", Staticjs274a4e2cbchunkjs)
	})

	route.GET("/static/js/2.74a4e2cb.chunk.js.LICENSE.txt", func(c *gin.Context) {
		c.Data(200, "text/plain; charset=utf-8", Staticjs274a4e2cbchunkjsLICENSEtxt)
	})

	route.GET("/static/js/2.74a4e2cb.chunk.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjs274a4e2cbchunkjsmap)
	})

	route.GET("/static/js/main.72a47914.chunk.js", func(c *gin.Context) {
		c.Data(200, "application/javascript", Staticjsmain72a47914chunkjs)
	})

	route.GET("/static/js/main.72a47914.chunk.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjsmain72a47914chunkjsmap)
	})

	route.GET("/static/js/runtime-main.33603a80.js", func(c *gin.Context) {
		c.Data(200, "application/javascript", Staticjsruntimemain33603a80js)
	})

	route.GET("/static/js/runtime-main.33603a80.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjsruntimemain33603a80jsmap)
	})
}
