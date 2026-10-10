package torrstor

import (
	"sync"

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

	// Поля размера и доступа пишутся при записи чанка и читаются при снятии
	// состояния кэша, вытеснении и отдаче пирам, поэтому идут под мьютексом.
	mu sync.RWMutex
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
	p.SetComplete(true)
	return nil
}

func (p *Piece) MarkNotComplete() error {
	p.SetComplete(false)
	return nil
}

func (p *Piece) Completion() storage.Completion {
	return storage.Completion{
		Complete: p.IsComplete(),
		Ok:       true,
	}
}

// SizeOf возвращает число записанных байт куска.
func (p *Piece) SizeOf() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Size
}

// AccessedAt возвращает время последнего обращения к куску.
func (p *Piece) AccessedAt() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Accessed
}

// IsComplete сообщает, записан ли кусок целиком.
func (p *Piece) IsComplete() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Complete
}

// AddSize увеличивает размер куска, не выходя за его длину.
func (p *Piece) AddSize(n int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Size += n
	if p.Size > p.cache.pieceLength {
		p.Size = p.cache.pieceLength
	}
}

// SetAccessed отмечает время последнего обращения к куску.
func (p *Piece) SetAccessed(t int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Accessed = t
}

// SetComplete задает признак полного куска.
func (p *Piece) SetComplete(v bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Complete = v
}

// Reset обнуляет размер и признак полноты после освобождения данных.
func (p *Piece) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Size = 0
	p.Complete = false
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
	if p.cache != nil {
		if t := p.cache.Torrent(); t != nil {
			t.Piece(p.Id).UpdateCompletion()
		}
	}
}
