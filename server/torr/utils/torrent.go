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
	"time"

	"server/log"
	"server/settings"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"golang.org/x/time/rate"
)

const trackersFetchTimeout = 5 * time.Second

var (
	// fallbackTrackers is used when BTsets is nil or DefaultTrackers is empty (JNI/tests).
	fallbackTrackers = parseTrackerLines(settings.DefaultTrackersText)
	fallbackTrackersMu sync.RWMutex

	loadedTrackers   []string
	loadTrackersOnce sync.Once
	loadedTrackersMu sync.Mutex
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

func GetDefTrackers() []string {
	loadNewTracker()
	loadedTrackersMu.Lock()
	defer loadedTrackersMu.Unlock()
	out := make([]string, len(loadedTrackers))
	copy(out, loadedTrackers)
	return out
}

// InvalidateTrackersCache clears the one-shot remote list cache so the next
// GetDefTrackers() reloads from current BTSets.
func InvalidateTrackersCache() {
	loadedTrackersMu.Lock()
	defer loadedTrackersMu.Unlock()
	loadTrackersOnce = sync.Once{}
	loadedTrackers = nil
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

func useDefaultTrackersFallback(reason string) {
	loadedTrackers = configuredDefaultTrackers()
	if reason != "" {
		log.TLogln("trackerslist fetch failed, using DefaultTrackers:", reason)
	}
}

func loadNewTracker() {
	loadTrackersOnce.Do(func() {
		local := configuredDefaultTrackers()
		url := configuredTrackersListURL()
		if url == "" {
			loadedTrackers = local
			return
		}

		client := &http.Client{Timeout: trackersFetchTimeout}
		resp, err := client.Get(url)
		if err != nil {
			useDefaultTrackersFallback(err.Error())
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			useDefaultTrackersFallback(fmt.Sprintf("status %d", resp.StatusCode))
			return
		}
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			useDefaultTrackersFallback(err.Error())
			return
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
			useDefaultTrackersFallback("empty list")
			return
		}
		loadedTrackers = append(remote, local...)
	})
}

// resetLoadedTrackersForTest clears cached trackers so tests can re-run loadNewTracker.
func resetLoadedTrackersForTest() {
	InvalidateTrackersCache()
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
