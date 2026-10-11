package timeindex

import "bytes"

// MPEG program streams — DVD rips as .vob, and plain .mpg — carry the presentation time in
// the header of every elementary stream packet, the same field MPEG-TS carries, just wrapped
// differently. Video packets are taken and the rest ignored, so the time read is the time of
// the picture rather than of whatever audio happened to be interleaved there.
//
// Like MPEG-TS the clock starts where the muxer left it, so the times are offset by a
// constant until Index.SetOrigin supplies it.

const psTail = 64 // enough to hold a packet header split across two reads

var psPrefix = []byte{0x00, 0x00, 0x01}

type mpegps struct {
	tail
}

func newMPEGPS() *mpegps { return &mpegps{} }

func (s *mpegps) reset() { s.drop() }

func (s *mpegps) feed(off int64, p []byte, emit func(int64, float64)) {
	buf, base := s.join(off, p)

	for i := 0; ; {
		hit := bytes.Index(buf[i:], psPrefix)
		if hit < 0 {
			break
		}
		at := i + hit
		if sec, ok := pesPTS(buf[at:]); ok {
			emit(base+int64(at), sec)
		}
		i = at + len(psPrefix)
	}

	s.hold(buf, base, psTail)
}

// pesPTS reads the presentation time out of a video packet whose start code has just been
// matched. Anything that is not a video stream, or carries no time, is passed over — as is
// a start code that turned out to be a coincidence inside picture data, which is what the
// shape checks on the header rule out.
func pesPTS(p []byte) (float64, bool) {
	if len(p) < 14 {
		return 0, false
	}
	stream := p[3]
	if stream < 0xE0 || stream > 0xEF { // video streams only
		return 0, false
	}
	if p[6]&0xC0 != 0x80 { // MPEG-2 packets mark themselves here
		return 0, false
	}
	if p[7]&0x80 == 0 { // no presentation time in this packet
		return 0, false
	}
	if int(p[8]) < 5 {
		return 0, false
	}
	t := p[9:14]
	// The high nibble is 0010 when only a presentation time follows and 0011 when a decode
	// time follows it, which is what a video packet normally carries.
	if t[0]&0xE0 != 0x20 || t[0]&0x01 == 0 || t[2]&0x01 == 0 || t[4]&0x01 == 0 {
		return 0, false // the marker bits the format puts between the digits
	}
	ticks := int64(t[0]&0x0E)<<29 |
		int64(t[1])<<22 |
		int64(t[2]&0xFE)<<14 |
		int64(t[3])<<7 |
		int64(t[4])>>1
	return float64(ticks) / 90000, true
}
