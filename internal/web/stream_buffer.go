package web

import (
	"bufio"
	"io"
)

type bufferedStreamReader struct {
	source io.ReadSeeker
	buffer *bufio.Reader
}

func newBufferedStreamReader(source io.ReadSeeker, size int) *bufferedStreamReader {
	return &bufferedStreamReader{source: source, buffer: bufio.NewReaderSize(source, size)}
}

func (r *bufferedStreamReader) Read(p []byte) (int, error) { return r.buffer.Read(p) }
func (r *bufferedStreamReader) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekCurrent {
		// Часть байтов уже прочитана в буфер, поэтому смещение уменьшается.
		offset -= int64(r.buffer.Buffered())
	}
	pos, err := r.source.Seek(offset, whence)
	if err == nil {
		r.buffer.Reset(r.source)
	}
	return pos, err
}
