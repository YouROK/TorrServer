package rutor

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"github.com/agnivade/levenshtein"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"server/log"
	"server/rutor/models"
	"server/rutor/torrsearch"
	"server/rutor/utils"
	"server/settings"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	torrs  []*models.TorrentDetails
	isStop bool
)

func Start() {
	go func() {
		if settings.BTsets.EnableRutorSearch {
			updateDB()
			isStop = false
			for !isStop {
				for i := 0; i < 3*60*60; i++ {
					time.Sleep(time.Second)
					if isStop {
						return
					}
				}
				updateDB()
			}
		}
	}()
}

func Stop() {
	isStop = true
	time.Sleep(time.Millisecond * 1500)
}

// https://github.com/yourok-0001/releases/raw/master/torr/rutor.ls
func updateDB() {
	log.TLogln("Update rutor db")
	filename := filepath.Join(settings.Path, "rutor.tmp")
	out, err := os.Create(filename)
	if err != nil {
		log.TLogln("Error create file rutor.tmp:", err)
		return
	}
	defer out.Close()
	resp, err := http.Get("https://github.com/yourok-0001/releases/raw/master/torr/rutor.ls")
	if err != nil {
		log.TLogln("Error connect to rutor db:", err)
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.TLogln("Error download rutor db:", err)
		return
	}

	err = os.Remove(filepath.Join(settings.Path, "rutor.ls"))
	if err != nil && !os.IsNotExist(err) {
		log.TLogln("Error remove old rutor db:", err)
		return
	}
	err = os.Rename(filename, filepath.Join(settings.Path, "rutor.ls"))
	if err != nil {
		log.TLogln("Error rename rutor db:", err)
		return
	}
	loadDB()
}

func loadDB() {
	log.TLogln("Load rutor db")
	buf, err := os.ReadFile("rutor.ls")
	if err == nil {
		r := flate.NewReader(bytes.NewReader(buf))
		buf, err = io.ReadAll(r)
		r.Close()
		if err == nil {
			var ftors []*models.TorrentDetails
			err = json.Unmarshal(buf, &ftors)
			if err == nil {
				torrs = ftors
				log.TLogln("Index rutor db")
				torrsearch.NewIndex(torrs)
			}
		}
	}
}

func Search(query string) []*models.TorrentDetails {
	matchedIDs := torrsearch.Search(query)
	if len(matchedIDs) == 0 {
		return nil
	}
	var list []*models.TorrentDetails
	for _, id := range matchedIDs {
		list = append(list, torrs[id])
	}

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
