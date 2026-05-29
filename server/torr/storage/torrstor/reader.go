package torrstor

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anacrolix/torrent"

	"server/log"
	"server/settings"
)

// readerIdleTimeout is how long a reader can be inactive before being deactivated
// when multiple readers are active. 5 minutes allows for pauses without stream drops.
const readerIdleTimeout = 300

type Reader struct {
	torrent.Reader
	// offset / readahead are read by getOffsetRange / checkReader on the
	// cache loop goroutine while Read / Seek mutate them on the HTTP
	// goroutine — must be atomic.
	offset    atomic.Int64
	readahead atomic.Int64
	file      *torrent.File

	cache *Cache
	// isClosed is read by Read / Seek and written by Close concurrently.
	isClosed atomic.Bool
	// ctx is cancelled by Close to unblock any in-flight Read.
	// anacrolix's bare reader.Close() only deletes the reader from the
	// torrent's reader-set; it does NOT signal waitReadable. Without
	// ReadContext + a cancellable context, a Read blocked in
	// waitAvailable would leak until the torrent itself is dropped.
	ctx       context.Context
	ctxCancel context.CancelFunc

	///Preload
	// lastAccess is read by checkReader (cache loop) and written by
	// Read / Seek (HTTP goroutine).
	lastAccess atomic.Int64
	isUse      bool
	mu         sync.Mutex
}

func newReader(file *torrent.File, cache *Cache) *Reader {
	r := new(Reader)
	r.file = file
	r.Reader = file.NewReader()
	r.ctx, r.ctxCancel = context.WithCancel(context.Background())

	r.SetReadahead(0)
	r.cache = cache
	r.isUse = true

	cache.muReaders.Lock()
	cache.readers[r] = struct{}{}
	cache.muReaders.Unlock()
	return r
}

func (r *Reader) Seek(offset int64, whence int) (n int64, err error) {
	if r.isClosed.Load() {
		return 0, io.ErrClosedPipe
	}
	switch whence {
	case io.SeekStart:
		r.offset.Store(offset)
	case io.SeekCurrent:
		r.offset.Add(offset)
	case io.SeekEnd:
		r.offset.Store(r.file.Length() + offset)
	}
	r.readerOn()
	n, err = r.Reader.Seek(offset, whence)
	r.offset.Store(n)
	r.lastAccess.Store(time.Now().Unix())
	return
}

func (r *Reader) Read(p []byte) (n int, err error) {
	// Distinguish "reader was closed" from a real EOF. Returning io.EOF
	// here would cause http.ServeContent to silently truncate the
	// response if Close races an in-flight Read.
	if r.isClosed.Load() {
		return 0, io.ErrClosedPipe
	}
	if r.file.Torrent() == nil || r.file.Torrent().Info() == nil {
		log.TLogln("Torrent closed and readed")
		return 0, io.ErrClosedPipe
	}
	r.readerOn()
	n, err = r.Reader.ReadContext(r.ctx, p)
	r.offset.Add(int64(n))
	r.lastAccess.Store(time.Now().Unix())
	return
}

func (r *Reader) SetReadahead(length int64) {
	if r.cache != nil && length > r.cache.capacity {
		length = r.cache.capacity
	}
	if r.isUse {
		r.Reader.SetReadahead(length)
	}
	r.readahead.Store(length)
}

func (r *Reader) Offset() int64 {
	return r.offset.Load()
}

func (r *Reader) Readahead() int64 {
	return r.readahead.Load()
}

func (r *Reader) Close() {
	// file reader close in gotorrent
	// this struct close in cache
	if !r.isClosed.CompareAndSwap(false, true) {
		// Already closed: idempotent.
		return
	}
	// Cancel the context FIRST so any in-flight ReadContext unblocks
	// promptly via anacrolix's tickleReaders path. Without this, the
	// goroutine running Read() would leak until the torrent is
	// dropped, since reader.Close() in the fork doesn't broadcast.
	if r.ctxCancel != nil {
		r.ctxCancel()
	}
	torr := r.file.Torrent()
	if torr != nil && len(torr.Files()) > 0 {
		r.Reader.Close()
	}
	go r.cache.getRemPieces()
}

func (r *Reader) getPiecesRange() Range {
	startOff, endOff := r.getOffsetRange()
	return Range{r.getPieceNum(startOff), r.getPieceNum(endOff), r.file}
}

func (r *Reader) getReaderPiece() int {
	return r.getPieceNum(r.offset.Load())
}

func (r *Reader) getReaderRAHPiece() int {
	return r.getPieceNum(r.offset.Load() + r.readahead.Load())
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

	off := r.offset.Load()
	beginOffset := off - (r.cache.capacity/readers)*(100-prc)/100
	endOffset := off + (r.cache.capacity/readers)*prc/100

	if beginOffset < 0 {
		beginOffset = 0
	}

	if endOffset > r.file.Length() {
		endOffset = r.file.Length()
	}
	return beginOffset, endOffset
}

func (r *Reader) checkReader() {
	if time.Now().Unix() > r.lastAccess.Load()+readerIdleTimeout && len(r.cache.readers) > 1 {
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
			r.Reader.Seek(r.offset.Load(), io.SeekStart)
		}
		r.SetReadahead(r.readahead.Load())
		r.isUse = true
	}
}

func (r *Reader) readerOff() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isUse {
		r.SetReadahead(0)
		r.isUse = false
		if r.offset.Load() > 0 {
			r.Reader.Seek(0, io.SeekStart)
		}
	}
}

func (r *Reader) getUseReaders() int {
	readers := 0
	if r.cache != nil {
		for reader := range r.cache.readers {
			if reader.isUse {
				readers++
			}
		}
	}
	return readers
}
