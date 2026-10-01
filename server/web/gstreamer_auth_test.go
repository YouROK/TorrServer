//go:build gst

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"server/settings"
	"server/web/auth"
)

func TestGStreamerRoutesRequireAuth(t *testing.T) {
	old := settings.HttpAuth
	settings.HttpAuth = true
	t.Cleanup(func() { settings.HttpAuth = old })

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(auth.BasicAuth(gin.Accounts{"user": "pass"}))
	setupGStreamerRoutes(r)

	paths := []string{
		"/gst/echo",
		"/gst/remove?hash=abc",
		"/gst/abc/heartbeat",
		"/gst/abc/probe?index=1",
		"/gst/abc/master.m3u8?index=1",
		"/gst/abc/video.m3u8",
		"/gst/abc/init.mp4",
		"/gst/abc/seg/0.m4s",
		"/gst/abc/subs/0.m3u8",
	}
	for _, p := range paths {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, p, nil))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("GET %s without credentials: %d, want 401", p, w.Code)
		}
	}

	// with credentials the request reaches the handler
	req := httptest.NewRequest(http.MethodGet, "/gst/remove?hash=abc", nil)
	req.SetBasicAuth("user", "pass")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusUnauthorized {
		t.Fatalf("GET /gst/remove with credentials: 401, want the handler's response")
	}
}
