package template

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newWebRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RouteWebPages(r)
	return r
}

func get(r *gin.Engine, path, ifNoneMatch string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if ifNoneMatch != "" {
		req.Header.Set("If-None-Match", ifNoneMatch)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestWebPagesCacheControl(t *testing.T) {
	r := newWebRouter()

	var static string
	for _, route := range r.Routes() {
		if strings.HasPrefix(route.Path, "/static/js/main.") && strings.HasSuffix(route.Path, ".js") {
			static = route.Path
		}
	}
	if static == "" {
		t.Fatal("no hashed main bundle route")
	}

	cases := map[string]string{
		"/":                 "no-cache",
		"/index.html":       "no-cache",
		"/site.webmanifest": "no-cache",
		static:              "public, max-age=31536000, immutable",
	}
	for path, want := range cases {
		w := get(r, path, "")
		if w.Code != http.StatusOK {
			t.Errorf("%s: status %d, want 200", path, w.Code)
		}
		if got := w.Header().Get("Cache-Control"); got != want {
			t.Errorf("%s: Cache-Control %q, want %q", path, got, want)
		}
		if etag := w.Header().Get("ETag"); !strings.HasPrefix(etag, `"`) || !strings.HasSuffix(etag, `"`) {
			t.Errorf("%s: ETag %q is not quoted", path, etag)
		}
	}
}

func TestWebPagesRevalidation(t *testing.T) {
	r := newWebRouter()
	etag := get(r, "/", "").Header().Get("ETag")

	for _, inm := range []string{etag, "W/" + etag, `"other", ` + etag, "*"} {
		w := get(r, "/", inm)
		if w.Code != http.StatusNotModified {
			t.Errorf("If-None-Match %q: status %d, want 304", inm, w.Code)
		}
		if w.Body.Len() != 0 {
			t.Errorf("If-None-Match %q: 304 has a body", inm)
		}
	}

	if w := get(r, "/", `"stale"`); w.Code != http.StatusOK || w.Body.Len() == 0 {
		t.Errorf("stale ETag: status %d, body %d bytes, want 200 with the page", w.Code, w.Body.Len())
	}
}
