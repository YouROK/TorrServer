package msx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"server/settings"
	"server/torr"

	"github.com/gin-gonic/gin"
)

func msxRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	oldAuth, oldPath := settings.HttpAuth, settings.Path
	settings.HttpAuth = false
	settings.Path = t.TempDir()
	t.Cleanup(func() {
		settings.HttpAuth = oldAuth
		settings.Path = oldPath
		paramMu.Lock()
		param = "menu:request:interaction:{SERVER}@{PREFIX}" + base + "/start.html"
		paramMu.Unlock()
	})
	r := gin.New()
	SetupRoute(r)
	return r
}

func TestMSXLocalRoutes(t *testing.T) {
	torr.InitApiHelper(torr.NewBTS())
	oldSets := settings.BTsets
	settings.BTsets = &settings.BTSets{}
	t.Cleanup(func() { settings.BTsets = oldSets })
	r := msxRouter(t)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/msx/start.json", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "TorrServer") {
		t.Fatalf("start %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/msx/start.json", strings.NewReader(`"custom"`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("post start %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/msx/start.json", nil))
	if !strings.Contains(w.Body.String(), "custom") {
		t.Fatalf("parameter not stored: %s", w.Body.String())
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/msx/start.json", strings.NewReader(`{`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad json %d", w.Code)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/msx/trn", nil))
	if w.Code != http.StatusOK || strings.TrimSpace(w.Body.String()) != "false" {
		t.Fatalf("trn %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/msx/trn", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "data is not set") {
		t.Fatalf("trn post %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/msx/proxy", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("proxy %d", w.Code)
	}

	if st, sc := trn(strings.Repeat("ab", 20)); st != "" || sc != "" {
		t.Fatalf("trn missing torrent %q %q", st, sc)
	}
}

func TestMSXStartConcurrent(t *testing.T) {
	r := msxRouter(t)
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/msx/start.json", nil))
			if w.Code != http.StatusOK {
				t.Errorf("get %d", w.Code)
			}
		}()
		go func(i int) {
			defer wg.Done()
			body := `"p"`
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/msx/start.json", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("post %d", w.Code)
			}
		}(i)
	}
	wg.Wait()
}
