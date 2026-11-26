package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
)

// The cache path is often the root of a USB drive, next to the user's own folders.
// Only cache folders of torrents that are no longer saved may be emptied on start.
func TestCleanCacheRemovesOnlyStaleHashFolders(t *testing.T) {
	settings.Path = t.TempDir()
	if err := settings.InitSets(false, false); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(settings.CloseDB)

	cacheDir := t.TempDir()
	settings.BTsets.UseDisk = true
	settings.BTsets.TorrentsSavePath = cacheDir
	settings.BTsets.RemoveCacheOnDrop = false

	saved := metainfo.NewHashFromHex("0123456789abcdef0123456789abcdef01234567")
	settings.AddTorrent(&settings.TorrentDB{TorrentSpec: &torrent.TorrentSpec{InfoHash: saved}})

	mk := func(dir string) string {
		path := filepath.Join(cacheDir, dir, "0")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	stale := mk("fedcba9876543210fedcba9876543210fedcba98")
	kept := mk(saved.HexString())
	user := mk("Interstellar.2014.2160p.UHD.BluRay.x265A") // 40 characters, not a hash

	cleanCache()

	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("cache of an unsaved torrent was not removed: %v", err)
	}
	if _, err := os.Stat(kept); err != nil {
		t.Errorf("cache of a saved torrent was removed: %v", err)
	}
	if _, err := os.Stat(user); err != nil {
		t.Errorf("a user folder with a 40-character name was emptied: %v", err)
	}
}
