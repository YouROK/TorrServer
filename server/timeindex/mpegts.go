package timeindex

// MPEG-TS is a stream of fixed-size packets, and some of them carry a program clock
// reference — the decoder's wall clock at that point in the stream. Packets are 188 bytes,
// or 192 in the Blu-ray flavour, where each is prefixed by four bytes of arrival time. A
// stream joined mid-file has to find the packet boundary first, which is done by requiring
// the sync byte to repeat at the expected stride.
//
// The clock does not start at zero — it starts wherever the muxer left it — so the times are
// offset by a constant until Index.SetOrigin supplies it.

const (
	tsSync    = 0x47
	tsPacket  = 188
	m2tsExtra = 4
	tsLockRun = 5 // packets that must line up before a boundary is believed
)

type mpegts struct {
	tail
	stride int // 188 or 192, zero until the flavour is known
	skew   int // bytes before the sync byte inside a packet
}

func newMPEGTS() *mpegts { return &mpegts{} }

func (t *mpegts) reset() { t.drop() }

func (t *mpegts) feed(off int64, p []byte, emit func(int64, float64)) {
	buf, base := t.join(off, p)

	start, ok := t.align(buf)
	if !ok {
		t.hold(buf, base, tsPacket*2)
		return
	}

	i := start
	for ; i+t.stride <= len(buf); i += t.stride {
		pkt := buf[i+t.skew : i+t.skew+tsPacket]
		if pkt[0] != tsSync {
			// The lock is stale — a discontinuity or a bad guess. Look for it again.
			if start, ok = t.align(buf[i:]); !ok {
				break
			}
			i += start - t.stride
			continue
		}
		if sec, ok := packetPCR(pkt); ok {
			emit(base+int64(i), sec)
		}
	}
	t.hold(buf[i:], base+int64(i), tsPacket*2)
}

// align finds where a packet starts, and which of the two packet sizes is in use. A single
// sync byte means nothing in the middle of video data, so several in a row are required.
func (t *mpegts) align(buf []byte) (int, bool) {
	for i := 0; i < len(buf); i++ {
		for _, stride := range []int{tsPacket, tsPacket + m2tsExtra} {
			if !syncRun(buf, i, stride) {
				continue
			}
			t.stride = stride
			t.skew = 0
			// In the Blu-ray flavour the four extra bytes come first, so the packet the
			// sync byte belongs to starts earlier — unless this is the file's first one.
			if stride == tsPacket+m2tsExtra && i >= m2tsExtra {
				t.skew = m2tsExtra
				return i - m2tsExtra, true
			}
			return i, true
		}
	}
	return 0, false
}

func syncRun(buf []byte, at, stride int) bool {
	for n := 0; n < tsLockRun; n++ {
		pos := at + n*stride
		if pos >= len(buf) {
			return n >= 2 // near the end of the buffer, take what lined up
		}
		if buf[pos] != tsSync {
			return false
		}
	}
	return true
}

// packetPCR reads the program clock reference out of a packet's adaptation field.
func packetPCR(pkt []byte) (float64, bool) {
	const (
		hasAdaptation = 0x20
		pcrFlag       = 0x10
	)
	if pkt[1]&0x80 != 0 { // transport error
		return 0, false
	}
	if pkt[3]&hasAdaptation == 0 {
		return 0, false
	}
	length := int(pkt[4])
	if length < 7 || 5+length > len(pkt) {
		return 0, false
	}
	if pkt[5]&pcrFlag == 0 {
		return 0, false
	}
	f := pkt[6:12]
	base := int64(f[0])<<25 | int64(f[1])<<17 | int64(f[2])<<9 | int64(f[3])<<1 | int64(f[4])>>7
	ext := int64(f[4]&0x01)<<8 | int64(f[5])
	return float64(base)/90000 + float64(ext)/27000000, true
}
