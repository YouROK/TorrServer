package template

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIndexHTMLRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RouteWebPages(r)

	for _, path := range []string{"/", "/index.html"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s: status %d, want 200", path, w.Code)
		}
		ct := w.Header().Get("Content-Type")
		if ct != "text/html; charset=utf-8" {
			t.Fatalf("%s: Content-Type %q", path, ct)
		}
		if len(w.Body.Bytes()) == 0 {
			t.Fatalf("%s: empty body", path)
		}
	}
}

func TestCacheControlFor(t *testing.T) {
	wantNoCache := "no-cache, must-revalidate"
	wantImmutable := "public, max-age=31536000, immutable"

	cases := []struct {
		path string
		want string
	}{
		{"pages/index.html", wantNoCache},
		{"pages/sw.js", wantNoCache},
		{"pages/workbox-abc123.js", wantNoCache},
		{"pages/site.webmanifest", wantNoCache},
		{"pages/static/useTranslation-BnNnnMzO.js", wantImmutable},
		{"pages/favicon.ico", "public, max-age=3600"},
	}
	for _, tc := range cases {
		if got := cacheControlFor(tc.path); got != tc.want {
			t.Errorf("cacheControlFor(%q)=%q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestIndexHTMLSetsCDNBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RouteWebPages(r)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if got := w.Header().Get("CDN-Cache-Control"); got != "no-store" {
		t.Fatalf("CDN-Cache-Control=%q, want no-store", got)
	}
	if got := w.Header().Get("Cloudflare-CDN-Cache-Control"); got != "no-store" {
		t.Fatalf("Cloudflare-CDN-Cache-Control=%q, want no-store", got)
	}
}
