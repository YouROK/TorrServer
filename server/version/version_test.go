package version

import (
	"testing"

	_ "github.com/anacrolix/torrent"
)

func TestGetTorrentVersion(t *testing.T) {
	got := GetTorrentVersion()
	if got == "" {
		t.Fatal("expected the replaced anacrolix/torrent module version")
	}
}
