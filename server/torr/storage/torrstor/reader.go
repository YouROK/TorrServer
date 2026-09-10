package torrstor

import (
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anacrolix/torrent"

	"server/log"
	"server/settings"
	"server/timeindex"
)

type Reader struct {
	torrent.Reader
	offset    int64
	readahead int64
	file      *torrent.File

	// Playback position tracking (resume feature).
	// anchor is the offset of the first actual Read — i.e. the client's Range target.
	// It is captured on Read, not Seek: http.ServeContent always seeks to EOF then to 0
	// to measure the size before seeking to the requested range, so Seek is not a
	// reliable signal of where playback really starts.
	anchor    int64
	anchorSet bool
	firstRead time.Time
	lastRead  time.Time
	posMu     sync.Mutex

	// index turns the byte offsets this reader hands out into playback time; feeder is this
	// connection's own way into it, filled from the bytes as they go past.
	index  *timeindex.Index
	feeder *timeindex.Feeder

	// Where the picture is, in film time, tracked as the session runs. See progress.go.
	pos progress

	// Which side each read waited on, added up between one step of the reckoning and the next.
	// Guarded by posMu, like everything else the tracker reads.
	clientWait time.Duration
	supplyWait time.Duration
	readEnded  time.Time

	// An identity for the cache's per-file trail, which tells this reader's entries from those
	// of the other connections a player opens to the same file.
	id int64

	cache    *Cache
	isClosed bool

	///Preload
	lastAccess int64
	isUse      bool
	mu         sync.Mutex
}

var readerSeq int64

func newReader(file *torrent.File, cache *Cache) *Reader {
	r := new(Reader)
	r.id = atomic.AddInt64(&readerSeq, 1)
	r.file = file
	r.Reader = file.NewReader()
	r.index = cache.TimeIndex(file.Path())
	r.feeder = r.index.Feeder()

	r.SetReadahead(0)
	r.cache = cache
	r.isUse = true

	cache.muReaders.Lock()
	cache.readers[r] = struct{}{}
	cache.muReaders.Unlock()
	return r
}

func (r *Reader) Seek(offset int64, whence int) (n int64, err error) {
	if r.isClosed {
		return 0, io.EOF
	}
	switch whence {
	case io.SeekStart:
		r.offset = offset
	case io.SeekCurrent:
		r.offset += offset
	case io.SeekEnd:
		r.offset = r.file.Length() + offset
	}
	r.readerOn()
	n, err = r.Reader.Seek(offset, whence)
	r.offset = n
	r.lastAccess = time.Now().Unix()
	return
}

func (r *Reader) Read(p []byte) (n int, err error) {
	err = io.EOF
	if r.isClosed {
		return
	}
	if r.file.Torrent() != nil && r.file.Torrent().Info() != nil {
		r.readerOn()
		// Which side is holding things up. Between one read returning and the next being
		// asked for, the HTTP layer is writing to the client, and if that takes a while it is
		// because the client is not draining — it has nowhere to put more, so it is full.
		// Inside the read, the wait is for the torrent to supply, and that says nothing about
		// the client at all.
		//
		// Borrowed from how congestion control keeps its bandwidth estimates honest: a sample
		// taken while the sender had nothing to send describes the sender, not the path, and
		// is marked and set aside rather than averaged in. The same distinction is what
		// separates a client that is full from one that is starving, and from outside the two
		// look identical — the head crawls either way.
		began := time.Now()
		if !r.readEnded.IsZero() {
			r.clientWait += began.Sub(r.readEnded)
		}
		n, err = r.Reader.Read(p)
		done := time.Now()
		r.supplyWait += done.Sub(began)
		r.readEnded = done

		// samsung tv fix xvid/divx
		//if r.offset == 0 && len(p) >= 192 {
		//	str := strings.ToLower(string(p[112:116]))
		//	if str == "xvid" || str == "divx" {
		//		p[112] = 0x4D // M
		//		p[113] = 0x50 // P
		//		p[114] = 0x34 // 4
		//		p[115] = 0x56 // V
		//	}
		//	str = strings.ToLower(string(p[188:192]))
		//	if str == "xvid" || str == "divx" {
		//		p[188] = 0x4D // M
		//		p[189] = 0x50 // P
		//		p[190] = 0x34 // 4
		//		p[191] = 0x56 // V
		//	}
		//}

		r.trackPosition(n)
		if r.TimeIndex() != nil {
			r.feeder.Feed(r.offset, p[:n])
		}
		r.offset += int64(n)
		r.trackShown()
		r.lastAccess = time.Now().Unix()
	} else {
		log.TLogln("Torrent closed and readed")
	}
	return
}

// trackPosition records where playback started. It is called from Read before the offset is
// advanced, so r.offset is where this read began — and the first such offset is the byte the
// client asked for, which is where its picture starts.
func (r *Reader) trackPosition(n int) {
	if n <= 0 {
		return
	}
	now := time.Now()
	// Written from the reading goroutine and read from three others — the torrent status, the
	// cache state and the saving ticker. A time.Time is several words, so an unguarded read
	// can come out torn, and the session length gates both the saving and the anti-overshoot
	// ceiling.
	r.posMu.Lock()
	defer r.posMu.Unlock()
	if !r.anchorSet {
		r.anchor = r.offset
		r.anchorSet = true
		r.firstRead = now
	}
	r.lastRead = now
}

// File is what this reader streams.
func (r *Reader) File() *torrent.File {
	return r.file
}

// TimeIndex turns byte offsets in this file into playback time. It is nil for containers
// that carry no timestamps, and when the feature was switched off as the reader opened.
func (r *Reader) TimeIndex() *timeindex.Index {
	return r.index
}

const (
	// What a connection must have delivered before what it reports is worth handing to the
	// next one. This is measured in bytes rather than in seconds, and that is the whole point:
	// a time gate looks like it separates real playback from probes, and does not. A player
	// paused the moment it opens fills its buffer and stops, and on a fast link that is over
	// in sixteen seconds — under a twenty-second gate it hands on nothing, the connection that
	// follows starts believing the client holds nothing, and the position sits a whole buffer
	// in front of the picture for the rest of the session. Measured that way it ran 25 seconds
	// ahead and stayed there, which is the one thing that must not happen.
	//
	// The distinction that does hold is size. A header read is under a megabyte, ffprobe takes
	// a few; nothing that is actually feeding a player stops this short.
	handoverMinBytes = 32 << 20
)

// trackShown moves the picture on by however much the last read allows. The reasoning lives
// in progress.go; here it is only fed the film time at the read head, once a second.
func (r *Reader) trackShown() {
	index := r.TimeIndex()
	if index == nil || !r.anchorSet {
		return
	}
	now := time.Now()

	r.posMu.Lock()
	defer r.posMu.Unlock()
	if !r.pos.set {
		// The film time this session started at has to come from a timestamp this connection
		// found itself. The index belongs to the file and keeps what earlier sessions read,
		// so asking it what plays at the byte this one began on can be answered from far
		// ahead — a rewind lands in front of everything the previous session indexed.
		sec, ok := r.feeder.First()
		if !ok {
			return
		}
		off, ok := index.OffsetAt(sec)
		if !ok {
			off = r.anchor
		}
		// What the client was already holding when this connection opened, from what the
		// previous connection to the same file left behind.
		held, box, _ := r.cache.takeOver(r.file.Path(), off)
		r.pos.start(sec, off, r.firstRead, held, box, index)
	}
	if now.Sub(r.pos.at) < time.Second {
		return
	}
	// Between the timestamps, not rounded back to the last one: the tracker measures how the
	// position moves, and a reading that advances in steps looks like a picture that keeps
	// stalling and catching up.
	if head, ok := index.TimeBetween(r.offset); ok {
		r.pos.waited(r.clientWait, r.supplyWait)
		r.clientWait, r.supplyWait = 0, 0
		r.pos.step(head, r.offset, index, now)
	}
	// Leave a trail for whatever connection comes next: where this one has read to, and how
	// much the client is holding there. Read straight off the tracker — the lock is already
	// held here, and the public form would take it a second time and stop the reader dead.
	//
	// Only from a connection that has been streaming for a while. A player opens several at
	// once — one for the header, one for the picture, a probe alongside — and they all read
	// the same file. Letting the short ones write here mixes their offsets and their holdings
	// in with the one that is actually playing, and what the next connection then inherits is
	// nobody's buffer in particular.
	if r.offset-r.anchor >= handoverMinBytes {
		sec, _ := r.pos.screen()
		r.cache.noteRead(r.file.Path(), r.id, r.pos.handOn(), r.pos.box(), r.pos.pictureAt(), sec)
	}
}

// Tick moves the reckoning on without a read, and it is not optional. The tracker is
// otherwise driven only by Read, so the one event worth noticing — the client stopping,
// which it does when it has nowhere left to put anything — is the very event that stops
// anything from noticing it. Measured on a real player through a whole paused warm-up: not
// one step took a silence branch, because not one step ran. Everything built on silence was
// alive in the tests, where the clock is turned by hand, and dead everywhere else.
func (r *Reader) Tick() {
	if r == nil || r.isClosed {
		return
	}
	r.trackShown()
}

// FurthestScreen is the furthest the picture can have reached, going by where an earlier
// connection to this file left it and the time since. It bounds a fresh session, which would
// otherwise put the picture a whole buffer ahead of itself.
func (r *Reader) FurthestScreen() (float64, bool) {
	return r.cache.FurthestScreen(r.file.Path())
}

// Playback is what the tracker knows about one connection, read under one lock, so the
// status poll and the saver take a single consistent snapshot rather than a dozen.
type Playback struct {
	Started   bool    // the client has read something, so the anchor is known
	Anchor    int64   // byte the client began at
	Head      int64   // byte the reader has handed out up to
	Screen    int64   // byte the picture is at, never behind the anchor or past the file
	ScreenSec float64 // the same in film time, when the container carries timestamps
	HasTime   bool    // ScreenSec and Buffer were measured rather than taken from the settings
	Held      float64 // film the client holds ahead of the picture, in seconds
	Buffer    int64   // the same in bytes, or the configured size when it is not measured
	Session   float64 // seconds this connection has been streaming
}

// Playback reports where this connection's picture is and what the client holds ahead of it.
// The head is ahead of the picture by whatever the client keeps buffered; when the container
// carries timestamps that is a measurement, worked out in film time and converted back
// through the file's own byte positions. Otherwise it is the configured guess, which is all a
// file without timestamps allows.
func (r *Reader) Playback() Playback {
	r.posMu.Lock()
	pb := Playback{Started: r.anchorSet, Anchor: r.anchor, Head: r.offset}
	if r.anchorSet {
		pb.Session = r.lastRead.Sub(r.firstRead).Seconds()
	}
	if r.index != nil && r.pos.set {
		pb.ScreenSec, pb.HasTime = r.pos.screenSec, true
		pb.Held = r.pos.held()
	}
	r.posMu.Unlock()

	if pb.HasTime {
		if off, ok := r.index.OffsetAt(pb.ScreenSec); ok {
			pb.Screen, pb.Buffer = off, pb.Head-off
		} else {
			pb.HasTime, pb.Held = false, 0
		}
	}
	if !pb.HasTime {
		pb.Buffer = int64(settings.BTsets.BufferSizeMB) << 20
		pb.Screen = pb.Head - pb.Buffer
	}
	pb.Screen = min(max(pb.Screen, pb.Anchor), r.file.Length())
	return pb
}

// getScreenPiece is the piece the picture is in, as opposed to the one being read.
func (r *Reader) getScreenPiece() int {
	return r.getPieceNum(r.Playback().Screen)
}

func (r *Reader) SetReadahead(length int64) {
	if r.cache != nil && length > r.cache.capacity {
		length = r.cache.capacity
	}
	if r.isUse {
		r.Reader.SetReadahead(length)
	}
	r.readahead = length
}

func (r *Reader) Offset() int64 {
	return r.offset
}

func (r *Reader) Readahead() int64 {
	return r.readahead
}

func (r *Reader) Close() {
	// file reader close in gotorrent
	// this struct close in cache
	r.isClosed = true
	if len(r.file.Torrent().Files()) > 0 {
		r.Reader.Close()
	}
	go r.cache.getRemPieces()
}

func (r *Reader) getPiecesRange() Range {
	startOff, endOff := r.getOffsetRange()
	return Range{r.getPieceNum(startOff), r.getPieceNum(endOff), r.file}
}

func (r *Reader) getReaderPiece() int {
	return r.getPieceNum(r.offset)
}

func (r *Reader) getReaderRAHPiece() int {
	return r.getPieceNum(r.offset + r.readahead)
}

func (r *Reader) getPieceNum(offset int64) int {
	return int((offset + r.file.Offset()) / r.cache.pieceLength)
}

func (r *Reader) getOffsetRange() (int64, int64) {
	prc := int64(settings.BTsets.ReaderReadAHead)
	readers := int64(r.getUseReaders())
	if readers == 0 {
		readers = 1
	}

	beginOffset := r.offset - (r.cache.capacity/readers)*(100-prc)/100
	endOffset := r.offset + (r.cache.capacity/readers)*prc/100

	if beginOffset < 0 {
		beginOffset = 0
	}

	if endOffset > r.file.Length() {
		endOffset = r.file.Length()
	}
	return beginOffset, endOffset
}

func (r *Reader) checkReader() {
	if time.Now().Unix() > r.lastAccess+60 && r.cache.Readers() > 1 {
		r.readerOff()
	} else {
		r.readerOn()
	}
}

func (r *Reader) readerOn() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.isUse {
		if pos, err := r.Reader.Seek(0, io.SeekCurrent); err == nil && pos == 0 {
			r.Reader.Seek(r.offset, io.SeekStart)
		}
		r.SetReadahead(r.readahead)
		r.isUse = true
	}
}

func (r *Reader) readerOff() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isUse {
		r.SetReadahead(0)
		r.isUse = false
		if r.offset > 0 {
			r.Reader.Seek(0, io.SeekStart)
		}
	}
}

func (r *Reader) getUseReaders() int {
	return r.cache.GetUseReaders()
}
