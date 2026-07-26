package settings

import (
	"encoding/json"

	"server/log"
)

type Viewed struct {
	Hash      string  `json:"hash"`
	FileIndex int     `json:"file_index"`
	TimeCode  float64 `json:"timecode"`
	// Playback position details, filled by the auto-save feature. Offset/Length let a
	// client verify or recompute the position; Duration is the real media duration
	// (from ffprobe), so clients don't have to guess it from external metadata.
	Offset   int64   `json:"offset,omitempty"`
	Length   int64   `json:"length,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

// viewedRec is the persisted per-file record. Older databases hold a bare float
// (timecode) or, even older, an empty struct; readIndexes migrates both on read.
type viewedRec struct {
	TimeCode float64 `json:"tc"`
	Offset   int64   `json:"off,omitempty"`
	Length   int64   `json:"len,omitempty"`
	Duration float64 `json:"dur,omitempty"`
}

func readIndexes(buf []byte) map[int]viewedRec {
	if len(buf) == 0 {
		return map[int]viewedRec{}
	}
	recs := map[int]viewedRec{}
	if json.Unmarshal(buf, &recs) == nil {
		return recs
	}
	timecodes := map[int]float64{}
	if json.Unmarshal(buf, &timecodes) == nil {
		m := make(map[int]viewedRec, len(timecodes))
		for k, v := range timecodes {
			m[k] = viewedRec{TimeCode: v}
		}
		return m
	}
	legacy := map[int]struct{}{}
	m := map[int]viewedRec{}
	if json.Unmarshal(buf, &legacy) == nil {
		for k := range legacy {
			m[k] = viewedRec{}
		}
	}
	return m
}

func storeIndexes(hash string, m map[int]viewedRec) {
	buf, err := json.Marshal(m)
	if err != nil {
		log.TLogln("Error set viewed:", err)
		return
	}
	tdb.Set("Viewed", hash, buf)
}

func keepTimecode() bool {
	return BTsets != nil && (BTsets.TrackTimecode || BTsets.SavePosition)
}

func SetViewed(vv *Viewed) {
	rec := viewedRec{TimeCode: vv.TimeCode, Offset: vv.Offset, Length: vv.Length, Duration: vv.Duration}
	if !keepTimecode() {
		rec = viewedRec{}
	}

	m := readIndexes(tdb.Get("Viewed", vv.Hash))
	m[vv.FileIndex] = rec
	storeIndexes(vv.Hash, m)
}

// MarkViewed flags a file as viewed without touching an already stored playback
// position. Starting a stream must not wipe where the user left off.
func MarkViewed(hash string, fileIndex int) {
	m := readIndexes(tdb.Get("Viewed", hash))
	if _, ok := m[fileIndex]; ok {
		return
	}
	m[fileIndex] = viewedRec{}
	storeIndexes(hash, m)
}

func newViewed(hash string, fileIndex int, rec viewedRec) *Viewed {
	return &Viewed{
		Hash:      hash,
		FileIndex: fileIndex,
		TimeCode:  rec.TimeCode,
		Offset:    rec.Offset,
		Length:    rec.Length,
		Duration:  rec.Duration,
	}
}

func RemViewed(vv *Viewed) {
	buf := tdb.Get("Viewed", vv.Hash)
	m := readIndexes(buf)
	if vv.FileIndex != -1 {
		delete(m, vv.FileIndex)
		storeIndexes(vv.Hash, m)
	} else {
		tdb.Rem("Viewed", vv.Hash)
	}
}

func ListViewed(hash string) []*Viewed {
	if hash != "" {
		buf := tdb.Get("Viewed", hash)
		if len(buf) == 0 {
			return []*Viewed{}
		}
		m := readIndexes(buf)
		var ret []*Viewed
		for i, rec := range m {
			ret = append(ret, newViewed(hash, i, rec))
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
			for i, rec := range m {
				ret = append(ret, newViewed(key, i, rec))
			}
		}
		return ret
	}
}
