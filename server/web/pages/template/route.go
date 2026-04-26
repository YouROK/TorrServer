package template

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
)

func RouteWebPages(route gin.IRouter) {
	route.GET("/", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Indexhtml))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "text/html; charset=utf-8", Indexhtml)
	})

	route.GET("/apple-splash-1125-2436.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash11252436jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash11252436jpg)
	})

	route.GET("/apple-splash-1136-640.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash1136640jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash1136640jpg)
	})

	route.GET("/apple-splash-1170-2532.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash11702532jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash11702532jpg)
	})

	route.GET("/apple-splash-1242-2208.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash12422208jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash12422208jpg)
	})

	route.GET("/apple-splash-1242-2688.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash12422688jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash12422688jpg)
	})

	route.GET("/apple-splash-1284-2778.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash12842778jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash12842778jpg)
	})

	route.GET("/apple-splash-1334-750.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash1334750jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash1334750jpg)
	})

	route.GET("/apple-splash-1536-2048.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash15362048jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash15362048jpg)
	})

	route.GET("/apple-splash-1620-2160.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash16202160jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash16202160jpg)
	})

	route.GET("/apple-splash-1668-2224.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash16682224jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash16682224jpg)
	})

	route.GET("/apple-splash-1668-2388.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash16682388jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash16682388jpg)
	})

	route.GET("/apple-splash-1792-828.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash1792828jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash1792828jpg)
	})

	route.GET("/apple-splash-2048-1536.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash20481536jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash20481536jpg)
	})

	route.GET("/apple-splash-2048-2732.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash20482732jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash20482732jpg)
	})

	route.GET("/apple-splash-2160-1620.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash21601620jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash21601620jpg)
	})

	route.GET("/apple-splash-2208-1242.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash22081242jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash22081242jpg)
	})

	route.GET("/apple-splash-2224-1668.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash22241668jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash22241668jpg)
	})

	route.GET("/apple-splash-2388-1668.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash23881668jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash23881668jpg)
	})

	route.GET("/apple-splash-2436-1125.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash24361125jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash24361125jpg)
	})

	route.GET("/apple-splash-2532-1170.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash25321170jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash25321170jpg)
	})

	route.GET("/apple-splash-2688-1242.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash26881242jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash26881242jpg)
	})

	route.GET("/apple-splash-2732-2048.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash27322048jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash27322048jpg)
	})

	route.GET("/apple-splash-2778-1284.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash27781284jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash27781284jpg)
	})

	route.GET("/apple-splash-640-1136.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash6401136jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash6401136jpg)
	})

	route.GET("/apple-splash-750-1334.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash7501334jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash7501334jpg)
	})

	route.GET("/apple-splash-828-1792.jpg", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Applesplash8281792jpg))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/jpeg", Applesplash8281792jpg)
	})

	route.GET("/asset-manifest.json", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Assetmanifestjson))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "application/json", Assetmanifestjson)
	})

	route.GET("/browserconfig.xml", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Browserconfigxml))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "text/xml; charset=utf-8", Browserconfigxml)
	})

	route.GET("/dlnaicon-120.png", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Dlnaicon120png))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/png", Dlnaicon120png)
	})

	route.GET("/dlnaicon-48.png", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Dlnaicon48png))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/png", Dlnaicon48png)
	})

	route.GET("/favicon-16x16.png", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Favicon16x16png))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/png", Favicon16x16png)
	})

	route.GET("/favicon-32x32.png", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Favicon32x32png))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/png", Favicon32x32png)
	})

	route.GET("/favicon.ico", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Faviconico))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/vnd.microsoft.icon", Faviconico)
	})

	route.GET("/icon.png", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Iconpng))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/png", Iconpng)
	})

	route.GET("/index.html", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Indexhtml))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "text/html; charset=utf-8", Indexhtml)
	})

	route.GET("/logo.png", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Logopng))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/png", Logopng)
	})

	route.GET("/mstile-150x150.png", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Mstile150x150png))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "image/png", Mstile150x150png)
	})

	route.GET("/site.webmanifest", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Sitewebmanifest))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "application/manifest+json", Sitewebmanifest)
	})

	route.GET("/static/js/2.d0681903.chunk.js", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Staticjs2d0681903chunkjs))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "text/javascript; charset=utf-8", Staticjs2d0681903chunkjs)
	})

	route.GET("/static/js/2.d0681903.chunk.js.LICENSE.txt", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Staticjs2d0681903chunkjsLICENSEtxt))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "text/plain; charset=utf-8", Staticjs2d0681903chunkjsLICENSEtxt)
	})

	route.GET("/static/js/2.d0681903.chunk.js.map", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Staticjs2d0681903chunkjsmap))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "application/json", Staticjs2d0681903chunkjsmap)
	})

	route.GET("/static/js/main.53e2dc27.chunk.js", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Staticjsmain53e2dc27chunkjs))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "text/javascript; charset=utf-8", Staticjsmain53e2dc27chunkjs)
	})

	route.GET("/static/js/main.53e2dc27.chunk.js.map", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Staticjsmain53e2dc27chunkjsmap))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "application/json", Staticjsmain53e2dc27chunkjsmap)
	})

	route.GET("/static/js/runtime-main.5ed86a79.js", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Staticjsruntimemain5ed86a79js))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "text/javascript; charset=utf-8", Staticjsruntimemain5ed86a79js)
	})

	route.GET("/static/js/runtime-main.5ed86a79.js.map", func(c *gin.Context) {
		etag := fmt.Sprintf("%x", md5.Sum(Staticjsruntimemain5ed86a79jsmap))
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Header("ETag", etag)
		c.Data(200, "application/json", Staticjsruntimemain5ed86a79jsmap)
	})
}
