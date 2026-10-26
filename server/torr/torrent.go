package torr

import (
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"

	"server/torrshash"

	utils2 "server/utils"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/log"
	"server/settings"
	"server/torr/offline"
	"server/torr/state"
	cacheSt "server/torr/storage/state"
	"server/torr/storage/torrstor"
	"server/torr/utils"
)

type Torrent struct {
	Title    string
	Category string
	Poster   string
	Data     string
	*torrent.TorrentSpec

	Stat      state.TorrentStat
	Timestamp int64
	Size      int64

	*torrent.Torrent
	muTorrent sync.Mutex

	bt    *BTServer
	cache *torrstor.Cache

	lastTimeSpeed       time.Time
	DownloadSpeed       float64
	UploadSpeed         float64
	BytesReadUsefulData int64
	BytesWrittenData    int64

	PreloadSize    int64
	PreloadedBytes int64

	DurationSeconds float64
	BitRate         string

	expiredTime time.Time

	closed <-chan struct{}

	progressTicker *time.Ticker

	// offline mode ("", state.OfflineDownloading, state.OfflineReady)
	offlineState string
	offlineDir   string // base directory of the offline data ("" = current default)
	offlineUsed  int64  // unix time of the last use of the offline copy (muOffline)
	offlinePin   bool   // the offline copy is protected from the rotation (muOffline)
	muOffline    sync.RWMutex
	offStore     *offline.Storage // set for torrents that are downloaded in offline mode
}

func NewTorrent(spec *torrent.TorrentSpec, bt *BTServer) (*Torrent, error) {
	// https://github.com/anacrolix/torrent/issues/747
	if bt == nil || bt.client == nil {
		return nil, errors.New("BT client not connected")
	}

	// NewTorrent is called concurrently for the same torrent: every status request of the web UI
	// for a torrent that is not loaded yet starts one. The check and the registration must be one
	// step. A second AddTorrentSpec for a torrent that is already in the client initialises it
	// again (the storage is opened a second time and the pieces in flight are reset), which broke
	// running offline downloads and could crash the server.
	bt.newMu.Lock()
	defer bt.newMu.Unlock()
	bt.mu.RLock()
	existing := bt.torrents[spec.InfoHash]
	bt.mu.RUnlock()
	if existing != nil {
		// AddTorrentSpec used to merge the trackers of the new spec into a torrent
		// that is already loaded; keep doing that without a second AddTorrentSpec
		if len(spec.Trackers) > 0 {
			existing.muTorrent.Lock()
			gt := existing.Torrent
			existing.muTorrent.Unlock()
			if gt != nil {
				gt.AddTrackers(spec.Trackers)
			}
		}
		return existing, nil
	}

	switch settings.BTsets.RetrackersMode {
	case 1:
		spec.Trackers = append(spec.Trackers, [][]string{utils.GetDefTrackers()}...)
	case 2:
		spec.Trackers = nil
	case 3:
		spec.Trackers = [][]string{utils.GetDefTrackers()}
	}

	trackers := utils.GetTrackerFromFile()
	if len(trackers) > 0 {
		spec.Trackers = append(spec.Trackers, [][]string{trackers}...)
	}

	offlineState := ""
	offlineDir := ""
	if db := GetTorrentDB(spec.InfoHash); db != nil {
		offlineState = db.OfflineState()
		offlineDir = db.offlineBase()
		if len(spec.InfoBytes) == 0 && db.TorrentSpec != nil {
			spec.InfoBytes = db.TorrentSpec.InfoBytes
		}
	}

	// Offline torrents are stored by their own file storage. The storage is set on a
	// copy of the spec only: the original is persisted to the DB as JSON.
	addSpec := spec
	var offStore *offline.Storage
	if offlineState == state.OfflineDownloading {
		bt.mu.RLock()
		off := bt.offline
		bt.mu.RUnlock()
		if off == nil {
			return nil, errors.New("offline storage is not available")
		}
		off.Use(spec.InfoHash, offlineDir)
		offStore = off
		cp := *spec
		cp.Storage = off
		addSpec = &cp
	}

	// The read lock is taken and released explicitly, NOT with defer: the code below
	// takes bt.mu.Lock() for the torrents map, and sync.RWMutex is not reentrant,
	// so holding an RLock across that would self-deadlock.
	bt.mu.RLock()
	if bt.client == nil {
		bt.mu.RUnlock()
		return nil, errors.New("BT client not connected")
	}
	goTorrent, _, err := bt.client.AddTorrentSpec(addSpec)
	bt.mu.RUnlock()
	if err != nil {
		return nil, err
	}

	bt.mu.Lock()
	defer bt.mu.Unlock()
	if tor, ok := bt.torrents[spec.InfoHash]; ok {
		return tor, nil
	}

	timeout := time.Second * time.Duration(settings.BTsets.TorrentDisconnectTimeout)
	if timeout > time.Minute {
		timeout = time.Minute
	}

	torr := new(Torrent)
	torr.Torrent = goTorrent
	torr.Stat = state.TorrentAdded
	torr.lastTimeSpeed = time.Now()
	torr.bt = bt
	torr.closed = goTorrent.Closed()
	torr.TorrentSpec = spec
	torr.AddExpiredTime(timeout)
	torr.Timestamp = time.Now().Unix()
	torr.offlineState = offlineState
	torr.offStore = offStore
	if offlineState != "" {
		torr.offlineDir = offlineDir
	}

	go torr.watch()

	bt.torrents[spec.InfoHash] = torr
	if offlineState == state.OfflineDownloading {
		go torr.runOffline()
	}
	return torr, nil
}

func (t *Torrent) WaitInfo() bool {
	if t == nil {
		return false
	}
	// drop() (Close, "drop all torrents") sets t.Torrent to nil at any moment, e.g. on
	// "drop" from Lampa while stream requests wait here. Take the pointer once under
	// muTorrent and use only it below: it stays valid after Drop, and t.closed fires.
	t.muTorrent.Lock()
	torrent := t.Torrent
	t.muTorrent.Unlock()
	if torrent == nil {
		return false
	}

	// Close torrent if no info in 1 minute + TorrentDisconnectTimeout config option.
	// An offline download waits for the metadata of a rare torrent as long as it takes.
	var timeout <-chan time.Time
	if t.OfflineState() != state.OfflineDownloading {
		tm := time.NewTimer(time.Minute + time.Second*time.Duration(settings.BTsets.TorrentDisconnectTimeout))
		defer tm.Stop()
		timeout = tm.C
	}

	select {
	case <-torrent.GotInfo():
		if t.TorrentSpec != nil && len(t.TorrentSpec.InfoBytes) == 0 {
			t.TorrentSpec.InfoBytes = torrent.Metainfo().InfoBytes
		}
		if t.bt != nil && t.offStore == nil { // offline downloads have their own storage, no cache
			// bt.storage is replaced by Connect under bt.mu; snapshot it under the
			// lock and use the snapshot afterwards, because the storage object keeps
			// its own locking and is never closed. Reading the field unlocked was
			// reported by -race and can tear the interface value.
			t.bt.mu.RLock()
			storage := t.bt.storage
			t.bt.mu.RUnlock()
			if storage != nil {
				cache := storage.GetCache(t.Hash())
				if cache == nil {
					// torrent was dropped while waiting for info: CloseHash already removed the cache
					return false
				}
				t.cache = cache
				cache.SetTorrent(torrent)
			}
		}
		return true
	case <-t.closed:
		return false
	case <-timeout:
		return false
	}
}

func (t *Torrent) GotInfo() bool {
	// log.TLogln("GotInfo state:", t.Stat)
	if t == nil {
		return false
	}
	// A ready offline torrent has everything it needs on disk and no client torrent.
	if t.OfflineReady() {
		return true
	}
	// Every stream request calls this, often several at once for one file. Once the info
	// is in, only report it: flipping Stat through TorrentGettingInfo on each call raced
	// with Status, Preload and Close, and could turn a closed torrent back into a working one.
	t.muTorrent.Lock()
	switch {
	case t.Stat == state.TorrentClosed:
		t.muTorrent.Unlock()
		return false
	case t.Stat == state.TorrentPreload:
		// assume we have info in preload state
		// and dont override with TorrentWorking
		t.muTorrent.Unlock()
		return true
	case t.Stat == state.TorrentWorking && t.Torrent != nil && t.Torrent.Info() != nil:
		t.muTorrent.Unlock()
		t.AddExpiredTime(time.Second * time.Duration(settings.BTsets.TorrentDisconnectTimeout))
		return true
	}
	t.Stat = state.TorrentGettingInfo
	t.muTorrent.Unlock()

	if t.WaitInfo() {
		t.muTorrent.Lock()
		if t.Stat == state.TorrentGettingInfo {
			t.Stat = state.TorrentWorking
		}
		closed := t.Stat == state.TorrentClosed
		t.muTorrent.Unlock()
		if closed {
			return false
		}
		t.AddExpiredTime(time.Second * time.Duration(settings.BTsets.TorrentDisconnectTimeout))
		return true
	} else {
		t.Close()
		return false
	}
}

func (t *Torrent) AddExpiredTime(duration time.Duration) {
	newExpiredTime := time.Now().Add(duration)
	t.muTorrent.Lock()
	if t.expiredTime.Before(newExpiredTime) {
		t.expiredTime = newExpiredTime
	}
	t.muTorrent.Unlock()
}

func (t *Torrent) watch() {
	t.progressTicker = time.NewTicker(time.Second)
	defer t.progressTicker.Stop()

	for {
		select {
		case <-t.progressTicker.C:
			go t.progressEvent()
		case <-t.closed:
			return
		}
	}
}

func (t *Torrent) progressEvent() {
	if t.expired() {
		if t.TorrentSpec != nil {
			log.TLogln("Torrent close by timeout", t.TorrentSpec.InfoHash.HexString())
		}
		t.bt.RemoveTorrent(t.Hash())
		return
	}

	t.muTorrent.Lock()
	if t.Torrent != nil && t.Torrent.Info() != nil {
		st := t.Torrent.Stats()
		deltaDlBytes := st.BytesRead.Int64() - t.BytesReadUsefulData
		deltaUpBytes := st.BytesWritten.Int64() - t.BytesWrittenData
		deltaTime := time.Since(t.lastTimeSpeed).Seconds()

		t.DownloadSpeed = float64(deltaDlBytes) / deltaTime
		t.UploadSpeed = float64(deltaUpBytes) / deltaTime

		t.BytesReadUsefulData = st.BytesRead.Int64()
		t.BytesWrittenData = st.BytesWritten.Int64()

		if t.cache != nil {
			t.PreloadedBytes = t.cache.GetState().Filled
		}
	} else {
		t.DownloadSpeed = 0
		t.UploadSpeed = 0
	}
	t.muTorrent.Unlock()

	t.lastTimeSpeed = time.Now()
	t.updateRA()
}

func (t *Torrent) updateRA() {
	// t.muTorrent.Lock()
	// defer t.muTorrent.Unlock()
	// if t.Torrent != nil && t.Torrent.Info() != nil {
	// 	pieceLen := t.Torrent.Info().PieceLength
	// 	adj := pieceLen * int64(t.Torrent.Stats().ActivePeers) / int64(1+t.cache.Readers())
	// 	switch {
	// 	case adj < pieceLen:
	// 		adj = pieceLen
	// 	case adj > pieceLen*4:
	// 		adj = pieceLen * 4
	// 	}
	// 	go t.cache.AdjustRA(adj)
	// }
	if t.cache == nil { // offline torrents have no cache
		return
	}
	adj := int64(16 << 20) // 16 MB fixed RA
	go t.cache.AdjustRA(adj)
}

func (t *Torrent) expired() bool {
	if t.cache == nil {
		return false
	}
	t.muTorrent.Lock()
	stat, expiredTime := t.Stat, t.expiredTime
	t.muTorrent.Unlock()
	return t.cache.Readers() == 0 && expiredTime.Before(time.Now()) && (stat == state.TorrentWorking || stat == state.TorrentClosed)
}

func (t *Torrent) Files() []*torrent.File {
	if t.Torrent != nil && t.Torrent.Info() != nil {
		files := t.Torrent.Files()
		return files
	}
	return nil
}

func (t *Torrent) Hash() metainfo.Hash {
	if t.Torrent != nil {
		return t.Torrent.InfoHash()
	}
	if t.TorrentSpec != nil {
		return t.TorrentSpec.InfoHash
	}
	return [20]byte{}
}

func (t *Torrent) Length() int64 {
	if t.Info() == nil {
		return 0
	}
	return t.Torrent.Length()
}

func (t *Torrent) NewReader(file *torrent.File) *torrstor.Reader {
	if t.Stat == state.TorrentClosed || t.cache == nil {
		return nil
	}
	reader := t.cache.NewReader(file)
	return reader
}

func (t *Torrent) CloseReader(reader *torrstor.Reader) {
	if t.cache == nil {
		return
	}
	t.cache.CloseReader(reader)
	t.AddExpiredTime(time.Second * time.Duration(settings.BTsets.TorrentDisconnectTimeout))
}

func (t *Torrent) GetCache() *torrstor.Cache {
	return t.cache
}

func (t *Torrent) drop() {
	t.muTorrent.Lock()
	defer t.muTorrent.Unlock()
	if t.Torrent != nil {
		t.Torrent.Drop()
		t.Torrent = nil
	}
}

func (t *Torrent) Close() bool {
	if t == nil {
		return false
	}
	t.muTorrent.Lock()
	if t.Stat == state.TorrentClosed {
		t.muTorrent.Unlock()
		return true
	}
	if settings.ReadOnly && t.cache != nil && t.cache.GetUseReaders() > 0 {
		t.muTorrent.Unlock()
		return false
	}
	t.Stat = state.TorrentClosed
	t.muTorrent.Unlock()

	if t.bt != nil {
		t.bt.mu.Lock()
		if _, ok := t.bt.torrents[t.Hash()]; ok {
			delete(t.bt.torrents, t.Hash())
		}
		t.bt.mu.Unlock()
	}

	t.drop()
	return true
}

func (t *Torrent) Status() *state.TorrentStatus {
	t.muTorrent.Lock()

	st := new(state.TorrentStatus)

	st.Stat = t.Stat
	st.StatString = t.Stat.String()
	st.OfflineState = t.OfflineState()
	st.OfflinePinned = t.OfflinePinned()
	st.Title = t.Title
	st.Category = t.Category
	st.Poster = t.Poster
	st.Data = t.Data
	st.Timestamp = t.Timestamp
	st.TorrentSize = t.Size
	st.BitRate = t.BitRate
	st.DurationSeconds = t.DurationSeconds

	if t.TorrentSpec != nil {
		st.Hash = t.TorrentSpec.InfoHash.HexString()
	}
	if t.Torrent != nil {
		st.Name = t.Torrent.Name()
		st.Hash = t.Torrent.InfoHash().HexString()
		st.LoadedSize = t.Torrent.BytesCompleted()
		if t.offStore != nil && t.OfflineState() == state.OfflineDownloading {
			// Only verified pieces survive a restart; the received chunks of unfinished pieces do not.
			if _, _, b, ok := t.offStore.Stats(t.Torrent.InfoHash()); ok {
				st.LoadedSize = b
			}
		}

		st.PreloadedBytes = t.PreloadedBytes
		st.PreloadSize = t.PreloadSize
		st.DownloadSpeed = t.DownloadSpeed
		st.UploadSpeed = t.UploadSpeed

		tst := t.Torrent.Stats()
		st.BytesWritten = tst.BytesWritten.Int64()
		st.BytesWrittenData = tst.BytesWrittenData.Int64()
		st.BytesRead = tst.BytesRead.Int64()
		st.BytesReadData = tst.BytesReadData.Int64()
		st.BytesReadUsefulData = tst.BytesReadUsefulData.Int64()
		st.ChunksWritten = tst.ChunksWritten.Int64()
		st.ChunksRead = tst.ChunksRead.Int64()
		st.ChunksReadUseful = tst.ChunksReadUseful.Int64()
		st.ChunksReadWasted = tst.ChunksReadWasted.Int64()
		st.PiecesDirtiedGood = tst.PiecesDirtiedGood.Int64()
		st.PiecesDirtiedBad = tst.PiecesDirtiedBad.Int64()
		st.TotalPeers = tst.TotalPeers
		st.PendingPeers = tst.PendingPeers
		st.ActivePeers = tst.ActivePeers
		st.ConnectedSeeders = tst.ConnectedSeeders
		st.HalfOpenPeers = tst.HalfOpenPeers

		if t.Torrent.Info() != nil {
			st.TorrentSize = t.Torrent.Length()

			// Files returns the torrent's own slice, which also backs each piece's file list,
			// so sort a copy: sorting in place reorders it under concurrent readers.
			files := append([]*torrent.File(nil), t.Files()...)
			sort.Slice(files, func(i, j int) bool {
				return utils2.CompareStrings(files[i].Path(), files[j].Path())
			})
			for i, f := range files {
				st.FileStats = append(st.FileStats, &state.TorrentFileStat{
					Id:     i + 1, // in web id 0 is undefined
					Path:   f.Path(),
					Length: f.Length(),
				})
			}

			th := torrshash.New(st.Hash)
			th.AddField(torrshash.TagTitle, st.Title)
			th.AddField(torrshash.TagPoster, st.Poster)
			th.AddField(torrshash.TagCategory, st.Category)
			th.AddField(torrshash.TagSize, strconv.FormatInt(st.TorrentSize, 10))

			if t.TorrentSpec != nil {
				if len(t.TorrentSpec.Trackers) > 0 && len(t.TorrentSpec.Trackers[0]) > 0 {
					for _, tr := range t.TorrentSpec.Trackers[0] {
						th.AddField(torrshash.TagTracker, tr)
					}
				}
			}
			token, err := torrshash.Pack(th)
			if err == nil {
				st.TorrsHash = token
			}
		}
	}

	// The offline branches read the sidecar from disk and parse the metadata.
	// The web UI polls the status every second, so this I/O must not happen
	// under muTorrent, where it blocked every other reader of the torrent
	// (review remark N-3): the conditions are snapshotted under the lock and
	// the offline status is filled in without it.
	noClient := t.Torrent == nil
	offState := t.OfflineState()
	t.muTorrent.Unlock()

	if offState != "" {
		st.OfflineError = offlineErrorOf(t.Hash())
	}

	if noClient && (offState == state.OfflineReady || offState == state.OfflinePaused) {
		t.fillOfflineStatus(st)
	} else if offState == state.OfflinePaused {
		// a paused offline torrent that is loaded for online use: show what is saved on disk
		if lay, err := t.offlineLayout(); err == nil {
			st.LoadedSize = t.pausedVerified(lay)
		}
	}

	return st
}

func (t *Torrent) CacheState() *cacheSt.CacheState {
	if t.Torrent != nil && t.cache != nil {
		st := t.cache.GetState()
		st.Torrent = t.Status()
		return st
	}
	return nil
}
