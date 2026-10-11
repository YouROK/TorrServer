package template

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// serve writes an embedded web asset. Files under /static/ carry a content
// hash in their name, so a browser may keep them for good. Everything else
// (index.html, the manifest, icons) keeps its name across releases and must be
// revalidated, otherwise an upgraded server keeps showing the old web UI until
// the browser cache is cleared.
func serve(c *gin.Context, data []byte, etag, mime string, immutable bool) {
	if immutable {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		c.Header("Cache-Control", "no-cache")
	}
	c.Header("ETag", etag)
	if etagMatches(c.GetHeader("If-None-Match"), etag) {
		c.Status(http.StatusNotModified)
		return
	}
	c.Data(http.StatusOK, mime, data)
}

func etagMatches(header, etag string) bool {
	for _, tag := range strings.Split(header, ",") {
		tag = strings.TrimSpace(tag)
		if tag == "*" || strings.TrimPrefix(tag, "W/") == etag {
			return true
		}
	}
	return false
}
