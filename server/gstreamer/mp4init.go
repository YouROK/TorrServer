//go:build gst

package gstreamer

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"strconv"
	"strings"
)

type HLSVariantInfo struct {
	Codecs     string
	VideoRange string
	Width      int
	Height     int
	FrameRate  float64
}

type initBox struct {
	typ          string
	payloadStart int
	end          int
}

type initTrack struct {
	video      bool
	codec      string
	videoRange string
	width      int
	height     int
}

func readMP4InitInfo(data []byte) *HLSVariantInfo {
	if len(data) < 8 {
		return nil
	}
	moov, ok := findInitChild(data, 0, len(data), "moov")
	if !ok {
		return nil
	}
	result := &HLSVariantInfo{}
	for cursor := moov.payloadStart; ; {
		box, next, ok := readInitBox(data, cursor, moov.end)
		if !ok {
			break
		}
		cursor = next
		if box.typ != "trak" {
			continue
		}
		track := readInitTrack(data, box)
		if track == nil || track.codec == "" {
			continue
		}
		if track.video && result.Codecs == "" {
			result.Codecs = track.codec
			result.VideoRange = track.videoRange
			result.Width = track.width
			result.Height = track.height
		} else if !track.video {
			if result.Codecs == "" {
				result.Codecs = track.codec
			} else if !strings.Contains(result.Codecs, "mp4a.") {
				result.Codecs += "," + track.codec
			}
		}
	}
	if result.Codecs == "" {
		return nil
	}
	return result
}

func readInitTrack(data []byte, trak initBox) *initTrack {
	mdia, ok := findInitChild(data, trak.payloadStart, trak.end, "mdia")
	if !ok {
		return nil
	}
	hdlr, ok := findInitChild(data, mdia.payloadStart, mdia.end, "hdlr")
	if !ok || hdlr.payloadStart > hdlr.end-12 {
		return nil
	}
	handler := string(data[hdlr.payloadStart+8 : hdlr.payloadStart+12])
	isVideo := handler == "vide"
	if !isVideo && handler != "soun" {
		return nil
	}
	minf, ok := findInitChild(data, mdia.payloadStart, mdia.end, "minf")
	if !ok {
		return nil
	}
	stbl, ok := findInitChild(data, minf.payloadStart, minf.end, "stbl")
	if !ok {
		return nil
	}
	stsd, ok := findInitChild(data, stbl.payloadStart, stbl.end, "stsd")
	if !ok || stsd.payloadStart > stsd.end-8 {
		return nil
	}
	entryCount := int(binary.BigEndian.Uint32(data[stsd.payloadStart+4 : stsd.payloadStart+8]))
	cursor := stsd.payloadStart + 8
	for i := 0; i < entryCount; i++ {
		entry, next, ok := readInitBox(data, cursor, stsd.end)
		if !ok {
			break
		}
		cursor = next
		var result *initTrack
		if isVideo {
			result = readInitVideoEntry(data, entry)
		} else {
			result = readInitAudioEntry(data, entry)
		}
		if result != nil {
			return result
		}
	}
	return nil
}

func readInitVideoEntry(data []byte, entry initBox) *initTrack {
	fixedEnd := entry.payloadStart + 78
	if fixedEnd > entry.end {
		return nil
	}
	configType := map[string]string{"avc1": "avcC", "avc3": "avcC", "hvc1": "hvcC", "hev1": "hvcC", "av01": "av1C", "vp09": "vpcC"}[entry.typ]
	if configType == "" {
		return nil
	}
	config, ok := findInitChild(data, fixedEnd, entry.end, configType)
	if !ok {
		return nil
	}
	payload := data[config.payloadStart:config.end]
	codec := ""
	switch entry.typ {
	case "avc1", "avc3":
		if len(payload) >= 4 {
			codec = fmt.Sprintf("%s.%02X%02X%02X", entry.typ, payload[1], payload[2], payload[3])
		}
	case "hvc1", "hev1":
		codec = readHEVCCodec(entry.typ, payload)
	case "av01":
		codec = readAV1Codec(payload)
	case "vp09":
		codec = readVP9Codec(payload)
	}
	if codec == "" {
		return nil
	}
	videoRange := ""
	if colr, ok := findInitChild(data, fixedEnd, entry.end, "colr"); ok {
		videoRange = readInitVideoRange(data[colr.payloadStart:colr.end])
	}
	return &initTrack{
		video:      true,
		codec:      codec,
		videoRange: videoRange,
		width:      int(binary.BigEndian.Uint16(data[entry.payloadStart+24 : entry.payloadStart+26])),
		height:     int(binary.BigEndian.Uint16(data[entry.payloadStart+26 : entry.payloadStart+28])),
	}
}

func readInitAudioEntry(data []byte, entry initBox) *initTrack {
	if entry.typ != "mp4a" || entry.payloadStart > entry.end-28 {
		return nil
	}
	version := binary.BigEndian.Uint16(data[entry.payloadStart+8 : entry.payloadStart+10])
	childStart := entry.payloadStart + 28
	if version == 1 {
		childStart = entry.payloadStart + 44
	} else if version == 2 {
		childStart = entry.payloadStart + 64
	}
	esds, ok := findInitChild(data, childStart, entry.end, "esds")
	if !ok {
		return nil
	}
	codec := readAACCodec(data[esds.payloadStart:esds.end])
	if codec == "" {
		return nil
	}
	return &initTrack{codec: codec}
}

func readHEVCCodec(sampleEntry string, data []byte) string {
	if len(data) < 13 {
		return ""
	}
	profileByte := data[1]
	profileSpace := []string{"", "A", "B", "C"}[profileByte>>6]
	profileIDC := profileByte & 0x1f
	compatibility := bits.Reverse32(binary.BigEndian.Uint32(data[2:6]))
	tier := "L"
	if profileByte&0x20 != 0 {
		tier = "H"
	}
	result := fmt.Sprintf("%s.%s%d.%d.%s%d", sampleEntry, profileSpace, profileIDC, compatibility, tier, data[12])
	lastConstraint := 11
	for lastConstraint >= 6 && data[lastConstraint] == 0 {
		lastConstraint--
	}
	for i := 6; i <= lastConstraint; i++ {
		result += fmt.Sprintf(".%02X", data[i])
	}
	return result
}

func readAV1Codec(data []byte) string {
	if len(data) < 3 || data[0]&0x80 == 0 {
		return ""
	}
	profile := data[1] >> 5
	level := data[1] & 0x1f
	tier := "M"
	if data[2]&0x80 != 0 {
		tier = "H"
	}
	bitDepth := 8
	if data[2]&0x40 != 0 {
		bitDepth = 10
		if profile == 2 && data[2]&0x20 != 0 {
			bitDepth = 12
		}
	}
	return fmt.Sprintf("av01.%d.%02d%s.%02d", profile, level, tier, bitDepth)
}

func readVP9Codec(data []byte) string {
	offset := 0
	if len(data) >= 7 && data[0] <= 1 && data[1] == 0 && data[2] == 0 && data[3] == 0 {
		offset = 4
	}
	if len(data) < offset+3 {
		return ""
	}
	bitDepth := data[offset+2]
	if bitDepth > 16 {
		bitDepth >>= 4
	}
	return fmt.Sprintf("vp09.%02d.%02d.%02d", data[offset], data[offset+1], bitDepth)
}

func readAACCodec(data []byte) string {
	for i := min(4, len(data)); i < len(data)-2; i++ {
		if data[i] != 0x05 {
			continue
		}
		cursor := i + 1
		length, ok := readDescriptorLength(data, &cursor)
		if !ok || length < 2 || cursor > len(data)-length {
			continue
		}
		objectType := int(data[cursor] >> 3)
		if objectType == 31 {
			objectType = 32 + (int(data[cursor]&7) << 3) + int(data[cursor+1]>>5)
		}
		return "mp4a.40." + strconv.Itoa(objectType)
	}
	return ""
}

func readDescriptorLength(data []byte, cursor *int) (int, bool) {
	length := 0
	for i := 0; i < 4 && *cursor < len(data); i++ {
		value := data[*cursor]
		*cursor = *cursor + 1
		length = length<<7 | int(value&0x7f)
		if value&0x80 == 0 {
			return length, true
		}
	}
	return 0, false
}

func readInitVideoRange(data []byte) string {
	if len(data) < 10 || (string(data[:4]) != "nclx" && string(data[:4]) != "nclc") {
		return ""
	}
	switch binary.BigEndian.Uint16(data[6:8]) {
	case 16:
		return "PQ"
	case 18:
		return "HLG"
	case 1, 4, 5, 6, 7, 8, 13, 14, 15:
		return "SDR"
	default:
		return ""
	}
}

func findInitChild(data []byte, start int, end int, typ string) (initBox, bool) {
	for cursor := start; ; {
		box, next, ok := readInitBox(data, cursor, end)
		if !ok {
			return initBox{}, false
		}
		if box.typ == typ {
			return box, true
		}
		cursor = next
	}
}

func readInitBox(data []byte, cursor int, end int) (initBox, int, bool) {
	if cursor < 0 || end > len(data) || cursor > end-8 {
		return initBox{}, end, false
	}
	start := cursor
	size32 := binary.BigEndian.Uint32(data[start : start+4])
	typ := string(data[start+4 : start+8])
	headerSize := 8
	size := uint64(size32)
	if size32 == 1 {
		if start > end-16 {
			return initBox{}, end, false
		}
		size = binary.BigEndian.Uint64(data[start+8 : start+16])
		headerSize = 16
	} else if size32 == 0 {
		size = uint64(end - start)
	}
	if size < uint64(headerSize) || size > uint64(end-start) || size > uint64(^uint(0)>>1) {
		return initBox{}, end, false
	}
	boxEnd := start + int(size)
	return initBox{typ: typ, payloadStart: start + headerSize, end: boxEnd}, boxEnd, true
}
