// Package timeindex turns a byte offset in a media file into playback time.
//
// The time is not inferred from the average bitrate. It is read out of the container as the
// bytes go past: every format that carries absolute timestamps in-band — Matroska cluster
// timestamps, MPEG-TS program clock references, fragmented MP4 fragment times — is covered
// by watching the stream a client is already pulling. Nothing extra is downloaded, and a
// variable bitrate costs nothing, because no bitrate is ever assumed. Progressive MP4 keeps
// its table in the moov atom instead, and every player reads that atom before it can start,
// so it goes past as well.
//
// Lookups round down: the answer is the last timestamp at or before the offset asked about,
// never the next one. A caller that must not overshoot the picture can use it as is.
package timeindex

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// maxSamples caps memory: past it the index is thinned by dropping every second sample,
// which halves the resolution and keeps the coverage. A three-hour film fills a few thousand.
const maxSamples = 1 << 15

type sample struct {
	off int64
	sec float64
}

// parser reads timestamps out of one container format.
type parser interface {
	// feed hands over bytes that start at file offset off, calling emit for every timestamp
	// found. Implementations keep whatever tail they need to span the next call.
	feed(off int64, p []byte, emit func(off int64, sec float64))
	// reset drops any partial state, because the stream jumped to an unrelated offset.
	reset()
}

// tail is what a parser keeps between reads: the last few bytes of the previous one, so an
// element split across two reads is still seen whole.
type tail struct {
	carry   []byte
	carryAt int64
	scratch []byte
}

// join prepends the carry to p when the two are contiguous, and says where the result
// starts in the file. The joined buffer lives in a scratch slice reused across reads, so the
// join stops allocating once it has grown to the size of a read.
func (t *tail) join(off int64, p []byte) (buf []byte, base int64) {
	if len(t.carry) == 0 || t.carryAt+int64(len(t.carry)) != off {
		return p, off
	}
	t.scratch = append(append(t.scratch[:0], t.carry...), p...)
	return t.scratch, t.carryAt
}

// hold keeps at most keep bytes from the end of buf for the next read. The offset is
// adjusted before the reslice: afterwards the length is already keep and the correction
// comes out zero, leaving carryAt pointing at bytes that were dropped.
func (t *tail) hold(buf []byte, base int64, keep int) {
	if len(buf) > keep {
		base += int64(len(buf) - keep)
		buf = buf[len(buf)-keep:]
	}
	t.carry = append(t.carry[:0], buf...)
	t.carryAt = base
}

// drop forgets the carry, for a parser that is now collecting something larger instead.
func (t *tail) drop() { t.carry = t.carry[:0] }

// collector gathers one element that runs past the end of a read — a box, an index — until
// it has all of it.
type collector struct {
	buf  []byte
	at   int64
	want int64
}

// start begins collecting an element of size bytes at file offset at, of which head has
// already been read.
func (c *collector) start(at int64, head []byte, size int64) {
	c.at, c.want = at, size
	c.buf = append([]byte(nil), head...)
}

func (c *collector) collecting() bool { return c.want > 0 }

// take consumes from p what the element still needs, and reports how much that was and
// whether the element is now complete.
func (c *collector) take(p []byte) (n int, done bool) {
	n = min(len(p), int(c.want-int64(len(c.buf))))
	c.buf = append(c.buf, p[:n]...)
	return n, int64(len(c.buf)) >= c.want
}

// finish hands over the collected element and its offset, and stops collecting.
func (c *collector) finish() ([]byte, int64) {
	buf, at := c.buf, c.at
	c.buf, c.want = nil, 0
	return buf, at
}

func (c *collector) reset() { c.buf, c.want = nil, 0 }

// Index collects (offset, time) pairs seen in a file and answers questions about offsets.
// One index belongs to one file and is shared by everything reading it. The zero value is
// unusable; call New.
type Index struct {
	mu       sync.Mutex
	newSeen  func() parser
	source   string
	samples  []sample
	origin   float64 // timestamp the container gives to the first frame
	duration float64 // how long the media runs, zero until known
}

// Feeder reads one stream into the index. A player opens several connections at once — the
// header on one, the picture on another, a probe on a third — and each walks its own part of
// the file. What they find is shared; the parsing must not be. One parser fed from several
// places sees every switch between them as the stream jumping, and a format whose headers
// have to be followed in order never recovers from that.
type Feeder struct {
	index   *Index
	parser  parser
	emit    func(off int64, sec float64) // built once: a closure per read would be garbage
	nextOff int64
	started bool

	// The first timestamp this stream itself found. The index outlives any one stream and
	// holds timestamps from wherever the file has been read before, so asking it what plays
	// at a byte a stream has only just started on can be answered from somewhere else
	// entirely — a rewind lands in front of everything the last session indexed, and the
	// nearest answer is then wherever that session had got to.
	firstSec float64
	firstOK  bool
}

// First is the earliest timestamp this stream found, which is the film time it started at.
func (f *Feeder) First() (float64, bool) {
	if f == nil {
		return 0, false
	}
	f.index.mu.Lock()
	defer f.index.mu.Unlock()
	return f.firstSec - f.index.origin, f.firstOK
}

// New picks a parser from the file name. It returns nil for containers with no timestamps
// to read, and callers fall back to whatever estimate they have.
func New(name string) *Index {
	var build func() parser
	var source string
	switch strings.ToLower(filepath.Ext(name)) {
	case ".mkv", ".mka", ".mks", ".webm", ".mk3d":
		build, source = func() parser { return newMatroska() }, "matroska"
	case ".ts", ".m2ts", ".mts", ".tsv", ".m2t":
		build, source = func() parser { return newMPEGTS() }, "mpegts"
	case ".mp4", ".m4v", ".mov", ".m4a", ".qt":
		build, source = func() parser { return newMP4() }, "mp4"
	case ".vob", ".mpg", ".mpeg", ".mpe", ".m2p", ".ps", ".evo":
		build, source = func() parser { return newMPEGPS() }, "mpegps"
	case ".flv", ".f4v":
		build, source = func() parser { return newFLV() }, "flv"
	case ".avi", ".divx":
		build, source = func() parser { return newAVI() }, "avi"
	}
	if build == nil {
		return nil
	}
	return &Index{newSeen: build, source: source}
}

// Feeder opens a stream of its own into the index.
func (ix *Index) Feeder() *Feeder {
	if ix == nil {
		return nil
	}
	f := &Feeder{index: ix, parser: ix.newSeen()}
	f.emit = func(at int64, sec float64) {
		if !f.firstOK {
			f.firstSec, f.firstOK = sec, true
		}
		ix.add(at, sec)
	}
	return f
}

// Source names the container the timestamps come from, for display.
func (ix *Index) Source() string {
	if ix == nil {
		return ""
	}
	return ix.source
}

// SetDuration records how long the media runs, which is the one thing every timestamp in it
// has to fit inside. A parser that has been fooled tends to produce times far past the end,
// and without this they look as ordinary as any other.
func (ix *Index) SetDuration(seconds float64) {
	if ix == nil || seconds <= 0 {
		return
	}
	ix.mu.Lock()
	defer ix.mu.Unlock()
	ix.duration = seconds
}

// SetOrigin records the timestamp the container gives to the first frame. MPEG-TS clocks
// start wherever the muxer left them, so without this its times are offset by a constant;
// ffprobe reports it as the format start time. Formats that count from zero ignore it.
func (ix *Index) SetOrigin(seconds float64) {
	if ix == nil || seconds <= 0 {
		return
	}
	ix.mu.Lock()
	defer ix.mu.Unlock()
	ix.origin = seconds
}

// Feed hands this stream's bytes to the index, starting at file offset off.
func (f *Feeder) Feed(off int64, p []byte) {
	if f == nil || len(p) == 0 {
		return
	}
	if !f.started || off != f.nextOff {
		// A seek, or the first read: whatever the parser was holding belongs to a different
		// part of the file and would be read as if it ran into these bytes.
		f.parser.reset()
		f.started = true
	}
	f.nextOff = off + int64(len(p))

	f.index.mu.Lock()
	defer f.index.mu.Unlock()
	f.parser.feed(off, p, f.emit)
}

// add files one timestamp, keeping the samples ordered by offset.
func (ix *Index) add(off int64, sec float64) {
	if off < 0 || sec < 0 {
		return
	}
	if ix.duration > 0 && sec > ix.origin+ix.duration*1.05 {
		return // past the end of the media: not a timestamp this file contains
	}
	// Within one stream offsets only grow, so nearly every timestamp lands past the end and
	// the search is skipped for it.
	if n := len(ix.samples); n == 0 || ix.samples[n-1].off < off && ix.samples[n-1].sec <= sec {
		ix.samples = append(ix.samples, sample{off: off, sec: sec})
		ix.thin()
		return
	}
	at := sort.Search(len(ix.samples), func(i int) bool { return ix.samples[i].off >= off })
	if at < len(ix.samples) && ix.samples[at].off == off {
		return // already known
	}
	// Time and offset both only ever increase within a file, so a timestamp that does not fit
	// between its neighbours did not come from the container — it came from a byte pattern
	// that looked like a header. A parser that lets one through poisons every lookup near it.
	if at > 0 && ix.samples[at-1].sec > sec {
		return
	}
	if at < len(ix.samples) && ix.samples[at].sec < sec {
		return
	}
	ix.samples = append(ix.samples, sample{})
	copy(ix.samples[at+1:], ix.samples[at:])
	ix.samples[at] = sample{off: off, sec: sec}
	ix.thin()
}

// thin halves the index once it is over the cap, keeping the coverage at half the resolution.
func (ix *Index) thin() {
	if len(ix.samples) <= maxSamples {
		return
	}
	kept := ix.samples[:0]
	for i, s := range ix.samples {
		if i%2 == 0 {
			kept = append(kept, s)
		}
	}
	ix.samples = kept
}

// rel turns a container timestamp into playback time, which starts at zero.
func (ix *Index) rel(sec float64) float64 {
	return max(sec-ix.origin, 0)
}

// TimeAt is the playback time at a byte offset, taken from the last timestamp at or before
// it. The second result is false while nothing has been seen that early in the file.
func (ix *Index) TimeAt(off int64) (float64, bool) {
	if ix == nil {
		return 0, false
	}
	ix.mu.Lock()
	defer ix.mu.Unlock()

	at := sort.Search(len(ix.samples), func(i int) bool { return ix.samples[i].off > off })
	if at == 0 {
		return 0, false
	}
	return ix.rel(ix.samples[at-1].sec), true
}

// TimeBetween is the playback time at a byte offset, read between the timestamps on either
// side of it rather than rounded back to the one before. Timestamps stand seconds apart, and
// rounding back to them puts a step into every reading; a caller measuring how the position
// moves sees those steps as the picture stalling and restarting. Within the stretch between
// two timestamps the bitrate is whatever it is and does not change, so the point in between
// follows from the byte positions alone.
func (ix *Index) TimeBetween(off int64) (float64, bool) {
	if ix == nil {
		return 0, false
	}
	ix.mu.Lock()
	defer ix.mu.Unlock()

	at := sort.Search(len(ix.samples), func(i int) bool { return ix.samples[i].off > off })
	if at == 0 {
		return 0, false
	}
	before := ix.samples[at-1]
	sec := before.sec
	switch {
	case at < len(ix.samples):
		after := ix.samples[at]
		if span := after.off - before.off; span > 0 {
			sec += (after.sec - before.sec) * float64(off-before.off) / float64(span)
		}
	case at >= 2:
		// Past the last timestamp there is nothing to interpolate towards, and holding the
		// last one turns the answer into a staircase — which anything measuring how the
		// position moves reads as the picture stopping and starting. The slope carries on
		// from the two most recent timestamps.
		//
		// It is not held back to within one gap between them. The read head sits further out
		// than that most of the time, so a limit there is reached on nearly every reading and
		// the staircase comes straight back, costing a tenth of every step. Reading too far
		// ahead cannot push a position past the picture in any case: what is shown is capped
		// by the clock, and no amount of film arriving lets it outrun that.
		prev := ix.samples[at-2]
		if span := before.off - prev.off; span > 0 {
			sec += (before.sec - prev.sec) * float64(off-before.off) / float64(span)
		}
	}
	return ix.rel(sec), true
}

// TimeFrom is the playback time at the first timestamp at or after a byte offset. Where
// TimeAt answers "what is playing here at the latest", this answers "what is playing here at
// the earliest" — which is what a caller wants for the point a session started reading, since
// the timestamps in front of that point belong to a different part of the file entirely.
func (ix *Index) TimeFrom(off int64) (float64, bool) {
	if ix == nil {
		return 0, false
	}
	ix.mu.Lock()
	defer ix.mu.Unlock()

	at := sort.Search(len(ix.samples), func(i int) bool { return ix.samples[i].off >= off })
	if at >= len(ix.samples) {
		return 0, false
	}
	return ix.rel(ix.samples[at].sec), true
}

// OffsetAt is where in the file a playback time is found, taken from the last timestamp at
// or before it — the inverse of TimeAt, and rounding the same way.
func (ix *Index) OffsetAt(seconds float64) (int64, bool) {
	if ix == nil {
		return 0, false
	}
	ix.mu.Lock()
	defer ix.mu.Unlock()

	want := seconds + ix.origin
	at := sort.Search(len(ix.samples), func(i int) bool { return ix.samples[i].sec > want })
	if at == 0 {
		return 0, false
	}
	return ix.samples[at-1].off, true
}
