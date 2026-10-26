package torrstor

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
)

func TestWriteAtReadAtConcurrentWithClose(t *testing.T) {
	prev := settings.BTsets
	settings.BTsets = &settings.BTSets{}
	t.Cleanup(func() { settings.BTsets = prev })

	stor := NewStorage(1 << 20)
	c := NewCache(1<<20, stor)
	c.pieceLength = 1024
	c.pieceCount = 1
	p := &Piece{Id: 0, cache: c}
	mp := NewMemPiece(p)
	p.mPiece = mp
	c.pieces = map[int]*Piece{0: p}

	var wg sync.WaitGroup
	const writers = 8
	wg.Add(writers + 1)
	for i := 0; i < writers; i++ {
		go func() {
			defer wg.Done()
			buf := make([]byte, 64)
			for j := 0; j < 200; j++ {
				_, _ = mp.WriteAt(buf, 0)
				_, _ = mp.ReadAt(buf, 0)
			}
		}()
	}
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond)
		_ = c.Close()
	}()
	wg.Wait()

	if mp.buffer != nil {
		t.Fatal("after Close buffer must be nil")
	}
	n, err := mp.WriteAt([]byte("x"), 0)
	if n != 0 || !os.IsNotExist(err) {
		t.Fatalf("post-Close WriteAt: n=%d err=%v", n, err)
	}
}

func TestGetStateGetRemPiecesConcurrentWithClose(t *testing.T) {
	initTestSettings(t)

	var wg sync.WaitGroup
	const iters = 200
	for i := 0; i < iters; i++ {
		stor := NewStorage(1 << 20)
		var h metainfo.Hash
		h[0] = byte(i)
		_, _ = stor.OpenTorrent(testInfo(), h)
		cache := stor.GetCache(h)
		if cache == nil {
			t.Fatal("cache nil")
		}
		if p := cache.pieces[0]; p != nil && p.mPiece != nil {
			p.mPiece.buffer = make([]byte, cache.pieceLength)
			p.Size = cache.pieceLength
			cache.notePieceFilled(0)
		}

		wg.Add(3)
		go func() {
			defer wg.Done()
			_ = cache.GetState()
		}()
		go func() {
			defer wg.Done()
			_ = cache.getRemPieces()
		}()
		go func() {
			defer wg.Done()
			_ = cache.Close()
		}()
		wg.Wait()
	}
}

func TestWriteAtAfterCloseAllocs(t *testing.T) {
	prev := settings.BTsets
	settings.BTsets = &settings.BTSets{}
	t.Cleanup(func() { settings.BTsets = prev })

	stor := NewStorage(1 << 20)
	c := NewCache(1<<20, stor)
	c.pieceLength = 1 << 20
	c.pieceCount = 1
	p := &Piece{Id: 0, cache: c}
	mp := NewMemPiece(p)
	p.mPiece = mp
	c.pieces = map[int]*Piece{0: p}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	payload := []byte("x")
	allocs := testing.AllocsPerRun(1000, func() {
		_, _ = mp.WriteAt(payload, 0)
	})
	if allocs != 0 {
		t.Fatalf("WriteAt after Close must not allocate, got %v allocs/run", allocs)
	}
	if mp.buffer != nil {
		t.Fatal("buffer must stay nil")
	}
}

func BenchmarkWriteAtRejectedWhenClosed(b *testing.B) {
	prev := settings.BTsets
	settings.BTsets = &settings.BTSets{}
	b.Cleanup(func() { settings.BTsets = prev })

	stor := NewStorage(1 << 20)
	c := NewCache(1<<20, stor)
	c.pieceLength = 1 << 20
	c.pieceCount = 1
	p := &Piece{Id: 0, cache: c}
	mp := NewMemPiece(p)
	p.mPiece = mp
	c.pieces = map[int]*Piece{0: p}
	_ = c.Close()

	payload := []byte("x")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mp.WriteAt(payload, 0)
	}
}

func TestMarkCompleteConcurrentWithGetState(t *testing.T) {
	initTestSettings(t)

	stor := NewStorage(1 << 20)
	c := NewCache(1<<20, stor)
	c.pieceLength = 1024
	c.pieceCount = 1
	p := &Piece{Id: 0, cache: c, Size: 1024}
	mp := NewMemPiece(p)
	mp.buffer = make([]byte, 1024)
	p.mPiece = mp
	c.pieces = map[int]*Piece{0: p}
	c.notePieceFilled(0)

	var wg sync.WaitGroup
	const iters = 500
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			_ = p.MarkComplete()
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			_ = c.GetState()
			_ = c.getRemPieces()
		}
	}()
	wg.Wait()
}

func TestGetRemPiecesNoopAfterClose(t *testing.T) {
	prev := settings.BTsets
	settings.BTsets = &settings.BTSets{}
	t.Cleanup(func() { settings.BTsets = prev })

	stor := NewStorage(1 << 20)
	c := NewCache(1<<20, stor)
	c.Init(&metainfo.Info{
		Name:        "wd",
		PieceLength: 1 << 10,
		Length:      1 << 10,
		Pieces:      make([]byte, 20),
	}, metainfo.Hash{})

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if !c.isClosed.Load() {
		t.Fatal("expected isClosed after Close")
	}
	if rem := c.getRemPieces(); rem != nil {
		t.Fatalf("getRemPieces after Close: want nil, got %d", len(rem))
	}
	// Watchdog loop checks isClosed; second Close must stay idempotent.
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCompletionConcurrentWithMarkComplete(t *testing.T) {
	initTestSettings(t)

	c := &Cache{pieceLength: 1024, pieceCount: 1, activePieces: make(map[int]struct{})}
	p := &Piece{Id: 0, cache: c, Size: 1024}
	mp := NewMemPiece(p)
	mp.buffer = make([]byte, 1024)
	p.mPiece = mp

	var wg sync.WaitGroup
	const iters = 1000
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			_ = p.MarkComplete()
			_ = p.MarkNotComplete()
			mp.mu.Lock()
			mp.buffer = make([]byte, 1024)
			p.Size = 1024
			mp.mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			_ = p.Completion()
		}
	}()
	wg.Wait()
}

// cleanPieces must drop muRemove before UpdateCompletion. Client.Close holds the
// client lock then Cache.Close waits on muRemove — the preload teardown hang.
func TestCleanPiecesUnlocksBeforeCompletion(t *testing.T) {
	prev := settings.BTsets
	settings.BTsets = &settings.BTSets{}
	t.Cleanup(func() { settings.BTsets = prev })

	var clientMu sync.Mutex
	old := pieceCompletionUpdater
	pieceCompletionUpdater = func(*Piece) {
		clientMu.Lock()
		clientMu.Unlock()
	}
	t.Cleanup(func() { pieceCompletionUpdater = old })

	const plen int64 = 8 << 20
	stor := NewStorage(plen)
	c := NewCache(plen, stor) // capacity = one piece; three filled → eviction
	c.pieceLength = plen
	c.pieceCount = 3
	c.pieces = make(map[int]*Piece, 3)
	c.activePieces = make(map[int]struct{})
	for id := 0; id < 3; id++ {
		p := &Piece{Id: id, cache: c, Size: plen, Complete: true, Accessed: int64(id)}
		mp := NewMemPiece(p)
		mp.buffer = make([]byte, plen)
		p.mPiece = mp
		c.pieces[id] = p
		c.notePieceFilled(id)
	}
	c.filled.Store(3 * plen)

	done := make(chan struct{})
	go func() {
		defer close(done)
		clientMu.Lock()
		evictDone := make(chan struct{})
		go func() {
			defer close(evictDone)
			c.cleanPieces()
		}()
		// Old bug: cleanPieces held muRemove across completion (waits clientMu),
		// Close blocked on muRemove → deadlock. Sleep so eviction reaches completion.
		time.Sleep(20 * time.Millisecond)
		_ = c.Close()
		clientMu.Unlock()
		<-evictDone // before test cleanup clears pieceCompletionUpdater
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("deadlock: muRemove held across completion while Close held client lock")
	}
}

// Close holds muRemove → piece.mu → muActive; cleanPieces TryLock(muRemove).
// Hang = lock-order regression.
func TestCloseCleanPiecesNoDeadlock(t *testing.T) {
	prev := settings.BTsets
	settings.BTsets = &settings.BTSets{}
	t.Cleanup(func() { settings.BTsets = prev })

	done := make(chan struct{})
	go func() {
		defer close(done)
		const rounds = 100
		for i := 0; i < rounds; i++ {
			stor := NewStorage(4 << 10)
			c := NewCache(4<<10, stor) // small capacity → cleanPieces evicts
			c.pieceLength = 1024
			c.pieceCount = 8
			c.pieces = make(map[int]*Piece, 8)
			c.activePieces = make(map[int]struct{})
			// Fill map before any WriteAt (WriteAt may spawn cleanPieces).
			for id := 0; id < 8; id++ {
				p := &Piece{Id: id, cache: c}
				mp := NewMemPiece(p)
				mp.buffer = make([]byte, 1024)
				p.mPiece = mp
				p.Size = 1024
				c.pieces[id] = p
				c.notePieceFilled(id)
			}
			c.filled.Store(8 * 1024)

			var wg sync.WaitGroup
			wg.Add(4)
			go func() {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					c.cleanPieces()
				}
			}()
			go func() {
				defer wg.Done()
				buf := make([]byte, 64)
				for j := 0; j < 50; j++ {
					_, _ = c.pieces[j%8].WriteAt(buf, 0)
					_ = c.GetState()
				}
			}()
			go func() {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					c.cleanPieces()
					_ = c.getRemPieces()
				}
			}()
			go func() {
				defer wg.Done()
				time.Sleep(time.Millisecond)
				_ = c.Close()
			}()
			wg.Wait()
		}
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("deadlock: Close vs cleanPieces exceeded 30s")
	}
}
