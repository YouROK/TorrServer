package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"silo/internal/ffmpeg/profile"

	"github.com/gin-gonic/gin"
)

func TestResolveProfileDefaults(t *testing.T) {
	s := &Server{}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/transcode/abc/0", nil)

	prof, err := s.resolveProfile(c)
	if err != nil {
		t.Fatalf("resolveProfile: %v", err)
	}
	if prof.ID != profile.DefaultID {
		t.Errorf("profile id = %q, want %q", prof.ID, profile.DefaultID)
	}
}

func TestResolveProfileQueryOverrides(t *testing.T) {
	s := &Server{}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet,
		"/api/transcode/abc/0?video_codec=hevc&max_height=720&crf=20&audio_channels=2&container=mkv", nil)

	prof, err := s.resolveProfile(c)
	if err != nil {
		t.Fatalf("resolveProfile: %v", err)
	}

	if prof.Video.Codec != "hevc" {
		t.Errorf("video codec = %q, want hevc", prof.Video.Codec)
	}
	if prof.Video.MaxHeight != 720 {
		t.Errorf("max height = %d, want 720", prof.Video.MaxHeight)
	}
	if prof.Video.CRF != 20 {
		t.Errorf("crf = %d, want 20", prof.Video.CRF)
	}
	if prof.Video.BitrateKbps != 0 {
		t.Errorf("bitrate = %d, want it disabled by crf", prof.Video.BitrateKbps)
	}
	if prof.Audio.Channels != 2 {
		t.Errorf("audio channels = %d, want 2", prof.Audio.Channels)
	}
	if prof.Container != "mkv" {
		t.Errorf("container = %q, want mkv", prof.Container)
	}
}

func TestResolveProfileUnknown(t *testing.T) {
	s := &Server{}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/transcode/abc/0?profile=nope", nil)

	if _, err := s.resolveProfile(c); err == nil {
		t.Error("expected an error for an unknown profile")
	}
}

func TestParseSeek(t *testing.T) {
	cases := []struct {
		query string
		want  float64
	}{
		{query: "", want: 0},
		{query: "?t=90", want: 90},
		{query: "?t=12.5", want: 12.5},
		{query: "?start=30", want: 30},
		{query: "?t=-5", want: 0},
		{query: "?t=abc", want: 0},
	}

	for _, c := range cases {
		rec := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/x"+c.query, nil)

		if got := parseSeek(ctx); got != c.want {
			t.Errorf("parseSeek(%q) = %v, want %v", c.query, got, c.want)
		}
	}
}

func TestParseFileIdx(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{{Key: "fileIdx", Value: "3"}}

	idx, ok := parseFileIdx(c)
	if !ok || idx != 3 {
		t.Errorf("parseFileIdx = (%d, %v), want (3, true)", idx, ok)
	}
}

func TestParseFileIdxInvalid(t *testing.T) {
	for _, raw := range []string{"abc", "-1"} {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Params = gin.Params{{Key: "fileIdx", Value: raw}}

		if _, ok := parseFileIdx(c); ok {
			t.Errorf("parseFileIdx(%q) must fail", raw)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	}
}

func TestSecondsToDuration(t *testing.T) {
	cases := []struct {
		in   float64
		want time.Duration
	}{
		{in: 90, want: 90 * time.Second},
		{in: 0, want: 0},
		{in: -5, want: 0},
	}

	for _, c := range cases {
		if got := secondsToDuration(c.in); got != c.want {
			t.Errorf("secondsToDuration(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestIsSourceUnavailable(t *testing.T) {
	cases := []struct {
		msg  string
		want bool
	}{
		{msg: "torrent not found in database", want: true},
		{msg: "Connection to tcp://127.0.0.1:8090 failed: Connection refused", want: true},
		{msg: "Server returned 404 Not Found", want: true},
		{msg: "Invalid data found when processing input", want: true},
		{msg: "no such file or directory", want: true},
		{msg: "some internal failure", want: false},
	}

	for _, c := range cases {
		if got := isSourceUnavailable(errText(c.msg)); got != c.want {
			t.Errorf("isSourceUnavailable(%q) = %v, want %v", c.msg, got, c.want)
		}
	}
	if isSourceUnavailable(nil) {
		t.Error("isSourceUnavailable(nil) must be false")
	}
}

// errText реализует error для проверки разбора сообщений.
type errText string

func (e errText) Error() string { return string(e) }
