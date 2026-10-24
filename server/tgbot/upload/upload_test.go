package upload

import (
	"strings"
	"sync"
	"testing"

	"server/tgbot/config"
	"server/torr/state"
)

func TestNewTorrFileSizeLimits(t *testing.T) {
	old := config.Cfg
	t.Cleanup(func() { config.Cfg = old })

	config.Cfg = nil
	_, err := NewTorrFile(&Worker{torrentHash: "h"}, &state.TorrentFileStat{Path: "a.mkv", Length: 50*1024*1024 + 1})
	if err == nil || !strings.Contains(err.Error(), "50") && !strings.Contains(err.Error(), "upload_file_too_large_50mb") {
		t.Fatalf("cloud limit: %v", err)
	}

	config.Cfg = &config.Config{HostTG: "http://127.0.0.1:8081"}
	_, err = NewTorrFile(&Worker{}, &state.TorrentFileStat{Path: "a.mkv", Length: 2*1024*1024*1024 + 1})
	if err == nil || !strings.Contains(err.Error(), "2") && !strings.Contains(err.Error(), "upload_file_too_large_2gb") {
		t.Fatalf("local limit: %v", err)
	}
}

func TestTorrFileRemainingAndCancel(t *testing.T) {
	f := &TorrFile{size: 100, offset: 40, wrk: &Worker{isCancelled: true}}
	if f.Remaining() != 60 {
		t.Fatalf("remaining %d", f.Remaining())
	}
	n, err := f.Read(make([]byte, 8))
	if n != 0 || err != ErrStopped {
		t.Fatalf("read %d %v", n, err)
	}
	f.Close()
	f.Close()
}

func TestManagerCancelConcurrent(t *testing.T) {
	m := &Manager{working: map[int]*Worker{}}
	for i := 0; i < 50; i++ {
		m.working[i] = &Worker{id: i}
	}
	var wg sync.WaitGroup
	for g := 0; g < 40; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				m.Cancel(i)
			}
		}()
	}
	wg.Wait()
	for id, w := range m.working {
		if !w.isCancelled {
			t.Fatalf("worker %d not cancelled", id)
		}
	}
}
