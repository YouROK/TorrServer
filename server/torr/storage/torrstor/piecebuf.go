package torrstor

import "sync"

// piecePools recycles in-memory piece buffers by exact length so cache
// turnover does not allocate a fresh backing array every eviction.
// ponytail: one global pool per length. Split per cache if a process hosts
// many unusual lengths and this map shows up in profiles.
var piecePools sync.Map // int -> *sync.Pool

// pieceSlab is the pooled object. Storing the slice inside a pointer keeps
// Get/Put off the alloc path (a bare []byte in sync.Pool boxes a header).
type pieceSlab struct {
	b []byte
}

func piecePool(n int) *sync.Pool {
	if v, ok := piecePools.Load(n); ok {
		return v.(*sync.Pool)
	}
	p := &sync.Pool{New: func() any {
		return &pieceSlab{b: make([]byte, n)}
	}}
	actual, _ := piecePools.LoadOrStore(n, p)
	return actual.(*sync.Pool)
}

func getPieceBuf(n int) ([]byte, *pieceSlab) {
	s := piecePool(n).Get().(*pieceSlab)
	if cap(s.b) < n {
		s.b = make([]byte, n)
	}
	// Put clears the full cap before the slab is published again.
	return s.b[:n], s
}

func putPieceBuf(s *pieceSlab) {
	if s == nil {
		return
	}
	n := cap(s.b)
	if n == 0 {
		return
	}
	s.b = s.b[:n]
	clear(s.b)
	piecePool(n).Put(s)
}
