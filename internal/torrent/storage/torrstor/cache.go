package torrstor

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"silo/internal/log"
	"sort"
	"sync"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

// freeOSMemory возвращает свободную память операционной системе с троттлингом
func FreeOSMemGC() {
	runtime.GC()
	debug.FreeOSMemory()
}

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
	isClosed bool
	muRemove sync.Mutex
	torrent  *torrent.Torrent
}

func NewCache(capacity int64, s *Storage) *Cache {
	return &Cache{
		capacity: capacity,
		filled:   0,
		pieces:   make(map[int]*Piece),
		storage:  s,
		readers:  make(map[*Reader]struct{}),
	}
}

func (c *Cache) Init(info *metainfo.Info, hash metainfo.Hash) {
	log.Debugf("[TorrStor] Create cache for: %s (%s)", info.Name, hash.HexString())

	if c.capacity <= 0 {
		c.capacity = info.PieceLength * 4
	}

	c.pieceLength = info.PieceLength
	c.pieceCount = info.NumPieces()
	c.hash = hash

	cfg := c.storage.cfg
	if cfg.UseDisk {
		dir := filepath.Join(cfg.TorrentsSavePath, hash.HexString())
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Errorf("[TorrStor] Error creating cache directory: %v", err)
		}
	}

	for i := 0; i < c.pieceCount; i++ {
		c.pieces[i] = NewPiece(i, c)
	}
}

func (c *Cache) SetTorrent(torr *torrent.Torrent) {
	c.torrent = torr
}

func (c *Cache) Piece(m metainfo.Piece) storage.PieceImpl {
	if val, ok := c.pieces[m.Index()]; ok {
		return val
	}
	return &PieceFake{}
}

func (c *Cache) Close() error {
	if c.torrent != nil {
		log.Debugf("[TorrStor] Close cache for: %s (%s)", c.torrent.Name(), c.hash.HexString())
	} else {
		log.Debugf("[TorrStor] Close cache for: %s", c.hash.HexString())
	}
	c.isClosed = true

	delete(c.storage.caches, c.hash)

	cfg := c.storage.cfg
	if cfg.RemoveCacheOnDrop && cfg.UseDisk {
		name := filepath.Join(cfg.TorrentsSavePath, c.hash.HexString())
		if name != "" && name != "/" {
			for _, v := range c.pieces {
				if v.dPiece != nil {
					_ = os.Remove(v.dPiece.name)
				}
			}
			_ = os.Remove(name)
		}
	}

	c.muReaders.Lock()
	c.readers = nil
	c.pieces = nil
	c.muReaders.Unlock()

	FreeOSMemGC()
	return nil
}

func (c *Cache) removePiece(piece *Piece) {
	if !c.isClosed {
		piece.Release()
	}
}

// applyZones распределяет зону скачивания торрента между активными ридерами.
// Зона одного ридера это Capacity за вычетом запаса, поделенная на число ридеров.
func (c *Cache) applyZones() {
	if c == nil || c.torrent == nil {
		return
	}

	c.muReaders.Lock()
	readers := make([]*Reader, 0, len(c.readers))
	for r := range c.readers {
		if r.isUse {
			readers = append(readers, r)
		}
	}
	c.muReaders.Unlock()

	cfg := c.storage.cfg
	capacity := c.capacity
	if capacity <= 0 {
		capacity = cfg.Capacity
	}

	zone := cfg.ZoneBytes(capacity)
	// Мелкий кэш не делится на зону и запас: ридеру отдается весь буфер.
	if zone < 2*c.pieceLength {
		zone = capacity
	}
	next := cfg.NextBytes()
	readahead := cfg.ReadaheadBytes()

	if n := int64(len(readers)); n > 1 {
		zone /= n
		next /= n
		readahead /= n
	}

	for _, r := range readers {
		r.SetZones(next, readahead, zone)
	}
}

func (c *Cache) GetState() *CacheState {
	cState := new(CacheState)

	piecesState := make(map[int]ItemState)
	var fill int64 = 0

	if len(c.pieces) > 0 {
		for _, p := range c.pieces {
			if p.Size > 0 {
				fill += p.Size
				priority := 0
				if c.torrent != nil {
					priority = int(c.torrent.PieceState(p.Id).Priority)
				}
				piecesState[p.Id] = ItemState{
					Id:        p.Id,
					Size:      p.Size,
					Length:    c.pieceLength,
					Completed: p.Complete,
					Priority:  priority,
				}
			}
		}
	}

	readersState := make([]*ReaderState, 0)

	if c.Readers() > 0 {
		c.muReaders.Lock()
		for r := range c.readers {
			rng := r.getPiecesRange()
			pc := r.getReaderPiece()
			readersState = append(readersState, &ReaderState{
				Start:  rng.Start,
				End:    rng.End,
				Reader: pc,
			})
		}
		c.muReaders.Unlock()
	}

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
	if c.isRemove || c.isClosed {
		return
	}

	if !c.muRemove.TryLock() {
		return
	}
	defer c.muRemove.Unlock()

	c.isRemove = true
	defer func() { c.isRemove = false }()

	remPieces := c.getRemPieces()
	if c.filled > c.capacity {
		rems := (c.filled-c.capacity)/c.pieceLength + 1
		for _, p := range remPieces {
			c.removePiece(p)
			rems--
			if rems <= 0 {
				FreeOSMemGC()
				return
			}
		}
	}
}

func (c *Cache) getRemPieces() []*Piece {
	c.muReaders.Lock()
	readers := make([]*Reader, 0, len(c.readers))
	for r := range c.readers {
		readers = append(readers, r)
	}
	// Снимок карты кусков берется под тем же мьютексом, что и обнуление в Close:
	// иначе вытеснение читает карту, пока закрытие кэша ее очищает.
	pieces := make([]*Piece, 0, len(c.pieces))
	for _, p := range c.pieces {
		pieces = append(pieces, p)
	}
	c.muReaders.Unlock()

	ranges := make([]Range, 0)
	for _, r := range readers {
		r.checkReader()
		if r.isUse {
			ranges = append(ranges, r.getPiecesRange())
		}
	}
	ranges = mergeRange(ranges)

	piecesRemove := make([]*Piece, 0)
	fill := int64(0)

	for _, p := range pieces {
		id := p.Id
		if p.Size > 0 {
			fill += p.Size
		}
		if len(ranges) > 0 {
			if !inRanges(ranges, id) {
				if p.Size > 0 && !c.isIdInFileBE(ranges, id) {
					piecesRemove = append(piecesRemove, p)
				}
			}
		} else {
			if p.Size > 0 && !c.isIdInFileBE(ranges, id) {
				piecesRemove = append(piecesRemove, p)
			}
		}
	}

	sort.Slice(piecesRemove, func(i, j int) bool {
		if piecesRemove[i].Complete != piecesRemove[j].Complete {
			return piecesRemove[i].Complete
		}
		return piecesRemove[i].Accessed < piecesRemove[j].Accessed
	})

	c.filled = fill
	return piecesRemove
}

func (c *Cache) isIdInFileBE(ranges []Range, id int) bool {
	fileRangeNotDelete := int64(c.pieceLength)
	if fileRangeNotDelete < 8<<20 {
		fileRangeNotDelete = 8 << 20
	}

	for _, rng := range ranges {
		if rng.File == nil {
			continue
		}
		ss := int(rng.File.Offset() / c.pieceLength)
		se := int((rng.File.Offset() + fileRangeNotDelete) / c.pieceLength)

		es := int((rng.File.Offset() + rng.File.Length() - fileRangeNotDelete) / c.pieceLength)
		ee := int((rng.File.Offset() + rng.File.Length()) / c.pieceLength)

		if id >= ss && id < se || id > es && id <= ee {
			return true
		}
	}
	return false
}

func (c *Cache) NewReader(file *torrent.File) *Reader {
	r := newReader(file, c)
	c.applyZones()
	return r
}

func (c *Cache) GetUseReaders() int {
	if c == nil {
		return 0
	}
	c.muReaders.Lock()
	defer c.muReaders.Unlock()
	readers := 0
	for reader := range c.readers {
		if reader.isUse {
			readers++
		}
	}
	return readers
}

func (c *Cache) Readers() int {
	if c == nil {
		return 0
	}
	c.muReaders.Lock()
	defer c.muReaders.Unlock()
	if c.readers == nil {
		return 0
	}
	return len(c.readers)
}

func (c *Cache) CloseReader(r *Reader) {
	r.cache.muReaders.Lock()
	delete(r.cache.readers, r)
	r.cache.muReaders.Unlock()
	r.Close()
	go c.applyZones()
}

func (c *Cache) GetCapacity() int64 {
	if c == nil {
		return 0
	}
	return c.capacity
}
