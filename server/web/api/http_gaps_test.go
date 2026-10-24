package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	sets "server/settings"
)

func TestFileReaderSeekWhence(t *testing.T) {
	f := newFR(1000)
	if n, err := f.Seek(10, io.SeekStart); err != nil || n != 10 {
		t.Fatalf("start %d %v", n, err)
	}
	if n, err := f.Seek(5, io.SeekCurrent); err != nil || n != 15 {
		t.Fatalf("current %d %v", n, err)
	}
	if n, err := f.Seek(-2, io.SeekEnd); err != nil || n != 998 {
		t.Fatalf("end %d %v", n, err)
	}
	n, err := f.Read(make([]byte, 4))
	if err != nil || n != 4 || f.pos != 1002 {
		t.Fatalf("read n=%d pos=%d err=%v", n, f.pos, err)
	}
}

func TestDownloadHonorsRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/download/:size", download)
	req := httptest.NewRequest(http.MethodGet, "/download/1", nil)
	req.Header.Set("Range", "bytes=0-15")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusPartialContent || w.Body.Len() != 16 {
		t.Fatalf("range %d len %d", w.Code, w.Body.Len())
	}
}

func TestShutdownReadOnlyForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	old := sets.ReadOnly
	sets.ReadOnly = true
	t.Cleanup(func() { sets.ReadOnly = old })

	r := gin.New()
	r.GET("/shutdown", shutdown)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/shutdown", nil))
	if w.Code != http.StatusForbidden {
		t.Fatalf("shutdown %d", w.Code)
	}
}

func TestTMDBSettingsUninitialized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	old := sets.BTsets
	sets.BTsets = nil
	t.Cleanup(func() { sets.BTsets = old })

	r := gin.New()
	r.GET("/tmdb/settings", tmdbSettings)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tmdb/settings", nil))
	if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), "Settings not initialized") {
		t.Fatalf("tmdb %d %s", w.Code, w.Body.String())
	}
}

func TestSettingsRejectsFloatIntegers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/settings", settings)
	req := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(`{"action":"set","sets":{"CacheSize":64.5,"ConnectionsLimit":25}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("float sets %d %s", w.Code, w.Body.String())
	}
}

func TestCacheRejectsBadBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/cache", cache)
	for _, body := range []string{`{`, `{"action":"nope","hash":"ab"}`} {
		req := httptest.NewRequest(http.MethodPost, "/cache", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("cache %q %d %s", body, w.Code, w.Body.String())
		}
	}
}

func TestViewedRemRequiresHash(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/viewed", viewed)
	req := httptest.NewRequest(http.MethodPost, "/viewed", strings.NewReader(`{"action":"rem"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("rem %d %s", w.Code, w.Body.String())
	}
}
