package torr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
	"server/torr/state"
	"server/torr/storage/torrstor"
)

func TestStreamCanceledRequestReleasesReader(t *testing.T) {
	// The settings database singletons cannot be reopened after CloseDB.
	const isolatedEnv = "TORRSERVER_STREAM_CANCEL_TEST_ISOLATED"
	if os.Getenv(isolatedEnv) != "1" {
		ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestStreamCanceledRequestReleasesReader$")
		cmd.Env = append(os.Environ(), isolatedEnv+"=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("isolated stream test: %v\n%s", err, output)
		}
		return
	}
	dbDir := t.TempDir()
	settings.Path = dbDir
	if err := settings.InitSets(false, false); err != nil {
		t.Fatal(err)
	}
	defer settings.CloseDB()

	storage := torrstor.NewStorage(1 << 20)
	cfg := torrent.NewDefaultClientConfig()
	cfg.DefaultStorage = storage
	cfg.DataDir = dbDir
	cfg.ListenHost = func(string) string { return "127.0.0.1" }
	cfg.ListenPort = 0
	cfg.NoDefaultPortForwarding = true
	cfg.NoDHT = true
	cfg.DisableTrackers = true
	cfg.DisablePEX = true
	cfg.DisableUTP = true
	cfg.DisableIPv6 = true
	client, err := torrent.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	const pieceSize = 16 << 10
	info := metainfo.Info{Name: "missing.mkv", PieceLength: pieceSize, Length: pieceSize, Pieces: make([]byte, 20)}
	engine, err := client.AddTorrent(&metainfo.MetaInfo{InfoBytes: bencode.MustMarshal(info)})
	if err != nil {
		t.Fatal(err)
	}
	cache := storage.GetCache(engine.InfoHash())
	if cache == nil || engine.PieceState(0).Complete {
		t.Fatal("expected a torrent cache with a missing piece")
	}
	cache.SetTorrent(engine)
	tor := &Torrent{Torrent: engine, Stat: state.TorrentWorking, cache: cache}

	ctx, cancel := context.WithCancel(t.Context())
	req := httptest.NewRequest(http.MethodGet, "/stream", nil).WithContext(ctx)
	req.Header.Set("Range", "bytes=0-")
	resp := &streamCancelRecorder{ResponseRecorder: httptest.NewRecorder(), headers: make(chan int, 1)}
	done := make(chan struct{})
	var streamErr error
	go func() {
		defer close(done)
		streamErr = tor.Stream(1, req, resp)
	}()
	defer func() {
		cancel()
		engine.Drop()
		<-done
	}()

	select {
	case code := <-resp.headers:
		if code != http.StatusPartialContent || cache.Readers() != 1 {
			t.Fatalf("response status = %d, registered readers = %d, want 206 and 1", code, cache.Readers())
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Stream did not start serving the requested range")
	}
	select {
	case <-done:
		t.Fatalf("Stream returned before cancellation while the piece was missing: %v", streamErr)
	case <-time.After(50 * time.Millisecond):
	}

	cancel()
	select {
	case <-done:
		if streamErr != nil {
			t.Fatalf("Stream returned an error: %v", streamErr)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Stream still waiting for the missing piece after request cancellation")
	}
	if readers := cache.Readers(); readers != 0 {
		t.Fatalf("registered readers after Stream returned = %d, want 0", readers)
	}
}

type streamCancelRecorder struct {
	*httptest.ResponseRecorder
	headers chan int
}

func (r *streamCancelRecorder) WriteHeader(code int) {
	r.ResponseRecorder.WriteHeader(code)
	r.headers <- code
}
