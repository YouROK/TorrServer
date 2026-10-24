package torrstor

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/anacrolix/torrent/metainfo"

	"server/leakcheck"
	"server/settings"
)

// 50MB through a 4MB cache. The start/end keep window is 8MB per side
// (isIdInKeepWindow), so live buffers must stay near that window, not the
// bytes written.
func TestPieceCacheEvictionHeapCeiling(t *testing.T) {
	prev := settings.BTsets
	settings.BTsets = &settings.BTSets{UseDisk: false}
	t.Cleanup(func() { settings.BTsets = prev })

	const (
		pieceLen = 64 << 10
		capacity = 4 << 20
		total    = 50 << 20
	)
	nPieces := total / pieceLen

	baseG := runtime.NumGoroutine()
	before := leakcheck.Heap()

	stor := NewStorage(capacity)
	info := &metainfo.Info{
		Name:        "mem",
		PieceLength: pieceLen,
		Length:      total,
		Pieces:      make([]byte, 20*nPieces),
	}
	var h metainfo.Hash
	h[0] = 0x4d
	if _, err := stor.OpenTorrent(info, h); err != nil {
		t.Fatal(err)
	}
	c := stor.GetCache(h)
	if c == nil || c.pieceCount != nPieces {
		t.Fatalf("cache pieces %v", c)
	}

	payload := make([]byte, pieceLen)
	for i := range payload {
		payload[i] = byte(i)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for i := 0; i < nPieces; i++ {
			p := c.pieces[i%nPieces]
			if p == nil {
				continue
			}
			_, _ = p.ReadAt(buf, int64(i%1024))
		}
	}()
	for id := 0; id < nPieces; id++ {
		p := c.pieces[id]
		if _, err := p.WriteAt(payload, 0); err != nil {
			t.Fatal(err)
		}
		if err := p.MarkComplete(); err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()

	keep := c.pieceLength
	if keep < 8<<20 {
		keep = 8 << 20
	}
	se := int(keep / c.pieceLength)
	if se < 1 {
		se = 1
	}
	ceiling := int64(se*2)*c.pieceLength + capacity + c.pieceLength

	deadline := time.Now().Add(5 * time.Second)
	var liveN int
	var liveB int64
	for {
		c.cleanPieces()
		liveN, liveB = livePieceBuffers(c)
		if liveB <= ceiling || time.Now().After(deadline) {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	// Refresh the filled counter, which cleanPieces samples before it releases.
	_ = c.getRemPieces()
	if liveB > ceiling {
		t.Fatalf("live piece buffers %d bytes (%d pieces) over ceiling %d after writing %d", liveB, liveN, ceiling, total)
	}
	if liveB >= total/2 {
		t.Fatalf("eviction did not turn the cache over: live %d of %d written", liveB, total)
	}
	if filled := c.FilledBytes(); filled > ceiling {
		t.Fatalf("filled counter %d over ceiling %d", filled, ceiling)
	}
	c.muActive.Lock()
	active := len(c.activePieces)
	c.muActive.Unlock()
	if active > liveN+2 {
		t.Fatalf("active piece map %d > live buffers %d", active, liveN)
	}
	if c.Readers() != 0 {
		t.Fatalf("readers still registered: %d", c.Readers())
	}

	if err := stor.Close(); err != nil {
		t.Fatal(err)
	}
	if n, b := livePieceBuffers(c); n != 0 || b != 0 {
		t.Fatalf("closed cache retained %d buffers (%d bytes)", n, b)
	}

	after := leakcheck.Heap()
	const heapSlack = 16 << 20
	if after.HeapAlloc > before.HeapAlloc+heapSlack || after.HeapInuse > before.HeapInuse+heapSlack {
		t.Fatalf("heap alloc %d->%d inuse %d->%d after releasing %d of piece data",
			before.HeapAlloc, after.HeapAlloc, before.HeapInuse, after.HeapInuse, total)
	}
	leakcheck.AssertNoGoroutineLeak(t, baseG)
}

func livePieceBuffers(c *Cache) (n int, bytes int64) {
	for _, p := range c.pieces {
		if p == nil || p.mPiece == nil {
			continue
		}
		p.mPiece.mu.Lock()
		if p.mPiece.buffer != nil {
			n++
			bytes += int64(len(p.mPiece.buffer))
		}
		p.mPiece.mu.Unlock()
	}
	return n, bytes
}
