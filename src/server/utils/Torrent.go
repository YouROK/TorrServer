package utils

import (
	"encoding/base32"
	"errors"
	"math/rand"
	"time"

	"server/settings"

	"github.com/anacrolix/torrent"
	"golang.org/x/time/rate"
)

var trackers = []string{
	"http://bt4.t-ru.org/ann?magnet",
	"http://retracker.mgts.by:80/announce",
	"http://tracker.city9x.com:2710/announce",
	"http://tracker.electro-torrent.pl:80/announce",
	"http://tracker.internetwarriors.net:1337/announce",
	"http://tracker2.itzmx.com:6961/announce",
	"udp://46.148.18.250:2710",
	"udp://opentor.org:2710",
	"udp://public.popcorn-tracker.org:6969/announce",
	"udp://tracker.opentrackr.org:1337/announce",

	"http://bt.svao-ix.ru/announce",
}

func GetDefTrackers() []string {
	return trackers
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
	readahead := int64(float64(settings.Get().CacheSize) * 0.33)
	if readahead < 66*1024*1024 {
		readahead = int64(settings.Get().CacheSize)
		if readahead > 66*1024*1024 {
			readahead = 66 * 1024 * 1024
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
