package torrstor

import (
    "bytes"
    "testing"
)

func TestPartialPieceRAMAccounting(t *testing.T) {
    c := NewCache(96<<20, nil)
    c.pieceLength = 8 << 20
    c.pieceCount = 128
    for i := 0; i < c.pieceCount; i++ {
        p := &Piece{Id: i, cache: c}
        p.mPiece = NewMemPiece(p)
        c.pieces[i] = p
    }

    // Serialize asynchronous cleanup during this accounting probe.
    c.muRemove.Lock()
    defer c.muRemove.Unlock()
    defer func() {
        for _, p := range c.pieces {
            p.mPiece.Release()
        }
    }()

    block := bytes.Repeat([]byte{0x5a}, 16<<10)
    for _, p := range c.pieces {
        if _, err := p.mPiece.WriteAt(block, 0); err != nil {
            t.Fatal(err)
        }
    }
    c.getRemPieces() // recomputes the value used for eviction

    var allocated int64
    for _, p := range c.pieces {
        allocated += int64(cap(p.mPiece.buffer))
    }
    t.Logf("capacity=%d MiB, filled=%d MiB, allocated=%d MiB",
        c.capacity>>20, c.filled>>20, allocated>>20)
    if c.filled < allocated {
        t.Errorf("eviction accounts for %d bytes, but RAM buffers hold %d bytes",
            c.filled, allocated)
    }
}