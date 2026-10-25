package torr

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/anacrolix/torrent/metainfo"

	sets "server/settings"
)

func withQuota(t *testing.T, maxSize, minFree int64, rotation bool) {
	t.Helper()
	old := sets.BTsets
	sets.BTsets = &sets.BTSets{OfflineMaxSize: maxSize, OfflineMinFreeSpace: minFree, OfflineRotation: rotation}
	t.Cleanup(func() { sets.BTsets = old })
}

func TestCurrentOfflineQuotaClampsNegative(t *testing.T) {
	withQuota(t, -5, -7, true)
	q := currentOfflineQuota()
	if q.maxSize != 0 || q.minFree != 0 || !q.rotation {
		t.Fatalf("unexpected quota %+v", q)
	}
}

func TestEnsureOfflineRoomNoQuotaEnoughSpace(t *testing.T) {
	withQuota(t, 0, 0, false)
	r := offlineRoom{dir: t.TempDir(), total: 1 << 20, remaining: 1 << 20}
	if err := ensureOfflineRoom(r); err != nil {
		t.Fatalf("a 1 MiB download must fit without quotas: %v", err)
	}
}

func TestEnsureOfflineRoomNothingToWrite(t *testing.T) {
	withQuota(t, 1, 1<<50, false)
	if err := ensureOfflineRoom(offlineRoom{dir: t.TempDir()}); err != nil {
		t.Fatalf("remaining 0 must not be checked: %v", err)
	}
}

func TestEnsureOfflineRoomTorrentLargerThanQuota(t *testing.T) {
	// even with the rotation on nothing can help: it is refused up front
	withQuota(t, 1<<20, 0, true)
	err := ensureOfflineRoom(offlineRoom{dir: t.TempDir(), total: 2 << 20, remaining: 2 << 20})
	if err == nil || !strings.Contains(err.Error(), "offline quota exceeded") {
		t.Fatalf("want a quota error, got %v", err)
	}
}

func TestEnsureOfflineRoomFreeSpaceReserve(t *testing.T) {
	// a reserve no disk can satisfy: refused, and the text names the reserve
	withQuota(t, 0, 1<<50, false)
	err := ensureOfflineRoom(offlineRoom{dir: t.TempDir(), total: 1 << 20, remaining: 1 << 20})
	if err == nil {
		t.Fatal("the free space reserve must be enforced")
	}
	if !strings.Contains(err.Error(), "not enough disk space") || !strings.Contains(err.Error(), "free space reserve") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureOfflineRoomRotationWithoutCandidates(t *testing.T) {
	// the rotation is on, but there is nothing to remove: a clear error, nothing locked
	withQuota(t, 0, 1<<50, true)
	err := ensureOfflineRoom(offlineRoom{dir: t.TempDir(), total: 1 << 20, remaining: 1 << 20})
	if err == nil || !strings.Contains(err.Error(), "rotation cannot free enough space") {
		t.Fatalf("want a rotation error, got %v", err)
	}
}

func TestSameOfflineDir(t *testing.T) {
	if !sameOfflineDir("/a/b/../c", "/a/c") {
		t.Fatal("paths must be compared cleaned")
	}
	if sameOfflineDir("/a/c", "/a/d") {
		t.Fatal("different directories")
	}
}

func TestTryLockOfflineTransition(t *testing.T) {
	var h metainfo.Hash
	h[0] = 0x42
	unlock, ok := tryLockOfflineTransition(h)
	if !ok {
		t.Fatal("a free lock must be taken")
	}
	if _, ok2 := tryLockOfflineTransition(h); ok2 {
		t.Fatal("a held lock must not be taken twice")
	}
	unlock()
	unlock2, ok3 := tryLockOfflineTransition(h)
	if !ok3 {
		t.Fatal("a released lock must be taken again")
	}
	unlock2()
}

func TestOfflineBeginUseTracksActive(t *testing.T) {
	var h metainfo.Hash
	h[0] = 0x43
	end := offlineBeginUse(h)
	if !offlineInUse(h) {
		t.Fatal("the copy must be in use while a request reads it")
	}
	end()
	end() // idempotent
	if offlineInUse(h) {
		t.Fatal("the copy must be free after the request")
	}
	if atomic.LoadInt64(&offlineUseOf(h).last) == 0 {
		t.Fatal("the last use time must be recorded")
	}
	offlineUses.Delete(h)
}

func TestOfflineLastUsedFallbacks(t *testing.T) {
	var h metainfo.Hash
	h[0] = 0x44
	db := &Torrent{Timestamp: 100}
	if got := offlineLastUsed(db, h); got != 100 {
		t.Fatalf("falls back to the add time, got %d", got)
	}
	db.setOfflineUsedAt(200)
	if got := offlineLastUsed(db, h); got != 200 {
		t.Fatalf("stored use time wins, got %d", got)
	}
	atomic.StoreInt64(&offlineUseOf(h).last, 300) // watched
	if got := offlineLastUsed(db, h); got != 300 {
		t.Fatalf("in-memory use time wins when newer, got %d", got)
	}
	offlineUses.Delete(h)
}

func TestOfflineEvictable(t *testing.T) {
	cases := []struct {
		st            string
		pinned, inUse bool
		want          bool
	}{
		{"ready", false, false, true},
		{"ready", true, false, false}, // pinned
		{"ready", false, true, false}, // being watched
		{"downloading", false, false, false},
		{"paused", false, false, false},
		{"", false, false, false},
	}
	for _, c := range cases {
		if got := offlineEvictable(c.st, c.pinned, c.inUse); got != c.want {
			t.Errorf("offlineEvictable(%q, pinned=%v, inUse=%v) = %v, want %v", c.st, c.pinned, c.inUse, got, c.want)
		}
	}
}

func TestOfflinePinnedAccessors(t *testing.T) {
	tr := &Torrent{}
	if tr.OfflinePinned() {
		t.Fatal("not pinned by default")
	}
	tr.setOfflinePinned(true)
	if !tr.OfflinePinned() {
		t.Fatal("pin must be stored")
	}
}

func TestFinishedCopyIsNotWatched(t *testing.T) {
	var h metainfo.Hash
	h[0] = 0x45
	markOfflineFinished(h, 1000)
	if offlineWatchedAt(h) != 0 {
		t.Fatal("finishing a download must not protect the copy from the rotation")
	}
	touchOffline(h)
	if offlineWatchedAt(h) == 0 {
		t.Fatal("reading the copy must be remembered")
	}
	offlineUses.Delete(h)
}
