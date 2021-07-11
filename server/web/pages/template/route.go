package template

import (
	"github.com/gin-gonic/gin"
)

func RouteWebPages(route *gin.RouterGroup) {
	route.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Indexhtml)
	})

	route.GET("/mstile-150x150.png", func(c *gin.Context) {
		c.Data(200, "image/png", Mstile150x150png)
	})


	route.GET("/static/js/2.c8261c65.chunk.js", func(c *gin.Context) {
		c.Data(200, "application/javascript", Staticjs2c8261c65chunkjs)
	})


	route.GET("/apple-touch-icon.png", func(c *gin.Context) {
		c.Data(200, "image/png", Appletouchiconpng)
	})


	route.GET("/static/js/runtime-main.8bda5920.js", func(c *gin.Context) {
		c.Data(200, "application/javascript", Staticjsruntimemain8bda5920js)
	})


	route.GET("/favicon.ico", func(c *gin.Context) {
		c.Data(200, "image/vnd.microsoft.icon", Faviconico)
	})


	route.GET("/browserconfig.xml", func(c *gin.Context) {
		c.Data(200, "application/xml", Browserconfigxml)
	})


	route.GET("/favicon-16x16.png", func(c *gin.Context) {
		c.Data(200, "image/png", Favicon16x16png)
	})


	route.GET("/favicon-32x32.png", func(c *gin.Context) {
		c.Data(200, "image/png", Favicon32x32png)
	})


	route.GET("/index.html", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Indexhtml)
	})


	route.GET("/site.webmanifest", func(c *gin.Context) {
		c.Data(200, "application/manifest+json", Sitewebmanifest)
	})


	route.GET("/static/js/2.c8261c65.chunk.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjs2c8261c65chunkjsmap)
	})


	route.GET("/static/js/main.c71e4ba7.chunk.js", func(c *gin.Context) {
		c.Data(200, "application/javascript", Staticjsmainc71e4ba7chunkjs)
	})


	route.GET("/android-chrome-192x192.png", func(c *gin.Context) {
		c.Data(200, "image/png", Androidchrome192x192png)
	})


	route.GET("/static/js/runtime-main.8bda5920.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjsruntimemain8bda5920jsmap)
	})


	route.GET("/static/js/main.c71e4ba7.chunk.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjsmainc71e4ba7chunkjsmap)
	})


	route.GET("/asset-manifest.json", func(c *gin.Context) {
		c.Data(200, "application/json", Assetmanifestjson)
	})


	route.GET("/static/js/2.c8261c65.chunk.js.LICENSE.txt", func(c *gin.Context) {
		c.Data(200, "text/plain; charset=utf-8", Staticjs2c8261c65chunkjsLICENSEtxt)
	})


	route.GET("/android-chrome-512x512.png", func(c *gin.Context) {
		c.Data(200, "image/png", Androidchrome512x512png)
	})

}