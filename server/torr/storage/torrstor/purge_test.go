package torrstor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"server/settings"
)

const testHash = "2e8e44068b254814ea1a7d4969a9af1d78e0f51f"

func setupPurgeDirs(t *testing.T) (cacheDir string) {
	t.Helper()
	oldPath := settings.Path
	settings.Path = t.TempDir()
	t.Cleanup(func() { settings.Path = oldPath })

	cacheDir = t.TempDir()
	if err := os.MkdirAll(filepath.Join(cacheDir, testHash), 0o777); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"0", "1", "info"} {
		if err := os.WriteFile(filepath.Join(cacheDir, testHash, name), []byte("x"), 0o666); err != nil {
			t.Fatal(err)
		}
	}
	// not a torrent cache directory: must survive
	if err := os.MkdirAll(filepath.Join(cacheDir, "keep-me"), 0o777); err != nil {
		t.Fatal(err)
	}
	return cacheDir
}

func TestPurgeDiskCacheRemovesTorrentDirsOnly(t *testing.T) {
	cacheDir := setupPurgeDirs(t)

	if err := PurgeDiskCache(cacheDir); err != nil {
		t.Fatalf("PurgeDiskCache: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cacheDir, testHash)); !os.IsNotExist(err) {
		t.Fatalf("torrent cache dir still exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cacheDir, "keep-me")); err != nil {
		t.Fatalf("unrelated dir was removed: %v", err)
	}
	if _, err := os.Stat(purgeMarkerPath()); !os.IsNotExist(err) {
		t.Fatalf("purge marker left after a finished purge: %v", err)
	}
}

func TestResumeDiskCachePurgeFinishesInterruptedPurge(t *testing.T) {
	cacheDir := setupPurgeDirs(t)

	// a purge that was cut short leaves its marker behind
	data, _ := json.Marshal(purgeMarker{Path: cacheDir})
	if err := os.WriteFile(purgeMarkerPath(), data, 0o666); err != nil {
		t.Fatal(err)
	}

	ResumeDiskCachePurge()

	if _, err := os.Stat(filepath.Join(cacheDir, testHash)); !os.IsNotExist(err) {
		t.Fatalf("interrupted purge was not finished: %v", err)
	}
	if _, err := os.Stat(purgeMarkerPath()); !os.IsNotExist(err) {
		t.Fatalf("purge marker left after resume: %v", err)
	}
}

func TestResumeDiskCachePurgeWithoutMarker(t *testing.T) {
	cacheDir := setupPurgeDirs(t)

	ResumeDiskCachePurge()

	if _, err := os.Stat(filepath.Join(cacheDir, testHash)); err != nil {
		t.Fatalf("cache removed without a pending purge: %v", err)
	}
}
