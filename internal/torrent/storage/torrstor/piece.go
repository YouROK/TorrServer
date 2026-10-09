package torrstor

import (
	"github.com/anacrolix/torrent/storage"
)

type Piece struct {
	storage.PieceImpl `json:"-"`

	Id   int   `json:"-"`
	Size int64 `json:"size"`

	Complete bool  `json:"complete"`
	Accessed int64 `json:"accessed"`

	mPiece *MemPiece  `json:"-"`
	dPiece *DiskPiece `json:"-"`

	cache *Cache `json:"-"`
}

func NewPiece(id int, cache *Cache) *Piece {
	p := &Piece{
		Id:    id,
		cache: cache,
	}

	if !cache.storage.cfg.UseDisk {
		p.mPiece = NewMemPiece(p)
	} else {
		p.dPiece = NewDiskPiece(p)
	}
	return p
}

func (p *Piece) WriteAt(b []byte, off int64) (n int, err error) {
	if !p.cache.storage.cfg.UseDisk {
		return p.mPiece.WriteAt(b, off)
	}
	return p.dPiece.WriteAt(b, off)
}

func (p *Piece) ReadAt(b []byte, off int64) (n int, err error) {
	if !p.cache.storage.cfg.UseDisk {
		return p.mPiece.ReadAt(b, off)
	}
	return p.dPiece.ReadAt(b, off)
}

func (p *Piece) MarkComplete() error {
	p.Complete = true
	return nil
}

func (p *Piece) MarkNotComplete() error {
	p.Complete = false
	return nil
}

func (p *Piece) Completion() storage.Completion {
	return storage.Completion{
		Complete: p.Complete,
		Ok:       true,
	}
}

// Release освобождает память куска. Приоритет куска не трогается: движок выводит его
// из зон ридеров, поэтому вытесненный кусок перестает запрашиваться сам.
func (p *Piece) Release() {
	if !p.cache.storage.cfg.UseDisk {
		p.mPiece.Release()
	} else {
		p.dPiece.Release()
	}

	// Сообщаем движку, что данных больше нет. Флаги куска уже сброшены, поэтому
	// движок снимет кусок с раздачи и не будет отдавать его пирам.
	if p.cache != nil && p.cache.torrent != nil {
		p.cache.torrent.Piece(p.Id).UpdateCompletion()
	}
}
