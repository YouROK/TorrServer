package torrstor

import (
	"context"
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
		p.AddSize(ff.Size())
		p.SetComplete(ff.Size() == p.cache.pieceLength)
		p.SetAccessed(ff.ModTime().Unix())
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

	p.piece.AddSize(int64(n))
	p.piece.SetAccessed(time.Now().Unix())
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

	p.piece.SetAccessed(time.Now().Unix())
	if int64(len(b))+off >= p.piece.SizeOf() {
		p.piece.cache.tasks.Go(func(ctx context.Context) { p.piece.cache.cleanPieces() })
	}
	return n, nil
}

func (p *DiskPiece) Release() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.piece.Reset()

	_ = os.Remove(p.name)
}
