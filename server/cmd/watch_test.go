package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchDirStop(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent")
	watchDir(missing, nil)

	dir := t.TempDir()
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		watchDir(dir, stop)
		close(done)
	}()
	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.torrent"), []byte("not a torrent"), 0o644); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	close(stop)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("watchDir did not return")
	}
}
