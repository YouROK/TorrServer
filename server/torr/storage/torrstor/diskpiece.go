package torrstor

import (
	"io"
	"sync"
	"time"
)

type DiskPiece struct {
	piece *Piece

	mu sync.RWMutex
}

func NewDiskPiece(p *Piece) *DiskPiece {
	return &DiskPiece{piece: p}
}

func (p *DiskPiece) WriteAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := int64(p.piece.Id)
	pl := p.piece.cache.pieceLength
	poff := id * pl
	off += poff

	n = 0
	off, err = p.piece.cache.file.Seek(off, io.SeekStart)
	if err == nil {
		n, err = p.piece.cache.file.Write(b)
	}

	go p.piece.cache.loadPieces()
	p.piece.cache.saveInfo()

	p.piece.Size += int64(n)
	p.piece.Accessed = time.Now().Unix()
	return
}

func (p *DiskPiece) ReadAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := int64(p.piece.Id)
	pl := p.piece.cache.pieceLength
	poff := id * pl
	off += poff

	n = 0
	off, err = p.piece.cache.file.Seek(off, io.SeekStart)
	if err == nil {
		n, err = p.piece.cache.file.Read(b)
	}

	p.piece.Accessed = time.Now().Unix()
	return n, nil
}

func (p *DiskPiece) Release() {

}
