//go:build gst

package gstreamer

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func ibox(typ string, payload ...[]byte) []byte {
	body := bytes.Join(payload, nil)
	out := make([]byte, 8, 8+len(body))
	binary.BigEndian.PutUint32(out, uint32(8+len(body)))
	copy(out[4:], typ)
	return append(out, body...)
}

func hdlrBox(handler string) []byte {
	p := make([]byte, 12)
	copy(p[8:], handler)
	return ibox("hdlr", p)
}

func stsdBox(entries ...[]byte) []byte {
	head := make([]byte, 8)
	binary.BigEndian.PutUint32(head[4:], uint32(len(entries)))
	return ibox("stsd", append([][]byte{head}, entries...)...)
}

func trakBox(handler string, entries ...[]byte) []byte {
	return ibox("trak", ibox("mdia", hdlrBox(handler), ibox("minf", ibox("stbl", stsdBox(entries...)))))
}

func videoEntry(typ string, w, h uint16, children ...[]byte) []byte {
	fixed := make([]byte, 78)
	binary.BigEndian.PutUint16(fixed[24:], w)
	binary.BigEndian.PutUint16(fixed[26:], h)
	return ibox(typ, append([][]byte{fixed}, children...)...)
}

func audioEntry(version uint16, children ...[]byte) []byte {
	fixed := make([]byte, 28)
	binary.BigEndian.PutUint16(fixed[8:], version)
	switch version {
	case 1:
		fixed = append(fixed, make([]byte, 16)...)
	case 2:
		fixed = append(fixed, make([]byte, 36)...)
	}
	return ibox("mp4a", append([][]byte{fixed}, children...)...)
}

// esdsAACLC carries a DecSpecificInfo (tag 0x05) with objectType 2 (AAC-LC).
func esdsAAC(audioSpecific ...byte) []byte {
	p := []byte{0, 0, 0, 0, 0x03, 0x19, 0, 1, 0, 0x04, 0x11, 0x40, 0x15}
	p = append(p, make([]byte, 13)...)
	p = append(p, 0x05, byte(len(audioSpecific)))
	p = append(p, audioSpecific...)
	return ibox("esds", p)
}

func colrBox(transfer uint16) []byte {
	p := []byte("nclx")
	p = binary.BigEndian.AppendUint16(p, 1)
	p = binary.BigEndian.AppendUint16(p, transfer)
	p = binary.BigEndian.AppendUint16(p, 1)
	p = append(p, 0)
	return ibox("colr", p)
}

func initSegment(traks ...[]byte) []byte {
	return append(ibox("ftyp", []byte("isom\x00\x00\x02\x00")), ibox("moov", traks...)...)
}

func TestReadMP4InitInfoCodecs(t *testing.T) {
	avcC := ibox("avcC", []byte{1, 0x64, 0x00, 0x28, 0xFF})
	hvcC := ibox("hvcC", []byte{1, 0x02, 0x60, 0, 0, 0, 0x90, 0, 0, 0, 0, 0, 93, 0xF0})
	av1C := ibox("av1C", []byte{0x81, 0x08, 0x4C, 0})
	vpcC := ibox("vpcC", []byte{1, 0, 0, 0, 2, 10, 0xA0, 0})

	cases := []struct {
		name string
		data []byte
		want HLSVariantInfo
	}{
		{"h264+aac", initSegment(
			trakBox("vide", videoEntry("avc1", 1920, 1080, avcC)),
			trakBox("soun", audioEntry(0, esdsAAC(0x12, 0x10)))),
			HLSVariantInfo{Codecs: "avc1.640028,mp4a.40.2", Width: 1920, Height: 1080}},
		{"hevc hdr10", initSegment(
			trakBox("vide", videoEntry("hvc1", 3840, 2160, hvcC, colrBox(16))),
			trakBox("soun", audioEntry(0, esdsAAC(0x12, 0x10)))),
			HLSVariantInfo{Codecs: "hvc1.2.6.L93.90,mp4a.40.2", VideoRange: "PQ", Width: 3840, Height: 2160}},
		{"av1 hlg", initSegment(
			trakBox("vide", videoEntry("av01", 1280, 720, av1C, colrBox(18)))),
			HLSVariantInfo{Codecs: "av01.0.08M.10", VideoRange: "HLG", Width: 1280, Height: 720}},
		{"vp9 sdr", initSegment(
			trakBox("vide", videoEntry("vp09", 640, 360, vpcC, colrBox(1)))),
			HLSVariantInfo{Codecs: "vp09.02.10.10", VideoRange: "SDR", Width: 640, Height: 360}},
		{"audio only", initSegment(trakBox("soun", audioEntry(0, esdsAAC(0x12, 0x10)))),
			HLSVariantInfo{Codecs: "mp4a.40.2"}},
		{"aac in a version 1 sample entry", initSegment(
			trakBox("vide", videoEntry("avc3", 320, 240, avcC)),
			trakBox("soun", audioEntry(1, esdsAAC(0x2B, 0x92)))), // HE-AAC: 0x2B>>3 = 5
			HLSVariantInfo{Codecs: "avc3.640028,mp4a.40.5", Width: 320, Height: 240}},
		{"aac in a version 2 sample entry", initSegment(
			trakBox("soun", audioEntry(2, esdsAAC(0xF8, 0x20)))), // escape: 32 + 0<<3 + 1
			HLSVariantInfo{Codecs: "mp4a.40.33"}},
		{"second audio track does not add a second mp4a", initSegment(
			trakBox("vide", videoEntry("avc1", 1, 1, avcC)),
			trakBox("soun", audioEntry(0, esdsAAC(0x12, 0x10))),
			trakBox("soun", audioEntry(0, esdsAAC(0x2B, 0x92)))),
			HLSVariantInfo{Codecs: "avc1.640028,mp4a.40.2", Width: 1, Height: 1}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := readMP4InitInfo(c.data)
			if got == nil || *got != c.want {
				t.Fatalf("got %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestReadMP4InitInfoRejects(t *testing.T) {
	avcC := ibox("avcC", []byte{1, 0x64, 0x00, 0x28, 0xFF})
	good := initSegment(trakBox("vide", videoEntry("avc1", 1, 1, avcC)))

	cases := map[string][]byte{
		"nil":           nil,
		"short":         {0, 0, 0, 8},
		"no moov":       ibox("ftyp", []byte("isom")),
		"empty moov":    initSegment(),
		"subtitle trak": initSegment(trakBox("text", videoEntry("avc1", 1, 1, avcC))),
		"unknown codec": initSegment(trakBox("vide", videoEntry("zzzz", 1, 1, avcC))),
		"missing avcC":  initSegment(trakBox("vide", videoEntry("avc1", 1, 1))),
		"short avcC":    initSegment(trakBox("vide", videoEntry("avc1", 1, 1, ibox("avcC", []byte{1, 2})))),
		"no esds":       initSegment(trakBox("soun", audioEntry(0))),
		"esds w/o 0x05": initSegment(trakBox("soun", audioEntry(0, ibox("esds", make([]byte, 20))))),
		"non-mp4a audio": initSegment(trakBox("soun",
			ibox("ac-3", make([]byte, 40)))),
	}
	for name, data := range cases {
		if got := readMP4InitInfo(data); got != nil {
			t.Errorf("%s: got %+v, want nil", name, got)
		}
	}

	// Every truncation of a valid init segment must be handled without a panic
	// (these bytes come from the muxer, but also from a failing source).
	for n := 0; n < len(good); n++ {
		_ = readMP4InitInfo(good[:n])
	}
}

func TestReadInitBoxSizes(t *testing.T) {
	// 64-bit largesize.
	large := binary.BigEndian.AppendUint32(nil, 1)
	large = append(large, "free"...)
	large = binary.BigEndian.AppendUint64(large, 20)
	large = append(large, 1, 2, 3, 4)
	b, next, ok := readInitBox(large, 0, len(large))
	if !ok || b.typ != "free" || b.payloadStart != 16 || next != 20 {
		t.Errorf("largesize: %+v next=%d ok=%v", b, next, ok)
	}

	// size 0: box extends to the end of the parent.
	toEnd := append(binary.BigEndian.AppendUint32(nil, 0), "mdat"...)
	toEnd = append(toEnd, 9, 9, 9)
	if b, next, ok := readInitBox(toEnd, 0, len(toEnd)); !ok || b.end != len(toEnd) || next != len(toEnd) {
		t.Errorf("size 0: %+v next=%d ok=%v", b, next, ok)
	}

	// Sizes smaller than a header, or larger than the parent, are corrupt.
	for name, size := range map[string]uint32{"tiny": 4, "past parent": 100} {
		raw := binary.BigEndian.AppendUint32(nil, size)
		raw = append(raw, "free"...)
		if _, _, ok := readInitBox(raw, 0, len(raw)); ok {
			t.Errorf("%s accepted", name)
		}
	}
	// largesize header cut short, huge largesize, negative cursor, end past data.
	cut := append(binary.BigEndian.AppendUint32(nil, 1), "free"...)
	if _, _, ok := readInitBox(cut, 0, len(cut)); ok {
		t.Error("truncated largesize accepted")
	}
	huge := append(append(binary.BigEndian.AppendUint32(nil, 1), "free"...), 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF)
	if _, _, ok := readInitBox(huge, 0, len(huge)); ok {
		t.Error("2^64-1 largesize accepted")
	}
	if _, _, ok := readInitBox(toEnd, -1, len(toEnd)); ok {
		t.Error("negative cursor")
	}
	if _, _, ok := readInitBox(toEnd, 0, len(toEnd)+1); ok {
		t.Error("end past the data")
	}
}

func TestReadAV1AndVP9EdgeCases(t *testing.T) {
	if got := readAV1Codec([]byte{0x01, 0x08, 0x00}); got != "" {
		t.Errorf("av1C without the marker bit = %q", got)
	}
	if got := readAV1Codec([]byte{0x81, 0x08}); got != "" {
		t.Errorf("short av1C = %q", got)
	}
	// profile 2, 12-bit, high tier, level 13
	if got := readAV1Codec([]byte{0x81, 2<<5 | 13, 0xE0}); got != "av01.2.13H.12" {
		t.Errorf("12-bit = %q", got)
	}
	// 8-bit main tier
	if got := readAV1Codec([]byte{0x81, 0x00, 0x00}); got != "av01.0.00M.08" {
		t.Errorf("8-bit = %q", got)
	}
	// vpcC without the 4-byte version/flags header: fields start at 0.
	if got := readVP9Codec([]byte{0, 31, 8}); got != "vp09.00.31.08" {
		t.Errorf("headerless vpcC = %q", got)
	}
	if got := readVP9Codec([]byte{0, 31}); got != "" {
		t.Errorf("short vpcC = %q", got)
	}
}

func TestReadDescriptorLength(t *testing.T) {
	cases := []struct {
		in   []byte
		want int
		ok   bool
	}{
		{[]byte{0x05}, 5, true},
		{[]byte{0x81, 0x00}, 128, true},
		{[]byte{0x80, 0x80, 0x80, 0x22}, 34, true},
		{[]byte{0x80, 0x80, 0x80, 0x80, 0x01}, 0, false}, // more than 4 bytes
		{[]byte{0x80}, 0, false},                         // runs off the end
		{nil, 0, false},
	}
	for _, c := range cases {
		cur := 0
		got, ok := readDescriptorLength(c.in, &cur)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("readDescriptorLength(%v) = %d,%v want %d,%v", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestReadInitVideoRange(t *testing.T) {
	mk := func(tag string, transfer uint16) []byte {
		p := []byte(tag)
		p = binary.BigEndian.AppendUint16(p, 9)
		p = binary.BigEndian.AppendUint16(p, transfer)
		return append(p, 0, 0, 0)
	}
	for transfer, want := range map[uint16]string{16: "PQ", 18: "HLG", 1: "SDR", 13: "SDR", 99: ""} {
		if got := readInitVideoRange(mk("nclx", transfer)); got != want {
			t.Errorf("transfer %d = %q, want %q", transfer, got, want)
		}
	}
	if got := readInitVideoRange(mk("nclc", 16)); got != "PQ" {
		t.Errorf("nclc = %q", got)
	}
	if got := readInitVideoRange(mk("prof", 16)); got != "" {
		t.Errorf("icc profile colr = %q, want empty", got)
	}
	if got := readInitVideoRange([]byte("nclx")); got != "" {
		t.Errorf("short colr = %q", got)
	}
}
