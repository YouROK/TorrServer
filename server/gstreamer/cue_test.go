//go:build gst

package gstreamer

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

// ebmlID encodes a Matroska element id (the marker bits are part of the id).
func ebmlIDBytes(id uint64) []byte {
	var b []byte
	for v := id; v > 0; v >>= 8 {
		b = append([]byte{byte(v)}, b...)
	}
	return b
}

// ebmlSize encodes a data size with the shortest vint.
func ebmlSize(n uint64) []byte {
	for length := 1; length <= 8; length++ {
		max := uint64(1)<<(7*length) - 2 // all-ones is "unknown"
		if n <= max {
			out := make([]byte, length)
			v := n | uint64(1)<<(7*length)
			for i := length - 1; i >= 0; i-- {
				out[i] = byte(v)
				v >>= 8
			}
			return out
		}
	}
	panic("size too large")
}

func ebml(id uint64, payload ...[]byte) []byte {
	body := bytes.Join(payload, nil)
	out := append(ebmlIDBytes(id), ebmlSize(uint64(len(body)))...)
	return append(out, body...)
}

func ebmlUint(id, v uint64) []byte {
	var b []byte
	for ; v > 0; v >>= 8 {
		b = append([]byte{byte(v)}, b...)
	}
	if len(b) == 0 {
		b = []byte{0}
	}
	return ebml(id, b)
}

func trackEntry(number, trackType uint64) []byte {
	return ebml(trackEntryID, ebmlUint(trackNumberID, number), ebmlUint(trackTypeID, trackType))
}

func cuePoint(timeTicks, track uint64) []byte {
	return ebml(cuePointID,
		ebmlUint(cueTimeID, timeTicks),
		ebml(cueTrackPositionsID, ebmlUint(cueTrackID, track), ebmlUint(cueClusterPositionID, 4096)))
}

func seekEntry(target, position uint64) []byte {
	return ebml(seekID, ebml(seekEntryID, ebmlIDBytes(target)), ebmlUint(seekPositionID, position))
}

// seekEntry16 always writes a 2-byte position so a SeekHead has the same size
// whatever positions it points at.
func seekEntry16(target, position uint64) []byte {
	return ebml(seekID, ebml(seekEntryID, ebmlIDBytes(target)),
		ebml(seekPositionID, []byte{byte(position >> 8), byte(position)}))
}

func TestEBMLSizeHelperRoundTrip(t *testing.T) {
	for _, n := range []uint64{0, 1, 126, 127, 16382, 16383, 1 << 20, 1 << 40} {
		data := append(ebmlIDBytes(infoID), ebmlSize(n)...)
		data = append(data, make([]byte, 1)...) // readEBMLElement needs idLength+1 bytes
		el, ok := readEBMLElement(data, 0)
		if !ok || el.id != infoID || el.size != n || el.unknownSize {
			t.Errorf("size %d: %+v ok=%v", n, el, ok)
		}
	}
}

func TestEBMLVintLength(t *testing.T) {
	cases := []struct {
		first byte
		max   int
		want  int
		ok    bool
	}{
		{0x80, 4, 1, true},
		{0x40, 4, 2, true},
		{0x10, 4, 4, true},
		{0x08, 4, 5, false}, // an id may be at most 4 bytes
		{0x00, 8, 9, false}, // no marker bit at all
		{0x01, 8, 8, true},
	}
	for _, c := range cases {
		got, ok := ebmlVintLength(c.first, c.max)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("ebmlVintLength(%#x,%d) = %d,%v want %d,%v", c.first, c.max, got, ok, c.want, c.ok)
		}
	}
}

func TestReadEBMLElementRejectsTruncatedAndUnknown(t *testing.T) {
	if _, ok := readEBMLElement(nil, 0); ok {
		t.Error("empty")
	}
	if _, ok := readEBMLElement([]byte{0x1A, 0x45}, 0); ok {
		t.Error("id cut short")
	}
	if _, ok := readEBMLElement([]byte{0xAE}, 0); ok {
		t.Error("no size byte")
	}
	if _, ok := readEBMLElement([]byte{0xAE, 0x01, 0, 0}, 0); ok {
		t.Error("8-byte size cut short")
	}
	if _, ok := readEBMLElement([]byte{0xAE, 0x81}, -1); ok {
		t.Error("negative offset")
	}
	el, ok := readEBMLElement([]byte{0xAE, 0xFF}, 0) // 1-byte all-ones size
	if !ok || !el.unknownSize {
		t.Errorf("unknown size 1 byte: %+v %v", el, ok)
	}
	el, ok = readEBMLElement([]byte{0xAE, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, 0)
	if !ok || !el.unknownSize {
		t.Errorf("unknown size 8 bytes: %+v %v", el, ok)
	}
}

func TestReadEBMLUnsigned(t *testing.T) {
	data := ebmlUint(timestampScaleID, 1_000_000)
	el, _ := readEBMLElement(data, 0)
	if v, ok := readEBMLUnsigned(data, el); !ok || v != 1_000_000 {
		t.Fatalf("got %d %v", v, ok)
	}
	if _, ok := readEBMLUnsigned(data[:len(data)-1], el); ok {
		t.Error("element runs past the buffer")
	}
	if _, ok := readEBMLUnsigned(data, ebmlElement{dataOffset: 0, size: 9}); ok {
		t.Error("9-byte integers are invalid")
	}
	if _, ok := readEBMLUnsigned(data, ebmlElement{}); ok {
		t.Error("zero-size integer")
	}
}

func TestParseVideoTrackNumberPicksFirstVideo(t *testing.T) {
	tracks := ebml(tracksID, trackEntry(1, 2), trackEntry(2, 1), trackEntry(3, 1))
	el, _ := readEBMLElement(tracks, 0)
	if got := parseVideoTrackNumber(tracks, el); got != 2 {
		t.Errorf("video track = %d, want 2 (first type-1 entry)", got)
	}

	audioOnly := ebml(tracksID, trackEntry(1, 2))
	el, _ = readEBMLElement(audioOnly, 0)
	if got := parseVideoTrackNumber(audioOnly, el); got != 0 {
		t.Errorf("audio-only = %d", got)
	}
}

func TestParseTimestampScale(t *testing.T) {
	info := ebml(infoID, ebml(0x4489, []byte{0x40, 0x45}), ebmlUint(timestampScaleID, 500_000))
	el, _ := readEBMLElement(info, 0)
	scale := defaultTimestampScaleNS
	parseTimestampScale(info, el, &scale)
	if scale != 500_000 {
		t.Errorf("scale = %d", scale)
	}

	// Absent or zero keeps the default.
	for name, info := range map[string][]byte{
		"absent": ebml(infoID, ebml(0x4489, []byte{0x40, 0x45})),
		"zero":   ebml(infoID, ebmlUint(timestampScaleID, 0)),
	} {
		el, _ := readEBMLElement(info, 0)
		scale := defaultTimestampScaleNS
		parseTimestampScale(info, el, &scale)
		if scale != defaultTimestampScaleNS {
			t.Errorf("%s: scale = %d, want default", name, scale)
		}
	}
}

func TestParseSeekHeadFirstEntryWins(t *testing.T) {
	head := ebml(seekHeadID,
		seekEntry(cuesID, 9000),
		seekEntry(tracksID, 120),
		seekEntry(cuesID, 1),                      // duplicate target is ignored
		ebml(seekID, ebmlUint(seekPositionID, 5)), // no target
	)
	el, _ := readEBMLElement(head, 0)
	positions := map[uint64]uint64{}
	parseSeekHead(head, el, positions)
	want := map[uint64]uint64{cuesID: 9000, tracksID: 120}
	if !reflect.DeepEqual(positions, want) {
		t.Fatalf("positions = %v, want %v", positions, want)
	}
}

func TestParseCueTimeline(t *testing.T) {
	const scale = 1_000_000 // 1 tick = 1ms
	cues := ebml(cuesID,
		cuePoint(10_000, 1),
		cuePoint(0, 1),
		cuePoint(5_000, 1),
		cuePoint(5_000, 1),  // duplicate time collapses
		cuePoint(7_000, 2),  // another track
		cuePoint(99_000, 1), // beyond the file duration
	)
	el, _ := readEBMLElement(cues, 0)
	tl := parseCueTimeline(cues, el, 1, scale, 20_000*scale)
	if tl == nil {
		t.Fatal("nil timeline")
	}
	want := []CueSegment{
		{0, 5_000 * scale},
		{5_000 * scale, 10_000 * scale},
		{10_000 * scale, 20_000 * scale},
	}
	if !reflect.DeepEqual(tl.Segments, want) {
		t.Fatalf("segments = %+v, want %+v", tl.Segments, want)
	}
	if tl.MaxDurationNS != 10_000*scale || tl.TimestampScaleNS != scale {
		t.Errorf("max=%d scale=%d", tl.MaxDurationNS, tl.TimestampScaleNS)
	}
	if s, ok := tl.Segment(1); !ok || s.DurationNS() != 5_000*scale {
		t.Errorf("Segment(1) = %+v %v", s, ok)
	}
	if _, ok := tl.Segment(3); ok {
		t.Error("Segment past the end")
	}
	if _, ok := tl.Segment(-1); ok {
		t.Error("Segment(-1)")
	}
	var nilTL *CueTimeline
	if _, ok := nilTL.Segment(0); ok {
		t.Error("nil timeline")
	}
}

func TestParseCueTimelineNeedsTwoVideoCues(t *testing.T) {
	cues := ebml(cuesID, cuePoint(0, 1), cuePoint(5, 2))
	el, _ := readEBMLElement(cues, 0)
	if tl := parseCueTimeline(cues, el, 1, 1_000_000, 1_000_000_000); tl != nil {
		t.Fatalf("one video cue gave %+v", tl)
	}
}

func TestParseCueTimelineOverflowIsSkipped(t *testing.T) {
	// cueTime * scale would wrap around uint64 and look like an early cue.
	cues := ebml(cuesID, cuePoint(0, 1), cuePoint(1<<62, 1), cuePoint(1000, 1))
	el, _ := readEBMLElement(cues, 0)
	tl := parseCueTimeline(cues, el, 1, 1_000_000, 10_000*1_000_000)
	if tl == nil {
		t.Fatal("nil")
	}
	for _, s := range tl.Segments {
		if s.EndNS > 10_000*1_000_000 {
			t.Errorf("segment past duration: %+v", s)
		}
	}
	if len(tl.Segments) != 2 {
		t.Errorf("segments = %+v, want the 0 and 1000 cues only", tl.Segments)
	}
}

func TestParseCueTimelineTruncatedCuePoint(t *testing.T) {
	cues := ebml(cuesID, cuePoint(0, 1), cuePoint(5, 1))
	// Cut inside the last cue point but keep the parent's declared size.
	el, _ := readEBMLElement(cues, 0)
	if tl := parseCueTimeline(cues[:len(cues)-3], el, 1, 1_000_000, 1<<40); tl != nil {
		t.Fatalf("truncated cues gave %+v", tl)
	}
}

func TestParseMatroskaPrefixStopsAtUnknownSizeCluster(t *testing.T) {
	prefix := append(
		ebml(seekHeadID, seekEntry(cuesID, 777)),
		ebml(infoID, ebmlUint(timestampScaleID, 250_000))...)
	prefix = append(prefix, ebml(tracksID, trackEntry(3, 1))...)
	// An unknown-size Cluster header: everything after it is unparseable.
	prefix = append(prefix, 0x1F, 0x43, 0xB6, 0x75, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF)

	positions := map[uint64]uint64{}
	scale := defaultTimestampScaleNS
	var video uint64
	parseMatroskaPrefix(prefix, 0, positions, &scale, &video)
	if positions[cuesID] != 777 || scale != 250_000 || video != 3 {
		t.Fatalf("positions=%v scale=%d video=%d", positions, scale, video)
	}
}

func TestFindMatroskaSegment(t *testing.T) {
	header := ebml(ebmlID, ebml(0x4282, []byte("matroska")))
	seg := ebml(matroskaSegmentID, ebml(infoID))
	s, ok := findMatroskaSegment(append(append([]byte{}, header...), seg...))
	if !ok || s.id != matroskaSegmentID || s.offset != len(header) {
		t.Fatalf("segment %+v ok=%v", s, ok)
	}
	if _, ok := findMatroskaSegment(seg); ok {
		t.Error("data without an EBML header")
	}
	if _, ok := findMatroskaSegment(header); ok {
		t.Error("EBML header without a segment")
	}
	if _, ok := findMatroskaSegment(append(append([]byte{}, header...), ebml(infoID)...)); ok {
		t.Error("second element is not a segment")
	}
	if _, ok := findMatroskaSegment(header[:len(header)-2]); ok {
		t.Error("truncated EBML header")
	}
}

// rangeServer serves body with real Range semantics and counts requests.
func rangeServer(t *testing.T, body []byte, hits *atomic.Int32) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits != nil {
			hits.Add(1)
		}
		http.ServeContent(w, r, "x.mkv", time.Time{}, bytes.NewReader(body))
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestReadHTTPRange(t *testing.T) {
	body := []byte("0123456789abcdef")
	url := rangeServer(t, body, nil)
	ctx := t.Context()

	got, err := readHTTPRange(ctx, url, 4, 6)
	if err != nil || string(got) != "456789" {
		t.Fatalf("got %q, %v", got, err)
	}
	// Starting at 0 a server may answer 200 with the whole body.
	whole := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	defer whole.Close()
	if got, err := readHTTPRange(ctx, whole.URL, 0, 4); err != nil || string(got) != "0123" {
		t.Fatalf("200 at offset 0: %q, %v", got, err)
	}
	// ...but a 200 for offset>0 means the server ignored Range: refuse it.
	if _, err := readHTTPRange(ctx, whole.URL, 4, 4); err == nil {
		t.Error("200 for a nonzero offset must fail")
	}
	// Short body.
	if _, err := readHTTPRange(ctx, url, 10, 100); err == nil {
		t.Error("range past EOF must fail")
	}
	for _, c := range []struct {
		off int64
		n   int
	}{{-1, 1}, {0, 0}, {0, -5}} {
		if _, err := readHTTPRange(ctx, url, c.off, c.n); err == nil {
			t.Errorf("invalid range %d/%d accepted", c.off, c.n)
		}
	}
	if _, err := readHTTPRange(ctx, "://bad", 0, 1); err == nil {
		t.Error("bad url")
	}
}

func TestReadHTTPRangeRejectsMismatchedContentRange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Range", "bytes 0-3/16")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write([]byte("0123"))
	}))
	defer srv.Close()
	if _, err := readHTTPRange(t.Context(), srv.URL, 8, 4); err == nil {
		t.Fatal("206 whose Content-Range starts elsewhere must be rejected")
	}
}

// buildMKV returns: EBML header, Segment{SeekHead, Info, Tracks, Cues}, with
// the SeekHead positions relative to the Segment data as the format requires.
func buildMKV(cuePoints ...[]byte) []byte {
	header := ebml(ebmlID, ebml(0x4282, []byte("matroska")))
	info := ebml(infoID, ebmlUint(timestampScaleID, 1_000_000))
	tracks := ebml(tracksID, trackEntry(1, 1), trackEntry(2, 2))
	cues := ebml(cuesID, cuePoints...)

	seekHead := func(infoPos, tracksPos, cuesPos uint64) []byte {
		return ebml(seekHeadID,
			seekEntry16(infoID, infoPos), seekEntry16(tracksID, tracksPos), seekEntry16(cuesID, cuesPos))
	}
	seekLen := uint64(len(seekHead(0, 0, 0)))
	infoPos := seekLen
	tracksPos := infoPos + uint64(len(info))
	cuesPos := tracksPos + uint64(len(tracks))
	segment := ebml(matroskaSegmentID, seekHead(infoPos, tracksPos, cuesPos), info, tracks, cues)
	return append(header, segment...)
}

func TestReadMatroskaCueTimelineEndToEnd(t *testing.T) {
	file := buildMKV(cuePoint(0, 1), cuePoint(4_000, 1), cuePoint(9_000, 1), cuePoint(2_000, 2))
	var hits atomic.Int32
	url := rangeServer(t, file, &hits)

	const dur = 12_000 * 1_000_000
	tl := readMatroskaCueTimeline(url, int64(len(file)), dur)
	if tl == nil {
		t.Fatal("nil timeline")
	}
	want := []CueSegment{{0, 4_000_000_000}, {4_000_000_000, 9_000_000_000}, {9_000_000_000, dur}}
	if !reflect.DeepEqual(tl.Segments, want) {
		t.Fatalf("segments = %+v, want %+v", tl.Segments, want)
	}

	if readMatroskaCueTimeline("", int64(len(file)), dur) != nil {
		t.Error("empty url")
	}
	if readMatroskaCueTimeline(url, int64(len(file)), 0) != nil {
		t.Error("zero duration")
	}
}

func TestReadMatroskaCueTimelineNotMatroska(t *testing.T) {
	url := rangeServer(t, bytes.Repeat([]byte("mp4 is not mkv "), 100), nil)
	if readMatroskaCueTimeline(url, 1500, 1e9) != nil {
		t.Fatal("non-matroska source gave a timeline")
	}
}

func TestReadMatroskaCueTimelineNoVideoTrack(t *testing.T) {
	// Tracks that hold audio only: no video cues can be matched.
	header := ebml(ebmlID, ebml(0x4282, []byte("matroska")))
	tracks := ebml(tracksID, trackEntry(1, 2))
	cues := ebml(cuesID, cuePoint(0, 1), cuePoint(5, 1))
	file := append(header, ebml(matroskaSegmentID, tracks, cues)...)
	url := rangeServer(t, file, nil)
	if readMatroskaCueTimeline(url, int64(len(file)), 1e10) != nil {
		t.Fatal("audio-only file gave a timeline")
	}
}

func TestReadMatroskaElementGuards(t *testing.T) {
	file := buildMKV(cuePoint(0, 1), cuePoint(10, 1))
	url := rangeServer(t, file, nil)
	ctx := t.Context()
	segment, _ := findMatroskaSegment(file)
	segOff := int64(segment.dataOffset)

	// Find the real Cues position through the SeekHead.
	positions := map[uint64]uint64{}
	var scale, video uint64
	parseMatroskaPrefix(file, segment.dataOffset, positions, &scale, &video)
	cuesPos := positions[cuesID]

	if data := readMatroskaElement(ctx, url, segOff, cuesPos, cuesID, 1<<20, int64(len(file))); len(data) == 0 {
		t.Fatal("valid cues not read")
	}
	if readMatroskaElement(ctx, url, segOff, cuesPos, tracksID, 1<<20, int64(len(file))) != nil {
		t.Error("wrong element id accepted")
	}
	if readMatroskaElement(ctx, url, segOff, cuesPos, cuesID, 4, int64(len(file))) != nil {
		t.Error("element larger than the cap accepted")
	}
	if readMatroskaElement(ctx, url, segOff, uint64(len(file))+10, cuesID, 1<<20, int64(len(file))) != nil {
		t.Error("position past EOF accepted")
	}
	if readMatroskaElement(ctx, url, segOff, ^uint64(0), cuesID, 1<<20, 0) != nil {
		t.Error("overflowing position accepted")
	}
	if readMatroskaElement(ctx, url, -1, 0, cuesID, 1<<20, 0) != nil {
		t.Error("negative segment offset accepted")
	}
	// A content length shorter than the real element must stop the read.
	if readMatroskaElement(ctx, url, segOff, cuesPos, cuesID, 1<<20, int64(segOff)+int64(cuesPos)+4) != nil {
		t.Error("element extending past contentLength accepted")
	}
}
