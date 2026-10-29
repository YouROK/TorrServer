package utils

import (
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"
)

var (
	lastFreeOSMemGC    atomic.Int64
	freeOSMemGCPending atomic.Bool

	freeOSMemGCInterval = 4 * time.Second
	freeOSMemGCFunc     = FreeOSMemGC
)

func FreeOSMem() {
	debug.FreeOSMemory()
}

// ponytail: sync STW; torrstor uses FreeOSMemGCThrottled — migrate btserver/rutor callers the same way if Disconnect/index churn spikes CPU.
func FreeOSMemGC() {
	runtime.GC()
	debug.FreeOSMemory()
}

// FreeOSMemGCThrottled coalesces burst rem/Close reclaim: at most one async GC
// per interval. A call inside the window runs at its end instead of being
// dropped, otherwise Cache.Close right after an eviction never returns the cache.
func FreeOSMemGCThrottled() {
	if !freeOSMemGCPending.CompareAndSwap(false, true) {
		return
	}
	wait := time.Duration(lastFreeOSMemGC.Load() + int64(freeOSMemGCInterval) - time.Now().UnixNano())
	time.AfterFunc(max(wait, 0), func() {
		lastFreeOSMemGC.Store(time.Now().UnixNano())
		freeOSMemGCPending.Store(false)
		freeOSMemGCFunc()
	})
}
