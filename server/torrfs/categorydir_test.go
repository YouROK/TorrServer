package torrfs

import (
	"testing"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/torr"
)

func torrentWithTitle(title string, b byte) *torr.Torrent {
	var h metainfo.Hash
	h[0] = b
	return &torr.Torrent{Title: title, TorrentSpec: &torrent.TorrentSpec{InfoHash: h}}
}

func TestTorrentNamesTellDuplicateTitlesApart(t *testing.T) {
	torrs := []*torr.Torrent{
		torrentWithTitle("Same", 0x11),
		torrentWithTitle("Other", 0x22),
		torrentWithTitle("Same", 0x33),
		torrentWithTitle("a/b", 0x44),
		torrentWithTitle("a_b", 0x55),
	}
	want := []string{"Same [11000000]", "Other", "Same [33000000]", "a_b [44000000]", "a_b [55000000]"}
	got := torrentNames(torrs)
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("name %d = %q, want %q", i, got[i], want[i])
		}
	}
}
