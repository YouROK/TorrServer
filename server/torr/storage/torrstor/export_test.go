package torrstor

import "io"

// Test hooks: re-export internal fields/methods so torrstor_test
// (external test package) can drive whitebox scenarios without forcing
// us to move every test inside the package.

func (c *Cache) CleanPiecesForTest() { c.cleanPieces() }

func (c *Cache) PieceSizeForTest(id int) int64 {
	if p, ok := c.pieces[id]; ok {
		return p.Size.Load()
	}
	return -1
}

func (c *Cache) PieceCountForTest() int    { return c.pieceCount }
func (c *Cache) PieceLengthForTest() int64 { return c.pieceLength }

// SeekReaderForTest moves r.offset directly so tests can position a
// reader without needing an underlying file with real data. Touches
// only the mirrored offset that getPiecesRange uses; does NOT call the
// underlying anacrolix Reader.Seek (which would require data).
func SeekReaderForTest(r *Reader, off int64) {
	r.offset.Store(off)
}

// MarkReaderUseForTest forces r.isUse so getPiecesRange contributions
// are included in cleanPieces decisions.
func MarkReaderUseForTest(r *Reader, use bool) {
	r.isUse = use
}

// ForceReaderOffForTest invokes the otherwise time-gated readerOff
// path directly. In production it fires after 60s of inactivity AND
// when there is more than one reader on the cache.
func ForceReaderOffForTest(r *Reader) {
	r.readerOff()
}

// ReaderOffsetForTest exposes r.offset (the mirrored logical position)
// so tests can verify readerOff/readerOn don't lose it.
func ReaderOffsetForTest(r *Reader) int64 { return r.offset.Load() }

// UnderlyingReaderPosForTest returns the position of the embedded
// anacrolix torrent.Reader by doing a SeekCurrent(0). This is the
// "real" position the next direct Read on the embedded reader would
// use — distinct from r.offset, which torrstor.Reader maintains
// itself.
func UnderlyingReaderPosForTest(r *Reader) int64 {
	pos, err := r.Reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return -1
	}
	return pos
}

// WaitClearPriorityForTest blocks until any in-flight clearPriority
// goroutine has released c.muPrio. Used by regression-gates that
// previously had to sleep ~1s to wait out the (now removed) artificial
// delay in clearPriority.
func (c *Cache) WaitClearPriorityForTest() {
	c.muPrio.Lock()
	c.muPrio.Unlock() //nolint:staticcheck // intentional acquire-release as a barrier
}

// FillPieceForTest pretends a piece was downloaded by going through
// the real WriteAt path (which exercises the same code production hits
// when anacrolix flushes a piece into storage). Returns the bytes
// actually written.
func (c *Cache) FillPieceForTest(id int, b []byte) (int, error) {
	p, ok := c.pieces[id]
	if !ok {
		return 0, io.ErrShortBuffer
	}
	return p.WriteAt(b, 0)
}

