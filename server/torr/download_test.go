package torr

import (
	"testing"

	"server/torr/storage/torrstor"
)

func TestFileRange(t *testing.T) {
	tests := []struct {
		name                        string
		pieceLength, offset, length int64
		want                        torrstor.PieceRange
	}{
		{"aligned", 100, 0, 300, torrstor.PieceRange{Begin: 0, End: 3}},
		{"straddles pieces", 100, 150, 100, torrstor.PieceRange{Begin: 1, End: 3}},
		{"inside one piece", 100, 120, 10, torrstor.PieceRange{Begin: 1, End: 2}},
		{"empty file", 100, 120, 0, torrstor.PieceRange{}},
	}
	for _, tt := range tests {
		if got := fileRange(tt.pieceLength, tt.offset, tt.length); got != tt.want {
			t.Errorf("%s: fileRange = %+v, want %+v", tt.name, got, tt.want)
		}
	}
}

func TestFileCompletedBytes(t *testing.T) {
	complete := map[int]bool{1: true, 2: true}
	isComplete := func(id int) bool { return complete[id] }

	// file covers bytes [150, 350): half of piece 1, all of piece 2, half of piece 3
	if got := fileCompletedBytes(100, 150, 200, isComplete); got != 150 {
		t.Errorf("fileCompletedBytes = %d, want 150", got)
	}
	if got := fileCompletedBytes(100, 400, 100, isComplete); got != 0 {
		t.Errorf("fileCompletedBytes for an uncached file = %d, want 0", got)
	}
}

func TestSeedableShare(t *testing.T) {
	const big = seedMinFileSize * 4
	tests := []struct {
		name              string
		length, completed int64
		want              bool
	}{
		{"half cached", big, big / 2, true},
		{"less than half", big, big/2 - 1, false},
		{"small file fully cached", seedMinFileSize - 1, seedMinFileSize - 1, false},
	}
	for _, tt := range tests {
		if got := seedableShare(tt.length, tt.completed); got != tt.want {
			t.Errorf("%s: seedableShare = %v, want %v", tt.name, got, tt.want)
		}
	}
}
