package msx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestMSXFilesAndTorrentStamp(t *testing.T) {
	torr.InitApiHelper(torr.NewBTS())
	oldSets := settings.BTsets
	settings.BTsets = &settings.BTSets{}
	t.Cleanup(func() { settings.BTsets = oldSets })
	r := msxRouter(t)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/files", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("files get %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(`{`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad files %d", w.Code)
	}

	file := filepath.Join(settings.Path, "note.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(file)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("file not dir %d %s", w.Code, w.Body.String())
	}

	dir := t.TempDir()
	body, _ = json.Marshal(dir)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("symlink %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/files", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), dir) {
		t.Fatalf("linked %d %s", w.Code, w.Body.String())
	}

	hash := strings.Repeat("ef", 20)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/msx/trn", strings.NewReader(`{"Data":"`+hash+`"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "stamp") {
		t.Fatalf("stamp %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/msx/trn?hash="+hash, nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "action") {
		t.Fatalf("query stamp %d %s", w.Code, w.Body.String())
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "1" {
			http.Error(w, "missing header", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("proxied"))
	}))
	defer upstream.Close()
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/msx/proxy?url="+upstream.URL+"&header=X-Test:1&header=nocolon", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "proxied") {
		t.Fatalf("proxy %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/msx/proxy?url=http://127.0.0.1:1", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("proxy fail %d", w.Code)
	}

	link := filepath.Join(settings.Path, files)
	if err := os.Remove(link); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if err := os.WriteFile(link, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/files", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("files not link %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(`""`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("clear files %d %s", w.Code, w.Body.String())
	}
	missing, _ := json.Marshal(filepath.Join(settings.Path, "no-such-dir"))
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(string(missing)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing dir %d %s", w.Code, w.Body.String())
	}
	blocked := filepath.Join(settings.Path, files)
	if err := os.MkdirAll(filepath.Join(blocked, "inner"), 0o755); err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(`""`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("remove dir %d %s", w.Code, w.Body.String())
	}
}
