package tgbot

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"server/leakcheck"
	"server/settings"
	"server/tgbot/config"
)

func TestStartStopNoGoroutineLeak(t *testing.T) {
	dir := t.TempDir()
	oldPath, oldCfg := settings.Path, config.Cfg
	settings.Path = dir
	t.Cleanup(func() {
		Stop()
		settings.Path, config.Cfg = oldPath, oldCfg
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		method := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		switch method {
		case "getMe":
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": map[string]any{"id": 1, "is_bot": true, "username": "leakbot", "first_name": "L"}})
		case "getUpdates":
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": []any{}})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": true})
		}
	}))
	t.Cleanup(srv.Close)

	buf, _ := json.Marshal(&config.Config{HostTG: srv.URL, WhiteIds: []int64{}, BlackIds: []int64{}})
	if err := os.WriteFile(filepath.Join(dir, "tg.cfg"), buf, 0o600); err != nil {
		t.Fatal(err)
	}

	base := runtime.NumGoroutine()
	if err := Start("123:leak"); err != nil {
		t.Fatal(err)
	}
	Stop()
	leakcheck.AssertNoGoroutineLeak(t, base)
}
