package torrstor

import (
	"context"
	"io"
	"testing"
)

// fakeTorrentReader stands in for the anacrolix reader: it only tracks the position the
// torrstor reader moves it to.
type fakeTorrentReader struct{ pos int64 }

func (f *fakeTorrentReader) Read([]byte) (int, error)                         { return 0, io.EOF }
func (f *fakeTorrentReader) ReadContext(context.Context, []byte) (int, error) { return 0, io.EOF }
func (f *fakeTorrentReader) Close() error                                     { return nil }
func (f *fakeTorrentReader) SetReadahead(int64)                               {}
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
	backend := &fakeTorrentReader{pos: at}
	r := &Reader{Reader: backend, isUse: true, offset: at}

	r.beginIO()
	r.readerOff()
	if !r.isUse || backend.pos != at {
		t.Fatalf("parked with a read in flight: isUse=%v position=%d, want true and %d", r.isUse, backend.pos, at)
	}
	r.endIO()

	r.readerOff()
	if r.isUse || backend.pos != 0 {
		t.Fatalf("an idle reader was not parked: isUse=%v position=%d", r.isUse, backend.pos)
	}

	r.beginIO()
	if !r.isUse || backend.pos != at {
		t.Fatalf("the next read did not resume at the reader's position: isUse=%v position=%d, want %d", r.isUse, backend.pos, at)
	}
	r.endIO()
}
