package torrstor

import (
	"path/filepath"
	"testing"
)

func TestCachedTorrentsListsDirsWithMetainfo(t *testing.T) {
	cacheDir := setupDiskCacheSettings(t)
	const withInfo = "2e8e44068b254814ea1a7d4969a9af1d78e0f51f"
	const piecesOnly = "3f8e44068b254814ea1a7d4969a9af1d78e0f52a"

	writeCacheFile(t, filepath.Join(cacheDir, withInfo, metainfoFileName), 10)
	writeCacheFile(t, filepath.Join(cacheDir, withInfo, "0"), 10)
	writeCacheFile(t, filepath.Join(cacheDir, piecesOnly, "0"), 10)
	writeCacheFile(t, filepath.Join(cacheDir, "not-a-hash", metainfoFileName), 10)

	got := CachedTorrents()
	if len(got) != 1 || got[0] != withInfo {
		t.Fatalf("CachedTorrents = %v, want [%s]", got, withInfo)
	}
}

func TestCachedTorrentsEmptyDir(t *testing.T) {
	setupDiskCacheSettings(t)
	if got := CachedTorrents(); got != nil {
		t.Fatalf("CachedTorrents on an empty dir = %v, want nil", got)
	}
}
