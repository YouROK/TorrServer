package torrstor

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"server/log"
	"server/settings"
)

type DiskPiece struct {
	piece *Piece

	name string

	mu sync.RWMutex
}

func NewDiskPiece(p *Piece) *DiskPiece {
	name := filepath.Join(settings.BTsets.TorrentsSavePath, p.cache.hash.HexString(), strconv.Itoa(p.Id))
	ff, err := os.Stat(name)
	if err == nil {
		p.Size.Store(ff.Size())
		p.Complete.Store(ff.Size() == p.cache.pieceLength)
		p.Accessed.Store(ff.ModTime().Unix())
	}
	return &DiskPiece{piece: p, name: name}
}

func (p *DiskPiece) WriteAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ff, err := os.OpenFile(p.name, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		log.TLogln("Error open file:", err)
		return 0, err
	}
	defer ff.Close()
	n, err = ff.WriteAt(b, off)

	p.piece.addSize(int64(n))
	p.piece.Accessed.Store(time.Now().Unix())
	return
}

func (p *DiskPiece) ReadAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ff, err := os.OpenFile(p.name, os.O_RDONLY, 0o644)
	if os.IsNotExist(err) {
		return 0, io.EOF
	}
	if err != nil {
		log.TLogln("Error open file:", err)
		return 0, err
	}
	defer ff.Close()

	n, err = ff.ReadAt(b, off)

	p.piece.Accessed.Store(time.Now().Unix())
	if int64(len(b))+off >= p.piece.Size.Load() {
		go p.piece.cache.cleanPieces()
	}
	return n, nil
}

func (p *DiskPiece) Release() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.piece.Size.Store(0)
	p.piece.Complete.Store(false)

	os.Remove(p.name)
}
