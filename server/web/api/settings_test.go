package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	sets "server/settings"
	"server/torr"
)

// "set" without "sets" used to pass nil to torr.SetSettings and panic.
func TestSettingsSetWithoutSetsIsRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/settings", settings)

	req := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(`{"action":"set"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), `"error"`) {
		t.Errorf("no error in body: %q", w.Body.String())
	}
}

func TestSettingsTrackTimecode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/settings", settings)

	dir, err := os.MkdirTemp("", "ts-settings-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sets.CloseDB()
		os.RemoveAll(dir)
	})
	if sets.Args == nil {
		sets.Args = &sets.ExecArgs{}
	}
	sets.Path = dir
	if err := sets.InitSets(false, false); err != nil {
		t.Fatal(err)
	}

	bt := torr.NewBTS()
	torr.InitApiHelper(bt)

	body := `{"action":"set","sets":{"CacheSize":67108864,"TrackTimecode":true}}`
	req := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("set status = %d, want %d", w.Code, http.StatusOK)
	}

	if !sets.BTsets.TrackTimecode {
		t.Errorf("BTsets.TrackTimecode = false, want true")
	}

	reqGet := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(`{"action":"get"}`))
	reqGet.Header.Set("Content-Type", "application/json")
	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", wGet.Code, http.StatusOK)
	}

	var res sets.BTSets
	if err := json.Unmarshal(wGet.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !res.TrackTimecode {
		t.Errorf("res.TrackTimecode = false, want true")
	}
}
