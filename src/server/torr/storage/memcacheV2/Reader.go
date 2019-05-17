package memcacheV2

import "github.com/anacrolix/torrent"

type Reader struct {
	torrent.Reader

	pos int64
}

func NewReader(file torrent.File) *Reader {
	r := new(Reader)
	r.Reader = file.NewReader()
	return r
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.Read(p)
	r.pos += int64(n)
	return
}

func (r *Reader) Seek(offset int64, whence int) (ret int64, err error) {
	ret, err = r.Reader.Seek(offset, whence)
	r.pos = ret
	return
}
