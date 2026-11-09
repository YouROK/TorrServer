//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"runtime"
	"sync"
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

func TestHEVCHDRCapsProbeRetainsInitialMetadata(t *testing.T) {
	conf := DefaultConfig()
	setupGStreamer(conf)
	api, err := loadGST(conf)
	if err != nil {
		t.Skipf("gstreamer runtime is not available: %v", err)
	}
	if err := api.init(); err != nil {
		t.Fatal(err)
	}
	t.Logf("native GStreamer %s on %s/%s", api.version, runtime.GOOS, runtime.GOARCH)
	var capsFromString func(string) uintptr
	var capsEqual func(uintptr, uintptr) int32
	var pushEvent func(uintptr, uintptr) int32
	var eventSeqnum func(uintptr) uint32
	var setSeqnum func(uintptr, uint32)
	var ref func(uintptr) uintptr
	purego.RegisterLibFunc(&capsFromString, api.handles[1], "gst_caps_from_string")
	purego.RegisterLibFunc(&capsEqual, api.handles[1], "gst_caps_is_equal")
	purego.RegisterLibFunc(&pushEvent, api.handles[1], "gst_pad_push_event")
	purego.RegisterLibFunc(&eventSeqnum, api.handles[1], "gst_event_get_seqnum")
	purego.RegisterLibFunc(&setSeqnum, api.handles[1], "gst_event_set_seqnum")
	purego.RegisterLibFunc(&ref, api.handles[1], "gst_mini_object_ref")
	const base = "video/x-h265,stream-format=(string)hvc1,alignment=(string)au,width=(int)1920,height=(int)1080,codec_data=(buffer)010203"
	const hdr = ",mastering-display-info=(string)arbitrary-display,content-light-level=(string)731:219"
	for _, initial := range []string{hdr, ",mastering-display-info=(uint)47,content-light-level=(guint64)913", ",mastering-display-info=(string)other-display,content-light-level=(string)921:107", ",mastering-display-info=(string)display-only", ",content-light-level=(string)619:71", ""} {
		t.Run(initial, func(t *testing.T) {
			pipeline, err := api.parseLaunch("identity name=probe ! fakesink name=out async=false")
			if err != nil {
				t.Fatal(err)
			}
			defer api.objectUnref(pipeline)
			pad := gstPipelineElementPad(api, pipeline, "probe", "src")
			state := &hevcHDRCapsProbeState{api: api}
			registration := addVideoPadProbe(api, pad, gstPadProbeTypeEventDownstream, hevcHDRCapsProbeCallback, state)
			if registration == nil {
				t.Fatal("failed to install HDR probe")
			}
			defer removeVideoPadProbe(&registration)
			downstream := gstPipelineElementPad(api, pipeline, "out", "sink")
			var expected uintptr
			var received int
			callback := purego.NewCallback(func(_ purego.CDecl, _ uintptr, info uintptr, _ uintptr) uintptr {
				event := gstProbeInfoData(info)
				if gstEventType(event) == gstEventCaps {
					var caps uintptr
					api.gstEventParseCaps(event, unsafe.Pointer(&caps))
					if capsEqual(caps, expected) == 0 {
						t.Error("downstream caps differ from expected typed fields")
					}
					if eventSeqnum(event) != 1234 {
						t.Error("CAPS event sequence number changed")
					}
					received++
				}
				return gstPadProbeOK
			})
			capture := addVideoPadProbe(api, downstream, gstPadProbeTypeEventDownstream, callback, nil)
			defer removeVideoPadProbe(&capture)
			defer api.elementSetState(pipeline, gstStateNull)
			if api.elementSetState(pipeline, gstStatePlaying) == gstStateChangeFailure {
				t.Fatal("failed to activate pipeline")
			}
			for i, fields := range []string{initial, initial, ",mastering-display-info=(string)arbitrary-display", ",content-light-level=(string)731:219", ",mastering-display-info=(string)different,content-light-level=(string)991:83", ""} {
				if i == 3 {
					flush := &gstEventABI{eventType: gstEventFlushStart}
					info, keep := testPadProbeInfo(gstPadProbeTypeEventDownstream, uintptr(unsafe.Pointer(flush)))
					if hevcHDRCapsPadProbe(purego.CDecl{}, 0, info, registration.token) != gstPadProbeOK {
						t.Error("flush rejected")
					}
					runtime.KeepAlive(keep)
					runtime.KeepAlive(flush)
				}
				current := base
				if i > 1 {
					current = "video/x-h265,stream-format=(string)hev1,alignment=(string)au,width=(int)1280,height=(int)720,codec_data=(buffer)040506"
				}
				caps := capsFromString(current + fields)
				expected = capsFromString(current + initial)
				if caps == 0 || expected == 0 {
					t.Fatal("failed to parse synthetic test caps")
				}
				event := api.gstEventNewCaps(caps)
				api.miniObjectUnref(caps)
				setSeqnum(event, 1234)
				ref(event) // Keep the original alive to verify copy-on-write.
				before := received
				if pushEvent(pad, event) == 0 {
					t.Error("CAPS event rejected")
				}
				if received != before+1 {
					t.Error("CAPS event did not reach downstream probe")
				}
				var original uintptr
				api.gstEventParseCaps(event, unsafe.Pointer(&original))
				unchanged := capsFromString(current + fields)
				if capsEqual(original, unchanged) == 0 {
					t.Error("original event was mutated")
				}
				api.miniObjectUnref(unchanged)
				api.miniObjectUnref(event)
				api.miniObjectUnref(expected)
			}
			flush := &gstEventABI{eventType: gstEventFlushStart}
			buffer := &gstBufferABI{pts: 17, dts: 13}
			for _, item := range []struct {
				kind uint32
				data uintptr
			}{
				{gstPadProbeTypeEventDownstream, uintptr(unsafe.Pointer(flush))},
				{gstPadProbeTypeBuffer, uintptr(unsafe.Pointer(buffer))},
			} {
				info, keep := testPadProbeInfo(item.kind, item.data)
				if hevcHDRCapsPadProbe(purego.CDecl{}, 0, info, registration.token) != gstPadProbeOK || gstProbeInfoData(info) != item.data {
					t.Error("non-CAPS data changed")
				}
				runtime.KeepAlive(keep)
			}
			runtime.KeepAlive(flush)
			runtime.KeepAlive(buffer)
			// Hold one extra reference so teardown ownership can be checked safely.
			if state.initialCaps == 0 {
				t.Fatal("initial HDR snapshot was not captured")
			}
			snapshot := ref(state.initialCaps)
			removeVideoPadProbe(&registration)
			if got := (*gstMiniObjectABI)(unsafe.Pointer(snapshot)).refCount; got != 1 {
				t.Errorf("snapshot references after removal=%d, want 1", got)
			}
			api.miniObjectUnref(snapshot)
		})
	}
}

func TestHEVCHDRCapsProbeRemovalWaitsForCallback(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	removed := make(chan struct{})
	callbackDone := make(chan struct{})
	cleanupDone := make(chan struct{})
	var mu sync.Mutex
	var released []uintptr
	api := &gstAPI{
		gstEventParseCaps: func(_ uintptr, caps unsafe.Pointer) {
			close(entered)
			<-release
			*(*uintptr)(caps) = 41
		},
		gstMiniObjectCopy:  func(uintptr) uintptr { return 42 },
		gstMiniObjectUnref: func(object uintptr) { mu.Lock(); released = append(released, object); mu.Unlock() },
		gstPadRemoveProbe:  func(uintptr, uintptr) { close(removed) },
		gstObjectUnref:     func(uintptr) {},
	}
	state := &hevcHDRCapsProbeState{api: api}
	registration := &gstPadProbeRegistration{api: api, pad: 1, token: uintptr(videoProbeNextToken.Add(1)), state: state}
	registration.id.Store(2)
	videoProbeStates.Store(registration.token, state)
	event := &gstEventABI{eventType: gstEventCaps}
	info, keep := testPadProbeInfo(gstPadProbeTypeEventDownstream, uintptr(unsafe.Pointer(event)))
	go func() {
		defer close(callbackDone)
		hevcHDRCapsPadProbe(purego.CDecl{}, 0, info, registration.token)
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("callback did not start")
	}
	go func() {
		target := registration
		removeVideoPadProbe(&target)
		close(cleanupDone)
	}()
	select {
	case <-removed:
	case <-time.After(time.Second):
		t.Fatal("removal did not start")
	}
	mu.Lock()
	if len(released) != 0 {
		t.Error("snapshot released while callback was using it")
	}
	mu.Unlock()
	close(release)
	select {
	case <-callbackDone:
	case <-time.After(time.Second):
		t.Fatal("callback did not finish")
	}
	select {
	case <-cleanupDone:
	case <-time.After(time.Second):
		t.Fatal("cleanup did not finish")
	}
	if len(released) != 1 || released[0] != 42 || !state.closed || state.initialCaps != 0 {
		t.Fatalf("snapshot cleanup: released=%v closed=%v caps=%d", released, state.closed, state.initialCaps)
	}
	if _, ok := videoProbeStates.Load(registration.token); ok {
		t.Fatal("probe state remains registered")
	}
	runtime.KeepAlive(event)
	runtime.KeepAlive(keep)
}
