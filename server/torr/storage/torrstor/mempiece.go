package torrstor

import (
	"io"
	"os"
	"sync"
	"time"
)

type MemPiece struct {
	piece *Piece

	buffer []byte
	slab   *pieceSlab
	mu     sync.RWMutex
}

func NewMemPiece(p *Piece) *MemPiece {
	return &MemPiece{piece: p}
}

func (p *MemPiece) WriteAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// isClosed under piece mu: Close sets isClosed before Release (same mu),
	// so we either finish then Release nils buffer, or see closed and skip make.
	if p.piece.cache.isClosed.Load() {
		return 0, os.ErrNotExist
	}
	if p.buffer == nil {
		go p.piece.cache.cleanPieces()
		n := int(p.piece.cache.pieceLength)
		if n <= 0 {
			p.buffer = make([]byte, p.piece.cache.pieceLength)
		} else {
			p.buffer, p.slab = getPieceBuf(n)
		}
	}
	n = copy(p.buffer[off:], b[:])
	// Accumulate: chunks arrive out of order, so a high-water mark would count
	// gaps as filled. MarkNotComplete resets Size on retransmit.
	p.piece.Size += int64(n)
	plen := p.piece.cache.pieceByteLength(p.piece.Id)
	if plen > 0 && p.piece.Size > plen {
		p.piece.Size = plen
	}
	if p.piece.Size > 0 {
		p.piece.cache.notePieceFilled(p.piece.Id)
	}
	p.piece.Accessed = time.Now().Unix()
	return
}

func (p *MemPiece) ReadAt(b []byte, off int64) (n int, err error) {
	// Lock (not RLock): we write Accessed; concurrent ReadAt under RLock races.
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.piece.cache.isClosed.Load() || p.buffer == nil {
		// Closed / released while hashing: os.IsNotExist is the expected anacrolix path
		// (io.Copy turns io.EOF into err=nil and logs "unexpected error hashing").
		return 0, os.ErrNotExist
	}

	size := len(b)
	if size+int(off) > len(p.buffer) {
		size = len(p.buffer) - int(off)
		if size < 0 {
			size = 0
		}
	}
	if len(p.buffer) < int(off) || len(p.buffer) < int(off)+size {
		return 0, io.EOF
	}
	n = copy(b, p.buffer[int(off):int(off)+size])
	p.piece.Accessed = time.Now().Unix()
	// Do not evict while anacrolix is hashing (Complete is still false).
	if p.piece.Complete && int64(len(b))+off >= p.piece.Size {
		go p.piece.cache.cleanPieces()
	}
	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}

func (p *MemPiece) Release() {
	p.mu.Lock()
	// Nil the slice before unlock so a ReadAt that was waiting cannot observe
	// it, then recycle only after unlock so that ReadAt's copy happens-before
	// the buffer is cleared and published to the pool.
	slab := p.slab
	p.buffer = nil
	p.slab = nil
	p.piece.Size = 0
	p.piece.Complete = false
	// Closed torrents drop the bytes instead of parking them in the pool,
	// so FreeOSMem after Close can return the cache to the OS.
	closed := p.piece.cache.isClosed.Load()
	p.piece.cache.notePieceEmpty(p.piece.Id)
	p.mu.Unlock()
	if !closed {
		putPieceBuf(slab)
	}
}
