package torrstor

import (
	"errors"
	"io"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
)

type Piece struct {
	storage.PieceImpl

	Id   int
	Size int64

	complete bool
	accessed int64
	buffer   []byte

	cache *Cache
}

func (p *Piece) WriteAt(b []byte, off int64) (n int, err error) {
	if p.buffer == nil {
		go p.cache.cleanPieces()
		p.buffer = make([]byte, p.cache.pieceLength, p.cache.pieceLength)
	}
	n = copy(p.buffer[off:], b[:])
	p.Size += int64(n)
	p.accessed = time.Now().Unix()
	return
}

func (p *Piece) ReadAt(b []byte, off int64) (n int, err error) {
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
	n = copy(b, p.buffer[int(off) : int(off)+size][:])
	p.accessed = time.Now().Unix()
	if int64(len(b))+off >= p.Size {
		go p.cache.cleanPieces()
	}
	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}

func (p *Piece) MarkComplete() error {
	if len(p.buffer) == 0 {
		return errors.New("piece is not complete")
	}
	p.complete = true
	return nil
}

func (p *Piece) MarkNotComplete() error {
	p.complete = false
	return nil
}

func (p *Piece) Completion() storage.Completion {
	return storage.Completion{
		Complete: p.complete && len(p.buffer) > 0,
		Ok:       true,
	}
}

func (p *Piece) Release() {
	if p.buffer != nil {
		p.buffer = nil
	}
	p.Size = 0
	p.complete = false

	p.cache.torrent.Piece(p.Id).SetPriority(torrent.PiecePriorityNone)
}
