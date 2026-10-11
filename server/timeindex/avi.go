package timeindex

import (
	"bytes"
	"encoding/binary"
)

// AVI is the one container that states no time anywhere in the picture data. What it has is
// a fixed frame rate and, at the end of the file, an index listing every chunk in order — so
// counting video chunks in that index gives the frame number, and the frame number divided
// by the rate gives the time. The index is not fetched for this: a player has to read it
// before it can seek at all, so it goes past like everything else.
//
// Old rips are worth the trouble. A two-pass encode is variable bitrate by definition — it
// moves bits from the quiet scenes to the busy ones — so guessing from the average is no more
// accurate here than on a modern remux.

const (
	aviTail   = 64
	aviMaxIdx = 64 << 20 // an index this large is a ten-hour file; past it, give up
	aviEntry  = 16       // id, flags, offset, length
	aviGap    = 0.5      // seconds between the entries kept, the index is per-frame
)

var (
	aviAvih = []byte("avih")
	aviStrh = []byte("strh")
	aviMovi = []byte("movi")
	aviIdx1 = []byte("idx1")
)

type avi struct {
	tail
	index   collector // the idx1 chunk being collected until it is complete
	fps     float64
	movi    int64 // where the movi four-character code sits, which the index counts from
	video   [2]byte
	haveVid bool
	streams int
	done    bool
}

func newAVI() *avi { return &avi{video: [2]byte{'0', '0'}} }

func (a *avi) reset() {
	a.drop()
	a.index.reset()
}

func (a *avi) feed(off int64, p []byte, emit func(int64, float64)) {
	if a.index.collecting() {
		n, done := a.index.take(p)
		if !done {
			return
		}
		idx, _ := a.index.finish()
		a.readIndex(idx, emit)
		off, p = off+int64(n), p[n:]
		if len(p) == 0 {
			return
		}
	}

	buf, base := a.join(off, p)

	a.readHeaders(buf, base)

	if !a.done {
		if hit := bytes.Index(buf, aviIdx1); hit >= 0 && hit+8 <= len(buf) {
			size := int64(binary.LittleEndian.Uint32(buf[hit+4 : hit+8]))
			if size >= aviEntry && size <= aviMaxIdx {
				body := buf[hit+8:]
				if int64(len(body)) >= size {
					a.readIndex(body[:size], emit)
				} else {
					a.index.start(base+int64(hit+8), body, size)
					a.drop()
					return
				}
			}
		}
	}

	a.hold(buf, base, aviTail)
}

// readHeaders picks up the frame rate, which stream carries the picture, and where the movi
// list starts — the three things the index means nothing without.
func (a *avi) readHeaders(buf []byte, base int64) {
	if a.movi == 0 {
		if hit := bytes.Index(buf, aviMovi); hit >= 0 {
			a.movi = base + int64(hit)
		}
	}
	if a.fps == 0 {
		if hit := bytes.Index(buf, aviAvih); hit >= 0 && hit+12 <= len(buf) {
			if us := binary.LittleEndian.Uint32(buf[hit+8 : hit+12]); us > 0 {
				a.fps = 1e6 / float64(us)
			}
		}
	}
	// Stream headers come in the order the streams are numbered, and the index names chunks
	// by that number, so both the rate and the number are taken from the video one.
	for i := 0; !a.haveVid; {
		hit := bytes.Index(buf[i:], aviStrh)
		if hit < 0 || i+hit+40 > len(buf) {
			break
		}
		at := i + hit
		index := a.streams
		a.streams++
		if string(buf[at+8:at+12]) == "vids" {
			scale := binary.LittleEndian.Uint32(buf[at+28 : at+32])
			rate := binary.LittleEndian.Uint32(buf[at+32 : at+36])
			if scale > 0 && rate > 0 {
				a.fps = float64(rate) / float64(scale)
			}
			a.video = [2]byte{byte('0' + index/10), byte('0' + index%10)}
			a.haveVid = true
		}
		i = at + 4
	}
}

// readIndex walks the index and turns the running count of video chunks into a time.
func (a *avi) readIndex(idx []byte, emit func(int64, float64)) {
	if a.fps <= 0 || a.done {
		return
	}
	// The offsets are stated from the movi four-character code, but some writers store plain
	// file offsets instead. Which it is shows in the first entry: a file offset cannot point
	// in front of the list it belongs to.
	base := a.movi
	if len(idx) >= aviEntry {
		if int64(binary.LittleEndian.Uint32(idx[8:12])) > a.movi {
			base = 0
		}
	}

	frame := 0
	last := -aviGap
	for at := 0; at+aviEntry <= len(idx); at += aviEntry {
		id := idx[at : at+4]
		if id[0] != a.video[0] || id[1] != a.video[1] || (id[2] != 'd' && id[2] != 'w') {
			continue
		}
		if id[2] == 'd' && id[3] != 'c' && id[3] != 'b' {
			continue
		}
		sec := float64(frame) / a.fps
		if sec-last >= aviGap {
			emit(base+int64(binary.LittleEndian.Uint32(idx[at+8:at+12])), sec)
			last = sec
		}
		frame++
	}
	a.done = frame > 0
}
