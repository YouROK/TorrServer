package torrstor

import (
	"sort"
	"sync"

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

	pieces     map[int]*Piece
	bufferPull *BufferPool

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
	c.bufferPull = NewBufferPool(c.pieceLength, c.capacity)

	for i := 0; i < c.pieceCount; i++ {
		c.pieces[i] = &Piece{
			Id:     i,
			Length: info.Piece(i).Length(),
			Hash:   info.Piece(i).Hash().HexString(),
			cache:  c,
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
	c.bufferPull = nil

	c.muReaders.Lock()
	c.readers = nil
	c.muReaders.Unlock()

	utils.FreeOSMemGC()
	return nil
}

func (c *Cache) removePiece(piece *Piece) {
	piece.Release()
	utils.FreeOSMemGC()
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
			fill += p.Length
			piecesState[p.Id] = state.ItemState{
				Id:        p.Id,
				Size:      p.Size,
				Length:    p.Length,
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
		rems := (c.filled - c.capacity) / c.pieceLength
		for _, p := range remPieces {
			c.removePiece(p)
			rems--
			if rems <= 0 {
				break
			}
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
				piece := c.torrent.Piece(id)
				if piece.State().Priority != torrent.PiecePriorityNone {
					piece.SetPriority(torrent.PiecePriorityNone)
				}
				if p.Size > 0 {
					piecesRemove = append(piecesRemove, p)
				}
			}
		} else {
			piece := c.torrent.Piece(id)
			if piece.State().Priority != torrent.PiecePriorityNone {
				piece.SetPriority(torrent.PiecePriorityNone)
			}
		}
	}

	sort.Slice(piecesRemove, func(i, j int) bool {
		return piecesRemove[i].accessed < piecesRemove[j].accessed
	})

	c.filled = fill
	return piecesRemove
}
