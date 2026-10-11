package torrstor

import (
	"sync"

	"github.com/anacrolix/torrent/metainfo"
)

// maxHandovers is how many files keep a record of what their last client was holding. Only the
// files paused on the devices in the house matter; the oldest record goes first.
const maxHandovers = 100

// The records live outside the cache, for as long as the server runs. They used to go with the
// torrent's cache — and ViMu drops its connection once its buffer is full and comes back when it
// needs more, after a pause many minutes later. A torrent with no readers is closed after a
// minute, so the connection that followed found nothing to inherit, took the client for empty and
// measured only what it added on top: 137MB of a 300MB buffer, the position twenty seconds in
// front of the picture for the rest of the film. A paused player holds its buffer for as long as
// it stays paused, so the record is kept without an expiry; where a new connection begins is
// what decides whether it applies.
var handovers = struct {
	sync.Mutex
	m map[string]*handover
}{m: map[string]*handover{}}

func handoverKey(hash metainfo.Hash, path string) string {
	return hash.HexString() + "/" + path
}

// handoverFor returns the record for a file, creating it if asked to, and must be called with
// handovers held.
func handoverFor(key string, create bool) *handover {
	h := handovers.m[key]
	if h != nil || !create {
		return h
	}
	if len(handovers.m) >= maxHandovers {
		var oldest string
		for k, v := range handovers.m {
			if oldest == "" || v.at.Before(handovers.m[oldest].at) {
				oldest = k
			}
		}
		delete(handovers.m, oldest)
	}
	h = &handover{}
	handovers.m[key] = h
	return h
}

// resetHandovers forgets every record; for tests.
func resetHandovers() {
	handovers.Lock()
	handovers.m = map[string]*handover{}
	handovers.Unlock()
}
