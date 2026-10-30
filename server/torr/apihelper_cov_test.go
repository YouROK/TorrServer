package torr

import (
	"bytes"
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"server/settings"
	"server/torr/offlinetest"
)

type plainReadSeeker struct{ *bytes.Reader }

type ctxReadSeeker struct {
	*bytes.Reader
	got context.Context
}

func (c *ctxReadSeeker) ReadContext(ctx context.Context, p []byte) (int, error) {
	c.got = ctx
	return c.Read(p)
}

func TestBindStreamReadContextPassthroughAndWrap(t *testing.T) {
	if got := bindStreamReadContext(context.Background(), nil); got != nil {
		t.Fatalf("nil content = %v", got)
	}
	plain := plainReadSeeker{bytes.NewReader([]byte("abc"))}
	if got := bindStreamReadContext(nil, plain); got != plain { //nolint:staticcheck // nil ctx is the guarded case
		t.Fatal("nil ctx must return content unchanged")
	}
	if got := bindStreamReadContext(context.Background(), plain); got != plain {
		t.Fatal("content without ReadContext must be returned unchanged")
	}

	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, 1)
	src := &ctxReadSeeker{Reader: bytes.NewReader([]byte("abc"))}
	wrapped := bindStreamReadContext(ctx, src)
	if _, ok := wrapped.(streamCtxReader); !ok {
		t.Fatalf("wrapped type = %T", wrapped)
	}
	buf := make([]byte, 3)
	if n, err := wrapped.Read(buf); err != nil || n != 3 {
		t.Fatalf("Read = %d, %v", n, err)
	}
	if src.got != ctx {
		t.Fatal("Read did not pass the request context to ReadContext")
	}
}

func TestReadOnlyGuardsSkipClient(t *testing.T) {
	offlinetest.Prepare(t)
	bt := NewBTS()
	InitApiHelper(bt)
	settings.ReadOnly = true
	settings.BTsets.UseDisk = true
	settings.BTsets.TorrentsSavePath = t.TempDir()
	hash := strings.Repeat("ab", 20)
	keep := filepath.Join(settings.BTsets.TorrentsSavePath, hash)
	if err := os.MkdirAll(keep, 0o755); err != nil {
		t.Fatal(err)
	}
	before := *settings.BTsets

	start := time.Now()
	SetDefSettings()
	SetSettings(&settings.BTSets{CacheSize: 1})
	RemTorrent(hash)
	if time.Since(start) > 500*time.Millisecond {
		t.Fatal("read-only guards must return immediately")
	}
	if bt.client != nil {
		t.Fatal("read-only guard connected the client")
	}
	if settings.BTsets.CacheSize != before.CacheSize || settings.BTsets.UseDisk != before.UseDisk {
		t.Fatal("read-only guard changed BTsets")
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("read-only RemTorrent removed cache dir: %v", err)
	}
}

func TestShutdownEmbeddedCallsStopWithoutExit(t *testing.T) {
	offlinetest.Prepare(t)
	InitApiHelper(NewBTS())
	settings.Embedded = true
	calls := 0
	settings.EmbeddedStop = func() { calls++ }
	Shutdown()
	if calls != 1 {
		t.Fatalf("EmbeddedStop calls = %d, want 1", calls)
	}
	settings.EmbeddedStop = nil
	Shutdown() // nil stop hook is tolerated
}

// One BT connect covers the lookup paths that need a live client.
func TestAPIHelperLookupsAndRemoval(t *testing.T) {
	offlinetest.Prepare(t)
	bt := NewBTS()
	InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(bt.Disconnect)

	show := offlinetest.WriteShow(t)
	hash := show.Spec.InfoHash

	// Unknown hash: neither in memory nor in the DB.
	if tor := GetTorrent(hash.HexString()); tor != nil {
		t.Fatalf("unknown hash returned %v", tor)
	}

	// DB-only row: GetTorrent returns the stored row and loads it in the background.
	row := &Torrent{TorrentSpec: show.Spec, Title: "stored", Category: "tv", Data: "{}"}
	AddTorrentDB(row)
	got := GetTorrent(hash.HexString())
	if got == nil || got.Title != "stored" {
		t.Fatalf("DB-only GetTorrent = %+v", got)
	}
	deadline := time.Now().Add(5 * time.Second)
	for bt.GetTorrent(hash) == nil {
		if time.Now().After(deadline) {
			t.Fatal("background load never registered the torrent")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if mem := GetTorrent(hash.HexString()); mem == nil || mem == got {
		t.Fatalf("in-memory GetTorrent = %p, DB row %p", mem, got)
	}

	// RemTorrent on a hash that is only on disk and in the DB.
	other := offlinetest.WriteShow(t)
	otherHex := other.Spec.InfoHash.HexString()
	AddTorrentDB(&Torrent{TorrentSpec: other.Spec, Title: "gone", Data: "{}"})
	settings.BTsets.UseDisk = true
	settings.BTsets.TorrentsSavePath = t.TempDir()
	dir := filepath.Join(settings.BTsets.TorrentsSavePath, otherHex)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	RemTorrent(otherHex)
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("cache dir not removed: %v", err)
	}
	if GetTorrentDB(other.Spec.InfoHash) != nil {
		t.Fatal("DB row not removed")
	}

	// RemTorrent on the in-memory torrent: close, remove cache dir, drop the row.
	dir = filepath.Join(settings.BTsets.TorrentsSavePath, hash.HexString())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	RemTorrent(hash.HexString())
	if bt.GetTorrent(hash) != nil {
		t.Fatal("torrent still in memory")
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("cache dir not removed: %v", err)
	}
	if GetTorrentDB(hash) != nil {
		t.Fatal("DB row not removed")
	}
	RemTorrent("not-a-hash") // malformed hash is ignored
}

// reconnect retries while the listen port is busy and succeeds once it frees up.
func TestReconnectRetriesUntilPortFrees(t *testing.T) {
	offlinetest.Prepare(t)
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	settings.TorAddr = ln.Addr().String()
	bt := NewBTS()
	InitApiHelper(bt)
	t.Cleanup(bt.Disconnect)

	go func() {
		time.Sleep(1500 * time.Millisecond)
		_ = ln.Close()
	}()
	reconnect()
	bt.mu.RLock()
	connected := bt.client != nil
	bt.mu.RUnlock()
	if !connected {
		t.Fatalf("reconnect left the client nil with port %d freed", port)
	}
}
