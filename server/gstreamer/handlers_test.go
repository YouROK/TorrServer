//go:build gst

package gstreamer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type masterInitRunner struct {
	task  *Task
	calls int
}

func (r *masterInitRunner) EnsureInit(context.Context, int, int) error {
	r.calls++
	r.task.initMu.Lock()
	r.task.initMP4 = []byte{1}
	r.task.variant = &HLSVariantInfo{
		Codecs: "hvc1.1.6.L153.B0,mp4a.40.2",
		Width:  3840,
		Height: 2160,
	}
	r.task.initMu.Unlock()
	return nil
}

func (r *masterInitRunner) GetSegment(context.Context, int, int) (Segment, error) {
	return Segment{}, nil
}

func (r *masterInitRunner) Seek(float64) bool { return true }
func (r *masterInitRunner) Frozen()           {}
func (r *masterInitRunner) Dispose()          {}
func (r *masterInitRunner) IsFrozen() bool    { return false }

func TestSetupRouteDoesNotConflict(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	service := NewService(DefaultConfig())
	defer service.Dispose()

	service.SetupRoute(router)

	routes := router.Routes()
	registered := make(map[string]bool, len(routes))
	for _, route := range routes {
		registered[route.Method+" "+route.Path] = true
	}

	for _, path := range []string{
		"/gst/remove",
		"/gst/echo",
		"/gst/:hash/heartbeat",
		"/gst/:hash/probe",
		"/gst/:hash/master.m3u8",
		"/gst/:hash/init.mp4",
		"/gst/:hash/seg/*segment",
	} {
		if !registered["GET "+path] {
			t.Fatalf("route for %s was not registered", path)
		}
	}
}

func TestProbeRouteReturnsCachedProbe(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	service := &Service{
		conf:       Config{}.normalized(),
		tasks:      make(map[string]*Task),
		probeCache: make(map[string]probeCacheEntry),
	}
	service.setCachedProbe("hash", "1", ProbeInfo{
		DurationNS: 60 * int64(time.Second),
		Container:  "Matroska",
		Tracks: []TrackInfo{
			{Type: "video", CapsName: "video/x-h264", Width: 1920, Height: 1080},
		},
	})
	service.SetupRoute(router)

	request := httptest.NewRequest(http.MethodGet, "/gst/hash/probe?index=1", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("probe route status=%d body=%q", response.Code, response.Body.String())
	}

	var probe ProbeInfo
	if err := json.Unmarshal(response.Body.Bytes(), &probe); err != nil {
		t.Fatal(err)
	}
	if probe.Container != "Matroska" || !probe.IsH264() || probe.Video().Width != 1920 {
		t.Fatalf("probe route returned unexpected probe: %+v", probe)
	}
}

func TestHeartbeatReturnsTorrentDetailsResponse(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	service := &Service{
		conf: Config{}.normalized(),
		tasks: map[string]*Task{
			"hash": {
				ID:         "hash",
				lastActive: time.Now().UTC(),
			},
		},
	}
	service.SetupRoute(router)

	request := httptest.NewRequest(http.MethodGet, "/gst/hash/heartbeat", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("heartbeat status=%d body=%q", response.Code, response.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["Hash"] != "hash" {
		t.Fatalf("heartbeat returned unexpected body: %v", body)
	}
}

func TestMasterEnsuresInitBeforeWritingCodecMetadata(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	conf := Config{}.normalized()
	task := &Task{
		ID:              "hash",
		FileID:          "1",
		Audio:           0,
		Config:          conf,
		LastSentSegment: -1,
		lastActive:      time.Now().UTC(),
		Probe: ProbeInfo{
			DurationNS: int64(2 * time.Hour),
			FileSize:   24 * 1024 * 1024 * 1024,
			Tracks: []TrackInfo{
				{Type: "video", CapsName: "video/x-h265", Width: 3840, Height: 2160},
				{Type: "audio", Index: 0, CapsName: "audio/mpeg"},
			},
		},
	}
	runner := &masterInitRunner{task: task}
	task.runner = runner
	service := &Service{
		conf:       conf,
		tasks:      map[string]*Task{"hash": task},
		probeCache: make(map[string]probeCacheEntry),
	}
	service.SetupRoute(router)

	request := httptest.NewRequest(http.MethodGet, "/gst/hash/master.m3u8?index=1&audio=0", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("master status=%d body=%q", response.Code, response.Body.String())
	}
	if runner.calls != 1 {
		t.Fatalf("EnsureInit calls=%d, want 1", runner.calls)
	}
	if !strings.Contains(response.Body.String(), `CODECS="hvc1.1.6.L153.B0,mp4a.40.2"`) {
		t.Fatalf("master did not use codec metadata from init.mp4: %q", response.Body.String())
	}
}

func TestHLSBandwidthDoesNotOverflowForLargeFile(t *testing.T) {
	task := &Task{
		Config: Config{}.normalized(),
		Probe: ProbeInfo{
			FileSize:   24 * 1024 * 1024 * 1024,
			DurationNS: int64(2 * time.Hour),
		},
	}

	bandwidth, average := task.hlsBandwidth()
	const wantAverage int64 = 28_633_116
	const wantBandwidth int64 = 42_949_674
	if average != wantAverage || bandwidth != wantBandwidth {
		t.Fatalf("hlsBandwidth() = (%d, %d), want (%d, %d)", bandwidth, average, wantBandwidth, wantAverage)
	}
}

func TestParseSingleRange(t *testing.T) {
	tests := []struct {
		header string
		total  int64
		start  int64
		end    int64
		ok     bool
	}{
		{header: "bytes=0-9", total: 100, start: 0, end: 9, ok: true},
		{header: "bytes=10-", total: 100, start: 10, end: 99, ok: true},
		{header: "bytes=-5", total: 100, start: 95, end: 99, ok: true},
		{header: "bytes=100-101", total: 100, ok: false},
		{header: "items=0-1", total: 100, ok: false},
		{header: "bytes=0-1,2-3", total: 100, ok: false},
	}

	for _, test := range tests {
		start, end, ok := parseSingleRange(test.header, test.total)
		if ok != test.ok || start != test.start || end != test.end {
			t.Fatalf("parseSingleRange(%q, %d) = (%d, %d, %v), want (%d, %d, %v)",
				test.header, test.total, start, end, ok, test.start, test.end, test.ok)
		}
	}
}

func TestWriteSegmentWritesPartialContentHeaders(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(http.MethodGet, "/segment", nil)
	c.Request.Header.Set("Range", "bytes=2-5")

	if err := writeSegment(c, Segment{Header: []byte("header"), Payloads: [][]byte{[]byte("payload")}}); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusPartialContent {
		t.Fatalf("status=%d, want %d", response.Code, http.StatusPartialContent)
	}
	if got, want := response.Header().Get("Content-Range"), "bytes 2-5/13"; got != want {
		t.Fatalf("Content-Range=%q, want %q", got, want)
	}
	if got, want := response.Header().Get("Content-Length"), "4"; got != want {
		t.Fatalf("Content-Length=%q, want %q", got, want)
	}
	if got, want := response.Body.String(), "ader"; got != want {
		t.Fatalf("body=%q, want %q", got, want)
	}
}

func TestHLSQuotedRemovesControlCharacters(t *testing.T) {
	if got, want := hlsQuoted("line\r\n\"name\""), `line  \"name\"`; got != want {
		t.Fatalf("hlsQuoted()=%q, want %q", got, want)
	}
}

func TestStartSegmentIndex(t *testing.T) {
	tests := []struct {
		seconds        int
		segmentSeconds int
		count          int
		want           int
	}{
		{seconds: 0, segmentSeconds: 6, count: 1000, want: 0},
		{seconds: 600, segmentSeconds: 6, count: 1000, want: 100},
		{seconds: 599, segmentSeconds: 6, count: 1000, want: 99},
		{seconds: 9999, segmentSeconds: 6, count: 100, want: 100},
		{seconds: 600, segmentSeconds: 0, count: 1000, want: 0},
	}

	for _, test := range tests {
		got := startSegmentIndex(test.seconds, test.segmentSeconds, test.count)
		if got != test.want {
			t.Fatalf("startSegmentIndex(%d, %d, %d) = %d, want %d", test.seconds, test.segmentSeconds, test.count, got, test.want)
		}
	}
}

func TestParseSegmentIndexRejectsNegative(t *testing.T) {
	if _, err := parseSegmentIndex("/-1.m4s"); err == nil {
		t.Fatal("negative segment index was accepted")
	}
}

func TestParseSegmentIndexAllowsNonNegative(t *testing.T) {
	tests := map[string]int{
		"/0.m4s": 0,
		"42.m4s": 42,
	}

	for value, want := range tests {
		got, err := parseSegmentIndex(value)
		if err != nil {
			t.Fatalf("parseSegmentIndex(%q) returned error: %v", value, err)
		}
		if got != want {
			t.Fatalf("parseSegmentIndex(%q) = %d, want %d", value, got, want)
		}
	}
}
