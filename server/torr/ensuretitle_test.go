package torr

import (
	"sync"
	"testing"
)

func TestEnsureTitleKeepsExistingAndIsRaceFree(t *testing.T) {
	tr := &Torrent{}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			tr.EnsureTitle(func() string { return "fallback" })
		}()
		go func() {
			defer wg.Done()
			tr.muTorrent.Lock() // what SetTorrent does
			tr.Title = "renamed"
			tr.muTorrent.Unlock()
		}()
	}
	wg.Wait()
	tr.EnsureTitle(func() string { return "late" })
	if tr.Title != "renamed" && tr.Title != "fallback" {
		t.Fatalf("title = %q", tr.Title)
	}

	tr.Title = "keep"
	called := false
	tr.EnsureTitle(func() string { called = true; return "x" })
	if called || tr.Title != "keep" {
		t.Fatalf("existing title replaced: %q called=%v", tr.Title, called)
	}
}
