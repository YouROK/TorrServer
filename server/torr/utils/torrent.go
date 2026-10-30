package utils

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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

const (
	trackersFetchTimeout = 5 * time.Second
	trackersBodyLimit    = 256 * 1024
)

// trackersRefreshInterval controls how often the remote trackers list is
// re-fetched in the background. On refresh failure the existing cache is kept.
var trackersRefreshInterval = 12 * time.Hour

// Dedicated client: a Client with a nil Transport uses http.DefaultTransport,
// and CloseIdleConnections on that closes every idle connection in the process.
var trackersHTTPClient = &http.Client{
	Timeout: trackersFetchTimeout,
	Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 3 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		TLSHandshakeTimeout:   3 * time.Second,
		ResponseHeaderTimeout: 3 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		MaxIdleConns:          4,
	},
}

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

	lifeMu      sync.Mutex
	lifeCtx     context.Context
	lifeCancel  context.CancelFunc
	fetchMu     sync.Mutex
	fetchCancel context.CancelFunc
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
	buf, err := os.ReadFile(filepath.Join(settings.Path, "trackers.txt"))
	if err != nil {
		return nil
	}
	return SanitizeTrackers(strings.Split(string(buf), "\n"))
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
	startRefreshLoop()
	startPrefetch()
}

// StopTrackers cancels the refresh loop and any in-flight list fetch.
// The cached list is dropped; the next PrefetchTrackers starts a new fetch.
func StopTrackers() {
	lifeMu.Lock()
	cancel := lifeCancel
	lifeCtx, lifeCancel = nil, nil
	lifeMu.Unlock()
	if cancel != nil {
		cancel()
	}
	fetchMu.Lock()
	cancelFetch := fetchCancel
	fetchCancel = nil
	fetchMu.Unlock()
	if cancelFetch != nil {
		cancelFetch()
	}
	trackersFetchGen.Add(1)
	prefetchMu.Lock()
	prefetchStartedGen = ^uint64(0)
	prefetchMu.Unlock()
	trackersMu.Lock()
	loadedTrackers = nil
	trackersMu.Unlock()
	trackersHTTPClient.CloseIdleConnections()
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
	return SanitizeTrackers(strings.Split(text, "\n"))
}

// SanitizeTrackers drops malformed, unsupported, and raw-IP https/wss URLs
// and collapses duplicates. It does not cap the list.
func SanitizeTrackers(lines []string) []string {
	return sanitizeTrackers(lines)
}

func sanitizeTrackers(lines []string) []string {
	seen := make(map[string]struct{}, len(lines))
	var out []string
	for _, line := range lines {
		stored, key, ok := acceptTracker(line)
		if !ok {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, stored)
	}
	return out
}

func acceptTracker(raw string) (stored, key string, ok bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "#") || len(raw) > 2048 {
		return "", "", false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return "", "", false
	}
	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "udp", "http", "https", "ws", "wss":
	default:
		return "", "", false
	}
	if port := u.Port(); port != "" {
		n, convErr := strconv.Atoi(port)
		if convErr != nil || n < 1 || n > 65535 {
			return "", "", false
		}
	}
	host := u.Hostname()
	port := u.Port()
	ip := net.ParseIP(host)
	// Raw-IP https/wss has no usable SNI name. Dial+TLS to a dead address sits
	// for the tracker client's 15s timeout, per torrent.
	if ip != nil && (scheme == "https" || scheme == "wss") {
		return "", "", false
	}
	if ip != nil {
		host = ip.String()
	} else {
		host = strings.ToLower(host)
	}
	u.Scheme = scheme
	u.Host = host
	if port != "" {
		u.Host = net.JoinHostPort(host, port)
	}
	u.Path = strings.TrimRight(u.Path, "/")
	u.Fragment = ""
	stored = u.String()
	if ip != nil {
		// ponytail: raw-IP trackers collapse on host:port. Split by path if one IP hosts two trackers.
		key = "ip:" + host + ":" + u.Port()
		return stored, key, true
	}
	return stored, stored, true
}

// PinUDPFamily rewrites udp://<IPv4> to udp4:// and udp://[IPv6] to udp6://.
// The fork starts a udp4 and a udp6 scraper for every udp:// URL. For a literal
// address the other family can never resolve (getIP skips it), so that scraper
// only fails with "no acceptable ips" every five minutes until the torrent
// closes. Host names stay udp://: the family is unknown without DNS. The input
// slice is not modified.
func PinUDPFamily(urls []string) []string {
	if urls == nil {
		return nil
	}
	out := make([]string, len(urls))
	for i, raw := range urls {
		out[i] = pinUDPFamily(raw)
	}
	return out
}

func pinUDPFamily(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "udp" {
		return raw
	}
	ip := net.ParseIP(u.Hostname())
	if ip == nil {
		return raw
	}
	if ip.To4() != nil {
		u.Scheme = "udp4"
	} else {
		u.Scheme = "udp6"
	}
	return u.String()
}

// mergeTrackers appends every remote URL that is not already in local, then
// the local list. remoteCount is how many remote URLs were kept.
func mergeTrackers(remote, local []string) (merged []string, remoteCount int) {
	local = sanitizeTrackers(local)
	remote = sanitizeTrackers(remote)
	seen := make(map[string]struct{}, len(local)+len(remote))
	for _, t := range local {
		if _, key, ok := acceptTracker(t); ok {
			seen[key] = struct{}{}
		}
	}
	var head []string
	for _, t := range remote {
		_, key, ok := acceptTracker(t)
		if !ok {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		head = append(head, t)
	}
	return append(head, local...), len(head)
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

	go fetchTrackersAsync(trackersContext(), gen, urls, local)
}

func trackersContext() context.Context {
	lifeMu.Lock()
	defer lifeMu.Unlock()
	if lifeCtx == nil {
		return context.Background()
	}
	return lifeCtx
}

func startRefreshLoop() {
	lifeMu.Lock()
	defer lifeMu.Unlock()
	if lifeCtx != nil {
		return
	}
	lifeCtx, lifeCancel = context.WithCancel(context.Background())
	go trackersRefreshLoop(lifeCtx)
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

func fetchTrackersAsync(ctx context.Context, gen uint64, urls []string, local []string) {
	merged, usedURL, remoteCount, err := fetchTrackersFromURLs(ctx, gen, urls, local)
	if err != nil {
		if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			return
		}
		setLoadedTrackers(gen, local, "trackerslist fetch failed, using DefaultTrackers: "+err.Error())
		return
	}
	setLoadedTrackers(gen, merged, fmt.Sprintf("trackerslist loaded from %s: %d remote + %d local", usedURL, remoteCount, len(merged)-remoteCount))
}

func fetchTrackersFromURLs(ctx context.Context, gen uint64, urls []string, local []string) ([]string, string, int, error) {
	if len(urls) == 0 {
		return nil, "", 0, fmt.Errorf("no trackers list URLs")
	}
	if err := ctx.Err(); err != nil {
		return nil, "", 0, err
	}
	ctx, cancel := context.WithCancel(ctx)
	fetchMu.Lock()
	// A superseded generation must not cancel the fetch that replaced it.
	// Doing so dropped loadedTrackers until the next invalidate.
	if err := ctx.Err(); err != nil || trackersFetchGen.Load() != gen {
		fetchMu.Unlock()
		cancel()
		if err != nil {
			return nil, "", 0, err
		}
		return nil, "", 0, context.Canceled
	}
	if fetchCancel != nil {
		fetchCancel()
	}
	fetchCancel = cancel
	fetchMu.Unlock()
	defer cancel()

	var errs []string
	for _, listURL := range urls {
		if ctx.Err() != nil {
			return nil, "", 0, ctx.Err()
		}
		merged, remoteCount, err := fetchTrackersFromURL(ctx, listURL, local)
		if err == nil {
			return merged, listURL, remoteCount, nil
		}
		if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			return nil, "", 0, err
		}
		log.TLogln("trackerslist fetch failed (" + listURL + "): " + err.Error())
		errs = append(errs, listURL+": "+err.Error())
	}
	return nil, "", 0, fmt.Errorf("all URLs failed: %s", strings.Join(errs, "; "))
}

func fetchTrackersFromURL(ctx context.Context, listURL string, local []string) ([]string, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := trackersHTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("status %d", resp.StatusCode)
	}
	buf, err := io.ReadAll(io.LimitReader(resp.Body, trackersBodyLimit))
	if err != nil {
		return nil, 0, err
	}
	remote := sanitizeTrackers(strings.Split(string(buf), "\n"))
	if len(remote) == 0 {
		return nil, 0, fmt.Errorf("empty list")
	}
	merged, remoteCount := mergeTrackers(remote, local)
	return merged, remoteCount, nil
}

func trackersRefreshLoop(ctx context.Context) {
	timer := time.NewTimer(trackersRefreshInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			refreshTrackers()
			timer.Reset(trackersRefreshInterval)
		}
	}
}

// refreshTrackers re-fetches the remote trackers list once and replaces the
// cache on success. On failure, or if the cache was invalidated meanwhile,
// the existing cache is kept.
func refreshTrackers() {
	urls := configuredTrackersListURLs()
	if len(urls) == 0 {
		return
	}
	gen := trackersFetchGen.Load()
	local := configuredDefaultTrackers()
	ctx := trackersContext()
	merged, usedURL, remoteCount, err := fetchTrackersFromURLs(ctx, gen, urls, local)
	if err != nil {
		if ctx.Err() == nil && !errors.Is(err, context.Canceled) {
			log.TLogln("trackerslist refresh failed:", err.Error())
		}
		return
	}
	if !updateLoadedTrackersIfCurrent(gen, merged) {
		return
	}
	log.TLogln(fmt.Sprintf("trackerslist refreshed from %s: %d remote + %d local", usedURL, remoteCount, len(merged)-remoteCount))
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
