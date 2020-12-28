package torrstor

import (
	sync "github.com/sasha-s/go-deadlock"
	"io"

	"github.com/anacrolix/torrent"
	"server/log"
)

type Reader struct {
	torrent.Reader
	offset    int64
	readahead int64
	file      *torrent.File

	cache    *Cache
	isClosed bool
	mu       sync.Mutex

	///Preload
	muPreload sync.Mutex
}

func newReader(file *torrent.File, cache *Cache) *Reader {
	r := new(Reader)
	r.file = file
	r.Reader = file.NewReader()

	r.SetReadahead(0)
	r.cache = cache

	cache.muReaders.Lock()
	cache.readers[r] = struct{}{}
	cache.muReaders.Unlock()
	return r
}

func (r *Reader) Seek(offset int64, whence int) (n int64, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	n, err = r.Reader.Seek(offset, whence)
	r.offset = n
	return
}

func (r *Reader) Read(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	err = io.EOF
	if r.isClosed {
		return
	}
	if r.file.Torrent() != nil && r.file.Torrent().Info() != nil {
		n, err = r.Reader.Read(p)
		r.offset += int64(n)
		go r.preload()
	} else {
		log.TLogln("Torrent closed and readed")
	}
	return
}

func (r *Reader) SetReadahead(length int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Reader.SetReadahead(length)
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
	// TODO провверить как будут закрываться ридеры
	r.mu.Lock()
	r.isClosed = true
	if len(r.file.Torrent().Files()) > 0 {
		r.Reader.Close()
	}
	r.mu.Unlock()
	go r.cache.getRemPieces()
}

func (c *Cache) NewReader(file *torrent.File) *Reader {
	return newReader(file, c)
}

func (c *Cache) Readers() int {
	if c == nil {
		return 0
	}
	c.muReaders.Lock()
	defer c.muReaders.Unlock()
	if c == nil || c.readers == nil {
		return 0
	}
	return len(c.readers)
}

func (c *Cache) CloseReader(r *Reader) {
	r.cache.muReaders.Lock()
	r.Close()
	delete(r.cache.readers, r)
	r.cache.muReaders.Unlock()
}
