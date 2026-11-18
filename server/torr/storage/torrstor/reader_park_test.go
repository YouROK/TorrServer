package torrstor

import (
	"context"
	"io"
	"testing"
)

// fakeTorrentReader stands in for the anacrolix reader: it only tracks the position the
// torrstor reader moves it to.
type fakeTorrentReader struct {
	pos       int64
	readahead int64
}

func (f *fakeTorrentReader) Read([]byte) (int, error)                         { return 0, io.EOF }
func (f *fakeTorrentReader) ReadContext(context.Context, []byte) (int, error) { return 0, io.EOF }
func (f *fakeTorrentReader) Close() error                                     { return nil }
func (f *fakeTorrentReader) SetReadahead(n int64)                             { f.readahead = n }
func (f *fakeTorrentReader) SetResponsive()                                   {}

func (f *fakeTorrentReader) Seek(off int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.pos = off
	case io.SeekCurrent:
		f.pos += off
	}
	return f.pos, nil
}

// The idle sweep parks a reader by seeking its anacrolix reader to byte 0. Under a read in
// progress that moves the read itself: anacrolix adds the bytes it returns to the new
// position, and the stream carries on from the head of the file.
func TestReaderOffDoesNotParkAReaderWithIOInFlight(t *testing.T) {
	const at = 900 << 20
	const ahead = 16 << 20
	backend := &fakeTorrentReader{pos: at, readahead: ahead}
	cache := NewCache(32<<20, nil)
	r := &Reader{backend: backend, isUse: true, offset: at, readahead: ahead, cache: cache}
	cache.readers[r] = struct{}{}
	cache.readers[&Reader{}] = struct{}{}

	r.beginIO()
	r.checkReader()
	if !r.isUse || backend.pos != at {
		t.Fatalf("parked with a read in flight: isUse=%v position=%d, want true and %d", r.isUse, backend.pos, at)
	}
	r.endIO()

	r.checkReader()
	if r.isUse || backend.pos != 0 || backend.readahead != 0 {
		t.Fatalf("idle reader was not parked: isUse=%v position=%d readahead=%d", r.isUse, backend.pos, backend.readahead)
	}

	const delta = 4096
	pos, err := r.Seek(delta, io.SeekCurrent)
	if err != nil || pos != at+delta || !r.isUse || backend.pos != pos {
		t.Fatalf("relative seek did not resume at the saved offset: position=%d, backend=%d, active=%v, err=%v", pos, backend.pos, r.isUse, err)
	}
	if backend.readahead != ahead {
		t.Fatalf("resumed readahead = %d, want %d", backend.readahead, ahead)
	}
}
