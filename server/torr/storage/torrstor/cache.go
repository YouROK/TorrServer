package torrstor

import (
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
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
	filled   atomic.Int64
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

	isClosed atomic.Bool
	// stopWatch wakes priorityWatchdog. A 5s sleep survived Close and kept one
	// goroutine per dropped torrent.
	stopWatch chan struct{}
	muRemove  sync.Mutex
	torrent   *torrent.Torrent
	muTor     sync.Mutex
}

func NewCache(capacity int64, storage *Storage) *Cache {
	ret := &Cache{
		capacity:     capacity,
		pieces:       make(map[int]*Piece),
		activePieces: make(map[int]struct{}),
		storage:      storage,
		readers:      make(map[*Reader]struct{}),
		stopWatch:    make(chan struct{}),
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

	if c.stopWatch == nil {
		c.stopWatch = make(chan struct{})
	}
	if !c.isClosed.Load() {
		go c.priorityWatchdog()
	}
}

// priorityWatchdog parks idle readers while a torrent stays open.
// Piece reads also reach checkReader through getRemPieces, but a paused
// player does not read, so without this sweep the reader would keep its
// download zone. Piece priorities themselves come from the torrent client's
// reader tiers, not from this sweep.
func (c *Cache) priorityWatchdog() {
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()
	for {
		if c.isClosed.Load() {
			return
		}
		select {
		case <-c.stopWatch:
			return
		case <-tick.C:
		}
		if c.isClosed.Load() || c.torrentPtr() == nil {
			continue
		}
		if c.GetUseReaders() > 0 {
			c.getRemPieces()
		}
	}
}

func (c *Cache) stopWatchdog() {
	if c.stopWatch == nil {
		return
	}
	select {
	case <-c.stopWatch:
	default:
		close(c.stopWatch)
	}
}

func (c *Cache) SetTorrent(torr *torrent.Torrent) {
	if c == nil {
		return
	}
	c.muTor.Lock()
	c.torrent = torr
	c.muTor.Unlock()
}

func (c *Cache) torrentPtr() *torrent.Torrent {
	if c == nil {
		return nil
	}
	c.muTor.Lock()
	defer c.muTor.Unlock()
	return c.torrent
}

/** Actual byte length of piece id (last piece may be shorter than PiecesLength). */
func (c *Cache) pieceByteLength(id int) int64 {
	if id < 0 || id >= c.pieceCount || c.pieceLength <= 0 {
		return c.pieceLength
	}
	total := c.totalLength
	if tor := c.torrentPtr(); tor != nil {
		if tl := tor.Length(); tl > 0 {
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
	if c == nil {
		return nil
	}
	if !c.isClosed.CompareAndSwap(false, true) {
		return nil
	}
	c.stopWatchdog()
	if tor := c.torrentPtr(); tor != nil {
		log.TLogln("Close cache for:", tor.Name(), c.hash)
	} else {
		log.TLogln("Close cache for:", c.hash)
	}

	c.storage.removeCache(c.hash)

	if settings.BTsets.RemoveCacheOnDrop {
		name := filepath.Join(settings.BTsets.TorrentsSavePath, c.hash.HexString())
		if name != "" && name != "/" {
			for _, v := range c.pieces {
				if v != nil && v.dPiece != nil {
					_ = os.Remove(v.dPiece.name)
				}
			}
			_ = os.Remove(name)
		}
	}

	// Drop piece buffers before FreeOSMem so rem/ffp churn can return RAM.
	// MemPiece/DiskPiece.Release only — Piece.Release also touches the torrent client.
	// Hold muRemove so cleanPieces cannot interleave removePiece with this release.
	// Keep c.pieces (do not nil): getRemPieces/GetState may still read the field;
	// buffers are already nil and isClosed blocks new allocs.
	c.muRemove.Lock()
	for _, p := range c.pieces {
		if p == nil {
			continue
		}
		if p.mPiece != nil {
			p.mPiece.Release()
		}
		if p.dPiece != nil {
			p.dPiece.Release()
		}
	}
	c.filled.Store(0)
	c.muRemove.Unlock()

	c.muReaders.Lock()
	c.readers = nil
	c.muReaders.Unlock()

	c.muActive.Lock()
	c.activePieces = nil
	c.muActive.Unlock()

	utils.FreeOSMemGCThrottled()
	return nil
}

func (c *Cache) removePiece(piece *Piece) {
	if !c.isClosed.Load() {
		piece.Release()
	}
}

func (c *Cache) AdjustRA(readahead int64) {
	if c == nil {
		return
	}
	if settings.BTsets != nil && settings.BTsets.CacheSize == 0 {
		c.capacity = readahead * 3
	}
	c.applyZones(readahead)
}

// applyZones gives each active reader a share of the cache as its download zone.
func (c *Cache) applyZones(readaheadBudget int64) {
	if c == nil || c.isClosed.Load() {
		return
	}
	c.muReaders.Lock()
	readers := make([]*Reader, 0, len(c.readers))
	for r := range c.readers {
		if r != nil && r.using() && !r.closed() {
			readers = append(readers, r)
		}
	}
	c.muReaders.Unlock()
	n := len(readers)
	if n == 0 {
		return
	}
	next, rah, zone := readerWindows(c.capacity, n, readAheadPercent(), readaheadBudget)
	for _, r := range readers {
		r.SetZones(next, rah, zone)
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
	if c == nil || c.isClosed.Load() {
		return cState
	}

	piecesState := make(map[int]state.ItemState)
	var fill int64

	readersState := make([]*state.ReaderState, 0)
	priorityWindow := make(map[int]struct{})

	if c.Readers() > 0 {
		for _, r := range c.snapshotReaders() {
			// Closed readers are gone; a nil file cannot yield a piece range
			// unless tests set overrideLen.
			if r.closed() || (r.file == nil && r.overrideLen <= 0) {
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
				Active: r.using(),
			})
			if !r.using() {
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
	}

	c.muActive.Lock()
	activeIDs := make([]int, 0, len(c.activePieces))
	for id := range c.activePieces {
		activeIDs = append(activeIDs, id)
	}
	c.muActive.Unlock()

	stale := make([]int, 0)
	for _, id := range activeIDs {
		if c.isClosed.Load() {
			return cState
		}
		p, ok := c.pieces[id]
		if !ok || p == nil {
			stale = append(stale, id)
			continue
		}
		sz, _, _ := p.snapshotStats()
		if sz <= 0 {
			stale = append(stale, id)
			continue
		}
		fill += sz
		priority := 0
		if tor := c.torrentPtr(); tor != nil {
			priority = int(tor.PieceState(p.Id).Priority)
		}
		plen := c.pieceByteLength(p.Id)
		piecesState[p.Id] = state.ItemState{
			Id:     p.Id,
			Size:   sz,
			Length: plen,
			// Completed follows bytes present — not MarkComplete alone (avoids green empty cells).
			Completed: plen > 0 && sz >= plen,
			Priority:  priority,
		}
	}
	for _, id := range stale {
		c.notePieceEmpty(id)
	}

	// Emit Size==0 pieces inside the reader priority window so the web snake
	// can show H/R/N/A labels before bytes arrive.
	if tor := c.torrentPtr(); tor != nil {
		for id := range priorityWindow {
			if _, exists := piecesState[id]; exists {
				continue
			}
			p, ok := c.pieces[id]
			if !ok || p == nil {
				continue
			}
			sz, _, _ := p.snapshotStats()
			plen := c.pieceByteLength(id)
			piecesState[id] = state.ItemState{
				Id:        id,
				Size:      sz,
				Length:    plen,
				Completed: plen > 0 && sz >= plen,
				Priority:  int(tor.PieceState(id).Priority),
			}
		}
	}

	if !c.isClosed.Load() {
		c.filled.Store(fill)
	}
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
	if c.isClosed.Load() {
		return
	}
	// TryLock serializes eviction; Close takes the same muRemove around Release.
	if !c.muRemove.TryLock() {
		return
	}
	defer c.muRemove.Unlock()

	remPieces := c.getRemPieces()
	filled := c.filled.Load()
	if filled > c.capacity {
		rems := (filled-c.capacity)/c.pieceLength + 1
		for _, p := range remPieces {
			c.removePiece(p)
			rems--
			if rems <= 0 {
				utils.FreeOSMemGCThrottled()
				return
			}
		}
	}
}

func (c *Cache) snapshotReaders() []*Reader {
	if c == nil {
		return nil
	}
	c.muReaders.Lock()
	defer c.muReaders.Unlock()
	if c.readers == nil {
		return nil
	}
	readers := make([]*Reader, 0, len(c.readers))
	for r := range c.readers {
		readers = append(readers, r)
	}
	return readers
}

func (c *Cache) getRemPieces() []*Piece {
	if c.isClosed.Load() {
		return nil
	}

	readers := c.snapshotReaders()

	// Collect read ranges from active readers
	ranges := make([]Range, 0)
	for _, r := range readers {
		if r.closed() {
			continue
		}
		r.checkReader()
		if r.using() {
			ranges = append(ranges, r.getPiecesRange())
		}
	}
	ranges = mergeRange(ranges)

	c.muActive.Lock()
	activeIDs := make([]int, 0, len(c.activePieces))
	for id := range c.activePieces {
		activeIDs = append(activeIDs, id)
	}
	c.muActive.Unlock()

	piecesRemove := make([]*Piece, 0)
	fill := int64(0)

	// Only filled pieces can be evicted — O(active), not O(pieceCount).
	for _, id := range activeIDs {
		if c.isClosed.Load() {
			return nil
		}
		p, ok := c.pieces[id]
		if !ok || p == nil {
			continue
		}
		sz, _, _ := p.snapshotStats()
		if sz > 0 {
			fill += sz
		}
		if c.pieceEvictable(p, ranges) {
			piecesRemove = append(piecesRemove, p)
		}
	}

	// Complete pieces first, oldest first; a piece still downloading goes last. The torrent
	// client never re-requests chunks it has received, so evicting their bytes makes a
	// responsive reader that comes back to them read zeros.
	sort.Slice(piecesRemove, func(i, j int) bool {
		_, ci, ai := piecesRemove[i].snapshotStats()
		_, cj, aj := piecesRemove[j].snapshotStats()
		if ci != cj {
			return ci
		}
		return ai < aj
	})

	if !c.isClosed.Load() {
		c.filled.Store(fill)
	}
	return piecesRemove
}

// pieceEvictable reports whether a filled piece may be Released.
// Pieces being hashed (full Size, not yet Complete) must not be dropped —
// ReadAt would then return a short copy and anacrolix logs
// "unexpected error hashing piece ... %!s(<nil>)".
func (c *Cache) pieceEvictable(p *Piece, ranges []Range) bool {
	if p == nil {
		return false
	}
	sz, complete, _ := p.snapshotStats()
	if sz <= 0 {
		return false
	}
	plen := c.pieceByteLength(p.Id)
	if plen > 0 && sz >= plen && !complete {
		return false
	}
	if inRanges(ranges, p.Id) || c.isIdInKeepWindow(p.Id, ranges) {
		return false
	}
	return true
}

// isIdInKeepWindow is the begin/end cache keep-range. When there are no
// readers, isIdInFileBE cannot see a File, so fall back to torrent start/end.
func (c *Cache) isIdInKeepWindow(id int, ranges []Range) bool {
	if c.isIdInFileBE(ranges, id) {
		return true
	}
	if len(ranges) != 0 || c.pieceLength <= 0 {
		return false
	}
	keep := c.pieceLength
	if keep < 8<<20 {
		keep = 8 << 20
	}
	se := int(keep / c.pieceLength)
	if se < 1 {
		se = 1
	}
	if id >= 0 && id < se {
		return true
	}
	if c.pieceCount > 0 {
		es := c.pieceCount - se
		if es < 0 {
			es = 0
		}
		if id >= es && id < c.pieceCount {
			return true
		}
	}
	return false
}

func (c *Cache) isIdInFileBE(ranges []Range, id int) bool {
	// keep 8/16 MB
	FileRangeNotDelete := int64(c.pieceLength)
	if FileRangeNotDelete < 8<<20 {
		FileRangeNotDelete = 8 << 20
	}
	if c.pieceLength <= 0 {
		return false
	}

	for _, rng := range ranges {
		if rng.File == nil {
			continue
		}
		if fileEdgePiece(id, rng.File.Offset(), rng.File.Length(), c.pieceLength, FileRangeNotDelete) {
			return true
		}
	}
	return false
}

// fileEdgePiece reports whether piece id overlaps the first or last keep bytes of a file.
// The tail check is inclusive of the last piece: an exclusive end dropped that piece
// whenever the file ended on a piece boundary and the keep window was a single piece.
func fileEdgePiece(id int, fileOff, fileLen, pieceLen, keep int64) bool {
	if id < 0 || fileLen <= 0 || pieceLen <= 0 || keep <= 0 {
		return false
	}
	if keep > fileLen {
		keep = fileLen
	}
	end := fileOff + fileLen
	return pieceOverlaps(id, fileOff, fileOff+keep, pieceLen) || pieceOverlaps(id, end-keep, end, pieceLen)
}

func pieceOverlaps(id int, start, end, pieceLen int64) bool {
	if end <= start || pieceLen <= 0 {
		return false
	}
	p0 := int64(id) * pieceLen
	return p0 < end && p0+pieceLen > start
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
		if reader.using() {
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
	// Uses r.cache (not c): safe with nil Cache receiver after Torrent.drop.
	if r == nil || r.cache == nil {
		return
	}
	r.cache.muReaders.Lock()
	r.Close()
	delete(r.cache.readers, r)
	r.cache.muReaders.Unlock()
	// Reader.Close already runs getRemPieces. Rebalance the readers that remain.
	r.cache.AdjustRA(ReaderReadaheadBytes)
}

func (c *Cache) GetCapacity() int64 {
	if c == nil {
		return 0
	}
	return c.capacity
}

func (c *Cache) FilledBytes() int64 {
	if c == nil {
		return 0
	}
	return c.filled.Load()
}

func (c *Cache) PieceCount() int {
	if c == nil {
		return 0
	}
	return c.pieceCount
}

func (c *Cache) IsClosed() bool {
	if c == nil {
		return true
	}
	return c.isClosed.Load()
}
