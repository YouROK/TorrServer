package torrstor

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

// A read waits for data no peer has. When the consumer goes away (the player closed the
// HTTP connection), the read has to end, otherwise the reader keeps its priorities and
// keeps the torrent alive.
func newReaderTestTorrent(t *testing.T) (*Cache, *torrent.Torrent) {
	t.Helper()
	initTestSettings(t)
	stor := NewStorage(1 << 20)

	cfg := torrent.NewDefaultClientConfig()
	cfg.DefaultStorage = stor
	cfg.DataDir = t.TempDir()
	cfg.ListenHost = func(string) string { return "127.0.0.1" }
	cfg.ListenPort = 0
	cfg.NoDHT = true
	cfg.DisableTrackers = true
	cfg.DisableUTP = true
	cfg.DisableIPv6 = true
	cfg.NoDefaultPortForwarding = true
	cl, err := torrent.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cl.Close() })

	infoBytes, err := bencode.Marshal(testInfo())
	if err != nil {
		t.Fatal(err)
	}
	tor, err := cl.AddTorrent(&metainfo.MetaInfo{InfoBytes: infoBytes})
	if err != nil {
		t.Fatal(err)
	}
	<-tor.GotInfo()

	cache := stor.GetCache(tor.InfoHash())
	if cache == nil {
		t.Fatal("no cache for the torrent")
	}
	return cache, tor
}

func TestReaderReadStopsWhenContextIsCanceled(t *testing.T) {
	cache, tor := newReaderTestTorrent(t)
	r := cache.NewReader(tor.Files()[0])
	defer cache.CloseReader(r)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r.SetContext(ctx)

	done := make(chan error, 1)
	go func() {
		_, err := r.Read(make([]byte, 1<<10))
		done <- err
	}()

	select {
	case err := <-done:
		t.Fatalf("Read returned without data: %v", err)
	case <-time.After(300 * time.Millisecond):
	}

	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Read returned %v, want context.Canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Read is still waiting for data after the context was canceled")
	}
}

func TestCloseReaderCancelsBlockedRead(t *testing.T) {
	cache, tor := newReaderTestTorrent(t)
	r := cache.NewReader(tor.Files()[0])
	t.Cleanup(r.Close)
	done := make(chan error, 1)
	go func() {
		_, err := r.Read(make([]byte, 1024))
		done <- err
	}()

	// Wait until Read has been admitted. There are no peers or cached data.
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	deadline := time.After(3 * time.Second)
	for {
		r.mu.Lock()
		reading := r.inFlight > 0
		r.mu.Unlock()
		if reading {
			break
		}
		select {
		case err := <-done:
			t.Fatalf("Read returned before close: %v", err)
		case <-tick.C:
		case <-deadline:
			t.Fatal("Read did not start")
		}
	}

	cache.CloseReader(r)
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Read = %v, want context.Canceled", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Read is still waiting after CloseReader")
	}
	if readers := cache.Readers(); readers != 0 {
		t.Fatalf("registered readers after drain = %d, want 0", readers)
	}
}

type drainingTestReader struct {
	torrent.Reader
	started    chan struct{}
	canceled   chan struct{}
	release    chan struct{}
	closeCount atomic.Int32
}

func (r *drainingTestReader) ReadContext(ctx context.Context, _ []byte) (int, error) {
	close(r.started)
	<-ctx.Done()
	close(r.canceled)
	<-r.release
	return 0, ctx.Err()
}

func (r *drainingTestReader) Close() error {
	r.closeCount.Add(1)
	return r.Reader.Close()
}

func TestCloseReaderKeepsReaderRegisteredWhileDraining(t *testing.T) {
	cache, tor := newReaderTestTorrent(t)
	reader := cache.NewReader(tor.Files()[0])
	backend := &drainingTestReader{
		Reader: reader.backend, started: make(chan struct{}),
		canceled: make(chan struct{}), release: make(chan struct{}),
	}
	reader.backend = backend
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reader.SetContext(ctx)
	releaseRead := sync.OnceFunc(func() { close(backend.release) })
	defer releaseRead()
	t.Cleanup(reader.Close)

	done := make(chan error, 1)
	go func() {
		_, err := reader.Read(make([]byte, 1))
		done <- err
	}()
	select {
	case <-backend.started:
	case <-time.After(3 * time.Second):
		t.Fatal("backend read did not start")
	}

	reader.Close()
	reader.Close()
	select {
	case <-backend.canceled:
	case <-time.After(3 * time.Second):
		t.Fatal("reader close did not cancel the read")
	}
	if ctx.Err() != nil || cache.Readers() != 1 || cache.GetUseReaders() != 1 || backend.closeCount.Load() != 0 {
		t.Fatal("close canceled the caller or removed the reader before I/O drained")
	}
	if n, err := reader.Read(make([]byte, 1)); n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("read after close = %d, %v, want EOF", n, err)
	}
	if _, err := reader.Seek(0, io.SeekStart); !errors.Is(err, io.EOF) {
		t.Fatalf("seek after close = %v, want EOF", err)
	}

	releaseRead()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Read = %v, want context.Canceled", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("read did not drain")
	}
	reader.Close()
	if cache.Readers() != 0 || backend.closeCount.Load() != 1 {
		t.Fatal("reader should be unregistered and the backend closed once")
	}
}

func TestCacheRejectsReaderAfterDrop(t *testing.T) {
	cache, tor := newReaderTestTorrent(t)
	file := tor.Files()[0]
	reader := cache.NewReader(file)
	t.Cleanup(reader.Close)
	opened := make(chan *Reader, 1)
	go func() { opened <- cache.NewReader(file) }()
	tor.Drop()
	select {
	case racingReader := <-opened:
		if racingReader != nil {
			if _, err := racingReader.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
				t.Fatalf("read from a reader opened during Drop = %v, want EOF", err)
			}
			racingReader.Close()
		}
	case <-time.After(3 * time.Second):
		t.Fatal("opening a reader did not finish after Drop")
	}
	if got := cache.NewReader(file); got != nil {
		got.Close()
		t.Fatal("closed cache admitted a reader")
	}
	if _, err := reader.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("read after drop = %v, want EOF", err)
	}
	if cache.Readers() != 0 {
		t.Fatal("cache kept readers registered after Drop")
	}
}
