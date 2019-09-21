package torr

import (
	"io"

	"github.com/anacrolix/torrent"
	"server/torr/storage/memcacheV2"
)

type Reader struct {
	torrent.Reader

	offset int64

	file  *torrent.File
	cache *memcacheV2.Cache
}

func NewReader(file *torrent.File, cache *memcacheV2.Cache) *Reader {
	r := new(Reader)
	r.file = file
	r.Reader = file.NewReader()
	r.Reader.SetReadahead(0)
	r.cache = cache
	return r
}

func (r *Reader) Seek(offset int64, whence int) (n int64, err error) {
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
	r.cache.SetPos(int(n))
	return
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.offset += int64(n)
	r.cache.SetPos(int(n))
	return
}
