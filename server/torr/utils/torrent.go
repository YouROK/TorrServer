package utils

import (
	"encoding/base32"
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

var defTrackers = []string{
	"http://retracker.local/announce",
	"http://bt4.t-ru.org/ann?magnet",
	"http://retracker.mgts.by:80/announce",
	"http://tracker.city9x.com:2710/announce",
	"http://tracker.electro-torrent.pl:80/announce",
	"http://tracker.internetwarriors.net:1337/announce",
	"http://tracker2.itzmx.com:6961/announce",
	"udp://opentor.org:2710",
	"udp://public.popcorn-tracker.org:6969/announce",
	"udp://tracker.opentrackr.org:1337/announce",
	"http://bt.svao-ix.ru/announce",
	"udp://explodie.org:6969/announce",
	"wss://tracker.btorrent.xyz",
	"wss://tracker.openwebtorrent.com",
}

var loadedTrackers []string
var lastTrackerUpdate time.Time
var trackersLock sync.Mutex

func SaveUniqueTrackers(trackers [][]string) {
	// Count total trackers for logging
	totalInput := 0
	for _, tier := range trackers {
		totalInput += len(tier)
	}

	if len(trackers) == 0 || totalInput == 0 {
		log.TLogln("[Trackers] SaveUniqueTrackers called with empty trackers")
		return
	}

	log.TLogln("[Trackers] SaveUniqueTrackers called with", len(trackers), "tiers,", totalInput, "trackers total")

	trackersLock.Lock()
	defer trackersLock.Unlock()

	name := filepath.Join(settings.Path, "trackers.txt")
	log.TLogln("[Trackers] File path:", name)

	existing := make(map[string]bool)

	// Read existing trackers to avoid duplicates
	buf, err := os.ReadFile(name)
	if err == nil {
		lines := strings.Split(string(buf), "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" {
				existing[l] = true
			}
		}
		log.TLogln("[Trackers] Existing trackers in file:", len(existing))
	} else {
		log.TLogln("[Trackers] File not found or read error:", err, "- will create new file")
	}

	var newTrackers []string
	for _, tier := range trackers {
		for _, tr := range tier {
			tr = strings.TrimSpace(tr)
			if tr != "" && (strings.HasPrefix(tr, "udp") || strings.HasPrefix(tr, "http") || strings.HasPrefix(tr, "wss")) {
				if !existing[tr] {
					newTrackers = append(newTrackers, tr)
					existing[tr] = true // Mark as added to avoid duplicates in the same batch
				}
			}
		}
	}

	if len(newTrackers) > 0 {
		log.TLogln("[Trackers] Writing", len(newTrackers), "new trackers to file")
		f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.TLogln("[Trackers] ERROR: Cannot open file for writing:", err)
			return
		}
		defer f.Close()

		for _, tr := range newTrackers {
			f.WriteString(tr + "\n")
		}
		log.TLogln("[Trackers] Successfully wrote", len(newTrackers), "trackers")
	} else {
		log.TLogln("[Trackers] No new trackers to add (all", totalInput, "already exist)")
	}
}

func GetTrackerFromFile() []string {
	name := filepath.Join(settings.Path, "trackers.txt")
	buf, err := os.ReadFile(name)
	if err == nil {
		list := strings.Split(string(buf), "\n")
		var ret []string
		for _, l := range list {
			l = strings.TrimSpace(l)
			if strings.HasPrefix(l, "udp") || strings.HasPrefix(l, "http") || strings.HasPrefix(l, "wss") {
				ret = append(ret, l)
			}
		}
		if len(ret) > 0 {
			log.TLogln("[Trackers] Loaded", len(ret), "trackers from", name)
		}
		return ret
	} else if !os.IsNotExist(err) {
		log.TLogln("[Trackers] Warning: could not read trackers file:", err)
	}
	return nil
}

func GetDefTrackers() []string {
	loadNewTracker()
	if len(loadedTrackers) == 0 {
		return defTrackers
	}
	return loadedTrackers
}

func loadNewTracker() {
	// Check if we need to refresh: empty or older than 30 days
	if len(loadedTrackers) > 0 && time.Since(lastTrackerUpdate) < 30*24*time.Hour {
		return
	}
	resp, err := http.Get("https://raw.githubusercontent.com/ngosang/trackerslist/master/trackers_best_ip.txt")
	if err == nil {
		defer resp.Body.Close()
		buf, err := io.ReadAll(resp.Body)
		if err == nil {
			arr := strings.Split(string(buf), "\n")
			var ret []string
			for _, s := range arr {
				s = strings.TrimSpace(s)
				if len(s) > 0 {
					ret = append(ret, s)
				}
			}
			loadedTrackers = append(ret, defTrackers...)
			lastTrackerUpdate = time.Now()
		}
	}
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
