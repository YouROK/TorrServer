package torrstor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
)

func countCompletionUpdates(t *testing.T) *int {
	t.Helper()
	var n int
	prev := pieceCompletionUpdater
	pieceCompletionUpdater = func(*Piece) { n++ }
	t.Cleanup(func() { pieceCompletionUpdater = prev })
	return &n
}

func TestPieceReleaseMemClearsBufferAndNotifies(t *testing.T) {
	testBTSets(t)
	settings.BTsets.UseDisk = false
	updates := countCompletionUpdates(t)
	c := &Cache{
		pieces:       map[int]*Piece{},
		activePieces: map[int]struct{}{},
		readers:      map[*Reader]struct{}{},
		pieceLength:  16, pieceCount: 1, totalLength: 16, capacity: 1 << 20,
	}
	p := NewPiece(0, c)
	c.pieces[0] = p
	if _, err := p.WriteAt([]byte("hello"), 0); err != nil {
		t.Fatal(err)
	}

	p.Release()
	if *updates != 1 {
		t.Fatalf("completion updates = %d, want 1", *updates)
	}
	if _, err := p.ReadAt(make([]byte, 5), 0); !os.IsNotExist(err) {
		t.Fatalf("read after release: %v", err)
	}

	// A closed cache must not call back into the client.
	c.isClosed.Store(true)
	p.Release()
	if *updates != 1 {
		t.Fatalf("closed cache notified: %d", *updates)
	}
}

func TestPieceReleaseDiskRemovesFile(t *testing.T) {
	testBTSets(t)
	dir := t.TempDir()
	settings.BTsets.UseDisk = true
	settings.BTsets.TorrentsSavePath = dir
	updates := countCompletionUpdates(t)
	c := &Cache{
		pieces:       map[int]*Piece{},
		activePieces: map[int]struct{}{},
		hash:         metainfo.Hash{7},
		pieceLength:  8, pieceCount: 1, totalLength: 8, capacity: 1 << 20,
	}
	if err := os.MkdirAll(filepath.Join(dir, c.hash.HexString()), 0o777); err != nil {
		t.Fatal(err)
	}
	p := NewPiece(0, c)
	c.pieces[0] = p
	if _, err := p.WriteAt([]byte("abcdefgh"), 0); err != nil {
		t.Fatal(err)
	}
	if path := p.dPiece.name; path == "" {
		t.Fatal("disk piece has no file name")
	} else if _, err := os.Stat(path); err != nil {
		t.Fatalf("piece file missing before release: %v", err)
	} else {
		p.Release()
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("piece file kept after release: %v", err)
		}
	}
	if *updates != 1 {
		t.Fatalf("completion updates = %d", *updates)
	}

	var nilPiece *Piece
	if got := nilPiece.releaseBuffers(); got != "" {
		t.Fatalf("nil piece path %q", got)
	}
	if got := (&Piece{}).releaseBuffers(); got != "" {
		t.Fatalf("bare disk piece path %q", got)
	}
}

// Without a torrent pointer there is no client lock to take: evictPartial falls
// back to a plain Release.
func TestEvictPartialWithoutTorrentReleases(t *testing.T) {
	testBTSets(t)
	settings.BTsets.UseDisk = false
	updates := countCompletionUpdates(t)
	c := &Cache{
		pieces:       map[int]*Piece{},
		activePieces: map[int]struct{}{},
		readers:      map[*Reader]struct{}{},
		pieceLength:  16, pieceCount: 1, totalLength: 16, capacity: 1 << 20,
	}
	p := NewPiece(0, c)
	c.pieces[0] = p
	if _, err := p.WriteAt([]byte("hello"), 0); err != nil {
		t.Fatal(err)
	}
	evictPartial(p)
	if *updates != 1 {
		t.Fatalf("completion updates = %d", *updates)
	}
	if _, err := p.ReadAt(make([]byte, 5), 0); !os.IsNotExist(err) {
		t.Fatalf("read after evict: %v", err)
	}
}

func TestPieceFakeAlwaysFailsAndStaysIncomplete(t *testing.T) {
	var f PieceFake
	if _, err := f.ReadAt(nil, 0); err == nil {
		t.Error("ReadAt")
	}
	if _, err := f.WriteAt(nil, 0); err == nil {
		t.Error("WriteAt")
	}
	if f.MarkComplete() == nil || f.MarkNotComplete() == nil {
		t.Error("Mark*")
	}
	if c := f.Completion(); c.Complete || !c.Ok {
		t.Errorf("completion %+v", c)
	}
}

func TestCacheAccessors(t *testing.T) {
	var nilCache *Cache
	if nilCache.GetCapacity() != 0 || nilCache.FilledBytes() != 0 || nilCache.PieceCount() != 0 || !nilCache.IsClosed() {
		t.Fatal("nil cache accessors")
	}
	nilCache.SetTorrent(nil) // must not panic

	c := &Cache{capacity: 123, pieceCount: 4}
	c.filled.Store(55)
	if c.GetCapacity() != 123 || c.FilledBytes() != 55 || c.PieceCount() != 4 || c.IsClosed() {
		t.Fatal("cache accessors")
	}
	c.isClosed.Store(true)
	if !c.IsClosed() {
		t.Fatal("IsClosed after close")
	}
	c.SetTorrent(nil)
	if c.torrentPtr() != nil {
		t.Fatal("torrentPtr")
	}
}
