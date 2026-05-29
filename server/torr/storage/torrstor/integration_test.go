package torrstor_test

// Integration test that uses a real anacrolix/torrent.Client wired against
// TorrServer's torrstor.Storage, but with all peer/network IO disabled.
// The point is to reproduce streaming-side bugs (silent EOF on torrent drop,
// behavior with zero peers, etc.) deterministically and without any sockets.

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
	"server/torr/storage/torrstor"
)

// --- testutil-style helpers, copy-pasted from anacrolix/torrent/internal/testutil
// (we can't import the upstream internal package). Builds a tiny single-file
// torrent with an in-memory metainfo. Pieces are NOT placed on disk: we want
// the swarm to "have nothing".

const greetingFileName = "greeting"
const greetingFileContents = "hello,\x00world\n"

func makeTinyTorrent(t *testing.T) (*metainfo.MetaInfo, metainfo.Info) {
	t.Helper()
	info := metainfo.Info{
		Name:        greetingFileName,
		PieceLength: 5, // forces 3 pieces over the 13-byte file
		Length:      int64(len(greetingFileContents)),
	}
	err := info.GeneratePieces(func(metainfo.FileInfo) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(greetingFileContents)), nil
	})
	if err != nil {
		t.Fatalf("GeneratePieces: %v", err)
	}
	mi := &metainfo.MetaInfo{}
	mi.InfoBytes, err = bencode.Marshal(info)
	if err != nil {
		t.Fatalf("Marshal info: %v", err)
	}
	return mi, info
}

func ensureSettingsForIntegration(t *testing.T) {
	t.Helper()
	if settings.BTsets == nil {
		settings.BTsets = &settings.BTSets{
			CacheSize:                64 << 20,
			ConnectionsLimit:         25,
			TorrentDisconnectTimeout: 30,
			ReaderReadAHead:          95,
			UseDisk:                  false,
		}
	}
}

// newIsolatedClient builds a torrent.Client that cannot talk to any peers:
// no TCP, no uTP, no DHT, no upload, no port forwarding, listen on an
// ephemeral port. With no trackers in the spec and no DHT, no peer can ever
// be discovered — perfect "0 seeds" environment.
func newIsolatedClient(t *testing.T, st *torrstor.Storage) *torrent.Client {
	t.Helper()
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = t.TempDir()
	cfg.NoDHT = true
	cfg.DisableTCP = true
	cfg.DisableUTP = true
	cfg.NoUpload = true
	cfg.NoDefaultPortForwarding = true
	cfg.ListenPort = 0
	cfg.DefaultStorage = st
	cfg.Seed = false

	cl, err := torrent.NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { cl.Close() })
	return cl
}

// TestReader_DropTorrentDuringRead documents the contract of
// torrstor.Reader.Read when the underlying torrent is dropped while a
// Read is in flight: the upstream anacrolix Reader returns
// "torrent closed" and torrstor surfaces it as-is. http.ServeContent
// will see a real (non-EOF) error after a partial body and abort the
// response — the player stops, but at least there is an error rather
// than a clean truncation.
//
// This is regression documentation, NOT a bug fix gate. The "silent
// EOF" surface lives on different code paths (see
// TestReader_ClosedReader_ReturnsSilentEOF and
// TestReader_TorrentInfoNil_ReturnsSilentEOF below).
func TestReader_DropTorrentDuringRead_DocumentsContract(t *testing.T) {
	ensureSettingsForIntegration(t)

	st := torrstor.NewStorage(64 << 20)
	cl := newIsolatedClient(t, st)

	mi, _ := makeTinyTorrent(t)
	spec := &torrent.TorrentSpec{
		InfoBytes: mi.InfoBytes,
		InfoHash:  mi.HashInfoBytes(),
	}
	tt, _, err := cl.AddTorrentSpec(spec)
	if err != nil {
		t.Fatalf("AddTorrentSpec: %v", err)
	}

	// Wait for info to settle (we provided InfoBytes so this should be
	// nearly immediate).
	select {
	case <-tt.GotInfo():
	case <-time.After(5 * time.Second):
		t.Fatal("torrent did not GotInfo in 5s")
	}

	cache := st.GetCache(tt.InfoHash())
	if cache == nil {
		t.Fatal("cache is nil after GotInfo")
	}
	cache.SetTorrent(tt)

	if len(tt.Files()) == 0 {
		t.Fatal("no files in test torrent")
	}
	file := tt.Files()[0]

	reader := cache.NewReader(file)
	if reader == nil {
		t.Fatal("NewReader returned nil")
	}

	// Drop torrent in another goroutine, then attempt a read.
	// With no peers and no data on disk, Read would otherwise block
	// forever — Drop should unblock it (one way or another).
	go func() {
		time.Sleep(50 * time.Millisecond)
		tt.Drop()
	}()

	done := make(chan struct{})
	var n int
	var rerr error
	go func() {
		buf := make([]byte, 8)
		n, rerr = reader.Read(buf)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Reader.Read did not return within 10s after Drop — would hang the streaming response indefinitely")
	}

	// Document the actual contract: a real error ("torrent closed")
	// surfaces, with n==0. http.ServeContent will abort the response
	// (which is correct behavior — the alternative is silent
	// truncation, see other tests in this file).
	if n != 0 {
		t.Fatalf("expected n==0 after Drop, got n=%d err=%v", n, rerr)
	}
	if rerr == nil {
		t.Fatal("expected non-nil error after Drop, got nil")
	}
	if errors.Is(rerr, io.EOF) {
		t.Fatalf("REGRESSION: Read returned io.EOF after Drop — clients will see a SILENTLY TRUNCATED response. err=%v", rerr)
	}
	t.Logf("OK: Read returned (n=0, err=%q) after Drop — surfaces as a real error to http.ServeContent.", rerr.Error())
}

// (TestReader_ClosedReader_ReturnsSilentEOF was promoted out of racebug
// after fixing bugs #2/#13/#15 — see TestReader_ClosedReader_NoSilentEOF
// at the bottom of this file.)




// TestReader_BlocksWith0Peers verifies the most basic streaming
// invariant: with zero peers and no data on disk, Reader.Read MUST
// block (waiting for data) — it must NEVER return (0, io.EOF) or any
// other "give up" signal that http.ServeContent would interpret as
// end-of-file. If this test fails (Read returns), it means the player
// will see a truncated response purely because of slow/missing peers,
// which is exactly the bug class the user is hitting.
func TestReader_BlocksWith0Peers(t *testing.T) {
	ensureSettingsForIntegration(t)

	st := torrstor.NewStorage(64 << 20)
	cl := newIsolatedClient(t, st)

	mi, _ := makeTinyTorrent(t)
	spec := &torrent.TorrentSpec{
		InfoBytes: mi.InfoBytes,
		InfoHash:  mi.HashInfoBytes(),
	}
	tt, _, err := cl.AddTorrentSpec(spec)
	if err != nil {
		t.Fatalf("AddTorrentSpec: %v", err)
	}
	select {
	case <-tt.GotInfo():
	case <-time.After(5 * time.Second):
		t.Fatal("no GotInfo")
	}
	cache := st.GetCache(tt.InfoHash())
	cache.SetTorrent(tt)

	reader := cache.NewReader(tt.Files()[0])

	done := make(chan struct{})
	var n int
	var rerr error
	go func() {
		buf := make([]byte, 8)
		n, rerr = reader.Read(buf)
		close(done)
	}()

	blocked := false
	select {
	case <-done:
		if errors.Is(rerr, io.EOF) {
			t.Fatalf("BUG: Read returned (n=%d, io.EOF) with 0 peers and no data — "+
				"http.ServeContent will deliver a truncated/empty body to the client.", n)
		}
		t.Logf("Read returned early but with non-EOF error (acceptable, surfaces to client): n=%d err=%v", n, rerr)
	case <-time.After(2 * time.Second):
		blocked = true
		t.Logf("OK: Read blocked for 2s with 0 peers (correct streaming behavior).")
	}

	// Drain the reader goroutine deterministically before test cleanup
	// runs (otherwise client.Close() races r.isClosed; that's bug #13,
	// see TestReader_IsClosed_DataRace).
	if blocked {
		tt.Drop()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Read did not unblock after Drop")
		}
	}
}

// (TestReader_IsClosed_DataRace was promoted out of racebug after fixing
// bugs #13/#14/#15 — see TestReader_CloseUnblocksRead_NoRace below.)

// TestReader_CloseUnblocksRead_NoRace is a regression gate for bugs
// #13 (data race on Reader.isClosed) and #14 (Close() did not unblock
// an in-flight Read; goroutine leaked until the torrent was dropped).
//
// It runs under -race; if the unsynchronized field comes back, the
// race detector fails the build. If Close ever stops cancelling the
// context (or the embedded reader's Close), the 1s timeout fires.
func TestReader_CloseUnblocksRead_NoRace(t *testing.T) {
	ensureSettingsForIntegration(t)

	st := torrstor.NewStorage(64 << 20)
	cl := newIsolatedClient(t, st)

	mi, _ := makeTinyTorrent(t)
	spec := &torrent.TorrentSpec{
		InfoBytes: mi.InfoBytes,
		InfoHash:  mi.HashInfoBytes(),
	}
	tt, _, err := cl.AddTorrentSpec(spec)
	if err != nil {
		t.Fatalf("AddTorrentSpec: %v", err)
	}
	select {
	case <-tt.GotInfo():
	case <-time.After(5 * time.Second):
		t.Fatal("no GotInfo")
	}
	cache := st.GetCache(tt.InfoHash())
	cache.SetTorrent(tt)

	reader := cache.NewReader(tt.Files()[0])

	done := make(chan struct{})
	go func() {
		buf := make([]byte, 8)
		_, _ = reader.Read(buf)
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)
	reader.Close()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		// Best-effort recovery so we don't leak the goroutine into
		// the next test.
		tt.Drop()
		<-done
		t.Fatal("REGRESSION (bug #14): Close() did not unblock in-flight Read within 1s — http.ServeContent goroutines will accumulate on every cancelled stream.")
	}
}

// TestReader_ClosedReader_NoSilentEOF is a regression gate for bug #2
// (and the related #13/#15 races): a Read on a closed reader MUST NOT
// return (0, io.EOF), since http.ServeContent treats EOF as a clean
// end-of-stream and silently truncates the response body.
func TestReader_ClosedReader_NoSilentEOF(t *testing.T) {
	ensureSettingsForIntegration(t)

	st := torrstor.NewStorage(64 << 20)
	cl := newIsolatedClient(t, st)

	mi, _ := makeTinyTorrent(t)
	spec := &torrent.TorrentSpec{
		InfoBytes: mi.InfoBytes,
		InfoHash:  mi.HashInfoBytes(),
	}
	tt, _, err := cl.AddTorrentSpec(spec)
	if err != nil {
		t.Fatalf("AddTorrentSpec: %v", err)
	}
	select {
	case <-tt.GotInfo():
	case <-time.After(5 * time.Second):
		t.Fatal("no GotInfo")
	}
	cache := st.GetCache(tt.InfoHash())
	cache.SetTorrent(tt)

	reader := cache.NewReader(tt.Files()[0])
	reader.Close()
	// Let the cache cleanup goroutine spawned by Reader.Close finish
	// before the test cleanup tears the client down.
	time.Sleep(100 * time.Millisecond)

	buf := make([]byte, 8)
	n, rerr := reader.Read(buf)

	if n == 0 && errors.Is(rerr, io.EOF) {
		t.Fatalf("REGRESSION (bug #2): Read on closed reader returned (0, io.EOF). http.ServeContent will silently truncate the response. err=%v", rerr)
	}
	if rerr == nil {
		t.Fatalf("expected non-nil error after Close, got n=%d err=nil", n)
	}
	t.Logf("OK: Read on closed reader returned (n=%d, err=%v).", n, rerr)
}

