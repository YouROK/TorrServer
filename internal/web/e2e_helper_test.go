package web

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"silo/internal/ffmpeg"
	"silo/internal/torrent"
	"silo/internal/user"
)

// startTestServer поднимает настоящий HTTP-сервер на свободном порту.
// Транскодирование требует работающего сервера: ffmpeg читает источник по HTTP.
func startTestServer(t *testing.T, s *Server) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	s.setSourceAddr(ln.Addr())

	srv := &http.Server{Handler: s.router}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	})

	return "http://" + ln.Addr().String()
}

// ffmpegPathForTest возвращает путь к ffmpeg для тестов.
func ffmpegPathForTest(t *testing.T) string {
	t.Helper()

	for _, candidate := range []string{
		filepath.Join("..", "..", "dist", "ffmpeg"),
		"/usr/bin/ffmpeg",
	} {
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate
		}
	}
	return ""
}

// newTestModule создаёт модуль транскодирования на найденном ffmpeg.
func newTestModule(t *testing.T) *ffmpeg.FFmpeg {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping transcode test in short mode")
	}
	if ffmpegPathForTest(t) == "" {
		t.Skip("ffmpeg is not available")
	}

	module := ffmpeg.New(ffmpeg.DefaultConfig(), t.TempDir())
	if !module.Enabled() {
		t.Skipf("transcoding is unavailable: %s", module.Reason())
	}
	t.Cleanup(func() {
		module.StopAll()
		module.Wait()
	})
	return module
}

// makeMediaFile создаёт короткий файл для сквозных проверок.
func makeMediaFile(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "source.mkv")
	cmd := exec.Command(ffmpegPathForTest(t),
		"-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc=s=320x240:r=25:d=4",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=4",
		"-c:v", "libx264", "-preset", "ultrafast",
		"-c:a", "aac",
		"-y", path,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create test media: %v (%s)", err, out)
	}
	return path
}

// serveLocalFile подменяет источник так, чтобы ffmpeg читал локальный файл.
// Синтетическая раздача без пиров данные не отдаёт, поэтому для сквозных
// проверок транскодирования файл берётся с диска.
func serveLocalFile(t *testing.T, s *Server, path string) {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open media file: %v", err)
	}
	_ = file.Close()

	s.SetSourceOpener(func(_ *user.User, _ string, _ int) (io.ReadSeekCloser, *torrent.TorrentFileStat, error) {
		f, err := os.Open(path)
		if err != nil {
			return nil, nil, err
		}
		stat, err := f.Stat()
		if err != nil {
			_ = f.Close()
			return nil, nil, err
		}
		return f, &torrent.TorrentFileStat{
			Id:     0,
			Path:   path,
			Name:   filepath.Base(path),
			Length: stat.Size(),
		}, nil
	})
}
