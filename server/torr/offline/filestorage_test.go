package offline

import (
	"container/list"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anacrolix/torrent/metainfo"
)

func TestParseSidecar(t *testing.T) {
	pieces := 4
	valid := make([]byte, 0, len(sidecarMagic)+pieces+4)
	valid = append(valid, sidecarMagic...)
	valid = append(valid, []byte("1010")...)
	crc := crc32.ChecksumIEEE(valid)
	valid = append(valid, byte(crc>>24), byte(crc>>16), byte(crc>>8), byte(crc))

	bits, ok := parseSidecar(valid, pieces)
	if !ok || string(bits) != "1010" {
		t.Fatalf("valid sidecar not accepted: bits=%q ok=%v", bits, ok)
	}

	corrupt := append([]byte(nil), valid...)
	corrupt[len(sidecarMagic)] = '0' // break the bitmap, keep the CRC
	if _, ok := parseSidecar(corrupt, pieces); ok {
		t.Fatal("sidecar with a CRC mismatch must be rejected")
	}

	truncated := valid[:len(valid)-1]
	if _, ok := parseSidecar(truncated, pieces); ok {
		t.Fatal("sidecar with a wrong length must be rejected")
	}

	if _, ok := parseSidecar(valid, pieces+1); ok {
		t.Fatal("sidecar must be rejected when the piece count does not match")
	}

	legacy := []byte("0110")
	if bits, ok := parseSidecar(legacy, 4); !ok || string(bits) != "0110" {
		t.Fatalf("legacy sidecar not accepted: bits=%q ok=%v", bits, ok)
	}
	if _, ok := parseSidecar(legacy[:3], 4); ok {
		t.Fatal("legacy sidecar with a wrong length must be rejected")
	}
}

func TestSidecarProgress(t *testing.T) {
	base := t.TempDir()
	ih := testHash()
	const (
		pieceLen = 100
		pieces   = 10
		total    = 950 // the last piece is 50 bytes
	)
	writeBits := func(bits string) {
		t.Helper()
		buf := make([]byte, 0, len(sidecarMagic)+pieces+4)
		buf = append(buf, sidecarMagic...)
		buf = append(buf, bits...)
		crc := crc32.ChecksumIEEE(buf)
		buf = append(buf, byte(crc>>24), byte(crc>>16), byte(crc>>8), byte(crc))
		if err := os.WriteFile(sidecarPath(base, ih), buf, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeBits("1111111111")
	if b, ok := SidecarProgress(base, ih, pieceLen, pieces, total); !ok || b != 950 {
		t.Fatalf("fully complete sidecar: got %d bytes (ok=%v), want 950", b, ok)
	}

	writeBits("1111111110")
	if b, ok := SidecarProgress(base, ih, pieceLen, pieces, total); !ok || b != 900 {
		t.Fatalf("all but the last piece: got %d bytes (ok=%v), want 900", b, ok)
	}

	writeBits("0000000000")
	if b, ok := SidecarProgress(base, ih, pieceLen, pieces, total); !ok || b != 0 {
		t.Fatalf("empty sidecar: got %d bytes (ok=%v), want 0", b, ok)
	}

	if err := os.Remove(sidecarPath(base, ih)); err != nil {
		t.Fatal(err)
	}
	if b, ok := SidecarProgress(base, ih, pieceLen, pieces, total); ok || b != 0 {
		t.Fatalf("missing sidecar: got %d bytes (ok=%v), want 0,false", b, ok)
	}
}

func TestVerifiedBytes(t *testing.T) {
	const (
		pieceLen = 100
		pieces   = 10
		total    = 950 // the last piece is 50 bytes
	)
	ts := &torrentStore{
		pieceLen: pieceLen,
		total:    total,
		state:    make([]uint8, pieces),
	}
	complete := func(n int) {
		for i := range ts.state {
			if i < n {
				ts.state[i] = pieceComplete
			} else {
				ts.state[i] = pieceNotComplete
			}
		}
	}
	complete(5)
	if got := ts.verifiedBytes(); got != 500 {
		t.Fatalf("5 complete pieces: got %d, want 500", got)
	}
	complete(10)
	if got := ts.verifiedBytes(); got != 950 {
		t.Fatalf("all pieces complete: got %d, want 950 (the last piece is shorter)", got)
	}
	complete(9) // the last piece is not complete
	if got := ts.verifiedBytes(); got != 900 {
		t.Fatalf("9 complete pieces: got %d, want 900", got)
	}
	complete(0)
	if got := ts.verifiedBytes(); got != 0 {
		t.Fatalf("nothing complete: got %d, want 0", got)
	}
}

// sidecar with every piece marked complete + no files on disk: the state must
// be distrusted and the pieces re-verified by hashing. This is the failure
// mode of review remark V-1 (a stale sidecar made a re-download instantly
// "finish" with no files on disk).
func TestNewTorrentStoreDistrustsCompleteSidecarWithMissingFiles(t *testing.T) {
	base := t.TempDir()
	ih := testHash()
	info := &metainfo.Info{
		Name:        "Show",
		PieceLength: 32768,
		Pieces:      make([]byte, 40), // 2 pieces
		Files:       []metainfo.FileInfo{{Path: []string{"movie.mkv"}, Length: 40000}},
	}
	writeBits(t, base, ih, "11")

	s := NewStorage(base)
	ts := newTorrentStore(s, base, info, ih)
	defer ts.Close()

	if got := ts.completeCount(); got != 0 {
		t.Fatalf("complete sidecar with missing files: %d pieces reported complete, want 0", got)
	}
	// no files on disk at all: the store falls back to the "new download" mode
	for i, st := range ts.state {
		if st != pieceNotComplete {
			t.Fatalf("piece %d: state %d, want pieceNotComplete (fresh download)", i, st)
		}
	}
}

func TestNewTorrentStoreTrustsCompleteSidecarWithIntactFiles(t *testing.T) {
	base := t.TempDir()
	ih := testHash()
	info := &metainfo.Info{
		Name:        "Show",
		PieceLength: 32768,
		Pieces:      make([]byte, 40), // 2 pieces
		Files:       []metainfo.FileInfo{{Path: []string{"movie.mkv"}, Length: 40000}},
	}
	writeBits(t, base, ih, "11")
	path := filepath.Join(TorrentRoot(base, ih), "Show", "movie.mkv")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, 40000), 0o644); err != nil {
		t.Fatal(err)
	}

	s := NewStorage(base)
	ts := newTorrentStore(s, base, info, ih)
	defer ts.Close()

	if got := ts.completeCount(); got != 2 {
		t.Fatalf("complete sidecar with intact files: %d pieces reported complete, want 2", got)
	}
}

// sidecar says complete, but one file was truncated on disk: the state must be
// distrusted as well (the partial case of review remark V-1).
func TestNewTorrentStoreDistrustsCompleteSidecarWithBrokenFile(t *testing.T) {
	base := t.TempDir()
	ih := testHash()
	info := &metainfo.Info{
		Name:        "Show",
		PieceLength: 32768,
		Pieces:      make([]byte, 40), // 2 pieces
		Files:       []metainfo.FileInfo{{Path: []string{"movie.mkv"}, Length: 40000}},
	}
	writeBits(t, base, ih, "11")
	path := filepath.Join(TorrentRoot(base, ih), "Show", "movie.mkv")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, 100), 0o644); err != nil {
		t.Fatal(err)
	}

	s := NewStorage(base)
	ts := newTorrentStore(s, base, info, ih)
	defer ts.Close()

	if got := ts.completeCount(); got != 0 {
		t.Fatalf("complete sidecar with a truncated file: %d pieces reported complete, want 0", got)
	}
	// files exist on disk (with a wrong size): every piece must be re-verified by hashing
	for i, st := range ts.state {
		if st != pieceUnknown {
			t.Fatalf("piece %d: state %d, want pieceUnknown (must be verified by hashing)", i, st)
		}
	}
}

func writeBits(t *testing.T, base string, ih metainfo.Hash, bits string) {
	t.Helper()
	buf := make([]byte, 0, len(sidecarMagic)+len(bits)+4)
	buf = append(buf, sidecarMagic...)
	buf = append(buf, bits...)
	crc := crc32.ChecksumIEEE(buf)
	buf = append(buf, byte(crc>>24), byte(crc>>16), byte(crc>>8), byte(crc))
	if err := os.WriteFile(sidecarPath(base, ih), buf, 0o644); err != nil {
		t.Fatal(err)
	}
}

// guard against accidental changes of the sidecar format
func TestSidecarMagic(t *testing.T) {
	if sidecarMagic != "TSOFF1\n" {
		t.Fatalf("sidecarMagic = %q, want %q", sidecarMagic, "TSOFF1\n")
	}
	if !strings.HasSuffix(sidecarPath("/base", testHash()), ".pieces") {
		t.Fatal("sidecar file name must end with .pieces")
	}
}

// lruStore builds a torrentStore over count real files of 16 bytes each.
func lruStore(t *testing.T, count int) (*torrentStore, string) {
	t.Helper()
	base := t.TempDir()
	files := make([]File, count)
	for i := range files {
		name := fmt.Sprintf("f%02d.bin", i)
		path := filepath.Join(base, name)
		if err := os.WriteFile(path, []byte("0123456789abcdef"), 0o644); err != nil {
			t.Fatal(err)
		}
		files[i] = File{Path: name, Disk: path, Offset: int64(i) * 16, Length: 16}
	}
	ts := &torrentStore{
		owner:    NewStorage(base),
		hash:     testHash(),
		files:    files,
		pieceLen: 16,
		total:    int64(count) * 16,
		state:    []uint8{pieceNotComplete},
		sidecar:  filepath.Join(base, testHash().HexString()+".pieces"),
		stop:     make(chan struct{}),
		handles:  make(map[int]*fileHandle),
		lru:      list.New(),
	}
	return ts, base
}

// TestLRUEviction checks that the open file descriptors are evicted least
// recently used and that the soft limit holds for sequential reads
// (review remark N-2).
func TestLRUEviction(t *testing.T) {
	const count = 80
	ts, _ := lruStore(t, count)

	for i := 0; i < count; i++ {
		buf := make([]byte, 16)
		if _, err := ts.readAt(buf, int64(i)*16); err != nil {
			t.Fatalf("read file %d: %v", i, err)
		}
		if len(ts.handles) > maxOpenFiles {
			t.Fatalf("after reading file %d: %d handles cached, limit is %d", i, len(ts.handles), maxOpenFiles)
		}
	}

	// the handles of the early files were evicted: reading them again must
	// transparently reopen them
	buf := make([]byte, 16)
	if _, err := ts.readAt(buf, 0); err != nil {
		t.Fatalf("re-read of an evicted file: %v", err)
	}

	// pinned handles are not evicted: the limit is soft while they are in use
	hs := make([]*fileHandle, 0, maxOpenFiles+6)
	for i := 0; i < maxOpenFiles+6; i++ {
		h, err := ts.acquire(i, false)
		if err != nil {
			t.Fatal(err)
		}
		hs = append(hs, h)
	}
	if len(ts.handles) != maxOpenFiles+6 {
		t.Fatalf("pinned handles: %d cached, want %d (the limit must be soft)", len(ts.handles), maxOpenFiles+6)
	}
	for _, h := range hs {
		ts.release(h)
	}

	// everything unpinned again: the next acquire evicts back under the limit
	h, err := ts.acquire(count-1, false)
	if err != nil {
		t.Fatal(err)
	}
	ts.release(h)
	if len(ts.handles) > maxOpenFiles {
		t.Fatalf("after release: %d handles cached, limit is %d", len(ts.handles), maxOpenFiles)
	}

	// Close must not deadlock with pinned handles and must close everything
	for i := 0; i < maxOpenFiles+2; i++ {
		h, err := ts.acquire(i, false)
		if err != nil {
			t.Fatal(err)
		}
		hs = append(hs, h)
	}
	if err := ts.Close(); err != nil {
		t.Fatal(err)
	}
	for _, h := range hs {
		ts.release(h)
	}
	if _, err := ts.acquire(0, false); err == nil {
		t.Fatal("acquire after Close must fail")
	}
}

// TestNonCachedWriteHandleIsClosed: a write that arrives while the cached
// read-only handle of the same file is pinned by a reader gets a separate,
// non-cached descriptor. It must be closed on release (it used to leak).
func TestNonCachedWriteHandleIsClosed(t *testing.T) {
	ts, _ := lruStore(t, 1)

	ro, err := ts.acquire(0, false) // pinned read-only handle (a reader)
	if err != nil {
		t.Fatal(err)
	}
	w, err := ts.acquire(0, true)
	if err != nil {
		t.Fatal(err)
	}
	if w == ro || w.elem != nil || !w.rw {
		t.Fatalf("expected a separate non-cached writable handle, got %+v", w)
	}
	ts.release(w)
	if err := w.f.Close(); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("non-cached handle was not closed on release (descriptor leak): Close() = %v", err)
	}
	// the pinned reader is unaffected
	if _, err := ro.f.ReadAt(make([]byte, 4), 0); err != nil {
		t.Fatalf("pinned read-only handle must stay usable: %v", err)
	}
	ts.release(ro)

	// writeAt through the same path must not leak either
	ro, err = ts.acquire(0, false)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := ts.writeAt([]byte("XY"), 0); err != nil || n != 2 {
		t.Fatalf("writeAt: n=%d err=%v", n, err)
	}
	ts.release(ro)

	// a non-cached handle released after Close is closed as well
	ro, _ = ts.acquire(0, false)
	w, err = ts.acquire(0, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := ts.Close(); err != nil {
		t.Fatal(err)
	}
	ts.release(w)
	if err := w.f.Close(); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("non-cached handle released after Close was not closed: %v", err)
	}
	ts.release(ro)
}

// TestSaveCompletionSyncsWrittenFiles: the piece state is saved only after the
// files that were written to have been fsynced.
func TestSaveCompletionSyncsWrittenFiles(t *testing.T) {
	ts, _ := lruStore(t, 2)
	if _, err := ts.writeAt([]byte("zz"), 16); err != nil { // second file
		t.Fatal(err)
	}
	ts.smu.Lock()
	_, marked := ts.dirtyFiles[1]
	ts.smu.Unlock()
	if !marked {
		t.Fatal("a written file must be remembered for the next fsync")
	}
	p := &pieceImpl{ts: ts, idx: 0}
	if err := p.MarkComplete(); err != nil {
		t.Fatal(err)
	}
	if err := ts.saveCompletion(); err != nil {
		t.Fatalf("saveCompletion: %v", err)
	}
	ts.smu.Lock()
	left := len(ts.dirtyFiles)
	ts.smu.Unlock()
	if left != 0 {
		t.Fatalf("%d files still marked dirty after the flush", left)
	}
	if _, err := os.Stat(ts.sidecar); err != nil {
		t.Fatalf("sidecar was not written: %v", err)
	}
}

// TestSaveCompletionDoesNotSaveStateWhenSyncFails: when the data cannot be
// flushed the state must not be written (it would vouch for unflushed data) and
// must stay dirty for a retry.
func TestSaveCompletionDoesNotSaveStateWhenSyncFails(t *testing.T) {
	ts, _ := lruStore(t, 1)
	ts.markFileDirty(0)
	// a directory in place of the file: it cannot be opened for writing and syncing
	if err := os.Remove(ts.files[0].Disk); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(ts.files[0].Disk, 0o755); err != nil {
		t.Fatal(err)
	}
	p := &pieceImpl{ts: ts, idx: 0}
	if err := p.MarkComplete(); err != nil {
		t.Fatal(err)
	}
	if err := ts.saveCompletion(); err == nil {
		t.Fatal("saveCompletion must fail when a file cannot be synced")
	}
	if _, err := os.Stat(ts.sidecar); !os.IsNotExist(err) {
		t.Fatalf("the state must not be saved after a failed sync: %v", err)
	}
	ts.cmu.Lock()
	dirty := ts.dirty
	ts.cmu.Unlock()
	if !dirty {
		t.Fatal("the state must stay dirty for a retry")
	}
}
