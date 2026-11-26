package torrstor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

// A read waits for data no peer has. When the consumer goes away (the player closed the
// HTTP connection), the read has to end, otherwise the reader keeps its priorities and
// keeps the torrent alive.
func TestReaderReadStopsWhenContextIsCanceled(t *testing.T) {
	initTestSettings(t)
	stor := NewStorage(1 << 20)

	cfg := torrent.NewDefaultClientConfig()
	cfg.DefaultStorage = stor
	cfg.DataDir = t.TempDir()
	cfg.ListenPort = 0
	cfg.NoDHT = true
	cfg.DisableUTP = true
	cfg.DisableIPv6 = true
	cfg.NoDefaultPortForwarding = true
	cl, err := torrent.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close()

	infoBytes, err := bencode.Marshal(testInfo())
	if err != nil {
		t.Fatal(err)
	}
	tor, err := cl.AddTorrent(&metainfo.MetaInfo{InfoBytes: infoBytes})
	if err != nil {
		t.Fatal(err)
	}
	<-tor.GotInfo()

	cache := stor.GetCache(tor.InfoHash())
	if cache == nil {
		t.Fatal("no cache for the torrent")
	}
	r := cache.NewReader(tor.Files()[0])
	defer cache.CloseReader(r)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r.SetContext(ctx)

	done := make(chan error, 1)
	go func() {
		_, err := r.Read(make([]byte, 1<<10))
		done <- err
	}()

	select {
	case err := <-done:
		t.Fatalf("Read returned without data: %v", err)
	case <-time.After(300 * time.Millisecond):
	}

	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Read returned %v, want context.Canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Read is still waiting for data after the context was canceled")
	}
}
