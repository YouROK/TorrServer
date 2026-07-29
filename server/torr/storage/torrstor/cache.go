package torrstor

import (
	"os"
	"path/filepath"
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
	/** Total torrent length from metainfo — used for last-piece Length before SetTorrent. */
	totalLength int64

	pieces map[int]*Piece

	// activePieces tracks ids with Size > 0 for O(active) GetState (not O(pieceCount)).
	activePieces map[int]struct{}
	muActive     sync.Mutex

	readers   map[*Reader]struct{}
	muReaders sync.Mutex

	isRemove bool
	isClosed bool
	muRemove sync.Mutex
	// muPrio serializes clearPriority and setLoadPriority so that the priority
	// reset of a reader that has just closed cannot wipe the priorities a
	// freshly created reader has already set.
	muPrio  sync.Mutex
	torrent *torrent.Torrent
}

func NewCache(capacity int64, storage *Storage) *Cache {
	ret := &Cache{
		capacity:     capacity,
		filled:       0,
		pieces:       make(map[int]*Piece),
		activePieces: make(map[int]struct{}),
		storage:      storage,
		readers:      make(map[*Reader]struct{}),
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
	c.totalLength = info.TotalLength()
	c.hash = hash

	if settings.BTsets.UseDisk {
		name := filepath.Join(settings.BTsets.TorrentsSavePath, hash.HexString())
		err := os.MkdirAll(name, 0o777)
		if err != nil {
			log.TLogln("Error create dir:", err)
		}
	}

	for i := 0; i < c.pieceCount; i++ {
		c.pieces[i] = NewPiece(i, c)
	}
}

func (c *Cache) SetTorrent(torr *torrent.Torrent) {
	c.torrent = torr
}

/** Actual byte length of piece id (last piece may be shorter than PiecesLength). */
func (c *Cache) pieceByteLength(id int) int64 {
	if id < 0 || id >= c.pieceCount || c.pieceLength <= 0 {
		return c.pieceLength
	}
	total := c.totalLength
	if c.torrent != nil {
		if tl := c.torrent.Length(); tl > 0 {
			total = tl
		}
	}
	if total <= 0 {
		return c.pieceLength
	}
	start := int64(id) * c.pieceLength
	remain := total - start
	if remain <= 0 {
		return c.pieceLength
	}
	if remain < c.pieceLength {
		return remain
	}
	return c.pieceLength
}

func (c *Cache) Piece(m metainfo.Piece) storage.PieceImpl {
	if val, ok := c.pieces[m.Index()]; ok {
		return val
	}
	return &PieceFake{}
}

func (c *Cache) Close() error {
	if c.torrent != nil {
		log.TLogln("Close cache for:", c.torrent.Name(), c.hash)
	} else {
		log.TLogln("Close cache for:", c.hash)
	}
	c.isClosed = true

	c.storage.removeCache(c.hash)

	if settings.BTsets.RemoveCacheOnDrop {
		name := filepath.Join(settings.BTsets.TorrentsSavePath, c.hash.HexString())
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

	c.muActive.Lock()
	c.activePieces = nil
	c.muActive.Unlock()

	utils.FreeOSMemGC()
	return nil
}

func (c *Cache) removePiece(piece *Piece) {
	if !c.isClosed {
		piece.Release()
	}
}

func (c *Cache) AdjustRA(readahead int64) {
	if c == nil {
		return
	}
	if settings.BTsets.CacheSize == 0 {
		c.capacity = readahead * 3
	}
	if c.Readers() > 0 {
		c.muReaders.Lock()
		for r := range c.readers {
			r.SetReadahead(readahead)
		}
		c.muReaders.Unlock()
	}
}

func (c *Cache) notePieceFilled(id int) {
	c.muActive.Lock()
	if c.activePieces == nil {
		c.activePieces = make(map[int]struct{})
	}
	c.activePieces[id] = struct{}{}
	c.muActive.Unlock()
}

func (c *Cache) notePieceEmpty(id int) {
	c.muActive.Lock()
	delete(c.activePieces, id)
	c.muActive.Unlock()
}

func (c *Cache) GetState() *state.CacheState {
	cState := new(state.CacheState)

	piecesState := make(map[int]state.ItemState)
	var fill int64

	readersState := make([]*state.ReaderState, 0)
	priorityWindow := make(map[int]struct{})

	if c.Readers() > 0 {
		c.muReaders.Lock()
		for r := range c.readers {
			// Closed readers are gone; a nil file cannot yield a piece range.
			if r.isClosed || r.file == nil {
				continue
			}
			r.checkReader()
			rng := r.getPiecesRange()
			pc := r.getReaderPiece()
			// Idle readers are still reported so the UI can show a frozen playhead
			// instead of silently dropping the square.
			readersState = append(readersState, &state.ReaderState{
				Start:  rng.Start,
				End:    rng.End,
				Reader: pc,
				Active: r.isUse,
			})
			if !r.isUse {
				continue
			}
			// Inclusive End (matches inRanges). Small margin so debug labels appear
			// on neighbouring queued pieces just outside the strict reader window.
			const margin = 5
			start := rng.Start - margin
			if start < 0 {
				start = 0
			}
			end := rng.End + margin
			if end >= c.pieceCount {
				end = c.pieceCount - 1
			}
			for id := start; id <= end; id++ {
				priorityWindow[id] = struct{}{}
			}
		}
		c.muReaders.Unlock()
	}

	c.muActive.Lock()
	activeIDs := make([]int, 0, len(c.activePieces))
	for id := range c.activePieces {
		activeIDs = append(activeIDs, id)
	}
	c.muActive.Unlock()

	stale := make([]int, 0)
	for _, id := range activeIDs {
		p, ok := c.pieces[id]
		if !ok || p == nil || p.Size <= 0 {
			stale = append(stale, id)
			continue
		}
		fill += p.Size
		priority := 0
		if c.torrent != nil {
			priority = int(c.torrent.PieceState(p.Id).Priority)
		}
		plen := c.pieceByteLength(p.Id)
		piecesState[p.Id] = state.ItemState{
			Id:     p.Id,
			Size:   p.Size,
			Length: plen,
			// Completed follows bytes present — not MarkComplete alone (avoids green empty cells).
			Completed: plen > 0 && p.Size >= plen,
			Priority:  priority,
		}
	}
	for _, id := range stale {
		c.notePieceEmpty(id)
	}

	// Emit Size==0 pieces inside the reader priority window so the web snake
	// can show H/R/N/A labels before bytes arrive.
	if c.torrent != nil {
		for id := range priorityWindow {
			if _, exists := piecesState[id]; exists {
				continue
			}
			p, ok := c.pieces[id]
			if !ok || p == nil {
				continue
			}
			plen := c.pieceByteLength(id)
			piecesState[id] = state.ItemState{
				Id:        id,
				Size:      p.Size,
				Length:    plen,
				Completed: plen > 0 && p.Size >= plen,
				Priority:  int(c.torrent.PieceState(id).Priority),
			}
		}
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

	// Protection against concurrent deletion
	if !c.muRemove.TryLock() {
		return // Cleanup is already in progress in another goroutine
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
				utils.FreeOSMemGC()
				return
			}
		}
	}
}

func (c *Cache) getRemPieces() []*Piece {
	// Copy readers without a long lock
	c.muReaders.Lock()
	readers := make([]*Reader, 0, len(c.readers))
	for r := range c.readers {
		readers = append(readers, r)
	}
	c.muReaders.Unlock()

	// Collect read ranges from active readers
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

	// Determine which chunks can be deleted
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
		} else {
			// When preloading, clear everything except the beginning and end of the file
			if p.Size > 0 && !c.isIdInFileBE(ranges, id) {
				piecesRemove = append(piecesRemove, p)
			}
		}
	}

	c.clearPriority()
	c.setLoadPriority(ranges)

	// Sort by last access time (oldest first)
	sort.Slice(piecesRemove, func(i, j int) bool {
		return piecesRemove[i].Accessed < piecesRemove[j].Accessed
	})

	c.filled = fill
	return piecesRemove
}

func (c *Cache) setLoadPriority(ranges []Range) {
	if c.torrent == nil || c.pieces == nil {
		return
	}
	c.muPrio.Lock()
	defer c.muPrio.Unlock()
	c.muReaders.Lock()
	for r := range c.readers {
		if !r.isUse {
			continue
		}
		if c.isIdInFileBE(ranges, r.getReaderPiece()) {
			continue
		}
		readerPos := r.getReaderPiece()
		readerRAHPos := r.getReaderRAHPiece()
		end := r.getPiecesRange().End
		count := settings.BTsets.ConnectionsLimit / len(c.readers) // max concurrent loading blocks
		limit := 0
		for i := readerPos; i < end && limit < count; i++ {
			if !c.pieces[i].Complete {
				if i == readerPos {
					c.torrent.Piece(i).SetPriority(torrent.PiecePriorityNow)
				} else if i == readerPos+1 {
					c.torrent.Piece(i).SetPriority(torrent.PiecePriorityNext)
				} else if i > readerPos && i <= readerRAHPos {
					c.torrent.Piece(i).SetPriority(torrent.PiecePriorityReadahead)
				} else if i > readerRAHPos && i <= readerRAHPos+5 && c.torrent.PieceState(i).Priority != torrent.PiecePriorityHigh {
					c.torrent.Piece(i).SetPriority(torrent.PiecePriorityHigh)
				} else if i > readerRAHPos+5 && c.torrent.PieceState(i).Priority != torrent.PiecePriorityNormal {
					c.torrent.Piece(i).SetPriority(torrent.PiecePriorityNormal)
				}
				limit++
			}
		}
	}
	c.muReaders.Unlock()
}

func (c *Cache) isIdInFileBE(ranges []Range, id int) bool {
	// keep 8/16 MB
	FileRangeNotDelete := int64(c.pieceLength)
	if FileRangeNotDelete < 8<<20 {
		FileRangeNotDelete = 8 << 20
	}

	for _, rng := range ranges {
		ss := int(rng.File.Offset() / c.pieceLength)
		se := int((rng.File.Offset() + FileRangeNotDelete) / c.pieceLength)

		es := int((rng.File.Offset() + rng.File.Length() - FileRangeNotDelete) / c.pieceLength)
		ee := int((rng.File.Offset() + rng.File.Length()) / c.pieceLength)

		if id >= ss && id < se || id > es && id <= ee {
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
	r.Close()
	delete(r.cache.readers, r)
	r.cache.muReaders.Unlock()
	go c.clearPriority()
}

func (c *Cache) clearPriority() {
	if c.torrent == nil {
		return
	}
	// This used to sleep for a second before clearing priorities. A reader
	// created during that window could have its PiecePriorityNow/Next/Readahead
	// reset to None right after setLoadPriority had assigned them, starving the
	// player. A mutex provides the same ordering without the race window.
	c.muPrio.Lock()
	defer c.muPrio.Unlock()
	ranges := make([]Range, 0)
	c.muReaders.Lock()
	for r := range c.readers {
		r.checkReader()
		if r.isUse {
			ranges = append(ranges, r.getPiecesRange())
		}
	}
	c.muReaders.Unlock()
	ranges = mergeRange(ranges)

	for id := range c.pieces {
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

func (c *Cache) GetCapacity() int64 {
	if c == nil {
		return 0
	}
	return c.capacity
}
