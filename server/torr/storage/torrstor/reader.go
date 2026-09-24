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

	cache    *Cache
	isClosed bool

	// inFlight counts reads and seeks inside the underlying reader; guarded by mu.
	inFlight int

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
	r.beginIO()
	defer r.endIO()
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
		r.beginIO()
		defer r.endIO()
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

		r.offset += int64(n)
		r.lastAccess = time.Now().Unix()
	} else {
		log.TLogln("Torrent closed and readed")
	}
	return
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

// beginIO claims the reader for one read or seek and resumes it if the sweep parked it.
// Registering under mu is what keeps readerOff from parking it until endIO.
func (r *Reader) beginIO() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inFlight++
	r.resume()
}

func (r *Reader) endIO() {
	r.mu.Lock()
	r.inFlight--
	r.mu.Unlock()
}

func (r *Reader) readerOn() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resume()
}

// resume must be called with mu held.
func (r *Reader) resume() {
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
	// lastAccess is stamped only when a read returns, so a read blocked on the swarm makes a
	// busy reader look idle; seeking the anacrolix reader under it moves that read.
	if r.inFlight > 0 {
		return
	}
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
