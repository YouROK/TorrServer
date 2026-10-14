package torrstor

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/anacrolix/torrent"

	"server/log"
	"server/settings"
)

// Pinned pieces are downloaded in full into the long-term disk cache on request
// ("download file" / "download torrent") instead of following a reader. Each request
// is a job: a range of pieces. While a job has pieces left to download, priority
// resets leave its pieces alone and global eviction keeps the whole range, including
// the pieces already cached, so finishing a file never evicts its beginning. Once
// every job is done the pieces age out of the cache like any other. Unfinished jobs
// are persisted next to the cached pieces and resume when the torrent opens again.

const pinsFileName = "pins"

// PieceRange is a half-open range of piece indexes [Begin, End).
type PieceRange struct {
	Begin int `json:"begin"`
	End   int `json:"end"`
}

// Lock order: muPrio -> client lock -> muPinned. The torrent client calls
// Piece.MarkComplete -> pieceCompleted (muPinned) while holding its own lock, so
// nothing may call into the client (PieceState, SetPriority) while holding muPinned.

// PinPieces requests the incomplete pieces in r and keeps them pinned until they
// are downloaded.
func (c *Cache) PinPieces(r PieceRange) {
	if c == nil || c.torrent == nil {
		return
	}
	r = PieceRange{Begin: max(r.Begin, 0), End: min(r.End, c.pieceCount)}
	if r.Begin >= r.End {
		return
	}

	c.muPrio.Lock()
	incomplete := make([]int, 0, r.End-r.Begin)
	for id := r.Begin; id < r.End; id++ {
		if !c.torrent.PieceState(id).Complete {
			incomplete = append(incomplete, id)
		}
	}
	if len(incomplete) == 0 {
		c.muPrio.Unlock()
		return
	}

	c.muPinned.Lock()
	if c.pinned == nil {
		c.pinned = make(map[int]struct{})
	}
	for _, id := range incomplete {
		c.pinned[id] = struct{}{}
	}
	if !containsRange(c.pinJobs, r) {
		c.pinJobs = append(c.pinJobs, r)
	}
	c.muPinned.Unlock()

	for _, id := range incomplete {
		c.torrent.Piece(id).SetPriority(torrent.PiecePriorityNormal)
	}
	c.muPrio.Unlock()

	// a piece that completed between the snapshot above and pinning it was not
	// unpinned by pieceCompleted; do it now, or the job would never finish
	for _, id := range incomplete {
		if c.torrent.PieceState(id).Complete {
			c.pieceCompleted(id)
		}
	}
	c.savePins()
}

// UnpinPieces cancels the download of the pieces in r. Pieces a reader still
// needs keep the priority the reader gave them.
func (c *Cache) UnpinPieces(r PieceRange) {
	if c == nil {
		return
	}
	ranges := c.collectReaderRanges()

	c.muPrio.Lock()
	c.muPinned.Lock()
	jobs := c.pinJobs[:0]
	for _, job := range c.pinJobs {
		if job.Begin < r.Begin || job.End > r.End {
			jobs = append(jobs, job)
		}
	}
	c.pinJobs = jobs
	unpinned := make([]int, 0)
	for id := max(r.Begin, 0); id < min(r.End, c.pieceCount); id++ {
		if _, ok := c.pinned[id]; !ok || inPieceRanges(c.pinJobs, id) {
			continue
		}
		delete(c.pinned, id)
		unpinned = append(unpinned, id)
	}
	if len(c.pinned) == 0 {
		c.pinJobs = nil
	}
	c.muPinned.Unlock()

	if c.torrent != nil {
		for _, id := range unpinned {
			if !inRanges(ranges, id) && !c.torrent.PieceState(id).Complete {
				c.torrent.Piece(id).SetPriority(torrent.PiecePriorityNone)
			}
		}
	}
	c.muPrio.Unlock()
	c.savePins()
}

// PinnedPieces returns how many pieces in r are still waiting to be downloaded.
func (c *Cache) PinnedPieces(r PieceRange) int {
	if c == nil {
		return 0
	}
	c.muPinned.Lock()
	defer c.muPinned.Unlock()
	if len(c.pinned) == 0 {
		return 0
	}
	count := 0
	for id := range c.pinned {
		if id >= r.Begin && id < r.End {
			count++
		}
	}
	return count
}

// HasPinned reports whether a download is in progress; the torrent must stay open
// until it finishes.
func (c *Cache) HasPinned() bool {
	if c == nil {
		return false
	}
	c.muPinned.Lock()
	defer c.muPinned.Unlock()
	return len(c.pinned) > 0
}

// isPinned reports whether a piece is waiting to be downloaded by a job.
func (c *Cache) isPinned(id int) bool {
	c.muPinned.Lock()
	defer c.muPinned.Unlock()
	_, ok := c.pinned[id]
	return ok
}

// inPinJob reports whether a piece belongs to an unfinished job; eviction keeps it.
func (c *Cache) inPinJob(id int) bool {
	c.muPinned.Lock()
	defer c.muPinned.Unlock()
	return inPieceRanges(c.pinJobs, id)
}

func inPieceRanges(ranges []PieceRange, id int) bool {
	for _, r := range ranges {
		if id >= r.Begin && id < r.End {
			return true
		}
	}
	return false
}

func containsRange(ranges []PieceRange, r PieceRange) bool {
	for _, have := range ranges {
		if have == r {
			return true
		}
	}
	return false
}

// pieceCompleted unpins a downloaded piece; the last one removes the pins file.
// The client calls it holding its lock, so it only touches muPinned and leaves
// the file to a goroutine.
func (c *Cache) pieceCompleted(id int) {
	c.muPinned.Lock()
	_, ok := c.pinned[id]
	if ok {
		delete(c.pinned, id)
	}
	done := ok && len(c.pinned) == 0
	if done {
		c.pinJobs = nil
	}
	c.muPinned.Unlock()
	if done {
		log.TLogln("Download to cache complete:", c.hash.HexString())
		go c.savePins()
	}
}

// restorePins re-pins the download saved by a previous session.
func (c *Cache) restorePins() {
	for _, r := range LoadPins(c.hash.HexString()) {
		c.PinPieces(r)
	}
}

func (c *Cache) pinnedRanges() []PieceRange {
	c.muPinned.Lock()
	defer c.muPinned.Unlock()
	return append([]PieceRange(nil), c.pinJobs...)
}

// savePins writes the current jobs. The snapshot is taken under muSavePins, so
// whichever call writes last writes the latest state.
func (c *Cache) savePins() {
	if !IsLongTermCache() {
		return
	}
	c.muSavePins.Lock()
	defer c.muSavePins.Unlock()
	path := pinsPath(c.hash.HexString())
	ranges := c.pinnedRanges()
	if len(ranges) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.TLogln("Error remove download state:", err)
		}
		return
	}
	data, err := json.Marshal(ranges)
	if err != nil {
		return
	}
	if err := os.WriteFile(path, data, 0o666); err != nil {
		log.TLogln("Error save download state:", err)
	}
}

func pinsPath(hashHex string) string {
	return filepath.Join(settings.BTsets.TorrentsSavePath, hashHex, pinsFileName)
}

// LoadPins returns the unfinished download saved for a torrent.
func LoadPins(hashHex string) []PieceRange {
	if !IsLongTermCache() {
		return nil
	}
	data, err := os.ReadFile(pinsPath(hashHex))
	if err != nil || len(data) == 0 {
		return nil
	}
	var ranges []PieceRange
	if err := json.Unmarshal(data, &ranges); err != nil {
		log.TLogln("Error read download state:", err)
		return nil
	}
	return ranges
}

// HasPins reports whether a torrent has an unfinished download to resume.
func HasPins(hashHex string) bool {
	return len(LoadPins(hashHex)) > 0
}
