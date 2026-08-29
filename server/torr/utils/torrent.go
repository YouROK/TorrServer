package utils

import (
	"encoding/base32"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"server/log"
	"server/settings"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"golang.org/x/time/rate"
)

const trackersFetchTimeout = 5 * time.Second

// trackersRefreshInterval controls how often the remote trackers list is
// re-fetched in the background. On refresh failure the existing cache is kept.
var trackersRefreshInterval = 12 * time.Hour

var (
	// fallbackTrackers is used when BTsets is nil or DefaultTrackers is empty (JNI/tests).
	fallbackTrackers   = parseTrackerLines(settings.DefaultTrackersText)
	fallbackTrackersMu sync.RWMutex

	trackersMu         sync.RWMutex
	loadedTrackers     []string // nil until first prefetch completes; GetDefTrackers uses local meanwhile
	trackersFetchGen   atomic.Uint64
	prefetchMu         sync.Mutex
	prefetchStartedGen uint64 = ^uint64(0)
	refreshLoopOnce    sync.Once
)

func SetDefTrackers(trackers []string) {
	fallbackTrackersMu.Lock()
	fallbackTrackers = append([]string(nil), trackers...)
	fallbackTrackersMu.Unlock()

	if settings.BTsets != nil {
		settings.BTsets.DefaultTrackers = strings.Join(trackers, "\n")
	}
	InvalidateTrackersCache()
}

func GetTrackerFromFile() []string {
	name := filepath.Join(settings.Path, "trackers.txt")
	buf, err := os.ReadFile(name)
	if err == nil {
		list := strings.Split(string(buf), "\n")
		var ret []string
		for _, l := range list {
			l = strings.TrimSpace(l)
			if strings.HasPrefix(l, "udp") || strings.HasPrefix(l, "http") {
				ret = append(ret, l)
			}
		}
		return ret
	}
	return nil
}

// GetDefTrackers returns cached trackers when available, otherwise local
// DefaultTrackers immediately without blocking on network I/O.
func GetDefTrackers() []string {
	trackersMu.RLock()
	if loadedTrackers != nil {
		out := make([]string, len(loadedTrackers))
		copy(out, loadedTrackers)
		trackersMu.RUnlock()
		return out
	}
	trackersMu.RUnlock()
	return configuredDefaultTrackers()
}

// PrefetchTrackers loads the remote trackers list in the background when
// configured. Safe to call multiple times; duplicate fetches for the same
// generation are deduplicated. Also starts the periodic refresh loop once.
func PrefetchTrackers() {
	startPrefetch()
	refreshLoopOnce.Do(func() {
		go trackersRefreshLoop()
	})
}

// InvalidateTrackersCache clears the remote list cache and starts a new
// background fetch from current BTSets.
func InvalidateTrackersCache() {
	trackersMu.Lock()
	loadedTrackers = nil
	trackersMu.Unlock()
	trackersFetchGen.Add(1)
	startPrefetch()
}

func parseTrackerLines(text string) []string {
	var ret []string
	for _, s := range strings.Split(text, "\n") {
		s = strings.TrimSpace(s)
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		if strings.HasPrefix(s, "udp") || strings.HasPrefix(s, "http") || strings.HasPrefix(s, "wss") {
			ret = append(ret, s)
		}
	}
	return ret
}

func configuredDefaultTrackers() []string {
	if settings.BTsets != nil && strings.TrimSpace(settings.BTsets.DefaultTrackers) != "" {
		parsed := parseTrackerLines(settings.BTsets.DefaultTrackers)
		if len(parsed) > 0 {
			return parsed
		}
	}
	fallbackTrackersMu.RLock()
	defer fallbackTrackersMu.RUnlock()
	out := make([]string, len(fallbackTrackers))
	copy(out, fallbackTrackers)
	return out
}

func configuredTrackersListURL() string {
	if settings.BTsets != nil {
		return strings.TrimSpace(settings.BTsets.TrackersListURL)
	}
	return settings.DefaultTrackersListURL
}

func startPrefetch() {
	gen := trackersFetchGen.Load()

	prefetchMu.Lock()
	if prefetchStartedGen == gen {
		prefetchMu.Unlock()
		return
	}
	prefetchStartedGen = gen
	prefetchMu.Unlock()

	local := configuredDefaultTrackers()
	url := configuredTrackersListURL()
	if url == "" {
		setLoadedTrackers(gen, local, "")
		return
	}

	go fetchTrackersAsync(gen, url, local)
}

func setLoadedTrackers(gen uint64, trackers []string, logMsg string) {
	if trackersFetchGen.Load() != gen {
		return
	}
	trackersMu.Lock()
	if trackersFetchGen.Load() != gen {
		trackersMu.Unlock()
		return
	}
	loadedTrackers = append([]string(nil), trackers...)
	trackersMu.Unlock()
	if logMsg != "" {
		log.TLogln(logMsg)
	}
}

func fetchTrackersAsync(gen uint64, url string, local []string) {
	merged, err := fetchTrackersFromURL(url, local)
	if err != nil {
		setLoadedTrackers(gen, local, "trackerslist fetch failed, using DefaultTrackers: "+err.Error())
		return
	}
	remoteCount := len(merged) - len(local)
	setLoadedTrackers(gen, merged, fmt.Sprintf("trackerslist loaded: %d remote + %d local", remoteCount, len(local)))
}

func fetchTrackersFromURL(url string, local []string) ([]string, error) {
	client := &http.Client{Timeout: trackersFetchTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var remote []string
	for _, s := range strings.Split(string(buf), "\n") {
		s = strings.TrimSpace(s)
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		remote = append(remote, s)
	}
	if len(remote) == 0 {
		return nil, fmt.Errorf("empty list")
	}
	return append(remote, local...), nil
}

func trackersRefreshLoop() {
	for {
		time.Sleep(trackersRefreshInterval)
		url := configuredTrackersListURL()
		if url == "" {
			continue
		}
		gen := trackersFetchGen.Load()
		local := configuredDefaultTrackers()
		merged, err := fetchTrackersFromURL(url, local)
		if err != nil {
			log.TLogln("trackerslist refresh failed:", err.Error())
			continue
		}
		if !updateLoadedTrackersIfCurrent(gen, merged) {
			continue
		}
		remoteCount := len(merged) - len(local)
		log.TLogln(fmt.Sprintf("trackerslist refreshed: %d remote + %d local", remoteCount, len(local)))
	}
}

func updateLoadedTrackersIfCurrent(gen uint64, trackers []string) bool {
	if trackersFetchGen.Load() != gen {
		return false
	}
	trackersMu.Lock()
	defer trackersMu.Unlock()
	if trackersFetchGen.Load() != gen {
		return false
	}
	loadedTrackers = append([]string(nil), trackers...)
	return true
}

func PeerIDRandom(peer string) string {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return peer + base32.StdEncoding.EncodeToString(randomBytes)[:20-len(peer)]
}

func Limit(i int) *rate.Limiter {
	l := rate.NewLimiter(rate.Inf, 0)
	if i > 0 {
		b := i
		if b < 16*1024 {
			b = 16 * 1024
		}
		l = rate.NewLimiter(rate.Limit(i), b)
	}
	return l
}

func OpenTorrentFile(path string) (*torrent.TorrentSpec, error) {
	minfo, err := metainfo.LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	info, err := minfo.UnmarshalInfo()
	if err != nil {
		return nil, err
	}

	// mag := minfo.Magnet(info.Name, minfo.HashInfoBytes())
	mag := minfo.Magnet(nil, &info)
	return &torrent.TorrentSpec{
		InfoBytes:   minfo.InfoBytes,
		Trackers:    [][]string{mag.Trackers},
		DisplayName: info.Name,
		InfoHash:    minfo.HashInfoBytes(),
	}, nil
}
