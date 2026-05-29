package torrstor_test

// Layer-B tests focused on three suspected playback-stop bugs:
//
//   #5 cleanPieces() can evict a piece that lies inside an active
//      reader's range — the player is reading it RIGHT NOW.
//
//   #4 clearPriority() sleeps 1s before clearing all priorities; if a
//      new Seek lands during that window, the just-installed priorities
//      get clobbered to PiecePriorityNone, starving the player.
//
//   #3 readerOff() does Seek(0) on the underlying anacrolix Reader
//      after 60s of inactivity; a slow buffering player that resumes
//      after the timeout will trigger a wasteful rewind to the start of
//      the file.

import (
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
	"server/torr/storage/torrstor"
)

// makeBigTorrent builds an in-memory torrent with `numPieces` pieces of
// `pieceLen` bytes each. Content is deterministic but irrelevant — we
// never actually serve real bytes (we hand-fill cache pieces in tests).
func makeBigTorrent(t *testing.T, pieceLen int, numPieces int) *metainfo.MetaInfo {
	t.Helper()
	total := pieceLen * numPieces
	body := strings.Repeat("a", total)
	info := metainfo.Info{
		Name:        "bigfile.bin",
		PieceLength: int64(pieceLen),
		Length:      int64(total),
	}
	if err := info.GeneratePieces(func(metainfo.FileInfo) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(body)), nil
	}); err != nil {
		t.Fatalf("GeneratePieces: %v", err)
	}
	mi := &metainfo.MetaInfo{}
	var err error
	mi.InfoBytes, err = bencode.Marshal(info)
	if err != nil {
		t.Fatalf("Marshal info: %v", err)
	}
	return mi
}

// fillPiece pretends `pieceLen` bytes were downloaded into piece `id`,
// so getRemPieces sees Size > 0 and considers it for eviction.
func fillPiece(t *testing.T, cache *torrstor.Cache, _ *torrent.Torrent, id int, pieceLen int) {
	t.Helper()
	buf := make([]byte, pieceLen)
	for i := range buf {
		buf[i] = byte(id)
	}
	if _, err := cache.FillPieceForTest(id, buf); err != nil {
		t.Fatalf("fill piece %d: %v", id, err)
	}
}

// fillPiecePartial writes `nBytes` (< pieceLen) at offset 0 of piece `id`.
// The crucial property: MemPiece.WriteAt still allocates the full
// pieceLength buffer on first touch, even when only nBytes are written.
// Used to reproduce the "filled-undercount" memory bloat fixed in
// Cache.pieceCost.
func fillPiecePartial(t *testing.T, cache *torrstor.Cache, id, nBytes int) {
	t.Helper()
	buf := make([]byte, nBytes)
	for i := range buf {
		buf[i] = byte(id)
	}
	if _, err := cache.FillPieceForTest(id, buf); err != nil {
		t.Fatalf("partial fill piece %d: %v", id, err)
	}
}

// setupCache builds a torrent + cache with a custom capacity. Returns
// the torrent and cache; cleanup is wired through t.Cleanup.
func setupCache(t *testing.T, capacity int64, pieceLen int, numPieces int) (*torrent.Torrent, *torrstor.Cache) {
	t.Helper()
	ensureSettingsForIntegration(t)

	st := torrstor.NewStorage(capacity)
	cl := newIsolatedClient(t, st)

	mi := makeBigTorrent(t, pieceLen, numPieces)
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
	if cache == nil {
		t.Fatal("cache is nil")
	}
	cache.SetTorrent(tt)
	return tt, cache
}

// --------------------------------------------------------------------
// Bug #5: piece in active reader range must NOT be evicted.
// --------------------------------------------------------------------

// TestCleanPieces_DoesNotEvictPieceInActiveRange verifies the most
// fundamental cache invariant: a piece that is currently inside the
// active reader's range is OFF-LIMITS for eviction. If cleanPieces()
// removes it, the player is reading garbage / an empty piece and the
// stream stalls or aborts.
//
// Setup matters: the reader is registered and positioned BEFORE we
// start filling pieces, so each MemPiece.WriteAt-spawned cleanPieces()
// goroutine sees the active range and protects pieces inside it.
// (Registering the reader after the fill burst — as the original
// racebug test did — measures a different thing: the eviction policy's
// behaviour during preload, where it is correct to evict freely.)
func TestCleanPieces_DoesNotEvictPieceInActiveRange(t *testing.T) {
	const pieceLen = 64 * 1024
	const numPieces = 32
	const capacity = int64(pieceLen * 4) // only 4 pieces fit

	saved := settings.BTsets.ReaderReadAHead
	settings.BTsets.ReaderReadAHead = 95
	defer func() { settings.BTsets.ReaderReadAHead = saved }()

	tt, cache := setupCache(t, capacity, pieceLen, numPieces)

	reader := cache.NewReader(tt.Files()[0])
	const protectedPiece = 8
	torrstor.SeekReaderForTest(reader, int64(protectedPiece*pieceLen))
	torrstor.MarkReaderUseForTest(reader, true)

	// Now flood the cache; with capacity == 4 pieces and 16 fills the
	// eviction loop will run hard. Piece `protectedPiece` is inside
	// the reader range and must survive every sweep.
	for id := 0; id < 16; id++ {
		fillPiece(t, cache, tt, id, pieceLen)
	}
	// Drain the spawned goroutines.
	time.Sleep(100 * time.Millisecond)
	cache.CleanPiecesForTest()

	if got := cache.PieceSizeForTest(protectedPiece); got <= 0 {
		t.Fatalf("BUG #5 REPRODUCED: piece %d was evicted while reader is on it (Size=%d).",
			protectedPiece, got)
	}
}

// --------------------------------------------------------------------
// Bug #16: cleanPieces / getRemPieces race under WriteAt-spawned
// concurrent calls.
// --------------------------------------------------------------------

// TestCleanPieces_NoRace_ConcurrentWrites is the race-clean regression
// gate for bug #16. Each MemPiece.WriteAt on a fresh piece spawns a
// `go cleanPieces()`. Before the fix, those goroutines mutated
// c.isRemove, c.filled, p.Size, p.Accessed and p.Complete without
// synchronisation, so under -race a burst of fills produced data-race
// reports. After the fix (atomic.Bool isRemove/isClosed, atomic.Int64
// Size/Accessed, atomic.Bool Complete) the burst must complete cleanly.
//
// Run with `go test -race` — failure means the race came back.
func TestCleanPieces_NoRace_ConcurrentWrites(t *testing.T) {
	const pieceLen = 64 * 1024
	const numPieces = 32
	const capacity = int64(pieceLen * 4)

	tt, cache := setupCache(t, capacity, pieceLen, numPieces)

	for id := 0; id < 16; id++ {
		fillPiece(t, cache, tt, id, pieceLen)
	}
	// Let any goroutines started by WriteAt finish before t.Cleanup
	// tears the cache down (otherwise their access to a closed cache
	// would itself be racy with Close).
	time.Sleep(150 * time.Millisecond)
}



// --------------------------------------------------------------------
// Bug #4: clearPriority sleeps 1s, then clobbers priorities. If a Seek
// lands in that window, the new piece priorities get reset to None.
// --------------------------------------------------------------------

// TestClearPriority_NoArtificialSleepWindow is a regression-gate for
// bug #4. The original clearPriority did `time.Sleep(time.Second)`
// before sweeping pieces, leaving a 1s window where:
//   - any new setLoadPriority writes from a Seek could be clobbered,
//   - the goroutine itself raced with concurrent setLoadPriority for
//     the per-piece Storage.SetPriority call.
//
// The fix removes the sleep and serializes clearPriority with
// setLoadPriority via Cache.muPrio. This test asserts that after
// CloseReader, the spawned clearPriority goroutine quiesces in well
// under the old 1s window — proving the sleep is gone.
//
// If somebody re-introduces the sleep (or otherwise makes the priority
// sweep slow), this test FAILS, gating the regression.
func TestClearPriority_NoArtificialSleepWindow(t *testing.T) {
	const pieceLen = 64 * 1024
	const numPieces = 8

	tt, cache := setupCache(t, int64(pieceLen*4), pieceLen, numPieces)

	r1 := cache.NewReader(tt.Files()[0])
	r2 := cache.NewReader(tt.Files()[0])
	_ = r2

	start := time.Now()
	cache.CloseReader(r1)            // spawns clearPriority goroutine
	cache.WaitClearPriorityForTest() // blocks until it releases muPrio
	elapsed := time.Since(start)

	// With the bug present, elapsed >= ~1s (the sleep). The fix makes
	// it microseconds; allow generous slack for CI noise.
	if elapsed > 250*time.Millisecond {
		t.Fatalf("REGRESSION #4: clearPriority took %v; expected <250ms. "+
			"Has the time.Sleep been re-introduced, or is the sweep otherwise slow?",
			elapsed)
	}
	t.Logf("clearPriority quiesced in %v (well under the old 1s window)", elapsed)

	// CloseReader also spawns `go r.cache.getRemPieces()` from inside
	// Reader.Close; that goroutine is fire-and-forget and reads
	// settings.BTsets.UseDisk. If it survives into the next test which
	// also mutates UseDisk, -race reports a write/read race.
	// WaitClearPriorityForTest only awaits the clearPriority goroutine,
	// not the bare getRemPieces, so wait a touch more here to drain.
	time.Sleep(100 * time.Millisecond)
}

// --------------------------------------------------------------------
// Bug #3: readerOff seeks to 0 after 60s of inactivity.
// --------------------------------------------------------------------

// TestReaderOff_DoesNotRewindUnderlyingReader is the regression-gate
// for bug #3. The original readerOff did `r.Reader.Seek(0, SeekStart)`
// on the underlying anacrolix reader after 60s of inactivity, causing:
//   - wasteful piece-priority churn (drop pieces around real offset,
//     then re-request them when readerOn Seeks back),
//   - a transient window where underlying pos != r.offset, so any
//     direct read on the embedded Reader would have started at byte 0.
//
// The fix removes the Seek(0): readerOff now only drops readahead,
// which is sufficient to relinquish bandwidth claims without churn.
//
// This test pre-positions the underlying reader at a non-zero offset,
// triggers readerOff, and asserts the underlying position is preserved.
// If somebody re-introduces the Seek(0), this FAILS.
func TestReaderOff_DoesNotRewindUnderlyingReader(t *testing.T) {
	const pieceLen = 64 * 1024
	const numPieces = 8

	tt, cache := setupCache(t, int64(pieceLen*4), pieceLen, numPieces)

	r1 := cache.NewReader(tt.Files()[0])
	r2 := cache.NewReader(tt.Files()[0]) // need >1 readers, see checkReader
	_ = r2

	const wantOff = int64(3 * pieceLen)

	// Position both the logical mirror AND the underlying anacrolix
	// reader at wantOff. Real Seek on the underlying reader is what
	// makes a future regression observable.
	if _, err := r1.Seek(wantOff, io.SeekStart); err != nil {
		t.Fatalf("seed Seek: %v", err)
	}
	torrstor.MarkReaderUseForTest(r1, true)

	// Sanity: underlying reader is now at wantOff.
	if pos := torrstor.UnderlyingReaderPosForTest(r1); pos != wantOff {
		t.Fatalf("setup: underlying reader pos = %d, want %d", pos, wantOff)
	}

	// Trigger the "60s elapsed" path.
	torrstor.ForceReaderOffForTest(r1)

	// Logical offset must be preserved (always was, even with bug).
	if got := torrstor.ReaderOffsetForTest(r1); got != wantOff {
		t.Fatalf("readerOff lost logical offset: want %d, got %d", wantOff, got)
	}

	// REGRESSION GATE: underlying pos must NOT have been rewound to 0.
	if pos := torrstor.UnderlyingReaderPosForTest(r1); pos != wantOff {
		t.Fatalf("REGRESSION #3: readerOff rewound underlying reader to %d "+
			"(want %d). The Seek(0, SeekStart) churn is back.", pos, wantOff)
	}

	// Sanity: a fresh Seek still works after readerOff.
	got, err := r1.Seek(wantOff, io.SeekStart)
	if err != nil {
		t.Fatalf("post-readerOff Seek failed: %v", err)
	}
	if got != wantOff {
		t.Fatalf("post-readerOff Seek returned %d, want %d", got, wantOff)
	}

	// silence unused checks
	_ = errors.New
	_ = sync.Once{}
}

// --------------------------------------------------------------------
// Filled-undercount memory bloat: when many pieces are partially-written
// (multi-client / HLS chunked access pattern), MemPiece allocates a
// full-pieceLength buffer on first touch but Piece.Size reflects only
// the bytes actually written. Before the fix, getRemPieces summed Size
// into c.filled, so c.filled vastly understated real RAM usage and the
// `c.filled > c.capacity` cleanPieces trigger never fired. RSS grew
// without bound — observed as ~6GB torrserver killed by OOM under
// >2 concurrent clients on a 7.6GB host.
//
// The fix charges full pieceLength per touched piece when UseDisk=false.
// This test fills more pieces than fit in capacity with tiny partial
// writes, runs cleanPieces, and asserts the cache actually evicts down
// to (approximately) capacity. With the bug present, no eviction
// happens at all.
// --------------------------------------------------------------------

func TestCleanPieces_EvictsPartiallyFilledPieces(t *testing.T) {
	const pieceLen = 64 * 1024
	const numPieces = 32
	const capacityPieces = 4
	const capacity = int64(pieceLen * capacityPieces)
	const partialBytes = pieceLen / 10 // 6.4 KB written per touched piece
	const touchedPieces = 16           // 4x the capacity budget

	// Force UseDisk=false so MemPiece is in play (the bug only bites
	// memory mode; disk-mode Size from os.Stat reflects real on-disk
	// bytes, which is the correct accounting for that path).
	savedDisk := settings.BTsets.UseDisk
	settings.BTsets.UseDisk = false
	defer func() { settings.BTsets.UseDisk = savedDisk }()

	_, cache := setupCache(t, capacity, pieceLen, numPieces)

	// Partial-fill `touchedPieces` distinct pieces. Each WriteAt
	// allocates a full pieceLength buffer; Size only advances by
	// partialBytes. No reader is registered, so getRemPieces takes the
	// "on preload clean" branch and every touched piece is a candidate
	// for eviction.
	for id := 0; id < touchedPieces; id++ {
		fillPiecePartial(t, cache, id, partialBytes)
	}

	// Drain WriteAt-spawned cleanPieces goroutines, then force a
	// synchronous sweep so we don't race the assertion.
	time.Sleep(100 * time.Millisecond)
	cache.CleanPiecesForTest()

	// Count how many pieces are still "alive" (Size > 0).
	live := 0
	for id := 0; id < touchedPieces; id++ {
		if cache.PieceSizeForTest(id) > 0 {
			live++
		}
	}

	// With the fix, cleanPieces sees fill = touchedPieces*pieceLength
	// (1MB > capacity 256KB) and evicts down to ~capacityPieces.
	// Allow some slack: the eviction loop stops as soon as fill drops
	// under capacity, and pieceCost rounding can leave 1-2 extra alive.
	const maxAlive = capacityPieces + 2
	if live > maxAlive {
		t.Fatalf("REGRESSION (filled undercount): %d/%d partially-filled "+
			"pieces survived cleanPieces; want <=%d (capacity=%d pieces). "+
			"Has Cache.pieceCost been reverted to charging Size for "+
			"UseDisk=false? That lets RAM grow without bound under "+
			"multi-client load and trips OOM.",
			live, touchedPieces, maxAlive, capacityPieces)
	}
	t.Logf("OK: cleanPieces evicted down to %d/%d live pieces "+
		"(capacity=%d, partial=%d/%d bytes per piece)",
		live, touchedPieces, capacityPieces, partialBytes, pieceLen)
}
