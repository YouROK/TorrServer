package torrstor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
)

func setupDiskCacheSettings(t *testing.T) (cacheDir string) {
	t.Helper()
	old := settings.BTsets
	cacheDir = t.TempDir()
	settings.BTsets = &settings.BTSets{UseDisk: true, LongTermCache: true, TorrentsSavePath: cacheDir}
	t.Cleanup(func() { settings.BTsets = old })
	return cacheDir
}

func writeCacheFile(t *testing.T, path string, size int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0o666); err != nil {
		t.Fatal(err)
	}
}

func TestDiskCacheWalkersCountPieceFilesOnly(t *testing.T) {
	cacheDir := setupDiskCacheSettings(t)
	dir := filepath.Join(cacheDir, testHash)
	writeCacheFile(t, filepath.Join(dir, "0"), 10)
	writeCacheFile(t, filepath.Join(dir, "7"), 20)
	writeCacheFile(t, filepath.Join(dir, "1"), 0)      // empty: not cached
	writeCacheFile(t, filepath.Join(dir, "info"), 100) // metainfo
	writeCacheFile(t, filepath.Join(dir, "pins"), 100) // download state
	writeCacheFile(t, filepath.Join(dir, "junk"), 100) // not a piece
	writeCacheFile(t, filepath.Join(dir, "sub", "3"), 100)
	writeCacheFile(t, filepath.Join(cacheDir, "not-a-hash", "0"), 100)

	if got := diskCacheUsage(); got != 30 {
		t.Errorf("diskCacheUsage = %d, want 30", got)
	}
	refs := scanDiskCachePieces()
	if len(refs) != 2 {
		t.Fatalf("scanDiskCachePieces = %+v, want pieces 0 and 7", refs)
	}
	sizes := CachedPieceSizes(testHash)
	if len(sizes) != 2 || sizes[0] != 10 || sizes[7] != 20 {
		t.Errorf("CachedPieceSizes = %v, want map[0:10 7:20]", sizes)
	}
	var total int64
	for _, ref := range refs {
		if ref.hash != testHash || sizes[ref.pieceId] != ref.size {
			t.Errorf("scanDiskCachePieces ref %+v does not match CachedPieceSizes", ref)
		}
		total += ref.size
	}
	if total != diskCacheUsage() {
		t.Errorf("scan total %d != diskCacheUsage %d", total, diskCacheUsage())
	}
}

func newDiskTestCache(t *testing.T) *Cache {
	t.Helper()
	c := NewCache(0, NewStorage(0))
	c.pieceLength = 100
	c.pieceCount = 3
	c.totalLength = 250
	c.hash = metainfo.NewHashFromHex(testHash)
	return c
}

func TestMarkNotCompleteRemovesFailedDiskPiece(t *testing.T) {
	cacheDir := setupDiskCacheSettings(t)
	c := newDiskTestCache(t)
	name := filepath.Join(cacheDir, testHash, "2")
	writeCacheFile(t, name, 50) // the last piece is 50 bytes

	p := NewPiece(2, c)
	if !p.Complete || p.Size != 50 {
		t.Fatalf("full piece file: Complete=%v Size=%d, want complete 50", p.Complete, p.Size)
	}
	before := diskUsed.Load()
	if err := p.MarkNotComplete(); err != nil {
		t.Fatal(err)
	}
	if p.Complete || p.Size != 0 {
		t.Errorf("after MarkNotComplete: Complete=%v Size=%d, want false 0", p.Complete, p.Size)
	}
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		t.Errorf("failed piece file still on disk: %v", err)
	}
	if got := before - diskUsed.Load(); got != 50 {
		t.Errorf("disk usage dropped by %d, want 50", got)
	}
	// after a restart the piece must not be taken for complete
	if p := NewPiece(2, c); p.Complete {
		t.Error("piece is complete after reopening")
	}
}

func TestMarkNotCompleteKeepsPartialDiskPiece(t *testing.T) {
	cacheDir := setupDiskCacheSettings(t)
	c := newDiskTestCache(t)
	name := filepath.Join(cacheDir, testHash, "0")
	writeCacheFile(t, name, 40)

	p := NewPiece(0, c)
	before := diskUsed.Load()
	if err := p.MarkNotComplete(); err != nil {
		t.Fatal(err)
	}
	if p.Size != 40 {
		t.Errorf("partial piece Size = %d, want 40", p.Size)
	}
	if _, err := os.Stat(name); err != nil {
		t.Errorf("partial piece file removed: %v", err)
	}
	if diskUsed.Load() != before {
		t.Errorf("disk usage changed for a kept piece")
	}
}

func TestEvictionKeepsCardFiles(t *testing.T) {
	cacheDir := setupDiskCacheSettings(t)
	settings.BTsets.DiskCacheSize = 10
	dir := filepath.Join(cacheDir, testHash)
	writeCacheFile(t, filepath.Join(dir, "0"), 100)
	writeCacheFile(t, filepath.Join(dir, "info"), 5)
	writeCacheFile(t, filepath.Join(dir, "pins"), 5)
	diskUsedSync.Store(0) // force a rescan

	NewStorage(0).cleanGlobal()

	if _, err := os.Stat(filepath.Join(dir, "0")); !os.IsNotExist(err) {
		t.Errorf("piece over the quota not evicted: %v", err)
	}
	for _, name := range []string{"info", "pins"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("%s removed by eviction: %v", name, err)
		}
	}
}
