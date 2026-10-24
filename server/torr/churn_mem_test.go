package torr

import (
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anacrolix/torrent"

	"server/leakcheck"
	"server/settings"
	"server/torr/offlinetest"
)

func TestTorrentChurnHeapAndGoroutines(t *testing.T) {
	offlinetest.Prepare(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "udp://127.0.0.1:1/announce\n")
	}))
	t.Cleanup(srv.Close)
	settings.BTsets.TrackersListURL = srv.URL

	show := offlinetest.WriteShow(t)
	seed := offlinetest.Seed(t, show)
	peer := seed.Peer()
	if peer.Port == 0 {
		t.Fatal("seeder has no port")
	}
	// Seeder and the tracker server stay up for the whole test. Disconnect
	// stops the tracker refresh loop Connect started.
	baseG := runtime.NumGoroutine()
	bt := NewBTS()
	InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatal(err)
	}

	const iters = 100
	var warm, done runtime.MemStats
	for i := 0; i < iters; i++ {
		if err := churnOnce(t, show.Spec, peer); err != nil {
			t.Fatalf("iter %d: %v", i, err)
		}
		if i == 9 || i == iters-1 {
			stats := leakcheck.Heap()
			if i == 9 {
				warm = stats
			} else {
				done = stats
			}
		}
	}

	const objectSlack = uint64(20000)
	if done.HeapObjects > warm.HeapObjects+objectSlack {
		t.Fatalf("HeapObjects %d -> %d across %d add/drop cycles", warm.HeapObjects, done.HeapObjects, iters-10)
	}
	const heapSlack = uint64(24 << 20)
	if done.HeapAlloc > warm.HeapAlloc+heapSlack {
		t.Fatalf("HeapAlloc %d -> %d", warm.HeapAlloc, done.HeapAlloc)
	}
	if n := len(bt.ListTorrents()); n != 0 {
		t.Fatalf("BTServer still tracks %d torrents", n)
	}

	connected := runtime.NumGoroutine()
	bt.Disconnect()
	leakcheck.AssertNoGoroutineLeak(t, baseG)
	if connected < baseG {
		t.Fatalf("connected goroutines %d < baseline %d", connected, baseG)
	}
}

func churnOnce(t *testing.T, spec *torrent.TorrentSpec, peer torrent.Peer) error {
	t.Helper()
	tor, err := AddTorrent(spec, "Show", "", "", "tv")
	if err != nil {
		return err
	}
	if !tor.GotInfo() {
		return io.ErrUnexpectedEOF
	}
	tor.AddPeers([]torrent.Peer{peer})

	var file *torrent.File
	for _, f := range tor.Files() {
		if strings.HasSuffix(f.Path(), "episode1.mkv") {
			file = f
			break
		}
	}
	if file == nil {
		t.Fatal("no mkv")
	}

	var wg sync.WaitGroup
	errc := make(chan error, 5)
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := tor.NewReader(file)
			if r == nil {
				errc <- io.ErrClosedPipe
				return
			}
			defer tor.CloseReader(r)
			r.SetResponsive()
			buf := make([]byte, 1024)
			if _, err := io.ReadFull(r, buf); err != nil {
				errc <- err
				return
			}
			if _, err := r.Seek(100, io.SeekStart); err != nil {
				errc <- err
				return
			}
			if _, err := io.ReadFull(r, buf); err != nil {
				errc <- err
			}
		}()
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("readers stuck")
	}
	close(errc)
	for err := range errc {
		if err != nil {
			return err
		}
	}

	id := 1
	for _, f := range tor.Status().FileStats {
		if strings.HasSuffix(f.Path, "episode1.mkv") {
			id = f.Id
		}
	}
	tor.Preload(id, 8<<10)
	if cache := tor.GetCache(); cache != nil && cache.Readers() != 0 {
		t.Fatalf("readers left after preload: %d", cache.Readers())
	}
	RemTorrent(tor.Hash().HexString())
	if GetTorrent(tor.Hash().HexString()) != nil {
		t.Fatal("dropped torrent still in BTServer")
	}
	return nil
}
