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

type Reader struct {
	torrent.Reader
	offset    int64
	readahead int64
	file      *torrent.File
	// overrideLen is used by tests when file is nil (file.Length is unexported).
	overrideLen int64

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
	// Without this a fresh reader looks idle since epoch and is switched off at once.
	r.lastAccess = time.Now().Unix()

	cache.muReaders.Lock()
	cache.readers[r] = struct{}{}
	cache.muReaders.Unlock()
	return r
}

func (r *Reader) fileLen() int64 {
	if r.file != nil {
		return r.file.Length()
	}
	return r.overrideLen
}

func (r *Reader) fileOffset() int64 {
	if r.file != nil {
		return r.file.Offset()
	}
	return 0
}

func (r *Reader) Seek(offset int64, whence int) (n int64, err error) {
	if r.closed() {
		return 0, io.EOF
	}
	switch whence {
	case io.SeekStart:
		r.setOffset(offset)
	case io.SeekCurrent:
		r.setOffset(r.Offset() + offset)
	case io.SeekEnd:
		// ServeContent probes size with Seek(0, SeekEnd). Return the length but
		// leave the logical playhead alone — otherwise the snake jumps to EOF
		// until the following SeekStart, and we must not pull pieces from the
		// tail of a 100GB+ file.
		if offset == 0 {
			r.readerOn()
			r.mu.Lock()
			r.lastAccess = time.Now().Unix()
			r.mu.Unlock()
			return r.fileLen(), nil
		}
		r.setOffset(r.fileLen() + offset)
	}
	r.beginIO()
	defer r.endIO()
	n, err = r.Reader.Seek(offset, whence)
	r.touch(n)
	return
}

func (r *Reader) Read(p []byte) (n int, err error) {
	return r.ReadContext(context.Background(), p)
}

// ReadContext is Read, but it returns when ctx is canceled. anacrolix Reader.Read
// waits on context.Background, and Reader.Close does not wake that wait, so a
// client that has already gone away (ffprobe timeout, player stop) would pin the
// reader and its piece window until the torrent produced bytes.
func (r *Reader) ReadContext(ctx context.Context, p []byte) (n int, err error) {
	err = io.EOF
	if r.closed() {
		return
	}
	if r.file != nil && r.file.Torrent() != nil && r.file.Torrent().Info() != nil {
		r.beginIO()
		defer r.endIO()
		n, err = r.Reader.ReadContext(ctx, p)
		r.mu.Lock()
		r.offset += int64(n)
		r.lastAccess = time.Now().Unix()
		r.mu.Unlock()
	} else {
		log.TLogln("Torrent closed and readed")
	}
	return
}

func (r *Reader) SetReadahead(length int64) {
	if r.cache != nil && length > r.cache.capacity {
		length = r.cache.capacity
	}
	r.mu.Lock()
	r.readahead = length
	rd := r.Reader
	use := r.isUse && rd != nil
	r.mu.Unlock()
	if use {
		rd.SetReadahead(length)
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

func (r *Reader) Close() {
	// file reader close in gotorrent
	// this struct close in cache
	r.mu.Lock()
	r.isClosed = true
	r.mu.Unlock()
	if r.Reader != nil && r.file != nil && r.file.Torrent() != nil && len(r.file.Torrent().Files()) > 0 {
		_ = r.Reader.Close()
	}
	if r.cache != nil {
		go r.cache.getRemPieces()
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
	off := r.offset + r.readahead
	r.mu.Unlock()
	return r.getPieceNum(off)
}

func (r *Reader) getPieceNum(offset int64) int {
	return int((offset + r.fileOffset()) / r.cache.pieceLength)
}

func (r *Reader) getOffsetRange() (int64, int64) {
	prc := int64(95)
	if settings.BTsets != nil {
		prc = int64(settings.BTsets.ReaderReadAHead)
	}
	readers := int64(r.getUseReaders())
	if readers == 0 {
		readers = 1
	}

	off := r.Offset()
	beginOffset := off - (r.cache.capacity/readers)*(100-prc)/100
	endOffset := off + (r.cache.capacity/readers)*prc/100

	if beginOffset < 0 {
		beginOffset = 0
	}

	if endOffset > r.fileLen() {
		endOffset = r.fileLen()
	}
	return beginOffset, endOffset
}

// minReaderIdleTimeout keeps a paused player from dropping its cache window immediately.
const minReaderIdleTimeout = 60

func (r *Reader) idleTimeout() int64 {
	timeout := int64(0)
	if settings.BTsets != nil {
		timeout = int64(settings.BTsets.TorrentDisconnectTimeout)
	}
	if timeout < minReaderIdleTimeout {
		timeout = minReaderIdleTimeout
	}
	return timeout
}

// checkReader flips the reader between active and idle. A lone reader must be
// able to go idle too, otherwise a stopped player keeps driving the playhead.
func (r *Reader) checkReader() {
	r.mu.Lock()
	last := r.lastAccess
	r.mu.Unlock()
	if time.Now().Unix() > last+r.idleTimeout() {
		r.readerOff()
	} else {
		r.readerOn()
	}
}

func (r *Reader) closed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.isClosed
}

func (r *Reader) using() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.isUse
}

func (r *Reader) setOffset(n int64) {
	r.mu.Lock()
	r.offset = n
	r.mu.Unlock()
}

func (r *Reader) touch(n int64) {
	r.mu.Lock()
	r.offset = n
	r.lastAccess = time.Now().Unix()
	r.mu.Unlock()
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

// resume must be called with mu held. It drops the lock around the underlying
// reader: that Seek can run cache cleanup, which locks mu again.
func (r *Reader) resume() {
	if r.isUse {
		return
	}
	rd := r.Reader
	off := r.offset
	rah := r.readahead
	r.isUse = true
	if rd == nil {
		return
	}
	r.mu.Unlock()
	if pos, err := rd.Seek(0, io.SeekCurrent); err == nil && pos == 0 {
		_, _ = rd.Seek(off, io.SeekStart)
	}
	rd.SetReadahead(rah)
	r.mu.Lock()
}

func (r *Reader) readerOff() {
	r.mu.Lock()
	// lastAccess is stamped only when a read returns, so a read blocked on the swarm makes a
	// busy reader look idle; seeking the anacrolix reader under it moves that read.
	if r.inFlight > 0 || !r.isUse {
		r.mu.Unlock()
		return
	}
	r.readahead = 0
	r.isUse = false
	off := r.offset
	rd := r.Reader
	r.mu.Unlock()
	if rd == nil {
		return
	}
	rd.SetReadahead(0)
	if off > 0 {
		_, _ = rd.Seek(0, io.SeekStart)
	}
}

func (r *Reader) getUseReaders() int {
	if r.cache == nil {
		return 0
	}
	return r.cache.GetUseReaders()
}
