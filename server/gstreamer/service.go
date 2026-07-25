//go:build gst

package gstreamer

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"server/settings"
	"server/torr"
	torrstate "server/torr/state"

	"golang.org/x/sync/singleflight"
)

var (
	ErrBadSource               = errors.New("bad gstreamer source")
	ErrUnsupportedContainer    = errors.New("unsupported container; only Matroska/WebM is supported")
	ErrUnsupportedVideo        = errors.New("unsupported video codec")
	ErrUnsupportedHDRTransfer  = errors.New("HDR tone mapping requires a PQ or HLG base layer")
	ErrProbeUnavailable        = errors.New("gst-discoverer returned no stream info")
	ErrPipelineUnavailable     = errors.New("gstreamer runtime is unavailable")
	ErrSegmentNotReady         = errors.New("segment is not ready")
	ErrTaskNotFound            = errors.New("gstreamer task not found")
	ErrServiceClosed           = errors.New("gstreamer service is closed")
	ErrInvalidIdentifier       = errors.New("invalid gstreamer task id")
	ErrEarlyEndOfStream        = errors.New("gstreamer reached EOS before the expected end")
	ErrEndOfStreamExhausted    = errors.New("gstreamer end of stream is exhausted")
	ErrTruncatedMP4Fragment    = errors.New("truncated mp4 fragment at end of stream")
	ErrUndecodableEOSRemainder = errors.New("undecodable mp4 eos remainder")
)

type Service struct {
	conf Config

	mu    sync.RWMutex
	tasks map[string]*Task

	probeMu    sync.Mutex
	probeCache map[string]probeCacheEntry
	probeCalls singleflight.Group
	taskCalls  singleflight.Group

	cleanupRunning atomic.Bool
	disposed       atomic.Bool
	stopCleanup    chan struct{}
}

const probeCacheTTL = time.Hour

type probeCacheEntry struct {
	probe     ProbeInfo
	expiresAt time.Time
}

func (s *Service) currentConfig() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.conf
}

func (s *Service) updateConfig(conf Config) {
	if s.disposed.Load() {
		return
	}
	conf = conf.normalized()

	s.mu.Lock()
	s.conf = conf
	evicted := s.evictTasksForLimitLocked("")
	s.mu.Unlock()

	disposeTasks(evicted)
}

func NewService(conf Config) *Service {
	conf = conf.normalized()

	service := &Service{
		conf:        conf,
		tasks:       make(map[string]*Task),
		probeCache:  make(map[string]probeCacheEntry),
		stopCleanup: make(chan struct{}),
	}
	go service.cleanupLoop()
	return service
}

func (s *Service) GetOrAdd(hash string, fileID string, audio int) (*Task, error) {
	if hash == "" || fileID == "" {
		return nil, ErrBadSource
	}

	for {
		if s.disposed.Load() {
			return nil, ErrServiceClosed
		}
		value, err, _ := s.taskCalls.Do(hash, func() (any, error) {
			return s.getOrAdd(hash, fileID, audio)
		})
		if err != nil {
			return nil, err
		}
		task, ok := value.(*Task)
		if !ok {
			return nil, errors.New("gstreamer task creation returned an invalid result")
		}
		if taskMatchesRequest(task, hash, fileID, audio) {
			return task, nil
		}
	}
}

func (s *Service) getOrAdd(hash string, fileID string, audio int) (*Task, error) {
	if s.disposed.Load() {
		return nil, ErrServiceClosed
	}
	conf := s.currentConfig()
	sourceURL := sourceURL(conf, hash, fileID)
	id := hash

	s.mu.RLock()
	task := s.tasks[id]
	if task != nil && task.FileID == fileID && task.Audio == audio && !task.IsDisposed() {
		task.UpdateLastActive()
		s.mu.RUnlock()
		return task, nil
	}
	s.mu.RUnlock()

	probe, err := s.Probe(hash, fileID)
	if err != nil {
		return nil, err
	}
	var cue *CueTimeline
	if shouldUseCueTimeline(conf, probe) {
		cue = readMatroskaCueTimeline(sourceURL, probe.FileSize, probe.DurationNS)
	}

	task, err = NewTask(id, fileID, audio, sourceURL, probe, cue, conf)
	if err != nil {
		return nil, err
	}

	var replaced *Task
	var evicted []*Task

	s.mu.Lock()
	if s.disposed.Load() {
		s.mu.Unlock()
		task.Dispose()
		return nil, ErrServiceClosed
	}
	existing := s.tasks[id]
	if existing != nil &&
		existing.FileID == fileID &&
		existing.Audio == audio &&
		!existing.IsDisposed() {
		existing.UpdateLastActive()
		s.mu.Unlock()

		task.Dispose()
		return existing, nil
	}

	replaced = existing
	s.tasks[id] = task
	evicted = s.evictTasksForLimitLocked(id)
	s.mu.Unlock()

	if replaced != nil {
		replaced.Dispose()
	}
	disposeTasks(evicted)

	return task, nil
}

func taskMatchesRequest(task *Task, hash string, fileID string, audio int) bool {
	return task != nil && !task.IsDisposed() && task.ID == hash && task.FileID == fileID && task.Audio == audio
}

func shouldUseCueTimeline(conf Config, probe ProbeInfo) bool {
	if !probe.IsMatroskaContainer() {
		return false
	}
	switch {
	case probe.IsH264():
		return !conf.TranscodeH264
	case probe.IsH265():
		return !conf.TranscodeH265
	case probe.IsAV1():
		return !conf.TranscodeAV1
	case probe.IsVP9():
		return !conf.TranscodeVP9
	default:
		return false
	}
}

func (s *Service) evictTasksForLimitLocked(protectedID string) []*Task {
	limit := s.conf.normalized().MaxTasks
	if limit <= 0 || len(s.tasks) <= limit {
		return nil
	}

	evicted := make([]*Task, 0, len(s.tasks)-limit)
	for len(s.tasks) > limit {
		id, task := oldestEvictableTask(s.tasks, protectedID)
		if id == "" {
			break
		}
		delete(s.tasks, id)
		if task != nil {
			evicted = append(evicted, task)
		}
	}
	return evicted
}

func oldestEvictableTask(tasks map[string]*Task, protectedID string) (string, *Task) {
	var oldestID string
	var oldestTask *Task
	var oldestActive time.Time
	oldestDisposed := false

	for id, task := range tasks {
		if id == protectedID {
			continue
		}
		if task == nil {
			return id, nil
		}

		disposed := task.IsDisposed()
		lastActive := task.LastActive()
		if oldestID == "" || (disposed && !oldestDisposed) || (disposed == oldestDisposed && lastActive.Before(oldestActive)) {
			oldestID = id
			oldestTask = task
			oldestActive = lastActive
			oldestDisposed = disposed
		}
	}
	return oldestID, oldestTask
}

func disposeTasks(tasks []*Task) {
	for _, task := range tasks {
		if task != nil {
			task.Dispose()
		}
	}
}

func (s *Service) Probe(hash string, fileID string) (ProbeInfo, error) {
	if hash == "" || fileID == "" {
		return ProbeInfo{}, ErrBadSource
	}
	if s.disposed.Load() {
		return ProbeInfo{}, ErrServiceClosed
	}

	if probe, ok, err := s.cachedProbe(hash, fileID); ok {
		return probe, err
	}

	key := probeCacheKey(hash, fileID)
	value, err, _ := s.probeCalls.Do(key, func() (any, error) {
		if s.disposed.Load() {
			return ProbeInfo{}, ErrServiceClosed
		}
		if cached, found, err := s.cachedProbe(hash, fileID); found {
			return cached, err
		}
		conf := s.currentConfig()
		result, err := probeSource(sourceURL(conf, hash, fileID), conf)
		if err != nil {
			return ProbeInfo{}, err
		}
		result = refreshProbeFileSize(result, hash, fileID)
		if err := validateProbe(result, conf); err != nil {
			return ProbeInfo{}, err
		}
		if s.disposed.Load() {
			return ProbeInfo{}, ErrServiceClosed
		}
		s.setCachedProbe(hash, fileID, result)
		return result, nil
	})
	if err != nil {
		return ProbeInfo{}, err
	}
	if s.disposed.Load() {
		return ProbeInfo{}, ErrServiceClosed
	}
	probe := refreshProbeFileSize(value.(ProbeInfo), hash, fileID)
	if err := validateProbe(probe, s.currentConfig()); err != nil {
		return ProbeInfo{}, err
	}
	s.setCachedProbe(hash, fileID, probe)
	return probe, nil
}

func (s *Service) cachedProbe(hash string, fileID string) (ProbeInfo, bool, error) {
	probe, ok := s.getCachedProbe(hash, fileID)
	if !ok {
		return ProbeInfo{}, false, nil
	}

	probe = refreshProbeFileSize(probe, hash, fileID)
	if err := validateProbe(probe, s.currentConfig()); err != nil {
		return ProbeInfo{}, true, err
	}
	s.setCachedProbe(hash, fileID, probe)
	return probe, true, nil
}

func validateProbe(probe ProbeInfo, conf Config) error {
	if len(probe.Tracks) == 0 || probe.Video() == nil {
		return ErrProbeUnavailable
	}
	if conf.HDRToSDR && probe.Video().IsHDRVideo() && probe.Video().VideoTransfer != "pq" && probe.Video().VideoTransfer != "hlg" {
		return ErrUnsupportedHDRTransfer
	}
	transcodeAVI := probe.IsAVIContainer() && conf.TranscodeAVI
	if !probe.IsMatroskaContainer() && !transcodeAVI {
		name := strings.TrimSpace(probe.Container)
		if name == "" {
			name = "<unknown>"
		}
		return fmt.Errorf("%w: %s", ErrUnsupportedContainer, name)
	}
	supported := probe.IsH264() || probe.IsH265() || probe.IsAV1() || probe.IsVP9() ||
		(probe.IsVP8() && conf.TranscodeVP8) || transcodeAVI
	if !supported {
		return ErrUnsupportedVideo
	}
	return nil
}

func torrentFileSize(hash string, fileID string) (size int64) {
	index, err := strconv.Atoi(fileID)
	if err != nil || index <= 0 {
		return 0
	}

	tor := getTorrentForGStreamer(hash)
	if tor == nil {
		return 0
	}

	if size := torrentStatusFileSize(tor.Status(), index); size > 0 {
		return size
	}
	if tor.Torrent == nil {
		return 0
	}
	if !tor.GotInfo() {
		return 0
	}

	return torrentStatusFileSize(tor.Status(), index)
}

type heartbeatState struct {
	Hash    string                   `json:"Hash"`
	Torrent *torrstate.TorrentStatus `json:"Torrent,omitempty"`
}

func torrentHeartbeatState(hash string) (state any) {
	state = heartbeatState{Hash: hash}

	defer func() {
		if recover() != nil {
			state = heartbeatState{Hash: hash}
		}
	}()

	tor := getTorrentForGStreamer(hash)
	if tor == nil {
		return state
	}

	cacheState := tor.CacheState()
	if cacheState != nil {
		return cacheState
	}

	return heartbeatState{
		Hash:    hash,
		Torrent: tor.Status(),
	}
}

func dropTorrentForGStreamer(hash string) {
	defer func() {
		_ = recover()
	}()

	if hash == "" {
		return
	}
	torr.DropTorrent(hash)
}

func getTorrentForGStreamer(hash string) (tor *torr.Torrent) {
	defer func() {
		if recover() != nil {
			tor = nil
		}
	}()

	if hash == "" {
		return nil
	}
	return torr.GetTorrent(hash)
}

func torrentStatusFileSize(status *torrstate.TorrentStatus, index int) int64 {
	if status == nil {
		return 0
	}
	for _, file := range status.FileStats {
		if file != nil && file.Id == index && file.Length > 0 {
			return file.Length
		}
	}
	return 0
}

func refreshProbeFileSize(probe ProbeInfo, hash string, fileID string) ProbeInfo {
	if size := torrentFileSize(hash, fileID); size > 0 {
		probe.FileSize = size
	}
	return probe
}

func (s *Service) getCachedProbe(hash string, fileID string) (ProbeInfo, bool) {
	key := probeCacheKey(hash, fileID)
	now := time.Now().UTC()

	s.probeMu.Lock()
	defer s.probeMu.Unlock()

	entry, ok := s.probeCache[key]
	if !ok {
		return ProbeInfo{}, false
	}
	if !now.Before(entry.expiresAt) {
		delete(s.probeCache, key)
		return ProbeInfo{}, false
	}
	return cloneProbeInfo(entry.probe), true
}

func (s *Service) setCachedProbe(hash string, fileID string, probe ProbeInfo) {
	if s.disposed.Load() {
		return
	}
	key := probeCacheKey(hash, fileID)

	s.probeMu.Lock()
	defer s.probeMu.Unlock()
	if s.disposed.Load() {
		return
	}

	if s.probeCache == nil {
		s.probeCache = make(map[string]probeCacheEntry)
	}
	s.probeCache[key] = probeCacheEntry{
		probe:     cloneProbeInfo(probe),
		expiresAt: time.Now().UTC().Add(probeCacheTTL),
	}
}

func (s *Service) cleanupProbeCache(now time.Time) {
	s.probeMu.Lock()
	defer s.probeMu.Unlock()

	for key, entry := range s.probeCache {
		if !now.Before(entry.expiresAt) {
			delete(s.probeCache, key)
		}
	}
}

func probeCacheKey(hash string, fileID string) string {
	return hash + "\x00" + fileID
}

func cloneProbeInfo(probe ProbeInfo) ProbeInfo {
	if len(probe.Tracks) == 0 {
		return probe
	}
	probe.Tracks = append([]TrackInfo(nil), probe.Tracks...)
	return probe
}

func (s *Service) Get(id string) *Task {
	if id == "" || s.disposed.Load() {
		return nil
	}

	s.mu.RLock()
	task := s.tasks[id]
	if task == nil || task.IsDisposed() {
		s.mu.RUnlock()
		return nil
	}
	task.UpdateLastActive()
	s.mu.RUnlock()
	return task
}

func (s *Service) TryRemove(id string) bool {
	task, ok := s.detachTask(id, nil)
	if !ok {
		return false
	}

	task.Dispose()
	return true
}

func (s *Service) detachTask(id string, expected *Task) (*Task, bool) {
	if id == "" {
		return nil, false
	}

	s.mu.Lock()
	task := s.tasks[id]
	if task == nil || (expected != nil && task != expected) {
		s.mu.Unlock()
		return nil, false
	}

	delete(s.tasks, id)
	s.mu.Unlock()
	return task, true
}

func (s *Service) tryRemoveExpectedInactive(id string, expected *Task, cutoff time.Time) bool {
	if id == "" || expected == nil {
		return false
	}

	s.mu.Lock()
	task := s.tasks[id]
	if task != expected || !task.LastActive().Before(cutoff) {
		s.mu.Unlock()
		return false
	}
	delete(s.tasks, id)
	s.mu.Unlock()

	task.Dispose()
	return true
}

func (s *Service) Dispose() {
	if !s.disposed.CompareAndSwap(false, true) {
		return
	}
	if s.stopCleanup != nil {
		close(s.stopCleanup)
	}

	s.mu.Lock()
	tasks := s.tasks
	s.tasks = make(map[string]*Task)
	s.mu.Unlock()

	s.probeMu.Lock()
	s.probeCache = make(map[string]probeCacheEntry)
	s.probeMu.Unlock()

	for _, task := range tasks {
		task.Dispose()
	}
}

func (s *Service) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						gstErrorf("inactive cleanup panic: %v", recovered)
					}
				}()
				s.cleanupInactive()
			}()
		case <-s.stopCleanup:
			return
		}
	}
}

func (s *Service) cleanupInactive() {
	if s.disposed.Load() {
		return
	}
	if !s.cleanupRunning.CompareAndSwap(false, true) {
		return
	}
	defer s.cleanupRunning.Store(false)

	now := time.Now().UTC()

	type snapshotEntry struct {
		id   string
		task *Task
	}
	s.mu.RLock()
	snapshot := make([]snapshotEntry, 0, len(s.tasks))
	for id, task := range s.tasks {
		snapshot = append(snapshot, snapshotEntry{id: id, task: task})
	}
	s.mu.RUnlock()

	conf := s.currentConfig()
	inactiveDuration := conf.inactiveDuration()
	removeAfter := inactiveDuration + 20*time.Minute
	freezeCutoff := now.Add(-inactiveDuration)
	removeCutoff := now.Add(-removeAfter)

	for _, entry := range snapshot {
		id, task := entry.id, entry.task
		lastActive := task.LastActive()
		if lastActive.Before(removeCutoff) {
			s.tryRemoveExpectedInactive(id, task, removeCutoff)
			continue
		}
		if lastActive.Before(freezeCutoff) && s.isCurrentTask(id, task) {
			task.FreezeIfInactive(freezeCutoff)
		}
	}

	s.cleanupProbeCache(now)
}

func (s *Service) isCurrentTask(id string, expected *Task) bool {
	s.mu.RLock()
	current := s.tasks[id]
	s.mu.RUnlock()
	return current == expected
}

func sourceURL(conf Config, hash string, fileID string) string {
	if conf.normalized().Source == "play" {
		return playURL(hash, fileID)
	}
	return streamURL(hash, fileID)
}

func streamURL(hash string, fileID string) string {
	return "http://127.0.0.1:" + settings.Port + "/stream/?link=" + url.QueryEscape(hash) + "&index=" + url.QueryEscape(fileID) + "&play"
}

func playURL(hash string, fileID string) string {
	return "http://127.0.0.1:" + settings.Port + "/play/" + url.PathEscape(hash) + "/" + url.PathEscape(fileID)
}
