//go:build gst

package gstreamer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/bits"
)

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
	boxMfra = uint32('m')<<24 | uint32('f')<<16 | uint32('r')<<8 | uint32('a')

	handlerVideo = uint32('v')<<24 | uint32('i')<<16 | uint32('d')<<8 | uint32('e')
	handlerAudio = uint32('s')<<24 | uint32('o')<<16 | uint32('u')<<8 | uint32('n')

	tfhdBaseDataOffsetPresent         uint32 = 0x000001
	tfhdSampleDescriptionIndexPresent uint32 = 0x000002
	tfhdDefaultSampleDurationPresent  uint32 = 0x000008
	tfhdDefaultSampleSizePresent      uint32 = 0x000010
	tfhdDefaultSampleFlagsPresent     uint32 = 0x000020
	tfhdDurationIsEmpty               uint32 = 0x010000
	tfhdDefaultBaseIsMoof             uint32 = 0x020000
	trunDataOffsetPresent             uint32 = 0x000001
	trunFirstSampleFlagsPresent       uint32 = 0x000004
	trunSampleDurationPresent         uint32 = 0x000100
	trunSampleSizePresent             uint32 = 0x000200
	trunSampleFlagsPresent            uint32 = 0x000400
	trunCompositionTimeOffsetPresent  uint32 = 0x000800
	maxTrunSamples                           = 1_000_000
	canonicalTfhdSize                        = 20
	smallBufferGrowthQuantum                 = 64 * 1024
	largeBufferGrowthQuantum                 = 1024 * 1024
	largeBufferThreshold                     = 1024 * 1024
)

type linearGrowthBuffer struct {
	data []byte
}

func (b *linearGrowthBuffer) Bytes() []byte {
	return b.data
}

func (b *linearGrowthBuffer) Len() int {
	return len(b.data)
}

func (b *linearGrowthBuffer) Cap() int {
	return cap(b.data)
}

func (b *linearGrowthBuffer) Reset() {
	b.data = b.data[:0]
}

func (b *linearGrowthBuffer) Truncate(length int) {
	if length < 0 || length > len(b.data) {
		panic("gstreamer buffer truncation out of range")
	}
	b.data = b.data[:length]
}

func (b *linearGrowthBuffer) Grow(additional int) {
	if additional < 0 || additional > int(^uint(0)>>1)-len(b.data) {
		panic("gstreamer buffer size overflow")
	}
	required := len(b.data) + additional
	if required <= cap(b.data) {
		return
	}

	nextCapacity := roundedBufferCapacity(required)
	if len(b.data) == 0 {
		b.data = nil
		b.data = make([]byte, 0, nextCapacity)
		return
	}

	next := make([]byte, len(b.data), nextCapacity)
	copy(next, b.data)
	b.data = next
}

func (b *linearGrowthBuffer) Write(data []byte) (int, error) {
	start := len(b.data)
	b.Grow(len(data))
	b.data = b.data[:start+len(data)]
	copy(b.data[start:], data)
	return len(data), nil
}

func roundedBufferCapacity(required int) int {
	if required <= 0 {
		return 0
	}

	quantum := smallBufferGrowthQuantum
	if required >= largeBufferThreshold {
		quantum = largeBufferGrowthQuantum
	}
	remainder := required % quantum
	if remainder == 0 {
		return required
	}
	delta := quantum - remainder
	if required > int(^uint(0)>>1)-delta {
		return required
	}
	return required + delta
}

type trexInfo struct {
	descriptionIndex uint32
	duration         uint32
	size             uint32
	flags            uint32
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

type tfhdInfo struct {
	version uint8
	trackID uint32

	sampleDescriptionIndex    uint32
	hasSampleDescriptionIndex bool

	defaultDuration uint32
	defaultSize     uint32
	defaultFlags    uint32
	hasDefaultFlags bool
}

type mp4Sample struct {
	duration                 uint32
	size                     uint32
	flags                    uint32
	compositionTimeOffsetRaw uint32
}

type mp4Run struct {
	sourceDataOffset    int32
	hasSourceDataOffset bool

	samples                   []mp4Sample
	version                   uint8
	hasCompositionTimeOffsets bool

	duration            uint64
	dataSize            uint64
	payloadOffset       int64
	outputOffset        int64
	startsWithSync      bool
	hasInferredDuration bool
}

type canonicalTfhd [canonicalTfhdSize]byte

type mp4Fragment struct {
	trackID                uint32
	timescale              uint32
	decodeTime             uint64
	duration               uint64
	startsWithSync         bool
	sampleDescriptionIndex uint32
	tfhd                   canonicalTfhd
	runs                   []mp4Run
	payloadStart           int
	payloadLen             int
	hasInferredDuration    bool
}

func (f *mp4Fragment) endTime() uint64 {
	return f.decodeTime + f.duration
}

type mp4BoxReader struct {
	onInit    func([]byte)
	onSegment func(Segment)

	segmentSeconds float64
	segmentDiff    int
	cueMode        bool

	init       bytes.Buffer
	sourceMoof bytes.Buffer
	sourceStyp bytes.Buffer
	deferred   linearGrowthBuffer

	video []mp4Fragment
	audio []mp4Fragment

	sourcePayload   linearGrowthBuffer
	segmentHeader   bytes.Buffer
	segmentPayloads [][]byte
	prefix          bytes.Buffer
	prefixActive    bool

	pendingStorage mp4Fragment
	pending        *mp4Fragment
	styp           []byte

	boxHeader         [16]byte
	boxHeaderLength   int
	boxHeaderRequired int

	currentBoxType      uint32
	currentBoxRemaining uint64
	currentTarget       boxTarget

	initDone              bool
	moovCompleted         bool
	sourceMfraDone        bool
	sourceFinalMoovDone   bool
	sourcePayloadFromMoof int64
	currentPayloadStart   int

	videoTrack trackInfo
	audioTrack trackInfo

	videoSampleDurationHint          uint32
	videoSampleDurationHintTimescale uint32
	audioSampleDurationHint          uint32
	audioSampleDurationHintTimescale uint32

	tfdtOffsetSeconds      float64
	sequence               uint32
	lastVideoEndTime       uint64
	completedVideoSegments int

	hasTargetSegment     bool
	targetSegmentStartNS uint64
	targetSegmentEndNS   uint64
	targetToleranceNS    uint64
}

func Mp4BoxReader(onInit func([]byte), onSegment func(Segment), segmentSeconds float64, segmentDiff int, cueMode bool) *mp4BoxReader {
	reader := &mp4BoxReader{
		onInit:            onInit,
		onSegment:         onSegment,
		segmentSeconds:    segmentSeconds,
		segmentDiff:       max(segmentDiff, 0),
		cueMode:           cueMode,
		boxHeaderRequired: 8,
		sequence:          1,
	}
	if math.IsNaN(segmentSeconds) || math.IsInf(segmentSeconds, 0) || segmentSeconds <= 0 {
		reader.segmentSeconds = 6
	}
	reader.sourceMoof.Grow(16 * 1024)
	reader.sourceStyp.Grow(128)
	reader.deferred.Grow(64 * 1024)
	reader.segmentHeader.Grow(64 * 1024)
	return reader
}

func (r *mp4BoxReader) SeekReset(seconds float64) {
	r.initDone = false
	r.moovCompleted = false
	r.sourceMfraDone = false
	r.sourceFinalMoovDone = false
	r.videoTrack = trackInfo{}
	r.audioTrack = trackInfo{}
	r.sequence = 1
	r.styp = r.styp[:0]
	r.videoSampleDurationHint = 0
	r.videoSampleDurationHintTimescale = 0
	r.audioSampleDurationHint = 0
	r.audioSampleDurationHintTimescale = 0
	r.lastVideoEndTime = 0
	r.completedVideoSegments = 0
	r.hasTargetSegment = false
	r.targetSegmentStartNS = 0
	r.targetSegmentEndNS = 0
	r.targetToleranceNS = 0

	if !math.IsNaN(seconds) && !math.IsInf(seconds, 0) && seconds > 0 {
		r.tfdtOffsetSeconds = seconds
	} else {
		r.tfdtOffsetSeconds = 0
	}

	r.init.Reset()
	r.sourceMoof.Reset()
	r.sourceStyp.Reset()
	r.deferred.Reset()
	r.segmentHeader.Reset()
	r.segmentPayloads = r.segmentPayloads[:0]
	r.clearSource()
	r.clearFragments()
	r.resetPrefix()
	r.resetBoxState()
}

func (r *mp4BoxReader) SetTimelineOffsetNS(offsetNS uint64) {
	if offsetNS == math.MaxUint64 {
		r.tfdtOffsetSeconds = 0
		return
	}
	r.tfdtOffsetSeconds = float64(offsetNS) / 1_000_000_000
}

func (r *mp4BoxReader) SetTargetSegment(startNS uint64, endNS uint64, toleranceNS uint64) error {
	if !r.cueMode {
		return nil
	}
	if endNS <= startNS {
		return errors.New("invalid cue segment range")
	}
	r.targetSegmentStartNS = startNS
	r.targetSegmentEndNS = endNS
	r.targetToleranceNS = max(toleranceNS, 1)
	r.hasTargetSegment = true
	return nil
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

func (r *mp4BoxReader) TryBuildEndOfStreamRemainder() (bool, error) {
	if r.cueMode && !r.hasTargetSegment {
		return false, nil
	}

	videoCount := len(r.video)
	audioCount := len(r.audio)

	if videoCount == 0 && audioCount == 0 {
		return false, nil
	}

	if videoCount > 0 && !r.video[0].startsWithSync {
		return false, nil
	}

	finalVideoSegment := videoCount > 0
	if videoCount > 0 && audioCount > 0 {
		videoEnd := r.video[videoCount-1].endTime()
		selected, err := r.selectAudioCount(videoEnd)
		if err != nil {
			return false, err
		}
		if selected > 0 {
			audioCount = selected
		}
	}

	if err := r.buildSegment(videoCount, audioCount, true); err != nil {
		return false, err
	}
	if finalVideoSegment && len(r.video) == 0 && len(r.audio) > 0 {
		for i := range r.audio {
			r.audio[i] = mp4Fragment{}
		}
		r.audio = r.audio[:0]
	}
	return true, nil
}

func (r *mp4BoxReader) undecodableEOSRemainderError() error {
	start := 0.0
	if len(r.video) > 0 && r.video[0].timescale != 0 {
		start = float64(r.video[0].decodeTime) / float64(r.video[0].timescale)
	}
	return fmt.Errorf(
		"%w: leftover video starts with a non-sync sample at %.6fs",
		ErrUndecodableEOSRemainder,
		start,
	)
}

func (r *mp4BoxReader) EndOfStreamError() error {
	if r.boxHeaderLength != 0 {
		return fmt.Errorf(
			"%w: incomplete top-level box header: have=%d, need=%d",
			ErrTruncatedMP4Fragment,
			r.boxHeaderLength,
			r.boxHeaderRequired,
		)
	}

	if r.currentBoxRemaining != 0 {
		return fmt.Errorf(
			"%w: incomplete %s box: missing=%d body bytes",
			ErrTruncatedMP4Fragment,
			fourCC(r.currentBoxType),
			r.currentBoxRemaining,
		)
	}

	if r.pending != nil {
		return fmt.Errorf(
			"%w: moof for track_ID=%d has no complete mdat",
			ErrTruncatedMP4Fragment,
			r.pending.trackID,
		)
	}

	if r.sourcePayload.Len() != 0 {
		return fmt.Errorf(
			"%w: %d uncommitted mdat payload bytes",
			ErrTruncatedMP4Fragment,
			r.sourcePayload.Len(),
		)
	}

	if r.sourceMoof.Len() != 0 {
		return fmt.Errorf(
			"%w: %d uncommitted moof bytes",
			ErrTruncatedMP4Fragment,
			r.sourceMoof.Len(),
		)
	}

	if r.deferred.Len() != 0 {
		return fmt.Errorf(
			"%w: %d unprocessed deferred bytes",
			ErrTruncatedMP4Fragment,
			r.deferred.Len(),
		)
	}

	return nil
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
		if r.currentBoxType == boxMfra {
			return errors.New("mfra appeared before init was completed")
		}

		r.currentTarget = boxTargetInit
		r.writeCurrentBoxData(r.boxHeader[:headerSize])
		return nil
	}

	if r.sourceFinalMoovDone {
		return fmt.Errorf("unexpected top-level mp4 box after terminal moov: %s", fourCC(r.currentBoxType))
	}
	if r.sourceMfraDone && r.currentBoxType != boxMoov {
		return fmt.Errorf("only the rewritten terminal moov is allowed after mfra; got %s", fourCC(r.currentBoxType))
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

		r.currentPayloadStart = r.sourcePayload.Len()
		if r.currentBoxRemaining <= uint64(maxGStreamerSampleBytes) {
			r.sourcePayload.Grow(int(r.currentBoxRemaining))
		}
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

	case boxMfra:
		if r.pending != nil {
			return errors.New("mfra cannot appear between moof and mdat")
		}
		if r.sourceMfraDone {
			return errors.New("duplicate terminal mfra")
		}
		return nil

	case boxMoov:
		if !r.sourceMfraDone {
			return errors.New("unexpected moov after mp4 initialization")
		}
		if r.sourceFinalMoovDone {
			return errors.New("duplicate terminal moov")
		}
		if r.pending != nil {
			return errors.New("final moov cannot appear between moof and mdat")
		}
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
		if len(r.styp) == 0 && r.sourceStyp.Len() > 0 {
			r.styp = append(r.styp[:0], r.sourceStyp.Bytes()...)
		}
		r.sourceStyp.Reset()
		return false, nil

	case boxMfra:
		r.sourceMfraDone = true
		return false, nil

	case boxMoov:
		if r.initDone {
			if !r.sourceMfraDone || r.sourceFinalMoovDone {
				return false, errors.New("unexpected moov after mp4 initialization")
			}
			r.sourceFinalMoovDone = true
			return false, nil
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
	videoHint := uint32(0)
	if r.videoSampleDurationHintTimescale == r.videoTrack.timescale {
		videoHint = r.videoSampleDurationHint
	}
	audioHint := uint32(0)
	if r.audioSampleDurationHintTimescale == r.audioTrack.timescale {
		audioHint = r.audioSampleDurationHint
	}

	var fragment mp4Fragment
	if err := parseSourceMoof(r.sourceMoof.Bytes(), r.videoTrack, r.audioTrack, videoHint, audioHint, &fragment); err != nil {
		return fmt.Errorf("unable to parse source moof: %w", err)
	}
	if err := r.resolvePreviousInferredDuration(&fragment); err != nil {
		return err
	}

	if !fragment.hasInferredDuration {
		if duration := lastSampleDuration(&fragment); duration != 0 {
			if fragment.trackID == r.videoTrack.id {
				r.videoSampleDurationHint = duration
				r.videoSampleDurationHintTimescale = fragment.timescale
			} else if fragment.trackID == r.audioTrack.id {
				r.audioSampleDurationHint = duration
				r.audioSampleDurationHintTimescale = fragment.timescale
			}
		}
	}

	r.pendingStorage = fragment
	r.pending = &r.pendingStorage
	r.sourcePayloadFromMoof = int64(r.sourceMoof.Len())
	return nil
}

func (r *mp4BoxReader) resolvePreviousInferredDuration(current *mp4Fragment) error {
	if current == nil {
		return errors.New("current fragment is nil")
	}

	var fragments []mp4Fragment
	switch current.trackID {
	case r.videoTrack.id:
		fragments = r.video
	case r.audioTrack.id:
		fragments = r.audio
	default:
		return nil
	}
	if len(fragments) == 0 {
		return nil
	}

	previous := &fragments[len(fragments)-1]
	if !previous.hasInferredDuration {
		return nil
	}

	var inferredRun *mp4Run
	for i := len(previous.runs) - 1; i >= 0; i-- {
		if previous.runs[i].hasInferredDuration {
			inferredRun = &previous.runs[i]
			break
		}
	}
	if inferredRun == nil || len(inferredRun.samples) == 0 {
		return errors.New("inferred sample duration marker is missing")
	}

	sampleIndex := len(inferredRun.samples) - 1
	sample := inferredRun.samples[sampleIndex]
	oldDuration := uint64(sample.duration)
	if previous.duration < uint64(sample.duration) {
		return errors.New("invalid inferred sample duration")
	}
	sampleStart := previous.decodeTime + previous.duration - uint64(sample.duration)
	if current.decodeTime < sampleStart {
		return fmt.Errorf("inferred sample moves decode time backwards: track=%d previous_tfdt=%d next_tfdt=%d", current.trackID, previous.decodeTime, current.decodeTime)
	}
	exactDuration := current.decodeTime - sampleStart
	if exactDuration > math.MaxUint32 {
		return errors.New("inferred sample duration exceeds uint32")
	}

	sample.duration = uint32(exactDuration)
	inferredRun.samples[sampleIndex] = sample
	inferredRun.duration = inferredRun.duration - oldDuration + exactDuration
	previous.duration = previous.duration - oldDuration + exactDuration
	for i := range previous.runs {
		previous.runs[i].hasInferredDuration = false
	}
	previous.hasInferredDuration = false

	if exactDuration != 0 {
		if current.trackID == r.videoTrack.id {
			r.videoSampleDurationHint = uint32(exactDuration)
			r.videoSampleDurationHintTimescale = current.timescale
		} else {
			r.audioSampleDurationHint = uint32(exactDuration)
			r.audioSampleDurationHintTimescale = current.timescale
		}
	}
	return nil
}

func lastSampleDuration(fragment *mp4Fragment) uint32 {
	if fragment == nil {
		return 0
	}
	for runIndex := len(fragment.runs) - 1; runIndex >= 0; runIndex-- {
		for sampleIndex := len(fragment.runs[runIndex].samples) - 1; sampleIndex >= 0; sampleIndex-- {
			if duration := fragment.runs[runIndex].samples[sampleIndex].duration; duration != 0 {
				return duration
			}
		}
	}
	return 0
}

func (r *mp4BoxReader) completeMdat() error {
	if r.pending == nil {
		return errors.New("completed mdat has no source moof")
	}

	payloadLen := r.sourcePayload.Len() - r.currentPayloadStart
	if err := attachPayload(r.pending, r.currentPayloadStart, payloadLen, r.sourcePayloadFromMoof); err != nil {
		return err
	}

	switch {
	case r.pending.trackID == r.videoTrack.id:
		r.video = append(r.video, *r.pending)

	case r.audioTrack.id != 0 && r.pending.trackID == r.audioTrack.id:
		r.audio = append(r.audio, *r.pending)

	default:
		return fmt.Errorf("unsupported track_ID=%d", r.pending.trackID)
	}

	r.pending = nil
	r.pendingStorage = mp4Fragment{}
	r.sourcePayloadFromMoof = 0
	r.currentPayloadStart = 0
	r.sourceMoof.Reset()
	return nil
}

func attachPayload(fragment *mp4Fragment, payloadStart int, payloadLen int, payloadFromMoof int64) error {
	var expected int64

	if payloadStart < 0 || payloadLen < 0 {
		return errors.New("negative source mdat payload range")
	}

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

	if expected != int64(payloadLen) {
		return fmt.Errorf("source mdat size mismatch: trun=%d, mdat=%d", expected, payloadLen)
	}

	fragment.payloadStart = payloadStart
	fragment.payloadLen = payloadLen
	return nil
}

func (r *mp4BoxReader) tryBuildSegment() (bool, error) {
	videoCount, err := r.selectVideoCount()
	if err != nil || videoCount == 0 {
		return false, err
	}

	audioCount := 0
	if r.audioTrack.id != 0 {
		videoEnd := r.video[videoCount-1].endTime()
		audioCount, err = r.selectAudioCount(videoEnd)
		if err != nil || audioCount == 0 {
			return false, err
		}
	}

	if err := r.buildSegment(videoCount, audioCount, false); err != nil {
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
	if r.cueMode {
		return r.selectCueVideoCount()
	}

	target, err := toUnits(r.segmentSeconds, r.videoTrack.timescale)
	if err != nil {
		return 0, err
	}

	takeFirstSyncBoundary := false
	if r.completedVideoSegments > 0 && r.segmentDiff > 0 {
		expectedEnd := uint64(r.completedVideoSegments) * uint64(r.segmentSeconds) * uint64(r.videoTrack.timescale)
		diff := uint64(r.segmentSeconds+float64(r.segmentDiff)) * uint64(r.videoTrack.timescale)
		if r.lastVideoEndTime > expectedEnd && r.lastVideoEndTime-expectedEnd >= diff {
			takeFirstSyncBoundary = true
		}
	}

	var duration uint64
	selectedCount := 0
	for i := 0; i+1 < len(r.video); i++ {
		duration += r.video[i].duration
		if !takeFirstSyncBoundary {
			if duration >= target && r.video[i+1].startsWithSync {
				return i + 1, nil
			}
			continue
		}

		if duration <= target {
			if r.video[i+1].startsWithSync {
				selectedCount = i + 1
				if duration == target {
					return selectedCount, nil
				}
			}
			continue
		}
		if selectedCount > 0 {
			return selectedCount, nil
		}
		if r.video[i+1].startsWithSync {
			return i + 1, nil
		}
	}

	return 0, nil
}

func (r *mp4BoxReader) selectCueVideoCount() (int, error) {
	if !r.hasTargetSegment || len(r.video) < 2 {
		return 0, nil
	}
	firstPresentation, err := firstPresentationTime(r.video[0])
	if err != nil {
		return 0, err
	}
	targetDuration := r.targetSegmentEndNS - r.targetSegmentStartNS
	for i := 1; i < len(r.video); i++ {
		boundary := r.video[i]
		if !boundary.startsWithSync {
			continue
		}
		boundaryPresentation, err := firstPresentationTime(boundary)
		if err != nil {
			return 0, err
		}
		if boundaryPresentation <= firstPresentation {
			return 0, errors.New("cue presentation timeline is not increasing")
		}
		durationNS, err := timelineNanoseconds(uint64(boundaryPresentation-firstPresentation), r.videoTrack.timescale)
		if err != nil {
			return 0, err
		}
		if durationNS < targetDuration && targetDuration-durationNS > r.targetToleranceNS {
			continue
		}
		if durationNS > targetDuration && durationNS-targetDuration > r.targetToleranceNS {
			return 0, fmt.Errorf("cue sync boundary duration is %d, expected %d", durationNS, targetDuration)
		}
		return i, nil
	}
	return 0, nil
}

func firstPresentationTime(fragment mp4Fragment) (int64, error) {
	if len(fragment.runs) == 0 || len(fragment.runs[0].samples) == 0 {
		return 0, errors.New("video fragment has no first sample")
	}
	run := fragment.runs[0]
	sample := run.samples[0]
	compositionOffset := int64(sample.compositionTimeOffsetRaw)
	if run.version == 1 {
		compositionOffset = int64(int32(sample.compositionTimeOffsetRaw))
	}
	if fragment.decodeTime > math.MaxInt64 {
		return 0, errors.New("video decode time exceeds int64")
	}
	result := int64(fragment.decodeTime)
	if compositionOffset > 0 && result > math.MaxInt64-compositionOffset {
		return 0, errors.New("presentation time overflow")
	}
	if compositionOffset < 0 && result < math.MinInt64-compositionOffset {
		return 0, errors.New("presentation time underflow")
	}
	return result + compositionOffset, nil
}

func (r *mp4BoxReader) selectAudioCount(videoEnd uint64) (int, error) {
	if len(r.audio) == 0 {
		return 0, nil
	}

	for i := range r.audio {
		splitSamples, reached, err := r.findAudioSplitSampleCount(
			&r.audio[i],
			videoEnd,
		)
		if err != nil {
			return 0, err
		}
		if !reached {
			continue
		}

		totalSamples, err := fragmentSampleCount(r.audio[i])
		if err != nil {
			return 0, err
		}
		if splitSamples <= 0 || splitSamples > totalSamples {
			return 0, errors.New("invalid audio split sample count")
		}

		if splitSamples < totalSamples {
			left, right, err := splitFragmentAtSample(
				r.audio[i],
				splitSamples,
			)
			if err != nil {
				return 0, err
			}

			r.audio[i] = left
			r.audio = append(r.audio, mp4Fragment{})
			copy(r.audio[i+2:], r.audio[i+1:])
			r.audio[i+1] = right
		}

		return i + 1, nil
	}

	return 0, nil
}

func (r *mp4BoxReader) findAudioSplitSampleCount(
	fragment *mp4Fragment,
	videoEnd uint64,
) (int, bool, error) {
	if fragment == nil {
		return 0, false, errors.New("audio fragment is nil")
	}
	if fragment.timescale == 0 || r.videoTrack.timescale == 0 {
		return 0, false, errors.New("audio/video timescale is zero")
	}

	audioEnd := fragment.decodeTime
	sampleCount := 0

	for _, run := range fragment.runs {
		if len(run.samples) == 0 {
			return 0, false, errors.New("audio trun has no parsed samples")
		}

		for _, sample := range run.samples {
			if math.MaxUint64-audioEnd < uint64(sample.duration) {
				return 0, false, errors.New("audio sample timeline overflow")
			}

			audioEnd += uint64(sample.duration)
			sampleCount++

			if scaledGreaterOrEqual(
				audioEnd,
				r.videoTrack.timescale,
				videoEnd,
				fragment.timescale,
			) {
				return sampleCount, true, nil
			}
		}
	}

	if audioEnd != fragment.endTime() {
		return 0, false, errors.New("audio sample durations do not match fragment duration")
	}

	return sampleCount, false, nil
}

func fragmentSampleCount(fragment mp4Fragment) (int, error) {
	total := 0

	for _, run := range fragment.runs {
		if len(run.samples) == 0 {
			return 0, errors.New("fragment trun has no parsed samples")
		}
		total += len(run.samples)
	}

	if total == 0 {
		return 0, errors.New("fragment contains no samples")
	}
	return total, nil
}

func splitFragmentAtSample(
	fragment mp4Fragment,
	splitSamples int,
) (mp4Fragment, mp4Fragment, error) {
	var left mp4Fragment
	var right mp4Fragment

	totalSamples, err := fragmentSampleCount(fragment)
	if err != nil {
		return left, right, err
	}
	if splitSamples <= 0 || splitSamples >= totalSamples {
		return left, right, fmt.Errorf(
			"split sample must be inside fragment: split=%d total=%d",
			splitSamples,
			totalSamples,
		)
	}

	left = mp4Fragment{
		trackID:                fragment.trackID,
		timescale:              fragment.timescale,
		decodeTime:             fragment.decodeTime,
		sampleDescriptionIndex: fragment.sampleDescriptionIndex,
		tfhd:                   fragment.tfhd,
	}
	right = mp4Fragment{
		trackID:                fragment.trackID,
		timescale:              fragment.timescale,
		sampleDescriptionIndex: fragment.sampleDescriptionIndex,
		tfhd:                   fragment.tfhd,
	}
	sourceRuns := fragment.runs
	left.runs = sourceRuns[:0]
	right.runs = make([]mp4Run, 0, len(sourceRuns))

	remaining := splitSamples
	var leftBytes uint64
	var rightBytes uint64
	var leftDuration uint64
	var rightDuration uint64

	for runIndex := range sourceRuns {
		sourceRun := sourceRuns[runIndex]
		if len(sourceRun.samples) == 0 {
			return mp4Fragment{}, mp4Fragment{}, errors.New(
				"cannot split a run without parsed samples",
			)
		}

		take := 0
		if remaining > 0 {
			take = minInt(remaining, len(sourceRun.samples))
		}

		if take > 0 {
			leftRun, err := buildExplicitRun(
				sourceRun,
				sourceRun.samples[:take],
			)
			if err != nil {
				return mp4Fragment{}, mp4Fragment{}, err
			}
			if leftBytes > math.MaxInt64 ||
				leftRun.dataSize > uint64(math.MaxInt64)-leftBytes {
				return mp4Fragment{}, mp4Fragment{}, errors.New(
					"left audio payload offset overflow",
				)
			}

			leftRun.payloadOffset = int64(leftBytes)
			leftBytes += leftRun.dataSize
			leftDuration += leftRun.duration
			left.runs = append(left.runs, leftRun)
			remaining -= take
		}

		if take < len(sourceRun.samples) {
			rightRun, err := buildExplicitRun(
				sourceRun,
				sourceRun.samples[take:],
			)
			if err != nil {
				return mp4Fragment{}, mp4Fragment{}, err
			}
			if rightBytes > math.MaxInt64 ||
				rightRun.dataSize > uint64(math.MaxInt64)-rightBytes {
				return mp4Fragment{}, mp4Fragment{}, errors.New(
					"right audio payload offset overflow",
				)
			}

			rightRun.payloadOffset = int64(rightBytes)
			rightBytes += rightRun.dataSize
			rightDuration += rightRun.duration
			right.runs = append(right.runs, rightRun)
		}
	}

	if remaining != 0 || len(left.runs) == 0 || len(right.runs) == 0 {
		return mp4Fragment{}, mp4Fragment{}, errors.New(
			"failed to partition fragment runs",
		)
	}
	if leftDuration+rightDuration != fragment.duration {
		return mp4Fragment{}, mp4Fragment{}, errors.New(
			"split fragment duration mismatch",
		)
	}
	if leftBytes+rightBytes != uint64(fragment.payloadLen) {
		return mp4Fragment{}, mp4Fragment{}, fmt.Errorf(
			"split payload mismatch: left=%d right=%d source=%d",
			leftBytes,
			rightBytes,
			fragment.payloadLen,
		)
	}
	if leftBytes > uint64(fragment.payloadLen) {
		return mp4Fragment{}, mp4Fragment{}, errors.New(
			"left payload exceeds source payload",
		)
	}

	splitByte := int(leftBytes)

	left.duration = leftDuration
	left.startsWithSync = left.runs[0].startsWithSync
	left.payloadStart = fragment.payloadStart
	left.payloadLen = splitByte

	if math.MaxUint64-fragment.decodeTime < leftDuration {
		return mp4Fragment{}, mp4Fragment{}, errors.New(
			"right fragment decode time overflow",
		)
	}

	right.decodeTime = fragment.decodeTime + leftDuration
	right.duration = rightDuration
	right.startsWithSync = right.runs[0].startsWithSync
	right.payloadStart = fragment.payloadStart + splitByte
	right.payloadLen = fragment.payloadLen - splitByte

	return left, right, nil
}

func (r *mp4BoxReader) buildSegment(videoCount int, audioCount int, allowSingleTrack bool) error {
	if videoCount < 0 || videoCount > len(r.video) {
		return errors.New("invalid video fragment count")
	}
	if audioCount < 0 || audioCount > len(r.audio) {
		return errors.New("invalid audio fragment count")
	}

	hasVideo := videoCount > 0
	hasAudio := audioCount > 0
	sourceHasAudio := r.audioTrack.id != 0

	if !hasVideo && !hasAudio {
		return errors.New("segment contains no fragments")
	}

	if hasAudio && !sourceHasAudio {
		return errors.New("video-only source unexpectedly contains audio fragments")
	}

	if !allowSingleTrack {
		if !hasVideo {
			return errors.New("regular segment must contain video")
		}
		if sourceHasAudio && !hasAudio {
			return errors.New("regular segment must contain audio for an audio/video source")
		}
	}

	if hasVideo {
		if err := validateTrack(r.video, videoCount); err != nil {
			return err
		}
	}
	if hasAudio {
		if err := validateTrack(r.audio, audioCount); err != nil {
			return err
		}
	}

	var payloadLength int64
	if hasVideo {
		assignOffsets(r.video, videoCount, &payloadLength)
	}
	if hasAudio {
		assignOffsets(r.audio, audioCount, &payloadLength)
	}

	var videoTrafSize int64
	if hasVideo {
		videoTrafSize = getTrafSize(r.video, videoCount)
	}
	var audioTrafSize int64
	if hasAudio {
		audioTrafSize = getTrafSize(r.audio, audioCount)
	}

	moofSize64 := int64(8+16) + videoTrafSize + audioTrafSize
	if moofSize64 > math.MaxUint32 {
		return errors.New("combined moof is too large")
	}

	moofSize := uint32(moofSize64)
	mdatHeaderSize := 8
	if uint64(payloadLength)+8 > math.MaxUint32 {
		mdatHeaderSize = 16
	}

	header := &r.segmentHeader
	header.Reset()
	header.Grow(int(moofSize64) + mdatHeaderSize + len(r.styp) + r.prefix.Len())

	if r.styp != nil {
		_, _ = header.Write(r.styp)
	}
	if r.prefixActive && r.prefix.Len() > 0 {
		_, _ = header.Write(r.prefix.Bytes())
	}

	writeHeader(header, moofSize, boxMoof)
	writeMfhd(header, r.sequence)
	r.sequence++
	if hasVideo {
		if err := r.writeTraf(header, r.video, videoCount, moofSize, mdatHeaderSize); err != nil {
			return err
		}
	}
	if hasAudio {
		if err := r.writeTraf(header, r.audio, audioCount, moofSize, mdatHeaderSize); err != nil {
			return err
		}
	}
	writeMdatHeader(header, uint64(payloadLength), mdatHeaderSize)

	var first mp4Fragment
	var last mp4Fragment
	if hasVideo {
		first = r.video[0]
		last = r.video[videoCount-1]
	} else {
		first = r.audio[0]
		last = r.audio[audioCount-1]
	}
	if first.timescale == 0 {
		return errors.New("segment first track has zero timescale")
	}
	if last.timescale == 0 {
		return errors.New("segment last track has zero timescale")
	}

	startNS, err := timelineNanoseconds(first.decodeTime, first.timescale)
	if err != nil {
		return err
	}
	endNS, err := timelineNanoseconds(last.endTime(), last.timescale)
	if err != nil {
		return err
	}
	offsetNS := uint64(0)
	if r.tfdtOffsetSeconds > 0 {
		offsetNS = uint64(math.Round(r.tfdtOffsetSeconds * 1_000_000_000))
	}
	if math.MaxUint64-startNS < offsetNS || math.MaxUint64-endNS < offsetNS {
		return errors.New("segment timeline offset overflow")
	}
	startNS += offsetNS
	endNS += offsetNS
	payloads, err := collectPayloadRangesInto(
		r.segmentPayloads[:0],
		r.sourcePayload.Bytes(),
		r.video,
		videoCount,
		r.audio,
		audioCount,
	)
	if err != nil {
		return err
	}
	r.segmentPayloads = payloads

	r.onSegment(Segment{
		Header:       header.Bytes(),
		Payloads:     r.segmentPayloads,
		StartNS:      startNS,
		EndNS:        endNS,
		StartSeconds: float64(startNS) / 1_000_000_000,
		EndSeconds:   float64(endNS) / 1_000_000_000,
	})
	if r.cueMode {
		r.hasTargetSegment = false
	}

	if hasVideo {
		r.lastVideoEndTime = last.endTime()
		r.completedVideoSegments++
		r.video = removeFragments(r.video, videoCount)
	}
	if hasAudio {
		r.audio = removeFragments(r.audio, audioCount)
	}
	r.resetPrefix()
	return nil
}

func validateTrack(fragments []mp4Fragment, count int) error {
	if count <= 0 || count > len(fragments) {
		return errors.New("invalid fragment count")
	}

	first := fragments[0]
	expected := first.endTime()

	for i := 1; i < count; i++ {
		current := fragments[i]
		if current.trackID != first.trackID ||
			current.timescale != first.timescale ||
			current.decodeTime != expected ||
			current.sampleDescriptionIndex != first.sampleDescriptionIndex ||
			current.tfhd != first.tfhd {
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

		*outputOffset += int64(fragment.payloadLen)
	}
}

func getTrafSize(fragments []mp4Fragment, count int) int64 {
	size := int64(8 + len(fragments[0].tfhd) + 20)
	for i := 0; i < count; i++ {
		for _, run := range fragments[i].runs {
			size += canonicalTrunSize(run)
		}
	}
	return size
}

func (r *mp4BoxReader) writeTraf(output *bytes.Buffer, fragments []mp4Fragment, count int, moofSize uint32, mdatHeaderSize int) error {
	size := getTrafSize(fragments, count)
	if size > math.MaxUint32 {
		return errors.New("combined traf is too large")
	}

	first := fragments[0]
	writeHeader(output, uint32(size), boxTraf)
	output.Write(first.tfhd[:])
	decodeTime, err := addTfdtOffset(first.decodeTime, first.timescale, r.tfdtOffsetSeconds)
	if err != nil {
		return err
	}
	writeTfdt(output, decodeTime)

	for i := 0; i < count; i++ {
		for j := range fragments[i].runs {
			run := &fragments[i].runs[j]
			dataOffset := int64(moofSize) + int64(mdatHeaderSize) + run.outputOffset
			if dataOffset < math.MinInt32 || dataOffset > math.MaxInt32 {
				return errors.New("trun.data_offset exceeds int32")
			}
			if err := writeCanonicalTrun(output, run, int32(dataOffset)); err != nil {
				return err
			}
		}
	}
	return nil
}

func canonicalTrunSize(run mp4Run) int64 {
	entrySize := int64(12)
	if run.hasCompositionTimeOffsets {
		entrySize = 16
	}
	return 20 + int64(len(run.samples))*entrySize
}

func writeCanonicalTrun(output *bytes.Buffer, run *mp4Run, dataOffset int32) error {
	if run == nil || len(run.samples) == 0 {
		return errors.New("cannot write trun without samples")
	}
	if run.version > 1 {
		return fmt.Errorf("unsupported trun version=%d", run.version)
	}

	size := canonicalTrunSize(*run)
	if size > math.MaxUint32 {
		return errors.New("canonical trun is too large")
	}

	flags := uint32(
		trunDataOffsetPresent |
			trunSampleDurationPresent |
			trunSampleSizePresent |
			trunSampleFlagsPresent,
	)
	if run.hasCompositionTimeOffsets {
		flags |= trunCompositionTimeOffsetPresent
	}

	var header [20]byte
	binary.BigEndian.PutUint32(header[0:4], uint32(size))
	binary.BigEndian.PutUint32(header[4:8], boxTrun)
	binary.BigEndian.PutUint32(header[8:12], uint32(run.version)<<24|flags)
	binary.BigEndian.PutUint32(header[12:16], uint32(len(run.samples)))
	binary.BigEndian.PutUint32(header[16:20], uint32(dataOffset))
	_, _ = output.Write(header[:])

	entrySize := 12
	if run.hasCompositionTimeOffsets {
		entrySize = 16
	}
	var entry [16]byte
	for _, sample := range run.samples {
		binary.BigEndian.PutUint32(entry[0:4], sample.duration)
		binary.BigEndian.PutUint32(entry[4:8], sample.size)
		binary.BigEndian.PutUint32(entry[8:12], sample.flags)
		if run.hasCompositionTimeOffsets {
			binary.BigEndian.PutUint32(entry[12:16], sample.compositionTimeOffsetRaw)
		}
		_, _ = output.Write(entry[:entrySize])
	}
	return nil
}

func parseSourceMoof(
	moof []byte,
	videoTrack trackInfo,
	audioTrack trackInfo,
	videoDurationHint uint32,
	audioDurationHint uint32,
	dst *mp4Fragment,
) error {
	if dst == nil {
		return errors.New("nil moof destination")
	}
	*dst = mp4Fragment{}

	rootPosition := 0
	rootType, moofHeader, moofBox, ok := tryReadBox(moof, &rootPosition)
	if !ok || rootType != boxMoof || rootPosition != len(moof) {
		return errors.New("buffer does not contain exactly one moof")
	}

	position := moofHeader
	trafCount := 0

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
			return errors.New("source moof must contain one traf")
		}

		if err := parseTraf(box, headerSize, videoTrack, audioTrack, videoDurationHint, audioDurationHint, dst); err != nil {
			return err
		}
	}

	if trafCount == 0 {
		return errors.New("traf was not found")
	}
	return nil
}

func parseTraf(
	traf []byte,
	trafHeader int,
	videoTrack trackInfo,
	audioTrack trackInfo,
	videoDurationHint uint32,
	audioDurationHint uint32,
	dst *mp4Fragment,
) error {
	if dst == nil {
		return errors.New("nil traf destination")
	}
	*dst = mp4Fragment{}

	var parsedTfhd tfhdInfo
	hasTfhd := false
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
			if hasTfhd {
				return errors.New("duplicate tfhd")
			}
			value, err := parseTfhd(box, headerSize)
			if err != nil {
				return err
			}
			parsedTfhd = value
			hasTfhd = true

		case boxTfdt:
			if hasTfdt {
				return errors.New("duplicate tfdt")
			}
			value, ok := readTfdt(box, headerSize)
			if !ok {
				return errors.New("invalid tfdt")
			}
			decodeTime = value
			hasTfdt = true

		case boxTrun:
			// Parsed after tfhd/trex defaults are known.

		default:
			return fmt.Errorf("unsupported box %s inside traf", fourCC(boxType))
		}
	}

	if !hasTfhd || !hasTfdt {
		return errors.New("tfhd/tfdt was not found")
	}

	var timescale uint32
	var trex trexInfo
	var durationHint uint32
	switch {
	case parsedTfhd.trackID == videoTrack.id:
		timescale = videoTrack.timescale
		trex = videoTrack.trex
		durationHint = videoDurationHint

	case audioTrack.id != 0 && parsedTfhd.trackID == audioTrack.id:
		timescale = audioTrack.timescale
		trex = audioTrack.trex
		durationHint = audioDurationHint

	default:
		return fmt.Errorf("unsupported track_ID=%d", parsedTfhd.trackID)
	}

	if timescale == 0 {
		return fmt.Errorf("timescale is zero for track_ID=%d", parsedTfhd.trackID)
	}

	sampleDescriptionIndex := parsedTfhd.sampleDescriptionIndex
	if !parsedTfhd.hasSampleDescriptionIndex {
		sampleDescriptionIndex = trex.descriptionIndex
	}
	if sampleDescriptionIndex == 0 {
		return fmt.Errorf("sample description index is absent for track_ID=%d", parsedTfhd.trackID)
	}

	defaultDuration := parsedTfhd.defaultDuration
	defaultDurationIsHint := false
	if defaultDuration == 0 {
		defaultDuration = trex.duration
	}
	if defaultDuration == 0 {
		defaultDuration = durationHint
		defaultDurationIsHint = defaultDuration != 0
	}

	defaultSize := parsedTfhd.defaultSize
	if defaultSize == 0 {
		defaultSize = trex.size
	}

	defaultFlags := parsedTfhd.defaultFlags
	if !parsedTfhd.hasDefaultFlags {
		defaultFlags = trex.flags
	}

	*dst = mp4Fragment{
		trackID:                parsedTfhd.trackID,
		timescale:              timescale,
		decodeTime:             decodeTime,
		sampleDescriptionIndex: sampleDescriptionIndex,
		tfhd:                   buildCanonicalTfhd(parsedTfhd.trackID, sampleDescriptionIndex),
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

		run, err := normalizeTrun(box, headerSize, defaultDuration, defaultSize, defaultFlags, defaultDurationIsHint)
		if err != nil {
			return err
		}
		if math.MaxUint64-duration < run.duration {
			return errors.New("fragment duration overflow")
		}
		duration += run.duration
		dst.hasInferredDuration = dst.hasInferredDuration || run.hasInferredDuration
		dst.runs = append(dst.runs, run)
	}

	if len(dst.runs) == 0 || duration == 0 {
		return errors.New("trun/duration was not found")
	}

	dst.duration = duration
	dst.startsWithSync = dst.runs[0].startsWithSync
	return nil
}

func parseTfhd(box []byte, headerSize int) (tfhdInfo, error) {
	var info tfhdInfo

	if len(box) < headerSize+8 {
		return info, errors.New("tfhd is too small")
	}

	versionFlags := binary.BigEndian.Uint32(box[headerSize : headerSize+4])
	info.version = uint8(versionFlags >> 24)
	flags := versionFlags & 0x00ffffff
	if info.version != 0 {
		return info, fmt.Errorf("unsupported tfhd version=%d", info.version)
	}

	const knownFlags = tfhdBaseDataOffsetPresent |
		tfhdSampleDescriptionIndexPresent |
		tfhdDefaultSampleDurationPresent |
		tfhdDefaultSampleSizePresent |
		tfhdDefaultSampleFlagsPresent |
		tfhdDurationIsEmpty |
		tfhdDefaultBaseIsMoof

	if unknown := flags &^ knownFlags; unknown != 0 {
		return info, fmt.Errorf("unsupported tfhd flags=0x%06x", unknown)
	}
	if flags&tfhdBaseDataOffsetPresent != 0 {
		return info, errors.New("tfhd.base-data-offset-present is not supported")
	}
	if flags&tfhdDurationIsEmpty != 0 {
		return info, errors.New("tfhd.duration-is-empty is not supported")
	}

	info.trackID = binary.BigEndian.Uint32(box[headerSize+4 : headerSize+8])
	if info.trackID == 0 {
		return info, errors.New("tfhd track_ID is zero")
	}

	cursor := headerSize + 8

	if flags&tfhdSampleDescriptionIndexPresent != 0 {
		value, ok := readUint32(box, &cursor)
		if !ok || value == 0 {
			return info, errors.New("invalid tfhd sample_description_index")
		}
		info.sampleDescriptionIndex = value
		info.hasSampleDescriptionIndex = true
	}

	if flags&tfhdDefaultSampleDurationPresent != 0 {
		value, ok := readUint32(box, &cursor)
		if !ok {
			return info, errors.New("invalid tfhd default_sample_duration")
		}
		info.defaultDuration = value
	}

	if flags&tfhdDefaultSampleSizePresent != 0 {
		value, ok := readUint32(box, &cursor)
		if !ok {
			return info, errors.New("invalid tfhd default_sample_size")
		}
		info.defaultSize = value
	}

	if flags&tfhdDefaultSampleFlagsPresent != 0 {
		value, ok := readUint32(box, &cursor)
		if !ok {
			return info, errors.New("invalid tfhd default_sample_flags")
		}
		info.defaultFlags = value
		info.hasDefaultFlags = true
	}

	if cursor != len(box) {
		return info, errors.New("invalid tfhd body")
	}

	return info, nil
}

func buildCanonicalTfhd(trackID uint32, sampleDescriptionIndex uint32) canonicalTfhd {
	var normalized canonicalTfhd
	binary.BigEndian.PutUint32(normalized[0:4], canonicalTfhdSize)
	binary.BigEndian.PutUint32(normalized[4:8], boxTfhd)
	binary.BigEndian.PutUint32(normalized[8:12], tfhdSampleDescriptionIndexPresent|tfhdDefaultBaseIsMoof)
	binary.BigEndian.PutUint32(normalized[12:16], trackID)
	binary.BigEndian.PutUint32(normalized[16:20], sampleDescriptionIndex)
	return normalized
}

func normalizeTrun(
	box []byte,
	headerSize int,
	defaultDuration uint32,
	defaultSize uint32,
	defaultFlags uint32,
	defaultDurationIsHint bool,
) (mp4Run, error) {
	var run mp4Run

	if len(box) < headerSize+8 {
		return run, errors.New("trun is too small")
	}

	versionFlags := binary.BigEndian.Uint32(box[headerSize : headerSize+4])
	version := uint8(versionFlags >> 24)
	flags := versionFlags & 0x00ffffff
	if version > 1 {
		return run, fmt.Errorf("unsupported trun version=%d", version)
	}

	const knownFlags = trunDataOffsetPresent |
		trunFirstSampleFlagsPresent |
		trunSampleDurationPresent |
		trunSampleSizePresent |
		trunSampleFlagsPresent |
		trunCompositionTimeOffsetPresent

	if unknown := flags &^ knownFlags; unknown != 0 {
		return run, fmt.Errorf("unsupported trun flags=0x%06x", unknown)
	}

	sampleCount := binary.BigEndian.Uint32(box[headerSize+4 : headerSize+8])

	if sampleCount == 0 {
		return run, errors.New("trun sample_count is zero")
	}

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
	hasSampleFlags := flags&trunSampleFlagsPresent != 0
	if hasFirstSampleFlags && hasSampleFlags {
		return run, errors.New("trun first_sample_flags and sample_flags are both present")
	}

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
	hasCompositionOffsets := flags&trunCompositionTimeOffsetPresent != 0

	if !hasDuration && defaultDuration == 0 {
		return run, errors.New("sample duration is absent")
	}
	if !hasSize && defaultSize == 0 {
		return run, errors.New("sample size is absent")
	}
	if sampleCount > maxTrunSamples {
		return run, fmt.Errorf("trun sample_count exceeds safety limit: %d", sampleCount)
	}

	fieldsPerSample := uint64(0)
	for _, present := range []bool{hasDuration, hasSize, hasSampleFlags, hasCompositionOffsets} {
		if present {
			fieldsPerSample++
		}
	}
	requiredLength := uint64(cursor) + uint64(sampleCount)*fieldsPerSample*4
	if requiredLength != uint64(len(box)) {
		return run, errors.New("invalid trun body")
	}

	samples := make([]mp4Sample, int(sampleCount))

	var duration uint64
	var dataSize uint64
	var firstSampleEffectiveFlags uint32

	for i := 0; i < int(sampleCount); i++ {
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

		effectiveFlags := defaultFlags
		if hasSampleFlags {
			value, ok := readUint32(box, &cursor)
			if !ok {
				return run, errors.New("invalid trun sample_flags")
			}
			effectiveFlags = value
		} else if i == 0 && hasFirstSampleFlags {
			effectiveFlags = firstSampleFlags
		}

		var compositionTimeOffsetRaw uint32
		if hasCompositionOffsets {
			value, ok := readUint32(box, &cursor)
			if !ok {
				return run, errors.New("invalid trun composition_time_offset")
			}
			compositionTimeOffsetRaw = value
		}

		if math.MaxUint64-duration < uint64(sampleDuration) {
			return run, errors.New("trun duration overflow")
		}
		if math.MaxUint64-dataSize < uint64(sampleSize) {
			return run, errors.New("trun payload size overflow")
		}

		duration += uint64(sampleDuration)
		dataSize += uint64(sampleSize)

		sample := mp4Sample{
			duration:                 sampleDuration,
			size:                     sampleSize,
			flags:                    effectiveFlags,
			compositionTimeOffsetRaw: compositionTimeOffsetRaw,
		}

		samples[i] = sample
		if i == 0 {
			firstSampleEffectiveFlags = effectiveFlags
		}
	}

	if cursor != len(box) || duration == 0 || dataSize == 0 {
		return run, errors.New("invalid trun body")
	}

	run.samples = samples
	run.version = version
	run.hasCompositionTimeOffsets = hasCompositionOffsets
	run.duration = duration
	run.dataSize = dataSize
	run.startsWithSync = isSyncSample(firstSampleEffectiveFlags)
	run.hasInferredDuration = !hasDuration && defaultDurationIsHint
	return run, nil
}

func buildExplicitRun(template mp4Run, samples []mp4Sample) (mp4Run, error) {
	var run mp4Run

	if len(samples) == 0 {
		return run, errors.New("cannot build trun without samples")
	}
	var duration uint64
	var dataSize uint64

	for _, sample := range samples {
		if math.MaxUint64-duration < uint64(sample.duration) {
			return run, errors.New("derived trun duration overflow")
		}
		if math.MaxUint64-dataSize < uint64(sample.size) {
			return run, errors.New("derived trun payload size overflow")
		}

		duration += uint64(sample.duration)
		dataSize += uint64(sample.size)
	}

	if duration == 0 || dataSize == 0 {
		return run, errors.New("invalid derived trun")
	}

	run.samples = samples
	run.duration = duration
	run.dataSize = dataSize
	run.startsWithSync = isSyncSample(samples[0].flags)
	run.version = template.version
	run.hasCompositionTimeOffsets = template.hasCompositionTimeOffsets
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

	video := trackInfo{
		id:        videoID,
		timescale: videoTimescale,
		trex:      findTrex(trexEntries, videoID),
	}

	var audio trackInfo
	if audioID != 0 {
		if audioTimescale == 0 {
			return trackInfo{}, trackInfo{}, errors.New("audio timescale is zero")
		}
		audio = trackInfo{
			id:        audioID,
			timescale: audioTimescale,
			trex:      findTrex(trexEntries, audioID),
		}
	}

	return video, audio, nil
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
	descriptionIndex := binary.BigEndian.Uint32(box[header+8 : header+12])

	if trackID == 0 || descriptionIndex == 0 {
		return trexEntry{}, false
	}

	return trexEntry{
		trackID: trackID,
		value: trexInfo{
			descriptionIndex: descriptionIndex,
			duration:         binary.BigEndian.Uint32(box[header+12 : header+16]),
			size:             binary.BigEndian.Uint32(box[header+16 : header+20]),
			flags:            binary.BigEndian.Uint32(box[header+20 : header+24]),
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

func toUnits(seconds float64, timescale uint32) (uint64, error) {
	value := seconds * float64(timescale)
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > float64(math.MaxUint64) {
		return 0, errors.New("invalid timeline value")
	}
	return uint64(math.Ceil(value)), nil
}

func timelineNanoseconds(value uint64, timescale uint32) (uint64, error) {
	if timescale == 0 {
		return 0, errors.New("timeline timescale is zero")
	}
	seconds := value / uint64(timescale)
	remainder := value % uint64(timescale)
	if seconds > math.MaxUint64/1_000_000_000 {
		return 0, errors.New("timeline nanoseconds overflow")
	}
	nanoseconds := seconds * 1_000_000_000
	fraction := remainder * 1_000_000_000 / uint64(timescale)
	if math.MaxUint64-nanoseconds < fraction {
		return 0, errors.New("timeline nanoseconds overflow")
	}
	return nanoseconds + fraction, nil
}

func addTfdtOffset(value uint64, timescale uint32, seconds float64) (uint64, error) {
	if seconds <= 0 {
		return value, nil
	}

	units := seconds * float64(timescale)
	if math.IsNaN(units) || math.IsInf(units, 0) || units < 0 || units > float64(math.MaxUint64) {
		return 0, errors.New("invalid tfdt offset")
	}

	offset := uint64(math.Round(units))
	if math.MaxUint64-value < offset {
		return 0, errors.New("tfdt offset overflow")
	}
	return value + offset, nil
}

func writeTfdt(output *bytes.Buffer, decodeTime uint64) {
	var box [20]byte
	binary.BigEndian.PutUint32(box[0:4], 20)
	binary.BigEndian.PutUint32(box[4:8], boxTfdt)
	binary.BigEndian.PutUint32(box[8:12], 0x01000000)
	binary.BigEndian.PutUint64(box[12:20], decodeTime)
	output.Write(box[:])
}

func writeMfhd(output *bytes.Buffer, sequence uint32) {
	var box [16]byte
	binary.BigEndian.PutUint32(box[0:4], 16)
	binary.BigEndian.PutUint32(box[4:8], boxMfhd)
	binary.BigEndian.PutUint32(box[8:12], 0)
	binary.BigEndian.PutUint32(box[12:16], sequence)
	output.Write(box[:])
}

func writeHeader(output *bytes.Buffer, size uint32, boxType uint32) {
	var header [8]byte
	binary.BigEndian.PutUint32(header[0:4], size)
	binary.BigEndian.PutUint32(header[4:8], boxType)
	output.Write(header[:])
}

func writeMdatHeader(output *bytes.Buffer, payloadLength uint64, headerSize int) {
	if headerSize == 8 {
		writeHeader(output, uint32(payloadLength+8), boxMdat)
		return
	}

	var header [16]byte
	binary.BigEndian.PutUint32(header[0:4], 1)
	binary.BigEndian.PutUint32(header[4:8], boxMdat)
	binary.BigEndian.PutUint64(header[8:16], payloadLength+16)
	output.Write(header[:])
}

func collectPayloadRangesInto(
	dst [][]byte,
	storage []byte,
	video []mp4Fragment,
	videoCount int,
	audio []mp4Fragment,
	audioCount int,
) ([][]byte, error) {
	var err error
	dst, err = appendFragmentPayloadRanges(dst, storage, video, videoCount)
	if err != nil {
		return dst, err
	}
	return appendFragmentPayloadRanges(dst, storage, audio, audioCount)
}

func appendFragmentPayloadRanges(dst [][]byte, storage []byte, fragments []mp4Fragment, count int) ([][]byte, error) {
	for i := 0; i < count; i++ {
		fragment := fragments[i]
		if fragment.payloadLen == 0 {
			continue
		}
		if fragment.payloadStart < 0 || fragment.payloadLen < 0 {
			return dst, errors.New("negative fragment payload range")
		}
		end := fragment.payloadStart + fragment.payloadLen
		if end < fragment.payloadStart || end > len(storage) {
			return dst, errors.New("fragment payload range exceeds storage")
		}
		dst = append(dst, storage[fragment.payloadStart:end])
	}
	return dst, nil
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

func (r *mp4BoxReader) ReclaimPayloads() {
	if r.sourcePayload.Len() == 0 {
		return
	}

	minStart := r.minQueuedPayloadStart()
	if minStart < 0 {
		r.sourcePayload.Reset()
		r.currentPayloadStart = 0
		return
	}
	if minStart == 0 {
		return
	}

	data := r.sourcePayload.Bytes()
	copy(data, data[minStart:])
	r.sourcePayload.Truncate(len(data) - minStart)

	shiftFragmentPayloadStarts(r.video, minStart)
	shiftFragmentPayloadStarts(r.audio, minStart)
	if r.pending != nil && r.pending.payloadLen > 0 {
		r.pending.payloadStart -= minStart
	}

	if r.currentPayloadStart >= minStart {
		r.currentPayloadStart -= minStart
	} else {
		r.currentPayloadStart = 0
	}
}

func (r *mp4BoxReader) ReleaseSegment() {
	clear(r.segmentPayloads)
	r.segmentPayloads = r.segmentPayloads[:0]
	r.ReclaimPayloads()
}

func (r *mp4BoxReader) minQueuedPayloadStart() int {
	minStart := -1
	minStart = minFragmentPayloadStart(minStart, r.video)
	minStart = minFragmentPayloadStart(minStart, r.audio)
	if r.pending != nil && r.pending.payloadLen > 0 {
		minStart = minPayloadStart(minStart, r.pending.payloadStart)
	}
	return minStart
}

func minFragmentPayloadStart(current int, fragments []mp4Fragment) int {
	for i := range fragments {
		if fragments[i].payloadLen <= 0 {
			continue
		}
		current = minPayloadStart(current, fragments[i].payloadStart)
	}
	return current
}

func minPayloadStart(current int, value int) int {
	if current < 0 || value < current {
		return value
	}
	return current
}

func shiftFragmentPayloadStarts(fragments []mp4Fragment, delta int) {
	for i := range fragments {
		if fragments[i].payloadLen > 0 {
			fragments[i].payloadStart -= delta
		}
	}
}

func (r *mp4BoxReader) resetPrefix() {
	r.prefix.Reset()
	r.prefixActive = false
}

func (r *mp4BoxReader) clearSource() {
	r.pending = nil
	r.pendingStorage = mp4Fragment{}
	r.sourcePayload.Reset()
	r.sourcePayloadFromMoof = 0
	r.currentPayloadStart = 0
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
