package memcache

import (
	"fmt"
	"sort"
	"sync"

	"server/torr/storage/state"
	"server/utils"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

type Cache struct {
	storage.TorrentImpl

	s *Storage

	capacity int64
	filled   int64
	hash     metainfo.Hash

	pieceLength int64
	pieceCount  int
	piecesBuff  int

	muPiece  sync.Mutex
	muRemove sync.Mutex
	isRemove bool

	pieces     map[int]*Piece
	bufferPull *BufferPool

	prcLoaded int
}

func NewCache(capacity int64, storage *Storage) *Cache {
	ret := &Cache{
		capacity: capacity,
		filled:   0,
		pieces:   make(map[int]*Piece),
		s:        storage,
	}

	return ret
}

func (c *Cache) Init(info *metainfo.Info, hash metainfo.Hash) {
	fmt.Println("Create cache for:", info.Name)
	//Min capacity of 2 pieces length
	cap := info.PieceLength * 2
	if c.capacity < cap {
		c.capacity = cap
	}
	c.pieceLength = info.PieceLength
	c.pieceCount = info.NumPieces()
	c.piecesBuff = int(c.capacity / c.pieceLength)
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

func (c *Cache) Piece(m metainfo.Piece) storage.PieceImpl {
	c.muPiece.Lock()
	defer c.muPiece.Unlock()
	if val, ok := c.pieces[m.Index()]; ok {
		return val
	}
	return nil
}

func (c *Cache) Close() error {
	c.isRemove = false
	fmt.Println("Close cache for:", c.hash)
	if _, ok := c.s.caches[c.hash]; ok {
		delete(c.s.caches, c.hash)
	}
	c.pieces = nil
	c.bufferPull = nil
	utils.FreeOSMemGC()
	return nil
}

func (c *Cache) GetState() state.CacheState {
	cState := state.CacheState{}
	cState.Capacity = c.capacity
	cState.PiecesLength = c.pieceLength
	cState.PiecesCount = c.pieceCount
	cState.Hash = c.hash.HexString()

	stats := make(map[int]state.ItemState, 0)
	c.muPiece.Lock()
	var fill int64 = 0
	for _, value := range c.pieces {
		stat := value.Stat()
		if stat.BufferSize > 0 {
			fill += stat.BufferSize
			stats[stat.Id] = stat
		}
	}
	c.filled = fill
	c.muPiece.Unlock()
	cState.Filled = c.filled
	cState.Pieces = stats
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
	if len(remPieces) > 0 && (c.capacity < c.filled || c.bufferPull.Len() <= 1) {
		remCount := int((c.filled - c.capacity) / c.pieceLength)
		if remCount < 1 {
			remCount = 1
		}
		if remCount > len(remPieces) {
			remCount = len(remPieces)
		}

		remPieces = remPieces[:remCount]

		for _, p := range remPieces {
			c.removePiece(p)
		}
	}
}

func (c *Cache) getRemPieces() []*Piece {
	pieces := make([]*Piece, 0)
	fill := int64(0)
	loading := 0
	used := make(map[int]struct{})
	for _, b := range c.bufferPull.buffs {
		if b.used {
			used[b.pieceId] = struct{}{}
		}
	}
	for u := range used {
		v := c.pieces[u]
		if v.Size > 0 {
			if v.Id > 0 {
				pieces = append(pieces, v)
			}
			fill += v.Size
			if !v.complete {
				loading++
			}
		}
	}
	c.filled = fill
	sort.Slice(pieces, func(i, j int) bool {
		return pieces[i].accessed.Before(pieces[j].accessed)
	})

	c.prcLoaded = prc(c.piecesBuff-loading, c.piecesBuff)
	return pieces
}

func (c *Cache) removePiece(piece *Piece) {
	c.muPiece.Lock()
	defer c.muPiece.Unlock()
	piece.Release()

	st := fmt.Sprintf("%v%% %v\t%s\t%s", c.prcLoaded, piece.Id, piece.accessed.Format("15:04:05.000"), piece.Hash)
	if c.prcLoaded >= 95 {
		fmt.Println("Clean memory GC:", st)
		utils.FreeOSMemGC()
	} else {
		fmt.Println("Clean memory:", st)
		utils.FreeOSMem()
	}
}

func prc(val, of int) int {
	return int(float64(val) * 100.0 / float64(of))
}
