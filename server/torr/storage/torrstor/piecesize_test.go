package torrstor

import "testing"

func TestPieceSize(t *testing.T) {
	c := &Cache{pieceLength: 100, totalLength: 250}
	for id, want := range []int64{100, 100, 50, 0} {
		if got := c.pieceSize(id); got != want {
			t.Errorf("pieceSize(%d) = %d, want %d", id, got, want)
		}
	}
	if got := (&Cache{pieceLength: 100}).pieceSize(5); got != 100 {
		t.Errorf("pieceSize without total length = %d, want the piece length", got)
	}
}
