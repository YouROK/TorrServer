//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"server/settings"

	"github.com/ebitengine/purego"
)

const (
	gstPadProbeDrop int32 = 0
	gstPadProbeOK   int32 = 1

	gstPadProbeTypeBuffer          int32 = 1 << 4
	gstPadProbeTypeEventDownstream int32 = 1 << 6
	gstPadProbeTypeEventUpstream   int32 = 1 << 7
	gstPadProbeTypeEventFlush      int32 = 1 << 8

	gstClockTimeNone uint64 = ^uint64(0)

	gstMiniObjectFlagsOffset = uintptr(16)
	gstMiniObjectSize        = uintptr(64)

	gstBufferPoolOffset      = gstMiniObjectSize
	gstBufferPtsOffset       = gstBufferPoolOffset + unsafe.Sizeof(uintptr(0))
	gstBufferDtsOffset       = gstBufferPtsOffset + 8
	gstBufferDurationOffset  = gstBufferDtsOffset + 8
	gstBufferOffsetOffset    = gstBufferDurationOffset + 8
	gstBufferOffsetEndOffset = gstBufferOffsetOffset + 8

	gstEventTypeOffset = gstMiniObjectSize

	gstBufferFlagDiscont   uint32 = 1 << 6
	gstBufferFlagCorrupted uint32 = 1 << 8
	gstBufferFlagHeader    uint32 = 1 << 10
	gstBufferFlagGap       uint32 = 1 << 11
	gstBufferFlagDroppable uint32 = 1 << 12
	gstBufferFlagDeltaUnit uint32 = 1 << 13

	gstProbeBranchVideoIn = "branch=video-in"
	gstProbeBranchAudioIn = "branch=audio-in"
	gstProbeBranchMuxOut  = "branch=mux-out"

	gstProbePhasePreroll  uint32 = 0
	gstProbePhasePostSeek uint32 = 1
)

var (
	nextPipelineGen atomic.Uint64
	nextProbeHandle atomic.Uint64

	gstProbeRegistryMu sync.RWMutex
	gstProbeRegistry   = make(map[uintptr]*gstPadProbeRef)

	gstProbeCallbackOnce    sync.Once
	gstProbeCallbackAddress uintptr

	probeLogOnce  sync.Once
	probeLogQueue = make(chan string, 1024)
)

type gstPipelineProbes struct {
	pipelineGen           uint64
	debug                 bool
	refs                  []*gstPadProbeRef
	released              atomic.Bool
	dropped               atomic.Uint64
	droppedLeadingBuffers atomic.Uint64
}

type gstPadProbeRef struct {
	pipelineGen               uint64
	branch                    string
	debug                     bool
	filterHEVCLeadingPictures bool
	pad                       uintptr
	id                        uintptr
	handle                    uintptr
	dropped                   *atomic.Uint64
	droppedLeadingBuffers     *atomic.Uint64

	bufferCount atomic.Uint64
	phase       atomic.Uint32
	released    atomic.Bool

	segmentMu  sync.RWMutex
	segment    gstSegmentSnapshot
	hasSegment bool

	filterMu                       sync.Mutex
	fragmentDurationNS             int64
	fragmentStartDecodeRunningTime int64
	fragmentStartSet               bool
	leadingBoundaryPTSRunningTime  int64
	leadingWindow                  bool
	fragmentBoundaryCount          uint64
}

type gstSegmentSnapshot struct {
	format      int32
	flags       uint32
	rate        float64
	appliedRate float64
	start       uint64
	stop        uint64
	time        uint64
	base        uint64
	offset      uint64
	duration    uint64
}

type gstBufferSnapshot struct {
	size      uintptr
	pts       uint64
	dts       uint64
	duration  uint64
	offset    uint64
	offsetEnd uint64
	flags     uint32
}

func newGSTPipelineProbes(r *gstRunner, pipeline uintptr, pipelineGen uint64, filterHEVCLeadingPictures bool, fragmentDuration time.Duration) *gstPipelineProbes {
	debug := settings.IsDebug()
	if !debug && !filterHEVCLeadingPictures {
		return nil
	}
	if debug {
		startProbeLogger()
	}

	probes := &gstPipelineProbes{pipelineGen: pipelineGen, debug: debug}
	if gstRuntime == nil || pipeline == 0 {
		return probes
	}

	mux := gstRuntime.binGetByName(pipeline, "mux")
	if mux == 0 {
		gstDebugf("pad probe mux lookup failed task=%s file=%s audio=%d pipelineGen=%d err=mp4mux element is not available", r.task.ID, r.task.FileID, r.audioIndex, pipelineGen)
		return probes
	}
	defer gstRuntime.objectUnref(mux)

	if debug || filterHEVCLeadingPictures {
		probes.add(r, mux, "video_0", gstProbeBranchVideoIn, filterHEVCLeadingPictures, fragmentDuration)
	}
	if debug {
		probes.add(r, mux, "audio_0", gstProbeBranchAudioIn, false, 0)
		probes.add(r, mux, "src", gstProbeBranchMuxOut, false, 0)
	}
	return probes
}

func (p *gstPipelineProbes) add(r *gstRunner, mux uintptr, padName string, branch string, filterHEVCLeadingPictures bool, fragmentDuration time.Duration) {
	pad := gstRuntime.elementGetStaticPad(mux, padName)
	if pad == 0 {
		gstDebugf("pad probe unavailable task=%s file=%s audio=%d pipelineGen=%d %s pad=%s", r.task.ID, r.task.FileID, r.audioIndex, p.pipelineGen, branch, padName)
		return
	}

	ref := &gstPadProbeRef{
		pipelineGen:               p.pipelineGen,
		branch:                    branch,
		debug:                     p.debug,
		filterHEVCLeadingPictures: filterHEVCLeadingPictures,
		pad:                       pad,
		handle:                    uintptr(nextProbeHandle.Add(1)),
		dropped:                   &p.dropped,
		droppedLeadingBuffers:     &p.droppedLeadingBuffers,
		fragmentDurationNS:        fragmentDuration.Nanoseconds(),
	}
	ref.phase.Store(gstProbePhasePreroll)

	gstProbeRegistryMu.Lock()
	gstProbeRegistry[ref.handle] = ref
	gstProbeRegistryMu.Unlock()

	mask := gstPadProbeTypeBuffer |
		gstPadProbeTypeEventDownstream |
		gstPadProbeTypeEventFlush
	ref.id = gstRuntime.padAddProbe(ref.pad, mask, gstPadProbeCallbackAddress(), ref.handle)
	if ref.id == 0 {
		ref.release()
		gstDebugf("pad probe install failed task=%s file=%s audio=%d pipelineGen=%d %s pad=%s", r.task.ID, r.task.FileID, r.audioIndex, p.pipelineGen, branch, padName)
		return
	}

	p.refs = append(p.refs, ref)
	gstDebugf("pad probe installed task=%s file=%s audio=%d pipelineGen=%d %s pad=%s probeID=%d", r.task.ID, r.task.FileID, r.audioIndex, p.pipelineGen, branch, padName, ref.id)
}

func (p *gstPipelineProbes) release() {
	if p == nil || !p.released.CompareAndSwap(false, true) {
		return
	}
	for _, ref := range p.refs {
		ref.release()
	}
	if p.debug {
		fragmentBoundaries := uint64(0)
		for _, ref := range p.refs {
			fragmentBoundaries += ref.fragmentBoundaryCountValue()
		}
		if droppedLeading := p.droppedLeadingBuffers.Load(); droppedLeading > 0 {
			gstDebugf("pad probe leading video buffers dropped pipelineGen=%d droppedLeadingBuffers=%d fragmentBoundaries=%d", p.pipelineGen, droppedLeading, fragmentBoundaries)
		} else if fragmentBoundaries > 0 {
			gstDebugf("pad probe HEVC fragment boundaries pipelineGen=%d fragmentBoundaries=%d droppedLeadingBuffers=0", p.pipelineGen, fragmentBoundaries)
		}
		if dropped := p.dropped.Load(); dropped > 0 {
			gstDebugf("pad probe diagnostic lines dropped pipelineGen=%d dropped=%d", p.pipelineGen, dropped)
		}
	}
	p.refs = nil
}

func (p *gstPipelineProbes) resetForSeek() {
	if p == nil || p.released.Load() {
		return
	}
	for _, ref := range p.refs {
		ref.bufferCount.Store(0)
		ref.phase.Store(gstProbePhasePostSeek)
		ref.segmentMu.Lock()
		ref.segment = gstSegmentSnapshot{}
		ref.hasSegment = false
		ref.segmentMu.Unlock()
		ref.resetFilterState()
	}
}

func (p *gstPadProbeRef) release() {
	if p == nil || !p.released.CompareAndSwap(false, true) {
		return
	}

	gstProbeRegistryMu.Lock()
	delete(gstProbeRegistry, p.handle)
	gstProbeRegistryMu.Unlock()

	if gstRuntime != nil {
		gstRuntime.padRemoveProbe(p.pad, p.id)
		gstRuntime.objectUnref(p.pad)
	}
	p.id = 0
	p.pad = 0
}

func gstPadProbeCallbackAddress() uintptr {
	gstProbeCallbackOnce.Do(func() {
		gstProbeCallbackAddress = purego.NewCallback(func(pad uintptr, info uintptr, userData uintptr) int32 {
			return gstPadProbeCallback(pad, info, userData)
		})
	})
	return gstProbeCallbackAddress
}

func gstPadProbeCallback(_ uintptr, info uintptr, userData uintptr) int32 {
	gstProbeRegistryMu.RLock()
	ref := gstProbeRegistry[userData]
	gstProbeRegistryMu.RUnlock()
	if ref == nil || ref.released.Load() || gstRuntime == nil {
		return gstPadProbeOK
	}

	probeType := readGstPadProbeInfoType(info)
	if probeType&gstPadProbeTypeBuffer != 0 {
		if buffer := gstRuntime.padProbeInfoBuffer(info); buffer != 0 {
			if ref.shouldDropLeadingBuffer(buffer) {
				return gstPadProbeDrop
			}
			ref.logBuffer(buffer)
		}
	}
	if probeType&gstPadProbeTypeEventDownstream != 0 {
		if event := gstRuntime.padProbeInfoEvent(info); event != 0 {
			ref.logEvent(event)
		}
	}
	return gstPadProbeOK
}

func readGstPadProbeInfoType(info uintptr) int32 {
	if info == 0 {
		return 0
	}
	return int32(readUint32AtAddress(info))
}

func (p *gstPadProbeRef) logEvent(event uintptr) {
	eventType := readGstEventType(event)
	name := gstRuntime.eventTypeName(eventType)
	if !isLoggedPadEvent(name) {
		return
	}

	label := eventLogLabel(name)
	if strings.EqualFold(name, "segment") {
		if segment, ok := gstRuntime.eventSegment(event); ok {
			p.segmentMu.Lock()
			p.segment = segment
			p.hasSegment = true
			p.segmentMu.Unlock()
			if !p.debug {
				return
			}
			p.probeLog(fmt.Sprintf("pad probe event pipelineGen=%d %s phase=%s event=%s format=%d flags=%d rate=%.6f appliedRate=%.6f start=%d stop=%d time=%d base=%d offset=%d duration=%d",
				p.pipelineGen, p.branch, p.phaseLabel(), label, segment.format, segment.flags, segment.rate, segment.appliedRate, segment.start, segment.stop, segment.time, segment.base, segment.offset, segment.duration))
			return
		}
	}

	if !p.debug {
		return
	}
	if strings.EqualFold(name, "caps") {
		p.probeLog(fmt.Sprintf("pad probe event pipelineGen=%d %s phase=%s event=%s caps=%q", p.pipelineGen, p.branch, p.phaseLabel(), label, gstRuntime.eventCapsString(event)))
		return
	}
	p.probeLog(fmt.Sprintf("pad probe event pipelineGen=%d %s phase=%s event=%s", p.pipelineGen, p.branch, p.phaseLabel(), label))
}

func (p *gstPadProbeRef) logBuffer(buffer uintptr) {
	if !p.debug {
		return
	}

	index := p.bufferCount.Add(1)
	if index > 30 {
		return
	}

	info := readGstBuffer(buffer)
	ptsRunningTime := p.runningTime(info.pts)
	dtsRunningTime := p.runningTime(info.dts)

	message := fmt.Sprintf("pad probe buffer pipelineGen=%d %s phase=%s index=%d size=%d PTS=%d DTS=%d duration=%d offset=%d offsetEnd=%d flags=%d DISCONT=%t DELTA_UNIT=%t HEADER=%t GAP=%t DROPPABLE=%t CORRUPTED=%t ptsValid=%t dtsValid=%t durationValid=%t ptsRunningTime=%s dtsRunningTime=%s",
		p.pipelineGen, p.branch, p.phaseLabel(), index, info.size, info.pts, info.dts, info.duration, info.offset, info.offsetEnd, info.flags,
		info.flags&gstBufferFlagDiscont != 0,
		info.flags&gstBufferFlagDeltaUnit != 0,
		info.flags&gstBufferFlagHeader != 0,
		info.flags&gstBufferFlagGap != 0,
		info.flags&gstBufferFlagDroppable != 0,
		info.flags&gstBufferFlagCorrupted != 0,
		gstClockTimeIsValid(info.pts),
		gstClockTimeIsValid(info.dts),
		gstClockTimeIsValid(info.duration),
		ptsRunningTime,
		dtsRunningTime,
	)

	if p.branch == gstProbeBranchMuxOut && index <= 12 {
		hexPrefix, fourCC := gstBufferPrefix(buffer)
		if hexPrefix != "" {
			message += " first16=" + hexPrefix
		}
		if fourCC != "" {
			message += " fourCC=" + fourCC
		}
	}
	p.probeLog(message)
}

func (p *gstPadProbeRef) shouldDropLeadingBuffer(buffer uintptr) bool {
	if !p.filterHEVCLeadingPictures || p.phase.Load() != gstProbePhasePostSeek {
		return false
	}

	info := readGstBuffer(buffer)
	if !gstClockTimeIsValid(info.pts) {
		return false
	}

	ptsRunningTime, ok := p.runningTimeValue(info.pts)
	if !ok {
		return false
	}

	dtsRunningTime, dtsOK := p.runningTimeValue(info.dts)
	decodeRunningTime := ptsRunningTime
	if dtsOK {
		decodeRunningTime = dtsRunningTime
	}

	if info.flags&gstBufferFlagDeltaUnit == 0 {
		p.updateHEVCFragmentBoundary(info, ptsRunningTime, dtsRunningTime, dtsOK, decodeRunningTime)
		return false
	}

	drop, boundaryIndex, boundaryPTSRunningTime := p.consumeHEVCLeadingDelta(ptsRunningTime)
	if !drop {
		return false
	}

	if p.droppedLeadingBuffers != nil {
		p.droppedLeadingBuffers.Add(1)
	}
	if p.debug {
		p.probeLog(fmt.Sprintf("pad probe drop leading video pipelineGen=%d phase=%s boundaryIndex=%d PTS=%d DTS=%d ptsRunningTime=%d boundaryPTSRunningTime=%d flags=%d",
			p.pipelineGen, p.phaseLabel(), boundaryIndex, info.pts, info.dts, ptsRunningTime, boundaryPTSRunningTime, info.flags))
	}
	return true
}

func (p *gstPadProbeRef) updateHEVCFragmentBoundary(info gstBufferSnapshot, ptsRunningTime int64, dtsRunningTime int64, dtsOK bool, decodeRunningTime int64) {
	reason := ""
	boundaryIndex := uint64(0)

	p.filterMu.Lock()
	if !p.fragmentStartSet {
		p.fragmentStartDecodeRunningTime = decodeRunningTime
		p.leadingBoundaryPTSRunningTime = ptsRunningTime
		p.leadingWindow = true
		p.fragmentStartSet = true
		p.fragmentBoundaryCount++
		boundaryIndex = p.fragmentBoundaryCount
		reason = "initial"
	} else if p.fragmentDurationNS > 0 && decodeRunningTime-p.fragmentStartDecodeRunningTime >= p.fragmentDurationNS {
		p.fragmentStartDecodeRunningTime = decodeRunningTime
		p.fragmentBoundaryCount++
		boundaryIndex = p.fragmentBoundaryCount
		reason = "fragment-duration"
	}
	p.filterMu.Unlock()

	if reason == "" || !p.debug {
		return
	}

	dtsText := "NONE"
	if dtsOK {
		dtsText = strconv.FormatInt(dtsRunningTime, 10)
	}
	p.probeLog(fmt.Sprintf("pad probe HEVC fragment boundary pipelineGen=%d phase=%s boundaryIndex=%d reason=%s PTS=%d DTS=%d ptsRunningTime=%d dtsRunningTime=%s fragmentDuration=%s",
		p.pipelineGen, p.phaseLabel(), boundaryIndex, reason, info.pts, info.dts, ptsRunningTime, dtsText, time.Duration(p.fragmentDurationNS)))
}

func (p *gstPadProbeRef) consumeHEVCLeadingDelta(ptsRunningTime int64) (bool, uint64, int64) {
	p.filterMu.Lock()
	defer p.filterMu.Unlock()

	if !p.leadingWindow {
		return false, p.fragmentBoundaryCount, p.leadingBoundaryPTSRunningTime
	}
	if ptsRunningTime < p.leadingBoundaryPTSRunningTime {
		return true, p.fragmentBoundaryCount, p.leadingBoundaryPTSRunningTime
	}
	p.leadingWindow = false
	return false, p.fragmentBoundaryCount, p.leadingBoundaryPTSRunningTime
}

func (p *gstPadProbeRef) resetFilterState() {
	p.filterMu.Lock()
	p.fragmentStartDecodeRunningTime = 0
	p.fragmentStartSet = false
	p.leadingBoundaryPTSRunningTime = 0
	p.leadingWindow = false
	p.fragmentBoundaryCount = 0
	p.filterMu.Unlock()
}

func (p *gstPadProbeRef) fragmentBoundaryCountValue() uint64 {
	p.filterMu.Lock()
	defer p.filterMu.Unlock()
	return p.fragmentBoundaryCount
}

func (p *gstPadProbeRef) runningTime(position uint64) string {
	runningTime, ok := p.runningTimeValue(position)
	if !ok {
		return "NONE"
	}
	return strconv.FormatInt(runningTime, 10)
}

func (p *gstPadProbeRef) runningTimeValue(position uint64) (int64, bool) {
	if !gstClockTimeIsValid(position) {
		return 0, false
	}

	p.segmentMu.RLock()
	segment := p.segment
	ok := p.hasSegment
	p.segmentMu.RUnlock()
	if !ok || segment.format != gstFormatTime || segment.rate == 0 {
		return 0, false
	}
	start := int64(segment.start)
	offset := int64(segment.offset)
	base := int64(segment.base)
	adjustedStart := start + offset
	delta := int64(position) - adjustedStart
	rate := math.Abs(segment.rate)
	if rate == 0 {
		return 0, false
	}
	return base + int64(float64(delta)/rate), true
}

func (p *gstPadProbeRef) phaseLabel() string {
	if p.phase.Load() == gstProbePhasePostSeek {
		return "post-seek"
	}
	return "preroll"
}

func (p *gstPadProbeRef) probeLog(message string) {
	select {
	case probeLogQueue <- message:
	default:
		if p.dropped != nil {
			p.dropped.Add(1)
		}
	}
}

func startProbeLogger() {
	probeLogOnce.Do(func() {
		go func() {
			for line := range probeLogQueue {
				gstDebugf("%s", line)
			}
		}()
	})
}

func readGstBuffer(buffer uintptr) gstBufferSnapshot {
	return gstBufferSnapshot{
		size:      gstRuntime.gstBufferGetSize(buffer),
		pts:       readUint64AtAddress(buffer + gstBufferPtsOffset),
		dts:       readUint64AtAddress(buffer + gstBufferDtsOffset),
		duration:  readUint64AtAddress(buffer + gstBufferDurationOffset),
		offset:    readUint64AtAddress(buffer + gstBufferOffsetOffset),
		offsetEnd: readUint64AtAddress(buffer + gstBufferOffsetEndOffset),
		flags:     readUint32AtAddress(buffer + gstMiniObjectFlagsOffset),
	}
}

func readGstEventType(event uintptr) int32 {
	return int32(readUint32AtAddress(event + gstEventTypeOffset))
}

func readGstSegment(segment uintptr) gstSegmentSnapshot {
	return gstSegmentSnapshot{
		flags:       readUint32AtAddress(segment),
		rate:        readFloat64AtAddress(segment + 8),
		appliedRate: readFloat64AtAddress(segment + 16),
		format:      int32(readUint32AtAddress(segment + 24)),
		base:        readUint64AtAddress(segment + 32),
		offset:      readUint64AtAddress(segment + 40),
		start:       readUint64AtAddress(segment + 48),
		stop:        readUint64AtAddress(segment + 56),
		time:        readUint64AtAddress(segment + 64),
		duration:    readUint64AtAddress(segment + 80),
	}
}

func gstBufferPrefix(buffer uintptr) (string, string) {
	if buffer == 0 || gstRuntime == nil {
		return "", ""
	}
	size := gstRuntime.gstBufferGetSize(buffer)
	if size == 0 {
		return "", ""
	}

	var mapInfo [128]byte
	if gstRuntime.gstBufferMap(buffer, unsafe.Pointer(&mapInfo[0]), gstMapRead) == 0 {
		return "", ""
	}
	defer gstRuntime.gstBufferUnmap(buffer, unsafe.Pointer(&mapInfo[0]))

	dataPtr, mapSize := gstMapInfoData(&mapInfo)
	if dataPtr == 0 || mapSize <= 0 {
		return "", ""
	}

	limit := mapSize
	if limit > 16 {
		limit = 16
	}
	data := unsafe.Slice((*byte)(unsafe.Pointer(dataPtr)), limit)
	return hexBytes(data), mp4FourCC(data)
}

func hexBytes(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, b := range data {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(fmt.Sprintf("%02x", b))
	}
	return sb.String()
}

func mp4FourCC(data []byte) string {
	if len(data) < 8 {
		return ""
	}
	fourCC := data[4:8]
	for _, b := range fourCC {
		if b < 0x20 || b > 0x7e {
			return ""
		}
	}
	return string(fourCC)
}

func isLoggedPadEvent(name string) bool {
	switch strings.ToLower(name) {
	case "stream-start", "caps", "segment", "flush-start", "flush-stop", "gap", "eos":
		return true
	default:
		return false
	}
}

func eventLogLabel(name string) string {
	return strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
}

func gstClockTimeIsValid(value uint64) bool {
	return value != gstClockTimeNone
}

func readUint32AtAddress(address uintptr) uint32 {
	return *(*uint32)(unsafe.Pointer(address))
}

func readUint64AtAddress(address uintptr) uint64 {
	return *(*uint64)(unsafe.Pointer(address))
}

func readFloat64AtAddress(address uintptr) float64 {
	return *(*float64)(unsafe.Pointer(address))
}
