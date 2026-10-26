package torr

import (
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

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

// NewTorrent must not add a torrent that startLoadFromDB is still adding: both
// would work on the same DB spec at once.
func TestNewTorrentWaitsForDBLoad(t *testing.T) {
	bt := NewBTS()
	bt.client = nil
	spec := &torrent.TorrentSpec{InfoHash: metainfo.NewHashFromHex("2e8e44068b254814ea1a7d4969a9af1d78e0f51f")}
	added := make(chan struct{})
	bt.loading[spec.InfoHash] = added

	done := make(chan struct{})
	go func() {
		_, _ = NewTorrent(spec, bt)
		close(done)
	}()
	select {
	case <-done:
		t.Fatal("NewTorrent returned while the background load was in progress")
	case <-time.After(100 * time.Millisecond):
	}
	close(added)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("NewTorrent did not return after the background load finished")
	}
}

func TestFileStatsFromInfoMatchesStatusOrder(t *testing.T) {
	info := &metainfo.Info{
		Name:        "Show",
		PieceLength: 1,
		Files: []metainfo.FileInfo{
			{Path: []string{"s01", "e10.mkv"}, Length: 10},
			{Path: []string{"s01", "e02.mkv"}, Length: 2},
			{Path: []string{"cover.jpg"}, Length: 1},
		},
	}
	files := fileStatsFromInfo(info)
	if len(files) != 3 {
		t.Fatalf("got %d files, want 3", len(files))
	}
	want := []string{"Show/cover.jpg", "Show/s01/e02.mkv", "Show/s01/e10.mkv"}
	for i, f := range files {
		if f.Path != want[i] || f.Id != i+1 {
			t.Errorf("file %d = %q id %d, want %q id %d", i, f.Path, f.Id, want[i], i+1)
		}
	}

	single := &metainfo.Info{Name: "movie.mkv", Length: 5, PieceLength: 1}
	files = fileStatsFromInfo(single)
	if len(files) != 1 || files[0].Path != "movie.mkv" || files[0].Length != 5 {
		t.Fatalf("single file torrent: %+v", files[0])
	}
}
