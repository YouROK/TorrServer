package torr

import (
	"bufio"
	"io"
	"sync"
)

// streamBufSize is the /stream read-ahead. Only this size is pooled: every
// player seek is a new request, and a fresh 1MiB per request was the largest
// TorrServer-owned allocator in a rapid-seek heap profile.
const streamBufSize = 1 << 20

var streamBufPool = sync.Pool{New: func() any { return bufio.NewReaderSize(nil, streamBufSize) }}

type bufferedStreamReader struct {
	source io.ReadSeeker
	buffer *bufio.Reader
	pooled bool
}

func newBufferedStreamReader(source io.ReadSeeker, size int) *bufferedStreamReader {
	if size != streamBufSize {
		return &bufferedStreamReader{source: source, buffer: bufio.NewReaderSize(source, size)}
	}
	b := streamBufPool.Get().(*bufio.Reader)
	b.Reset(source)
	return &bufferedStreamReader{source: source, buffer: b, pooled: true}
}

// release returns a pooled buffer. The reader must not be used afterwards.
func (r *bufferedStreamReader) release() {
	if !r.pooled {
		return
	}
	r.pooled = false
	// Drop the source so the pool does not pin the torrent reader.
	r.buffer.Reset(nil)
	streamBufPool.Put(r.buffer)
	r.buffer = nil
}

func (r *bufferedStreamReader) Read(p []byte) (int, error) { return r.buffer.Read(p) }

// WriteTo lets io.Copy drain the read-ahead buffer without a second scratch slice.
// ServeContent uses CopyN, which does not see WriterTo; the writer-side ReadFrom
// on ServeStreamContent covers that path.
func (r *bufferedStreamReader) WriteTo(w io.Writer) (int64, error) {
	return r.buffer.WriteTo(w)
}
func (r *bufferedStreamReader) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekCurrent {
		// The source has advanced past unread buffered bytes.
		offset -= int64(r.buffer.Buffered())
	}
	pos, err := r.source.Seek(offset, whence)
	if err == nil {
		r.buffer.Reset(r.source)
	}
	return pos, err
}
