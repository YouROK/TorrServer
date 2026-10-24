package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
)

func TestRemoveAllFiles(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "a"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	removeAllFiles(sub)
	if _, err := os.Stat(sub); !os.IsNotExist(err) {
		t.Fatal("directory still exists")
	}
	removeAllFiles(filepath.Join(dir, "missing"))
}

func TestCleanCache(t *testing.T) {
	old := settings.BTsets
	t.Cleanup(func() { settings.BTsets = old })

	settings.BTsets = &settings.BTSets{}
	cleanCache()
	settings.BTsets.UseDisk = true
	settings.BTsets.TorrentsSavePath = "/"
	cleanCache()
	settings.BTsets.TorrentsSavePath = ""
	cleanCache()

	dir := t.TempDir()
	hash := strings.Repeat("ab", 20)
	if err := os.Mkdir(filepath.Join(dir, hash), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, hash, "piece"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("k"), 0o644); err != nil {
		t.Fatal(err)
	}
	settings.BTsets = &settings.BTSets{UseDisk: true, TorrentsSavePath: dir, RemoveCacheOnDrop: true}
	cleanCache()
	if _, err := os.Stat(filepath.Join(dir, hash)); !os.IsNotExist(err) {
		t.Fatal("hash cache dir still exists")
	}
	if _, err := os.Stat(filepath.Join(dir, "keep.txt")); err != nil {
		t.Fatal(err)
	}

	kept := strings.Repeat("cd", 20)
	if err := os.Mkdir(filepath.Join(dir, kept), 0o755); err != nil {
		t.Fatal(err)
	}
	settings.BTsets.RemoveCacheOnDrop = false
	settings.BTsets.TorrentsSavePath = filepath.Join(dir, "missing")
	cleanCache()
	settings.BTsets.TorrentsSavePath = dir
	cleanCache()
	if _, err := os.Stat(filepath.Join(dir, kept)); !os.IsNotExist(err) {
		t.Fatal("unused hash dir kept")
	}

	oldPath, oldRO, oldWA := settings.Path, settings.ReadOnly, settings.SearchWA
	settings.Path = t.TempDir()
	t.Cleanup(func() {
		settings.CloseDB()
		settings.Path, settings.ReadOnly, settings.SearchWA = oldPath, oldRO, oldWA
	})
	if err := settings.InitSets(false, false); err != nil {
		t.Fatal(err)
	}
	keepHash := strings.Repeat("ef", 20)
	cacheDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(cacheDir, keepHash), 0o755); err != nil {
		t.Fatal(err)
	}
	settings.AddTorrent(&settings.TorrentDB{
		TorrentSpec: &torrent.TorrentSpec{InfoHash: metainfo.NewHashFromHex(keepHash), DisplayName: "Keep"},
		Title:       "Keep",
	})
	settings.BTsets.UseDisk = true
	settings.BTsets.RemoveCacheOnDrop = false
	settings.BTsets.TorrentsSavePath = cacheDir
	cleanCache()
	if _, err := os.Stat(filepath.Join(cacheDir, keepHash)); err != nil {
		t.Fatal("in-db cache was removed")
	}
	if err := os.WriteFile(filepath.Join(cacheDir, strings.Repeat("12", 20)), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	cleanCache()
}
