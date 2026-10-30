//go:build gst

package gstreamer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func playlistTask(segmentSeconds int, duration time.Duration, cue *CueTimeline) *Task {
	return &Task{
		ID:     "h",
		Config: Config{SegmentSeconds: segmentSeconds}.normalized(),
		Probe:  ProbeInfo{DurationNS: int64(duration)},
		Cue:    cue,
	}
}

func extinfs(playlist string) []string {
	var out []string
	for _, line := range strings.Split(playlist, "\n") {
		if strings.HasPrefix(line, "#EXTINF:") {
			out = append(out, strings.TrimSuffix(strings.TrimPrefix(line, "#EXTINF:"), ","))
		}
	}
	return out
}

func TestBuildTaskPlaylistFixedSegments(t *testing.T) {
	// 25s in 10s segments: 10, 10, 5.
	got := buildTaskPlaylist(playlistTask(10, 25*time.Second, nil), 0, 1)

	for _, want := range []string{
		"#EXTM3U\n#EXT-X-PLAYLIST-TYPE:VOD\n#EXT-X-VERSION:7\n",
		"#EXT-X-TARGETDURATION:10\n",
		"#EXT-X-MEDIA-SEQUENCE:0\n",
		"#EXT-X-MAP:URI=\"init.mp4?audio=1\"\n",
		"seg/0.m4s\n", "seg/1.m4s\n", "seg/2.m4s\n",
		"#EXT-X-ENDLIST\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if durs := extinfs(got); strings.Join(durs, " ") != "10 10 5" {
		t.Errorf("EXTINF = %v, want [10 10 5]", durs)
	}
	if strings.Contains(got, "seg/3.m4s") {
		t.Error("one segment too many")
	}
}

func TestBuildTaskPlaylistExactMultipleHasNoEmptyTail(t *testing.T) {
	got := buildTaskPlaylist(playlistTask(10, 20*time.Second, nil), 0, 0)
	if durs := extinfs(got); strings.Join(durs, " ") != "10 10" {
		t.Fatalf("EXTINF = %v, want [10 10]", durs)
	}
}

func TestBuildTaskPlaylistStartIndexMapsSeekPosition(t *testing.T) {
	task := playlistTask(10, 100*time.Second, nil)
	got := buildTaskPlaylist(task, 3, 2)
	if !strings.Contains(got, "#EXT-X-MEDIA-SEQUENCE:3\n") {
		t.Errorf("media sequence:\n%s", got)
	}
	// The init segment must be fetched for the same seek position.
	if !strings.Contains(got, `init.mp4?audio=2&seconds=30"`) {
		t.Errorf("init map lost the seek offset:\n%s", got)
	}
	if strings.Contains(got, "seg/2.m4s") || !strings.Contains(got, "seg/3.m4s") || !strings.Contains(got, "seg/9.m4s") {
		t.Errorf("segment range wrong:\n%s", got)
	}

	// Out-of-range indexes are clamped instead of producing negative or
	// past-the-end sequences.
	if got := buildTaskPlaylist(task, -5, 0); !strings.Contains(got, "MEDIA-SEQUENCE:0\n") || strings.Contains(got, "seconds=") {
		t.Errorf("negative start not clamped:\n%s", got)
	}
	past := buildTaskPlaylist(task, 999, 0)
	if !strings.Contains(past, "MEDIA-SEQUENCE:10\n") || len(extinfs(past)) != 0 {
		t.Errorf("start past the end not clamped:\n%s", past)
	}
}

func TestBuildTaskPlaylistUnknownDurationAssumes200Minutes(t *testing.T) {
	got := buildTaskPlaylist(playlistTask(6, 0, nil), 0, 0)
	if n := len(extinfs(got)); n != 2000 { // 200*60/6
		t.Fatalf("segments = %d, want 2000", n)
	}
}

func TestBuildTaskPlaylistFollowsCueTimeline(t *testing.T) {
	cue := &CueTimeline{
		Segments: []CueSegment{
			{0, 4_500_000_000},
			{4_500_000_000, 12_000_000_000},
			{12_000_000_000, 13_250_000_000},
		},
		MaxDurationNS: 7_500_000_000,
	}
	task := playlistTask(6, 13*time.Second+250*time.Millisecond, cue)
	got := buildTaskPlaylist(task, 0, 0)
	if durs := extinfs(got); strings.Join(durs, " ") != "4.5 7.5 1.25" {
		t.Errorf("EXTINF = %v", durs)
	}
	if !strings.Contains(got, "#EXT-X-TARGETDURATION:8\n") {
		t.Errorf("target duration must round the longest segment up:\n%s", got)
	}

	got = buildTaskPlaylist(task, 1, 0)
	if !strings.Contains(got, "seconds=4\"") { // 4.5s truncated for the URL
		t.Errorf("cue start offset:\n%s", got)
	}
	if len(extinfs(got)) != 2 {
		t.Errorf("EXTINF count = %d, want 2", len(extinfs(got)))
	}
}

func TestStartIndexForSeconds(t *testing.T) {
	plain := playlistTask(10, 95*time.Second, nil)
	for seconds, want := range map[int]int{-1: 0, 0: 0, 9: 0, 10: 1, 94: 9, 5000: 10} {
		if got := plain.startIndexForSeconds(seconds); got != want {
			t.Errorf("plain seconds=%d index=%d want %d", seconds, got, want)
		}
	}

	cued := playlistTask(10, 20*time.Second, &CueTimeline{Segments: []CueSegment{
		{0, 3_000_000_000}, {3_000_000_000, 9_000_000_000}, {9_000_000_000, 20_000_000_000},
	}})
	for seconds, want := range map[int]int{0: 0, 2: 0, 3: 1, 8: 1, 9: 2, 19: 2, 20: 3, 999: 3} {
		if got := cued.startIndexForSeconds(seconds); got != want {
			t.Errorf("cue seconds=%d index=%d want %d", seconds, got, want)
		}
	}
	if got := cued.segmentStartNS(2); got != 9_000_000_000 {
		t.Errorf("cue start = %d", got)
	}
	if got := plain.segmentStartNS(3); got != 30_000_000_000 {
		t.Errorf("fixed start = %d", got)
	}
	if got := plain.segmentStartNS(-1); got != 0 {
		t.Errorf("negative index start = %d", got)
	}
}

func TestBuildSubtitlePlaylistRewritesSegmentsToVTT(t *testing.T) {
	task := playlistTask(10, 25*time.Second, nil)
	got := buildSubtitlePlaylist(task, 4)
	if strings.Contains(got, "EXT-X-MAP") || strings.Contains(got, ".m4s") {
		t.Errorf("video-only lines survived:\n%s", got)
	}
	for _, want := range []string{"4/0.vtt\n", "4/1.vtt\n", "4/2.vtt\n", "#EXT-X-ENDLIST\n", "#EXT-X-TARGETDURATION:10\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Join(extinfs(got), " ") != "10 10 5" {
		t.Errorf("durations must match the video playlist: %v", extinfs(got))
	}
}

func TestHLSQuotedEscapesAttributeValues(t *testing.T) {
	cases := map[string]string{
		`plain`:        `plain`,
		`He said "hi"`: `He said \"hi\"`,
		`back\slash`:   `back\\slash`,
		"line\nbreak":  "line break",
		"tab\tand\x7f": "tab and ",
		`ёжик «ok»`:    `ёжик «ok»`,
	}
	for in, want := range cases {
		if got := hlsQuoted(in); got != want {
			t.Errorf("hlsQuoted(%q) = %q, want %q", in, got, want)
		}
	}
}

func routerFor(s *Service) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	s.SetupRoute(r)
	return r
}

func get(r http.Handler, target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
	return w
}

func TestVideoPlaylistRoute(t *testing.T) {
	task := playlistTask(10, 40*time.Second, nil)
	task.lastActive = time.Now().UTC()
	task.Audio = 3
	svc := &Service{conf: Config{}.normalized(), tasks: map[string]*Task{"h": task}}
	r := routerFor(svc)

	if w := get(r, "/gst/missing/video.m3u8"); w.Code != http.StatusNotFound {
		t.Errorf("unknown task = %d", w.Code)
	}

	w := get(r, "/gst/h/video.m3u8")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/vnd.apple.mpegurl") {
		t.Errorf("content type = %q", ct)
	}
	if cc := w.Header().Get("Cache-Control"); !strings.Contains(cc, "no-store") {
		t.Errorf("playlist must not be cached, Cache-Control = %q", cc)
	}
	if !strings.Contains(w.Body.String(), "init.mp4?audio=3") {
		t.Errorf("task audio not used as the default:\n%s", w.Body.String())
	}

	w = get(r, "/gst/h/video.m3u8?audio=1&seconds=20")
	if !strings.Contains(w.Body.String(), "MEDIA-SEQUENCE:2\n") || !strings.Contains(w.Body.String(), "init.mp4?audio=1&seconds=20") {
		t.Errorf("seek params ignored:\n%s", w.Body.String())
	}
	// A non-numeric query value falls back instead of failing the request.
	if w := get(r, "/gst/h/video.m3u8?seconds=abc"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "MEDIA-SEQUENCE:0\n") {
		t.Errorf("bad seconds: %d", w.Code)
	}
}

func TestSubtitleRoute(t *testing.T) {
	task := playlistTask(10, 25*time.Second, nil)
	task.lastActive = time.Now().UTC()
	svc := &Service{conf: Config{}.normalized(), tasks: map[string]*Task{"h": task}}
	r := routerFor(svc)

	if w := get(r, "/gst/missing/subs/0.m3u8"); w.Code != http.StatusNotFound {
		t.Errorf("unknown task = %d", w.Code)
	}

	w := get(r, "/gst/h/subs/2.m3u8")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "2/1.vtt") {
		t.Fatalf("playlist: %d\n%s", w.Code, w.Body.String())
	}

	for _, bad := range []string{
		"/gst/h/subs/-1.m3u8",
		"/gst/h/subs/x.m3u8",
		"/gst/h/subs/0/1.txt",
		"/gst/h/subs/0",
		"/gst/h/subs/0/1/2.vtt",
		"/gst/h/subs/a/1.vtt",
		"/gst/h/subs/0/b.vtt",
		"/gst/h/subs/0/-1.vtt",
	} {
		if w := get(r, bad); w.Code != http.StatusBadRequest {
			t.Errorf("%s = %d, want 400", bad, w.Code)
		}
	}

	// A track that doesn't exist yields a valid, empty cue file, not an error:
	// the player keeps polling segments of a subtitle rendition.
	w = get(r, "/gst/h/subs/0/1.vtt")
	if w.Code != http.StatusOK || !strings.HasPrefix(w.Body.String(), "WEBVTT\n") {
		t.Fatalf("vtt: %d %q", w.Code, w.Body.String())
	}
	if !strings.HasPrefix(w.Header().Get("Content-Type"), "text/vtt") {
		t.Errorf("content type = %q", w.Header().Get("Content-Type"))
	}
	// Segment 1 starts at 10s: MPEGTS = 10 * 90000.
	if !strings.Contains(w.Body.String(), "MPEGTS:900000") {
		t.Errorf("timestamp map for segment 1: %q", w.Body.String())
	}
}

func TestRemoveRoute(t *testing.T) {
	task, runner := newTrackedTask("h", time.Now().UTC())
	svc := &Service{conf: Config{}.normalized(), tasks: map[string]*Task{"h": task}}
	r := routerFor(svc)

	if w := get(r, "/gst/remove"); w.Code != http.StatusBadRequest {
		t.Errorf("no id = %d, want 400", w.Code)
	}
	if w := get(r, "/gst/remove?hash=nope"); w.Code != http.StatusNotFound {
		t.Errorf("unknown id = %d, want 404", w.Code)
	}
	if runner.disposed.Load() {
		t.Fatal("unrelated removal disposed a task")
	}
	// The task is removed and disposed directly; the route's torrent drop is a
	// no-op when the torrent isn't known.
	if w := get(r, "/gst/remove?id=h"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"success":true`) {
		t.Errorf("remove = %d %s", w.Code, w.Body.String())
	}
	if !runner.disposed.Load() || !task.IsDisposed() {
		t.Error("removed task was not disposed")
	}
	if svc.Get("h") != nil {
		t.Error("task still reachable")
	}
	if w := get(r, "/gst/remove?id=h"); w.Code != http.StatusNotFound {
		t.Errorf("second remove = %d, want 404", w.Code)
	}
}

func TestCleanupInactiveFreezesThenRemoves(t *testing.T) {
	now := time.Now().UTC()
	conf := Config{InactiveMinutes: 5}.normalized()

	fresh, freshRunner := newTrackedTask("fresh", now.Add(-time.Minute))
	idle, idleRunner := newTrackedTask("idle", now.Add(-10*time.Minute)) // past 5min: freeze
	dead, deadRunner := newTrackedTask("dead", now.Add(-26*time.Minute)) // past 5+20min: remove
	svc := &Service{
		conf:       conf,
		tasks:      map[string]*Task{"fresh": fresh, "idle": idle, "dead": dead},
		probeCache: map[string]probeCacheEntry{},
	}
	svc.probeCache["old"] = probeCacheEntry{expiresAt: now.Add(-time.Second)}
	svc.probeCache["live"] = probeCacheEntry{expiresAt: now.Add(time.Hour)}

	svc.cleanupInactive()

	if freshRunner.frozen.Load() || freshRunner.disposed.Load() {
		t.Error("recently used task was touched")
	}
	if !idleRunner.frozen.Load() || idleRunner.disposed.Load() {
		t.Errorf("idle task: frozen=%v disposed=%v, want frozen only", idleRunner.frozen.Load(), idleRunner.disposed.Load())
	}
	if !deadRunner.disposed.Load() {
		t.Error("long-dead task was not disposed")
	}
	if _, ok := svc.tasks["dead"]; ok {
		t.Error("long-dead task still registered")
	}
	if _, ok := svc.tasks["idle"]; !ok {
		t.Error("frozen task must stay registered so it can resume")
	}
	if _, ok := svc.probeCache["old"]; ok {
		t.Error("expired probe cache entry kept")
	}
	if _, ok := svc.probeCache["live"]; !ok {
		t.Error("live probe cache entry dropped")
	}

	// A disposed service does nothing.
	svc.Dispose()
	svc.cleanupInactive()
}
