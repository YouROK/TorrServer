package timeindex

import (
	"bytes"
	"encoding/binary"
)

// MP4 comes in two shapes. A fragmented file states the time at the head of every fragment,
// the same way Matroska does. A progressive one says nothing in the media data at all: its
// table of sample sizes, durations and chunk offsets lives in the moov atom. Both are read
// here, because both go past — a player has to read moov before it can start, whether it
// then plays from the beginning or seeks.
//
// Boxes are a length and a four-character type. A stream joined mid-file has no tree to walk,
// so the two top-level boxes worth having are found by name and confirmed by their length.

const (
	mp4MaxBox  = 64 << 20 // a moov atom bigger than this is not worth holding in memory
	mp4BoxTail = 16       // enough to span a box header split across two reads
)

type mp4 struct {
	tail
	box   collector // a box being collected until it is complete
	scale uint32    // media timescale of the track the offsets belong to
	table bool      // the progressive table has been read, nothing more to do
}

func newMP4() *mp4 { return &mp4{} }

func (m *mp4) reset() {
	m.drop()
	m.box.reset()
}

func (m *mp4) feed(off int64, p []byte, emit func(int64, float64)) {
	if m.box.collecting() {
		n, done := m.box.take(p)
		if !done {
			return // still filling
		}
		box, at := m.box.finish()
		m.complete(emit, box, at)
		off, p = off+int64(n), p[n:]
		if len(p) == 0 {
			return
		}
	}

	buf, base := m.join(off, p)

	// Every byte of the media data goes through here, so the two box names are looked for
	// with a pattern search rather than a comparison at every offset.
	for i := 4; i+4 <= len(buf); i++ {
		hit := nextBox(buf, i)
		if hit < 0 {
			break
		}
		i = hit
		if string(buf[i:i+4]) == "moov" && m.table {
			continue
		}
		size, header := boxSize(buf[i-4:])
		if size < int64(header)+8 || size > mp4MaxBox {
			continue
		}
		start := i - 4
		if int64(len(buf)-start) >= size {
			m.complete(emit, buf[start:start+int(size)], base+int64(start))
			i = start + int(size)
			continue
		}
		// The box runs past what has been read; collect the rest as it arrives.
		m.box.start(base+int64(start), buf[start:], size)
		m.drop()
		return
	}

	m.hold(buf, base, mp4BoxTail)
}

// nextBox finds the next moov or moof name at or after from, leaving room for the length
// that precedes it.
func nextBox(buf []byte, from int) int {
	best := -1
	for _, name := range [][]byte{boxMoov, boxMoof} {
		if hit := bytes.Index(buf[from:], name); hit >= 0 && (best < 0 || from+hit < best) {
			best = from + hit
		}
	}
	if best < 4 {
		return -1
	}
	return best
}

var (
	boxMoov = []byte("moov")
	boxMoof = []byte("moof")
)

// complete reads a box that is in hand in full, starting at file offset at.
func (m *mp4) complete(emit func(int64, float64), box []byte, at int64) {
	if len(box) < 8 {
		return
	}
	switch string(box[4:8]) {
	case "moov":
		m.readMoov(box[8:], at, emit)
	case "moof":
		m.readMoof(box[8:], at, emit)
	}
}

// readMoof takes the fragment's decode time from its first track fragment.
func (m *mp4) readMoof(body []byte, at int64, emit func(int64, float64)) {
	if m.scale == 0 {
		return // no timescale yet, the numbers would mean nothing
	}
	traf := findBox(body, "traf")
	tfdt := findBox(traf, "tfdt")
	if len(tfdt) < 8 {
		return
	}
	var decode uint64
	if tfdt[0] == 1 {
		if len(tfdt) < 12 {
			return
		}
		decode = binary.BigEndian.Uint64(tfdt[4:12])
	} else {
		decode = uint64(binary.BigEndian.Uint32(tfdt[4:8]))
	}
	emit(at, float64(decode)/float64(m.scale))
}

// readMoov reads the timescale, and for a progressive file the whole chunk table: where each
// run of samples sits in the file and what time it starts at.
func (m *mp4) readMoov(body []byte, _ int64, emit func(int64, float64)) {
	trak := videoTrack(body)
	if trak == nil {
		return
	}
	mdia := findBox(trak, "mdia")
	mdhd := findBox(mdia, "mdhd")
	if len(mdhd) < 20 {
		return
	}
	if mdhd[0] == 1 {
		if len(mdhd) < 28 {
			return
		}
		m.scale = binary.BigEndian.Uint32(mdhd[20:24])
	} else {
		m.scale = binary.BigEndian.Uint32(mdhd[12:16])
	}
	if m.scale == 0 {
		return
	}

	stbl := findBox(findBox(mdia, "minf"), "stbl")
	durations := readSTTS(findBox(stbl, "stts"))
	perChunk := readSTSC(findBox(stbl, "stsc"))
	offsets := readChunkOffsets(stbl)
	sizes, fixedSize := readSTSZ(findBox(stbl, "stsz"))
	if len(durations) == 0 || len(perChunk) == 0 || len(offsets) == 0 {
		return
	}

	// Chunks are laid out by size, around a megabyte each, which in a quiet stretch can be
	// half a minute of picture. The sample sizes say where inside a chunk each frame starts,
	// so the table is built frame by frame and thinned to a fixed rate instead.
	const minGap = 0.5 // seconds between entries, once a chunk has been entered

	var sampleIndex, ticks int64
	run, sampleAt := 0, 0
	for chunk := 0; chunk < len(offsets); chunk++ {
		for run+1 < len(perChunk) && chunk+1 >= perChunk[run+1].firstChunk {
			run++
		}
		at, lastEmit := offsets[chunk], -minGap
		for n := 0; n < perChunk[run].samples; n++ {
			for sampleAt < len(durations) && sampleIndex >= durations[sampleAt].through {
				sampleAt++
			}
			if sampleAt >= len(durations) {
				break
			}
			sec := float64(ticks) / float64(m.scale)
			if n == 0 || sec-lastEmit >= minGap {
				emit(at, sec)
				lastEmit = sec
			}
			at += sampleSize(sizes, fixedSize, sampleIndex)
			ticks += durations[sampleAt].delta
			sampleIndex++
		}
	}
	m.table = true
}

// readSTSZ returns the per-sample sizes, or a single size shared by every sample.
func readSTSZ(box []byte) ([]uint32, int64) {
	if len(box) < 12 {
		return nil, 0
	}
	if fixed := binary.BigEndian.Uint32(box[4:8]); fixed != 0 {
		return nil, int64(fixed)
	}
	count := int(binary.BigEndian.Uint32(box[8:12]))
	body := box[12:]
	out := make([]uint32, 0, count)
	for i := 0; i < count && (i+1)*4 <= len(body); i++ {
		out = append(out, binary.BigEndian.Uint32(body[i*4:i*4+4]))
	}
	return out, 0
}

func sampleSize(sizes []uint32, fixed, index int64) int64 {
	if fixed > 0 {
		return fixed
	}
	if index < 0 || index >= int64(len(sizes)) {
		return 0
	}
	return int64(sizes[index])
}

type sttsRun struct {
	through int64 // sample index this run ends before
	delta   int64
}

type stscRun struct {
	firstChunk int
	samples    int
}

func readSTTS(box []byte) []sttsRun {
	if len(box) < 8 {
		return nil
	}
	count := int(binary.BigEndian.Uint32(box[4:8]))
	body := box[8:]
	runs := make([]sttsRun, 0, count)
	var through int64
	for i := 0; i < count && (i+1)*8 <= len(body); i++ {
		n := int64(binary.BigEndian.Uint32(body[i*8 : i*8+4]))
		through += n
		runs = append(runs, sttsRun{through: through, delta: int64(binary.BigEndian.Uint32(body[i*8+4 : i*8+8]))})
	}
	return runs
}

func readSTSC(box []byte) []stscRun {
	if len(box) < 8 {
		return nil
	}
	count := int(binary.BigEndian.Uint32(box[4:8]))
	body := box[8:]
	runs := make([]stscRun, 0, count)
	for i := 0; i < count && (i+1)*12 <= len(body); i++ {
		runs = append(runs, stscRun{
			firstChunk: int(binary.BigEndian.Uint32(body[i*12 : i*12+4])),
			samples:    int(binary.BigEndian.Uint32(body[i*12+4 : i*12+8])),
		})
	}
	return runs
}

func readChunkOffsets(stbl []byte) []int64 {
	if box := findBox(stbl, "stco"); len(box) >= 8 {
		count := int(binary.BigEndian.Uint32(box[4:8]))
		body := box[8:]
		out := make([]int64, 0, count)
		for i := 0; i < count && (i+1)*4 <= len(body); i++ {
			out = append(out, int64(binary.BigEndian.Uint32(body[i*4:i*4+4])))
		}
		return out
	}
	if box := findBox(stbl, "co64"); len(box) >= 8 {
		count := int(binary.BigEndian.Uint32(box[4:8]))
		body := box[8:]
		out := make([]int64, 0, count)
		for i := 0; i < count && (i+1)*8 <= len(body); i++ {
			out = append(out, int64(binary.BigEndian.Uint64(body[i*8:i*8+8])))
		}
		return out
	}
	return nil
}

// videoTrack returns the trak whose handler says it carries pictures — the one whose chunk
// offsets account for nearly all of the file.
func videoTrack(moov []byte) []byte {
	for rest := moov; len(rest) >= 8; {
		size, header := boxSize(rest)
		if !stepsForward(size, header, len(rest)) {
			return nil
		}
		if string(rest[4:8]) == "trak" {
			trak := rest[header:size]
			hdlr := findBox(findBox(trak, "mdia"), "hdlr")
			if len(hdlr) >= 12 && string(hdlr[8:12]) == "vide" {
				return trak
			}
		}
		rest = rest[size:]
	}
	return nil
}

// findBox returns the body of the first child box of the given type, without recursing.
func findBox(parent []byte, kind string) []byte {
	for rest := parent; len(rest) >= 8; {
		size, header := boxSize(rest)
		if !stepsForward(size, header, len(rest)) {
			return nil
		}
		if string(rest[4:8]) == kind {
			return rest[header:size]
		}
		rest = rest[size:]
	}
	return nil
}

// stepsForward says whether a box header can be walked past. A header that cannot be read at
// all comes back as zero length, and a walk that treats that as a box makes no progress — the
// slice is resliced from nothing and the loop spins for ever. Since these boxes are found by
// searching for their names, a stray match inside picture data lands here with arbitrary
// bytes behind it, and that spin runs under the index lock inside a read.
func stepsForward(size int64, header, left int) bool {
	return header > 0 && size >= int64(header) && size <= int64(left)
}

// boxSize reads a box header, which states a 32-bit length, or one means the real length
// follows as 64 bits.
func boxSize(p []byte) (size int64, header int) {
	if len(p) < 8 {
		return 0, 0
	}
	size = int64(binary.BigEndian.Uint32(p[0:4]))
	header = 8
	if size == 1 {
		if len(p) < 16 {
			return 0, 0
		}
		size = int64(binary.BigEndian.Uint64(p[8:16]))
		header = 16
	}
	return size, header
}
