package torrstor

import (
	"bytes"
	"sync"
	"sync/atomic"
	"testing"
)

func testMemCache(n int64) (*Cache, *Piece, *MemPiece) {
	c := &Cache{
		capacity:     8 << 20,
		pieceLength:  n,
		pieceCount:   1,
		totalLength:  n,
		activePieces: map[int]struct{}{},
		readers:      map[*Reader]struct{}{},
		pieces:       map[int]*Piece{},
	}
	p := &Piece{Id: 0, cache: c}
	mp := NewMemPiece(p)
	p.mPiece = mp
	c.pieces[0] = p
	return c, p, mp
}

func TestPieceBufRecycledAndZeroed(t *testing.T) {
	const n = 1237
	_, _, mp := testMemCache(n)
	junk := bytes.Repeat([]byte{0xAB}, n)
	if _, err := mp.WriteAt(junk, 0); err != nil {
		t.Fatal(err)
	}
	ptr := &mp.buffer[0]
	mp.Release()
	if mp.buffer != nil {
		t.Fatal("release left the piece buffer installed")
	}
	if _, err := mp.WriteAt([]byte{0x01}, 0); err != nil {
		t.Fatal(err)
	}
	if len(mp.buffer) == 0 || &mp.buffer[0] != ptr {
		t.Fatal("evicted piece buffer was not recycled")
	}
	got := make([]byte, n)
	if _, err := mp.ReadAt(got, 0); err != nil {
		t.Fatal(err)
	}
	if got[0] != 0x01 {
		t.Fatalf("head %x", got[0])
	}
	for i, b := range got[1:] {
		if b != 0 {
			t.Fatalf("stale byte at %d: %x", i+1, b)
		}
	}
}

func TestMemPieceReadDuringEvict(t *testing.T) {
	const n = 4096
	_, p, mp := testMemCache(n)
	payload := bytes.Repeat([]byte{0x5A}, n)
	if _, err := mp.WriteAt(payload, 0); err != nil {
		t.Fatal(err)
	}
	if err := p.MarkComplete(); err != nil {
		t.Fatal(err)
	}

	var torn atomic.Bool
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, 128)
			for {
				select {
				case <-stop:
					return
				default:
				}
				rn, err := mp.ReadAt(buf, 0)
				if err != nil {
					continue
				}
				for _, b := range buf[:rn] {
					if b != 0x5A {
						torn.Store(true)
						return
					}
				}
			}
		}()
	}
	for i := 0; i < 400; i++ {
		mp.Release()
		if _, err := mp.WriteAt(payload, 0); err != nil {
			t.Fatal(err)
		}
		if err := p.MarkComplete(); err != nil {
			t.Fatal(err)
		}
	}
	close(stop)
	wg.Wait()
	if torn.Load() {
		t.Fatal("read observed torn or recycled piece bytes")
	}
}
