package torrstor

import (
	"os"

	"server/settings"

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

// snapshotStats reads Size/Complete/Accessed under the piece I/O mutex so
// GetState/getRemPieces do not race Release/WriteAt.
func (p *Piece) snapshotStats() (size int64, complete bool, accessed int64) {
	if p == nil {
		return 0, false, 0
	}
	switch {
	case p.mPiece != nil:
		p.mPiece.mu.RLock()
		defer p.mPiece.mu.RUnlock()
	case p.dPiece != nil:
		p.dPiece.mu.RLock()
		defer p.dPiece.mu.RUnlock()
	}
	return p.Size, p.Complete, p.Accessed
}

func (p *Piece) MarkComplete() error {
	// Same mutex as snapshotStats/Release so GetState does not race Complete.
	switch {
	case p.mPiece != nil:
		p.mPiece.mu.Lock()
		p.Complete = true
		p.mPiece.mu.Unlock()
	case p.dPiece != nil:
		p.dPiece.mu.Lock()
		p.Complete = true
		p.dPiece.mu.Unlock()
	default:
		p.Complete = true
	}
	return nil
}

func (p *Piece) MarkNotComplete() error {
	// Hash fail: drop buffer so eviction does not orphan Size<=0 RAM/disk.
	// Release clears Complete/Size under the piece mutex.
	if p.mPiece != nil {
		p.mPiece.Release()
	} else if p.dPiece != nil {
		p.dPiece.Release()
	} else {
		p.Complete = false
		p.Size = 0
	}
	return nil
}

func (p *Piece) Completion() storage.Completion {
	_, complete, _ := p.snapshotStats()
	return storage.Completion{
		Complete: complete,
		Ok:       true,
	}
}

func (p *Piece) Release() {
	if path := p.releaseBuffers(); path != "" {
		_ = os.Remove(path)
	}
	updatePieceCompletion(p)
}

// releaseBuffers clears piece bytes under the piece mutex.
// For disk pieces it returns the file path so the caller can os.Remove after
// dropping muRemove (Close/cleanPieces must not delete under that lock).
func (p *Piece) releaseBuffers() (diskPath string) {
	if p == nil {
		return ""
	}
	if !settings.BTsets.UseDisk {
		if p.mPiece != nil {
			p.mPiece.Release()
		}
		return ""
	}
	if p.dPiece != nil {
		return p.dPiece.releaseState()
	}
	return ""
}

// updatePieceCompletion tells the torrent client the piece bytes are gone.
// Forget, not UpdateCompletion: an evicted partial piece keeps its received
// chunks marked in the client, which then finishes it over zeros, fails the
// hash and drops a good peer.
// Must not run while holding Cache.muRemove: Client.Close holds the client lock
// and then Cache.Close waits on muRemove (see cleanPieces).
func updatePieceCompletion(p *Piece) {
	if p == nil || p.cache == nil || p.cache.isClosed.Load() {
		return
	}
	if pieceCompletionUpdater != nil {
		pieceCompletionUpdater(p)
		return
	}
	// Eviction runs before SetTorrent and after the client is gone. The pointer
	// is guarded by muTor; a bare field read raced Close and panicked on nil.
	if tor := p.cache.torrentPtr(); tor != nil {
		tor.Piece(p.Id).Forget()
	}
}

// pieceCompletionUpdater overrides the client Forget call in tests.
var pieceCompletionUpdater func(*Piece)
