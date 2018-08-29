package memcache

import (
	"errors"
	"io"
	"sync"
	"time"

	"server/torr/storage/state"

	"github.com/anacrolix/torrent/storage"
)

type Piece struct {
	storage.PieceImpl

	Id     int
	Hash   string
	Length int64
	Size   int64

	complete bool
	readed   bool
	accessed time.Time
	buffer   []byte
	bufIndex int

	mu    sync.RWMutex
	cache *Cache
}

func (p *Piece) WriteAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.buffer == nil {
		go p.cache.cleanPieces()
		p.buffer, p.bufIndex = p.cache.bufferPull.GetBuffer(p)
		if p.buffer == nil {
			return 0, errors.New("Can't get buffer write")
		}
	}
	n = copy(p.buffer[off:], b[:])
	p.Size += int64(n)
	p.accessed = time.Now()
	return
}

func (p *Piece) ReadAt(b []byte, off int64) (n int, err error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	size := len(b)
	if size+int(off) > len(p.buffer) {
		size = len(p.buffer) - int(off)
		if size < 0 {
			size = 0
		}
	}
	if len(p.buffer) < int(off) || len(p.buffer) < int(off)+size {
		return 0, io.ErrUnexpectedEOF
	}
	n = copy(b, p.buffer[int(off) : int(off)+size][:])
	p.accessed = time.Now()
	if int(off)+size >= len(p.buffer) {
		p.readed = true
	}
	if int64(len(b))+off >= p.Size {
		go p.cache.cleanPieces()
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
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.buffer != nil {
		p.buffer = nil
		p.cache.bufferPull.ReleaseBuffer(p.bufIndex)
		p.bufIndex = -1
	}
	p.Size = 0
	p.complete = false
}

func (p *Piece) Stat() state.ItemState {
	itm := state.ItemState{
		Id:         p.Id,
		Hash:       p.Hash,
		Accessed:   p.accessed,
		Completed:  p.complete,
		BufferSize: p.Size,
	}
	return itm
}
