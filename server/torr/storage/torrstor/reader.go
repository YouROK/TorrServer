package torrstor

import (
	"io"
	"math"
	"sync"
	"time"

	"github.com/anacrolix/torrent"

	"server/log"
	"server/settings"
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
	// Buffer auto-measure. A client tops its buffer up in bursts rather than reading
	// evenly, so the rate is only meaningful averaged over a window. Positions are
	// sampled once a second and the buffer is derived from them; once derived it is
	// kept and sampling stops.
	posMu     sync.Mutex
	samples   []posSample
	sampledAt time.Time
	buffer    int64
	bufferSet bool
	fillSecs  float64 // how long the lead took to build, for logging
	playRate  float64 // playback speed it was measured against, for logging

	cache    *Cache
	isClosed bool

	///Preload
	lastAccess int64
	isUse      bool
	mu         sync.Mutex
}

func newReader(file *torrent.File, cache *Cache) *Reader {
	r := new(Reader)
	r.file = file
	r.Reader = file.NewReader()

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
		n, err = r.Reader.Read(p)

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
		r.offset += int64(n)
		r.lastAccess = time.Now().Unix()
	} else {
		log.TLogln("Torrent closed and readed")
	}
	return
}

// Buffer measurement tuning. A client tops its buffer up in bursts, so a rate only means
// something when averaged over a window; playback speed is taken from the tail of the
// session, once two consecutive windows agree that it has settled.
const (
	steadyPhase = 30.0    // seconds of playback averaged into one rate window
	steadyDrift = 0.4     // how much two consecutive windows may differ and still count as settled
	maxSamples  = 20 * 60 // once a second, so twenty minutes of history
	minBuffer   = 4 << 20
	maxBuffer   = 1 << 30
)

// posSample is where the reader stood at a moment in time.
type posSample struct {
	at  time.Time
	off int64
}

// indexBefore returns the latest sample at least d seconds older than samples[i].
func indexBefore(samples []posSample, i int, d float64) int {
	for j := i - 1; j >= 0; j-- {
		if samples[i].at.Sub(samples[j].at).Seconds() >= d {
			return j
		}
	}
	return -1
}

// trackPosition records the playback anchor and the buffer-fill dynamics. Called from
// Read before the offset is advanced, so r.offset is where this read started.
func (r *Reader) trackPosition(n int) {
	if n <= 0 {
		return
	}
	now := time.Now()
	if !r.anchorSet {
		r.anchor = r.offset
		r.anchorSet = true
		r.firstRead = now
		r.lastRead = now
		return
	}
	r.lastRead = now
	if r.bufferSet || now.Sub(r.sampledAt) < time.Second {
		return
	}
	r.sampledAt = now
	r.posMu.Lock()
	if len(r.samples) < maxSamples {
		r.samples = append(r.samples, posSample{at: now, off: r.offset})
	}
	r.posMu.Unlock()
}

// Anchor is the offset where playback started (client's Range target), and whether it is known.
func (r *Reader) Anchor() (int64, bool) {
	return r.anchor, r.anchorSet
}

// SessionSeconds is how long this reader has actually been streaming.
func (r *Reader) SessionSeconds() float64 {
	if !r.anchorSet {
		return 0
	}
	return r.lastRead.Sub(r.firstRead).Seconds()
}

// FillStats exposes what the estimate was derived from: the reader's lead over playback,
// how long it took to build up, and the playback rate it was measured against.
func (r *Reader) FillStats() (lead int64, leadSeconds, rate float64, measured bool) {
	r.posMu.Lock()
	defer r.posMu.Unlock()
	return r.buffer, r.fillSecs, r.playRate, r.bufferSet
}

// BufferEstimate derives how much the client keeps buffered ahead of the picture. The
// buffer is by definition how far the reader runs ahead of playback, so it is the largest
// lead seen: bytes read beyond the start, minus what playback consumed in the same time.
// Taking the maximum means a network hiccup while filling cannot cut the measurement
// short, and the burst-and-idle pattern of topping the buffer up cannot either.
// The result is kept once found, and sampling stops.
func (r *Reader) BufferEstimate() (int64, bool) {
	r.posMu.Lock()
	defer r.posMu.Unlock()
	if r.bufferSet {
		return r.buffer, true
	}
	if !r.anchorSet || len(r.samples) < 2 {
		return 0, false
	}
	samples := r.samples
	last := len(samples) - 1

	// Playback speed, from the tail of the session. Two consecutive windows have to agree,
	// which is what tells filling (still racing ahead) apart from playing at its own pace.
	recent := indexBefore(samples, last, steadyPhase)
	if recent < 0 {
		return 0, false
	}
	earlier := indexBefore(samples, recent, steadyPhase)
	if earlier < 0 {
		return 0, false
	}
	play := rateBetween(samples[recent], samples[last])
	prev := rateBetween(samples[earlier], samples[recent])
	if play <= 0 || math.Abs(prev-play) > steadyDrift*math.Max(prev, play) {
		return 0, false
	}

	var lead int64
	var leadAt time.Time
	for _, sm := range samples {
		played := int64(play * sm.at.Sub(r.firstRead).Seconds())
		if ahead := sm.off - r.anchor - played; ahead > lead {
			lead, leadAt = ahead, sm.at
		}
	}
	if lead < minBuffer || lead > maxBuffer {
		return 0, false
	}
	// An error in the playback rate is multiplied by the time it took to build the lead,
	// so watch playback for at least that long before trusting the result.
	building := leadAt.Sub(r.firstRead).Seconds()
	if samples[last].at.Sub(leadAt).Seconds() < math.Max(steadyPhase, building) {
		return 0, false
	}

	r.buffer, r.bufferSet, r.fillSecs, r.playRate = lead, true, building, play
	r.samples = nil
	return lead, true
}

func rateBetween(from, to posSample) float64 {
	seconds := to.at.Sub(from.at).Seconds()
	if seconds <= 0 {
		return 0
	}
	return float64(to.off-from.off) / seconds
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
