package torrstor

import (
	"io"
	"sync"
	"time"

	"silo/internal/log"

	"github.com/anacrolix/torrent"
)

type Reader struct {
	torrent.Reader
	offset    int64
	readahead int64
	next      int64
	zone      int64
	file      *torrent.File

	cache    *Cache
	isClosed bool

	lastAccess int64
	isUse      bool
	mu         sync.Mutex
}

func newReader(file *torrent.File, cache *Cache) *Reader {
	r := new(Reader)
	r.file = file
	r.Reader = file.NewReader()

	r.cache = cache
	r.isUse = true
	r.SetResponsive()

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
		r.offset += int64(n)
		r.lastAccess = time.Now().Unix()
	} else {
		log.Debug("[TorrStor] Torrent closed while reading")
	}
	return
}

// SetZones задает зоны загрузки ридера: ближнюю, дальнюю и полную.
func (r *Reader) SetZones(next, readahead, zone int64) {
	if limit := r.cache.capacity; zone > limit {
		zone = limit
	}
	if readahead > zone {
		readahead = zone
	}
	if next > readahead {
		next = readahead
	}

	r.mu.Lock()
	r.next, r.readahead, r.zone = next, readahead, zone
	on := r.isUse
	r.mu.Unlock()

	if !on {
		return
	}
	r.applyZones(next, readahead, zone)
}

func (r *Reader) applyZones(next, readahead, zone int64) {
	r.Reader.SetNext(next)
	r.Reader.SetReadahead(readahead)
	r.Reader.SetZone(zone)
}

// Zoned сообщает, заданы ли ридеру зоны загрузки.
func (r *Reader) Zoned() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.zone > 0
}

func (r *Reader) Offset() int64 {
	return r.offset
}

func (r *Reader) Readahead() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.readahead
}

func (r *Reader) Close() {
	r.isClosed = true
	if r.file != nil && r.file.Torrent() != nil && len(r.file.Torrent().Files()) > 0 {
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
	r.mu.Lock()
	readahead := r.readahead
	r.mu.Unlock()
	return r.getPieceNum(r.offset + readahead)
}

// readerEnd возвращает байт, до которого ридер хочет держать данные.
func (r *Reader) readerEnd() int64 {
	_, end := r.getOffsetRange()
	return end
}

func (r *Reader) getPieceNum(offset int64) int {
	return int((offset + r.file.Offset()) / r.cache.pieceLength)
}

// getOffsetRange возвращает границы зоны загрузки ридера в байтах файла.
func (r *Reader) getOffsetRange() (int64, int64) {
	r.mu.Lock()
	zone := r.zone
	r.mu.Unlock()

	begin := r.offset
	if begin < 0 {
		begin = 0
	}
	end := begin + zone
	if end > r.file.Length() {
		end = r.file.Length()
	}
	return begin, end
}

func (r *Reader) checkReader() {
	if time.Now().Unix() > r.lastAccess+60 && r.cache.Readers() > 1 {
		r.readerOff()
	} else {
		r.readerOn()
	}
}

// readerOn возвращает ридеру его зоны загрузки после парковки.
func (r *Reader) readerOn() {
	r.mu.Lock()
	if r.isUse {
		r.mu.Unlock()
		return
	}
	if pos, err := r.Reader.Seek(0, io.SeekCurrent); err == nil && pos == 0 {
		r.Reader.Seek(r.offset, io.SeekStart)
	}
	r.isUse = true
	next, readahead, zone := r.next, r.readahead, r.zone
	r.mu.Unlock()

	r.applyZones(next, readahead, zone)
}

// readerOff снимает зоны ридера, чтобы движок перестал качать для него.
func (r *Reader) readerOff() {
	r.mu.Lock()
	if !r.isUse {
		r.mu.Unlock()
		return
	}
	r.isUse = false
	offset := r.offset
	r.mu.Unlock()

	r.applyZones(0, 0, 0)
	if offset > 0 {
		r.Reader.Seek(0, io.SeekStart)
	}
}

func (r *Reader) getUseReaders() int {
	return r.cache.GetUseReaders()
}
