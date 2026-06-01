package torrstor

import (
	"sync/atomic"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"server/settings"
)

// Piece's mutable hot fields (Size, Accessed, Complete) are written by
// torrent-package writer goroutines (via WriteAt / Release) and read
// concurrently by Cache bookkeeping (getRemPieces, GetState,
// setLoadPriority). They were plain fields and that was racy. They are
// now atomics; helper accessors hide the boilerplate so call sites
// stay readable.
type Piece struct {
	storage.PieceImpl `json:"-"`

	Id   int          `json:"-"`
	Size atomic.Int64 `json:"-"`

	Complete atomic.Bool  `json:"-"`
	Accessed atomic.Int64 `json:"-"`

	mPiece *MemPiece  `json:"-"`
	dPiece *DiskPiece `json:"-"`

	cache *Cache `json:"-"`
}

// addSize bumps Size by n, clamped to pieceLength. Returns the new size.
func (p *Piece) addSize(n int64) int64 {
	for {
		cur := p.Size.Load()
		next := cur + n
		if next > p.cache.pieceLength {
			next = p.cache.pieceLength
		}
		if p.Size.CompareAndSwap(cur, next) {
			return next
		}
	}
}

func NewPiece(id int, cache *Cache) *Piece {
	p := &Piece{
		Id:    id,
		cache: cache,
	}

	if !settings.BTsets.UseDisk {
		p.mPiece = NewMemPiece(p)
	} else {
		p.dPiece = NewDiskPiece(p)
	}
	return p
}

func (p *Piece) WriteAt(b []byte, off int64) (n int, err error) {
	if !settings.BTsets.UseDisk {
		return p.mPiece.WriteAt(b, off)
	} else {
		return p.dPiece.WriteAt(b, off)
	}
}

func (p *Piece) ReadAt(b []byte, off int64) (n int, err error) {
	if !settings.BTsets.UseDisk {
		return p.mPiece.ReadAt(b, off)
	} else {
		return p.dPiece.ReadAt(b, off)
	}
}

func (p *Piece) MarkComplete() error {
	p.Complete.Store(true)
	return nil
}

func (p *Piece) MarkNotComplete() error {
	p.Complete.Store(false)
	return nil
}

func (p *Piece) Completion() storage.Completion {
	return storage.Completion{
		Complete: p.Complete.Load(),
		Ok:       true,
	}
}

func (p *Piece) Release() {
	if !settings.BTsets.UseDisk {
		p.mPiece.Release()
	} else {
		p.dPiece.Release()
	}
	if !p.cache.isClosed.Load() && p.cache.torrent != nil {
		p.cache.torrent.Piece(p.Id).SetPriority(torrent.PiecePriorityNone)
		p.cache.torrent.Piece(p.Id).UpdateCompletion()
	}
}
