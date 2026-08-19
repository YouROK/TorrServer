package torrstor

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
)

// newDiskTestCache builds a cache pointing at a tmp dir with a fake hash so
// DiskPiece writes/reads its files there. isClosed is true so cleanPieces
// goroutines are no-ops (we don't have a real torrent attached).
func newDiskTestCache(t *testing.T, pieceLength int64) *Cache {
	t.Helper()
	ensureSettings(t)
	tmp := t.TempDir()

	// Configure global settings to point disk storage at the temp dir.
	prevPath := settings.BTsets.TorrentsSavePath
	prevUseDisk := settings.BTsets.UseDisk
	settings.BTsets.TorrentsSavePath = tmp
	settings.BTsets.UseDisk = true
	t.Cleanup(func() {
		settings.BTsets.TorrentsSavePath = prevPath
		settings.BTsets.UseDisk = prevUseDisk
	})

	var h metainfo.Hash
	for i := range h {
		h[i] = byte(i)
	}

	c := &Cache{
		capacity:    pieceLength * 10,
		pieceLength: pieceLength,
		pieces:      make(map[int]*Piece),
		readers:     make(map[*Reader]struct{}),
		hash:        h,
	}
	c.isClosed.Store(true)

	// Pre-create the per-torrent directory like Cache.Init would.
	if err := os.MkdirAll(filepath.Join(tmp, h.HexString()), 0o777); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	return c
}

func TestDiskPiece_WriteThenRead(t *testing.T) {
	c := newDiskTestCache(t, 64)
	p := newTestPiece(c, 0)
	dp := NewDiskPiece(p)

	want := []byte("hello, disk\n")
	n, err := dp.WriteAt(want, 0)
	if err != nil {
		t.Fatalf("WriteAt err: %v", err)
	}
	if n != len(want) {
		t.Fatalf("WriteAt n=%d want %d", n, len(want))
	}

	got := make([]byte, len(want))
	n, err = dp.ReadAt(got, 0)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("ReadAt err: %v", err)
	}
	if n != len(want) {
		t.Fatalf("ReadAt n=%d want %d", n, len(want))
	}
	if string(got) != string(want) {
		t.Fatalf("ReadAt got %q want %q", got, want)
	}
}

func TestDiskPiece_ReadMissingReturnsEOF(t *testing.T) {
	c := newDiskTestCache(t, 64)
	p := newTestPiece(c, 7)
	dp := NewDiskPiece(p)

	buf := make([]byte, 4)
	n, err := dp.ReadAt(buf, 0)
	if n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("ReadAt missing: n=%d err=%v, want n=0 err=EOF", n, err)
	}
}

// TestDiskPiece_ReadAt_SwallowsError documents the current bug (#8 in our
// findings): DiskPiece.ReadAt always returns nil error, even when the
// underlying os.File.ReadAt returned a real error like io.EOF.
//
// We detect this by writing 4 bytes and asking for 8: the underlying read
// returns (4, io.EOF), but DiskPiece returns (4, nil). After the bug is
// TestDiskPiece_ReadAt_PropagatesShortReadError gates the fix ported
// from upstream 71d05617: DiskPiece.ReadAt must propagate the underlying
// os.File.ReadAt error, not swallow it. Returning (n, nil) on a short
// read makes the anacrolix/torrent storage wrapper panic with "io.Copy
// will get stuck" (same class as anacrolix/torrent#430). MemPiece.ReadAt
// already returned io.EOF here; only DiskPiece was broken.
func TestDiskPiece_ReadAt_PropagatesShortReadError(t *testing.T) {
	c := newDiskTestCache(t, 64)
	p := newTestPiece(c, 0)
	dp := NewDiskPiece(p)

	if _, err := dp.WriteAt([]byte("abcd"), 0); err != nil {
		t.Fatalf("WriteAt: %v", err)
	}

	// Read more than was written: os.File.ReadAt returns io.EOF.
	buf := make([]byte, 8)
	n, err := dp.ReadAt(buf, 0)

	if err == nil {
		t.Fatalf("REGRESSION: DiskPiece.ReadAt swallowed the short-read "+
			"error; got (%d, nil), want (%d, io.EOF-or-similar). "+
			"anacrolix/torrent storage wrapper will panic 'io.Copy will "+
			"get stuck' on the (n, nil) response.", n, n)
	}
	t.Logf("OK: DiskPiece.ReadAt returned (n=%d, err=%v) for short read", n, err)
}

func TestDiskPiece_Release_RemovesFile(t *testing.T) {
	c := newDiskTestCache(t, 64)
	p := newTestPiece(c, 3)
	dp := NewDiskPiece(p)

	if _, err := dp.WriteAt([]byte("abcd"), 0); err != nil {
		t.Fatalf("WriteAt: %v", err)
	}
	fname := filepath.Join(settings.BTsets.TorrentsSavePath, c.hash.HexString(), strconv.Itoa(3))
	if _, err := os.Stat(fname); err != nil {
		t.Fatalf("file not created: %v", err)
	}

	dp.Release()

	if _, err := os.Stat(fname); !os.IsNotExist(err) {
		t.Fatalf("file still present after Release: err=%v", err)
	}
	if p.Size.Load() != 0 || p.Complete.Load() {
		t.Fatalf("piece state after Release: Size=%d Complete=%v", p.Size.Load(), p.Complete.Load())
	}
}
