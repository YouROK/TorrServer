package template

import (
	"crypto/md5"
	"fmt"
	"io/fs"
	"mime"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func init() {
	_ = mime.AddExtensionType(".map", "application/json")
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")
}

func RouteWebPages(route gin.IRouter) {
	const indexHTML = "pages/index.html"
	const indexMIME = "text/html; charset=utf-8"

	// Serve / and /index.html (Workbox precache / navigateFallback requests the latter).
	serveIndex := func(c *gin.Context) {
		serveEmbedded(c, indexHTML, indexMIME)
	}
	route.GET("/", serveIndex)
	route.GET("/index.html", serveIndex)

	// Walk embed.FS and register an explicit route for every file.
	// Explicit routes avoid catch-all wildcard conflicts with other gin routes.
	_ = fs.WalkDir(pages, "pages", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		urlPath := "/" + strings.TrimPrefix(path, "pages/")
		if urlPath == "/index.html" {
			return nil // already registered above
		}
		filePath := path // capture for closure
		mimeType := mimeFor(urlPath)
		route.GET(urlPath, func(c *gin.Context) {
			serveEmbedded(c, filePath, mimeType)
		})
		return nil
	})
}

func serveEmbedded(c *gin.Context, path, contentType string) {
	data, err := pages.ReadFile(path)
	if err != nil {
		c.Status(404)
		return
	}
	writeWithCache(c, path, data, contentType)
}

func writeWithCache(c *gin.Context, path string, data []byte, contentType string) {
	sum := md5.Sum(data)
	etag := fmt.Sprintf(`"%x"`, sum)
	cc := cacheControlFor(path)
	c.Header("Cache-Control", cc)
	// Cloudflare caches by Cache-Control unless told otherwise — SW/index must bypass the edge.
	if strings.HasPrefix(cc, "no-cache") || strings.HasPrefix(cc, "no-store") {
		c.Header("CDN-Cache-Control", "no-store")
		c.Header("Cloudflare-CDN-Cache-Control", "no-store")
	}
	c.Header("ETag", etag)
	c.Data(200, contentType, data)
}

// index.html / SW / manifest must revalidate so releases are not stuck behind caches.
// Hashed Vite assets under /static/ can be immutable.
func cacheControlFor(path string) string {
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(path))
	if base == "index.html" || ext == ".webmanifest" || base == "site.webmanifest" {
		return "no-cache, must-revalidate"
	}
	// Service worker scripts must never be long-cached — stale SW precaches 404 after deploys.
	if base == "sw.js" || strings.HasPrefix(base, "workbox-") {
		return "no-cache, must-revalidate"
	}
	if strings.Contains(path, "/static/") || strings.HasPrefix(path, "pages/static/") {
		return "public, max-age=31536000, immutable"
	}
	// favicons and other root assets — short cache
	return "public, max-age=3600"
}

func mimeFor(path string) string {
	m := mime.TypeByExtension(filepath.Ext(path))
	if m == "" {
		m = "application/octet-stream"
	}
	switch m {
	case "application/javascript", "application/xml":
		m += "; charset=utf-8"
	case "image/x-icon":
		m = "image/vnd.microsoft.icon"
	}
	return m
}
