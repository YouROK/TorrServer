package torr

import (
	"fmt"
	"io"

	"github.com/anacrolix/torrent"
	"server/log"
)

type Reader struct {
	torrent.Reader
	offset    int64
	readahead int64
	file      *torrent.File
	torr      *Torrent

	isClosed bool

	///Preload
	isPreload         bool
	endOffsetPreload  int64
	currOffsetPreload int64
}

func NewReader(torr *Torrent, file *torrent.File, readahead int64) *Reader {
	r := new(Reader)
	r.file = file
	r.Reader = file.NewReader()

	if readahead <= 0 {
		readahead = torr.Torrent.Info().PieceLength
	}
	r.SetReadahead(readahead)
	torr.GetCache().AddReader(r)
	r.torr = torr
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
	return
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.offset += int64(n)
	r.preload()
	return
}

func (r *Reader) SetReadahead(length int64) {
	r.Reader.SetReadahead(length)
	r.readahead = length
}

func (r *Reader) Offset() int64 {
	return r.offset
}

func (r *Reader) Readahead() int64 {
	return r.readahead
}

func (r *Reader) Close() error {
	r.isClosed = true
	return r.Reader.Close()
}

func (r *Reader) preload() {
	r.currOffsetPreload = r.offset
	capacity := r.torr.cache.GetCapacity()
	plength := r.torr.Info().PieceLength
	r.endOffsetPreload = r.offset + capacity - r.readahead - plength*3
	if r.endOffsetPreload > r.file.Length() {
		r.endOffsetPreload = r.file.Length()
	}
	if r.endOffsetPreload < r.readahead || r.isPreload {
		return
	}
	r.isPreload = true
	//TODO remove logs
	fmt.Println("Start buffering...")
	go func() {
		buffReader := r.file.NewReader()
		defer func() {
			r.isPreload = false
			buffReader.Close()
			fmt.Println("End buffering...")
		}()
		buffReader.SetReadahead(0)
		buffReader.Seek(r.currOffsetPreload, io.SeekStart)
		buff5mb := make([]byte, 1024)
		for r.currOffsetPreload < r.endOffsetPreload && !r.isClosed {
			off, err := buffReader.Read(buff5mb)
			if err != nil {
				log.TLogln("Error read e head buffer", err)
				return
			}
			r.currOffsetPreload += int64(off)
		}
	}()
}
