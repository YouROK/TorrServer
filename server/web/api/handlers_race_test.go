package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSnapshotRuntimeMemory(t *testing.T) {
	m := snapshotRuntimeMemory(7, 9, 2)
	if m.CacheFilled != 7 || m.CacheCapacity != 9 || m.CacheCount != 2 {
		t.Fatalf("%#v", m)
	}
	if m.NumGoroutine < 1 || m.GOMAXPROCS < 1 || m.Sys == 0 {
		t.Fatalf("%#v", m)
	}
	if m.LiveObjects != m.Mallocs-m.Frees {
		t.Fatalf("live %d mallocs %d frees %d", m.LiveObjects, m.Mallocs, m.Frees)
	}
}

func TestFfpRejectsEmptyHash(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/ffp", nil)
	c.Params = gin.Params{{Key: "hash", Value: ""}, {Key: "id", Value: "1"}}
	ffp(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestFfprobeStatusJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/ffp/status", nil)
	ffprobeStatus(c)
	if w.Code != http.StatusOK {
		t.Fatal(w.Code)
	}
	var body ffprobeStatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
}

func TestHandlersConcurrent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/torrents", torrents)
	r.GET("/ffp/status", ffprobeStatus)

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, body := range []string{"{", `{"action":"nope"}`, `{}`} {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/torrents", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				r.ServeHTTP(w, req)
				if w.Code != http.StatusBadRequest {
					t.Errorf("torrents %d %s", w.Code, w.Body.String())
				}
				_ = snapshotRuntimeMemory(0, 0, 0)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ffp/status", nil))
			if w.Code != http.StatusOK {
				t.Errorf("status %d", w.Code)
			}
		}()
	}
	wg.Wait()
}
