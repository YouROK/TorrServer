package ffmpeg

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestRealProbeOverHTTP проверяет, что ffprobe читает источник по HTTP.
// Так же модуль будет получать файл раздачи во время работы.
func TestRealProbeOverHTTP(t *testing.T) {
	f := setupRealModule(t)
	input := makeTestMedia(t)

	data, err := os.ReadFile(input)
	if err != nil {
		t.Fatalf("read test media: %v", err)
	}

	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		// ServeContent отдаёт Range-запросы, без которых не читается mp4 с moov в конце
		http.ServeContent(w, r, "source.mkv", time.Now(), bytes.NewReader(data))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := f.Probe(ctx, srv.URL+"/source/1")
	if err != nil {
		t.Fatalf("Probe over HTTP: %v", err)
	}
	if !info.HasVideo() || !info.HasAudio() {
		t.Fatalf("expected video and audio over HTTP, got %d streams", len(info.Streams))
	}
	if v := info.Video(); v.Codec != "h264" {
		t.Errorf("video codec = %q, want h264", v.Codec)
	}
	if hits == 0 {
		t.Error("source endpoint was never requested")
	}
}

// TestRealTranscodeFromHTTP проверяет полный путь: источник по HTTP, вывод в поток.
func TestRealTranscodeFromHTTP(t *testing.T) {
	f := setupRealModule(t)
	input := makeTestMedia(t)

	data, err := os.ReadFile(input)
	if err != nil {
		t.Fatalf("read test media: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "source.mkv", time.Now(), bytes.NewReader(data))
	}))
	defer srv.Close()

	o := f.DefaultTestOptions(srv.URL + "/source/1")
	o.Video.MaxWidth, o.Video.MaxHeight = 160, 120
	o.Container = "mp4"

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	j, err := f.Start(ctx, StartRequest{
		UserID:   "owner",
		Hash:     "bbccddeeff0011223344556677889900aabbccdd",
		FileIdx:  0,
		Options:  o,
		Duration: 4 * time.Second,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	total := 0
	buf := make([]byte, 32*1024)
	for {
		n, err := j.Output().Read(buf)
		total += n
		if err != nil {
			break
		}
	}

	select {
	case <-j.Done():
	case <-time.After(30 * time.Second):
		t.Fatal("transcoding did not finish in time")
	}

	if total < 1024 {
		t.Fatalf("read only %d bytes from the transcoded stream", total)
	}
	if j.State() != "done" {
		t.Errorf("state = %s, want done (stderr: %v)", j.State(), j.Stderr())
	}
}

// TestRealProbeOverHTTPRange проверяет, что источник отдаёт Range-запросы.
func TestRealProbeOverHTTPRange(t *testing.T) {
	f := setupRealModule(t)
	input := makeTestMedia(t)

	data, err := os.ReadFile(input)
	if err != nil {
		t.Fatalf("read test media: %v", err)
	}

	var ranged bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "" {
			ranged = true
		}
		http.ServeContent(w, r, "source.mkv", time.Now(), bytes.NewReader(data))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := f.Probe(ctx, srv.URL+"/source/1"); err != nil {
		t.Fatalf("Probe: %v", err)
	}
	if !ranged {
		t.Error("ffprobe did not issue a Range request")
	}
}

// TestRealProbeInvalidHTTP проверяет обработку недоступного источника.
func TestRealProbeInvalidHTTP(t *testing.T) {
	f := setupRealModule(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := f.Probe(ctx, srv.URL+"/missing"); err == nil {
		t.Error("expected an error for a missing source")
	}
}

// TestRealCacheDirCreated проверяет, что каталог транскодирования создаётся.
func TestRealCacheDirCreated(t *testing.T) {
	f := setupRealModule(t)

	dir := f.CacheDir()
	if dir == "" {
		t.Fatal("cache dir is empty")
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		t.Fatalf("cache dir %q was not created: %v", dir, err)
	}
}

// TestRealConfiguredCacheDir проверяет использование своего каталога.
func TestRealConfiguredCacheDir(t *testing.T) {
	path := realFFmpegPath(t)
	if path == "" || testing.Short() {
		t.Skip("ffmpeg binary is not available")
	}

	custom := filepath.Join(t.TempDir(), "custom-cache")
	cfg := DefaultConfig()
	cfg.CacheDir = custom

	f := newWithBinary(cfg, t.TempDir(), path)
	if !f.Enabled() {
		t.Fatalf("module is disabled: %s", f.Reason())
	}
	t.Cleanup(func() {
		f.StopAll()
		f.Wait()
	})

	if f.CacheDir() != custom {
		t.Errorf("cache dir = %q, want %q", f.CacheDir(), custom)
	}
	if st, err := os.Stat(custom); err != nil || !st.IsDir() {
		t.Fatalf("custom cache dir was not created: %v", err)
	}
}
