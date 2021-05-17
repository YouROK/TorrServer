package torrstor

import (
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

	file *os.File

	mu sync.RWMutex
}

func NewDiskPiece(p *Piece) *DiskPiece {
	name := filepath.Join(settings.BTsets.TorrentsSavePath, p.cache.hash.HexString(), strconv.Itoa(p.Id))
	ff, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.TLogln("Error open file:", err)
		return nil
	}
	return &DiskPiece{piece: p, file: ff}
}

func (p *DiskPiece) WriteAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	n, err = p.file.WriteAt(b, off)

	go p.piece.cache.LoadPiecesOnDisk()

	p.piece.Size += int64(n)
	p.piece.Accessed = time.Now().Unix()
	return
}

func (p *DiskPiece) ReadAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	n, err = p.file.ReadAt(b, off)

	p.piece.Accessed = time.Now().Unix()
	return n, nil
}

func (p *DiskPiece) Release() {
	p.file.Close()
}
