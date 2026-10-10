package ffmpeg

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"silo/internal/ffmpeg/args"
	"silo/internal/ffmpeg/profile"
)

// TestRealProfileResolveAndTranscode проверяет связку профиля с запуском процесса.
func TestRealProfileResolveAndTranscode(t *testing.T) {
	f := setupRealModule(t)
	input := makeTestMedia(t)

	data, err := os.ReadFile(input)
	if err != nil {
		t.Fatalf("read test media: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "source.mkv", time.Now(), newReaderAt(data))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	info, err := f.Probe(ctx, srv.URL+"/source/1")
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}

	// Проверяется выдача сплошного потока, поэтому протокол задаётся явно
	prof := profile.Default()
	prof.ID = "test-h264"
	prof.Protocol = args.ProtocolProgressive
	prof.Container = "mp4"
	prof.Video.Codec = "h264"
	prof.Video.Encoder = ""
	prof.Video.MaxWidth, prof.Video.MaxHeight = 160, 120
	prof.AllowVideoCopy = false

	options, err := prof.Resolve(srv.URL+"/source/1", info)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if options.Video.Encoder == "" {
		options.Video.Encoder = f.SelectVideoEncoder("h264", "")
	}
	if options.Audio.Encoder == "" {
		options.Audio.Encoder = f.SelectAudioEncoder(options.Audio.Codec)
	}

	j, err := f.Start(ctx, StartRequest{
		UserID:    "owner",
		Hash:      "ccddeeff0011223344556677889900aabbccddee",
		FileIdx:   0,
		ProfileID: prof.ID,
		Options:   options,
		Duration:  secondsToDuration(info.Duration),
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Вычитываем поток целиком, иначе ffmpeg встанет на заполненном канале
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
	case <-time.After(60 * time.Second):
		t.Fatal("transcoding did not finish in time")
	}

	if j.State() != "done" {
		t.Fatalf("state = %s, want done (stderr: %v)", j.State(), j.Stderr())
	}
	if total < 1024 {
		t.Fatalf("transcoded stream is too small: %d bytes", total)
	}
	if snap := j.Snapshot(); snap.OutputSize == 0 {
		t.Error("output size was not reported")
	}
}

// TestRealProfileCopiesMatchingCodec проверяет, что совпадающий кодек копируется.
func TestRealProfileCopiesMatchingCodec(t *testing.T) {
	f := setupRealModule(t)
	input := makeTestMedia(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	info, err := f.Probe(ctx, input)
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	if v := info.Video(); v == nil || v.Codec != "h264" {
		t.Skipf("test media codec is %v, expected h264", v)
	}

	prof := profile.Default()
	prof.ID = "test-copy"
	prof.AllowVideoCopy = true
	prof.Video.Codec = "h264"

	options, err := prof.Resolve(input, info)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if options.Video.Codec != "copy" {
		t.Errorf("video codec = %q, want copy for a matching source", options.Video.Codec)
	}
	if options.Audio.Codec != "copy" {
		t.Errorf("audio codec = %q, want copy for a matching source", options.Audio.Codec)
	}

	dir := t.TempDir()
	out := filepath.Join(dir, "copy.mkv")
	// Проверяется файл целиком, поэтому протокол задаётся явно
	options.Protocol = args.ProtocolProgressive
	options.Container = "mkv"
	options.Output = out

	j, err := f.Start(ctx, StartRequest{
		UserID:   "owner",
		Hash:     "ddeeff0011223344556677889900aabbccddeeff",
		Options:  options,
		Duration: secondsToDuration(info.Duration),
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	select {
	case <-j.Done():
	case <-time.After(60 * time.Second):
		t.Fatal("copy did not finish in time")
	}
	if j.State() != "done" {
		t.Fatalf("state = %s, want done (stderr: %v)", j.State(), j.Stderr())
	}

	outInfo, err := f.Probe(ctx, out)
	if err != nil {
		t.Fatalf("Probe result: %v", err)
	}
	if !outInfo.HasVideo() || !outInfo.HasAudio() {
		t.Error("copied file lost streams")
	}
	if v := outInfo.Video(); v.Width != 320 || v.Height != 240 {
		t.Errorf("resolution changed during copy: %dx%d", v.Width, v.Height)
	}
}

// TestRealHLSSessionPaths проверяет раскладку файлов HLS по каталогу сессии.
func TestRealHLSSessionPaths(t *testing.T) {
	f := setupRealModule(t)
	input := makeTestMedia(t)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	info, err := f.Probe(ctx, input)
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}

	prof := profile.Default()
	prof.ID = "test-hls"
	prof.Protocol = "hls"
	prof.Container = "ts"
	prof.Video.Codec = "h264"
	prof.Video.Encoder = f.SelectVideoEncoder("h264", "")
	prof.Video.MaxWidth, prof.Video.MaxHeight = 160, 120
	prof.AllowVideoCopy = false
	prof.Audio.Encoder = f.SelectAudioEncoder(prof.Audio.Codec)
	prof.HLS.SegmentLength = 2
	prof.HLS.ListSize = 3

	options, err := prof.Resolve(input, info)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	j, err := f.Start(ctx, StartRequest{
		ID:        "hls-session",
		UserID:    "owner",
		Hash:      "eeff0011223344556677889900aabbccddeeff00",
		ProfileID: prof.ID,
		Options:   options,
		Duration:  secondsToDuration(info.Duration),
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	select {
	case <-j.Done():
	case <-time.After(60 * time.Second):
		t.Fatal("hls transcoding did not finish in time")
	}
	if j.State() != "done" {
		t.Fatalf("state = %s, want done (stderr: %v)", j.State(), j.Stderr())
	}

	if j.PlaylistPath == "" {
		t.Fatal("playlist path was not set")
	}
	if _, err := os.Stat(j.PlaylistPath); err != nil {
		t.Fatalf("playlist was not created: %v", err)
	}
	if j.WorkDir != filepath.Dir(j.PlaylistPath) {
		t.Errorf("work dir = %q, want %q", j.WorkDir, filepath.Dir(j.PlaylistPath))
	}

	segments, err := filepath.Glob(filepath.Join(j.WorkDir, "seg*.ts"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(segments) == 0 {
		t.Fatal("no segments were produced")
	}
	if len(segments) > 6 {
		t.Errorf("segments = %d, expected the sliding window to cap them", len(segments))
	}
}
