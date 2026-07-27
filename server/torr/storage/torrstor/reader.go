package torrstor

import (
	"io"
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
	// Buffer auto-measure: the client fills its buffer as fast as it can, then stalls
	// and settles to the playback rate. Bytes read up to the first stall minus what was
	// played during that time is the client's buffer size.
	fillDone      bool
	fillBytes     int64
	fillSeconds   float64
	postFillBytes int64
	postFillTime  time.Time

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

// stallGap: a pause between reads longer than this means the client stopped pulling
// data because its buffer is full. steadyPhase is how much playing-at-its-own-pace has
// to be observed after that before the playback rate is trustworthy.
const (
	stallGap    = 2 * time.Second
	steadyPhase = 30.0 // seconds
)

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
	if !r.fillDone && now.Sub(r.lastRead) > stallGap {
		// first stall => the client buffer just got full
		r.fillDone = true
		r.fillBytes = r.offset - r.anchor
		r.fillSeconds = r.lastRead.Sub(r.firstRead).Seconds()
		r.postFillBytes = r.offset
		r.postFillTime = now
	}
	r.lastRead = now
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

// FillStats exposes what the buffer estimate is derived from: how much was read before
// the client first paused, over how long, and the playback rate measured afterwards.
func (r *Reader) FillStats() (fillBytes int64, fillSeconds, rate float64, filled bool) {
	if !r.fillDone {
		return 0, 0, 0, false
	}
	postSeconds := r.lastRead.Sub(r.postFillTime).Seconds()
	if postSeconds > 0 {
		rate = float64(r.offset-r.postFillBytes) / postSeconds
	}
	return r.fillBytes, r.fillSeconds, rate, true
}

// BufferEstimate infers the client's buffer size in bytes from the fill dynamics:
// after the buffer is full the client reads at the playback rate, so that rate times
// the fill duration is what was played while filling; the rest is the buffer.
// Returns false when the session is too short or the result is implausible.
func (r *Reader) BufferEstimate() (int64, bool) {
	if !r.fillDone || r.fillBytes <= 0 {
		return 0, false
	}
	postBytes := r.offset - r.postFillBytes
	postSeconds := r.lastRead.Sub(r.postFillTime).Seconds()
	// The playback rate is multiplied by the fill duration, so an error in it grows with
	// however long the fill took. Watch playback for at least as long as the fill lasted
	// (a slow line fills slowly), and never for less than steadyPhase.
	need := steadyPhase
	if r.fillSeconds > need {
		need = r.fillSeconds
	}
	if postBytes <= 0 || postSeconds < need {
		return 0, false
	}
	rate := float64(postBytes) / postSeconds
	buf := float64(r.fillBytes) - rate*r.fillSeconds
	if buf < 4<<20 || buf > 1<<30 {
		return 0, false
	}
	return int64(buf), true
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
