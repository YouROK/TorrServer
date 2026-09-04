package torrstor

import (
	"testing"

	"server/settings"
)

func TestGetStateSkipsReadersWithoutFile(t *testing.T) {
	c := &Cache{
		pieces:       make(map[int]*Piece),
		activePieces: make(map[int]struct{}),
		readers:      make(map[*Reader]struct{}),
		pieceCount:   10,
		pieceLength:  1024,
		capacity:     10 * 1024,
	}
	// A nil file cannot yield a piece range — must be skipped before range calc.
	c.readers[&Reader{isUse: false, cache: c}] = struct{}{}
	// Closed readers are gone regardless of their file.
	c.readers[&Reader{isUse: true, isClosed: true, cache: c}] = struct{}{}

	st := c.GetState()
	if len(st.Readers) != 0 {
		t.Fatalf("expected fileless and closed readers omitted, got %d", len(st.Readers))
	}
}

func TestReaderIdleTimeout(t *testing.T) {
	prev := settings.BTsets
	t.Cleanup(func() { settings.BTsets = prev })

	settings.BTsets = &settings.BTSets{TorrentDisconnectTimeout: 30}
	r := &Reader{}
	// Floor applies: a 30 s disconnect timeout must not evict a paused player early.
	if got := r.idleTimeout(); got != minReaderIdleTimeout {
		t.Fatalf("want floor %d, got %d", minReaderIdleTimeout, got)
	}

	settings.BTsets = &settings.BTSets{TorrentDisconnectTimeout: 600}
	if got := r.idleTimeout(); got != 600 {
		t.Fatalf("want 600, got %d", got)
	}

	settings.BTsets = nil
	if got := r.idleTimeout(); got != minReaderIdleTimeout {
		t.Fatalf("nil settings: want %d, got %d", minReaderIdleTimeout, got)
	}
}

func TestMemPieceSizeAccumulatesOutOfOrder(t *testing.T) {
	c := &Cache{
		pieces:       make(map[int]*Piece),
		activePieces: make(map[int]struct{}),
		pieceCount:   3,
		pieceLength:  1000,
		totalLength:  3000,
	}
	p := &Piece{Id: 0, cache: c}
	// Pre-allocated buffer keeps WriteAt off the cleanPieces goroutine (nil torrent).
	mp := &MemPiece{piece: p, buffer: make([]byte, c.pieceLength)}
	p.mPiece = mp

	if _, err := mp.WriteAt(make([]byte, 300), 700); err != nil {
		t.Fatal(err)
	}
	if _, err := mp.WriteAt(make([]byte, 300), 0); err != nil {
		t.Fatal(err)
	}
	// A high-water mark would report 1000 here and count the 400-byte gap as filled.
	if p.Size != 600 {
		t.Fatalf("want accumulated Size 600, got %d", p.Size)
	}

	if _, err := mp.WriteAt(make([]byte, 600), 400); err != nil {
		t.Fatal(err)
	}
	if p.Size != 1000 {
		t.Fatalf("Size must cap at piece length 1000, got %d", p.Size)
	}

	// Hash failure re-downloads the piece — accumulation must restart from zero.
	if err := p.MarkNotComplete(); err != nil {
		t.Fatal(err)
	}
	if p.Size != 0 {
		t.Fatalf("MarkNotComplete must reset Size, got %d", p.Size)
	}
}

func TestInRangesInclusiveEnd(t *testing.T) {
	ranges := []Range{{Start: 2, End: 4}}
	if !inRanges(ranges, 2) || !inRanges(ranges, 4) {
		t.Fatalf("expected Start and End inclusive")
	}
	if inRanges(ranges, 1) || inRanges(ranges, 5) {
		t.Fatalf("expected outside range to be false")
	}
}

func TestNotePieceActiveTracking(t *testing.T) {
	c := &Cache{
		pieces:       make(map[int]*Piece),
		activePieces: make(map[int]struct{}),
		pieceCount:   10,
	}
	c.pieces[3] = &Piece{Id: 3, Size: 100, cache: c}
	c.pieces[7] = &Piece{Id: 7, Size: 0, cache: c}

	c.notePieceFilled(3)
	c.notePieceFilled(7)
	c.notePieceEmpty(7)

	c.muActive.Lock()
	_, ok3 := c.activePieces[3]
	_, ok7 := c.activePieces[7]
	c.muActive.Unlock()
	if !ok3 {
		t.Fatal("expected piece 3 active")
	}
	if ok7 {
		t.Fatal("expected piece 7 inactive")
	}
}

func TestPieceByteLengthLastPiece(t *testing.T) {
	c := &Cache{
		pieces:       make(map[int]*Piece),
		activePieces: make(map[int]struct{}),
		pieceCount:   3,
		pieceLength:  1000,
		totalLength:  2500, // last piece = 500
		capacity:     3000,
	}
	if got := c.pieceByteLength(0); got != 1000 {
		t.Fatalf("piece 0: want 1000, got %d", got)
	}
	if got := c.pieceByteLength(2); got != 500 {
		t.Fatalf("last piece: want 500, got %d", got)
	}

	c.pieces[2] = &Piece{Id: 2, Size: 500, Complete: false, cache: c}
	c.notePieceFilled(2)
	st := c.GetState()
	item := st.Pieces[2]
	if item.Length != 500 {
		t.Fatalf("GetState Length: want 500, got %d", item.Length)
	}
	if !item.Completed {
		t.Fatal("expected Completed when Size >= last-piece Length")
	}

	// MarkComplete alone must not report Completed without full Size (snake green vs Filled).
	c.pieces[1] = &Piece{Id: 1, Size: 100, Complete: true, cache: c}
	c.notePieceFilled(1)
	st = c.GetState()
	partial := st.Pieces[1]
	if partial.Completed {
		t.Fatal("partial Size must not be Completed even if MarkComplete was set")
	}
	if partial.Size != 100 {
		t.Fatalf("want Size 100, got %d", partial.Size)
	}
}
