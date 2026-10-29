package utils

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestFreeOSMemGCThrottledRunsTrailingCall(t *testing.T) {
	var runs atomic.Int32
	oldFunc, oldInterval := freeOSMemGCFunc, freeOSMemGCInterval
	freeOSMemGCFunc = func() { runs.Add(1) }
	freeOSMemGCInterval = 100 * time.Millisecond
	lastFreeOSMemGC.Store(0)
	t.Cleanup(func() {
		freeOSMemGCFunc, freeOSMemGCInterval = oldFunc, oldInterval
		lastFreeOSMemGC.Store(0)
	})

	waitRuns := func(want int32) {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		for runs.Load() < want && time.Now().Before(deadline) {
			time.Sleep(5 * time.Millisecond)
		}
		if got := runs.Load(); got != want {
			t.Fatalf("runs = %d, want %d", got, want)
		}
	}

	FreeOSMemGCThrottled()
	waitRuns(1)

	// Eviction burst, then Close inside the window: one trailing run, not zero.
	start := time.Now()
	for range 3 {
		FreeOSMemGCThrottled()
	}
	waitRuns(2)
	if time.Since(start) < 50*time.Millisecond {
		t.Fatal("trailing run did not wait for the throttle window")
	}
	time.Sleep(200 * time.Millisecond)
	if got := runs.Load(); got != 2 {
		t.Fatalf("burst coalesced into %d runs, want 2", got)
	}
}
