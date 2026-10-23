// Package leakcheck is the shared goroutine and heap sampler for long-running tests.
package leakcheck

import (
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"testing"
	"time"
)

// AssertNoGoroutineLeak polls until the process is back to baseline or a second
// has passed. Background cleanup is allowed to exit; a leftover is a leak.
func AssertNoGoroutineLeak(t *testing.T, baseline int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		n := runtime.NumGoroutine()
		if n <= baseline {
			return
		}
		if time.Now().After(deadline) {
			t.Errorf("goroutine leak: have %d, baseline %d", n, baseline)
			_ = pprof.Lookup("goroutine").WriteTo(os.Stderr, 2)
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// Heap runs two GCs and returns memory to the OS before sampling.
func Heap() runtime.MemStats {
	runtime.GC()
	runtime.GC()
	debug.FreeOSMemory()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}
