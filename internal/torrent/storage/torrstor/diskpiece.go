package torrstor

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"silo/internal/log"
)

type DiskPiece struct {
	piece *Piece
	name  string
	mu    sync.RWMutex
}

func NewDiskPiece(p *Piece) *DiskPiece {
	cfg := p.cache.storage.cfg
	name := filepath.Join(cfg.TorrentsSavePath, p.cache.hash.HexString(), strconv.Itoa(p.Id))
	if ff, err := os.Stat(name); err == nil {
		p.Size = ff.Size()
		p.Complete = ff.Size() == p.cache.pieceLength
		p.Accessed = ff.ModTime().Unix()
	}
	return &DiskPiece{piece: p, name: name}
}

func (p *DiskPiece) WriteAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ff, err := os.OpenFile(p.name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf("[TorrStor] Error opening disk piece file: %v", err)
		return 0, err
	}
	defer ff.Close()

	n, err = ff.WriteAt(b, off)

	p.piece.Size += int64(n)
	if p.piece.Size > p.piece.cache.pieceLength {
		p.piece.Size = p.piece.cache.pieceLength
	}
	p.piece.Accessed = time.Now().Unix()
	return
}

func (p *DiskPiece) ReadAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ff, err := os.OpenFile(p.name, os.O_RDONLY, 0666)
	if os.IsNotExist(err) {
		return 0, io.EOF
	}
	if err != nil {
		log.Errorf("[TorrStor] Error opening disk piece file: %v", err)
		return 0, err
	}
	defer ff.Close()

	n, err = ff.ReadAt(b, off)

	p.piece.Accessed = time.Now().Unix()
	if int64(len(b))+off >= p.piece.Size {
		go p.piece.cache.cleanPieces()
	}
	return n, nil
}

func (p *DiskPiece) Release() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.piece.Size = 0
	p.piece.Complete = false

	_ = os.Remove(p.name)
}
