package template

import (
	"github.com/gin-gonic/gin"
)

func RouteWebPages(route gin.IRouter) {
	route.GET("/", func(c *gin.Context) {
		serve(c, Indexhtml, `"4f7132e23d27b343d96f197e3d47d741"`, "text/html; charset=utf-8", false)
	})

	route.GET("/apple-splash-1125-2436.jpg", func(c *gin.Context) {
		serve(c, Applesplash11252436jpg, `"c6da0cce59bc9301c6a4ea3a84f6fd50"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-1136-640.jpg", func(c *gin.Context) {
		serve(c, Applesplash1136640jpg, `"e1e04a04ad2299e0d764d2c9c4e452c8"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-1170-2532.jpg", func(c *gin.Context) {
		serve(c, Applesplash11702532jpg, `"176cc733fc6c155af447c917e23d2764"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-1242-2208.jpg", func(c *gin.Context) {
		serve(c, Applesplash12422208jpg, `"4671924b09f29483700d3ea2e39698e7"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-1242-2688.jpg", func(c *gin.Context) {
		serve(c, Applesplash12422688jpg, `"f7a3e51c4f580fff41d7471d35399e34"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-1284-2778.jpg", func(c *gin.Context) {
		serve(c, Applesplash12842778jpg, `"34651f734d70e56497ee63a13c41e36e"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-1334-750.jpg", func(c *gin.Context) {
		serve(c, Applesplash1334750jpg, `"9a4bdfd8eda5a4167ad6aac51b01595e"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-1536-2048.jpg", func(c *gin.Context) {
		serve(c, Applesplash15362048jpg, `"967e7ce3903cb95d53d27b70b6593427"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-1620-2160.jpg", func(c *gin.Context) {
		serve(c, Applesplash16202160jpg, `"0673c798f199c62d64ddbfd670fc5a23"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-1668-2224.jpg", func(c *gin.Context) {
		serve(c, Applesplash16682224jpg, `"e654b5ff9d2cf5f953bafb1e0f2d3951"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-1668-2388.jpg", func(c *gin.Context) {
		serve(c, Applesplash16682388jpg, `"654845e8b706d06b132d4372b989fb54"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-1792-828.jpg", func(c *gin.Context) {
		serve(c, Applesplash1792828jpg, `"be685ad658fc32a06a1c5195a4fd2752"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-2048-1536.jpg", func(c *gin.Context) {
		serve(c, Applesplash20481536jpg, `"28e75ef301c0ee321fa5195841687918"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-2048-2732.jpg", func(c *gin.Context) {
		serve(c, Applesplash20482732jpg, `"222fb34afd905334d4e044b8cd68be28"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-2160-1620.jpg", func(c *gin.Context) {
		serve(c, Applesplash21601620jpg, `"84b9df12deb2fb9d86ec4a0a64bb13cd"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-2208-1242.jpg", func(c *gin.Context) {
		serve(c, Applesplash22081242jpg, `"3affbad7ec4b62c1f974c13b33c310e9"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-2224-1668.jpg", func(c *gin.Context) {
		serve(c, Applesplash22241668jpg, `"1fc69f033c00a39b1474ee1e4e005052"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-2388-1668.jpg", func(c *gin.Context) {
		serve(c, Applesplash23881668jpg, `"a4dc9a719066191a228d5d1c0fc01044"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-2436-1125.jpg", func(c *gin.Context) {
		serve(c, Applesplash24361125jpg, `"942bf5dcfc61b0e7af5f790fae499f9d"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-2532-1170.jpg", func(c *gin.Context) {
		serve(c, Applesplash25321170jpg, `"313a7f8ede39ec775dbf6af1c445278a"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-2688-1242.jpg", func(c *gin.Context) {
		serve(c, Applesplash26881242jpg, `"d803adabd8e93e48f6793d87011b5e72"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-2732-2048.jpg", func(c *gin.Context) {
		serve(c, Applesplash27322048jpg, `"b71f66b2a15ff51e117462768534d104"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-2778-1284.jpg", func(c *gin.Context) {
		serve(c, Applesplash27781284jpg, `"39004877aa422269dde1de601ca8f5d2"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-640-1136.jpg", func(c *gin.Context) {
		serve(c, Applesplash6401136jpg, `"1cd9a7937471c6a8d7b7da739e9aa6cd"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-750-1334.jpg", func(c *gin.Context) {
		serve(c, Applesplash7501334jpg, `"2463c1a06a023d1683b9eb6c62000db0"`, "image/jpeg", false)
	})

	route.GET("/apple-splash-828-1792.jpg", func(c *gin.Context) {
		serve(c, Applesplash8281792jpg, `"771ca6d30d3cc285f83e9e757412f968"`, "image/jpeg", false)
	})

	route.GET("/asset-manifest.json", func(c *gin.Context) {
		serve(c, Assetmanifestjson, `"3451f0806913f9df7a2b5d5cd7c661b4"`, "application/json", false)
	})

	route.GET("/browserconfig.xml", func(c *gin.Context) {
		serve(c, Browserconfigxml, `"2192958eb6798cc05ecc87827dd6dc22"`, "application/xml; charset=utf-8", false)
	})

	route.GET("/dlnaicon-120.png", func(c *gin.Context) {
		serve(c, Dlnaicon120png, `"a741da374199bad465c9e2576da666d0"`, "image/png", false)
	})

	route.GET("/dlnaicon-48.png", func(c *gin.Context) {
		serve(c, Dlnaicon48png, `"94fb147303f8625b3f0dd99a385d72d8"`, "image/png", false)
	})

	route.GET("/favicon-16x16.png", func(c *gin.Context) {
		serve(c, Favicon16x16png, `"1cabe79322b89571937189d95cbe968b"`, "image/png", false)
	})

	route.GET("/favicon-32x32.png", func(c *gin.Context) {
		serve(c, Favicon32x32png, `"d34d53cbac4d729b0f49677fecbb48ac"`, "image/png", false)
	})

	route.GET("/favicon.ico", func(c *gin.Context) {
		serve(c, Faviconico, `"80c0375465582708d791bcab9b4053af"`, "image/vnd.microsoft.icon", false)
	})

	route.GET("/icon.png", func(c *gin.Context) {
		serve(c, Iconpng, `"4425f6f2b52d2deaf2374ff63c682bcc"`, "image/png", false)
	})

	route.GET("/index.html", func(c *gin.Context) {
		serve(c, Indexhtml, `"4f7132e23d27b343d96f197e3d47d741"`, "text/html; charset=utf-8", false)
	})

	route.GET("/logo.png", func(c *gin.Context) {
		serve(c, Logopng, `"f5db7f222c6f454cdf40c0793ef529f3"`, "image/png", false)
	})

	route.GET("/lord-icon-2.0.2.js", func(c *gin.Context) {
		serve(c, Lordicon202js, `"c4f3c8819396edbaab975c9a6c9ed1d4"`, "application/javascript; charset=utf-8", false)
	})

	route.GET("/mstile-150x150.png", func(c *gin.Context) {
		serve(c, Mstile150x150png, `"5701d873784a3f27c3dffb8a15736079"`, "image/png", false)
	})

	route.GET("/site.webmanifest", func(c *gin.Context) {
		serve(c, Sitewebmanifest, `"e1cb30ce5123a7f926497f48075ad250"`, "application/manifest+json", false)
	})

	route.GET("/static/js/2.7a202c12.chunk.js", func(c *gin.Context) {
		serve(c, Staticjs27a202c12chunkjs, `"0ad1c473f77e7e1d60620956fa2a10ab"`, "application/javascript; charset=utf-8", true)
	})

	route.GET("/static/js/2.7a202c12.chunk.js.LICENSE.txt", func(c *gin.Context) {
		serve(c, Staticjs27a202c12chunkjsLICENSEtxt, `"9c4a958bc09e18ebd8a6506ecc2d412e"`, "text/plain; charset=utf-8", true)
	})

	route.GET("/static/js/2.7a202c12.chunk.js.map", func(c *gin.Context) {
		serve(c, Staticjs27a202c12chunkjsmap, `"ba6f3698e36620ac76c053fa00ab50ea"`, "application/json", true)
	})

	route.GET("/static/js/main.7ce39df0.chunk.js", func(c *gin.Context) {
		serve(c, Staticjsmain7ce39df0chunkjs, `"1492b53a0c5f0ca49bbbe3c5507f339f"`, "application/javascript; charset=utf-8", true)
	})

	route.GET("/static/js/main.7ce39df0.chunk.js.map", func(c *gin.Context) {
		serve(c, Staticjsmain7ce39df0chunkjsmap, `"f3ea59b025ec3dcd32ca41d042b64e97"`, "application/json", true)
	})

	route.GET("/static/js/runtime-main.5ed86a79.js", func(c *gin.Context) {
		serve(c, Staticjsruntimemain5ed86a79js, `"d5d4d3d6f864fb55593f45530f39dbe7"`, "application/javascript; charset=utf-8", true)
	})

	route.GET("/static/js/runtime-main.5ed86a79.js.map", func(c *gin.Context) {
		serve(c, Staticjsruntimemain5ed86a79jsmap, `"5722955299ce6cf67023d3d3c404f49d"`, "application/json", true)
	})
}
