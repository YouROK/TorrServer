package torrstor

import (
	"sync"
	"testing"

	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
)

func testInfo() *metainfo.Info {
	return &metainfo.Info{
		Name:        "test",
		PieceLength: 1 << 10,
		Length:      1 << 20,
		Pieces:      make([]byte, (1<<10)*20),
	}
}

func initTestSettings(t *testing.T) {
	if settings.BTsets == nil {
		settings.BTsets = &settings.BTSets{}
		t.Cleanup(func() { settings.BTsets = nil })
	}
}

// Storage.caches: anacrolix-driven Cache.Close (delete) racing
// OpenTorrent/GetCache from other goroutines.
func TestStorageCachesConcurrentOpenClose(t *testing.T) {
	initTestSettings(t)
	stor := NewStorage(1 << 20)

	var wg sync.WaitGroup
	const iters = 500

	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			var h metainfo.Hash
			h[0] = byte(i)
			stor.OpenTorrent(testInfo(), h)
			cache := stor.GetCache(h)
			if cache != nil {
				// anacrolix calls TorrentImpl.Close from its own goroutine
				cache.Close()
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			var h metainfo.Hash
			h[0] = byte(i)
			stor.GetCache(h)
			var h2 metainfo.Hash
			h2[0] = byte(i)
			h2[1] = 1
			stor.OpenTorrent(testInfo(), h2)
			stor.CloseHash(h2)
		}
	}()
	wg.Wait()
}

// Cache.pieces: GetState/cleanPieces iterating while Close nils the map.
func TestCachePiecesConcurrentStateClose(t *testing.T) {
	initTestSettings(t)

	var wg sync.WaitGroup
	const iters = 300

	for i := 0; i < iters; i++ {
		stor := NewStorage(1 << 20)
		var h metainfo.Hash
		h[0] = byte(i)
		stor.OpenTorrent(testInfo(), h)
		cache := stor.GetCache(h)

		wg.Add(2)
		go func() {
			defer wg.Done()
			cache.GetState()
			cache.cleanPieces()
		}()
		go func() {
			defer wg.Done()
			cache.Close()
		}()
		wg.Wait()
	}
}
