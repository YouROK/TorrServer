package timeindex

import "bytes"

// Matroska writes the picture's time in the header of every cluster, as a count of
// TimestampScale ticks since the start of the segment — an absolute time, not a delta, so a
// cluster read in the middle of a file says on its own where playback is. Clusters are a few
// seconds apart, which is the resolution this gives.
//
// The elements are EBML: an id, a length, then the payload. Cluster is 1F 43 B6 75 and its
// Timestamp child is E7. Both are found by scanning, since a stream joined mid-file has no
// element tree to walk down, and a match is confirmed by requiring the timestamp to sit
// where the format says it does.

const (
	clusterTail  = 96        // bytes held back so an element split across two reads is still seen
	defaultScale = 1000000   // ns per tick, the value virtually every muxer writes
	maxCluster   = 512 << 20 // a stated length past this came from a stray match, not a cluster
	headerReach  = 4 << 20   // how far into the file the segment header can reasonably sit
)

var (
	clusterID = []byte{0x1F, 0x43, 0xB6, 0x75}
	infoID    = []byte{0x15, 0x49, 0xA9, 0x66}
	scaleID   = []byte{0x2A, 0xD7, 0xB1}
)

type matroska struct {
	tail
	scale     int64 // nanoseconds per timestamp tick
	scaleRead bool  // taken from the file, whatever the value turned out to be

	// Clusters chain: each states its own length, so the next one has to begin exactly where
	// the previous one ends. Over tens of gigabytes a four-byte pattern turns up inside frame
	// data by chance, and such a stray never lands on that boundary — which is the difference
	// between an index that can be trusted and one that reports the fourth minute of a film
	// two hours in. A cluster is therefore held back until the following one confirms it.
	next    int64 // where the next cluster must start
	pendOff int64
	pendSec float64
	pending bool
	locked  bool // the chain has been confirmed once, so a cluster on the boundary stands alone
	lastSec float64
}

func newMatroska() *matroska { return &matroska{scale: defaultScale} }

func (m *matroska) reset() {
	m.drop()
	m.pending = false
	m.locked = false
	m.next = 0
}

func (m *matroska) feed(off int64, p []byte, emit func(int64, float64)) {
	buf, base := m.join(off, p)

	// TimestampScale lives in the segment header, at the very front of the file. Looking for
	// it anywhere else means matching three bytes against picture data, and a match there
	// scales every timestamp in the file by whatever number happened to follow — which is how
	// a two-hour film came to report sixty-eight hours.
	if base < headerReach && !m.scaleRead {
		m.readScale(buf)
	}

	for i := 0; ; {
		hit := bytes.Index(buf[i:], clusterID)
		if hit < 0 {
			break
		}
		at := i + hit
		i = at + len(clusterID)

		ticks, body, ok := clusterTicks(buf[at+len(clusterID):])
		if !ok {
			continue
		}
		off := base + int64(at)
		sec := float64(ticks) * float64(m.scale) / 1e9
		if sec+1 < m.lastSec {
			continue // a cluster cannot play before the one in front of it
		}

		if body < 0 {
			// A cluster of unknown length, which live muxers write. There is no boundary to
			// check the next one against, so it stands on the timestamp alone.
			emit(off, sec)
			m.lastSec, m.pending = sec, false
			continue
		}
		end := off + int64(len(clusterID)) + body

		if m.pending || m.locked {
			switch {
			case off == m.next:
				// Lands exactly where the previous cluster said it would.
				if m.pending {
					emit(m.pendOff, m.pendSec)
					m.lastSec, m.pending, m.locked = m.pendSec, false, true
				}
				if m.locked {
					emit(off, sec)
					m.lastSec = sec
					m.next = end
					continue
				}
			case off < m.next:
				continue // inside the cluster being read: a stray match in frame data
			default:
				m.locked = false // the chain broke, start it again from here
			}
		}
		// Not on a confirmed boundary: hold this one until the next cluster points back at it.
		m.pendOff, m.pendSec, m.next, m.pending = off, sec, end, true
	}

	m.hold(buf, base, clusterTail)
}

// readScale picks up TimestampScale if the segment header happens to go past. Without it the
// default stands, which is what every muxer in practice writes anyway.
//
// It is only looked for inside the Info element, where the format puts it. The header ends a
// few kilobytes into the file and the stretch searched runs to four megabytes, so nearly all
// of what is searched is picture data, and three bytes match there by chance: measured on a
// 4K episode, the real scale sat at byte 4156 and a stray at byte 2851054 read 12523939, which
// multiplied every timestamp in the film by twelve and a half. The stray was believed because
// the real value equalled the default, which used to pass for "not read yet".
func (m *matroska) readScale(buf []byte) {
	hit := bytes.Index(buf, infoID)
	if hit < 0 {
		return
	}
	info := buf[hit+len(infoID):]
	size, n, ok := readVint(info, true)
	if !ok || size < 0 {
		return
	}
	info = info[n:]
	if size < int64(len(info)) {
		info = info[:size]
	}
	hit = bytes.Index(info, scaleID)
	if hit < 0 {
		return
	}
	rest := info[hit+len(scaleID):]
	size, n, ok = readVint(rest, true)
	if !ok || size < 1 || size > 8 || int64(len(rest)) < int64(n)+size {
		return
	}
	var value int64
	for _, b := range rest[n : int64(n)+size] {
		value = value<<8 | int64(b)
	}
	if value >= 1000 && value <= 1e9 {
		m.scale, m.scaleRead = value, true
	}
}

// clusterTicks reads the Timestamp out of a cluster whose id has just been matched, and the
// length of everything that follows the id, so the caller knows where the next cluster has to
// start. A length of -1 is the unknown-size form. The timestamp is the first child the format
// allows to carry a value, so anything else in front of it means the id was a coincidence
// inside frame data and the match is dropped.
func clusterTicks(rest []byte) (ticks, body int64, ok bool) {
	size, n, ok := readVint(rest, true) // cluster length, may be the unknown-size form
	if !ok {
		return 0, 0, false
	}
	body = -1
	if size >= 0 {
		if size > maxCluster {
			return 0, 0, false
		}
		body = int64(n) + size
	}
	// CRC-32 and Void are allowed in front of anything, and muxers do put a checksum there,
	// so they are stepped over. Anything else means the id was a coincidence inside frame
	// data and the match is dropped.
	children := rest[n:]
	for len(children) > 0 {
		if children[0] == 0xBF || children[0] == 0xEC {
			size, sn, ok := readVint(children[1:], true)
			if !ok || size < 0 || int64(len(children)) < int64(1+sn)+size {
				return 0, 0, false
			}
			children = children[int64(1+sn)+size:]
			continue
		}
		if children[0] != 0xE7 {
			return 0, 0, false
		}
		size, sn, ok := readVint(children[1:], true)
		if !ok || size < 1 || size > 8 || int64(len(children)) < int64(1+sn)+size {
			return 0, 0, false
		}
		for _, b := range children[1+sn : int64(1+sn)+size] {
			ticks = ticks<<8 | int64(b)
		}
		return ticks, body, true
	}
	return 0, 0, false
}

// readVint reads an EBML variable-length integer. Sizes carry a length marker that is not
// part of the value and is stripped; ids keep theirs. The unknown-size form (all value bits
// set) returns -1, which callers only ever use for its length.
func readVint(p []byte, stripMarker bool) (value int64, length int, ok bool) {
	if len(p) == 0 || p[0] == 0 {
		return 0, 0, false
	}
	length = 1
	for mask := byte(0x80); p[0]&mask == 0; mask >>= 1 {
		length++
	}
	if length > 8 || len(p) < length {
		return 0, 0, false
	}
	first := p[0]
	if stripMarker {
		first &^= byte(0x80) >> (length - 1)
	}
	value = int64(first)
	unknown := first == byte(0x80)>>(length-1)-1
	for _, b := range p[1:length] {
		value = value<<8 | int64(b)
		unknown = unknown && b == 0xFF
	}
	if unknown {
		return -1, length, true
	}
	return value, length, true
}
