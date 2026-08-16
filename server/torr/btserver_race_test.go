package torr

import (
	"sync"
	"testing"

	"github.com/anacrolix/torrent/metainfo"
)

// Mimics the real writers (NewTorrent insert, Torrent.Close delete), which
// hold bt.mu, racing against the public read methods used by HTTP handlers.
func TestBTServerTorrentsConcurrentAccess(t *testing.T) {
	bt := NewBTS()

	var wg sync.WaitGroup
	const iters = 5000

	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			var h metainfo.Hash
			h[0] = byte(i)
			h[1] = byte(i >> 8)
			bt.mu.Lock()
			bt.torrents[h] = &Torrent{}
			bt.mu.Unlock()
			bt.mu.Lock()
			delete(bt.torrents, h)
			bt.mu.Unlock()
		}
	}()
	go func() {
		defer wg.Done()
		var h metainfo.Hash
		for i := 0; i < iters; i++ {
			h[0] = byte(i)
			h[1] = byte(i >> 8)
			bt.GetTorrent(h)
			bt.ListTorrents()
		}
	}()
	wg.Wait()
}
