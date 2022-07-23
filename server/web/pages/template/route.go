package template

import (
	"github.com/gin-gonic/gin"
)

func RouteWebPages(route *gin.RouterGroup) {
	route.GET("/", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Indexhtml)
	})

	route.GET("/apple-icon-180.png", func(c *gin.Context) {
		c.Data(200, "image/png", Appleicon180png)
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
		c.Data(200, "application/xml; charset=utf-8", Browserconfigxml)
	})

	route.GET("/dlnaicon-120.png", func(c *gin.Context) {
		c.Data(200, "image/png", Dlnaicon120png)
	})

	route.GET("/dlnaicon-48.png", func(c *gin.Context) {
		c.Data(200, "image/png", Dlnaicon48png)
	})

	route.GET("/favicon-196.png", func(c *gin.Context) {
		c.Data(200, "image/png", Favicon196png)
	})

	route.GET("/index.html", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", Indexhtml)
	})

	route.GET("/logo.png", func(c *gin.Context) {
		c.Data(200, "image/png", Logopng)
	})

	route.GET("/manifest-icon-192.maskable.png", func(c *gin.Context) {
		c.Data(200, "image/png", Manifesticon192maskablepng)
	})

	route.GET("/manifest-icon-512.maskable.png", func(c *gin.Context) {
		c.Data(200, "image/png", Manifesticon512maskablepng)
	})

	route.GET("/site.webmanifest", func(c *gin.Context) {
		c.Data(200, "application/manifest+json", Sitewebmanifest)
	})

	route.GET("/static/js/2.b5e598b9.chunk.js", func(c *gin.Context) {
		c.Data(200, "application/javascript; charset=utf-8", Staticjs2b5e598b9chunkjs)
	})

	route.GET("/static/js/2.b5e598b9.chunk.js.LICENSE.txt", func(c *gin.Context) {
		c.Data(200, "text/plain; charset=utf-8", Staticjs2b5e598b9chunkjsLICENSEtxt)
	})

	route.GET("/static/js/2.b5e598b9.chunk.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjs2b5e598b9chunkjsmap)
	})

	route.GET("/static/js/main.7a603c10.chunk.js", func(c *gin.Context) {
		c.Data(200, "application/javascript; charset=utf-8", Staticjsmain7a603c10chunkjs)
	})

	route.GET("/static/js/main.7a603c10.chunk.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjsmain7a603c10chunkjsmap)
	})

	route.GET("/static/js/runtime-main.40b0cc71.js", func(c *gin.Context) {
		c.Data(200, "application/javascript; charset=utf-8", Staticjsruntimemain40b0cc71js)
	})

	route.GET("/static/js/runtime-main.40b0cc71.js.map", func(c *gin.Context) {
		c.Data(200, "application/json", Staticjsruntimemain40b0cc71jsmap)
	})
}
