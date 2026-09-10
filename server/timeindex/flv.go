package timeindex

import "encoding/binary"

// FLV is a chain of tags, and every tag states its own time in milliseconds. Reading it is
// the easiest of the lot; the only work is finding where a tag starts when the stream was
// joined in the middle, which is done by walking from a candidate and checking that each tag
// lands exactly where the previous one said it would.

const (
	flvTagHeader = 11
	flvLockRun   = 4 // tags that must chain correctly before a boundary is believed
	flvTail      = 4 << 10
)

type flv struct {
	tail
	locked bool
	next   int64 // file offset the following tag starts at
}

func newFLV() *flv { return &flv{} }

func (f *flv) reset() {
	f.drop()
	f.locked = false
}

func (f *flv) feed(off int64, p []byte, emit func(int64, float64)) {
	buf, base := f.join(off, p)

	i := 0
	if !f.locked {
		start, ok := f.align(buf, base)
		if !ok {
			f.hold(buf, base, flvTail)
			return
		}
		i = start
	} else if f.next >= base {
		i = int(f.next - base)
	} else {
		f.locked = false
		f.hold(buf, base, flvTail)
		return
	}

	for i+flvTagHeader <= len(buf) {
		size, sec, ok := flvTag(buf[i:])
		if !ok {
			f.locked = false
			break
		}
		if buf[i] == 9 { // video
			emit(base+int64(i), sec)
		}
		i += size
	}
	if f.locked {
		f.next = base + int64(i)
	}
	i = min(i, len(buf))
	f.hold(buf[i:], base+int64(i), flvTail)
}

// align looks for a run of tags that chain into one another, which a coincidence in the
// middle of a video frame will not do.
func (f *flv) align(buf []byte, base int64) (int, bool) {
	if len(buf) >= 13 && string(buf[:3]) == "FLV" {
		// Counted as 64 bits: on a 32-bit build an offset past two gigabytes turns negative,
		// and a negative start reaches back off the front of the slice.
		start := int64(binary.BigEndian.Uint32(buf[5:9])) + 4 // header, then the first back-pointer
		if start >= flvTagHeader && start < int64(len(buf)) {
			f.locked, f.next = true, base+start
			return int(start), true
		}
	}
	for i := 0; i+flvTagHeader <= len(buf); i++ {
		at, runs := i, 0
		for runs < flvLockRun {
			size, _, ok := flvTag(buf[at:])
			if !ok || at+size+4 > len(buf) {
				break
			}
			// The four bytes after a tag repeat its length, which is what makes the chain
			// checkable without trusting a single header.
			if int(binary.BigEndian.Uint32(buf[at+size-4:at+size])) != size-4 {
				break
			}
			at += size
			runs++
		}
		if runs >= flvLockRun || (runs >= 2 && at+flvTagHeader > len(buf)) {
			f.locked, f.next = true, base+int64(i)
			return i, true
		}
	}
	return 0, false
}

// flvTag reads one tag header: its total length including the back-pointer that follows it,
// and the time it plays at.
func flvTag(p []byte) (size int, sec float64, ok bool) {
	if len(p) < flvTagHeader {
		return 0, 0, false
	}
	kind := p[0] &^ 0x20 // the top bits mark encryption, not a different tag
	if kind != 8 && kind != 9 && kind != 18 {
		return 0, 0, false
	}
	body := int(p[1])<<16 | int(p[2])<<8 | int(p[3])
	if body < 0 || body > 16<<20 {
		return 0, 0, false
	}
	ms := int64(p[7])<<24 | int64(p[4])<<16 | int64(p[5])<<8 | int64(p[6])
	return flvTagHeader + body + 4, float64(ms) / 1000, true
}
