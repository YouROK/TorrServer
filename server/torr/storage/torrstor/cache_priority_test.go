package torrstor

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/anacrolix/torrent"

	"server/settings"
)

func testBTSets(t *testing.T) {
	t.Helper()
	prev := settings.BTsets
	settings.BTsets = &settings.BTSets{
		ConnectionsLimit:         50,
		ReaderReadAHead:          95,
		TorrentDisconnectTimeout: 60,
		CacheSize:                64 << 20,
	}
	t.Cleanup(func() { settings.BTsets = prev })
}

func TestPrioIDsToClearIsOAssigned(t *testing.T) {
	const n = 50000
	c := &Cache{
		pieces:       make(map[int]*Piece, n),
		activePieces: make(map[int]struct{}),
		prioPieces:   map[int]struct{}{10: {}, 20: {}, 30: {}},
		readers:      make(map[*Reader]struct{}),
		pieceCount:   n,
		pieceLength:  1 << 20,
	}
	for i := 0; i < n; i++ {
		c.pieces[i] = &Piece{Id: i, cache: c}
	}
	ids := c.prioIDsToClear(nil)
	if len(ids) != 3 {
		t.Fatalf("clearPriority must visit assigned ids only, got %d", len(ids))
	}
	c.clearPriority()
	if len(c.prioPieces) != 0 {
		t.Fatalf("expected prioPieces cleared, got %d", len(c.prioPieces))
	}
}

func TestGetRemPiecesUsesActivePiecesOnly(t *testing.T) {
	testBTSets(t)
	const n = 50000
	c := &Cache{
		pieces:       make(map[int]*Piece, n),
		activePieces: make(map[int]struct{}),
		prioPieces:   make(map[int]struct{}),
		readers:      make(map[*Reader]struct{}),
		pieceCount:   n,
		pieceLength:  1 << 20,
		capacity:     64 << 20,
	}
	for i := 0; i < n; i++ {
		c.pieces[i] = &Piece{Id: i, cache: c}
	}
	c.pieces[10].Size = 100
	c.pieces[20].Size = 100
	c.notePieceFilled(10)
	c.notePieceFilled(20)
	// Size>0 but not tracked as active — O(pieceCount) walk would evict this.
	c.pieces[999].Size = 500

	start := time.Now()
	rem := c.getRemPieces()
	if time.Since(start) > time.Second {
		t.Fatal("getRemPieces walked too many pieces")
	}
	got := map[int]bool{}
	for _, p := range rem {
		got[p.Id] = true
	}
	if !got[10] || !got[20] {
		t.Fatalf("expected active pieces 10 and 20 in rem, got %v", got)
	}
	if got[999] {
		t.Fatal("piece 999 is not in activePieces and must not be scanned")
	}
}

func TestSetLoadPriorityArmsFileStart(t *testing.T) {
	testBTSets(t)
	const pieceLen int64 = 2 << 20
	c := &Cache{
		pieces:       make(map[int]*Piece),
		activePieces: make(map[int]struct{}),
		prioPieces:   make(map[int]struct{}),
		readers:      make(map[*Reader]struct{}),
		pieceCount:   100,
		pieceLength:  pieceLen,
		capacity:     64 << 20,
	}
	for i := 0; i < 40; i++ {
		c.pieces[i] = &Piece{Id: i, cache: c}
	}
	r := &Reader{
		cache:       c,
		isUse:       true,
		offset:      0,
		overrideLen: 64 << 20,
		lastAccess:  time.Now().Unix(),
	}
	c.readers[r] = struct{}{}

	c.setLoadPriority(nil)
	if _, ok := c.prioPieces[0]; !ok {
		t.Fatal("reader at offset 0 must get PiecePriorityNow")
	}
}

func TestGetStateWithActiveReaderDoesNotDeadlock(t *testing.T) {
	testBTSets(t)
	c := &Cache{
		pieces:       make(map[int]*Piece),
		activePieces: make(map[int]struct{}),
		prioPieces:   make(map[int]struct{}),
		readers:      make(map[*Reader]struct{}),
		pieceCount:   10,
		pieceLength:  1024,
		capacity:     10 * 1024,
	}
	r := &Reader{
		cache:       c,
		isUse:       true,
		offset:      0,
		overrideLen: 4096,
		lastAccess:  time.Now().Unix(),
	}
	c.readers[r] = struct{}{}

	done := make(chan struct{})
	go func() {
		_ = c.GetState()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("GetState deadlocked on muReaders via getPiecesRange")
	}
}

func TestSeekEndSizeProbeDoesNotForward(t *testing.T) {
	const fileLen int64 = 193450314699
	stub := &seekRecordingReader{}
	r := &Reader{
		Reader:      stub,
		overrideLen: fileLen,
		isUse:       true,
		lastAccess:  time.Now().Unix(),
	}
	n, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		t.Fatal(err)
	}
	if n != fileLen {
		t.Fatalf("SeekEnd size probe: want %d, got %d", fileLen, n)
	}
	if r.offset != fileLen {
		t.Fatalf("offset: want %d, got %d", fileLen, r.offset)
	}
	if stub.seeks != 0 {
		t.Fatalf("size probe must not Seek the torrent reader, got %d calls", stub.seeks)
	}

	n, err = r.Seek(-100, io.SeekEnd)
	if err != nil {
		t.Fatal(err)
	}
	if stub.seeks != 1 {
		t.Fatalf("Seek(-n, SeekEnd) must forward, got %d calls", stub.seeks)
	}
	if n != -100 {
		t.Fatalf("forwarded Seek return: want -100, got %d", n)
	}
}

type seekRecordingReader struct {
	seeks int
}

var _ torrent.Reader = (*seekRecordingReader)(nil)

func (s *seekRecordingReader) Read([]byte) (int, error) { return 0, io.EOF }
func (s *seekRecordingReader) Close() error             { return nil }
func (s *seekRecordingReader) Seek(offset int64, whence int) (int64, error) {
	s.seeks++
	return offset, nil
}
func (s *seekRecordingReader) ReadContext(_ context.Context, _ []byte) (int, error) {
	return 0, io.EOF
}
func (s *seekRecordingReader) SetReadahead(int64) {}
func (s *seekRecordingReader) SetResponsive()     {}

func TestEmptyRangesKeepFileStart(t *testing.T) {
	testBTSets(t)
	c := &Cache{
		pieces:       make(map[int]*Piece),
		activePieces: make(map[int]struct{}),
		prioPieces:   make(map[int]struct{}),
		readers:      make(map[*Reader]struct{}),
		pieceCount:   100,
		pieceLength:  1 << 20,
		capacity:     64 << 20,
	}
	c.pieces[0] = &Piece{Id: 0, Size: 100, cache: c}
	c.pieces[50] = &Piece{Id: 50, Size: 100, cache: c}
	c.notePieceFilled(0)
	c.notePieceFilled(50)

	rem := c.getRemPieces()
	got := map[int]bool{}
	for _, p := range rem {
		got[p.Id] = true
	}
	if got[0] {
		t.Fatal("piece 0 is file start and must survive reader disconnect")
	}
	if !got[50] {
		t.Fatal("piece 50 is outside keep-window and should be evictable")
	}
}

func TestDoNotEvictHashPendingPiece(t *testing.T) {
	testBTSets(t)
	const plen int64 = 1 << 20
	c := &Cache{
		pieces:       make(map[int]*Piece),
		activePieces: make(map[int]struct{}),
		prioPieces:   make(map[int]struct{}),
		readers:      make(map[*Reader]struct{}),
		pieceCount:   100,
		pieceLength:  plen,
		totalLength:  plen * 100,
		capacity:     64 << 20,
	}
	c.pieces[50] = &Piece{Id: 50, Size: plen, Complete: false, cache: c}
	c.notePieceFilled(50)

	rem := c.getRemPieces()
	for _, p := range rem {
		if p.Id == 50 {
			t.Fatal("fully written incomplete piece must not be released during hash")
		}
	}
}

func TestDoNotEvictPrioPieces(t *testing.T) {
	testBTSets(t)
	c := &Cache{
		pieces:       make(map[int]*Piece),
		activePieces: make(map[int]struct{}),
		prioPieces:   map[int]struct{}{50: {}},
		readers:      make(map[*Reader]struct{}),
		pieceCount:   100,
		pieceLength:  1 << 20,
		capacity:     64 << 20,
	}
	c.pieces[50] = &Piece{Id: 50, Size: 100, cache: c}
	c.notePieceFilled(50)

	rem := c.getRemPieces()
	for _, p := range rem {
		if p.Id == 50 {
			t.Fatal("piece with assigned priority must not be evicted")
		}
	}
}

func TestSetTorrentNilReceiver(t *testing.T) {
	var c *Cache
	c.SetTorrent(nil)
}

func TestCloseIdempotent(t *testing.T) {
	prev := settings.BTsets
	settings.BTsets = &settings.BTSets{}
	t.Cleanup(func() { settings.BTsets = prev })
	c := NewCache(1<<20, NewStorage(1<<20))
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReadAtNilBufferIsNotExist(t *testing.T) {
	c := &Cache{pieceLength: 1024, pieceCount: 1, totalLength: 1024}
	p := &Piece{Id: 0, cache: c}
	mp := NewMemPiece(p)
	n, err := mp.ReadAt(make([]byte, 100), 0)
	if n != 0 {
		t.Fatalf("want 0 bytes, got %d", n)
	}
	if !os.IsNotExist(err) {
		t.Fatalf("want os.ErrNotExist, got %v", err)
	}
}
