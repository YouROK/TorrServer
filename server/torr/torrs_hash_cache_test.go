package torr

import (
	"strings"
	"testing"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/torr/state"
)

func TestCachedTorrsHashReusesTokenUntilMetaChanges(t *testing.T) {
	hex := strings.Repeat("a", 40)
	h := metainfo.NewHashFromHex(hex)
	tor := &Torrent{
		Title:    "Show A",
		Poster:   "https://img/a.jpg",
		Category: "movie",
		Size:     10,
		Stat:     state.TorrentInDB,
		TorrentSpec: &torrent.TorrentSpec{
			InfoHash: h,
			Trackers: [][]string{{"http://t1/announce"}},
		},
	}

	first := tor.Status()
	if first.TorrsHash == "" {
		t.Fatal("expected packed token")
	}
	second := tor.Status()
	if second.TorrsHash != first.TorrsHash {
		t.Fatal("unchanged meta must reuse the packed token")
	}
	if tor.torrsHashKey == "" {
		t.Fatal("expected fingerprint stored")
	}

	tor.Title = "Show B"
	third := tor.Status()
	if third.TorrsHash == "" || third.TorrsHash == first.TorrsHash {
		t.Fatal("title change must repack")
	}
}
