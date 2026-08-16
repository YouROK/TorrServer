//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"runtime"
	"testing"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
)

func TestVideoProbePublicABILayout(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Fatal("GStreamer bridge requires a 64-bit target")
	}
	if got := unsafe.Sizeof(gstMiniObjectABI{}); got != 64 {
		t.Fatalf("GstMiniObject size=%d, want 64", got)
	}
	if got := unsafe.Offsetof(gstBufferABI{}.pts); got != 72 {
		t.Fatalf("GstBuffer.pts offset=%d, want 72", got)
	}
	if got := unsafe.Offsetof(gstBufferABI{}.dts); got != 80 {
		t.Fatalf("GstBuffer.dts offset=%d, want 80", got)
	}
	if got := unsafe.Offsetof(gstEventABI{}.eventType); got != 64 {
		t.Fatalf("GstEvent.type offset=%d, want 64", got)
	}
	if got := unsafe.Offsetof(gstSegmentABI{}.format); got != 24 {
		t.Fatalf("GstSegment.format offset=%d, want 24", got)
	}
	if got := unsafe.Offsetof(gstSegmentABI{}.start); got != 48 {
		t.Fatalf("GstSegment.start offset=%d, want 48", got)
	}
	if got := unsafe.Offsetof(gstSegmentABI{}.time); got != 64 {
		t.Fatalf("GstSegment.time offset=%d, want 64", got)
	}
	if got := unsafe.Offsetof(gstPadProbeInfoWindowsABI{}.data); got != 8 {
		t.Fatalf("Windows GstPadProbeInfo.data offset=%d, want 8", got)
	}
	if got := unsafe.Offsetof(gstPadProbeInfoUnixABI{}.data); got != 16 {
		t.Fatalf("Unix GstPadProbeInfo.data offset=%d, want 16", got)
	}
}

func TestVideoStartProbeUsesFirstAcceptableBufferTimestamp(t *testing.T) {
	buffer := &gstBufferABI{pts: uint64(95_000_000_000), dts: gstClockTimeNone}
	info, keepInfo := testPadProbeInfo(gstPadProbeTypeBuffer, uintptr(unsafe.Pointer(buffer)))
	state := &videoStartProbeState{
		requestedNS:   100_000_000_000,
		maxBackDiffNS: 6_000_000_000,
	}
	state.actualNS.Store(gstClockTimeNone)
	registration := &gstPadProbeRegistration{api: &gstAPI{}}
	registration.id.Store(17)
	registration.state = state
	state.registration = registration
	token := uintptr(videoProbeNextToken.Add(1))
	videoProbeStates.Store(token, state)
	t.Cleanup(func() { videoProbeStates.Delete(token) })

	result := videoStartPadProbe(purego.CDecl{}, 0, info, token)
	if result != gstPadProbeRemove {
		t.Fatalf("probe result=%d, want REMOVE", result)
	}
	if got := state.actualNS.Load(); got != buffer.pts {
		t.Fatalf("actual timestamp=%d, want %d", got, buffer.pts)
	}
	if got := registration.id.Load(); got != 0 {
		t.Fatalf("probe id=%d after REMOVE, want 0", got)
	}

	runtime.KeepAlive(buffer)
	runtime.KeepAlive(keepInfo)
}

func TestVideoStartProbeRejectsTimestampTooFarBehind(t *testing.T) {
	state := &videoStartProbeState{
		requestedNS:   100_000_000_000,
		maxBackDiffNS: 6_000_000_000,
	}
	if state.accepts(93_999_999_999) {
		t.Fatal("timestamp before the allowed seek window was accepted")
	}
	if !state.accepts(94_000_000_000) {
		t.Fatal("timestamp at the allowed seek window was rejected")
	}
}

func TestVideoSegmentClipProbeDropsPreSeekBuffer(t *testing.T) {
	buffer := &gstBufferABI{pts: uint64(91_000_000_000), dts: gstClockTimeNone}
	info, keepInfo := testPadProbeInfo(gstPadProbeTypeBuffer, uintptr(unsafe.Pointer(buffer)))
	state := &videoSegmentClipProbeState{requestedStart: 92_000_000_000}
	state.segmentStart.Store(92_000_000_000)
	registration := &gstPadProbeRegistration{api: &gstAPI{}, state: state}
	state.registration = registration
	token := uintptr(videoProbeNextToken.Add(1))
	videoProbeStates.Store(token, state)
	t.Cleanup(func() { videoProbeStates.Delete(token) })

	if result := videoSegmentClipPadProbe(purego.CDecl{}, 0, info, token); result != gstPadProbeDrop {
		t.Fatalf("probe result=%d, want DROP", result)
	}
	buffer.pts = 92_000_000_000
	if result := videoSegmentClipPadProbe(purego.CDecl{}, 0, info, token); result != gstPadProbeOK {
		t.Fatalf("probe result=%d at seek boundary, want OK", result)
	}

	runtime.KeepAlive(buffer)
	runtime.KeepAlive(keepInfo)
}

func TestVideoSegmentClipProbeUsesDownstreamTimeSegment(t *testing.T) {
	segment := &gstSegmentABI{
		format: gstFormatTime,
		start:  92_759_000_000,
		time:   92_759_000_000,
	}
	event := &gstEventABI{eventType: gstEventSegment}
	info, keepInfo := testPadProbeInfo(gstPadProbeTypeEventDownstream, uintptr(unsafe.Pointer(event)))
	api := &gstAPI{
		gstEventParseSegment: func(_ uintptr, output unsafe.Pointer) {
			*(*uintptr)(output) = uintptr(unsafe.Pointer(segment))
		},
	}
	state := &videoSegmentClipProbeState{requestedStart: gstClockTimeNone}
	state.segmentStart.Store(gstClockTimeNone)
	registration := &gstPadProbeRegistration{api: api, state: state}
	state.registration = registration
	token := uintptr(videoProbeNextToken.Add(1))
	videoProbeStates.Store(token, state)
	t.Cleanup(func() { videoProbeStates.Delete(token) })

	if result := videoSegmentClipPadProbe(purego.CDecl{}, 0, info, token); result != gstPadProbeOK {
		t.Fatalf("probe result=%d, want OK", result)
	}
	if got := state.segmentStart.Load(); got != segment.start {
		t.Fatalf("segment start=%d, want %d", got, segment.start)
	}

	runtime.KeepAlive(segment)
	runtime.KeepAlive(event)
	runtime.KeepAlive(keepInfo)
}

func TestRemoveVideoPadProbeReleasesNativeRegistration(t *testing.T) {
	var removedPad, removedID, unrefed uintptr
	api := &gstAPI{
		gstPadRemoveProbe: func(pad uintptr, id uintptr) {
			removedPad = pad
			removedID = id
		},
		gstObjectUnref: func(object uintptr) {
			unrefed = object
		},
	}
	registration := &gstPadProbeRegistration{api: api, pad: 11, token: 12}
	registration.id.Store(13)
	videoProbeStates.Store(registration.token, struct{}{})
	target := registration

	removeVideoPadProbe(&target)
	if target != nil {
		t.Fatal("registration was not cleared")
	}
	if removedPad != 11 || removedID != 13 || unrefed != 11 {
		t.Fatalf("removed pad/id/unref=%d/%d/%d, want 11/13/11", removedPad, removedID, unrefed)
	}
	if _, ok := videoProbeStates.Load(registration.token); ok {
		t.Fatal("probe state remains registered")
	}
}

func TestVideoStartProbeNativeCallback(t *testing.T) {
	conf := DefaultConfig()
	setupGStreamer(conf)
	api, err := loadGST(conf)
	if err != nil {
		t.Skipf("gstreamer runtime is not available: %v", err)
	}
	if err := api.init(); err != nil {
		t.Fatalf("gst init failed: %v", err)
	}

	pipeline, err := api.parseLaunch("fakesrc num-buffers=1 do-timestamp=true ! identity name=probe ! fakesink")
	if err != nil {
		t.Fatalf("parse launch failed: %v", err)
	}
	defer api.objectUnref(pipeline)

	pad := gstPipelineElementPad(api, pipeline, "probe", "src")
	if pad == 0 {
		t.Fatal("test probe pad is not available")
	}
	state := &videoStartProbeState{maxBackDiffNS: uint64(time.Second)}
	state.actualNS.Store(gstClockTimeNone)
	registration := addVideoPadProbe(api, pad, gstPadProbeTypeBuffer, videoStartProbeCallback, state)
	if registration == nil {
		t.Fatal("failed to install native video probe")
	}
	state.registration = registration
	defer removeVideoPadProbe(&registration)

	bus := api.pipelineGetBus(pipeline)
	if bus == 0 {
		t.Fatal("test pipeline bus is not available")
	}
	defer api.objectUnref(bus)
	defer api.elementSetState(pipeline, gstStateNull)
	if result := api.elementSetState(pipeline, gstStatePlaying); result == gstStateChangeFailure {
		t.Fatal("failed to start test pipeline")
	}
	message := api.gstBusTimedPopFiltered(bus, uint64(5*time.Second), gstMessageEOS|gstMessageError)
	if message == 0 {
		t.Fatal("test pipeline did not finish")
	}
	api.miniObjectUnref(message)
	if got := state.actualNS.Load(); got == gstClockTimeNone {
		t.Fatal("native callback did not observe a buffer timestamp")
	}
}

func testPadProbeInfo(probeType uint32, data uintptr) (uintptr, any) {
	if runtime.GOOS == "windows" {
		info := &gstPadProbeInfoWindowsABI{probeType: probeType, data: data}
		return uintptr(unsafe.Pointer(info)), info
	}
	info := &gstPadProbeInfoUnixABI{probeType: probeType, data: data}
	return uintptr(unsafe.Pointer(info)), info
}
