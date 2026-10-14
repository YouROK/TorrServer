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
		p.Size = ff.Size()
		p.Complete = ff.Size() == p.cache.pieceSize(p.Id)
		p.Accessed = ff.ModTime().Unix()
	}
	return &DiskPiece{piece: p, name: name}
}

func (p *DiskPiece) WriteAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ff, err := os.OpenFile(p.name, os.O_RDWR|os.O_CREATE, 0o666)
	if err != nil {
		log.TLogln("Error open file:", err)
		return 0, err
	}
	defer ff.Close()
	n, err = ff.WriteAt(b, off)

	before := p.piece.Size
	p.piece.Size += int64(n)
	if p.piece.Size > p.piece.cache.pieceLength {
		p.piece.Size = p.piece.cache.pieceLength
	}
	addDiskCacheUsed(p.piece.Size - before)
	p.piece.Accessed = time.Now().Unix()
	// the write itself bumps the file mtime that global eviction sorts by;
	// check the quota once per piece, not on every chunk
	if full := p.piece.cache.pieceSize(p.piece.Id); IsLongTermCache() && before < full && p.piece.Size >= full {
		go p.piece.cache.storage.cleanGlobal()
	}
	return
}

func (p *DiskPiece) ReadAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ff, err := os.OpenFile(p.name, os.O_RDONLY, 0o666)
	if os.IsNotExist(err) {
		return 0, io.EOF
	}
	if err != nil {
		log.TLogln("Error open file:", err)
		return 0, err
	}
	defer ff.Close()

	n, err = ff.ReadAt(b, off)

	now := time.Now()
	// global eviction sorts by file mtime; refresh it at most once a minute per piece
	// instead of issuing a syscall on every read
	if IsLongTermCache() && now.Unix()-p.piece.Accessed >= 60 {
		_ = os.Chtimes(p.name, now, now)
	}
	p.piece.Accessed = now.Unix()
	if int64(len(b))+off >= p.piece.Size {
		go p.piece.cache.cleanPieces()
	}
	return n, err
}

func (p *DiskPiece) Release() {
	p.mu.Lock()
	defer p.mu.Unlock()

	size := p.piece.Size
	p.piece.Size = 0
	p.piece.Complete = false

	if err := os.Remove(p.name); err == nil {
		addDiskCacheUsed(-size)
	}
}

// discardFailed removes a piece file that failed the hash check. The client
// downloads all of its chunks again, but a full-size file left behind would be
// taken for a complete piece by NewDiskPiece after a restart and never rechecked.
// A partly written file is kept: its chunks are still marked as received.
func (p *DiskPiece) discardFailed() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.piece.Size < p.piece.cache.pieceSize(p.piece.Id) {
		return
	}
	err := os.Remove(p.name)
	if err != nil && !os.IsNotExist(err) {
		// Windows refuses to remove a file that is open elsewhere
		err = os.Truncate(p.name, 0)
	}
	if err != nil && !os.IsNotExist(err) {
		log.TLogln("Error remove failed piece:", err)
		return
	}
	addDiskCacheUsed(-p.piece.Size)
	p.piece.Size = 0
}
