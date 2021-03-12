package torrstor

import (
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
)

type Piece struct {
	storage.PieceImpl

	Id   int
	Hash string
	Size int64

	complete bool
	readed   bool
	accessed int64
	buffer   []byte

	mu    sync.RWMutex
	cache *Cache
}

func (p *Piece) WriteAt(b []byte, off int64) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.buffer == nil {
		go p.cache.cleanPieces()
		p.buffer = make([]byte, p.cache.pieceLength, p.cache.pieceLength)
	}
	n = copy(p.buffer[off:], b[:])
	//samsung tv fix xvid/divx
	if p.Id == 0 && off < 192 {
		str := strings.ToLower(string(p.buffer[112:116]))
		if str == "xvid" || str == "divx" {
			p.buffer[112] = 0x4D //M
			p.buffer[113] = 0x50 //P
			p.buffer[114] = 0x34 //4
			p.buffer[115] = 0x56 //V
		}
		str = strings.ToLower(string(p.buffer[188:192]))
		if str == "xvid" || str == "divx" {
			p.buffer[188] = 0x4D //M
			p.buffer[189] = 0x50 //P
			p.buffer[190] = 0x34 //4
			p.buffer[191] = 0x56 //V
		}

		println(string(p.buffer[110:192]))
	}
	p.Size += int64(n)
	p.accessed = time.Now().Unix()
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
		return 0, io.EOF
	}
	n = copy(b, p.buffer[int(off) : int(off)+size][:])
	p.accessed = time.Now().Unix()
	if int(off)+size >= len(p.buffer) {
		p.readed = true
	}
	if int64(len(b))+off >= p.Size {
		go p.cache.cleanPieces()
	}
	if n == 0 && err == nil {
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
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.buffer != nil {
		p.buffer = nil
	}
	p.Size = 0
	p.complete = false

	p.cache.torrent.Piece(p.Id).SetPriority(torrent.PiecePriorityNone)
}
