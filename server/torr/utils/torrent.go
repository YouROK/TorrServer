package utils

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"io"
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

// defaultTrackersListURLs is the built-in mirror chain. Tests stub this to
// avoid hitting the network when a custom URL fails or is empty.
var defaultTrackersListURLs = append([]string(nil), settings.DefaultTrackersListURLs...)

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

func configuredTrackersListURLs() []string {
	var custom string
	if settings.BTsets != nil {
		custom = strings.TrimSpace(settings.BTsets.TrackersListURL)
	}

	seen := make(map[string]struct{}, len(defaultTrackersListURLs)+1)
	var urls []string
	add := func(u string) {
		if u == "" {
			return
		}
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		urls = append(urls, u)
	}
	add(custom)
	for _, u := range defaultTrackersListURLs {
		add(u)
	}
	return urls
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
	urls := configuredTrackersListURLs()
	if len(urls) == 0 {
		setLoadedTrackers(gen, local, "")
		return
	}

	go fetchTrackersAsync(gen, urls, local)
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

func fetchTrackersAsync(gen uint64, urls []string, local []string) {
	merged, usedURL, err := fetchTrackersFromURLs(urls, local)
	if err != nil {
		setLoadedTrackers(gen, local, "trackerslist fetch failed, using DefaultTrackers: "+err.Error())
		return
	}
	remoteCount := len(merged) - len(local)
	setLoadedTrackers(gen, merged, fmt.Sprintf("trackerslist loaded from %s: %d remote + %d local", usedURL, remoteCount, len(local)))
}

func fetchTrackersFromURLs(urls []string, local []string) ([]string, string, error) {
	if len(urls) == 0 {
		return nil, "", fmt.Errorf("no trackers list URLs")
	}
	var errs []string
	for _, url := range urls {
		merged, err := fetchTrackersFromURL(url, local)
		if err == nil {
			return merged, url, nil
		}
		log.TLogln("trackerslist fetch failed (" + url + "): " + err.Error())
		errs = append(errs, url+": "+err.Error())
	}
	return nil, "", fmt.Errorf("all URLs failed: %s", strings.Join(errs, "; "))
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
		urls := configuredTrackersListURLs()
		if len(urls) == 0 {
			continue
		}
		gen := trackersFetchGen.Load()
		local := configuredDefaultTrackers()
		merged, usedURL, err := fetchTrackersFromURLs(urls, local)
		if err != nil {
			log.TLogln("trackerslist refresh failed:", err.Error())
			continue
		}
		if !updateLoadedTrackersIfCurrent(gen, merged) {
			continue
		}
		remoteCount := len(merged) - len(local)
		log.TLogln(fmt.Sprintf("trackerslist refreshed from %s: %d remote + %d local", usedURL, remoteCount, len(local)))
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
