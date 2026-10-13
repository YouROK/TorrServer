//go:build gst

package gstreamer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestFetchExternalSubtitles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("track") {
		case "":
			fmt.Fprint(w, `[{"index":7,"name":"ep01.rus.srt","title":"Exmel - ep01.rus.srt","lang":"ru","format":"srt"},{"index":9,"name":"ep01.eng.ass","lang":"en","format":"ass"}]`)
		case "7":
			if r.URL.Query().Get("format") != "vtt" {
				http.Error(w, "vtt expected", http.StatusBadRequest)
				return
			}
			fmt.Fprint(w, "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nПривет\n\n")
		default:
			http.Error(w, "boom", http.StatusGatewayTimeout)
		}
	}))
	defer server.Close()

	tracks, failed, err := fetchExternalSubtitles(context.Background(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if failed != 1 {
		t.Fatalf("failed = %d, want 1 (the second track returns 504)", failed)
	}
	if len(tracks) != 1 {
		t.Fatalf("got %d tracks, want 1: %+v", len(tracks), tracks)
	}
	track := tracks[0]
	if track.Index != externalSubtitleBase || track.Name != "Exmel - ep01.rus.srt" || track.Language != "ru" {
		t.Fatalf("unexpected track: %+v", track)
	}

	value, ready, _ := track.store.renderVTT(0, 6_000_000_000)
	if !ready {
		t.Fatal("external subtitle range is not ready")
	}
	if !strings.Contains(value, "00:00:01.000 --> 00:00:02.000\nПривет") {
		t.Fatalf("rendered VTT does not contain the cue: %q", value)
	}
}

func TestFetchExternalSubtitlesListError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	if _, _, err := fetchExternalSubtitles(context.Background(), server.URL); err == nil {
		t.Fatal("expected an error when the list request fails")
	}
}

func TestMasterPlaylistListsExternalSubtitles(t *testing.T) {
	task := &Task{
		ID:     "hash",
		Config: Config{Subtitles: true}.normalized(),
		Probe: ProbeInfo{
			DurationNS: int64(time.Minute),
			Tracks:     []TrackInfo{{Type: "video", CapsName: "video/x-h264", Width: 1920, Height: 1080}},
		},
	}
	task.setExternalSubtitles([]externalSubtitle{
		{Index: 1000, Name: `ep01 "rus".srt`, Language: "ru", store: newSubtitleStore()},
		{Index: 1001, Name: `ep01 "rus".srt`, Language: "", store: newSubtitleStore()},
	})

	playlist := buildVariantPlaylist(task, 0, 0)
	for _, want := range []string{
		`#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID="subs",NAME="ep01 \"rus\".srt",LANGUAGE="ru"`,
		`URI="/gst/hash/subs/1000.m3u8"`,
		`NAME="ep01 \"rus\".srt (2)",LANGUAGE="und"`,
		`URI="/gst/hash/subs/1001.m3u8"`,
		`SUBTITLES="subs"`,
	} {
		if !strings.Contains(playlist, want) {
			t.Errorf("master playlist does not contain %q:\n%s", want, playlist)
		}
	}
}

func TestMasterPlaylistWithoutExternalSubtitles(t *testing.T) {
	task := &Task{
		ID:     "hash",
		Config: Config{Subtitles: true}.normalized(),
		Probe: ProbeInfo{
			DurationNS: int64(time.Minute),
			Tracks:     []TrackInfo{{Type: "video", CapsName: "video/x-h264", Width: 1920, Height: 1080}},
		},
	}

	if playlist := buildVariantPlaylist(task, 0, 0); strings.Contains(playlist, "SUBTITLES") {
		t.Fatalf("master playlist mentions subtitles without any track:\n%s", playlist)
	}
}

func TestExternalSubtitleSegmentIsReady(t *testing.T) {
	task, _ := subtitleTestTask() // 6-second segments, one minute of video
	store := newSubtitleStore()
	store.appendVTT("00:00:07.000 --> 00:00:08.000\nhello\n\n", 0, 0)
	store.advanceVideoReadTo(^uint64(0))
	task.setExternalSubtitles([]externalSubtitle{
		{Index: 1000, Name: "a.srt", Language: "ru", store: store},
	})

	// Segment 1 covers 6s..12s: the cue is rendered relative to the segment start.
	value, ready := task.SubtitleVTT(1000, 1)
	if !ready {
		t.Fatal("external subtitle segment is not ready")
	}
	if !strings.Contains(value, "00:00:01.000 --> 00:00:02.000\nhello") {
		t.Fatalf("segment does not contain the cue: %q", value)
	}

	// Segment 0 (0s..6s) has no cues but is ready too.
	value, ready = task.SubtitleVTT(1000, 0)
	if !ready || strings.Contains(value, "-->") {
		t.Fatalf("segment 0: ready=%v value=%q", ready, value)
	}

	// Embedded tracks are still served from their own stores.
	if task.subtitleStore(2) == nil {
		t.Fatal("embedded subtitle store was lost")
	}
}

func TestFetchExternalSubtitlesPermanentErrorsAreNotRetried(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("track") {
		case "":
			fmt.Fprint(w, `[{"index":1,"name":"a.ass"},{"index":2,"name":"b.srt"},{"index":3,"name":"c.srt"}]`)
		case "1":
			http.Error(w, "no dialogue lines found", http.StatusUnprocessableEntity)
		case "2":
			http.Error(w, "not found", http.StatusNotFound)
		default:
			http.Error(w, "slow down", http.StatusTooManyRequests)
		}
	}))
	defer server.Close()

	tracks, failed, err := fetchExternalSubtitles(context.Background(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 0 {
		t.Fatalf("got %d tracks, want 0", len(tracks))
	}
	// 422 and 404 are permanent; only 429 is worth a retry.
	if failed != 1 {
		t.Fatalf("failed = %d, want 1 (only the 429 track)", failed)
	}
}

func TestFetchExternalSubtitlesTooLargeIsSkipped(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("track") == "" {
			fmt.Fprint(w, `[{"index":1,"name":"huge.srt"}]`)
			return
		}
		// One byte over the limit: must be rejected, not silently cut.
		fmt.Fprint(w, "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\n")
		fmt.Fprint(w, strings.Repeat("a", maxExternalSubtitleSize))
	}))
	defer server.Close()

	tracks, failed, err := fetchExternalSubtitles(context.Background(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 0 || failed != 0 {
		t.Fatalf("tracks = %d, failed = %d; want an oversized file skipped without a retry", len(tracks), failed)
	}
}

func TestLoadExternalSubtitlesStopsRetrying(t *testing.T) {
	var listRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("track") == "" {
			listRequests.Add(1)
			fmt.Fprint(w, `[{"index":1,"name":"a.srt"}]`)
			return
		}
		http.Error(w, "boom", http.StatusGatewayTimeout)
	}))
	defer server.Close()

	task := &Task{ID: "hash", FileID: "1", Config: Config{Subtitles: true}.normalized()}
	for i := 0; i < maxExternalSubtitleAttempts+2; i++ {
		task.loadExternalSubtitles(context.Background(), server.URL)
	}
	if got := listRequests.Load(); got != maxExternalSubtitleAttempts {
		t.Fatalf("list requested %d times, want %d", got, maxExternalSubtitleAttempts)
	}
}

func TestLoadExternalSubtitlesPermanentErrorLoadsOnce(t *testing.T) {
	var listRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("track") == "" {
			listRequests.Add(1)
			fmt.Fprint(w, `[{"index":1,"name":"signs.ass"}]`)
			return
		}
		http.Error(w, "no dialogue lines found", http.StatusUnprocessableEntity)
	}))
	defer server.Close()

	task := &Task{ID: "hash", FileID: "1", Config: Config{Subtitles: true}.normalized()}
	for i := 0; i < 3; i++ {
		task.loadExternalSubtitles(context.Background(), server.URL)
	}
	if got := listRequests.Load(); got != 1 {
		t.Fatalf("list requested %d times, want 1", got)
	}
	if len(task.externalSubtitles()) != 0 {
		t.Fatalf("unexpected tracks: %+v", task.externalSubtitles())
	}
}

func TestExternalSubtitleNamesAreUniqueWithEmbeddedTracks(t *testing.T) {
	task := &Task{
		ID:     "hash",
		Config: Config{Subtitles: true}.normalized(),
		Probe: ProbeInfo{
			Tracks: []TrackInfo{
				{Type: "subtitle", Codec: "subrip", Index: 2},
				{Type: "subtitle", Codec: "subrip", Index: 3, Title: "Russian"},
			},
		},
	}
	task.setExternalSubtitles([]externalSubtitle{
		{Index: 1000, Name: "Subtitle 2", store: newSubtitleStore()},
		{Index: 1001, Name: "Russian", store: newSubtitleStore()},
		{Index: 1002, Name: "Russian", store: newSubtitleStore()},
	})

	want := []string{"Subtitle 2 (2)", "Russian (2)", "Russian (3)"}
	got := task.externalSubtitles()
	if len(got) != len(want) {
		t.Fatalf("got %d tracks, want %d", len(got), len(want))
	}
	for i, track := range got {
		if track.Name != want[i] {
			t.Errorf("track %d: name = %q, want %q", i, track.Name, want[i])
		}
	}
}
