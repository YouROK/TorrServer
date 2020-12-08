package torrstor

import (
	"io"

	"github.com/anacrolix/torrent"
)

type Reader struct {
	torrent.Reader
	offset    int64
	readahead int64
	file      *torrent.File

	cache    *Cache
	isClosed bool

	///Preload
	isPreload         bool
	endOffsetPreload  int64
	currOffsetPreload int64
}

func NewReader(file *torrent.File, cache *Cache) *Reader {
	r := new(Reader)
	r.file = file
	r.Reader = file.NewReader()

	r.SetReadAHead(0)
	r.cache = cache
	r.cache.readers[r] = struct{}{}
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
		r.offset = r.file.Length() - offset
	}
	n, err = r.Reader.Seek(offset, whence)
	r.offset = n
	return
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.isClosed {
		return 0, io.EOF
	}
	n, err = r.Reader.Read(p)
	r.offset += int64(n)
	go r.preload()
	return
}

func (r *Reader) SetReadAHead(length int64) {
	r.Reader.SetReadahead(length)
	r.readahead = length
}

func (r *Reader) Offset() int64 {
	return r.offset
}

func (r *Reader) ReadAHead() int64 {
	return r.readahead
}

func (r *Reader) Close() error {
	r.isClosed = true
	delete(r.cache.readers, r)
	return r.Reader.Close()
}

func (c *Cache) NewReader(file *torrent.File) *Reader {
	return NewReader(file, c)
}

func (c *Cache) Readers() int {
	if c == nil || c.readers == nil {
		return 0
	}
	return len(c.readers)
}
