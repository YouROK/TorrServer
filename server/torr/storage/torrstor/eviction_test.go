package torrstor

import "testing"

// A piece that is still downloading is offered for eviction only after every complete one,
// even when it is older. The torrent client keeps counting the chunks it has already
// received as present and never requests them again, so once their bytes are dropped a
// responsive reader that comes back to that spot is served zeros.
func TestGetRemPiecesEvictsDownloadingPiecesLast(t *testing.T) {
	c := &Cache{
		pieces:  make(map[int]*Piece),
		readers: make(map[*Reader]struct{}),
	}
	// Pieces 0 and 1 are the oldest and only partly downloaded.
	for id := 0; id < 6; id++ {
		c.pieces[id] = &Piece{Id: id, Size: 1 << 10, Complete: id >= 2, Accessed: int64(id), cache: c}
	}

	order := c.getRemPieces()

	if len(order) != len(c.pieces) {
		t.Fatalf("got %d eviction candidates, want %d", len(order), len(c.pieces))
	}
	complete := len(order) - 2
	for i, p := range order {
		if (i < complete) != p.Complete {
			t.Fatalf("eviction order %v: piece %d (complete=%v) at position %d", ids(order), p.Id, p.Complete, i)
		}
		if i > 0 && order[i-1].Complete == p.Complete && order[i-1].Accessed > p.Accessed {
			t.Fatalf("eviction order %v is not oldest first within a group", ids(order))
		}
	}
}

func ids(pieces []*Piece) []int {
	out := make([]int, len(pieces))
	for i, p := range pieces {
		out[i] = p.Id
	}
	return out
}
