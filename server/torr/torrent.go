package torr

import (
	"errors"
	"path/filepath"
	"regexp"
	"server/torrshash"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	utils2 "server/utils"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/log"
	"server/settings"
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

	// Auto-delete after 3 hours
	createdAt time.Time
}

func NewTorrent(spec *torrent.TorrentSpec, bt *BTServer) (*Torrent, error) {
	// https://github.com/anacrolix/torrent/issues/747
	if bt == nil || bt.client == nil {
		return nil, errors.New("BT client not connected")
	}

	// Auto-save trackers from the source (magnet or .torrent file)
	utils.SaveUniqueTrackers(spec.Trackers)

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

	goTorrent, _, err := bt.client.AddTorrentSpec(spec)
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
	torr.createdAt = time.Now()

	go torr.watch()

	bt.torrents[spec.InfoHash] = torr
	return torr, nil
}

func (t *Torrent) WaitInfo() bool {
	if t == nil || t.Torrent == nil {
		return false
	}

	// Close torrent if no info in 1 minute + TorrentDisconnectTimeout config option
	tm := time.NewTimer(time.Minute + time.Second*time.Duration(settings.BTsets.TorrentDisconnectTimeout))

	select {
	case <-t.Torrent.GotInfo():
		if t.bt != nil && t.bt.storage != nil {
			t.cache = t.bt.storage.GetCache(t.Hash())
			t.cache.SetTorrent(t.Torrent)
		}
		return true
	case <-t.closed:
		return false
	case <-tm.C:
		return false
	}
}

func (t *Torrent) GotInfo() bool {
	// log.TLogln("GotInfo state:", t.Stat)
	if t == nil || t.Stat == state.TorrentClosed {
		return false
	}
	// assume we have info in preload state
	// and dont override with TorrentWorking
	if t.Stat == state.TorrentPreload {
		return true
	}
	t.Stat = state.TorrentGettingInfo
	if t.WaitInfo() {
		t.Stat = state.TorrentWorking
		t.AddExpiredTime(time.Second * time.Duration(settings.BTsets.TorrentDisconnectTimeout))
		return true
	} else {
		t.Close()
		return false
	}
}

func (t *Torrent) AddExpiredTime(duration time.Duration) {
	newExpiredTime := time.Now().Add(duration)
	if t.expiredTime.Before(newExpiredTime) {
		t.expiredTime = newExpiredTime
	}
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
	adj := int64(16 << 20) // 16 MB fixed RA
	go t.cache.AdjustRA(adj)
}

func (t *Torrent) expired() bool {
	// Auto-delete after 3 hours regardless of activity
	// Check if createdAt is not zero value (was initialized)
	if !t.createdAt.IsZero() && time.Since(t.createdAt) > 3*time.Hour {
		return true
	}
	// Check cache is not nil before accessing Readers()
	if t.cache == nil {
		return false
	}
	return t.cache.Readers() == 0 && t.expiredTime.Before(time.Now()) && (t.Stat == state.TorrentWorking || t.Stat == state.TorrentClosed)
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
	if t.Stat == state.TorrentClosed {
		return nil
	}
	reader := t.cache.NewReader(file)
	return reader
}

func (t *Torrent) CloseReader(reader *torrstor.Reader) {
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
	if settings.ReadOnly && t.cache != nil && t.cache.GetUseReaders() > 0 {
		return false
	}
	t.Stat = state.TorrentClosed

	if t.bt != nil {
		t.bt.mu.Lock()
		delete(t.bt.torrents, t.Hash())
		t.bt.mu.Unlock()
	}

	t.drop()
	return true
}

func (t *Torrent) Status() *state.TorrentStatus {
	t.muTorrent.Lock()
	defer t.muTorrent.Unlock()

	st := new(state.TorrentStatus)

	st.Stat = t.Stat
	st.StatString = t.Stat.String()
	st.Title = t.Title
	st.Category = t.Category
	st.Poster = t.Poster
	st.Data = t.Data
	st.Timestamp = t.Timestamp
	st.TorrentSize = t.Size
	st.BitRate = t.BitRate
	st.DurationSeconds = t.DurationSeconds

	// Set CreatedAt and calculate time until auto-delete (3 hours)
	if !t.createdAt.IsZero() {
		st.CreatedAt = t.createdAt.Unix()
		elapsed := time.Since(t.createdAt)
		threeHours := 3 * time.Hour
		if elapsed < threeHours {
			st.TimeUntilDelete = int64((threeHours - elapsed).Seconds())
		} else {
			st.TimeUntilDelete = 0 // Already expired
		}
	}

	if t.TorrentSpec != nil {
		st.Hash = t.TorrentSpec.InfoHash.HexString()
	}
	if t.Torrent != nil {
		st.Name = t.Torrent.Name()
		st.Hash = t.Torrent.InfoHash().HexString()
		st.LoadedSize = t.Torrent.BytesCompleted()

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

			files := t.Files()
			filesCopy := make([]*torrent.File, len(files))
			copy(filesCopy, files)
			files = filesCopy

			sort.Slice(files, func(i, j int) bool {
				return utils2.CompareStrings(files[i].Path(), files[j].Path())
			})

			// Check for Smart Indexing (Movie Mode)
			// Trigger if Category is explicitly "movie"
			// OR if the largest file dominates the torrent (> 85% of total size), implying it's the main movie file.
			if len(files) > 1 {
				totalSize := int64(0)
				maxSize := int64(0)
				maxIndex := -1

				for i, f := range files {
					fLen := f.Length()
					totalSize += fLen
					if fLen > maxSize {
						maxSize = fLen
						maxIndex = i
					}
				}

				isMovie := st.Category == "movie"
				if !isMovie && totalSize > 0 {
					// Heuristic: If largest file is > 85% of total size
					if float64(maxSize) > float64(totalSize)*0.85 {
						isMovie = true
					}
				}

				if isMovie && maxIndex > 0 {
					largest := files[maxIndex]
					// remove from current position
					files = append(files[:maxIndex], files[maxIndex+1:]...)
					// prepend
					files = append([]*torrent.File{largest}, files...)
				}
			}

			// Collect unique file extensions
			extensionsMap := make(map[string]bool)

			addExt := func(path string) {
				ext := filepath.Ext(path)
				if len(ext) > 0 {
					ext = strings.ToUpper(ext[1:])
					if ext != "" {
						extensionsMap[ext] = true
					}
				}
			}

			// Smart Indexing (TV Series Mode)
			// Exclude "anime" to keep default indexing as requested
			catLower := strings.ToLower(st.Category)
			if strings.Contains(catLower, "tv") && !strings.Contains(catLower, "anime") {
				// Priority 1: "Season...Episode" structure (Folder/File) - e.g. "Season 12/Episode 01.mkv"
				reSeasonEp := regexp.MustCompile(`(?i)Season\W*(\d+).*\WEpisode\W*(\d+)`)

				// Priority 2: Standard S...E... (e.g. S12E01)
				reSE := regexp.MustCompile(`(?i)\bS(\d+)(?:[^0-9E]+)?E(\d+)\b`)

				// Priority 3: X notation (e.g. 12x01)
				reX := regexp.MustCompile(`(?i)\b(\d+)x(\d+)\b`)

				parseID := func(s string) int {
					// Try Season/Episode full words first (often in folder names)
					matches := reSeasonEp.FindStringSubmatch(s)
					if len(matches) == 3 {
						season, _ := strconv.Atoi(matches[1])
						episode, _ := strconv.Atoi(matches[2])
						if season > 0 && episode > 0 {
							return season*100 + episode
						}
					}

					// Try S...E...
					matches = reSE.FindStringSubmatch(s)
					if len(matches) == 3 {
						season, _ := strconv.Atoi(matches[1])
						episode, _ := strconv.Atoi(matches[2])
						if season > 0 && episode > 0 {
							return season*100 + episode
						}
					}

					// Try 12x01
					matches = reX.FindStringSubmatch(s)
					if len(matches) == 3 {
						season, _ := strconv.Atoi(matches[1])
						episode, _ := strconv.Atoi(matches[2])
						if season > 0 && episode > 0 {
							return season*100 + episode
						}
					}

					return 0
				}

				// Single File
				if len(files) == 1 {
					f := files[0]
					id := parseID(f.Path())
					if id == 0 {
						id = parseID(t.Title)
					}
					if id == 0 {
						id = parseID(st.Category)
					}

					if id > 0 {
						st.FileStats = append(st.FileStats, &state.TorrentFileStat{
							Id:     id,
							Path:   f.Path(),
							Length: f.Length(),
						})
						addExt(f.Path())
						goto FinishStatus
					}
				} else {
					// Multiple Files
					customIDs := make(map[int]*torrent.File)
					usedIndices := make(map[int]bool)

					for i, f := range files {
						id := parseID(f.Path())
						if id > 0 {
							customIDs[id] = f
							usedIndices[i] = true
						}
					}

					if len(customIDs) > 0 {
						st.FileStats = make([]*state.TorrentFileStat, 0, len(files))

						var sortedIDs []int
						for id := range customIDs {
							sortedIDs = append(sortedIDs, id)
						}
						sort.Ints(sortedIDs)

						for _, id := range sortedIDs {
							f := customIDs[id]
							st.FileStats = append(st.FileStats, &state.TorrentFileStat{
								Id:     id,
								Path:   f.Path(),
								Length: f.Length(),
							})
							addExt(f.Path())
						}

						defaultID := 10000
						for i, f := range files {
							if !usedIndices[i] {
								st.FileStats = append(st.FileStats, &state.TorrentFileStat{
									Id:     defaultID,
									Path:   f.Path(),
									Length: f.Length(),
								})
								defaultID++
								addExt(f.Path())
							}
						}

						// Sort by ID
						sort.Slice(st.FileStats, func(i, j int) bool {
							return st.FileStats[i].Id < st.FileStats[j].Id
						})
						goto FinishStatus
					}
				}
			}

			// Default Logic (Mixed with Extension Collection)
			for i, f := range files {
				st.FileStats = append(st.FileStats, &state.TorrentFileStat{
					Id:     i + 1, // in web id 0 is undefined
					Path:   f.Path(),
					Length: f.Length(),
				})
				addExt(f.Path())
			}
		FinishStatus:

			// Convert map to sorted slice
			for ext := range extensionsMap {
				st.FileExtensions = append(st.FileExtensions, ext)
			}
			sort.Strings(st.FileExtensions)

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

	return st
}

func (t *Torrent) StatusLight() *state.TorrentStatus {
	t.muTorrent.Lock()
	defer t.muTorrent.Unlock()

	st := new(state.TorrentStatus)

	st.Stat = t.Stat
	st.StatString = t.Stat.String()
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
		st.DownloadSpeed = t.DownloadSpeed
		st.UploadSpeed = t.UploadSpeed

		tst := t.Torrent.Stats()
		st.TotalPeers = tst.TotalPeers
		st.ActivePeers = tst.ActivePeers
		st.ConnectedSeeders = tst.ConnectedSeeders

		if t.Torrent.Info() != nil {
			st.TorrentSize = t.Torrent.Length()
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
	// Return status even without cache (e.g. during GotInfo phase)
	// so frontend can get file_stats as soon as info is available
	st := &cacheSt.CacheState{}
	st.Torrent = t.Status()
	return st
}
