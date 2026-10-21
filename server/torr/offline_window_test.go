package torr

import (
	"testing"
	"time"
)

func TestNewOfflineWindowSizing(t *testing.T) {
	// 128 MiB of 1 MiB pieces -> 128 pieces in the window
	w := newOfflineWindow(1<<20, 1000)
	if w.base != 128 || w.size != 128 {
		t.Fatalf("window of 1MiB pieces: base=%d size=%d, want 128", w.base, w.size)
	}
	// fewer pieces than the window: the window is capped by the piece count
	w = newOfflineWindow(256<<10, 10)
	if w.base != 10 || w.size != 10 || w.total != 10 {
		t.Fatalf("small torrent: base=%d size=%d total=%d, want 10", w.base, w.size, w.total)
	}
	// no piece length known yet: the minimal window of 8 pieces is used
	w = newOfflineWindow(0, 50)
	if w.base != 8 || w.size != 8 {
		t.Fatalf("unknown piece length: base=%d size=%d, want 8", w.base, w.size)
	}
}

func TestOfflineWindowProgress(t *testing.T) {
	w := newOfflineWindow(1<<20, 1000) // base = 128
	t0 := time.Now()

	w.progress(10, t0) // the first report establishes the baseline
	if w.size != w.base {
		t.Fatalf("first report: size=%d, want base %d", w.size, w.base)
	}

	w.progress(10, t0.Add(time.Second)) // no new pieces: nothing changes
	if w.size != w.base {
		t.Fatalf("no progress: size=%d, want base %d", w.size, w.base)
	}

	w.progress(11, t0.Add(2*time.Second)) // a piece got verified: collapse back
	if w.size != w.base || w.lastVerified != 11 {
		t.Fatalf("new verified piece: size=%d lastVerified=%d, want size %d and 11", w.size, w.lastVerified, w.base)
	}

	// stalled longer than offlineWindowStall: the window doubles
	w.progress(11, t0.Add(2*time.Second+offlineWindowStall+time.Second))
	if w.size != 2*w.base {
		t.Fatalf("stalled: size=%d, want %d", w.size, 2*w.base)
	}

	// more stalls double it further, but never beyond the piece count
	now := t0.Add(2 * time.Second)
	for w.size < w.total {
		now = now.Add(offlineWindowStall + time.Second)
		w.progress(11, now)
	}
	if w.size != w.total {
		t.Fatalf("repeated stalls: size=%d, want %d", w.size, w.total)
	}
	now = now.Add(offlineWindowStall + time.Second)
	w.progress(11, now)
	if w.size != w.total {
		t.Fatalf("stall past the cap: size=%d, want %d", w.size, w.total)
	}

	// new progress collapses the window again
	w.progress(12, now.Add(time.Second))
	if w.size != w.base || w.lastVerified != 12 {
		t.Fatalf("progress after stall: size=%d lastVerified=%d, want size %d and 12", w.size, w.lastVerified, w.base)
	}
}
