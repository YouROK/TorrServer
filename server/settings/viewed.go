package settings

import (
	"encoding/json"

	"server/log"
)

type Viewed struct {
	Hash      string  `json:"hash"`
	FileIndex int     `json:"file_index"`
	TimeCode  float64 `json:"timecode"`
}

func readIndexes(buf []byte) map[int]float64 {
	m := map[int]float64{}
	if json.Unmarshal(buf, &m) == nil {
		return m
	}
	var old map[int]struct{}
	if json.Unmarshal(buf, &old) == nil {
		for k := range old {
			m[k] = 0
		}
	}
	return m
}

func SetViewed(vv *Viewed) {
	val := vv.TimeCode
	if BTsets == nil || !BTsets.TrackTimecode {
		val = 0
	}

	m := readIndexes(tdb.Get("Viewed", vv.Hash))
	m[vv.FileIndex] = val
	buf, err := json.Marshal(m)
	if err == nil {
		tdb.Set("Viewed", vv.Hash, buf)
	} else {
		log.TLogln("Error set viewed:", err)
	}
}

func RemViewed(vv *Viewed) {
	buf := tdb.Get("Viewed", vv.Hash)
	m := readIndexes(buf)
	if vv.FileIndex != -1 {
		delete(m, vv.FileIndex)
		buf, err := json.Marshal(m)
		if err == nil {
			tdb.Set("Viewed", vv.Hash, buf)
		}
	} else {
		tdb.Rem("Viewed", vv.Hash)
	}
}

func ListViewed(hash string) []*Viewed {
	var err error
	if hash != "" {
		buf := tdb.Get("Viewed", hash)
		if len(buf) == 0 {
			return []*Viewed{}
		}
		m := readIndexes(buf)
		var ret []*Viewed
		for i, tc := range m {
			ret = append(ret, &Viewed{Hash: hash, FileIndex: i, TimeCode: tc})
		}
		return ret
	} else {
		var ret []*Viewed
		keys := tdb.List("Viewed")
		for _, key := range keys {
			buf := tdb.Get("Viewed", key)
			if len(buf) == 0 {
				continue
			}
			m := readIndexes(buf)
			for i, tc := range m {
				ret = append(ret, &Viewed{Hash: key, FileIndex: i, TimeCode: tc})
			}
		}
		return ret
	}

	log.TLogln("Error list viewed:", err)
	return []*Viewed{}
}
