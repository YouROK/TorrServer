package torr

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anacrolix/torrent/metainfo"

	"server/log"
	sets "server/settings"
	"server/torr/offline"
	"server/torr/state"
)

// Disk quotas of the offline mode.
//
// Two independent limits (both 0 = disabled, see settings BTSets):
//
//   - OfflineMinFreeSpace: this much free space must stay on the disk that
//     hosts the offline directory after the new download has been accounted for;
//   - OfflineMaxSize: the offline torrents together (ready + downloading, and
//     the already downloaded part of paused ones) may not occupy more than this.
//
// When a new offline download does not fit and OfflineRotation is on, the least
// recently used READY offline copies are removed (the torrent itself stays in the
// list, as a normal streaming torrent) until it fits. Nothing is removed when the
// space cannot be freed in full: the plan is made first, and executed only if it
// is sufficient. Copies that are pinned (SetOfflinePinned), being watched, or watched
// during the last offlineEvictGrace, and torrents that are not ready are never removed.

// OnOfflineEvicted is called (in its own goroutine) after the rotation removed the
// offline copy of a torrent to make room for another one.
var OnOfflineEvicted func(hash, title string)

const (
	// A copy used this recently is not a rotation candidate: players open many short
	// connections, and the user most likely continues watching.
	offlineEvictGrace = 15 * time.Minute
	// The last-use time is written to the DB at most this often per torrent.
	offlineTouchPersistEvery = 10 * time.Minute
)

// offlineQuotaMu serialises the quota decisions (check + eviction + the write of the new
// offline state), so two torrents switched to the offline mode at the same moment cannot
// both be admitted into the same free room. Lock order: transition lock of the torrent
// -> offlineQuotaMu -> TRY-lock of the transition lock of a victim (never blocking, so
// no cycle is possible).
var offlineQuotaMu sync.Mutex

func lockOfflineQuota() func() {
	offlineQuotaMu.Lock()
	return offlineQuotaMu.Unlock
}

func tryLockOfflineTransition(hash metainfo.Hash) (func(), bool) {
	v, _ := offlineTransitions.LoadOrStore(hash, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	if !mu.TryLock() {
		return nil, false
	}
	return mu.Unlock, true
}

// ---- last use bookkeeping -------------------------------------------------

func (t *Torrent) offlineUsedAt() int64 {
	t.muOffline.RLock()
	defer t.muOffline.RUnlock()
	return t.offlineUsed
}

func (t *Torrent) setOfflineUsedAt(ts int64) {
	t.muOffline.Lock()
	t.offlineUsed = ts
	t.muOffline.Unlock()
}

// OfflinePinned reports whether the offline copy is protected from the quota rotation.
func (t *Torrent) OfflinePinned() bool {
	t.muOffline.RLock()
	defer t.muOffline.RUnlock()
	return t.offlinePin
}

func (t *Torrent) setOfflinePinned(v bool) {
	t.muOffline.Lock()
	t.offlinePin = v
	t.muOffline.Unlock()
}

// SetOfflinePinned pins (true) or unpins a torrent that is in the offline mode: a pinned
// copy is never removed by the rotation. The pin is dropped together with the offline mode.
func SetOfflinePinned(hashHex string, pinned bool) (*Torrent, error) {
	if sets.ReadOnly {
		return nil, errors.New("read-only DB mode")
	}
	hash := metainfo.NewHashFromHex(hashHex)
	unlock := lockOfflineTransition(hash)
	defer unlock()
	db := GetTorrentDB(hash)
	if db == nil {
		return nil, errors.New("torrent not found")
	}
	if db.OfflineState() == "" {
		return nil, errors.New("the torrent is not in the offline mode")
	}
	if !sets.SetOfflinePinned(hash, pinned) {
		return nil, errors.New("cannot save the torrent")
	}
	if live := bts.GetTorrent(hash); live != nil {
		live.setOfflinePinned(pinned)
	}
	db.setOfflinePinned(pinned)
	return db, nil
}

// offlineEvictable tells whether the rotation may remove the copy.
func offlineEvictable(offState string, pinned, inUse bool) bool {
	return offState == state.OfflineReady && !pinned && !inUse
}

// offlineUse is the in-memory usage record of an offline copy. The int64 fields come
// first and are only accessed atomically.
type offlineUse struct {
	last   int64 // unix time of the last use
	saved  int64 // unix time the value was last written to the DB
	active int32 // number of requests reading the copy right now
}

var offlineUses sync.Map // metainfo.Hash -> *offlineUse

func offlineUseOf(hash metainfo.Hash) *offlineUse {
	if v, ok := offlineUses.Load(hash); ok {
		return v.(*offlineUse)
	}
	v, _ := offlineUses.LoadOrStore(hash, &offlineUse{})
	return v.(*offlineUse)
}

// offlineBeginUse marks the copy as being read (it is not a rotation candidate meanwhile).
// The returned function must be called when the request is over.
func offlineBeginUse(hash metainfo.Hash) func() {
	u := offlineUseOf(hash)
	atomic.AddInt32(&u.active, 1)
	touchOffline(hash)
	var once sync.Once
	return func() {
		once.Do(func() {
			atomic.AddInt32(&u.active, -1)
			touchOffline(hash)
		})
	}
}

func offlineInUse(hash metainfo.Hash) bool {
	return atomic.LoadInt32(&offlineUseOf(hash).active) > 0
}

// markOfflineFinished is called when the download completes: the caller stores the time in the
// record itself, so the throttle of the DB writes is told that it is already saved. It does NOT
// count as watching: a freshly finished copy that nobody has opened is not protected from the rotation.
func markOfflineFinished(hash metainfo.Hash, ts int64) {
	atomic.StoreInt64(&offlineUseOf(hash).saved, ts)
}

// offlineWatchedAt is when a request last read the copy since the server started (0 = never).
func offlineWatchedAt(hash metainfo.Hash) int64 {
	return atomic.LoadInt64(&offlineUseOf(hash).last)
}

func touchOffline(hash metainfo.Hash) {
	u := offlineUseOf(hash)
	ts := time.Now().Unix()
	atomic.StoreInt64(&u.last, ts)
	saved := atomic.LoadInt64(&u.saved)
	if ts-saved < int64(offlineTouchPersistEvery/time.Second) {
		return
	}
	if !atomic.CompareAndSwapInt64(&u.saved, saved, ts) {
		return
	}
	go persistOfflineUsed(hash, ts)
}

// persistOfflineUsed stores the last-use time, so the LRU order survives a restart. It is
// a read-modify-write of the torrent record and therefore runs under the transition lock.
func persistOfflineUsed(hash metainfo.Hash, ts int64) {
	if sets.ReadOnly || !sets.IsDBOpen() {
		return
	}
	unlock := lockOfflineTransition(hash)
	defer unlock()
	db := GetTorrentDB(hash)
	if db == nil || db.OfflineState() == "" {
		return
	}
	db.setOfflineUsedAt(ts)
	AddTorrentDB(db)
}

// offlineLastUsed orders the copies for the rotation (least recently used first): the last
// time it was watched, falling back to the time it finished downloading / was added.
func offlineLastUsed(db *Torrent, hash metainfo.Hash) int64 {
	last := db.offlineUsedAt()
	if m := atomic.LoadInt64(&offlineUseOf(hash).last); m > last {
		last = m
	}
	if last == 0 {
		last = db.Timestamp
	}
	return last
}

// ---- quota arithmetic -----------------------------------------------------

type offlineQuota struct {
	maxSize  int64
	minFree  int64
	rotation bool
}

func currentOfflineQuota() offlineQuota {
	b := sets.BTsets
	if b == nil {
		return offlineQuota{}
	}
	q := offlineQuota{maxSize: b.OfflineMaxSize, minFree: b.OfflineMinFreeSpace, rotation: b.OfflineRotation}
	if q.maxSize < 0 {
		q.maxSize = 0
	}
	if q.minFree < 0 {
		q.minFree = 0
	}
	return q
}

func sameOfflineDir(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// offlineCommitted is the space the offline copy occupies or is going to occupy:
// the whole torrent for ready and downloading ones (a running download is counted in
// full, otherwise several parallel downloads would all be admitted into the same
// room), the downloaded part for paused ones.
func offlineCommitted(db *Torrent, hash metainfo.Hash) int64 {
	total := offlineTotalLength(db)
	switch db.OfflineState() {
	case state.OfflineReady, state.OfflineDownloading:
		return total
	case state.OfflinePaused:
		if rest := offlineRemainingBytes(db, hash); total > rest {
			return total - rest
		}
	}
	return 0
}

// offlineUsageOthers sums the other offline torrents: the committed size of all of
// them, and the bytes that running downloads in dir still have to write (they are
// not on the disk yet, but the free space is already spoken for).
func offlineUsageOthers(self metainfo.Hash, dir string) (used, reserved int64) {
	for hash, db := range ListTorrentsDB() {
		if hash == self {
			continue
		}
		st := db.OfflineState()
		if st == "" {
			continue
		}
		used += offlineCommitted(db, hash)
		if st == state.OfflineDownloading && sameOfflineDir(db.offlineBase(), dir) {
			reserved += offlineRemainingBytes(db, hash)
		}
	}
	return used, reserved
}

// offlineRoom describes a download that is about to start (or resume).
type offlineRoom struct {
	self      metainfo.Hash
	dir       string
	total     int64 // size the torrent will occupy when finished (0 = unknown yet)
	remaining int64 // bytes that still have to be written
}

type offlineVictim struct {
	hash   metainfo.Hash
	db     *Torrent
	size   int64
	unlock func()
}

// ensureOfflineRoom checks both quotas and the physical free space for the download and,
// when the rotation is enabled, makes room by removing the least recently used ready
// copies. The caller MUST hold offlineQuotaMu (lockOfflineQuota) and keep holding it
// until the new state of the torrent is saved. It must not hold the transition lock of
// another torrent that could become a victim.
func ensureOfflineRoom(r offlineRoom) error {
	if r.remaining <= 0 {
		return nil
	}
	q := currentOfflineQuota()
	if q.maxSize > 0 && r.total > q.maxSize {
		return fmt.Errorf("offline quota exceeded: the torrent needs %s, the limit for all offline torrents is %s",
			bytesHuman(r.total), bytesHuman(q.maxSize))
	}

	used, reserved := offlineUsageOthers(r.self, r.dir)
	var needSize, needFree int64
	if q.maxSize > 0 {
		needSize = used + r.total - q.maxSize
	}
	free, ferr := offline.FreeSpace(r.dir)
	freeKnown := ferr == nil
	if !freeKnown {
		log.TLogln("offline: cannot determine the free space on", r.dir, ferr)
	} else {
		needFree = r.remaining + offlineSpaceMargin + q.minFree + reserved - free
	}
	if needSize <= 0 && needFree <= 0 {
		return nil
	}

	base := offlineRoomError(r, q, used, reserved, free, needSize)
	if !q.rotation {
		return base
	}

	victims, st := planOfflineRotation(r, q, &needSize, &needFree)
	if needSize > 0 || needFree > 0 {
		for _, v := range victims {
			v.unlock()
		}
		// Say which limit is still not met: `base` names only the size quota, but the
		// free space of the disk may be what the rotation could not satisfy.
		var missing []string
		if needSize > 0 {
			missing = append(missing, fmt.Sprintf("%s over the size quota", bytesHuman(needSize)))
		}
		if needFree > 0 {
			missing = append(missing, fmt.Sprintf("%s of free space on %s", bytesHuman(needFree), r.dir))
		}
		err := fmt.Errorf("%v; the rotation cannot free enough space: still missing %s; %s (only ready, not pinned offline torrents that were not watched in the last %d min can be removed; for the free space only those stored in the same directory help)",
			base, strings.Join(missing, " and "), st, int(offlineEvictGrace/time.Minute))
		log.TLogln("offline: rotation failed:", err)
		return err
	}
	for _, v := range victims {
		evictOffline(v)
	}
	return nil
}

func offlineRoomError(r offlineRoom, q offlineQuota, used, reserved, free, needSize int64) error {
	if needSize > 0 {
		return fmt.Errorf("offline quota exceeded: limit %s, used by the other offline torrents %s, this torrent needs %s",
			bytesHuman(q.maxSize), bytesHuman(used), bytesHuman(r.total))
	}
	need := r.remaining + offlineSpaceMargin + reserved + q.minFree
	extra := ""
	if q.minFree > 0 {
		extra = fmt.Sprintf(", including the %s free space reserve", bytesHuman(q.minFree))
	}
	return fmt.Errorf("not enough disk space for the offline download: need %s%s, available %s (%s)",
		bytesHuman(need), extra, bytesHuman(free), r.dir)
}

// planOfflineRotation picks the victims, oldest use first, and locks their transition
// locks (try-lock: a copy in the middle of a transition is skipped). Nothing is removed
// here; the caller decides after looking at the remaining needs.
func planOfflineRotation(r offlineRoom, q offlineQuota, needSize, needFree *int64) ([]offlineVictim, rotationStats) {
	var st rotationStats
	type cand struct {
		hash metainfo.Hash
		db   *Torrent
		last int64
	}
	var cands []cand
	for hash, db := range ListTorrentsDB() {
		if hash == r.self {
			continue
		}
		off := db.OfflineState()
		switch {
		case off == "":
			continue
		case off != state.OfflineReady:
			st.notReady++
		case db.OfflinePinned():
			st.pinned++
		case offlineInUse(hash):
			st.inUse++
		default:
			cands = append(cands, cand{hash, db, offlineLastUsed(db, hash)})
		}
	}
	st.candidates = len(cands)
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].last != cands[j].last {
			return cands[i].last < cands[j].last
		}
		return cands[i].db.Timestamp < cands[j].db.Timestamp
	})

	cutoff := time.Now().Add(-offlineEvictGrace).Unix()
	var victims []offlineVictim
	for _, c := range cands {
		if *needSize <= 0 && *needFree <= 0 {
			break
		}
		// Only real reading protects a copy (the player may reconnect any moment); the
		// order above also uses the finish time, which does not protect anything.
		if w := offlineWatchedAt(c.hash); w > cutoff {
			st.recent++
			if len(st.recentInfo) < 3 {
				st.recentInfo = append(st.recentInfo, fmt.Sprintf("%q watched %d min ago",
					c.db.Title, (time.Now().Unix()-w)/60))
			}
			continue
		}
		size := offlineTotalLength(c.db)
		if size <= 0 {
			st.noSize++
			continue
		}
		helpsSize := *needSize > 0
		// removing a copy on another disk does not help the free space of this one
		helpsFree := *needFree > 0 && sameOfflineDir(c.db.offlineBase(), r.dir)
		if !helpsSize && !helpsFree {
			st.otherDir++
			continue
		}
		unlock, ok := tryLockOfflineTransition(c.hash)
		if !ok {
			st.busy++
			continue
		}
		// the state may have changed before the lock was taken
		fresh := GetTorrentDB(c.hash)
		if fresh == nil || !offlineEvictable(fresh.OfflineState(), fresh.OfflinePinned(), offlineInUse(c.hash)) {
			unlock()
			st.changed++
			continue
		}
		st.evicted++
		victims = append(victims, offlineVictim{hash: c.hash, db: fresh, size: size, unlock: unlock})
		if helpsSize {
			*needSize -= size
		}
		if helpsFree {
			*needFree -= size
		}
	}
	return victims, st
}

// rotationStats explains what the rotation saw; it goes to the error and the log.
type rotationStats struct {
	candidates int // ready, not pinned, not being read
	notReady   int // downloading / paused: their space cannot be freed
	pinned     int
	inUse      int
	recent     int // watched (or finished) within offlineEvictGrace
	noSize     int // the size of the copy is unknown
	otherDir   int // stored in another directory: does not help the free space here
	busy       int // in the middle of a state transition
	changed    int
	evicted    int // planned for removal
	recentInfo []string
}

func (s rotationStats) String() string {
	out := fmt.Sprintf("offline torrents seen: %d removable candidates (planned %d), skipped: %d not ready, %d pinned, %d in use, %d watched in the last %d min, %d in another directory, %d busy, %d without known size",
		s.candidates, s.evicted, s.notReady, s.pinned, s.inUse, s.recent, int(offlineEvictGrace/time.Minute), s.otherDir, s.busy+s.changed, s.noSize)
	if len(s.recentInfo) > 0 {
		out += " [" + strings.Join(s.recentInfo, "; ") + "]"
	}
	return out
}

// evictOffline removes the offline copy of a victim (its transition lock is held and is
// released by disableOffline). The torrent stays in the list as a normal one.
func evictOffline(v offlineVictim) {
	title := v.db.Title
	log.TLogln("offline: rotation removes the copy of", v.hash.HexString(), title, bytesHuman(v.size))
	if _, err := disableOffline(v.hash, v.unlock); err != nil {
		log.TLogln("offline: rotation cannot remove", v.hash.HexString(), err)
		return
	}
	if cb := OnOfflineEvicted; cb != nil {
		go cb(v.hash.HexString(), title)
	}
}

// ---- state for the API ----------------------------------------------------

// OfflineQuotaInfo is the state of the offline storage for the settings page.
type OfflineQuotaInfo struct {
	MaxSize      int64  `json:"max_size"`       // 0 = unlimited
	MinFreeSpace int64  `json:"min_free_space"` // 0 = no reserve
	Rotation     bool   `json:"rotation"`
	Used         int64  `json:"used"` // occupied or reserved by offline torrents
	Free         int64  `json:"free"` // free space of the offline disk, -1 = unknown
	Dir          string `json:"dir"`
}

func GetOfflineQuotaInfo() OfflineQuotaInfo {
	q := currentOfflineQuota()
	dir := offlineBaseDir()
	info := OfflineQuotaInfo{MaxSize: q.maxSize, MinFreeSpace: q.minFree, Rotation: q.rotation, Free: -1, Dir: dir}
	if bts != nil && sets.IsDBOpen() {
		for hash, db := range ListTorrentsDB() {
			if db.OfflineState() != "" {
				info.Used += offlineCommitted(db, hash)
			}
		}
	}
	// the directory may not exist yet: ask the nearest existing parent
	probe := dir
	for i := 0; i < 64; i++ {
		if _, err := os.Stat(probe); err == nil {
			if free, err := offline.FreeSpace(probe); err == nil {
				info.Free = free
			}
			break
		}
		parent := filepath.Dir(probe)
		if parent == probe {
			break
		}
		probe = parent
	}
	return info
}
