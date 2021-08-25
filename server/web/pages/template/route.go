package template

import (
	"github.com/gin-gonic/gin"
)

func RouteWebPages(route *gin.RouterGroup) {
	route.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Indexhtml)
	})

	route.GET("/android-chrome-512x512.png", func(c *gin.Context) {
		c.Data(200, "image/png", Androidchrome512x512png)
	})


	route.GET("/favicon-32x32.png", func(c *gin.Context) {
		c.Data(200, "image/png", Favicon32x32png)
	})


	route.GET("/site.webmanifest", func(c *gin.Context) {
		c.Data(200, "application/manifest+json", Sitewebmanifest)
	})


	route.GET("/static/js/main.0baba7bd.chunk.js", func(c *gin.Context) {
		c.Data(200, "text/javascript; charset=utf-8", Staticjsmain0baba7bdchunkjs)
	})


	route.GET("/static/js/main.0baba7bd.chunk.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjsmain0baba7bdchunkjsmap)
	})


	route.GET("/apple-touch-icon.png", func(c *gin.Context) {
		c.Data(200, "image/png", Appletouchiconpng)
	})


	route.GET("/asset-manifest.json", func(c *gin.Context) {
		c.Data(200, "application/json", Assetmanifestjson)
	})


	route.GET("/index.html", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Indexhtml)
	})


	route.GET("/static/js/2.ed783265.chunk.js", func(c *gin.Context) {
		c.Data(200, "text/javascript; charset=utf-8", Staticjs2ed783265chunkjs)
	})


	route.GET("/static/js/2.ed783265.chunk.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjs2ed783265chunkjsmap)
	})


	route.GET("/browserconfig.xml", func(c *gin.Context) {
		c.Data(200, "text/xml; charset=utf-8", Browserconfigxml)
	})


	route.GET("/favicon-16x16.png", func(c *gin.Context) {
		c.Data(200, "image/png", Favicon16x16png)
	})


	route.GET("/mstile-150x150.png", func(c *gin.Context) {
		c.Data(200, "image/png", Mstile150x150png)
	})


	route.GET("/static/js/2.ed783265.chunk.js.LICENSE.txt", func(c *gin.Context) {
		c.Data(200, "text/plain; charset=utf-8", Staticjs2ed783265chunkjsLICENSEtxt)
	})


	route.GET("/static/js/runtime-main.8bda5920.js", func(c *gin.Context) {
		c.Data(200, "text/javascript; charset=utf-8", Staticjsruntimemain8bda5920js)
	})


	route.GET("/static/js/runtime-main.8bda5920.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjsruntimemain8bda5920jsmap)
	})


	route.GET("/android-chrome-192x192.png", func(c *gin.Context) {
		c.Data(200, "image/png", Androidchrome192x192png)
	})


	route.GET("/favicon.ico", func(c *gin.Context) {
		c.Data(200, "image/vnd.microsoft.icon", Faviconico)
	})

}