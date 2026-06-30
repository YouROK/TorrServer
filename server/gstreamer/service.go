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
)

var (
	ErrBadSource               = errors.New("bad gstreamer source")
	ErrUnsupportedContainer    = errors.New("unsupported container; only Matroska/WebM is supported")
	ErrUnsupportedVideo        = errors.New("unsupported video codec")
	ErrProbeUnavailable        = errors.New("gst-discoverer returned no stream info")
	ErrPipelineDisabled        = errors.New("gstreamer support is not built in")
	ErrSegmentNotReady         = errors.New("segment is not ready")
	ErrTaskNotFound            = errors.New("gstreamer task not found")
	ErrInvalidIdentifier       = errors.New("invalid gstreamer task id")
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

	cleanupRunning atomic.Bool
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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conf = conf.normalized()
}

func NewService(conf Config) *Service {
	conf = conf.normalized()

	service := &Service{
		conf:        conf,
		tasks:       make(map[string]*Task),
		probeCache:  make(map[string]probeCacheEntry),
		stopCleanup: make(chan struct{}),
	}
	cleanupGSTTempFiles()
	go service.cleanupLoop()
	return service
}

func (s *Service) GetOrAdd(hash string, fileID string, audio int) (*Task, error) {
	if hash == "" || fileID == "" {
		return nil, ErrBadSource
	}

	sourceURL := sourceURL(s.conf, hash, fileID)
	id := hash

	s.mu.RLock()
	task := s.tasks[id]
	s.mu.RUnlock()

	if task != nil && task.FileID == fileID && task.Audio == audio && !task.IsDisposed() {
		task.UpdateLastActive()
		return task, nil
	}

	probe, err := s.Probe(hash, fileID)
	if err != nil {
		return nil, err
	}

	task, err = NewTask(id, hash, fileID, audio, sourceURL, probe, s.conf)
	if err != nil {
		return nil, err
	}

	var replaced *Task

	s.mu.Lock()
	existing := s.tasks[id]
	if existing != nil &&
		existing.FileID == fileID &&
		existing.Audio == audio &&
		!existing.IsDisposed() {
		s.mu.Unlock()

		task.Dispose()
		existing.UpdateLastActive()
		return existing, nil
	}

	replaced = existing
	s.tasks[id] = task
	s.mu.Unlock()

	if replaced != nil {
		replaced.Dispose()
	}

	return task, nil
}

func (s *Service) Probe(hash string, fileID string) (ProbeInfo, error) {
	if hash == "" || fileID == "" {
		return ProbeInfo{}, ErrBadSource
	}

	probe, ok := s.getCachedProbe(hash, fileID)
	if !ok {
		var err error
		probe, err = probeSource(sourceURL(s.conf, hash, fileID), s.conf)
		if err != nil {
			return ProbeInfo{}, err
		}
		probe = refreshProbeFileSize(probe, hash, fileID)
		if err := validateProbe(probe); err != nil {
			return ProbeInfo{}, err
		}
		s.setCachedProbe(hash, fileID, probe)
		return probe, nil
	}

	probe = refreshProbeFileSize(probe, hash, fileID)
	s.setCachedProbe(hash, fileID, probe)
	return probe, nil
}

func validateProbe(probe ProbeInfo) error {
	if len(probe.Tracks) == 0 || probe.Video() == nil {
		return ErrProbeUnavailable
	}
	if !probe.IsMatroskaContainer() {
		name := strings.TrimSpace(probe.Container)
		if name == "" {
			name = "<unknown>"
		}
		return fmt.Errorf("%w: %s", ErrUnsupportedContainer, name)
	}
	if !probe.IsH264() && !probe.IsH265() && !probe.IsAV1() && !probe.IsVP9() {
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
	key := probeCacheKey(hash, fileID)

	s.probeMu.Lock()
	defer s.probeMu.Unlock()

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
	if id == "" {
		return nil
	}

	s.mu.RLock()
	task := s.tasks[id]
	s.mu.RUnlock()

	if task == nil {
		return nil
	}
	if task.IsDisposed() {
		return nil
	}

	task.UpdateLastActive()
	return task
}

func (s *Service) TryRemove(id string) bool {
	task, ok := s.detachTask(id, nil)
	if !ok {
		return false
	}

	task.Dispose()
	if s.isEmpty() {
		cleanupGSTTempFiles()
	}
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

func (s *Service) tryRemoveExpected(id string, expected *Task) bool {
	task, ok := s.detachTask(id, expected)
	if !ok {
		return false
	}

	task.Dispose()
	if s.isEmpty() {
		cleanupGSTTempFiles()
	}
	return true
}

func (s *Service) Dispose() {
	closeOnce(s.stopCleanup)

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
					_ = recover()
				}()
				s.cleanupInactive()
			}()
		case <-s.stopCleanup:
			return
		}
	}
}

func (s *Service) cleanupInactive() {
	if !s.cleanupRunning.CompareAndSwap(false, true) {
		return
	}
	defer s.cleanupRunning.Store(false)

	now := time.Now().UTC()

	s.mu.RLock()
	snapshot := make(map[string]*Task, len(s.tasks))
	for id, task := range s.tasks {
		snapshot[id] = task
	}
	s.mu.RUnlock()

	inactiveDuration := s.conf.inactiveDuration()
	removeAfter := inactiveDuration + 20*time.Minute

	for id, task := range snapshot {
		lastActive := task.LastActive()
		if now.After(lastActive.Add(removeAfter)) {
			s.tryRemoveExpected(id, task)
			continue
		}
		if !task.IsFrozen() &&
			now.After(lastActive.Add(inactiveDuration)) &&
			s.isCurrentTask(id, task) {
			task.Frozen()
		}
	}

	if s.isEmpty() {
		cleanupGSTTempFiles()
	}
	s.cleanupProbeCache(now)
}

func (s *Service) isEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tasks) == 0
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

func closeOnce(ch chan struct{}) {
	defer func() { _ = recover() }()
	close(ch)
}
