package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
}
