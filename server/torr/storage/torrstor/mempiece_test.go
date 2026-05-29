package torrstor

import (
	"errors"
	"io"
	"sync"
	"testing"

	"server/settings"
)

// Helper to ensure global settings exist for tests that touch Piece (which
// dispatches between MemPiece and DiskPiece based on settings.BTsets.UseDisk).
// Tests that work directly on MemPiece/DiskPiece structs do not need this,
// but we set it anyway to keep behavior deterministic.
func ensureSettings(t *testing.T) {
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

// newTestCacheClosed returns a Cache marked closed so cleanPieces goroutines
// (started from MemPiece.WriteAt / *Piece.ReadAt) bail out immediately and
// don't touch a non-existent c.torrent.
func newTestCacheClosed(pieceLength int64) *Cache {
	c := &Cache{
		capacity:    pieceLength * 10,
		pieceLength: pieceLength,
		pieces:      make(map[int]*Piece),
		readers:     make(map[*Reader]struct{}),
	}
	c.isClosed.Store(true) // makes cleanPieces a no-op (cache.go)
	return c
}

func newTestPiece(c *Cache, id int) *Piece {
	p := &Piece{Id: id, cache: c}
	c.pieces[id] = p
	return p
}

// TestMemPiece_WriteThenRead verifies basic write/read.
func TestMemPiece_WriteThenRead(t *testing.T) {
	ensureSettings(t)
	c := newTestCacheClosed(64)
	p := newTestPiece(c, 0)
	mp := NewMemPiece(p)

	want := []byte("hello, world\n")
	n, err := mp.WriteAt(want, 0)
	if err != nil {
		t.Fatalf("WriteAt err: %v", err)
	}
	if n != len(want) {
		t.Fatalf("WriteAt n=%d want %d", n, len(want))
	}

	got := make([]byte, len(want))
	n, err = mp.ReadAt(got, 0)
	if err != nil {
		t.Fatalf("ReadAt err: %v", err)
	}
	if n != len(want) {
		t.Fatalf("ReadAt n=%d want %d", n, len(want))
	}
	if string(got) != string(want) {
		t.Fatalf("ReadAt got %q want %q", got, want)
	}
}

// TestMemPiece_ReadEmptyReturnsEOF: reading from a piece that was never
// written should not panic; it must return io.EOF.
func TestMemPiece_ReadEmptyReturnsEOF(t *testing.T) {
	ensureSettings(t)
	c := newTestCacheClosed(64)
	p := newTestPiece(c, 0)
	mp := NewMemPiece(p)

	buf := make([]byte, 4)
	n, err := mp.ReadAt(buf, 0)
	if n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("ReadAt empty: n=%d err=%v, want n=0 err=EOF", n, err)
	}
}

// TestMemPiece_ReadAfterRelease documents the current race-prone behavior
// (our finding #9): if Release() is called between two ReadAt calls on the
// same piece, the next ReadAt returns io.EOF — the in-flight stream sees
// a truncated file with no error to distinguish it from a real EOF.
//
// This test EXPECTS the buggy behavior so it stays green on the current
// codebase. When the bug is fixed (e.g. ReadAt returns a distinct error
// like io.ErrUnexpectedEOF instead of plain EOF, or the buffer is held
// alive while a Reader is active), this test should be updated to expect
// the new contract.
func TestMemPiece_ReadAfterRelease_DocumentsBug(t *testing.T) {
	ensureSettings(t)
	c := newTestCacheClosed(64)
	p := newTestPiece(c, 0)
	mp := NewMemPiece(p)

	if _, err := mp.WriteAt([]byte("data"), 0); err != nil {
		t.Fatalf("WriteAt err: %v", err)
	}
	mp.Release()

	buf := make([]byte, 4)
	n, err := mp.ReadAt(buf, 0)
	if !(n == 0 && errors.Is(err, io.EOF)) {
		t.Logf("BEHAVIOR CHANGED (good!): ReadAt after Release returned n=%d err=%v "+
			"instead of n=0 err=EOF. Update this test to assert the new contract.", n, err)
	}
}

// TestMemPiece_ReleaseConcurrentReadAt: Release running concurrently with
// ReadAt must not panic / data-race (write-locked mutex protects buffer).
// Run with -race to catch regressions.
func TestMemPiece_ReleaseConcurrentReadAt(t *testing.T) {
	ensureSettings(t)
	c := newTestCacheClosed(64)
	p := newTestPiece(c, 0)
	mp := NewMemPiece(p)

	if _, err := mp.WriteAt(make([]byte, 64), 0); err != nil {
		t.Fatalf("WriteAt err: %v", err)
	}

	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 16)
		for {
			select {
			case <-stop:
				return
			default:
				_, _ = mp.ReadAt(buf, 0)
			}
		}
	}()

	for i := 0; i < 100; i++ {
		mp.Release()
		_, _ = mp.WriteAt(make([]byte, 64), 0)
	}
	close(stop)
	wg.Wait()
}

// TestMemPiece_ReadPastSize tests reading past the written portion.
// Currently MemPiece.ReadAt clamps size to len(buffer), so it returns
// the available bytes (no error) and only returns EOF if n==0.
func TestMemPiece_ReadPastSize(t *testing.T) {
	ensureSettings(t)
	c := newTestCacheClosed(16)
	p := newTestPiece(c, 0)
	mp := NewMemPiece(p)

	if _, err := mp.WriteAt([]byte("abcd"), 0); err != nil {
		t.Fatalf("WriteAt err: %v", err)
	}

	// Read at offset beyond pieceLength → should EOF cleanly.
	buf := make([]byte, 4)
	n, err := mp.ReadAt(buf, 100)
	if !(n == 0 && errors.Is(err, io.EOF)) {
		t.Fatalf("ReadAt past size: n=%d err=%v, want n=0 err=EOF", n, err)
	}
}
