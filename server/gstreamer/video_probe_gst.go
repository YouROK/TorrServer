//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	gstClockTimeNone = ^uint64(0)

	gstPadProbeTypeBuffer          uint32 = 1 << 4
	gstPadProbeTypeEventDownstream uint32 = 1 << 6

	gstPadProbeDrop   uintptr = 0
	gstPadProbeOK     uintptr = 1
	gstPadProbeRemove uintptr = 2

	gstEventFlushStart uint32 = 2563
	gstEventSegment    uint32 = 17934
)

// These public GStreamer structures are ABI-stable. TorrServer only builds the
// native bridge for 64-bit targets, so their pointer-sized fields are 8 bytes.
type gstMiniObjectABI struct {
	typeID          uintptr
	refCount        int32
	lockState       int32
	flags           uint32
	_               uint32
	copyFunction    uintptr
	disposeFunction uintptr
	freeFunction    uintptr
	privateUint     uint32
	_               uint32
	privatePointer  uintptr
}

type gstBufferABI struct {
	miniObject gstMiniObjectABI
	pool       uintptr
	pts        uint64
	dts        uint64
}

type gstEventABI struct {
	miniObject gstMiniObjectABI
	eventType  uint32
}

type gstSegmentABI struct {
	flags       uint32
	_           uint32
	rate        float64
	appliedRate float64
	format      int32
	_           uint32
	base        uint64
	offset      uint64
	start       uint64
	stop        uint64
	time        uint64
	position    uint64
	duration    uint64
}

type gstPadProbeInfoABI struct {
	probeType uint32
}

type gstPadProbeInfoWindowsABI struct {
	probeType uint32
	id        uint32
	data      uintptr
}

type gstPadProbeInfoUnixABI struct {
	probeType uint32
	_         uint32
	id        uint64
	data      uintptr
}

type gstPadProbeRegistration struct {
	api   *gstAPI
	pad   uintptr
	token uintptr
	id    atomic.Uint64
	state any
}

type videoStartProbeState struct {
	registration  *gstPadProbeRegistration
	requestedNS   uint64
	maxBackDiffNS uint64
	actualNS      atomic.Uint64
}

type videoSegmentClipProbeState struct {
	registration   *gstPadProbeRegistration
	requestedStart uint64
	segmentStart   atomic.Uint64
}

var (
	videoProbeStates    sync.Map
	videoProbeNextToken atomic.Uint64

	videoStartProbeCallback = purego.NewCallback(videoStartPadProbe)
	videoClipProbeCallback  = purego.NewCallback(videoSegmentClipPadProbe)
)

func (r *gstRunner) installVideoSeekProbes(pipeline uintptr, requestedNS uint64, accurate bool) {
	r.removeVideoSeekProbes()
	if gstRuntime == nil || pipeline == 0 || gstRuntime.gstPadAddProbe == nil || gstRuntime.gstPadRemoveProbe == nil {
		return
	}

	r.installVideoStartProbe(pipeline, requestedNS)
	clipStart := uint64(gstClockTimeNone)
	if accurate {
		clipStart = requestedNS
	}
	r.installVideoSegmentClipProbe(pipeline, clipStart)
}

func (r *gstRunner) installVideoStartProbe(pipeline uintptr, requestedNS uint64) {
	pad := gstPipelineElementPad(gstRuntime, pipeline, "mq", "src_0")
	if pad == 0 {
		return
	}

	maxBackDiffNS := uint64(max(r.task.Config.SegmentSeconds, 1)) * 1_000_000_000
	if r.task.Cue != nil {
		maxBackDiffNS = r.task.Cue.MaxDurationNS
	}
	state := &videoStartProbeState{
		requestedNS:   requestedNS,
		maxBackDiffNS: maxBackDiffNS,
	}
	state.actualNS.Store(gstClockTimeNone)
	r.videoStartProbe = addVideoPadProbe(
		gstRuntime,
		pad,
		gstPadProbeTypeEventDownstream|gstPadProbeTypeBuffer,
		videoStartProbeCallback,
		state,
	)
	if r.videoStartProbe != nil {
		state.registration = r.videoStartProbe
	}
}

func (r *gstRunner) installVideoSegmentClipProbe(pipeline uintptr, requestedStart uint64) {
	passthrough := !videoIsTranscoded(r.task.Config, r.task.Probe)
	if !passthrough || (!r.task.Probe.IsH264() && !r.task.Probe.IsH265()) {
		return
	}

	pad := gstPipelineElementPad(gstRuntime, pipeline, "video_timestamper", "src")
	if pad == 0 {
		return
	}

	state := &videoSegmentClipProbeState{requestedStart: requestedStart}
	state.segmentStart.Store(requestedStart)
	r.videoClipProbe = addVideoPadProbe(
		gstRuntime,
		pad,
		gstPadProbeTypeEventDownstream|gstPadProbeTypeBuffer,
		videoClipProbeCallback,
		state,
	)
	if r.videoClipProbe != nil {
		state.registration = r.videoClipProbe
	}
}

func gstPipelineElementPad(api *gstAPI, pipeline uintptr, elementName string, padName string) uintptr {
	if api == nil || api.gstBinGetByName == nil || api.gstElementGetStaticPad == nil || api.gstObjectUnref == nil {
		return 0
	}
	element := api.binGetByName(pipeline, elementName)
	if element == 0 {
		return 0
	}
	defer api.objectUnref(element)
	return api.elementGetStaticPad(element, padName)
}

func addVideoPadProbe(api *gstAPI, pad uintptr, mask uint32, callback uintptr, state any) *gstPadProbeRegistration {
	if api == nil || pad == 0 || callback == 0 || api.gstPadAddProbe == nil {
		if api != nil && pad != 0 && api.gstObjectUnref != nil {
			api.objectUnref(pad)
		}
		return nil
	}

	token := uintptr(videoProbeNextToken.Add(1))
	if token == 0 {
		token = uintptr(videoProbeNextToken.Add(1))
	}
	registration := &gstPadProbeRegistration{
		api:   api,
		pad:   pad,
		token: token,
		state: state,
	}
	videoProbeStates.Store(token, state)
	id := api.gstPadAddProbe(pad, mask, callback, token, 0)
	if id == 0 {
		videoProbeStates.Delete(token)
		if api.gstObjectUnref != nil {
			api.objectUnref(pad)
		}
		return nil
	}
	registration.id.Store(uint64(id))
	return registration
}

func (r *gstRunner) removeVideoSeekProbes() {
	removeVideoPadProbe(&r.videoStartProbe)
	removeVideoPadProbe(&r.videoClipProbe)
}

func removeVideoPadProbe(target **gstPadProbeRegistration) {
	registration := *target
	*target = nil
	if registration == nil {
		return
	}

	id := registration.id.Swap(0)
	if id != 0 && registration.api != nil && registration.api.gstPadRemoveProbe != nil {
		registration.api.gstPadRemoveProbe(registration.pad, uintptr(id))
	}
	videoProbeStates.Delete(registration.token)
	if registration.api != nil && registration.api.gstObjectUnref != nil {
		registration.api.objectUnref(registration.pad)
	}
}

func (r *gstRunner) applyPendingVideoStart() {
	registration := r.videoStartProbe
	if registration == nil {
		return
	}
	state, ok := registration.state.(*videoStartProbeState)
	if !ok {
		return
	}
	clockTime := state.actualNS.Swap(gstClockTimeNone)
	if clockTime == gstClockTimeNone {
		return
	}

	seconds := float64(clockTime) / 1_000_000_000
	if r.reader != nil {
		r.reader.SetTimelineOffsetNS(clockTime)
	}
	r.positionSeekSeconds = seconds
	r.setPosition(seconds)
	gstTaskDebugf(r.task, "video seek start requested=%.3fs actual=%.3fs", float64(state.requestedNS)/1_000_000_000, seconds)
}

func videoStartPadProbe(_ purego.CDecl, _ uintptr, info uintptr, userData uintptr) (result uintptr) {
	result = gstPadProbeOK
	defer func() {
		if recover() != nil {
			result = gstPadProbeOK
		}
	}()

	value, ok := videoProbeStates.Load(userData)
	if !ok {
		return gstPadProbeRemove
	}
	state, ok := value.(*videoStartProbeState)
	if !ok || state.registration == nil {
		return gstPadProbeRemove
	}

	probeType := gstProbeInfoType(info)
	if probeType&gstPadProbeTypeEventDownstream != 0 {
		event := gstProbeInfoData(info)
		if clockTime, ok := gstSegmentClockTime(state.registration.api, event); ok && state.accepts(clockTime) {
			state.actualNS.Store(clockTime)
		}
		return gstPadProbeOK
	}
	if probeType&gstPadProbeTypeBuffer == 0 {
		return gstPadProbeOK
	}

	buffer := gstProbeInfoData(info)
	clockTime, ok := gstBufferClockTime(buffer)
	if !ok || !state.accepts(clockTime) {
		return gstPadProbeOK
	}
	state.actualNS.Store(clockTime)
	state.registration.id.Store(0)
	return gstPadProbeRemove
}

func (state *videoStartProbeState) accepts(clockTime uint64) bool {
	return state.requestedNS <= state.maxBackDiffNS || clockTime >= state.requestedNS-state.maxBackDiffNS
}

func videoSegmentClipPadProbe(_ purego.CDecl, _ uintptr, info uintptr, userData uintptr) (result uintptr) {
	result = gstPadProbeOK
	defer func() {
		if recover() != nil {
			result = gstPadProbeOK
		}
	}()

	value, ok := videoProbeStates.Load(userData)
	if !ok {
		return gstPadProbeRemove
	}
	state, ok := value.(*videoSegmentClipProbeState)
	if !ok || state.registration == nil {
		return gstPadProbeRemove
	}

	probeType := gstProbeInfoType(info)
	if probeType&gstPadProbeTypeEventDownstream != 0 {
		event := gstProbeInfoData(info)
		switch gstEventType(event) {
		case gstEventFlushStart:
			state.segmentStart.Store(state.requestedStart)
		case gstEventSegment:
			if start, _, ok := gstTimeSegment(state.registration.api, event); ok {
				if state.requestedStart != gstClockTimeNone && start < state.requestedStart {
					start = state.requestedStart
				}
				state.segmentStart.Store(start)
			}
		}
		return gstPadProbeOK
	}
	if probeType&gstPadProbeTypeBuffer == 0 {
		return gstPadProbeOK
	}

	buffer := gstProbeInfoData(info)
	pts, ok := gstBufferPTS(buffer)
	start := state.segmentStart.Load()
	if ok && start != gstClockTimeNone && pts < start {
		return gstPadProbeDrop
	}
	return gstPadProbeOK
}

func gstProbeInfoType(info uintptr) uint32 {
	if info == 0 {
		return 0
	}
	return (*gstPadProbeInfoABI)(unsafe.Pointer(info)).probeType
}

func gstProbeInfoData(info uintptr) uintptr {
	if info == 0 {
		return 0
	}
	if runtime.GOOS == "windows" {
		return (*gstPadProbeInfoWindowsABI)(unsafe.Pointer(info)).data
	}
	return (*gstPadProbeInfoUnixABI)(unsafe.Pointer(info)).data
}

func gstEventType(event uintptr) uint32 {
	if event == 0 {
		return 0
	}
	return (*gstEventABI)(unsafe.Pointer(event)).eventType
}

func gstBufferPTS(buffer uintptr) (uint64, bool) {
	if buffer == 0 {
		return 0, false
	}
	pts := (*gstBufferABI)(unsafe.Pointer(buffer)).pts
	return pts, pts != gstClockTimeNone
}

func gstBufferClockTime(buffer uintptr) (uint64, bool) {
	if buffer == 0 {
		return 0, false
	}
	native := (*gstBufferABI)(unsafe.Pointer(buffer))
	if native.pts != gstClockTimeNone {
		return native.pts, true
	}
	if native.dts != gstClockTimeNone {
		return native.dts, true
	}
	return 0, false
}

func gstSegmentClockTime(api *gstAPI, event uintptr) (uint64, bool) {
	start, segmentTime, ok := gstTimeSegment(api, event)
	if !ok {
		return 0, false
	}
	if segmentTime != gstClockTimeNone {
		return segmentTime, true
	}
	return start, start != gstClockTimeNone
}

func gstTimeSegment(api *gstAPI, event uintptr) (start uint64, segmentTime uint64, ok bool) {
	if api == nil || event == 0 || gstEventType(event) != gstEventSegment || api.gstEventParseSegment == nil {
		return gstClockTimeNone, gstClockTimeNone, false
	}
	var segmentPointer uintptr
	api.gstEventParseSegment(event, unsafe.Pointer(&segmentPointer))
	if segmentPointer == 0 {
		return gstClockTimeNone, gstClockTimeNone, false
	}
	segment := (*gstSegmentABI)(unsafe.Pointer(segmentPointer))
	if segment.format != gstFormatTime {
		return gstClockTimeNone, gstClockTimeNone, false
	}
	return segment.start, segment.time, true
}
