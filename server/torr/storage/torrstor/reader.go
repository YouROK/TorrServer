package torrstor

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/anacrolix/torrent"

	"server/log"
	"server/settings"
)

// Reader keeps a torrent reader registered until its I/O has drained.
// Callers must serialize reads and seeks on the same Reader.
type Reader struct {
	backend   torrent.Reader
	offset    int64
	readahead int64
	file      *torrent.File

	cache    *Cache
	isClosed bool

	// Cancel reads on either caller cancellation or reader close.
	ctx        context.Context
	cancelRead context.CancelFunc

	// inFlight counts reads and seeks inside the underlying reader; guarded by mu.
	inFlight int

	///Preload
	lastAccess int64
	isUse      bool
	mu         sync.Mutex
}

func newReader(file *torrent.File, cache *Cache) *Reader {
	if cache.isClosed.Load() {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	r := &Reader{
		backend:    file.NewReader(),
		file:       file,
		cache:      cache,
		isUse:      true,
		ctx:        ctx,
		cancelRead: cancel,
	}

	cache.muReaders.Lock()
	if cache.isClosed.Load() {
		cache.muReaders.Unlock()
		cancel()
		_ = r.backend.Close()
		return nil
	}
	cache.readers[r] = struct{}{}
	cache.muReaders.Unlock()
	return r
}

func (r *Reader) Seek(offset int64, whence int) (n int64, err error) {
	if _, ok := r.beginIO(); !ok {
		return 0, io.EOF
	}
	defer r.endIO()
	n, err = r.backend.Seek(offset, whence)
	r.mu.Lock()
	if err == nil {
		r.offset = n
	}
	r.lastAccess = time.Now().Unix()
	r.mu.Unlock()
	return
}

func (r *Reader) Read(p []byte) (n int, err error) {
	ctx, ok := r.beginIO()
	if !ok {
		return 0, io.EOF
	}
	defer r.endIO()
	err = io.EOF
	if r.file.Torrent() != nil && r.file.Torrent().Info() != nil {
		n, err = r.backend.ReadContext(ctx, p)

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

		r.mu.Lock()
		r.offset += int64(n)
		r.lastAccess = time.Now().Unix()
		r.mu.Unlock()
	} else {
		log.TLogln("Torrent closed and readed")
	}
	return
}

// SetContext binds reads to the consumer's lifetime. Call it before starting I/O.
func (r *Reader) SetContext(ctx context.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isClosed || r.inFlight != 0 {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	r.cancelRead()
	r.ctx, r.cancelRead = context.WithCancel(ctx)
}

func (r *Reader) SetReadahead(length int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isClosed || r.cache.isClosed.Load() {
		return
	}
	if length > r.cache.capacity {
		length = r.cache.capacity
	}
	if r.isUse {
		r.backend.SetReadahead(length)
	}
	r.readahead = length
}

func (r *Reader) SetResponsive() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.isClosed && !r.cache.isClosed.Load() {
		r.backend.SetResponsive()
	}
}

func (r *Reader) Offset() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.offset
}

func (r *Reader) Readahead() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.readahead
}

func (r *Reader) active() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.isUse && (!r.isClosed || r.inFlight > 0)
}

func (r *Reader) Close() {
	r.mu.Lock()
	if r.isClosed {
		r.mu.Unlock()
		return
	}
	r.isClosed = true
	cancel := r.cancelRead
	finish := r.inFlight == 0
	r.mu.Unlock()

	cancel()
	if finish {
		r.finishClose()
	}
}

// finishClose is called once, after the last operation has left the backend.
func (r *Reader) finishClose() {
	_ = r.backend.Close()
	c := r.cache
	c.muReaders.Lock()
	delete(c.readers, r)
	c.muReaders.Unlock()
	if !c.isClosed.Load() {
		go c.clearPriority()
	}
}

func (r *Reader) getPiecesRange() Range {
	startOff, endOff := r.getOffsetRange()
	return Range{r.getPieceNum(startOff), r.getPieceNum(endOff), r.file}
}

func (r *Reader) getReaderPiece() int {
	return r.getPieceNum(r.Offset())
}

func (r *Reader) getReaderRAHPiece() int {
	r.mu.Lock()
	defer r.mu.Unlock()
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

	offset := r.Offset()
	beginOffset := offset - (r.cache.capacity/readers)*(100-prc)/100
	endOffset := offset + (r.cache.capacity/readers)*prc/100

	if beginOffset < 0 {
		beginOffset = 0
	}

	if endOffset > r.file.Length() {
		endOffset = r.file.Length()
	}
	return beginOffset, endOffset
}

func (r *Reader) checkReader() {
	multipleReaders := r.cache.Readers() > 1
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isClosed || r.cache.isClosed.Load() {
		return
	}
	if multipleReaders && time.Now().Unix() > r.lastAccess+60 {
		r.park()
		return
	}
	r.resume()
}

// beginIO claims the reader for one read or seek and resumes it if the sweep parked it.
// Registering under mu keeps the idle sweep from parking it until endIO.
func (r *Reader) beginIO() (context.Context, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isClosed || r.cache.isClosed.Load() {
		return nil, false
	}
	r.inFlight++
	r.resume()
	return r.ctx, true
}

func (r *Reader) endIO() {
	r.mu.Lock()
	r.inFlight--
	finish := r.inFlight == 0 && r.isClosed
	r.mu.Unlock()
	if finish {
		r.finishClose()
	}
}

// resume must be called with mu held.
func (r *Reader) resume() {
	if !r.isUse {
		if pos, err := r.backend.Seek(0, io.SeekCurrent); err == nil && pos == 0 {
			r.backend.Seek(r.offset, io.SeekStart)
		}
		r.backend.SetReadahead(r.readahead)
		r.isUse = true
	}
}

// park must be called with mu held.
func (r *Reader) park() {
	// lastAccess is stamped only when a read returns, so a read blocked on the swarm makes a
	// busy reader look idle. Seeking the anacrolix reader under it moves that read.
	if r.inFlight > 0 {
		return
	}
	if r.isUse {
		r.backend.SetReadahead(0)
		r.isUse = false
		if r.offset > 0 {
			r.backend.Seek(0, io.SeekStart)
		}
	}
}

func (r *Reader) getUseReaders() int {
	return r.cache.GetUseReaders()
}
