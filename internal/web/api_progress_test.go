package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"silo/internal/config"
	"silo/internal/user"

	"github.com/gin-gonic/gin"
)

func postProgress(t *testing.T, s *Server, u *user.User, hash, idx, body string) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/torrents/"+hash+"/files/"+idx+"/progress", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "hash", Value: hash}, {Key: "idx", Value: idx}}
	c.Set("user", u)

	s.handleSetFileProgress(c)
	return rec
}

func decodeTracked(t *testing.T, rec *httptest.ResponseRecorder) bool {
	t.Helper()

	var resp struct {
		Tracked bool `json:"tracked"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp.Tracked
}

func TestHandleSetFileProgress(t *testing.T) {
	s, owner := setupAddTestEnvWith(t, func(cfg *config.Config) {
		cfg.Library.TrackProgress = true
	})

	hash := "0123456789abcdef0123456789abcdef01234567"
	if err := s.userSvc.AddTorrent(owner, hash, "Movie", "", ""); err != nil {
		t.Fatalf("AddTorrent: %v", err)
	}

	rec := postProgress(t, s, owner, hash, "2", `{"position": 75.5, "duration": 1800}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !decodeTracked(t, rec) {
		t.Error("tracked = false with tracking on")
	}

	ut, err := s.userSvc.GetUserTorrent(owner.ID, hash)
	if err != nil {
		t.Fatalf("GetUserTorrent: %v", err)
	}
	if got := ut.Progress[2]; got.Position != 75.5 || got.Duration != 1800 {
		t.Errorf("Progress[2] = %+v, want {75.5 1800}", got)
	}

	for name, body := range map[string]string{
		"no position":       `{"duration": 10}`,
		"negative position": `{"position": -5}`,
		"broken json":       `{"position":`,
	} {
		if rec := postProgress(t, s, owner, hash, "2", body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", name, rec.Code)
		}
	}

	if rec := postProgress(t, s, owner, hash, "x", `{"position": 1}`); rec.Code != http.StatusBadRequest {
		t.Errorf("bad index: expected 400, got %d", rec.Code)
	}
}

func TestHandleSetFileProgressDisabled(t *testing.T) {
	s, owner := setupAddTestEnv(t)

	hash := "0123456789abcdef0123456789abcdef01234567"
	if err := s.userSvc.AddTorrent(owner, hash, "Movie", "", ""); err != nil {
		t.Fatalf("AddTorrent: %v", err)
	}

	rec := postProgress(t, s, owner, hash, "0", `{"position": 30}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if decodeTracked(t, rec) {
		t.Error("tracked = true with tracking off")
	}

	ut, _ := s.userSvc.GetUserTorrent(owner.ID, hash)
	if !ut.IsFileViewed(0) {
		t.Error("file 0 must be viewed")
	}
	if len(ut.Progress) != 0 {
		t.Errorf("Progress = %v, want empty", ut.Progress)
	}
}
