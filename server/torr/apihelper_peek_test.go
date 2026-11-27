package torr

import (
	"testing"
	"time"

	"github.com/anacrolix/torrent/metainfo"
)

func TestPeekTorrentDoesNotExtendExpiration(t *testing.T) {
	oldBTS := bts
	t.Cleanup(func() {
		bts = oldBTS
	})

	var hash metainfo.Hash
	hash[0] = 0xab

	deadline := time.Now().Add(-time.Minute)
	tr := &Torrent{
		expiredTime: deadline,
	}

	bts = &BTServer{
		torrents: map[metainfo.Hash]*Torrent{
			hash: tr,
		},
	}

	for i := 0; i < 10; i++ {
		if got := PeekTorrent(hash.HexString()); got != tr {
			t.Fatal("PeekTorrent did not return the registered torrent")
		}
	}

	if !tr.expiredTime.Equal(deadline) {
		t.Fatalf("expiration changed from %v to %v",
			deadline, tr.expiredTime)
	}

	var missing metainfo.Hash
	missing[0] = 0xcd

	if got := PeekTorrent(missing.HexString()); got != nil {
		t.Fatal("PeekTorrent returned an unregistered torrent")
	}

	if got := PeekTorrent("zz"); got != nil {
		t.Fatal("PeekTorrent accepted a malformed hash")
	}
}
