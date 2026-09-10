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
	"server/timeindex"
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

	// pieces content is immutable after Init; muPieces guards the map
	// reference itself (nilled in Close), so holders of a snapshot may
	// safely iterate it without the lock
	pieces   map[int]*Piece
	muPieces sync.RWMutex

	readers   map[*Reader]struct{}
	muReaders sync.RWMutex

	// One time index per file, not per reader: a player opens a separate connection for the
	// header and another for playback, and the header is where some formats keep the only
	// copy of what their timestamps mean.
	indexes   map[string]*timeindex.Index
	muIndexes sync.Mutex

	// What a client was holding when its connection ended, kept per file so the next one can
	// pick it up. A player that pauses long enough loses the connection and opens another,
	// and the new one would otherwise start from the only assumption available to it — that
	// nothing is buffered — while the player still holds everything it had.
	handovers  map[string]*handover
	muHandover sync.Mutex

	isRemove atomic.Bool
	isClosed atomic.Bool
	muRemove sync.Mutex
	// muPrio serializes clearPriority and setLoadPriority so that the priority
	// reset of a reader that has just closed cannot wipe the priorities a
	// freshly created reader has already set.
	muPrio  sync.Mutex
	torrent *torrent.Torrent
}

func NewCache(capacity int64, storage *Storage) *Cache {
	ret := &Cache{
		capacity:  capacity,
		filled:    0,
		pieces:    make(map[int]*Piece),
		storage:   storage,
		readers:   make(map[*Reader]struct{}),
		indexes:   make(map[string]*timeindex.Index),
		handovers: make(map[string]*handover),
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

	go c.priorityWatchdog()
}

// priorityWatchdog re-arms piece priorities while readers are active.
//
// setLoadPriority is only reached through the cache cleanup path, which is
// driven by piece reads and writes (see mempiece.go and diskpiece.go). Should
// priorities ever end up cleared while a reader still needs data, nothing is
// downloaded, so no piece I/O happens, so cleanup never runs and the
// priorities are never restored - the torrent stalls indefinitely with peers
// connected. Re-arming them periodically breaks that cycle regardless of how
// the priorities were lost.
func (c *Cache) priorityWatchdog() {
	for {
		time.Sleep(5 * time.Second)
		if c.isClosed.Load() {
			return
		}
		if c.torrent == nil {
			continue
		}
		if c.GetUseReaders() > 0 {
			c.getRemPieces()
		}
	}
}

func (c *Cache) SetTorrent(torr *torrent.Torrent) {
	c.torrent = torr
}

func (c *Cache) getPieces() map[int]*Piece {
	c.muPieces.RLock()
	defer c.muPieces.RUnlock()
	return c.pieces
}

func (c *Cache) readersSnapshot() []*Reader {
	c.muReaders.RLock()
	defer c.muReaders.RUnlock()
	list := make([]*Reader, 0, len(c.readers))
	for r := range c.readers {
		list = append(list, r)
	}
	return list
}

// TimeIndex is the shared time index for a file, created on first use. It is nil for
// containers that carry no timestamps to read.
func (c *Cache) TimeIndex(path string) *timeindex.Index {
	if c == nil || !settings.BTsets.SmartTimecode {
		return nil
	}
	c.muIndexes.Lock()
	defer c.muIndexes.Unlock()
	if c.indexes == nil {
		return nil // the cache is closing; assigning here would panic on a nil map
	}
	ix, ok := c.indexes[path]
	if !ok {
		ix = timeindex.New(path)
		c.indexes[path] = ix
	}
	return ix
}

// handover is where a file was last being read, and how much the client reading it was
// holding at the time.
type handover struct {
	by      int64     // which connection these readings belong to
	holding int64     // what the client had in hand there, in bytes
	size    int64     // and the size of its buffer, which an interruption does not change
	picture int64     // byte the picture was at, as last reported
	sec     float64   // and the film time there, kept only for the moves-no-faster-than-the-clock bound
	at      time.Time // when that was
}

// noteRead records where a connection has read to, what it was holding, and where the
// picture was.
func (c *Cache) noteRead(path string, by, holding, size, picture int64, sec float64) {
	if c == nil {
		return
	}
	c.muHandover.Lock()
	defer c.muHandover.Unlock()
	if c.handovers == nil {
		return
	}
	h := c.handovers[path]
	if h == nil {
		h = &handover{}
		c.handovers[path] = h
	}
	h.at = time.Now()
	// The most it was seen holding, not what it holds at this instant. Handing over the
	// instant value was tried and reverted: it left the next connection with nothing to
	// inherit, and a connection that believes the client holds nothing puts the picture at
	// the read head — a whole buffer past what is on screen. Overstating leaves the position
	// behind instead, which is the side to be wrong on.
	//
	// The most within one connection, though, not for as long as the file stays open. Kept
	// across connections it is a ratchet: the largest reading of all is the one taken during
	// the opening fill, before anything has settled, and it then stands as the ceiling for
	// every session that follows however long they run. Measured, the first session peaked at
	// 47.7 seconds against a client holding 37, and two reconnections later the figure had
	// climbed from 37 to 42 to 48 with the ceiling never once binding.
	if h.by != by {
		h.by, h.holding, h.picture, h.sec = by, holding, picture, sec
		if size > h.size {
			h.size = size // the box belongs to the device, not to the connection
		}
		return
	}
	if holding > h.holding {
		h.holding = holding
	}
	if size > h.size {
		h.size = size
	}
	// The picture as it stands, not the furthest it has ever been said to be. A picture only
	// moves forward, so keeping the largest reading looks harmless — but it makes a ratchet of
	// every excursion: one reading that ran ahead is kept for the rest of the file's life and
	// handed to the next connection, which then holds its position there until the head
	// catches up. Measured at seventeen seconds in front, decaying over fifteen.
	h.picture, h.sec = picture, sec
}

// How far outside the stretch the previous connection was serving a new one may begin and
// still be taken for the same client coming back. It only has to be wider than the slop
// between where a player says it resumes and where it actually asks for; a seek is a jump of
// minutes and lands nowhere near.
const rejoinMargin = 256 << 20

// takeOver reports how much the client watching this file was last seen holding, so a
// connection opening now can start from that rather than from nothing. startOff is the byte
// the new connection begins at.
//
// Whether the buffer belongs to whoever turned up is decided by where they turned up. A
// player that lost its connection carries on from somewhere inside what it already had: at
// its own picture at the earliest, at the byte the server had reached at the latest. Anything
// outside that stretch is a seek or another device, and inherits nothing.
//
// Deciding it by how fast film then arrived was tried first and is gone. The reasoning was
// sound — a client already full has nowhere to put a burst — but the measurement is not there
// to support it. On eight minutes of undisturbed playback the fill rate ran from 0.46 to 1.69
// with nothing whatsoever happening, and the rule meant to catch a client filling from empty
// fired three times. Every reconnection was therefore refused, and the buffer remeasured from
// a fill that counted the refilled pipe as the client's own: 37 seconds became 56.
func (c *Cache) takeOver(path string, startOff int64) (held, size int64, ok bool) {
	if c == nil {
		return 0, 0, false
	}
	c.muHandover.Lock()
	defer c.muHandover.Unlock()
	h := c.handovers[path]
	if h == nil || h.holding <= 0 {
		return 0, 0, false
	}
	// No expiry. A buffer size is a property of the device, and an hour's pause does not
	// change it — while letting the record lapse would have the next connection assume the
	// client holds nothing, which puts the picture at the read head, a whole buffer past
	// where it really is. The record lives as long as the torrent's cache, and goes when
	// that does.
	from, to := h.picture-rejoinMargin, h.picture+h.holding+rejoinMargin
	if startOff < from || startOff > to {
		log.TLogln("[Handover] starting at", startOff>>20, "MB, outside", max(from, 0)>>20, "..", to>>20,
			"MB — not the same playback, nothing carried over")
		return 0, 0, false
	}
	// Not what the last connection reckoned the client was holding — where it reckoned the
	// picture was. What lies between there and the byte this connection opens on is what the
	// client still has, and that is a different quantity.
	//
	// The difference is everything in transit when the line went down: what the server had
	// read and written out but the player had not yet received. Measured on a 300MB player it
	// was 368MB — as much again as the buffer itself, sitting in socket queues between the
	// two. It is counted as held while the connection lives, correctly, since it is film that
	// has left the head and not been shown. But it dies with the connection, and the player
	// asks for it again. Carrying the old figure across therefore counted it twice: 37.9
	// seconds became 61.5 the moment the line was restored, and stayed there.
	//
	// Where the player resumes says exactly how much survived, because that is where what it
	// holds runs out. Never more than the connection before was holding — the same figure
	// cannot grow by being handed on.
	held = startOff - h.picture
	if held > h.holding {
		held = h.holding
	}
	if held <= 0 {
		log.TLogln("[Handover] resumes at the picture itself — nothing was still in hand")
		return 0, 0, false
	}
	// Two different quantities, and conflating them is a slow leak in the dangerous direction.
	// What survived the line going down says where the picture is. The size of the box says
	// how much the device holds, and a dropped connection does not shrink a device. Carrying
	// only what survived and calling it the new size cost about seven percent per
	// reconnection — 302MB, then 286, then 265 — and every megabyte lost that way moves the
	// picture forward.
	log.TLogln("[Handover] carried over:", held>>20, "MB still in hand of a", h.size>>20,
		"MB buffer, last seen", int(time.Since(h.at).Seconds()), "s ago")
	return held, h.size, true
}

// FurthestScreen is the furthest the picture can have reached by now, going by where the last
// connection to this file left it and how much time has passed since.
//
// A connection that opens knows nothing of what came before it, and its own reckoning starts
// from the assumption that the client holds nothing — so the moment a player drops its
// connection and opens another, the picture appears to leap forward by a whole buffer. What
// cannot happen, whatever the client is doing, is the picture moving faster than the clock.
// That bound survives any reconnection, because it needs to know nothing about the client.
func (c *Cache) FurthestScreen(path string) (float64, bool) {
	if c == nil {
		return 0, false
	}
	c.muHandover.Lock()
	defer c.muHandover.Unlock()
	h := c.handovers[path]
	if h == nil || h.sec <= 0 {
		return 0, false
	}
	return h.sec + time.Since(h.at).Seconds(), true
}

// ReaderList is a snapshot of the readers streaming from this cache.
func (c *Cache) ReaderList() []*Reader {
	if c == nil {
		return nil
	}
	return c.readersSnapshot()
}

func (c *Cache) Piece(m metainfo.Piece) storage.PieceImpl {
	if val, ok := c.getPieces()[m.Index()]; ok {
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
	c.isClosed.Store(true)

	c.storage.removeCache(c.hash)

	if settings.BTsets.RemoveCacheOnDrop {
		name := filepath.Join(settings.BTsets.TorrentsSavePath, c.hash.HexString())
		if name != "" && name != "/" {
			for _, v := range c.getPieces() {
				if v.dPiece != nil {
					os.Remove(v.dPiece.name)
				}
			}
			os.Remove(name)
		}
	}

	c.muReaders.Lock()
	c.readers = nil
	c.muReaders.Unlock()

	c.muIndexes.Lock()
	c.indexes = nil
	c.muIndexes.Unlock()

	c.muHandover.Lock()
	c.handovers = nil
	c.muHandover.Unlock()

	c.muPieces.Lock()
	c.pieces = nil
	c.muPieces.Unlock()

	utils.FreeOSMemGC()
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
	if settings.BTsets.CacheSize == 0 {
		c.capacity = readahead * 3
	}
	for _, r := range c.readersSnapshot() {
		r.SetReadahead(readahead)
	}
}

func (c *Cache) GetState() *state.CacheState {
	cState := new(state.CacheState)

	piecesState := make(map[int]state.ItemState, 0)
	var fill int64 = 0

	for _, p := range c.getPieces() {
		if p.Size > 0 {
			fill += p.Size
			piecesState[p.Id] = state.ItemState{
				Id:        p.Id,
				Size:      p.Size,
				Length:    c.pieceLength,
				Completed: p.Complete,
				Priority:  int(c.torrent.PieceState(p.Id).Priority),
			}
		}
	}

	readersState := make([]*state.ReaderState, 0)

	for _, r := range c.readersSnapshot() {
		rng := r.getPiecesRange()
		pc := r.getReaderPiece()
		readersState = append(readersState, &state.ReaderState{
			Start:  rng.Start,
			End:    rng.End,
			Reader: pc,
			Screen: r.getScreenPiece(),
		})
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
	if c.isRemove.Load() || c.isClosed.Load() {
		return
	}

	// Protection against concurrent deletion
	if !c.muRemove.TryLock() {
		return // Cleanup is already in progress in another goroutine
	}
	defer c.muRemove.Unlock()

	c.isRemove.Store(true)
	defer func() { c.isRemove.Store(false) }()

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
	readers := c.readersSnapshot()

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
	for id, p := range c.getPieces() {
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
	readers := c.readersSnapshot()
	pieces := c.getPieces()
	if len(readers) == 0 || pieces == nil {
		return
	}
	c.muPrio.Lock()
	defer c.muPrio.Unlock()
	for _, r := range readers {
		if !r.isUse {
			continue
		}
		if c.isIdInFileBE(ranges, r.getReaderPiece()) {
			continue
		}
		readerPos := r.getReaderPiece()
		readerRAHPos := r.getReaderRAHPiece()
		end := r.getPiecesRange().End
		count := settings.BTsets.ConnectionsLimit / len(readers) // max concurrent loading blocks
		limit := 0
		for i := readerPos; i < end && limit < count; i++ {
			if !pieces[i].Complete {
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
	c.muReaders.RLock()
	defer c.muReaders.RUnlock()
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
	c.muReaders.RLock()
	defer c.muReaders.RUnlock()
	return len(c.readers)
}

func (c *Cache) CloseReader(r *Reader) {
	r.cache.muReaders.Lock()
	delete(r.cache.readers, r)
	r.cache.muReaders.Unlock()
	// Reader.Close touches anacrolix internals, keep it outside muReaders
	r.Close()
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
	for _, r := range c.readersSnapshot() {
		r.checkReader()
		if r.isUse {
			ranges = append(ranges, r.getPiecesRange())
		}
	}
	ranges = mergeRange(ranges)

	for id := range c.getPieces() {
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
