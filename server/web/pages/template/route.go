package template

import (
	"github.com/gin-gonic/gin"
)

func RouteWebPages(route *gin.RouterGroup) {
	route.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Indexhtml)
	})

	route.GET("/apple-splash-1125-2436.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash11252436jpg)
	})

	route.GET("/apple-splash-1136-640.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash1136640jpg)
	})

	route.GET("/apple-splash-1170-2532.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash11702532jpg)
	})

	route.GET("/apple-splash-1242-2208.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash12422208jpg)
	})

	route.GET("/apple-splash-1242-2688.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash12422688jpg)
	})

	route.GET("/apple-splash-1284-2778.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash12842778jpg)
	})

	route.GET("/apple-splash-1334-750.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash1334750jpg)
	})

	route.GET("/apple-splash-1536-2048.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash15362048jpg)
	})

	route.GET("/apple-splash-1620-2160.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash16202160jpg)
	})

	route.GET("/apple-splash-1668-2224.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash16682224jpg)
	})

	route.GET("/apple-splash-1668-2388.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash16682388jpg)
	})

	route.GET("/apple-splash-1792-828.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash1792828jpg)
	})

	route.GET("/apple-splash-2048-1536.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash20481536jpg)
	})

	route.GET("/apple-splash-2048-2732.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash20482732jpg)
	})

	route.GET("/apple-splash-2160-1620.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash21601620jpg)
	})

	route.GET("/apple-splash-2208-1242.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash22081242jpg)
	})

	route.GET("/apple-splash-2224-1668.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash22241668jpg)
	})

	route.GET("/apple-splash-2388-1668.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash23881668jpg)
	})

	route.GET("/apple-splash-2436-1125.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash24361125jpg)
	})

	route.GET("/apple-splash-2532-1170.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash25321170jpg)
	})

	route.GET("/apple-splash-2688-1242.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash26881242jpg)
	})

	route.GET("/apple-splash-2732-2048.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash27322048jpg)
	})

	route.GET("/apple-splash-2778-1284.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash27781284jpg)
	})

	route.GET("/apple-splash-640-1136.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash6401136jpg)
	})

	route.GET("/apple-splash-750-1334.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash7501334jpg)
	})

	route.GET("/apple-splash-828-1792.jpg", func(c *gin.Context) {
		c.Data(200, "image/jpeg", Applesplash8281792jpg)
	})

	route.GET("/asset-manifest.json", func(c *gin.Context) {
		c.Data(200, "application/json", Assetmanifestjson)
	})

	route.GET("/browserconfig.xml", func(c *gin.Context) {
		c.Data(200, "text/xml; charset=utf-8", Browserconfigxml)
	})

	route.GET("/dlnaicon-120.png", func(c *gin.Context) {
		c.Data(200, "image/png", Dlnaicon120png)
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
		c.Data(200, "image/vnd.microsoft.icon", Faviconico)
	})

	route.GET("/icon.png", func(c *gin.Context) {
		c.Data(200, "image/png", Iconpng)
	})

	route.GET("/index.html", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Indexhtml)
	})

	route.GET("/logo.png", func(c *gin.Context) {
		c.Data(200, "image/png", Logopng)
	})

	route.GET("/mstile-150x150.png", func(c *gin.Context) {
		c.Data(200, "image/png", Mstile150x150png)
	})

	route.GET("/site.webmanifest", func(c *gin.Context) {
		c.Data(200, "application/manifest+json", Sitewebmanifest)
	})

	route.GET("/static/js/2.41a752aa.chunk.js", func(c *gin.Context) {
		c.Data(200, "text/javascript; charset=utf-8", Staticjs241a752aachunkjs)
	})

	route.GET("/static/js/2.41a752aa.chunk.js.LICENSE.txt", func(c *gin.Context) {
		c.Data(200, "text/plain; charset=utf-8", Staticjs241a752aachunkjsLICENSEtxt)
	})

	route.GET("/static/js/2.41a752aa.chunk.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjs241a752aachunkjsmap)
	})

	route.GET("/static/js/main.b1d76117.chunk.js", func(c *gin.Context) {
		c.Data(200, "text/javascript; charset=utf-8", Staticjsmainb1d76117chunkjs)
	})

	route.GET("/static/js/main.b1d76117.chunk.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjsmainb1d76117chunkjsmap)
	})

	route.GET("/static/js/runtime-main.f542387e.js", func(c *gin.Context) {
		c.Data(200, "text/javascript; charset=utf-8", Staticjsruntimemainf542387ejs)
	})

	route.GET("/static/js/runtime-main.f542387e.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjsruntimemainf542387ejsmap)
	})
}
