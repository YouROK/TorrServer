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
	next      int64
	readahead int64
	zone      int64
}

func (f *fakeTorrentReader) Read([]byte) (int, error)                         { return 0, io.EOF }
func (f *fakeTorrentReader) ReadContext(context.Context, []byte) (int, error) { return 0, io.EOF }
func (f *fakeTorrentReader) Close() error                                     { return nil }
func (f *fakeTorrentReader) SetReadahead(n int64)                             { f.readahead = n }
func (f *fakeTorrentReader) SetNext(n int64)                                  { f.next = n }
func (f *fakeTorrentReader) SetZone(n int64)                                  { f.zone = n }
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

	r.next, r.readahead, r.zone = 4<<20, 16<<20, 32<<20
	r.readerOff()
	if r.isUse || backend.pos != 0 {
		t.Fatalf("an idle reader was not parked: isUse=%v position=%d", r.isUse, backend.pos)
	}
	if backend.next != 0 || backend.readahead != 0 || backend.zone != 0 {
		t.Fatalf("parked reader kept a download window: next %d readahead %d zone %d", backend.next, backend.readahead, backend.zone)
	}
	if r.next != 4<<20 || r.readahead != 16<<20 || r.zone != 32<<20 {
		t.Fatal("parking forgot the windows resume needs")
	}

	r.beginIO()
	if !r.isUse || backend.pos != at {
		t.Fatalf("the next read did not resume at the reader's position: isUse=%v position=%d, want %d", r.isUse, backend.pos, at)
	}
	if backend.next != 4<<20 || backend.readahead != 16<<20 || backend.zone != 32<<20 {
		t.Fatalf("resume windows: next %d readahead %d zone %d", backend.next, backend.readahead, backend.zone)
	}
	r.endIO()
}

// getRemPieces must not park: cleanPieces holds muRemove and park takes the client lock.
func TestGetRemPiecesDoesNotParkReaders(t *testing.T) {
	backend := &fakeTorrentReader{pos: 1 << 20, next: 4 << 20, readahead: 16 << 20, zone: 32 << 20}
	c := &Cache{
		pieces:       make(map[int]*Piece),
		activePieces: make(map[int]struct{}),
		readers:      make(map[*Reader]struct{}),
		pieceLength:  1024,
		pieceCount:   4,
		capacity:     1 << 30,
	}
	r := &Reader{
		Reader:      backend,
		cache:       c,
		isUse:       true,
		offset:      1 << 20,
		overrideLen: 4 << 20,
		lastAccess:  1, // long idle vs minReaderIdleTimeout
		next:        4 << 20,
		readahead:   16 << 20,
		zone:        32 << 20,
	}
	c.readers[r] = struct{}{}

	_ = c.getRemPieces()
	if !r.using() {
		t.Fatal("getRemPieces parked an idle-expired reader")
	}
	if backend.pos != 1<<20 || backend.next != 4<<20 {
		t.Fatalf("getRemPieces touched anacrolix windows: pos=%d next=%d", backend.pos, backend.next)
	}
}

func TestParkIdleReadersParks(t *testing.T) {
	backend := &fakeTorrentReader{pos: 1 << 20, next: 4 << 20, readahead: 16 << 20, zone: 32 << 20}
	c := &Cache{
		readers: make(map[*Reader]struct{}),
	}
	r := &Reader{
		Reader:      backend,
		cache:       c,
		isUse:       true,
		offset:      1 << 20,
		overrideLen: 4 << 20,
		lastAccess:  1,
		next:        4 << 20,
		readahead:   16 << 20,
		zone:        32 << 20,
	}
	c.readers[r] = struct{}{}

	c.parkIdleReaders()
	if r.using() {
		t.Fatal("parkIdleReaders left idle reader active")
	}
	if backend.pos != 0 || backend.next != 0 || backend.readahead != 0 || backend.zone != 0 {
		t.Fatalf("park windows: pos=%d next=%d rah=%d zone=%d", backend.pos, backend.next, backend.readahead, backend.zone)
	}
}
