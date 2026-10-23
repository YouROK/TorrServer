package utils

import (
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"
)

var lastFreeOSMemGC atomic.Int64

func FreeOSMem() {
	debug.FreeOSMemory()
}

// ponytail: sync STW; torrstor uses FreeOSMemGCThrottled — migrate btserver/rutor callers the same way if Disconnect/index churn spikes CPU.
func FreeOSMemGC() {
	runtime.GC()
	debug.FreeOSMemory()
}

// FreeOSMemGCThrottled coalesces burst rem/Close reclaim (≥4s, async).
func FreeOSMemGCThrottled() {
	now := time.Now().UnixNano()
	last := lastFreeOSMemGC.Load()
	if now-last < 4*int64(time.Second) {
		return
	}
	if !lastFreeOSMemGC.CompareAndSwap(last, now) {
		return
	}
	go FreeOSMemGC()
}
