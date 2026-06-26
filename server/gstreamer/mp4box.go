package gstreamer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/bits"
)

const audioBoundaryToleranceSeconds = 0.100

type boxTarget int

const (
	boxTargetNone boxTarget = iota
	boxTargetInit
	boxTargetMoof
	boxTargetPayload
	boxTargetStyp
	boxTargetPrefix
)

const (
	boxStyp = uint32('s')<<24 | uint32('t')<<16 | uint32('y')<<8 | uint32('p')
	boxSidx = uint32('s')<<24 | uint32('i')<<16 | uint32('d')<<8 | uint32('x')
	boxEmsg = uint32('e')<<24 | uint32('m')<<16 | uint32('s')<<8 | uint32('g')
	boxFree = uint32('f')<<24 | uint32('r')<<16 | uint32('e')<<8 | uint32('e')
	boxPrft = uint32('p')<<24 | uint32('r')<<16 | uint32('f')<<8 | uint32('t')
	boxMoov = uint32('m')<<24 | uint32('o')<<16 | uint32('o')<<8 | uint32('v')
	boxMoof = uint32('m')<<24 | uint32('o')<<16 | uint32('o')<<8 | uint32('f')
	boxMdat = uint32('m')<<24 | uint32('d')<<16 | uint32('a')<<8 | uint32('t')
	boxMfhd = uint32('m')<<24 | uint32('f')<<16 | uint32('h')<<8 | uint32('d')
	boxTraf = uint32('t')<<24 | uint32('r')<<16 | uint32('a')<<8 | uint32('f')
	boxTfhd = uint32('t')<<24 | uint32('f')<<16 | uint32('h')<<8 | uint32('d')
	boxTfdt = uint32('t')<<24 | uint32('f')<<16 | uint32('d')<<8 | uint32('t')
	boxTrun = uint32('t')<<24 | uint32('r')<<16 | uint32('u')<<8 | uint32('n')
	boxTrak = uint32('t')<<24 | uint32('r')<<16 | uint32('a')<<8 | uint32('k')
	boxTkhd = uint32('t')<<24 | uint32('k')<<16 | uint32('h')<<8 | uint32('d')
	boxMdia = uint32('m')<<24 | uint32('d')<<16 | uint32('i')<<8 | uint32('a')
	boxMdhd = uint32('m')<<24 | uint32('d')<<16 | uint32('h')<<8 | uint32('d')
	boxHdlr = uint32('h')<<24 | uint32('d')<<16 | uint32('l')<<8 | uint32('r')
	boxMvex = uint32('m')<<24 | uint32('v')<<16 | uint32('e')<<8 | uint32('x')
	boxTrex = uint32('t')<<24 | uint32('r')<<16 | uint32('e')<<8 | uint32('x')

	handlerVideo = uint32('v')<<24 | uint32('i')<<16 | uint32('d')<<8 | uint32('e')
	handlerAudio = uint32('s')<<24 | uint32('o')<<16 | uint32('u')<<8 | uint32('n')

	tfhdBaseDataOffsetPresent         uint32 = 0x000001
	tfhdSampleDescriptionIndexPresent uint32 = 0x000002
	tfhdDefaultSampleDurationPresent  uint32 = 0x000008
	tfhdDefaultSampleSizePresent      uint32 = 0x000010
	tfhdDefaultSampleFlagsPresent     uint32 = 0x000020
	tfhdDefaultBaseIsMoof             uint32 = 0x020000
	trunDataOffsetPresent             uint32 = 0x000001
	trunFirstSampleFlagsPresent       uint32 = 0x000004
	trunSampleDurationPresent         uint32 = 0x000100
	trunSampleSizePresent             uint32 = 0x000200
	trunSampleFlagsPresent            uint32 = 0x000400
	trunCompositionTimeOffsetPresent  uint32 = 0x000800
)

type trexInfo struct {
	duration uint32
	size     uint32
	flags    uint32
}

type trackInfo struct {
	id        uint32
	timescale uint32
	trex      trexInfo
}

type trexEntry struct {
	trackID uint32
	value   trexInfo
}

type mp4Run struct {
	box                 []byte
	dataOffsetField     int
	sourceDataOffset    int32
	hasSourceDataOffset bool
	duration            uint64
	dataSize            uint64
	payloadOffset       int64
	outputOffset        int64
	startsWithSync      bool
}

type mp4Fragment struct {
	trackID        uint32
	timescale      uint32
	decodeTime     uint64
	duration       uint64
	startsWithSync bool
	tfhd           []byte
	runs           []mp4Run
	payload        []byte
}

func (f *mp4Fragment) endTime() uint64 {
	return f.decodeTime + f.duration
}

type mp4BoxReader struct {
	onInit    func([]byte)
	onSegment func(Segment)

	segmentSeconds float64

	init       bytes.Buffer
	sourceMoof bytes.Buffer
	sourceStyp bytes.Buffer
	deferred   bytes.Buffer

	video []mp4Fragment
	audio []mp4Fragment

	sourcePayload bytes.Buffer
	prefix        bytes.Buffer
	prefixActive  bool

	pending *mp4Fragment
	styp    []byte

	boxHeader         [16]byte
	boxHeaderLength   int
	boxHeaderRequired int

	currentBoxType      uint32
	currentBoxRemaining uint64
	currentTarget       boxTarget

	initDone              bool
	moovCompleted         bool
	sourcePayloadFromMoof int64

	videoTrack trackInfo
	audioTrack trackInfo

	tfdtOffsetSeconds float64
	sequence          uint32
}

func Mp4BoxReader(onInit func([]byte), onSegment func(Segment), segmentSeconds float64) *mp4BoxReader {
	reader := &mp4BoxReader{
		onInit:            onInit,
		onSegment:         onSegment,
		segmentSeconds:    segmentSeconds,
		boxHeaderRequired: 8,
		sequence:          1,
	}
	if math.IsNaN(segmentSeconds) || math.IsInf(segmentSeconds, 0) || segmentSeconds <= 0 {
		reader.segmentSeconds = 6
	}
	reader.sourceMoof.Grow(16 * 1024)
	reader.sourceStyp.Grow(128)
	reader.deferred.Grow(64 * 1024)
	return reader
}

func (r *mp4BoxReader) SeekReset(seconds float64) {
	r.initDone = false
	r.moovCompleted = false
	r.videoTrack = trackInfo{}
	r.audioTrack = trackInfo{}
	r.sequence = 1
	r.styp = nil

	if !math.IsNaN(seconds) && !math.IsInf(seconds, 0) && seconds > 0 {
		r.tfdtOffsetSeconds = seconds
	} else {
		r.tfdtOffsetSeconds = 0
	}

	r.init.Reset()
	r.sourceMoof.Reset()
	r.sourceStyp.Reset()
	r.deferred.Reset()
	r.clearSource()
	r.clearFragments()
	r.resetPrefix()
	r.resetBoxState()
}

func (r *mp4BoxReader) Push(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	completed, err := r.TryProcessDeferred()
	if err != nil {
		return err
	}
	if completed {
		_, _ = r.deferred.Write(data)
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
		_, _ = r.deferred.Write(data[consumed:])
	}
	return nil
}

func (r *mp4BoxReader) TryProcessDeferred() (bool, error) {
	completed, err := r.tryBuildSegment()
	if err != nil {
		return false, err
	}
	if completed {
		return true, nil
	}

	if r.deferred.Len() == 0 {
		return false, nil
	}

	data := r.deferred.Bytes()
	length := len(data)
	consumed, completed, err := r.processBytes(data)
	if err != nil {
		return false, err
	}

	if completed {
		r.keepDeferred(consumed)
		return true, nil
	}

	if consumed != length {
		return false, fmt.Errorf("mp4 parser consumed %d of %d deferred bytes", consumed, length)
	}

	r.deferred.Reset()
	return false, nil
}

func (r *mp4BoxReader) TryFlushFinalSegment() (bool, error) {
	if len(r.video) == 0 || len(r.audio) == 0 {
		return false, nil
	}
	if !r.video[0].startsWithSync {
		return false, nil
	}

	videoCount := len(r.video)
	videoEnd := r.video[videoCount-1].endTime()

	audioCount, err := r.selectAudioCount(videoEnd)
	if err != nil {
		return false, err
	}
	if audioCount == 0 {
		audioCount = len(r.audio)
	}
	if audioCount == 0 {
		return false, nil
	}

	if err := r.buildSegment(videoCount, audioCount); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mp4BoxReader) TakePendingSegment() (Segment, bool) {
	if r.pending == nil || r.sourcePayload.Len() == 0 {
		return Segment{}, false
	}

	payload := takeBuffer(&r.sourcePayload)
	mdatHeaderSize := 8
	if uint64(len(payload))+8 > math.MaxUint32 {
		mdatHeaderSize = 16
	}

	var header bytes.Buffer
	header.Grow(len(r.styp) + r.prefix.Len() + r.sourceMoof.Len() + mdatHeaderSize)

	if r.styp != nil {
		_, _ = header.Write(r.styp)
	}
	if r.prefixActive && r.prefix.Len() > 0 {
		_, _ = header.Write(r.prefix.Bytes())
	}
	_, _ = header.Write(r.sourceMoof.Bytes())
	writeMdatHeader(&header, uint64(len(payload)), mdatHeaderSize)

	startSeconds := 0.0
	if r.pending.timescale > 0 {
		startSeconds = float64(r.pending.decodeTime) / float64(r.pending.timescale)
	}

	r.clearSource()
	r.resetPrefix()

	return Segment{
		Header:       takeBuffer(&header),
		Payloads:     [][]byte{payload},
		StartSeconds: startSeconds,
	}, true
}

func (r *mp4BoxReader) processBytes(data []byte) (int, bool, error) {
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
					return position, false, errors.New("top-level mp4 box size=0 is not supported")
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

func (r *mp4BoxReader) beginBox(size uint64, headerSize int) error {
	if size < uint64(headerSize) {
		return errors.New("invalid mp4 box size")
	}
	if (r.currentBoxType == boxMoof || r.currentBoxType == boxMdat) && size > math.MaxInt32 {
		return errors.New("moof/mdat is too large")
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
			return errors.New("mdat appeared before init was completed")
		}

		r.currentTarget = boxTargetInit
		r.writeCurrentBoxData(r.boxHeader[:headerSize])
		return nil
	}

	switch r.currentBoxType {
	case boxMoof:
		if r.pending != nil {
			return errors.New("a new moof appeared before the previous mdat")
		}

		r.sourceMoof.Reset()
		r.sourcePayloadFromMoof = 0
		r.currentTarget = boxTargetMoof
		r.writeCurrentBoxData(r.boxHeader[:headerSize])
		return nil

	case boxMdat:
		if r.pending == nil {
			return errors.New("mdat does not follow a supported moof")
		}

		r.sourcePayload.Reset()
		r.sourcePayloadFromMoof += int64(headerSize)
		r.currentTarget = boxTargetPayload
		return nil

	case boxSidx:
		if r.pending != nil {
			r.sourcePayloadFromMoof += int64(size)
		}
		return nil

	case boxStyp:
		if r.pending != nil {
			return errors.New("styp cannot appear between moof and mdat")
		}

		r.sourceStyp.Reset()
		r.currentTarget = boxTargetStyp
		r.writeCurrentBoxData(r.boxHeader[:headerSize])
		return nil

	case boxEmsg, boxFree, boxPrft:
		if r.pending != nil {
			r.sourcePayloadFromMoof += int64(size)
		}

		r.prefixActive = true
		r.currentTarget = boxTargetPrefix
		r.writeCurrentBoxData(r.boxHeader[:headerSize])
		return nil
	}

	return fmt.Errorf("unsupported top-level mp4 box after init: %s", fourCC(r.currentBoxType))
}

func (r *mp4BoxReader) writeCurrentBoxData(data []byte) {
	if len(data) == 0 {
		return
	}

	switch r.currentTarget {
	case boxTargetInit:
		_, _ = r.init.Write(data)
	case boxTargetMoof:
		_, _ = r.sourceMoof.Write(data)
	case boxTargetPayload:
		_, _ = r.sourcePayload.Write(data)
	case boxTargetStyp:
		_, _ = r.sourceStyp.Write(data)
	case boxTargetPrefix:
		_, _ = r.prefix.Write(data)
	}
}

func (r *mp4BoxReader) completeBox() (bool, error) {
	switch r.currentBoxType {
	case boxStyp:
		if r.styp == nil && r.sourceStyp.Len() > 0 {
			r.styp = cloneBytes(r.sourceStyp.Bytes())
		}
		r.sourceStyp.Reset()
		return false, nil

	case boxMoov:
		if r.initDone {
			return false, errors.New("unexpected moov after mp4 initialization")
		}
		r.moovCompleted = true
		return false, nil

	case boxMoof:
		return false, r.completeMoof()

	case boxMdat:
		if err := r.completeMdat(); err != nil {
			return false, err
		}
		completed, err := r.tryBuildSegment()
		if err != nil {
			return false, err
		}
		return completed, nil
	}

	return false, nil
}

func (r *mp4BoxReader) completeInit() error {
	if !r.moovCompleted || r.init.Len() == 0 {
		return errors.New("incomplete mp4 initialization")
	}

	init := cloneBytes(r.init.Bytes())
	video, audio, err := parseInitTracks(init)
	if err != nil {
		return fmt.Errorf("unable to parse mp4 initialization: %w", err)
	}

	r.videoTrack = video
	r.audioTrack = audio
	r.initDone = true
	r.onInit(init)
	return nil
}

func (r *mp4BoxReader) completeMoof() error {
	fragment, err := parseSourceMoof(r.sourceMoof.Bytes(), r.videoTrack, r.audioTrack)
	if err != nil {
		return fmt.Errorf("unable to parse source moof: %w", err)
	}

	r.pending = fragment
	r.sourcePayloadFromMoof = int64(r.sourceMoof.Len())
	return nil
}

func (r *mp4BoxReader) completeMdat() error {
	if r.pending == nil {
		return errors.New("completed mdat has no source moof")
	}

	payload := takeBuffer(&r.sourcePayload)
	if err := attachPayload(r.pending, payload, r.sourcePayloadFromMoof); err != nil {
		return err
	}

	switch r.pending.trackID {
	case r.videoTrack.id:
		r.video = append(r.video, *r.pending)
	case r.audioTrack.id:
		r.audio = append(r.audio, *r.pending)
	default:
		return fmt.Errorf("unsupported track_ID=%d", r.pending.trackID)
	}

	r.pending = nil
	r.sourcePayloadFromMoof = 0
	r.sourceMoof.Reset()
	return nil
}

func attachPayload(fragment *mp4Fragment, payload []byte, payloadFromMoof int64) error {
	var expected int64

	for i := range fragment.runs {
		run := &fragment.runs[i]
		offset := expected
		if run.hasSourceDataOffset {
			offset = int64(run.sourceDataOffset) - payloadFromMoof
		}

		if offset != expected {
			return fmt.Errorf("non-contiguous source mdat: expected=%d, actual=%d", expected, offset)
		}
		if run.dataSize > uint64(math.MaxInt64) {
			return errors.New("trun payload is too large")
		}

		run.payloadOffset = offset
		expected = offset + int64(run.dataSize)
	}

	if expected != int64(len(payload)) {
		return fmt.Errorf("source mdat size mismatch: trun=%d, mdat=%d", expected, len(payload))
	}

	fragment.payload = payload
	return nil
}

func (r *mp4BoxReader) tryBuildSegment() (bool, error) {
	videoCount, err := r.selectVideoCount()
	if err != nil || videoCount == 0 {
		return false, err
	}

	videoEnd := r.video[videoCount-1].endTime()
	audioCount, err := r.selectAudioCount(videoEnd)
	if err != nil || audioCount == 0 {
		return false, err
	}

	if err := r.buildSegment(videoCount, audioCount); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mp4BoxReader) selectVideoCount() (int, error) {
	if len(r.video) == 0 {
		return 0, nil
	}

	if !r.video[0].startsWithSync {
		return 0, fmt.Errorf("video segment starts with a non-sync sample at %.6fs", float64(r.video[0].decodeTime)/float64(r.videoTrack.timescale))
	}

	target, err := toUnits(r.segmentSeconds, r.videoTrack.timescale)
	if err != nil {
		return 0, err
	}

	var duration uint64
	for i := 0; i+1 < len(r.video); i++ {
		duration += r.video[i].duration
		if duration >= target && r.video[i+1].startsWithSync {
			return i + 1, nil
		}
	}

	return 0, nil
}

func (r *mp4BoxReader) selectAudioCount(videoEnd uint64) (int, error) {
	if len(r.audio) == 0 {
		return 0, nil
	}

	tolerance, err := toUnits(audioBoundaryToleranceSeconds, r.audioTrack.timescale)
	if err != nil {
		return 0, err
	}

	for i := range r.audio {
		audioEnd := r.audio[i].endTime()
		withTolerance := audioEnd + tolerance
		if withTolerance < audioEnd {
			withTolerance = math.MaxUint64
		}

		if scaledGreaterOrEqual(withTolerance, r.videoTrack.timescale, videoEnd, r.audioTrack.timescale) {
			return i + 1, nil
		}
	}

	return 0, nil
}

func (r *mp4BoxReader) buildSegment(videoCount int, audioCount int) error {
	if err := validateTrack(r.video, videoCount); err != nil {
		return err
	}
	if err := validateTrack(r.audio, audioCount); err != nil {
		return err
	}

	var payloadLength int64
	assignOffsets(r.video, videoCount, &payloadLength)
	assignOffsets(r.audio, audioCount, &payloadLength)

	videoTrafSize := getTrafSize(r.video, videoCount)
	audioTrafSize := getTrafSize(r.audio, audioCount)
	moofSize64 := int64(8 + 16 + videoTrafSize + audioTrafSize)
	if moofSize64 > math.MaxUint32 {
		return errors.New("combined moof is too large")
	}

	moofSize := uint32(moofSize64)
	mdatHeaderSize := 8
	if uint64(payloadLength)+8 > math.MaxUint32 {
		mdatHeaderSize = 16
	}

	var header bytes.Buffer
	header.Grow(int(moofSize64) + mdatHeaderSize + len(r.styp) + r.prefix.Len())

	if r.styp != nil {
		_, _ = header.Write(r.styp)
	}
	if r.prefixActive && r.prefix.Len() > 0 {
		_, _ = header.Write(r.prefix.Bytes())
	}

	writeHeader(&header, moofSize, boxMoof)
	writeMfhd(&header, r.sequence)
	r.sequence++
	if err := r.writeTraf(&header, r.video, videoCount, moofSize, mdatHeaderSize); err != nil {
		return err
	}
	if err := r.writeTraf(&header, r.audio, audioCount, moofSize, mdatHeaderSize); err != nil {
		return err
	}
	writeMdatHeader(&header, uint64(payloadLength), mdatHeaderSize)

	startSeconds := float64(r.video[0].decodeTime) / float64(r.video[0].timescale)
	r.onSegment(Segment{
		Header:       takeBuffer(&header),
		Payloads:     collectPayloads(r.video, videoCount, r.audio, audioCount),
		StartSeconds: startSeconds,
	})

	r.video = removeFragments(r.video, videoCount)
	r.audio = removeFragments(r.audio, audioCount)
	r.resetPrefix()
	return nil
}

func validateTrack(fragments []mp4Fragment, count int) error {
	first := fragments[0]
	expected := first.endTime()

	for i := 1; i < count; i++ {
		current := fragments[i]
		if current.trackID != first.trackID ||
			current.timescale != first.timescale ||
			current.decodeTime != expected ||
			!bytes.Equal(current.tfhd, first.tfhd) {
			return fmt.Errorf("track %d fragments cannot be merged into one traf", first.trackID)
		}

		expected = current.endTime()
	}
	return nil
}

func assignOffsets(fragments []mp4Fragment, count int, outputOffset *int64) {
	for i := 0; i < count; i++ {
		fragment := &fragments[i]
		baseOffset := *outputOffset

		for j := range fragment.runs {
			fragment.runs[j].outputOffset = baseOffset + fragment.runs[j].payloadOffset
		}

		*outputOffset += int64(len(fragment.payload))
	}
}

func getTrafSize(fragments []mp4Fragment, count int) int64 {
	size := int64(8 + len(fragments[0].tfhd) + 20)
	for i := 0; i < count; i++ {
		for _, run := range fragments[i].runs {
			size += int64(len(run.box))
		}
	}
	return size
}

func (r *mp4BoxReader) writeTraf(output io.Writer, fragments []mp4Fragment, count int, moofSize uint32, mdatHeaderSize int) error {
	size := getTrafSize(fragments, count)
	if size > math.MaxUint32 {
		return errors.New("combined traf is too large")
	}

	first := fragments[0]
	writeHeader(output, uint32(size), boxTraf)
	_, _ = output.Write(first.tfhd)
	writeTfdt(output, addTfdtOffset(first.decodeTime, first.timescale, r.tfdtOffsetSeconds))

	for i := 0; i < count; i++ {
		for j := range fragments[i].runs {
			run := &fragments[i].runs[j]
			dataOffset := int64(moofSize) + int64(mdatHeaderSize) + run.outputOffset
			if dataOffset < math.MinInt32 || dataOffset > math.MaxInt32 {
				return errors.New("trun.data_offset exceeds int32")
			}
			writePatchedTrun(output, run, int32(dataOffset))
		}
	}
	return nil
}

func writePatchedTrun(output io.Writer, run *mp4Run, dataOffset int32) {
	box := run.box
	field := run.dataOffsetField
	_, _ = output.Write(box[:field])

	var value [4]byte
	binary.BigEndian.PutUint32(value[:], uint32(dataOffset))
	_, _ = output.Write(value[:])
	_, _ = output.Write(box[field+4:])
}

func parseSourceMoof(moof []byte, videoTrack trackInfo, audioTrack trackInfo) (*mp4Fragment, error) {
	rootPosition := 0
	rootType, moofHeader, moofBox, ok := tryReadBox(moof, &rootPosition)
	if !ok || rootType != boxMoof || rootPosition != len(moof) {
		return nil, errors.New("buffer does not contain exactly one moof")
	}

	position := moofHeader
	trafCount := 0
	var fragment *mp4Fragment

	for {
		boxType, headerSize, box, ok := tryReadBox(moofBox, &position)
		if !ok {
			break
		}
		if boxType != boxTraf {
			continue
		}

		trafCount++
		if trafCount > 1 {
			return nil, errors.New("source moof must contain one traf")
		}

		parsed, err := parseTraf(box, headerSize, videoTrack, audioTrack)
		if err != nil {
			return nil, err
		}
		fragment = parsed
	}

	if fragment == nil {
		return nil, errors.New("traf was not found")
	}
	return fragment, nil
}

func parseTraf(traf []byte, trafHeader int, videoTrack trackInfo, audioTrack trackInfo) (*mp4Fragment, error) {
	var trackID uint32
	var defaultDuration uint32
	var defaultSize uint32
	var defaultFlags uint32
	var hasDefaultFlags bool
	var tfhd []byte
	var decodeTime uint64
	var hasTfdt bool

	position := trafHeader
	for {
		boxType, headerSize, box, ok := tryReadBox(traf, &position)
		if !ok {
			break
		}

		switch boxType {
		case boxTfhd:
			if tfhd != nil {
				return nil, errors.New("duplicate tfhd")
			}
			normalized, parsedTrackID, parsedDuration, parsedSize, parsedFlags, parsedHasFlags, err := normalizeTfhd(box, headerSize)
			if err != nil {
				return nil, err
			}
			tfhd = normalized
			trackID = parsedTrackID
			defaultDuration = parsedDuration
			defaultSize = parsedSize
			defaultFlags = parsedFlags
			hasDefaultFlags = parsedHasFlags

		case boxTfdt:
			if hasTfdt {
				return nil, errors.New("duplicate tfdt")
			}
			value, ok := readTfdt(box, headerSize)
			if !ok {
				return nil, errors.New("invalid tfdt")
			}
			decodeTime = value
			hasTfdt = true

		case boxTrun:
			// Parsed after tfhd/trex defaults are known.

		default:
			return nil, fmt.Errorf("unsupported box %s inside traf", fourCC(boxType))
		}
	}

	if tfhd == nil || !hasTfdt {
		return nil, errors.New("tfhd/tfdt was not found")
	}

	var timescale uint32
	var trex trexInfo
	switch trackID {
	case videoTrack.id:
		timescale = videoTrack.timescale
		trex = videoTrack.trex
	case audioTrack.id:
		timescale = audioTrack.timescale
		trex = audioTrack.trex
	default:
		return nil, fmt.Errorf("unsupported track_ID=%d", trackID)
	}

	if timescale == 0 {
		return nil, fmt.Errorf("timescale is zero for track_ID=%d", trackID)
	}

	if defaultDuration == 0 {
		defaultDuration = trex.duration
	}
	if defaultSize == 0 {
		defaultSize = trex.size
	}
	if !hasDefaultFlags {
		defaultFlags = trex.flags
	}

	result := &mp4Fragment{
		trackID:    trackID,
		timescale:  timescale,
		decodeTime: decodeTime,
		tfhd:       tfhd,
	}

	var duration uint64
	position = trafHeader
	for {
		boxType, headerSize, box, ok := tryReadBox(traf, &position)
		if !ok {
			break
		}
		if boxType != boxTrun {
			continue
		}

		run, err := normalizeTrun(box, headerSize, defaultDuration, defaultSize, defaultFlags)
		if err != nil {
			return nil, err
		}
		duration += run.duration
		result.runs = append(result.runs, run)
	}

	if len(result.runs) == 0 || duration == 0 {
		return nil, errors.New("trun/duration was not found")
	}

	result.duration = duration
	result.startsWithSync = result.runs[0].startsWithSync
	return result, nil
}

func normalizeTfhd(box []byte, headerSize int) ([]byte, uint32, uint32, uint32, uint32, bool, error) {
	if len(box) < headerSize+8 {
		return nil, 0, 0, 0, 0, false, errors.New("tfhd is too small")
	}

	versionFlags := binary.BigEndian.Uint32(box[headerSize : headerSize+4])
	version := byte(versionFlags >> 24)
	flags := versionFlags & 0x00ffffff
	if flags&tfhdBaseDataOffsetPresent != 0 {
		return nil, 0, 0, 0, 0, false, errors.New("tfhd.base-data-offset-present is not supported")
	}

	trackID := binary.BigEndian.Uint32(box[headerSize+4 : headerSize+8])
	optionalStart := headerSize + 8
	cursor := optionalStart

	if flags&tfhdSampleDescriptionIndexPresent != 0 && !skip(box, &cursor, 4) {
		return nil, 0, 0, 0, 0, false, errors.New("invalid tfhd sample_description_index")
	}

	var defaultDuration uint32
	var defaultSize uint32
	var defaultFlags uint32

	if flags&tfhdDefaultSampleDurationPresent != 0 {
		value, ok := readUint32(box, &cursor)
		if !ok {
			return nil, 0, 0, 0, 0, false, errors.New("invalid tfhd default_sample_duration")
		}
		defaultDuration = value
	}

	if flags&tfhdDefaultSampleSizePresent != 0 {
		value, ok := readUint32(box, &cursor)
		if !ok {
			return nil, 0, 0, 0, 0, false, errors.New("invalid tfhd default_sample_size")
		}
		defaultSize = value
	}

	hasDefaultFlags := flags&tfhdDefaultSampleFlagsPresent != 0
	if hasDefaultFlags {
		value, ok := readUint32(box, &cursor)
		if !ok {
			return nil, 0, 0, 0, 0, false, errors.New("invalid tfhd default_sample_flags")
		}
		defaultFlags = value
	}

	if cursor != len(box) || trackID == 0 {
		return nil, 0, 0, 0, 0, false, errors.New("invalid tfhd body")
	}

	optionalLength := cursor - optionalStart
	size := 16 + optionalLength
	normalized := make([]byte, size)
	binary.BigEndian.PutUint32(normalized[0:4], uint32(size))
	binary.BigEndian.PutUint32(normalized[4:8], boxTfhd)
	binary.BigEndian.PutUint32(normalized[8:12], uint32(version)<<24|flags&^tfhdBaseDataOffsetPresent|tfhdDefaultBaseIsMoof)
	binary.BigEndian.PutUint32(normalized[12:16], trackID)
	copy(normalized[16:], box[optionalStart:cursor])

	return normalized, trackID, defaultDuration, defaultSize, defaultFlags, hasDefaultFlags, nil
}

func normalizeTrun(box []byte, headerSize int, defaultDuration uint32, defaultSize uint32, defaultFlags uint32) (mp4Run, error) {
	var run mp4Run
	if len(box) < headerSize+8 {
		return run, errors.New("trun is too small")
	}

	versionFlags := binary.BigEndian.Uint32(box[headerSize : headerSize+4])
	version := byte(versionFlags >> 24)
	flags := versionFlags & 0x00ffffff
	sampleCount := binary.BigEndian.Uint32(box[headerSize+4 : headerSize+8])

	cursor := headerSize + 8
	if flags&trunDataOffsetPresent != 0 {
		if len(box)-cursor < 4 {
			return run, errors.New("invalid trun data_offset")
		}
		run.sourceDataOffset = int32(binary.BigEndian.Uint32(box[cursor : cursor+4]))
		run.hasSourceDataOffset = true
		cursor += 4
	}

	hasFirstSampleFlags := flags&trunFirstSampleFlagsPresent != 0
	firstSampleFlags := defaultFlags
	if hasFirstSampleFlags {
		value, ok := readUint32(box, &cursor)
		if !ok {
			return run, errors.New("invalid trun first_sample_flags")
		}
		firstSampleFlags = value
	}

	hasDuration := flags&trunSampleDurationPresent != 0
	hasSize := flags&trunSampleSizePresent != 0
	if !hasDuration && defaultDuration == 0 {
		return run, errors.New("sample duration is absent")
	}
	if !hasSize && defaultSize == 0 {
		return run, errors.New("sample size is absent")
	}

	var duration uint64
	var dataSize uint64
	for i := uint32(0); i < sampleCount; i++ {
		sampleDuration := defaultDuration
		sampleSize := defaultSize

		if hasDuration {
			value, ok := readUint32(box, &cursor)
			if !ok {
				return run, errors.New("invalid trun sample_duration")
			}
			sampleDuration = value
		}

		if hasSize {
			value, ok := readUint32(box, &cursor)
			if !ok {
				return run, errors.New("invalid trun sample_size")
			}
			sampleSize = value
		}

		duration += uint64(sampleDuration)
		dataSize += uint64(sampleSize)

		if flags&trunSampleFlagsPresent != 0 {
			sampleFlags, ok := readUint32(box, &cursor)
			if !ok {
				return run, errors.New("invalid trun sample_flags")
			}
			if i == 0 && !hasFirstSampleFlags {
				firstSampleFlags = sampleFlags
			}
		}

		if flags&trunCompositionTimeOffsetPresent != 0 && !skip(box, &cursor, 4) {
			return run, errors.New("invalid trun composition_time_offset")
		}
	}

	if cursor != len(box) || sampleCount == 0 || duration == 0 || dataSize == 0 {
		return run, errors.New("invalid trun body")
	}

	hadOffset := flags&trunDataOffsetPresent != 0
	bodyLength := len(box) - headerSize
	normalizedSize := 8 + bodyLength
	if !hadOffset {
		normalizedSize += 4
	}

	normalized := make([]byte, normalizedSize)
	binary.BigEndian.PutUint32(normalized[0:4], uint32(normalizedSize))
	binary.BigEndian.PutUint32(normalized[4:8], boxTrun)
	binary.BigEndian.PutUint32(normalized[8:12], uint32(version)<<24|flags|trunDataOffsetPresent)
	binary.BigEndian.PutUint32(normalized[12:16], sampleCount)

	if hadOffset {
		copy(normalized[16:], box[headerSize+8:])
	} else {
		binary.BigEndian.PutUint32(normalized[16:20], 0)
		copy(normalized[20:], box[headerSize+8:])
	}

	run.box = normalized
	run.dataOffsetField = 16
	run.duration = duration
	run.dataSize = dataSize
	run.startsWithSync = isSyncSample(firstSampleFlags)
	return run, nil
}

func isSyncSample(flags uint32) bool {
	const nonSync = 0x00010000
	dependsOn := (flags >> 24) & 0x03
	return flags&nonSync == 0 && dependsOn != 1
}

func readTfdt(box []byte, headerSize int) (uint64, bool) {
	if len(box) < headerSize+8 {
		return 0, false
	}

	version := box[headerSize]
	offset := headerSize + 4
	switch version {
	case 1:
		if len(box) < offset+8 {
			return 0, false
		}
		return binary.BigEndian.Uint64(box[offset : offset+8]), true
	case 0:
		if len(box) < offset+4 {
			return 0, false
		}
		return uint64(binary.BigEndian.Uint32(box[offset : offset+4])), true
	}

	return 0, false
}

func parseInitTracks(init []byte) (trackInfo, trackInfo, error) {
	moov, moovHeader, ok := findBox(init, boxMoov)
	if !ok {
		return trackInfo{}, trackInfo{}, errors.New("moov was not found")
	}

	var videoID uint32
	var videoTimescale uint32
	var audioID uint32
	var audioTimescale uint32
	var trexEntries []trexEntry

	position := moovHeader
	for {
		boxType, header, box, ok := tryReadBox(moov, &position)
		if !ok {
			break
		}

		if boxType == boxTrak {
			trackID, timescale, handler := readTrack(box, header)
			switch handler {
			case handlerVideo:
				if videoID != 0 {
					return trackInfo{}, trackInfo{}, errors.New("multiple video tracks in mp4mux output")
				}
				videoID = trackID
				videoTimescale = timescale
			case handlerAudio:
				if audioID != 0 {
					return trackInfo{}, trackInfo{}, errors.New("multiple audio tracks in mp4mux output")
				}
				audioID = trackID
				audioTimescale = timescale
			}
			continue
		}

		if boxType != boxMvex {
			continue
		}

		mvexPosition := header
		for {
			childType, childHeader, child, ok := tryReadBox(box, &mvexPosition)
			if !ok {
				break
			}
			if childType != boxTrex {
				continue
			}

			entry, ok := readTrex(child, childHeader)
			if !ok {
				return trackInfo{}, trackInfo{}, errors.New("invalid trex")
			}
			trexEntries = append(trexEntries, entry)
		}
	}

	if videoID == 0 || videoTimescale == 0 {
		return trackInfo{}, trackInfo{}, errors.New("video track was not found through hdlr=vide")
	}
	if audioID == 0 || audioTimescale == 0 {
		return trackInfo{}, trackInfo{}, errors.New("audio track was not found through hdlr=soun")
	}

	return trackInfo{
			id:        videoID,
			timescale: videoTimescale,
			trex:      findTrex(trexEntries, videoID),
		}, trackInfo{
			id:        audioID,
			timescale: audioTimescale,
			trex:      findTrex(trexEntries, audioID),
		}, nil
}

func readTrack(trak []byte, trakHeader int) (uint32, uint32, uint32) {
	var trackID uint32
	var timescale uint32
	var handler uint32

	position := trakHeader
	for {
		boxType, header, box, ok := tryReadBox(trak, &position)
		if !ok {
			break
		}

		if boxType == boxTkhd {
			trackID = readTkhdTrackID(box, header)
			continue
		}
		if boxType != boxMdia {
			continue
		}

		mdiaPosition := header
		for {
			mdiaType, mdiaHeader, child, ok := tryReadBox(box, &mdiaPosition)
			if !ok {
				break
			}
			switch mdiaType {
			case boxMdhd:
				timescale = readMdhdTimescale(child, mdiaHeader)
			case boxHdlr:
				handler = readHandlerType(child, mdiaHeader)
			}
		}
	}

	return trackID, timescale, handler
}

func readTkhdTrackID(box []byte, header int) uint32 {
	if len(box) <= header {
		return 0
	}

	offset := -1
	switch box[header] {
	case 1:
		offset = header + 20
	case 0:
		offset = header + 12
	}

	if offset >= 0 && len(box) >= offset+4 {
		return binary.BigEndian.Uint32(box[offset : offset+4])
	}
	return 0
}

func readMdhdTimescale(box []byte, header int) uint32 {
	if len(box) <= header {
		return 0
	}

	offset := -1
	switch box[header] {
	case 1:
		offset = header + 20
	case 0:
		offset = header + 12
	}

	if offset >= 0 && len(box) >= offset+4 {
		return binary.BigEndian.Uint32(box[offset : offset+4])
	}
	return 0
}

func readHandlerType(box []byte, header int) uint32 {
	offset := header + 8
	if len(box) >= offset+4 {
		return binary.BigEndian.Uint32(box[offset : offset+4])
	}
	return 0
}

func readTrex(box []byte, header int) (trexEntry, bool) {
	if len(box) < header+24 {
		return trexEntry{}, false
	}

	trackID := binary.BigEndian.Uint32(box[header+4 : header+8])
	if trackID == 0 {
		return trexEntry{}, false
	}

	return trexEntry{
		trackID: trackID,
		value: trexInfo{
			duration: binary.BigEndian.Uint32(box[header+12 : header+16]),
			size:     binary.BigEndian.Uint32(box[header+16 : header+20]),
			flags:    binary.BigEndian.Uint32(box[header+20 : header+24]),
		},
	}, true
}

func findTrex(entries []trexEntry, trackID uint32) trexInfo {
	for _, entry := range entries {
		if entry.trackID == trackID {
			return entry.value
		}
	}
	return trexInfo{}
}

func findBox(data []byte, requiredType uint32) ([]byte, int, bool) {
	position := 0
	for position < len(data) {
		start := position
		boxType, header, _, ok := tryReadBox(data, &position)
		if !ok {
			return nil, 0, false
		}
		if boxType != requiredType {
			continue
		}
		return data[start:position], header, true
	}
	return nil, 0, false
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

func readUint32(data []byte, position *int) (uint32, bool) {
	if *position < 0 || len(data)-*position < 4 {
		return 0, false
	}
	value := binary.BigEndian.Uint32(data[*position : *position+4])
	*position += 4
	return value, true
}

func skip(data []byte, position *int, count int) bool {
	if count < 0 || *position < 0 || len(data)-*position < count {
		return false
	}
	*position += count
	return true
}

func toUnits(seconds float64, timescale uint32) (uint64, error) {
	value := seconds * float64(timescale)
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > float64(math.MaxUint64) {
		return 0, errors.New("invalid timeline value")
	}
	return uint64(math.Ceil(value)), nil
}

func addTfdtOffset(value uint64, timescale uint32, seconds float64) uint64 {
	if seconds <= 0 {
		return value
	}

	units := seconds * float64(timescale)
	if math.IsNaN(units) || math.IsInf(units, 0) || units < 0 || units > float64(math.MaxUint64) {
		panic("invalid tfdt offset")
	}

	offset := uint64(math.Round(units))
	if math.MaxUint64-value < offset {
		panic("tfdt offset overflow")
	}
	return value + offset
}

func writeTfdt(output io.Writer, decodeTime uint64) {
	var box [20]byte
	binary.BigEndian.PutUint32(box[0:4], 20)
	binary.BigEndian.PutUint32(box[4:8], boxTfdt)
	binary.BigEndian.PutUint32(box[8:12], 0x01000000)
	binary.BigEndian.PutUint64(box[12:20], decodeTime)
	_, _ = output.Write(box[:])
}

func writeMfhd(output io.Writer, sequence uint32) {
	var box [16]byte
	binary.BigEndian.PutUint32(box[0:4], 16)
	binary.BigEndian.PutUint32(box[4:8], boxMfhd)
	binary.BigEndian.PutUint32(box[8:12], 0)
	binary.BigEndian.PutUint32(box[12:16], sequence)
	_, _ = output.Write(box[:])
}

func writeHeader(output io.Writer, size uint32, boxType uint32) {
	var header [8]byte
	binary.BigEndian.PutUint32(header[0:4], size)
	binary.BigEndian.PutUint32(header[4:8], boxType)
	_, _ = output.Write(header[:])
}

func writeMdatHeader(output io.Writer, payloadLength uint64, headerSize int) {
	if headerSize == 8 {
		writeHeader(output, uint32(payloadLength+8), boxMdat)
		return
	}

	var header [16]byte
	binary.BigEndian.PutUint32(header[0:4], 1)
	binary.BigEndian.PutUint32(header[4:8], boxMdat)
	binary.BigEndian.PutUint64(header[8:16], payloadLength+16)
	_, _ = output.Write(header[:])
}

func collectPayloads(video []mp4Fragment, videoCount int, audio []mp4Fragment, audioCount int) [][]byte {
	payloads := make([][]byte, 0, videoCount+audioCount)
	payloads = appendFragmentPayloads(payloads, video, videoCount)
	payloads = appendFragmentPayloads(payloads, audio, audioCount)
	return payloads
}

func appendFragmentPayloads(payloads [][]byte, fragments []mp4Fragment, count int) [][]byte {
	for i := 0; i < count; i++ {
		if len(fragments[i].payload) > 0 {
			payloads = append(payloads, fragments[i].payload)
		}
	}
	return payloads
}

func removeFragments(fragments []mp4Fragment, count int) []mp4Fragment {
	if count <= 0 {
		return fragments
	}
	copy(fragments, fragments[count:])
	for i := len(fragments) - count; i < len(fragments); i++ {
		fragments[i] = mp4Fragment{}
	}
	return fragments[:len(fragments)-count]
}

func (r *mp4BoxReader) clearFragments() {
	for i := range r.video {
		r.video[i] = mp4Fragment{}
	}
	for i := range r.audio {
		r.audio[i] = mp4Fragment{}
	}
	r.video = r.video[:0]
	r.audio = r.audio[:0]
}

func (r *mp4BoxReader) resetPrefix() {
	r.prefix.Reset()
	r.prefixActive = false
}

func (r *mp4BoxReader) clearSource() {
	r.pending = nil
	r.sourcePayload.Reset()
	r.sourcePayloadFromMoof = 0
	r.sourceMoof.Reset()
}

func (r *mp4BoxReader) resetBoxState() {
	r.boxHeaderLength = 0
	r.boxHeaderRequired = 8
	r.currentBoxType = 0
	r.currentBoxRemaining = 0
	r.currentTarget = boxTargetNone
}

func (r *mp4BoxReader) keepDeferred(consumed int) {
	count := r.deferred.Len() - consumed
	if count <= 0 {
		r.deferred.Reset()
		return
	}

	data := r.deferred.Bytes()
	copy(data, data[consumed:])
	r.deferred.Truncate(count)
}

func scaledGreaterOrEqual(left uint64, leftScale uint32, right uint64, rightScale uint32) bool {
	leftHi, leftLo := bits.Mul64(left, uint64(leftScale))
	rightHi, rightLo := bits.Mul64(right, uint64(rightScale))
	if leftHi != rightHi {
		return leftHi > rightHi
	}
	return leftLo >= rightLo
}

func fourCC(boxType uint32) string {
	return string([]byte{
		byte(boxType >> 24),
		byte(boxType >> 16),
		byte(boxType >> 8),
		byte(boxType),
	})
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
