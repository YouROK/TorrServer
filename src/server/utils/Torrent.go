package utils

import (
	"encoding/base32"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"server/settings"

	"github.com/anacrolix/torrent"
	"golang.org/x/time/rate"
)

var defTrackers = []string{
	"http://retracker.local",

	"http://bt4.t-ru.org/ann?magnet",
	"http://retracker.mgts.by:80/announce",
	"http://tracker.city9x.com:2710/announce",
	"http://tracker.electro-torrent.pl:80/announce",
	"http://tracker.internetwarriors.net:1337/announce",
	"http://tracker2.itzmx.com:6961/announce",
	"udp4://46.148.18.250:2710",
	"udp://opentor.org:2710",
	"udp://public.popcorn-tracker.org:6969/announce",
	"udp://tracker.opentrackr.org:1337/announce",

	"http://bt.svao-ix.ru/announce",

	"udp://explodie.org:6969/announce",

	//https://github.com/ngosang/trackerslist/blob/master/trackers_best_ip.txt 24.12.2020
	"udp://93.158.213.92:1337/announce",
	"udp://151.80.120.114:2710/announce",
	"udp://31.131.4.62:6969/announce",
	"udp://194.182.165.153:6969/announce",
	"udp://208.83.20.20:6969/announce",
	"udp://184.105.151.164:6969/announce",
	"udp://62.210.97.59:1337/announce",
	"udp://176.113.71.60:6961/announce",
	"udp://51.15.55.204:1337/announce",
	"udp://89.234.156.205:451/announce",
	"udp://5.206.60.196:6969/announce",
	"udp://5.226.148.20:6969/announce",
	"udp://185.181.60.67:80/announce",
	"udp://37.235.174.46:2710/announce",
	"http://54.37.106.164:80/announce",
}

var loadedTrackers []string

func GetDefTrackers() []string {
	loadNewTracker()
	if len(loadedTrackers) == 0 {
		return defTrackers
	}
	return loadedTrackers
}

func loadNewTracker() {
	if len(loadedTrackers) > 0 {
		return
	}
	resp, err := http.Get("https://newtrackon.com/api/stable")
	if err == nil {
		buf, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			arr := strings.Split(string(buf), "\n")
			var ret []string
			for _, s := range arr {
				s = strings.TrimSpace(s)
				if len(s) > 0 {
					ret = append(ret, s)
				}
			}
			loadedTrackers = ret
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

func GotInfo(t *torrent.Torrent, timeout int) error {
	gi := t.GotInfo()
	select {
	case <-gi:
		return nil
	case <-time.Tick(time.Second * time.Duration(timeout)):
		return errors.New("timeout load torrent info")
	}
}

func GetReadahead() int64 {
	readahead := settings.Get().CacheSize - (138412032) //132mb
	if readahead < 69206016 {                           //66mb
		readahead = int64(float64(settings.Get().CacheSize) * 0.33)
		if readahead < 66*1024*1024 {
			readahead = int64(settings.Get().CacheSize)
			if readahead > 66*1024*1024 {
				readahead = 66 * 1024 * 1024
			}
		}
	}
	return readahead
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
