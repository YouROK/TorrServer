package settings

import (
	"encoding/json"
	"sync"

	"server/log"
)

type Viewed struct {
	Hash      string  `json:"hash"`
	FileIndex int     `json:"file_index"`
	TimeCode  float64 `json:"timecode"`
	// Duration is the real length of the media the timecode was taken from, when the server
	// saved the position itself. A client showing progress needs it: the runtime a catalogue
	// gives for a title is not the length of this particular file, which may have had its
	// intro cut or an advert left in.
	Duration float64 `json:"duration,omitempty"`
}

// viewedRec is the persisted per-file record.
type viewedRec struct {
	TimeCode float64 `json:"tc"`
	Duration float64 `json:"dur,omitempty"`
}

// readIndexes reads the records of one torrent's files. Databases written before positions
// were saved here hold a bare timecode per file, and older ones still an empty object; both
// are read as a record with what they carry.
func readIndexes(buf []byte) map[int]viewedRec {
	m := map[int]viewedRec{}
	if len(buf) == 0 {
		return m
	}
	if json.Unmarshal(buf, &m) == nil {
		return m
	}
	timecodes := map[int]float64{}
	if json.Unmarshal(buf, &timecodes) == nil {
		for k, tc := range timecodes {
			m[k] = viewedRec{TimeCode: tc}
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

// One writer at a time. The stored value is a blob holding every file of a torrent, so
// saving one of them reads the blob, changes an entry and writes it back — and two saves
// running together lose whichever finished first. With a position now saved every thirty
// seconds per stream, two files of the same torrent playing at once is enough.
var muViewed sync.Mutex

func SetViewed(vv *Viewed) {
	muViewed.Lock()
	defer muViewed.Unlock()

	rec := viewedRec{TimeCode: vv.TimeCode, Duration: vv.Duration}
	if BTsets == nil || !BTsets.TrackTimecode {
		rec = viewedRec{}
	}
	m := readIndexes(tdb.Get("Viewed", vv.Hash))
	m[vv.FileIndex] = rec
	storeIndexes(vv.Hash, m)
}

// MarkViewed flags a file as viewed without touching an already stored playback
// position. Starting a stream must not wipe where the user left off.
func MarkViewed(hash string, fileIndex int) {
	muViewed.Lock()
	defer muViewed.Unlock()

	m := readIndexes(tdb.Get("Viewed", hash))
	if _, ok := m[fileIndex]; ok {
		return
	}
	m[fileIndex] = viewedRec{}
	storeIndexes(hash, m)
}

func RemViewed(vv *Viewed) {
	muViewed.Lock()
	defer muViewed.Unlock()

	m := readIndexes(tdb.Get("Viewed", vv.Hash))
	if vv.FileIndex != -1 {
		delete(m, vv.FileIndex)
		storeIndexes(vv.Hash, m)
	} else {
		tdb.Rem("Viewed", vv.Hash)
	}
}

func ListViewed(hash string) []*Viewed {
	keys := []string{hash}
	if hash == "" {
		keys = tdb.List("Viewed")
	}
	ret := []*Viewed{}
	for _, key := range keys {
		buf := tdb.Get("Viewed", key)
		if len(buf) == 0 {
			continue
		}
		for i, rec := range readIndexes(buf) {
			ret = append(ret, &Viewed{Hash: key, FileIndex: i, TimeCode: rec.TimeCode, Duration: rec.Duration})
		}
	}
	return ret
}
