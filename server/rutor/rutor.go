package rutor

import (
	"compress/flate"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/agnivade/levenshtein"

	"server/log"
	"server/rutor/models"
	"server/rutor/torrsearch"
	"server/rutor/utils"
	"server/settings"
	utils2 "server/torr/utils"
)

var (
	mu    sync.RWMutex
	torrs []*models.TorrentDetails
	// gen counts Start and Stop calls. An updater goroutine runs for the value
	// Start gave it and exits as soon as gen moves, so a Stop that lands before
	// the goroutine's first step, or a Stop then Start inside one tick, never
	// leaves a stale updater behind.
	gen atomic.Uint64
	// updaters tracks running updater goroutines. Production code never waits
	// on it; tests do, to get a happens-before edge with the goroutine's reads
	// of settings before they restore them.
	updaters sync.WaitGroup

	// rutorDBURL is a var so tests can serve the index from httptest.
	rutorDBURL = "http://releases.yourok.ru/torr/rutor.ls"
)

func Start() {
	// Read the setting here, not in the goroutine: the caller owns BTsets and
	// a disabled search must not leave a goroutine behind.
	if !settings.BTsets.EnableRutorSearch {
		return
	}
	g := gen.Add(1)
	updaters.Add(1)
	go func() {
		defer updaters.Done()
		if !updateDB() {
			loadDB()
		}
		for gen.Load() == g {
			for i := 0; i < 3*60*60; i++ {
				time.Sleep(time.Second)
				if gen.Load() != g {
					return
				}
			}
			updateDB()
		}
	}()
}

func Stop() {
	mu.Lock()
	gen.Add(1)
	torrs = nil
	torrsearch.NewIndex(nil)
	mu.Unlock()
	utils2.FreeOSMemGC()
	time.Sleep(time.Millisecond * 1500)
}

// http://releases.yourok.ru/torr/rutor.ls
func updateDB() bool {
	log.TLogln("Update rutor db")

	fnOrig := filepath.Join(settings.Path, "rutor.ls")

	if fi, err := os.Stat(fnOrig); err == nil {
		if time.Since(fi.ModTime()) < time.Minute*175 /*2:55*/ {
			log.TLogln("Less 3 hours rutor db old")
			return false
		}
	}

	fnTmp := filepath.Join(settings.Path, "rutor.tmp")
	out, err := os.Create(fnTmp)
	if err != nil {
		log.TLogln("Error create file rutor.tmp:", err)
		return false
	}

	resp, err := http.Get(rutorDBURL)
	if err != nil {
		log.TLogln("Error connect to rutor db:", err)
		_ = out.Close()
		_ = os.Remove(fnTmp)
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		// An error page must not replace a good rutor.ls.
		log.TLogln("Error download rutor db: status", resp.Status)
		_ = out.Close()
		_ = os.Remove(fnTmp)
		return false
	}
	_, err = io.Copy(out, resp.Body)
	_ = out.Close()
	if err != nil {
		log.TLogln("Error download rutor db:", err)
		_ = os.Remove(fnTmp)
		return false
	}

	md5Tmp := utils.MD5File(fnTmp)
	md5Orig := utils.MD5File(fnOrig)
	if md5Tmp != md5Orig {
		err = os.Remove(fnOrig)
		if err != nil && !os.IsNotExist(err) {
			log.TLogln("Error remove old rutor db:", err)
			return false
		}
		err = os.Rename(fnTmp, fnOrig)
		if err != nil {
			log.TLogln("Error rename rutor db:", err)
			return false
		}
		loadDB()
		return true
	} else {
		_ = os.Remove(fnTmp)
	}
	return false
}

func loadDB() { loadDBAt(gen.Load()) }

// loadDBAt publishes the index only if gen is still g when decoding ends. A
// Stop (or Start) that lands meanwhile makes the load stale: it must not
// republish an index Stop just dropped.
func loadDBAt(g uint64) {
	log.TLogln("Load rutor db")
	ff, err := os.Open(filepath.Join(settings.Path, "rutor.ls"))
	if err == nil {
		defer func() { _ = ff.Close() }()
		r := flate.NewReader(ff)
		defer func() { _ = r.Close() }()
		var ftorrs []*models.TorrentDetails
		dec := json.NewDecoder(r)

		_, err := dec.Token()
		if err != nil {
			log.TLogln("Error read token rutor db:", err)
			return
		}

		for dec.More() {
			var torr *models.TorrentDetails
			err = dec.Decode(&torr)
			if err == nil {
				ftorrs = append(ftorrs, torr)
			}
		}
		mu.Lock()
		defer mu.Unlock()
		if gen.Load() != g {
			return
		}
		torrs = ftorrs
		log.TLogln("Index rutor db")
		torrsearch.NewIndex(torrs)
		log.TLogln("Torrents count:", len(torrs))
		log.TLogln("Indexed words:", len(torrsearch.GetIDX()))

	} else {
		log.TLogln("Error load rutor db:", err)
	}
	utils2.FreeOSMemGC()
}

func Search(query string) []*models.TorrentDetails {
	if !settings.BTsets.EnableRutorSearch {
		return nil
	}
	mu.RLock()
	matchedIDs := torrsearch.Search(query)
	if len(matchedIDs) == 0 {
		mu.RUnlock()
		return nil
	}
	var list []*models.TorrentDetails
	for _, id := range matchedIDs {
		list = append(list, torrs[id])
	}
	mu.RUnlock()

	hash := utils.ClearStr(query)

	sort.Slice(list, func(i, j int) bool {
		lhash := utils.ClearStr(strings.ToLower(list[i].Name+list[i].GetNames())) + strconv.Itoa(list[i].Year)
		lev1 := levenshtein.ComputeDistance(hash, lhash)
		lhash = utils.ClearStr(strings.ToLower(list[j].Name+list[j].GetNames())) + strconv.Itoa(list[j].Year)
		lev2 := levenshtein.ComputeDistance(hash, lhash)
		if lev1 == lev2 {
			return list[j].CreateDate.Before(list[i].CreateDate)
		}
		return lev1 < lev2
	})
	return list
}
