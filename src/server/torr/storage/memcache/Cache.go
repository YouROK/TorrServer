package memcache

import (
	"sync"

	"server/torr/storage/state"
	"server/utils"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"log"
	"server/torr/reader"
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
	muReader sync.Mutex
	isRemove bool

	pieces     map[int]*Piece
	bufferPull *BufferPool
	usedPieces map[int]struct{}

	readers map[*reader.Reader]struct{}
}

func NewCache(capacity int64, storage *Storage) *Cache {
	ret := &Cache{
		capacity:   capacity,
		filled:     0,
		pieces:     make(map[int]*Piece),
		readers:    make(map[*reader.Reader]struct{}),
		usedPieces: make(map[int]struct{}),
		s:          storage,
	}

	return ret
}

func (c *Cache) Init(info *metainfo.Info, hash metainfo.Hash) {
	log.Println("Create cache for:", info.Name)
	//Min capacity of 2 pieces length
	caps := info.PieceLength * 2
	if c.capacity < caps {
		c.capacity = caps
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
	defer func() {
		c.muPiece.Unlock()
		go utils.FreeOSMemGC(c.capacity)
	}()
	if val, ok := c.pieces[m.Index()]; ok {
		return val
	}
	return nil
}

func (c *Cache) Close() error {
	c.isRemove = false
	log.Println("Close cache for:", c.hash)
	if _, ok := c.s.caches[c.hash]; ok {
		delete(c.s.caches, c.hash)
	}
	c.pieces = nil
	c.bufferPull = nil
	c.readers = nil
	utils.FreeOSMemGC(0)
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

	bufPieces := c.getBufferedPieces()

	if len(bufPieces) > 0 && c.filled >= c.capacity {
		c.muReader.Lock()
		for reader := range c.readers {
			beg, end := c.getReaderPieces(reader)
			for id := range bufPieces {
				if id >= beg && id <= end {
					delete(bufPieces, id)
				}
			}
		}
		c.muReader.Unlock()
		if len(bufPieces) > 0 {
			for _, p := range bufPieces {
				p.Release()
			}
			bufPieces = nil
			go utils.FreeOSMemGC(c.capacity)
		}
	}
}

func (c *Cache) getBufferedPieces() map[int]*Piece {
	pieces := make(map[int]*Piece)
	fill := int64(0)
	used := c.usedPieces
	for u := range used {
		piece := c.pieces[u]
		if piece.Size > 0 {
			if piece.Id > 0 {
				pieces[piece.Id] = piece
			}
			fill += piece.Size
		}
	}
	c.filled = fill

	return pieces
}

func (c *Cache) removePiece(piece *Piece) {
	piece.Release()
	return
}

func (c *Cache) AddReader(r *reader.Reader) {
	c.muReader.Lock()
	defer c.muReader.Unlock()
	c.readers[r] = struct{}{}
}

func (c *Cache) RemReader(r *reader.Reader) {
	c.muReader.Lock()
	defer c.muReader.Unlock()
	delete(c.readers, r)
}

func (c *Cache) ReadersLen() int {
	if c == nil || c.readers == nil {
		return 0
	}
	return len(c.readers)
}

func (c *Cache) getReaderPieces(reader *reader.Reader) (begin, end int) {
	end = int((reader.Offset() + reader.Readahead()) / c.pieceLength)
	begin = int((reader.Offset() - c.capacity + reader.Readahead()) / c.pieceLength)
	return
}
