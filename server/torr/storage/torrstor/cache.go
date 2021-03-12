package torrstor

import (
	"sort"
	"sync"
	"time"

	"github.com/anacrolix/torrent"

	"server/log"
	"server/settings"
	"server/torr/storage/state"
	"server/torr/utils"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

type Cache struct {
	storage.TorrentImpl
	storage *Storage

	capacity int64
	filled   int64
	hash     metainfo.Hash

	pieceLength int64
	pieceCount  int

	pieces map[int]*Piece

	readers   map[*Reader]struct{}
	muReaders sync.Mutex

	isRemove bool
	muRemove sync.Mutex
	torrent  *torrent.Torrent
}

func NewCache(capacity int64, storage *Storage) *Cache {
	ret := &Cache{
		capacity: capacity,
		filled:   0,
		pieces:   make(map[int]*Piece),
		storage:  storage,
		readers:  make(map[*Reader]struct{}),
	}

	return ret
}

func (c *Cache) Init(info *metainfo.Info, hash metainfo.Hash) {
	log.TLogln("Create cache for:", info.Name, hash.HexString())
	if c.capacity == 0 {
		c.capacity = info.PieceLength * 4
	}

	c.pieceLength = info.PieceLength
	c.pieceCount = info.NumPieces()
	c.hash = hash

	for i := 0; i < c.pieceCount; i++ {
		c.pieces[i] = &Piece{
			Id:    i,
			Hash:  info.Piece(i).Hash().HexString(),
			cache: c,
		}
	}
}

func (c *Cache) SetTorrent(torr *torrent.Torrent) {
	c.torrent = torr
}

func (c *Cache) Piece(m metainfo.Piece) storage.PieceImpl {
	if val, ok := c.pieces[m.Index()]; ok {
		return val
	}
	return nil
}

func (c *Cache) Close() error {
	log.TLogln("Close cache for:", c.hash)
	if _, ok := c.storage.caches[c.hash]; ok {
		delete(c.storage.caches, c.hash)
	}
	c.pieces = nil

	c.muReaders.Lock()
	c.readers = nil
	c.muReaders.Unlock()

	utils.FreeOSMemGC()
	return nil
}

func (c *Cache) removePiece(piece *Piece) {
	piece.Release()
}

func (c *Cache) AdjustRA(readahead int64) {
	if settings.BTsets.CacheSize == 0 {
		c.capacity = readahead * 3
	}
	c.muReaders.Lock()
	for r, _ := range c.readers {
		r.SetReadahead(readahead)
	}
	c.muReaders.Unlock()
}

func (c *Cache) GetState() *state.CacheState {
	cState := new(state.CacheState)

	piecesState := make(map[int]state.ItemState, 0)
	var fill int64 = 0
	for _, p := range c.pieces {
		if p.Size > 0 {
			fill += p.Size
			piecesState[p.Id] = state.ItemState{
				Id:        p.Id,
				Size:      p.Size,
				Length:    c.pieceLength,
				Completed: p.complete,
			}
		}
	}

	readersState := make([]*state.ReaderState, 0)
	c.muReaders.Lock()
	for r, _ := range c.readers {
		rng := r.getPiecesRange()
		pc := r.getReaderPiece()
		readersState = append(readersState, &state.ReaderState{
			Start:  rng.Start,
			End:    rng.End,
			Reader: pc,
		})
	}
	c.muReaders.Unlock()

	c.filled = fill
	cState.Capacity = c.capacity
	cState.PiecesLength = c.pieceLength
	cState.PiecesCount = c.pieceCount
	cState.Hash = c.hash.HexString()
	cState.Filled = fill
	cState.Pieces = piecesState
	cState.Readers = readersState
	return cState
}

func (c *Cache) cleanPieces() {
	if c.isRemove {
		return
	}
	c.muRemove.Lock()
	if c.isRemove {
		c.muRemove.Unlock()
		return
	}
	c.isRemove = true
	defer func() { c.isRemove = false }()
	c.muRemove.Unlock()

	remPieces := c.getRemPieces()
	if c.filled > c.capacity {
		rems := (c.filled-c.capacity)/c.pieceLength + 1
		for _, p := range remPieces {
			c.removePiece(p)
			rems--
			if rems <= 0 {
				break
			}
		}
		if rems <= 0 {
			utils.FreeOSMemGC()
		}
	}
}

func (c *Cache) getRemPieces() []*Piece {
	piecesRemove := make([]*Piece, 0)
	fill := int64(0)

	ranges := make([]Range, 0)
	c.muReaders.Lock()
	for r, _ := range c.readers {
		ranges = append(ranges, r.getPiecesRange())
	}
	c.muReaders.Unlock()
	ranges = mergeRange(ranges)

	for id, p := range c.pieces {
		if p.Size > 0 {
			fill += p.Size
		}
		if len(ranges) > 0 {
			if !inRanges(ranges, id) {
				if p.Size > 0 && !c.isIdInFileBE(ranges, id) {
					piecesRemove = append(piecesRemove, p)
				}
			}
		}
	}

	for r, _ := range c.readers {
		pc := r.getReaderPiece()
		end := r.getPiecesRange().End
		limit := 5

		for pc <= end && limit > 0 {
			if !c.pieces[pc].complete {
				if c.torrent.PieceState(pc).Priority == torrent.PiecePriorityNone {
					c.torrent.Piece(pc).SetPriority(torrent.PiecePriorityNormal)
				}
				limit--
			}
			pc++
		}
	}

	sort.Slice(piecesRemove, func(i, j int) bool {
		return piecesRemove[i].accessed < piecesRemove[j].accessed
	})

	c.filled = fill
	return piecesRemove
}

func (c *Cache) isIdInFileBE(ranges []Range, id int) bool {
	for _, rng := range ranges {
		ss := int(rng.File.Offset() / c.pieceLength)
		se := int((5*1024*1024 + rng.File.Offset()) / c.pieceLength)

		es := int((rng.File.Offset() + rng.File.Length() - 5*1024*1024) / c.pieceLength)
		ee := int((rng.File.Offset() + rng.File.Length()) / c.pieceLength)

		if id >= ss && id <= se || id >= es && id <= ee {
			return true
		}
	}
	return false
}

//////////////////
// Reader section
////////

func (c *Cache) NewReader(file *torrent.File) *Reader {
	return newReader(file, c)
}

func (c *Cache) Readers() int {
	if c == nil {
		return 0
	}
	c.muReaders.Lock()
	defer c.muReaders.Unlock()
	if c == nil || c.readers == nil {
		return 0
	}
	return len(c.readers)
}

func (c *Cache) CloseReader(r *Reader) {
	r.cache.muReaders.Lock()
	r.Close()
	delete(r.cache.readers, r)
	r.cache.muReaders.Unlock()
	go c.updatePriority()
}

func (c *Cache) updatePriority() {
	time.Sleep(time.Second)
	ranges := make([]Range, 0)
	c.muReaders.Lock()
	for r, _ := range c.readers {
		ranges = append(ranges, r.getPiecesRange())
	}
	c.muReaders.Unlock()
	ranges = mergeRange(ranges)

	for id, _ := range c.pieces {
		if len(ranges) > 0 {
			if !inRanges(ranges, id) {
				if c.torrent.PieceState(id).Priority != torrent.PiecePriorityNone {
					c.torrent.Piece(id).SetPriority(torrent.PiecePriorityNone)
				}
			}
		} else {
			if c.torrent.PieceState(id).Priority != torrent.PiecePriorityNone {
				c.torrent.Piece(id).SetPriority(torrent.PiecePriorityNone)
			}
		}
	}
}
