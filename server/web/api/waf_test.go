package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	sets "server/settings"
	"server/web/auth"
	"server/web/waf"

	"github.com/gin-gonic/gin"
)

func withWAFAPITest(t *testing.T) string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	oldPath := sets.Path
	oldReadOnly := sets.ReadOnly
	oldHTTPAuth := sets.HttpAuth
	sets.Path = dir
	sets.ReadOnly = false
	sets.HttpAuth = false
	waf.Load()
	t.Cleanup(func() {
		sets.Path = oldPath
		sets.ReadOnly = oldReadOnly
		sets.HttpAuth = oldHTTPAuth
	})
	return dir
}

func serveWAFAPI(method, body string, middleware ...gin.HandlerFunc) *httptest.ResponseRecorder {
	r := gin.New()
	for _, handler := range middleware {
		r.Use(handler)
	}
	r.GET("/waf", getWAF)
	r.POST("/waf", updateWAF)
	req := httptest.NewRequest(method, "/waf", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestWAFAPIGetAndFullPost(t *testing.T) {
	withWAFAPITest(t)
	w := serveWAFAPI(http.MethodGet, "")
	if w.Code != http.StatusOK {
		t.Fatalf("GET status=%d body=%s", w.Code, w.Body)
	}
	var initial wafResponse
	if err := json.Unmarshal(w.Body.Bytes(), &initial); err != nil {
		t.Fatal(err)
	}
	if initial.Warnings == nil {
		t.Fatal("warnings must be an empty array, not null")
	}
	w = serveWAFAPI(http.MethodPost, `{"whitelist":"10.0.0.0/8","blacklist":"","referers":"example.com"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("POST status=%d body=%s", w.Code, w.Body)
	}
	snap := waf.GetSnapshot()
	if snap.WhitelistText != "10.0.0.0/8" || snap.ReferersText != "example.com" {
		t.Fatalf("snapshot=%+v", snap)
	}
}

func TestWAFAPIRejectsPartialAndMalformedPost(t *testing.T) {
	withWAFAPITest(t)
	if _, err := waf.Update(waf.ListsUpdate{Referers: "keep.example"}); err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{
		`{"whitelist":"10.0.0.0/8"}`,
		`{"whitelist":"","blacklist":"","referers":`,
	} {
		w := serveWAFAPI(http.MethodPost, body)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("body=%q status=%d response=%s", body, w.Code, w.Body)
		}
	}
	if got := waf.GetSnapshot().ReferersText; got != "keep.example" {
		t.Fatalf("partial POST mutated config: %q", got)
	}
}

func TestWAFAPIReadOnlyAndWriteFailure(t *testing.T) {
	dir := withWAFAPITest(t)
	if _, err := waf.Update(waf.ListsUpdate{Referers: "keep.example"}); err != nil {
		t.Fatal(err)
	}
	sets.ReadOnly = true
	w := serveWAFAPI(http.MethodPost, `{"whitelist":"","blacklist":"","referers":""}`)
	if w.Code != http.StatusForbidden {
		t.Fatalf("read-only status=%d body=%s", w.Code, w.Body)
	}

	sets.ReadOnly = false
	sets.Path = filepath.Join(dir, "missing", "directory")
	w = serveWAFAPI(http.MethodPost, `{"whitelist":"","blacklist":"","referers":""}`)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("write-failure status=%d body=%s", w.Code, w.Body)
	}
	if got := waf.GetSnapshot().ReferersText; got != "keep.example" {
		t.Fatalf("failed write mutated in-memory config: %q", got)
	}
}

func TestWAFAPICheckAuthRejectsMissingCredentials(t *testing.T) {
	withWAFAPITest(t)
	sets.HttpAuth = true
	w := serveWAFAPI(http.MethodGet, "", auth.CheckAuth())
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", w.Code, w.Body)
	}
}
