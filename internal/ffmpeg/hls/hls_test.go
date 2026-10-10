package hls

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseRollingPlaylist(t *testing.T) {
	p, err := Parse(filepath.Join("testdata", "rolling.m3u8"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if p.TargetDuration != 2 {
		t.Errorf("target duration = %d, want 2", p.TargetDuration)
	}
	if p.MediaSequence != 2 {
		t.Errorf("media sequence = %d, want 2", p.MediaSequence)
	}
	if len(p.Segments) != 3 {
		t.Fatalf("segments = %d, want 3", len(p.Segments))
	}
	if p.Segments[0].Name != "seg2.ts" {
		t.Errorf("first segment = %q, want seg2.ts", p.Segments[0].Name)
	}
	if p.Segments[0].Duration != 2 {
		t.Errorf("first duration = %v, want 2", p.Segments[0].Duration)
	}
	if !p.Ended {
		t.Error("playlist must be marked as ended")
	}
	if got := p.LastSegmentIndex(); got != 4 {
		t.Errorf("last index = %d, want 4", got)
	}
}

func TestParseVODPlaylist(t *testing.T) {
	p, err := Parse(filepath.Join("testdata", "vod.m3u8"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if p.TargetDuration != 6 {
		t.Errorf("target duration = %d, want 6", p.TargetDuration)
	}
	if len(p.Segments) != 3 {
		t.Fatalf("segments = %d, want 3", len(p.Segments))
	}
	if p.Segments[2].Duration != 3.5 {
		t.Errorf("last duration = %v, want 3.5", p.Segments[2].Duration)
	}
	if p.LastSegmentIndex() != 2 {
		t.Errorf("last index = %d, want 2", p.LastSegmentIndex())
	}
}

func TestParseMissingPlaylist(t *testing.T) {
	_, err := Parse(filepath.Join(t.TempDir(), "missing.m3u8"))
	if !errors.Is(err, ErrNotReady) {
		t.Errorf("error = %v, want %v", err, ErrNotReady)
	}
}

func TestParseEmptyPlaylist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.m3u8")
	if err := os.WriteFile(path, []byte("#EXTM3U\n#EXT-X-VERSION:3\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Parse(path); !errors.Is(err, ErrNoSegments) {
		t.Errorf("error = %v, want %v", err, ErrNoSegments)
	}
}

func TestRewriteUsesLocalNames(t *testing.T) {
	p := &Playlist{
		TargetDuration: 4,
		MediaSequence:  7,
		Segments: []Segment{
			{Name: "seg7.ts", Duration: 4},
			{Name: "seg8.ts", Duration: 3.5},
		},
	}

	out := p.Rewrite("/api/transcode/hls/session1/segment/", "")

	if !strings.Contains(out, "#EXT-X-TARGETDURATION:4") {
		t.Errorf("target duration is missing:\n%s", out)
	}
	if !strings.Contains(out, "#EXT-X-MEDIA-SEQUENCE:7") {
		t.Errorf("media sequence is missing:\n%s", out)
	}
	if !strings.Contains(out, "#EXTINF:4.000000,") {
		t.Errorf("first duration is missing:\n%s", out)
	}
	if !strings.Contains(out, "/api/transcode/hls/session1/segment/seg7.ts") {
		t.Errorf("segment prefix was not applied:\n%s", out)
	}
	if strings.Contains(out, "#EXT-X-ENDLIST") {
		t.Error("unfinished playlist must not contain ENDLIST")
	}
}

func TestRewriteKeepsEndlist(t *testing.T) {
	p := &Playlist{
		TargetDuration: 2,
		Segments:       []Segment{{Name: "seg0.ts", Duration: 2}},
		Ended:          true,
	}

	if !strings.Contains(p.Rewrite("", ""), "#EXT-X-ENDLIST") {
		t.Error("finished playlist must keep ENDLIST")
	}
}

func TestSegmentLocalName(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "seg0.ts", want: "seg0.ts"},
		{in: "folder/seg1.ts", want: "seg1.ts"},
		{in: "seg2.ts?token=abc", want: "seg2.ts"},
		{in: "seg3.ts#frag", want: "seg3.ts"},
	}

	for _, c := range cases {
		if got := (Segment{Name: c.in}).LocalName(); got != c.want {
			t.Errorf("LocalName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSafeSegmentPath(t *testing.T) {
	base := t.TempDir()

	ok, err := SafeSegmentPath(base, "seg0.ts")
	if err != nil {
		t.Fatalf("SafeSegmentPath: %v", err)
	}
	if filepath.Dir(ok) != base {
		t.Errorf("path %q is outside the base dir %q", ok, base)
	}

	for _, bad := range []string{
		"",
		"../secret.ts",
		"../../etc/passwd",
		"sub/seg0.ts",
		`sub\seg0.ts`,
		"..%2fsecret.ts",
	} {
		if _, err := SafeSegmentPath(base, bad); !errors.Is(err, ErrBadPath) {
			t.Errorf("SafeSegmentPath(%q) must be rejected, got %v", bad, err)
		}
	}
}

func TestWaitForSegmentReadyWhenNextExists(t *testing.T) {
	dir := t.TempDir()
	playlist := filepath.Join(dir, "index.m3u8")

	writeFile(t, filepath.Join(dir, "seg0.ts"), 1024)
	writeFile(t, filepath.Join(dir, "seg1.ts"), 1024)

	playlistContent := "#EXTM3U\n#EXT-X-TARGETDURATION:2\n#EXTINF:2.000000,\nseg0.ts\n#EXTINF:2.000000,\nseg1.ts\n"
	writeFile(t, playlist, int64(len(playlistContent)))
	_ = os.WriteFile(playlist, []byte(playlistContent), 0o644)

	if err := WaitForSegment(playlist, dir, "seg0.ts", 2*time.Second, nil); err != nil {
		t.Fatalf("WaitForSegment: %v", err)
	}
}

func TestWaitForSegmentTimesOut(t *testing.T) {
	dir := t.TempDir()
	playlist := filepath.Join(dir, "index.m3u8")
	_ = os.WriteFile(playlist, []byte("#EXTM3U\n"), 0o644)

	err := WaitForSegment(playlist, dir, "seg5.ts", 300*time.Millisecond, nil)
	if !errors.Is(err, ErrNotReady) {
		t.Errorf("error = %v, want %v", err, ErrNotReady)
	}
}

func TestWaitForSegmentBadName(t *testing.T) {
	dir := t.TempDir()

	err := WaitForSegment("", dir, "../escape.ts", time.Second, nil)
	if !errors.Is(err, ErrBadPath) {
		t.Errorf("error = %v, want %v", err, ErrBadPath)
	}
}

func TestWaitForSegmentStopsWhenDone(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "seg0.ts"), 2048)

	done := make(chan struct{})
	close(done)

	// Последний сегмент готов, даже если следующего нет
	if err := WaitForSegment("", dir, "seg0.ts", 2*time.Second, done); err != nil {
		t.Fatalf("WaitForSegment: %v", err)
	}
}

func TestBuildMaster(t *testing.T) {
	out := BuildMaster("main.m3u8?token=x", 3000000, "avc1.64001f,mp4a.40.2")

	if !strings.HasPrefix(out, "#EXTM3U") {
		t.Errorf("master playlist must start with EXTM3U:\n%s", out)
	}
	if !strings.Contains(out, "BANDWIDTH=3000000") {
		t.Errorf("bandwidth is missing:\n%s", out)
	}
	if !strings.Contains(out, `CODECS="avc1.64001f,mp4a.40.2"`) {
		t.Errorf("codecs are missing:\n%s", out)
	}
	if !strings.Contains(out, "main.m3u8?token=x") {
		t.Errorf("variant uri is missing:\n%s", out)
	}
}

func TestTrailingNumber(t *testing.T) {
	cases := []struct {
		in   string
		want int
		ok   bool
	}{
		{in: "seg0", want: 0, ok: true},
		{in: "seg42", want: 42, ok: true},
		{in: "seg", ok: false},
		{in: "", ok: false},
	}

	for _, c := range cases {
		got, ok := trailingNumber(c.in)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("trailingNumber(%q) = (%d, %v), want (%d, %v)", c.in, got, ok, c.want, c.ok)
		}
	}
}

// writeFile создаёт файл заданного размера.
func writeFile(t *testing.T, path string, size int64) {
	t.Helper()

	data := make([]byte, size)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func TestRewriteAddsSuffix(t *testing.T) {
	p := &Playlist{
		TargetDuration: 2,
		Segments:       []Segment{{Name: "seg0.ts", Duration: 2}},
	}

	out := p.Rewrite("/seg/", "?token=abc")

	if !strings.Contains(out, "/seg/seg0.ts?token=abc") {
		t.Errorf("suffix was not applied:\n%s", out)
	}
}

func TestParseFmp4Playlist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fmp4.m3u8")

	content := `#EXTM3U
#EXT-X-VERSION:7
#EXT-X-TARGETDURATION:2
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-PLAYLIST-TYPE:VOD
#EXT-X-MAP:URI="init.mp4"
#EXTINF:2.000000,
seg0.m4s
#EXTINF:2.000000,
seg1.m4s
#EXT-X-ENDLIST
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if p.Version != 7 {
		t.Errorf("version = %d, want 7", p.Version)
	}
	if p.InitSegment != "init.mp4" {
		t.Errorf("init segment = %q, want init.mp4", p.InitSegment)
	}
	if len(p.Segments) != 2 {
		t.Fatalf("segments = %d, want 2", len(p.Segments))
	}
	if p.Segments[0].Name != "seg0.m4s" {
		t.Errorf("first segment = %q, want seg0.m4s", p.Segments[0].Name)
	}
}

func TestRewriteFmp4IncludesMap(t *testing.T) {
	p := &Playlist{
		Version:        7,
		TargetDuration: 2,
		InitSegment:    "init.mp4",
		Segments:       []Segment{{Name: "seg0.m4s", Duration: 2}},
	}

	out := p.Rewrite("/api/transcode/hls/session/abc/segment/", "?token=x")

	if !strings.Contains(out, "#EXT-X-VERSION:7") {
		t.Errorf("version tag is missing:\n%s", out)
	}
	if !strings.Contains(out, `#EXT-X-MAP:URI="/api/transcode/hls/session/abc/segment/init.mp4?token=x"`) {
		t.Errorf("init segment URI is wrong:\n%s", out)
	}
	if !strings.Contains(out, "seg0.m4s?token=x") {
		t.Errorf("segment suffix is missing:\n%s", out)
	}
}

func TestRewriteWithoutInitSegment(t *testing.T) {
	p := &Playlist{
		TargetDuration: 2,
		Segments:       []Segment{{Name: "seg0.ts", Duration: 2}},
	}

	out := p.Rewrite("/seg/", "")

	if strings.Contains(out, "#EXT-X-MAP") {
		t.Errorf("mpegts playlists must not have EXT-X-MAP:\n%s", out)
	}
	// Без версии в исходнике подставляется третья
	if !strings.Contains(out, "#EXT-X-VERSION:3") {
		t.Errorf("default version is missing:\n%s", out)
	}
}

func TestParseInitSegment(t *testing.T) {
	cases := map[string]string{
		`#EXT-X-MAP:URI="init.mp4"`:                   "init.mp4",
		`#EXT-X-MAP:URI="init-1.mp4",BYTERANGE="1@2"`: "init-1.mp4",
		`#EXT-X-MAP:`: "",
	}

	for in, want := range cases {
		if got := parseInitSegment(in); got != want {
			t.Errorf("parseInitSegment(%q) = %q, want %q", in, got, want)
		}
	}
}
