package torr

import (
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

// TestRestartOfflineDownloadReleasesLockOnNoSpace: when the space check fails,
// restartOfflineDownload must return the error and leave the transition lock
// released exactly once. It used to call unlock() explicitly and then again
// through defer, which is a fatal "unlock of unlocked mutex".
func TestRestartOfflineDownloadReleasesLockOnNoSpace(t *testing.T) {
	infoBytes, err := bencode.Marshal(metainfo.Info{
		Name:        "huge",
		PieceLength: 16 << 10,
		Pieces:      make([]byte, 20),
		Length:      1 << 50, // does not fit on any test filesystem
	})
	if err != nil {
		t.Fatal(err)
	}
	hash := metainfo.HashBytes(infoBytes)
	db := &Torrent{TorrentSpec: &torrent.TorrentSpec{InfoHash: hash, InfoBytes: infoBytes}}
	db.setOfflineDir(t.TempDir())

	prev := bts
	bts = &BTServer{torrents: make(map[metainfo.Hash]*Torrent)}
	t.Cleanup(func() { bts = prev })

	if _, err := restartOfflineDownload(db, hash); err == nil {
		t.Fatal("restartOfflineDownload must fail when there is not enough disk space")
	}

	// the lock must be free again (and not unlocked twice)
	got := make(chan struct{})
	go func() {
		unlock := lockOfflineTransition(hash)
		unlock()
		close(got)
	}()
	select {
	case <-got:
	case <-time.After(2 * time.Second):
		t.Fatal("the offline transition lock was not released")
	}
}
