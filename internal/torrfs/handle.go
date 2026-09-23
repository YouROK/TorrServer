package torrfs

import (
	"io"
	"sync/atomic"
	"time"
)

// Handle - открытый файл торрента. Реализует io.ReadSeekCloser и передаётся в res.stream.
type Handle struct {
	reader  io.ReadSeekCloser
	name    string
	modTime time.Time
	closed  atomic.Bool
}

func NewHandle(r io.ReadSeekCloser, name string, modTime time.Time) *Handle {
	return &Handle{
		reader:  r,
		name:    name,
		modTime: modTime,
	}
}

func (h *Handle) Read(p []byte) (int, error) {
	return h.reader.Read(p)
}

func (h *Handle) Seek(offset int64, whence int) (int64, error) {
	return h.reader.Seek(offset, whence)
}

// Close идемпотентен - повторный вызов ничего не делает.
func (h *Handle) Close() error {
	if h.closed.Swap(true) {
		return nil
	}
	return h.reader.Close()
}

func (h *Handle) Name() string       { return h.name }
func (h *Handle) ModTime() time.Time { return h.modTime }
