package memcacheV2

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
	accessed int64
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
	p.accessed = time.Now().Unix() + 2000
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

	if int64(len(b))+off >= p.Size {
		go p.cache.cleanPieces()
		time.Sleep(time.Millisecond * 2000)
	}

	if p.complete {
		p.accessed = time.Now().Unix()
		p.cache.SetPos(p.Id)
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
	p.accessed = 0
	return nil
}

func (p *Piece) Completion() storage.Completion {
	return storage.Completion{
		Complete: p.complete,
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
