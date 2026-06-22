package gstreamer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
)

type boxTarget int

const (
	boxTargetNone boxTarget = iota
	boxTargetInit
	boxTargetMoof
	boxTargetAudio
	boxTargetVideo
)

const (
	boxStyp = uint32('s')<<24 | uint32('t')<<16 | uint32('y')<<8 | uint32('p')
	boxMoov = uint32('m')<<24 | uint32('o')<<16 | uint32('o')<<8 | uint32('v')
	boxMoof = uint32('m')<<24 | uint32('o')<<16 | uint32('o')<<8 | uint32('f')
	boxMdat = uint32('m')<<24 | uint32('d')<<16 | uint32('a')<<8 | uint32('t')
	boxTrak = uint32('t')<<24 | uint32('r')<<16 | uint32('a')<<8 | uint32('k')
	boxTkhd = uint32('t')<<24 | uint32('k')<<16 | uint32('h')<<8 | uint32('d')
	boxMdia = uint32('m')<<24 | uint32('d')<<16 | uint32('i')<<8 | uint32('a')
	boxMdhd = uint32('m')<<24 | uint32('d')<<16 | uint32('h')<<8 | uint32('d')
	boxTraf = uint32('t')<<24 | uint32('r')<<16 | uint32('a')<<8 | uint32('f')
	boxTfhd = uint32('t')<<24 | uint32('f')<<16 | uint32('h')<<8 | uint32('d')
	boxTfdt = uint32('t')<<24 | uint32('f')<<16 | uint32('d')<<8 | uint32('t')

	videoTrackID uint32 = 1
	audioTrackID uint32 = 2
)

type Mp4BoxReader struct {
	onInit    func([]byte)
	onSegment func(Segment)

	init     bytes.Buffer
	moof     bytes.Buffer
	deferred bytes.Buffer

	boxHeader         [16]byte
	boxHeaderLength   int
	boxHeaderRequired int

	currentBoxType      uint32
	currentBoxRemaining uint64
	currentTarget       boxTarget

	audioPart bytes.Buffer
	videoPart bytes.Buffer

	initDone      bool
	moovCompleted bool

	lastMoofTrackID uint32

	videoTimescale uint32
	audioTimescale uint32

	segmentStartSeconds float64
	tfdtOffsetSeconds   float64
}

func NewMp4BoxReader(onInit func([]byte), onSegment func(Segment)) *Mp4BoxReader {
	reader := &Mp4BoxReader{
		onInit:              onInit,
		onSegment:           onSegment,
		boxHeaderRequired:   8,
		segmentStartSeconds: -1,
	}
	reader.moof.Grow(16 * 1024)
	reader.deferred.Grow(64 * 1024)
	return reader
}

func (r *Mp4BoxReader) ResetSegment() {
	if r.currentBoxType == boxMdat && r.currentBoxRemaining > 0 {
		r.currentTarget = boxTargetNone
	}

	r.audioPart.Reset()
	r.videoPart.Reset()
	r.lastMoofTrackID = 0
	r.segmentStartSeconds = -1
}

func (r *Mp4BoxReader) SeekReset(seconds float64) {
	r.initDone = false
	r.moovCompleted = false

	r.videoTimescale = 0
	r.audioTimescale = 0
	r.lastMoofTrackID = 0
	r.segmentStartSeconds = -1

	if !math.IsNaN(seconds) && !math.IsInf(seconds, 0) && seconds > 0 {
		r.tfdtOffsetSeconds = seconds
	} else {
		r.tfdtOffsetSeconds = 0
	}

	r.init.Reset()
	r.moof.Reset()
	r.deferred.Reset()
	r.resetBoxState()
}

func (r *Mp4BoxReader) Push(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	completed, err := r.TryProcessDeferred()
	if err != nil {
		return err
	}
	if completed {
		r.deferred.Write(data)
		return nil
	}

	consumed, segmentCompleted, err := r.processBytes(data)
	if err != nil {
		return err
	}
	if !segmentCompleted {
		return nil
	}

	if consumed < len(data) {
		r.deferred.Write(data[consumed:])
	}
	return nil
}

func (r *Mp4BoxReader) processBytes(data []byte) (int, bool, error) {
	position := 0

	for position < len(data) {
		if r.boxHeaderLength < r.boxHeaderRequired {
			copyLen := minInt(r.boxHeaderRequired-r.boxHeaderLength, len(data)-position)
			copy(r.boxHeader[r.boxHeaderLength:r.boxHeaderLength+copyLen], data[position:position+copyLen])

			r.boxHeaderLength += copyLen
			position += copyLen

			if r.boxHeaderLength < r.boxHeaderRequired {
				break
			}

			if r.boxHeaderRequired == 8 {
				size32 := binary.BigEndian.Uint32(r.boxHeader[0:4])
				r.currentBoxType = binary.BigEndian.Uint32(r.boxHeader[4:8])

				if size32 == 1 {
					r.boxHeaderRequired = 16
					continue
				}

				if size32 == 0 {
					return position, false, errors.New("mp4 box with size=0 cannot be parsed before end of stream")
				}

				if err := r.beginBox(uint64(size32), 8); err != nil {
					return position, false, err
				}
			} else {
				size64 := binary.BigEndian.Uint64(r.boxHeader[8:16])
				if err := r.beginBox(size64, 16); err != nil {
					return position, false, err
				}
			}

			if r.currentBoxRemaining == 0 {
				completed, err := r.completeBox()
				if err != nil {
					return position, false, err
				}
				r.resetBoxState()

				if completed {
					return position, true, nil
				}
			}

			continue
		}

		bodySize := minInt(len(data)-position, int(minUint64(uint64(len(data)-position), r.currentBoxRemaining)))
		if bodySize <= 0 {
			break
		}

		r.writeCurrentBoxData(data[position : position+bodySize])

		position += bodySize
		r.currentBoxRemaining -= uint64(bodySize)

		if r.currentBoxRemaining == 0 {
			completed, err := r.completeBox()
			if err != nil {
				return position, false, err
			}
			r.resetBoxState()

			if completed {
				return position, true, nil
			}
		}
	}

	return position, false, nil
}

func (r *Mp4BoxReader) beginBox(size uint64, headerSize int) error {
	if size < uint64(headerSize) {
		return errors.New("invalid mp4 box size")
	}
	if r.currentBoxType == boxMoof && size > math.MaxInt32 {
		return errors.New("moof is too large")
	}

	r.currentBoxRemaining = size - uint64(headerSize)
	r.currentTarget = boxTargetNone

	if !r.initDone && (r.currentBoxType == boxStyp || r.currentBoxType == boxMoof) {
		if err := r.completeInit(); err != nil {
			return err
		}
	}

	if !r.initDone {
		if r.currentBoxType == boxMdat {
			return errors.New("bad mp4 init")
		}
		r.currentTarget = boxTargetInit
	} else if r.currentBoxType == boxMoof {
		r.moof.Reset()
		r.currentTarget = boxTargetMoof
	} else if r.currentBoxType == boxMdat {
		switch r.lastMoofTrackID {
		case audioTrackID:
			r.currentTarget = boxTargetAudio
		case videoTrackID:
			r.currentTarget = boxTargetVideo
		default:
			return errors.New("mp4 mdat does not follow a supported moof")
		}
	}

	r.writeCurrentBoxData(r.boxHeader[:headerSize])
	return nil
}

func (r *Mp4BoxReader) writeCurrentBoxData(data []byte) {
	if len(data) == 0 {
		return
	}

	switch r.currentTarget {
	case boxTargetInit:
		r.init.Write(data)
	case boxTargetMoof:
		r.moof.Write(data)
	case boxTargetAudio:
		r.audioPart.Write(data)
	case boxTargetVideo:
		r.videoPart.Write(data)
	}
}

func (r *Mp4BoxReader) completeBox() (bool, error) {
	switch r.currentBoxType {
	case boxMoov:
		if r.initDone {
			return false, errors.New("unexpected moov after mp4 initialization")
		}
		r.moovCompleted = true
		return false, nil

	case boxMoof:
		return false, r.completeMoof()

	case boxMdat:
		r.lastMoofTrackID = 0

		if r.audioPart.Len() <= 0 || r.videoPart.Len() <= 0 {
			return false, nil
		}

		r.onSegment(Segment{
			Audio:        takeBuffer(&r.audioPart),
			Video:        takeBuffer(&r.videoPart),
			StartSeconds: r.segmentStartSeconds,
		})

		return true, nil
	}

	return false, nil
}

func (r *Mp4BoxReader) completeInit() error {
	if !r.moovCompleted {
		return errors.New("mp4 initialization is incomplete: moov box was not found")
	}
	if r.init.Len() <= 0 {
		return errors.New("mp4 initialization is empty")
	}

	init := cloneBytes(r.init.Bytes())

	videoTimescale := getTrackTimescale(init, videoTrackID)
	audioTimescale := getTrackTimescale(init, audioTrackID)

	if videoTimescale == 0 {
		return errors.New("video track or its mdhd timescale was not found in moov")
	}
	if audioTimescale == 0 {
		return errors.New("audio track or its mdhd timescale was not found in moov")
	}

	r.videoTimescale = videoTimescale
	r.audioTimescale = audioTimescale
	r.initDone = true

	r.onInit(init)
	return nil
}

func (r *Mp4BoxReader) completeMoof() error {
	box := cloneBytes(r.moof.Bytes())

	trackID, decodeTime := getMoofTrackID(box)
	if trackID == 0 {
		return errors.New("mp4 moof does not contain a readable tfhd track id")
	}
	if trackID != videoTrackID && trackID != audioTrackID {
		return errors.New("unsupported mp4 track id")
	}

	r.lastMoofTrackID = trackID

	var timescale uint32
	if trackID == audioTrackID {
		timescale = r.audioTimescale
	} else {
		timescale = r.videoTimescale
		if r.videoTimescale > 0 && decodeTime != nil {
			r.segmentStartSeconds = float64(*decodeTime) / float64(r.videoTimescale)
		}
	}

	if r.tfdtOffsetSeconds > 0 && timescale > 0 {
		shiftTfdt(box, timescale, r.tfdtOffsetSeconds)
	}

	if trackID == audioTrackID {
		r.audioPart.Write(box)
	} else {
		r.videoPart.Write(box)
	}

	r.moof.Reset()
	return nil
}

func (r *Mp4BoxReader) resetBoxState() {
	r.boxHeaderLength = 0
	r.boxHeaderRequired = 8
	r.currentBoxType = 0
	r.currentBoxRemaining = 0
	r.currentTarget = boxTargetNone
}

func (r *Mp4BoxReader) TryProcessDeferred() (bool, error) {
	if r.deferred.Len() <= 0 {
		return false, nil
	}

	data := cloneBytes(r.deferred.Bytes())
	consumed, completed, err := r.processBytes(data)
	if err != nil {
		return false, err
	}

	if completed {
		r.deferred.Reset()
		if consumed < len(data) {
			r.deferred.Write(data[consumed:])
		}
		return true, nil
	}

	if consumed == len(data) {
		r.deferred.Reset()
		return false, nil
	}

	return false, errors.New("mp4 parser consumed partial deferred data without completing a segment")
}

func shiftTfdt(moof []byte, timescale uint32, offsetSeconds float64) bool {
	if len(moof) < 8 || timescale == 0 || offsetSeconds <= 0 {
		return false
	}

	offsetValue := offsetSeconds * float64(timescale)
	if math.IsNaN(offsetValue) || math.IsInf(offsetValue, 0) || offsetValue <= 0 || offsetValue > float64(math.MaxUint64) {
		return false
	}

	offset := uint64(math.Round(offsetValue))

	rootPosition := 0
	rootType, moofHeaderSize, moofBox, ok := tryReadBox(moof, &rootPosition)
	if !ok || rootType != boxMoof || rootPosition != len(moof) {
		return false
	}

	children := moofBox[moofHeaderSize:]
	position := 0

	for position < len(children) {
		childStart := position
		childType, childHeaderSize, child, ok := tryReadBox(children, &position)
		if !ok {
			return false
		}
		if childType != boxTraf {
			continue
		}

		trafStart := moofHeaderSize + childStart
		trafEnd := trafStart + len(child)
		return shiftTrafTfdt(moof, trafStart+childHeaderSize, trafEnd, offset)
	}

	return false
}

func shiftTrafTfdt(data []byte, start int, end int, offset uint64) bool {
	if start < 0 || end < start || end > len(data) {
		return false
	}

	traf := data[start:end]
	position := 0

	for position < len(traf) {
		childStart := position
		boxType, headerSize, box, ok := tryReadBox(traf, &position)
		if !ok {
			return false
		}
		if boxType != boxTfdt {
			continue
		}

		if len(box) < headerSize+8 {
			return false
		}

		fullBoxOffset := start + childStart + headerSize
		version := data[fullBoxOffset]
		valueOffset := fullBoxOffset + 4

		switch version {
		case 1:
			if len(box) < headerSize+12 {
				return false
			}
			value := binary.BigEndian.Uint64(data[valueOffset : valueOffset+8])
			if math.MaxUint64-value < offset {
				return false
			}
			binary.BigEndian.PutUint64(data[valueOffset:valueOffset+8], value+offset)
			return true

		case 0:
			value := binary.BigEndian.Uint32(data[valueOffset : valueOffset+4])
			result := uint64(value) + offset
			if result > math.MaxUint32 {
				return false
			}
			binary.BigEndian.PutUint32(data[valueOffset:valueOffset+4], uint32(result))
			return true
		}

		return false
	}

	return false
}

func getMoofTrackID(moof []byte) (uint32, *uint64) {
	rootPosition := 0
	rootType, moofHeaderSize, moofBox, ok := tryReadBox(moof, &rootPosition)
	if !ok || rootType != boxMoof || rootPosition != len(moof) {
		return 0, nil
	}

	children := moofBox[moofHeaderSize:]
	position := 0

	for position < len(children) {
		childStart := position
		childType, childHeaderSize, child, ok := tryReadBox(children, &position)
		if !ok {
			return 0, nil
		}
		if childType != boxTraf {
			continue
		}

		trafStart := moofHeaderSize + childStart
		trafEnd := trafStart + len(child)
		trackID, decodeTime := getTrafTrackID(moof, trafStart+childHeaderSize, trafEnd)
		if trackID != 0 {
			return trackID, decodeTime
		}
	}

	return 0, nil
}

func getTrafTrackID(data []byte, start int, end int) (uint32, *uint64) {
	if start < 0 || end < start || end > len(data) {
		return 0, nil
	}

	var trackID uint32
	var decodeTime *uint64

	traf := data[start:end]
	position := 0

	for position < len(traf) {
		boxType, headerSize, box, ok := tryReadBox(traf, &position)
		if !ok {
			return 0, nil
		}

		switch boxType {
		case boxTfhd:
			if len(box) < headerSize+8 {
				return 0, nil
			}
			trackID = binary.BigEndian.Uint32(box[headerSize+4 : headerSize+8])

		case boxTfdt:
			if len(box) < headerSize+8 {
				return 0, nil
			}
			version := box[headerSize]
			valueOffset := headerSize + 4

			switch version {
			case 1:
				if len(box) < headerSize+12 {
					return 0, nil
				}
				value := binary.BigEndian.Uint64(box[valueOffset : valueOffset+8])
				decodeTime = &value
			case 0:
				value := uint64(binary.BigEndian.Uint32(box[valueOffset : valueOffset+4]))
				decodeTime = &value
			default:
				return 0, nil
			}
		}
	}

	return trackID, decodeTime
}

func getTrackTimescale(init []byte, requiredTrackID uint32) uint32 {
	position := 0
	for {
		boxType, boxHeaderSize, box, ok := tryReadBox(init, &position)
		if !ok {
			break
		}
		if boxType != boxMoov {
			continue
		}

		moovPosition := boxHeaderSize
		for {
			childType, childHeaderSize, child, ok := tryReadBox(box, &moovPosition)
			if !ok {
				break
			}
			if childType != boxTrak {
				continue
			}

			var trackID uint32
			var timescale uint32
			trakPosition := childHeaderSize

			for {
				trakType, trakBoxHeaderSize, trakBox, ok := tryReadBox(child, &trakPosition)
				if !ok {
					break
				}

				switch trakType {
				case boxTkhd:
					if len(trakBox) <= trakBoxHeaderSize {
						continue
					}

					version := trakBox[trakBoxHeaderSize]
					var trackIDOffset int

					switch version {
					case 1:
						trackIDOffset = trakBoxHeaderSize + 20
					case 0:
						trackIDOffset = trakBoxHeaderSize + 12
					default:
						continue
					}

					if len(trakBox) >= trackIDOffset+4 {
						trackID = binary.BigEndian.Uint32(trakBox[trackIDOffset : trackIDOffset+4])
					}

				case boxMdia:
					timescale = getMdiaTimescale(trakBox)
				}
			}

			if trackID == requiredTrackID {
				return timescale
			}
		}
	}

	return 0
}

func getMdiaTimescale(mdia []byte) uint32 {
	rootPosition := 0
	rootType, mdiaHeaderSize, mdiaBox, ok := tryReadBox(mdia, &rootPosition)
	if !ok || rootType != boxMdia || rootPosition != len(mdia) {
		return 0
	}

	position := mdiaHeaderSize
	for {
		boxType, headerSize, box, ok := tryReadBox(mdiaBox, &position)
		if !ok {
			break
		}
		if boxType != boxMdhd {
			continue
		}
		if len(box) <= headerSize {
			return 0
		}

		version := box[headerSize]
		var timescaleOffset int

		switch version {
		case 1:
			timescaleOffset = headerSize + 20
		case 0:
			timescaleOffset = headerSize + 12
		default:
			return 0
		}

		if len(box) < timescaleOffset+4 {
			return 0
		}

		return binary.BigEndian.Uint32(box[timescaleOffset : timescaleOffset+4])
	}

	return 0
}

func tryReadBox(data []byte, position *int) (uint32, int, []byte, bool) {
	start := *position
	if start < 0 || start > len(data) || len(data)-start < 8 {
		return 0, 0, nil, false
	}

	size32 := binary.BigEndian.Uint32(data[start : start+4])
	boxType := binary.BigEndian.Uint32(data[start+4 : start+8])

	size := uint64(size32)
	headerSize := 8

	if size32 == 1 {
		if len(data)-start < 16 {
			return 0, 0, nil, false
		}
		size = binary.BigEndian.Uint64(data[start+8 : start+16])
		headerSize = 16
	} else if size32 == 0 {
		size = uint64(len(data) - start)
	}

	if size < uint64(headerSize) || size > math.MaxInt32 || size > uint64(len(data)-start) {
		return 0, 0, nil, false
	}

	boxSize := int(size)
	box := data[start : start+boxSize]
	*position = start + boxSize

	return boxType, headerSize, box, true
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func minUint64(a uint64, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
