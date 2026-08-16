//go:build gst

package gstreamer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	ebmlID               = uint64(0x1A45DFA3)
	matroskaSegmentID    = uint64(0x18538067)
	seekHeadID           = uint64(0x114D9B74)
	seekID               = uint64(0x4DBB)
	seekEntryID          = uint64(0x53AB)
	seekPositionID       = uint64(0x53AC)
	infoID               = uint64(0x1549A966)
	timestampScaleID     = uint64(0x2AD7B1)
	tracksID             = uint64(0x1654AE6B)
	trackEntryID         = uint64(0xAE)
	trackNumberID        = uint64(0xD7)
	trackTypeID          = uint64(0x83)
	cuesID               = uint64(0x1C53BB6B)
	cuePointID           = uint64(0xBB)
	cueTimeID            = uint64(0xB3)
	cueTrackPositionsID  = uint64(0xB7)
	cueTrackID           = uint64(0xF7)
	cueClusterPositionID = uint64(0xF1)

	defaultTimestampScaleNS = uint64(1_000_000)
	matroskaPrefixLength    = 4 * 1024 * 1024
	maxMatroskaMetadata     = 8 * 1024 * 1024
	maxMatroskaCues         = 64 * 1024 * 1024
	maxMatroskaCuePoints    = 1_000_000
)

type CueSegment struct {
	StartNS uint64 `json:"start_ns"`
	EndNS   uint64 `json:"end_ns"`
}

func (s CueSegment) DurationNS() uint64 { return s.EndNS - s.StartNS }

type CueTimeline struct {
	Segments         []CueSegment `json:"segments"`
	TimestampScaleNS uint64       `json:"timestamp_scale_ns"`
	MaxDurationNS    uint64       `json:"max_duration_ns"`
}

func (t *CueTimeline) Segment(index int) (CueSegment, bool) {
	if t == nil || index < 0 || index >= len(t.Segments) {
		return CueSegment{}, false
	}
	return t.Segments[index], true
}

type ebmlElement struct {
	id          uint64
	offset      int
	dataOffset  int
	size        uint64
	unknownSize bool
}

func (e ebmlElement) endOffset() uint64 { return uint64(e.dataOffset) + e.size }

func readMatroskaCueTimeline(sourceURL string, contentLength int64, durationNS int64) *CueTimeline {
	if sourceURL == "" || durationNS <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	prefixLength := matroskaPrefixLength
	if contentLength > 0 && contentLength < int64(prefixLength) {
		prefixLength = int(contentLength)
	}
	prefix, err := readHTTPRange(ctx, sourceURL, 0, prefixLength)
	if err != nil {
		return nil
	}

	segment, ok := findMatroskaSegment(prefix)
	if !ok {
		return nil
	}
	segmentDataOffset := int64(segment.dataOffset)
	positions := make(map[uint64]uint64)
	timestampScaleNS := defaultTimestampScaleNS
	videoTrackNumber := uint64(0)
	parseMatroskaPrefix(prefix, segment.dataOffset, positions, &timestampScaleNS, &videoTrackNumber)

	if videoTrackNumber == 0 {
		if position, exists := positions[tracksID]; exists {
			if data := readMatroskaElement(ctx, sourceURL, segmentDataOffset, position, tracksID, maxMatroskaMetadata, contentLength); len(data) > 0 {
				if element, ok := readEBMLElement(data, 0); ok {
					videoTrackNumber = parseVideoTrackNumber(data, element)
				}
			}
		}
	}
	if position, exists := positions[infoID]; exists {
		if data := readMatroskaElement(ctx, sourceURL, segmentDataOffset, position, infoID, maxMatroskaMetadata, contentLength); len(data) > 0 {
			if element, ok := readEBMLElement(data, 0); ok {
				parseTimestampScale(data, element, &timestampScaleNS)
			}
		}
	}

	cuesPosition, exists := positions[cuesID]
	if !exists || videoTrackNumber == 0 || timestampScaleNS == 0 {
		return nil
	}
	cues := readMatroskaElement(ctx, sourceURL, segmentDataOffset, cuesPosition, cuesID, maxMatroskaCues, contentLength)
	if len(cues) == 0 {
		return nil
	}
	element, ok := readEBMLElement(cues, 0)
	if !ok {
		return nil
	}
	return parseCueTimeline(cues, element, videoTrackNumber, timestampScaleNS, uint64(durationNS))
}

func parseMatroskaPrefix(data []byte, offset int, positions map[uint64]uint64, timestampScaleNS *uint64, videoTrackNumber *uint64) {
	for offset < len(data) {
		element, ok := readEBMLElement(data, offset)
		if !ok || element.unknownSize || element.endOffset() > uint64(len(data)) {
			return
		}
		switch element.id {
		case seekHeadID:
			parseSeekHead(data, element, positions)
		case infoID:
			parseTimestampScale(data, element, timestampScaleNS)
		case tracksID:
			*videoTrackNumber = parseVideoTrackNumber(data, element)
		}
		offset = int(element.endOffset())
	}
}

func parseSeekHead(data []byte, parent ebmlElement, positions map[uint64]uint64) {
	for offset, end := parent.dataOffset, int(parent.endOffset()); offset < end; {
		entry, ok := readEBMLElement(data, offset)
		if !ok || entry.unknownSize || entry.endOffset() > uint64(end) {
			return
		}
		if entry.id == seekID {
			var target, position uint64
			hasPosition := false
			for childOffset, childEnd := entry.dataOffset, int(entry.endOffset()); childOffset < childEnd; {
				child, ok := readEBMLElement(data, childOffset)
				if !ok || child.unknownSize || child.endOffset() > uint64(childEnd) {
					break
				}
				switch {
				case child.id == seekEntryID && child.size > 0 && child.size <= 4:
					target, _ = readEBMLUnsigned(data, child)
				case child.id == seekPositionID && child.size > 0 && child.size <= 8:
					position, _ = readEBMLUnsigned(data, child)
					hasPosition = true
				}
				childOffset = int(child.endOffset())
			}
			if target != 0 && hasPosition {
				if _, exists := positions[target]; !exists {
					positions[target] = position
				}
			}
		}
		offset = int(entry.endOffset())
	}
}

func parseTimestampScale(data []byte, parent ebmlElement, result *uint64) {
	for offset, end := parent.dataOffset, int(parent.endOffset()); offset < end; {
		child, ok := readEBMLElement(data, offset)
		if !ok || child.unknownSize || child.endOffset() > uint64(end) {
			return
		}
		if child.id == timestampScaleID && child.size > 0 && child.size <= 8 {
			if value, ok := readEBMLUnsigned(data, child); ok && value > 0 {
				*result = value
			}
			return
		}
		offset = int(child.endOffset())
	}
}

func parseVideoTrackNumber(data []byte, tracks ebmlElement) uint64 {
	for offset, end := tracks.dataOffset, int(tracks.endOffset()); offset < end; {
		entry, ok := readEBMLElement(data, offset)
		if !ok || entry.unknownSize || entry.endOffset() > uint64(end) {
			return 0
		}
		if entry.id == trackEntryID {
			var number, trackType uint64
			for childOffset, childEnd := entry.dataOffset, int(entry.endOffset()); childOffset < childEnd; {
				child, ok := readEBMLElement(data, childOffset)
				if !ok || child.unknownSize || child.endOffset() > uint64(childEnd) {
					break
				}
				if child.id == trackNumberID {
					number, _ = readEBMLUnsigned(data, child)
				} else if child.id == trackTypeID {
					trackType, _ = readEBMLUnsigned(data, child)
				}
				childOffset = int(child.endOffset())
			}
			if trackType == 1 && number > 0 {
				return number
			}
		}
		offset = int(entry.endOffset())
	}
	return 0
}

func parseCueTimeline(data []byte, cues ebmlElement, videoTrack, timestampScaleNS, durationNS uint64) *CueTimeline {
	times := make([]uint64, 0, 1024)
	for offset, end := cues.dataOffset, int(cues.endOffset()); offset < end; {
		point, ok := readEBMLElement(data, offset)
		if !ok || point.unknownSize || point.endOffset() > uint64(end) {
			return nil
		}
		if point.id == cuePointID {
			if cueTime, ok := readVideoCueTime(data, point, videoTrack); ok && cueTime <= ^uint64(0)/timestampScaleNS {
				timeNS := cueTime * timestampScaleNS
				if timeNS < durationNS {
					times = append(times, timeNS)
					if len(times) > maxMatroskaCuePoints {
						return nil
					}
				}
			}
		}
		offset = int(point.endOffset())
	}
	if len(times) < 2 {
		return nil
	}
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	segments := make([]CueSegment, 0, len(times))
	previous := times[0]
	for _, current := range times[1:] {
		if current <= previous {
			continue
		}
		segments = append(segments, CueSegment{StartNS: previous, EndNS: current})
		previous = current
	}
	if durationNS > previous {
		segments = append(segments, CueSegment{StartNS: previous, EndNS: durationNS})
	}
	if len(segments) == 0 {
		return nil
	}
	timeline := &CueTimeline{Segments: segments, TimestampScaleNS: timestampScaleNS}
	for _, segment := range segments {
		if duration := segment.DurationNS(); duration > timeline.MaxDurationNS {
			timeline.MaxDurationNS = duration
		}
	}
	return timeline
}

func readVideoCueTime(data []byte, point ebmlElement, videoTrack uint64) (uint64, bool) {
	var cueTime uint64
	hasTime, hasPosition := false, false
	for offset, end := point.dataOffset, int(point.endOffset()); offset < end; {
		child, ok := readEBMLElement(data, offset)
		if !ok || child.unknownSize || child.endOffset() > uint64(end) {
			return 0, false
		}
		if child.id == cueTimeID {
			cueTime, hasTime = readEBMLUnsigned(data, child)
		} else if child.id == cueTrackPositionsID && hasVideoCuePosition(data, child, videoTrack) {
			hasPosition = true
		}
		offset = int(child.endOffset())
	}
	return cueTime, hasTime && hasPosition
}

func hasVideoCuePosition(data []byte, positions ebmlElement, videoTrack uint64) bool {
	var track uint64
	hasCluster := false
	for offset, end := positions.dataOffset, int(positions.endOffset()); offset < end; {
		child, ok := readEBMLElement(data, offset)
		if !ok || child.unknownSize || child.endOffset() > uint64(end) {
			return false
		}
		if child.id == cueTrackID {
			track, _ = readEBMLUnsigned(data, child)
		} else if child.id == cueClusterPositionID {
			_, hasCluster = readEBMLUnsigned(data, child)
		}
		offset = int(child.endOffset())
	}
	return track == videoTrack && hasCluster
}

func readMatroskaElement(ctx context.Context, sourceURL string, segmentOffset int64, relative uint64, expectedID uint64, maxLength int, contentLength int64) []byte {
	if relative > uint64(^uint64(0)>>1) || segmentOffset < 0 || int64(relative) > int64(^uint64(0)>>1)-segmentOffset {
		return nil
	}
	offset := segmentOffset + int64(relative)
	headerLength := 16
	if contentLength > 0 && contentLength-offset < int64(headerLength) {
		headerLength = int(contentLength - offset)
	}
	if headerLength <= 0 {
		return nil
	}
	header, err := readHTTPRange(ctx, sourceURL, offset, headerLength)
	if err != nil {
		return nil
	}
	element, ok := readEBMLElement(header, 0)
	if !ok || element.id != expectedID || element.unknownSize || element.size > uint64(maxLength) {
		return nil
	}
	total := uint64(element.dataOffset-element.offset) + element.size
	if total > uint64(maxLength)+16 || total > uint64(^uint(0)>>1) {
		return nil
	}
	if contentLength > 0 && (offset > contentLength || int64(total) > contentLength-offset) {
		return nil
	}
	data, err := readHTTPRange(ctx, sourceURL, offset, int(total))
	if err != nil {
		return nil
	}
	return data
}

func readHTTPRange(ctx context.Context, sourceURL string, offset int64, length int) ([]byte, error) {
	if offset < 0 || length <= 0 || int64(length)-1 > math.MaxInt64-offset {
		return nil, errors.New("invalid byte range")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, offset+int64(length)-1))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent && !(offset == 0 && resp.StatusCode == http.StatusOK) {
		return nil, fmt.Errorf("range request returned %s", resp.Status)
	}
	if resp.StatusCode == http.StatusPartialContent {
		contentRange := resp.Header.Get("Content-Range")
		if !strings.HasPrefix(contentRange, "bytes "+strconv.FormatInt(offset, 10)+"-") {
			return nil, errors.New("invalid content-range")
		}
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(resp.Body, data); err != nil {
		return nil, err
	}
	return data, nil
}

func findMatroskaSegment(data []byte) (ebmlElement, bool) {
	ebml, ok := readEBMLElement(data, 0)
	if !ok || ebml.id != ebmlID || ebml.unknownSize || ebml.endOffset() > uint64(len(data)) {
		return ebmlElement{}, false
	}
	segment, ok := readEBMLElement(data, int(ebml.endOffset()))
	return segment, ok && segment.id == matroskaSegmentID
}

func readEBMLElement(data []byte, offset int) (ebmlElement, bool) {
	if offset < 0 || offset >= len(data) {
		return ebmlElement{}, false
	}
	idLength, ok := ebmlVintLength(data[offset], 4)
	if !ok || len(data)-offset < idLength+1 {
		return ebmlElement{}, false
	}
	id := uint64(0)
	for i := 0; i < idLength; i++ {
		id = id<<8 | uint64(data[offset+i])
	}
	sizeOffset := offset + idLength
	sizeLength, ok := ebmlVintLength(data[sizeOffset], 8)
	if !ok || len(data)-sizeOffset < sizeLength {
		return ebmlElement{}, false
	}
	marker := byte(0x80 >> (sizeLength - 1))
	size := uint64(data[sizeOffset] & (marker - 1))
	for i := 1; i < sizeLength; i++ {
		size = size<<8 | uint64(data[sizeOffset+i])
	}
	valueBits := sizeLength * 7
	unknown := uint64(1)<<valueBits - 1
	if valueBits == 56 {
		unknown = 0x00ff_ffff_ffff_ffff
	}
	return ebmlElement{id: id, offset: offset, dataOffset: sizeOffset + sizeLength, size: size, unknownSize: size == unknown}, true
}

func ebmlVintLength(first byte, maxLength int) (int, bool) {
	length, mask := 1, byte(0x80)
	for length <= maxLength && first&mask == 0 {
		mask >>= 1
		length++
	}
	return length, length <= maxLength
}

func readEBMLUnsigned(data []byte, element ebmlElement) (uint64, bool) {
	if element.size == 0 || element.size > 8 || element.endOffset() > uint64(len(data)) {
		return 0, false
	}
	value := uint64(0)
	for i, end := element.dataOffset, int(element.endOffset()); i < end; i++ {
		value = value<<8 | uint64(data[i])
	}
	return value, true
}
