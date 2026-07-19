//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"context"
	"errors"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func TestCreatePipelineArgsUsesSingleBufferAppSink(t *testing.T) {
	runner := &gstRunner{
		task: &Task{
			SourceURL: "http://127.0.0.1/video",
			Config:    Config{}.normalized(),
			Probe: ProbeInfo{
				Container: "Matroska",
				Tracks: []TrackInfo{
					{
						Type:     "video",
						Index:    0,
						CapsName: "video/x-h264",
					},
				},
			},
		},
		audioIndex: -1,
	}

	args := runner.createPipelineArgs()
	want := "appsink name=out emit-signals=false sync=false max-buffers=1"
	if !strings.Contains(args, want) {
		t.Fatalf("createPipelineArgs() appsink =\n%s\nwant %q", args, want)
	}
	if strings.Contains(args, "queue2") || strings.Contains(args, "temp-template=") {
		t.Fatalf("createPipelineArgs() must rely on TorrServer cache, got:\n%s", args)
	}
}

func TestCreatePipelineArgsUsesMultiqueueWithoutBranchQueueLimits(t *testing.T) {
	clearGStreamerRuntimeVersion(t)

	args := newVersionedVideoPipelineRunner(1.28).createPipelineArgs()

	for _, want := range []string{
		"multiqueue name=mq use-buffering=false max-size-buffers=5",
		"d.video_0 ! mq.sink_0",
		"mq.src_0 ! h264parse",
	} {
		if !strings.Contains(args, want) {
			t.Fatalf("createPipelineArgs() =\n%s\nwant %q", args, want)
		}
	}
	if strings.Contains(args, "d.video_0 ! queue ") || strings.Contains(args, " leaky=0 !") {
		t.Fatalf("createPipelineArgs() must not use per-branch queue limits:\n%s", args)
	}
}

func TestCreatePipelineArgsCopiesAACAudio(t *testing.T) {
	clearGStreamerRuntimeVersion(t)

	runner := newVersionedVideoPipelineRunner(1.28)
	runner.audioIndex = 1
	runner.task.Probe.Tracks = append(runner.task.Probe.Tracks,
		TrackInfo{Type: "audio", Index: 0, Codec: "AC-3", CapsName: "audio/x-ac3"},
		TrackInfo{Type: "audio", Index: 1, Codec: "MPEG-4 AAC", CapsName: "audio/mpeg"},
	)

	args := runner.createPipelineArgs()
	want := "d.audio_1 ! mq.sink_1 mq.src_1 ! aacparse ! audio/mpeg,mpegversion=4,stream-format=raw ! mux.audio_0"
	if !strings.Contains(args, want) {
		t.Fatalf("createPipelineArgs() =\n%s\nwant %q", args, want)
	}
	for _, forbidden := range []string{"decodebin ! audioconvert", "avenc_aac bitrate="} {
		if strings.Contains(args, forbidden) {
			t.Fatalf("AAC copy pipeline must not contain %q:\n%s", forbidden, args)
		}
	}
}

func TestCreatePipelineArgsEncodesNonAACAudio(t *testing.T) {
	clearGStreamerRuntimeVersion(t)

	runner := newVersionedVideoPipelineRunner(1.28)
	runner.audioIndex = 0
	runner.task.Probe.Tracks = append(runner.task.Probe.Tracks,
		TrackInfo{Type: "audio", Index: 0, Codec: "AC-3", CapsName: "audio/x-ac3"},
	)

	args := runner.createPipelineArgs()
	for _, want := range []string{
		"d.audio_0 ! mq.sink_1 mq.src_1 ! decodebin ! audioconvert",
		"avenc_aac bitrate=",
		"aacparse ! audio/mpeg,mpegversion=4,stream-format=raw,rate=48000,channels=2 ! mux.audio_0",
	} {
		if !strings.Contains(args, want) {
			t.Fatalf("createPipelineArgs() =\n%s\nwant %q", args, want)
		}
	}
}

func TestCreatePipelineArgsUsesProbeAudioCapsWhenConfigAuto(t *testing.T) {
	clearGStreamerRuntimeVersion(t)

	runner := newVersionedVideoPipelineRunner(1.28)
	runner.audioIndex = 0
	runner.task.Probe.Tracks = append(runner.task.Probe.Tracks,
		TrackInfo{
			Type:     "audio",
			Index:    0,
			Codec:    "AC-3",
			CapsName: "audio/x-ac3",
			Channels: 6,
			Rate:     44100,
		},
	)

	args := runner.createPipelineArgs()
	for _, want := range []string{
		"audio/x-raw,format=F32LE,layout=interleaved,rate=44100,channels=6,channel-mask=(bitmask)0x000000000000003f",
		"aacparse ! audio/mpeg,mpegversion=4,stream-format=raw,rate=44100,channels=6 ! mux.audio_0",
	} {
		if !strings.Contains(args, want) {
			t.Fatalf("createPipelineArgs() =\n%s\nwant %q", args, want)
		}
	}
}

func TestCreatePipelineArgsConfigOverridesProbeAudioCaps(t *testing.T) {
	clearGStreamerRuntimeVersion(t)

	runner := newVersionedVideoPipelineRunner(1.28)
	runner.audioIndex = 0
	runner.task.Config.AACChannels = 1
	runner.task.Config.AACSamplerate = 32000
	runner.task.Probe.Tracks = append(runner.task.Probe.Tracks,
		TrackInfo{
			Type:     "audio",
			Index:    0,
			Codec:    "AC-3",
			CapsName: "audio/x-ac3",
			Channels: 6,
			Rate:     44100,
		},
	)

	args := runner.createPipelineArgs()
	for _, want := range []string{
		"audio/x-raw,format=F32LE,layout=interleaved,rate=32000,channels=1",
		"aacparse ! audio/mpeg,mpegversion=4,stream-format=raw,rate=32000,channels=1 ! mux.audio_0",
	} {
		if !strings.Contains(args, want) {
			t.Fatalf("createPipelineArgs() =\n%s\nwant %q", args, want)
		}
	}
}

func TestCreatePipelineArgsMultiqueueUsesFixedLimits(t *testing.T) {
	clearGStreamerRuntimeVersion(t)

	runner := newVersionedVideoPipelineRunner(1.28)
	runner.task.Config.SegmentSeconds = 7

	args := runner.createPipelineArgs()
	want := "multiqueue name=mq use-buffering=false max-size-buffers=5 max-size-bytes=0 max-size-time=0"
	if !strings.Contains(args, want) {
		t.Fatalf("createPipelineArgs() =\n%s\nwant %q", args, want)
	}
	mqEnd := strings.Index(args, " d.video_0 ! mq.sink_0")
	if mqEnd < 0 {
		t.Fatalf("createPipelineArgs() missing video branch:\n%s", args)
	}
	mqStart := strings.Index(args, "multiqueue name=mq")
	if mqStart < 0 {
		t.Fatalf("createPipelineArgs() missing multiqueue:\n%s", args)
	}
	mqArgs := args[mqStart:mqEnd]
	if !strings.Contains(mqArgs, "max-size-bytes=0") || !strings.Contains(mqArgs, "max-size-time=0") {
		t.Fatalf("multiqueue must explicitly disable byte/time limits:\n%s", mqArgs)
	}
}

func TestCreatePipelineArgsGStreamer122OmitsNewerProperties(t *testing.T) {
	clearGStreamerRuntimeVersion(t)

	args := newVersionedVideoPipelineRunner(1.22).createPipelineArgs()

	if strings.Contains(args, " retry-backoff-factor=") || strings.Contains(args, " retry-backoff-max=") {
		t.Fatalf("GStreamer 1.22 pipeline must not contain 1.26 souphttpsrc retry-backoff properties:\n%s", args)
	}
	if !strings.Contains(args, " max-buffers=1") {
		t.Fatalf("GStreamer 1.22 pipeline must use one appsink buffer:\n%s", args)
	}
	if strings.Contains(args, " leaky-type=") {
		t.Fatalf("GStreamer 1.22 pipeline must not contain 1.28 appsink leaky-type property:\n%s", args)
	}
	if !strings.Contains(args, " drop=false") {
		t.Fatalf("GStreamer 1.22 pipeline must use appsink drop=false fallback:\n%s", args)
	}
}

func TestCreatePipelineArgsGStreamer124UsesSingleAppSinkBuffer(t *testing.T) {
	clearGStreamerRuntimeVersion(t)

	args := newVersionedVideoPipelineRunner(1.24).createPipelineArgs()

	if strings.Contains(args, " retry-backoff-factor=") || strings.Contains(args, " retry-backoff-max=") {
		t.Fatalf("GStreamer 1.24 pipeline must not contain 1.26 retry-backoff properties:\n%s", args)
	}
	if !strings.Contains(args, "max-buffers=1") || strings.Contains(args, "max-bytes=") {
		t.Fatalf("GStreamer 1.24 pipeline must use one appsink buffer without byte limit:\n%s", args)
	}
	if !strings.Contains(args, " drop=false") {
		t.Fatalf("GStreamer 1.24 pipeline must use drop=false fallback before 1.28:\n%s", args)
	}
}

func TestCreatePipelineArgsAppSinkOmitsByteLimit(t *testing.T) {
	clearGStreamerRuntimeVersion(t)

	runner := newVersionedVideoPipelineRunner(1.28)
	args := runner.createPipelineArgs()

	if !strings.Contains(args, " max-buffers=1") {
		t.Fatalf("pipeline must use max-buffers=1:\n%s", args)
	}
	if strings.Contains(args, " max-bytes=") || strings.Contains(args, " max-time=") {
		t.Fatalf("appsink must not use byte/time limits:\n%s", args)
	}
	if !strings.Contains(args, " leaky-type=none") {
		t.Fatalf("GStreamer 1.28+ pipeline must use appsink leaky-type:\n%s", args)
	}
}

func TestCreatePipelineArgsGStreamer126UsesSoupRetryBackoff(t *testing.T) {
	clearGStreamerRuntimeVersion(t)

	args := newVersionedVideoPipelineRunner(1.26).createPipelineArgs()

	if !strings.Contains(args, " retry-backoff-factor=0.5 retry-backoff-max=10") {
		t.Fatalf("GStreamer 1.26+ pipeline must contain souphttpsrc retry-backoff properties:\n%s", args)
	}
	if !strings.Contains(args, "max-buffers=1") || strings.Contains(args, "max-bytes=") {
		t.Fatalf("GStreamer 1.26 pipeline must use one appsink buffer without byte limit:\n%s", args)
	}
	if !strings.Contains(args, " drop=false") {
		t.Fatalf("GStreamer 1.26 pipeline must use drop=false fallback before 1.28:\n%s", args)
	}
}

func TestCreatePipelineArgsGStreamer128UsesAppSinkLeakyType(t *testing.T) {
	clearGStreamerRuntimeVersion(t)

	args := newVersionedVideoPipelineRunner(1.28).createPipelineArgs()

	if !strings.Contains(args, "max-buffers=1") || strings.Contains(args, "max-bytes=") {
		t.Fatalf("GStreamer 1.28 pipeline must use one appsink buffer without byte limit:\n%s", args)
	}
	if !strings.Contains(args, " leaky-type=none") {
		t.Fatalf("GStreamer 1.28+ pipeline must use appsink leaky-type:\n%s", args)
	}
	if strings.Contains(args, " drop=false") {
		t.Fatalf("GStreamer 1.28+ pipeline must not use deprecated appsink drop property:\n%s", args)
	}
}

func TestCreatePipelineArgsUsesRuntimeVersionWhenAvailable(t *testing.T) {
	previous := gstRuntime
	gstRuntime = &gstAPI{version: gstVersionInfo{major: 1, minor: 22}}
	t.Cleanup(func() {
		gstRuntime = previous
	})

	args := newVersionedVideoPipelineRunner(1.28).createPipelineArgs()

	if strings.Contains(args, " max-bytes=") || strings.Contains(args, " max-time=") || strings.Contains(args, " leaky-type=none") {
		t.Fatalf("runtime GStreamer 1.22 version must override config feature gates:\n%s", args)
	}
	if !strings.Contains(args, " max-buffers=1") {
		t.Fatalf("runtime GStreamer 1.22 pipeline must use one appsink buffer:\n%s", args)
	}
	if !strings.Contains(args, " drop=false") {
		t.Fatalf("runtime GStreamer 1.22 pipeline must use drop=false fallback:\n%s", args)
	}
}

func clearGStreamerRuntimeVersion(t *testing.T) {
	t.Helper()

	previous := gstRuntime
	gstRuntime = nil
	t.Cleanup(func() {
		gstRuntime = previous
	})
}

func newVersionedVideoPipelineRunner(gstVersion float64) *gstRunner {
	return &gstRunner{
		task: &Task{
			SourceURL: "http://127.0.0.1/video",
			Config: Config{
				GSTVersion: gstVersion,
			}.normalized(),
			Probe: ProbeInfo{
				Container: "Matroska",
				Tracks: []TrackInfo{
					{
						Type:     "video",
						Index:    0,
						CapsName: "video/x-h264",
					},
				},
			},
		},
		audioIndex: -1,
	}
}

func TestAACEncoderDefaultsToLibAV(t *testing.T) {
	runner := &gstRunner{}

	if got := runner.aacEncoder(); got != "avenc_aac" {
		t.Fatalf("aacEncoder() = %q, want avenc_aac", got)
	}
}

func TestGSTRuntimeRootsPreferUserPath(t *testing.T) {
	userRoot := makeFakeGSTRoot(t)

	roots := gstRuntimeRoots(Config{GSTPath: userRoot})
	if len(roots) == 0 {
		t.Fatalf("gstRuntimeRoots() returned no roots, want user path %q", userRoot)
	}
	if !sameTestPath(roots[0], userRoot) {
		t.Fatalf("gstRuntimeRoots()[0] = %q, want user path %q; roots=%v", roots[0], userRoot, roots)
	}

	probeRoots := gstDiscovererRoots(Config{GSTPath: userRoot})
	if len(probeRoots) == 0 {
		t.Fatalf("gstDiscovererRoots() returned no roots, want user path %q", userRoot)
	}
	if !sameTestPath(probeRoots[0], userRoot) {
		t.Fatalf("gstDiscovererRoots()[0] = %q, want user path %q; roots=%v", probeRoots[0], userRoot, probeRoots)
	}
}

func TestGSTRuntimeRootsSkipInvalidUserPath(t *testing.T) {
	invalidRoot := t.TempDir()

	for _, root := range gstRuntimeRoots(Config{GSTPath: invalidRoot}) {
		if sameTestPath(root, invalidRoot) {
			t.Fatalf("gstRuntimeRoots() included invalid user path %q: %v", invalidRoot, root)
		}
	}
	for _, root := range gstDiscovererRoots(Config{GSTPath: invalidRoot}) {
		if sameTestPath(root, invalidRoot) {
			t.Fatalf("gstDiscovererRoots() included invalid user path %q: %v", invalidRoot, root)
		}
	}
}

func makeFakeGSTRoot(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	var baseLib string
	switch runtime.GOOS {
	case "windows":
		baseLib = filepath.Join(root, "bin", "libgstreamer-1.0-0.dll")
	case "darwin":
		baseLib = filepath.Join(root, "lib", "libgstreamer-1.0.0.dylib")
	default:
		baseLib = filepath.Join(root, "lib", "libgstreamer-1.0.so.0")
	}
	if err := os.MkdirAll(filepath.Dir(baseLib), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(baseLib, []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func sameTestPath(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

func TestVideoOnlyPipelineHasNoAudioBranch(t *testing.T) {
	runner := &gstRunner{
		task: &Task{
			SourceURL: "http://127.0.0.1/video",
			Config:    Config{}.normalized(),
			Probe: ProbeInfo{
				Container: "Matroska",
				Tracks: []TrackInfo{
					{
						Type:     "video",
						Index:    0,
						CapsName: "video/x-h264",
					},
				},
			},
		},
		audioIndex: -1,
	}

	args := runner.createPipelineArgs()

	if strings.Contains(args, "d.audio_") {
		t.Fatalf("video-only pipeline contains audio branch:\n%s", args)
	}
	if !strings.Contains(args, "mux.video_0") {
		t.Fatalf("video branch is absent:\n%s", args)
	}
}

func TestCreatePipelineArgsTranscodesVP8(t *testing.T) {
	clearGStreamerRuntimeVersion(t)
	runner := newVersionedVideoPipelineRunner(1.28)
	runner.task.Probe.Tracks[0].CapsName = "video/x-vp8"
	runner.task.Config.TranscodeVP8 = true

	args := runner.createPipelineArgs()
	if !strings.Contains(args, "mq.src_0 ! decodebin ! videoconvert ! video/x-raw,format=I420 ! x264enc") {
		t.Fatalf("VP8 transcode branch is absent:\n%s", args)
	}
}

func TestCreatePipelineArgsTranscodesAVI(t *testing.T) {
	clearGStreamerRuntimeVersion(t)
	runner := newVersionedVideoPipelineRunner(1.28)
	runner.task.Probe.Container = "AVI"
	runner.task.Probe.ContainerCapsName = "video/x-msvideo"
	runner.task.Config.TranscodeAVI = true

	args := runner.createPipelineArgs()
	for _, want := range []string{"avidemux name=d", "mq.src_0 ! decodebin", "x264enc name=video_encoder"} {
		if !strings.Contains(args, want) {
			t.Fatalf("AVI transcode pipeline is missing %q:\n%s", want, args)
		}
	}
}

func TestCreatePipelineArgsToneMapsHDR(t *testing.T) {
	clearGStreamerRuntimeVersion(t)
	runner := newVersionedVideoPipelineRunner(1.28)
	runner.task.Probe.Tracks[0].CapsName = "video/x-h265"
	runner.task.Probe.Tracks[0].VideoTransfer = "pq"
	runner.task.Config.HDRToSDR = true
	runner.task.Config.UseGPU = false

	args := runner.createPipelineArgs()
	if !strings.Contains(args, "decodebin ! hdrtonemap transfer=pq use-opencl=false ! video/x-raw,format=I420 ! x264enc") {
		t.Fatalf("HDR tone mapping branch is absent:\n%s", args)
	}
}

func TestCreatePipelineArgsUsesX264Ultrafast(t *testing.T) {
	clearGStreamerRuntimeVersion(t)
	runner := newVersionedVideoPipelineRunner(1.28)
	runner.task.Config.TranscodeH264 = true
	runner.task.Config.X264Ultrafast = true

	args := runner.createPipelineArgs()
	if !strings.Contains(args, "x264enc name=video_encoder tune=zerolatency speed-preset=ultrafast") {
		t.Fatalf("x264 ultrafast preset is absent:\n%s", args)
	}
}

func TestSetPipelineStateRejectsAsyncTimeout(t *testing.T) {
	previous := gstRuntime
	t.Cleanup(func() {
		gstRuntime = previous
	})

	gstRuntime = &gstAPI{
		gstElementSetState: func(uintptr, int32) int32 {
			return gstStateChangeAsync
		},
		gstElementGetState: func(uintptr, unsafe.Pointer, unsafe.Pointer, uint64) int32 {
			return gstStateChangeAsync
		},
		gstBusTimedPopFiltered: func(uintptr, uint64, int32) uintptr {
			return 0
		},
	}

	err := (&gstRunner{}).setPipelineState(1, 2, gstStatePlaying)
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStartPipelineUsesActualQueriedSeekPosition(t *testing.T) {
	previous := gstRuntime
	t.Cleanup(func() {
		gstRuntime = previous
	})

	var seekPad uintptr
	var seekPosition int64
	seekPending := false
	gstRuntime = &gstAPI{
		gstParseLaunch: func(string, unsafe.Pointer) uintptr { return 1 },
		gstBinGetByName: func(_ uintptr, name string) uintptr {
			if name == "mq" {
				return 4
			}
			return 2
		},
		gstPipelineGetBus: func(uintptr) uintptr { return 3 },
		gstElementSetState: func(uintptr, int32) int32 {
			return gstStateChangeSuccess
		},
		gstElementGetState: func(uintptr, unsafe.Pointer, unsafe.Pointer, uint64) int32 {
			return gstStateChangeSuccess
		},
		gstElementGetStaticPad: func(element uintptr, name string) uintptr {
			if element != 4 || name != "src_0" {
				t.Fatalf("unexpected seek target: element=%d pad=%q", element, name)
			}
			return 5
		},
		gstEventNewSeek: func(rate float64, format int32, flags int32, startType int32, start int64, stopType int32, stop int64) uintptr {
			if rate != 1 || format != gstFormatTime || startType != gstSeekTypeSet || stopType != gstSeekTypeNone || stop != -1 {
				t.Fatalf("unexpected seek event: rate=%v format=%d flags=%d startType=%d start=%d stopType=%d stop=%d", rate, format, flags, startType, start, stopType, stop)
			}
			seekPosition = start
			return 6
		},
		gstPadSendEvent: func(pad uintptr, event uintptr) int32 {
			seekPad = pad
			if event != 6 {
				t.Fatalf("event=%d, want 6", event)
			}
			seekPending = true
			return 1
		},
		gstElementQueryPosition: func(_ uintptr, _ int32, cur unsafe.Pointer) int32 {
			*(*int64)(cur) = int64(16 * time.Second)
			return 1
		},
		gstBusTimedPopFiltered: func(_ uintptr, _ uint64, filter int32) uintptr {
			if filter == gstMessageAsyncDone && seekPending {
				seekPending = false
				return 7
			}
			return 0
		},
		gstObjectUnref:     func(uintptr) {},
		gstMiniObjectUnref: func(uintptr) {},
	}

	runner := &gstRunner{
		task: &Task{
			Config: Config{}.normalized(),
		},
	}
	actual, err := runner.startPipeline(12)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.stopPipeline()

	if actual != 16 {
		t.Fatalf("actual=%v, want 16", actual)
	}
	if seekPad != 5 {
		t.Fatalf("seek pad=%d, want mq.src_0", seekPad)
	}
	if seekPosition != int64(12*time.Second) {
		t.Fatalf("seek position=%d, want %d", seekPosition, int64(12*time.Second))
	}
	runner.stopPipeline()

	gstRuntime.gstElementQueryPosition = func(uintptr, int32, unsafe.Pointer) int32 {
		return 0
	}
	fallbackRunner := &gstRunner{
		task: &Task{Config: Config{}.normalized()},
	}
	actual, err = fallbackRunner.startPipeline(12)
	if err != nil {
		t.Fatalf("startPipeline failed when position query was unavailable: %v", err)
	}
	defer fallbackRunner.stopPipeline()
	if actual != 12 {
		t.Fatalf("fallback actual=%v, want requested position 12", actual)
	}
}

func TestReusePipelineUsesQueriedPositionOrRequestedFallback(t *testing.T) {
	previous := gstRuntime
	t.Cleanup(func() {
		gstRuntime = previous
	})

	tests := []struct {
		name          string
		queryResult   int32
		queryPosition int64
		want          float64
	}{
		{name: "queried position", queryResult: 1, queryPosition: int64(16 * time.Second), want: 16},
		{name: "unavailable position", queryResult: 0, queryPosition: int64(16 * time.Second), want: 12},
		{name: "negative position", queryResult: 1, queryPosition: -1, want: 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner, _ := newReusePipelineTestRunner(t, tt.queryResult, tt.queryPosition, 1, true)
			defer func() {
				runner.stopPipeline()
				runner.releaseTransientState()
			}()

			actual, err := runner.reusePipeline(12, false, pipelineStateTimeout)
			if err != nil {
				t.Fatalf("reusePipeline failed: %v", err)
			}
			if actual != tt.want {
				t.Fatalf("actual=%v, want %v", actual, tt.want)
			}
		})
	}
}

func TestReusePipelineContinuesWhenAsyncDoneIsNotObserved(t *testing.T) {
	previous := gstRuntime
	t.Cleanup(func() {
		gstRuntime = previous
	})

	runner, _ := newReusePipelineTestRunner(t, 1, int64(16*time.Second), 1, false)
	defer func() {
		runner.stopPipeline()
		runner.releaseTransientState()
	}()

	actual, err := runner.reusePipeline(12, false, time.Millisecond)
	if err != nil {
		t.Fatalf("reusePipeline failed without ASYNC_DONE: %v", err)
	}
	if actual != 16 {
		t.Fatalf("actual=%v, want queried position 16", actual)
	}
}

func TestReusePipelineRequiresVideoEncoderWhenTranscoding(t *testing.T) {
	previous := gstRuntime
	t.Cleanup(func() {
		gstRuntime = previous
	})

	runner, _ := newReusePipelineTestRunner(t, 1, int64(16*time.Second), 1, true)
	runner.task.Config.TranscodeH264 = true
	runner.task.Probe.Tracks = []TrackInfo{{Type: "video", CapsName: "video/x-h264"}}
	defer func() {
		runner.stopPipeline()
		runner.releaseTransientState()
	}()

	_, err := runner.reusePipeline(12, false, pipelineStateTimeout)
	if err == nil || !strings.Contains(err.Error(), "video encoder is not available") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReusePipelineFailureFreezesAndReleasesPipeline(t *testing.T) {
	previous := gstRuntime
	t.Cleanup(func() {
		gstRuntime = previous
	})

	runner, unrefs := newReusePipelineTestRunner(t, 1, int64(16*time.Second), 0, false)
	registration := &gstPadProbeRegistration{
		api:   gstRuntime,
		pad:   9,
		token: 10,
		state: struct{}{},
	}
	registration.id.Store(11)
	videoProbeStates.Store(registration.token, registration.state)
	t.Cleanup(func() { videoProbeStates.Delete(registration.token) })
	runner.videoStartProbe = registration

	if runner.Seek(12) {
		t.Fatal("Seek succeeded after the native seek event failed")
	}
	if !runner.IsFrozen() {
		t.Fatal("runner was not frozen after reusePipeline failure")
	}
	if runner.pipeline != 0 || runner.bus != 0 || runner.sink != 0 {
		t.Fatalf("native pipeline was retained: pipeline=%d bus=%d sink=%d", runner.pipeline, runner.bus, runner.sink)
	}
	if runner.reader != nil {
		t.Fatal("MP4 reader was retained after reusePipeline failure")
	}
	if runner.videoStartProbe != nil || runner.videoClipProbe != nil {
		t.Fatal("video probes were retained after reusePipeline failure")
	}
	if _, ok := videoProbeStates.Load(registration.token); ok {
		t.Fatal("video probe state remains registered after reusePipeline failure")
	}
	for _, handle := range []uintptr{1, 2, 3, 9} {
		if unrefs[handle] != 1 {
			t.Fatalf("native handle %d unref count=%d, want 1", handle, unrefs[handle])
		}
	}
}

func newReusePipelineTestRunner(t *testing.T, queryResult int32, queryPosition int64, sendSeekResult int32, emitAsyncDone bool) (*gstRunner, map[uintptr]int) {
	t.Helper()

	seekPending := false
	unrefs := make(map[uintptr]int)
	api := &gstAPI{}
	api.gstBinGetByName = func(_ uintptr, name string) uintptr {
		switch name {
		case "mux":
			return 4
		case "mq":
			return 5
		default:
			return 0
		}
	}
	api.gstElementSetState = func(_ uintptr, state int32) int32 {
		if state == gstStatePlaying {
			// reusePipeline has already consumed the seek message. Disabling the
			// fake bus prevents its watcher from spinning after this point.
			api.gstBusTimedPopFiltered = nil
		}
		return gstStateChangeSuccess
	}
	api.gstElementGetState = func(uintptr, unsafe.Pointer, unsafe.Pointer, uint64) int32 {
		return gstStateChangeSuccess
	}
	api.gstElementGetStaticPad = func(element uintptr, name string) uintptr {
		if element != 5 || name != "src_0" {
			t.Fatalf("unexpected seek target: element=%d pad=%q", element, name)
		}
		return 6
	}
	api.gstEventNewSeek = func(float64, int32, int32, int32, int64, int32, int64) uintptr {
		return 7
	}
	api.gstPadSendEvent = func(pad uintptr, event uintptr) int32 {
		if pad != 6 || event != 7 {
			t.Fatalf("unexpected seek event: pad=%d event=%d", pad, event)
		}
		seekPending = sendSeekResult != 0
		return sendSeekResult
	}
	api.gstElementQueryPosition = func(_ uintptr, format int32, position unsafe.Pointer) int32 {
		if format != gstFormatTime {
			t.Fatalf("query format=%d, want time", format)
		}
		*(*int64)(position) = queryPosition
		return queryResult
	}
	api.gstBusTimedPopFiltered = func(_ uintptr, _ uint64, filter int32) uintptr {
		if filter == gstMessageAsyncDone && seekPending && emitAsyncDone {
			seekPending = false
			return 8
		}
		return 0
	}
	api.gstPadRemoveProbe = func(uintptr, uintptr) {}
	api.gstObjectUnref = func(handle uintptr) {
		unrefs[handle]++
	}
	api.gstMiniObjectUnref = func(uintptr) {}
	gstRuntime = api

	task := &Task{Config: Config{}.normalized()}
	runner := &gstRunner{
		task:     task,
		pipeline: 1,
		bus:      2,
		sink:     3,
	}
	task.runner = runner
	runner.ensureTransientState()
	return runner, unrefs
}

func TestRunnerPositionUsesSegmentEnd(t *testing.T) {
	runner := &gstRunner{}
	runner.positionSeekSeconds = 10

	runner.acceptSegment(Segment{
		StartSeconds: 2,
		EndSeconds:   8,
	})

	if got := runner.position(); got != 8 {
		t.Fatalf("position=%v, want 8", got)
	}
}

func TestEarlyEndOfStreamValidation(t *testing.T) {
	tests := []struct {
		name       string
		durationNS int64
		position   float64
		wantEarly  bool
	}{
		{name: "unknown duration", position: 0},
		{name: "short media", durationNS: int64(60 * time.Second), position: 0},
		{name: "before threshold", durationNS: int64(10 * time.Minute), position: 479.999, wantEarly: true},
		{name: "at threshold", durationNS: int64(10 * time.Minute), position: 480},
		{name: "after threshold", durationNS: int64(10 * time.Minute), position: 599},
		{name: "invalid position", durationNS: int64(10 * time.Minute), position: math.NaN(), wantEarly: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &gstRunner{task: &Task{Probe: ProbeInfo{DurationNS: tt.durationNS}}}
			runner.setPosition(tt.position)

			err := runner.earlyEndOfStreamError()
			if got := errors.Is(err, ErrEarlyEndOfStream); got != tt.wantEarly {
				t.Fatalf("error=%v, early=%t, want %t", err, got, tt.wantEarly)
			}
		})
	}
}

func TestPipelineReadTimeout(t *testing.T) {
	if pipelineReadTimeout != 45*time.Second {
		t.Fatalf("pipelineReadTimeout=%s, want 45s", pipelineReadTimeout)
	}
}

func TestGetSegmentReturnsRuntimeBusError(t *testing.T) {
	previous := gstRuntime
	t.Cleanup(func() {
		gstRuntime = previous
	})

	messageText := append([]byte("decoder failed at runtime"), 0)
	gerror := make([]byte, 16)
	*(*uintptr)(unsafe.Pointer(&gerror[8])) = uintptr(unsafe.Pointer(&messageText[0]))

	messageReturned := false
	gstRuntime = &gstAPI{
		gstBusTimedPopFiltered: func(uintptr, uint64, int32) uintptr {
			if messageReturned {
				return 0
			}
			messageReturned = true
			return 77
		},
		gstMessageParseError: func(_ uintptr, errOut unsafe.Pointer, debugOut unsafe.Pointer) {
			*(*uintptr)(errOut) = uintptr(unsafe.Pointer(&gerror[0]))
			*(*uintptr)(debugOut) = 0
		},
		gstMiniObjectUnref: func(uintptr) {},
		gErrorFree:         func(uintptr) {},
		gFree:              func(uintptr) {},
		gstElementSetState: func(uintptr, int32) int32 {
			return gstStateChangeSuccess
		},
		gstObjectUnref: func(uintptr) {},
	}

	runner := &gstRunner{
		task: &Task{
			Config: Config{}.normalized(),
		},
		statePlaying: true,
		pipeline:     1,
		bus:          2,
		sink:         3,
		reader: Mp4BoxReader(
			func([]byte) {},
			func(Segment) {},
			6,
			0,
			false,
		),
	}

	_, err := runner.GetSegment(context.Background(), 0, 0)
	if err == nil || !strings.Contains(err.Error(), "decoder failed at runtime") {
		t.Fatalf("unexpected error: %v", err)
	}

	runtime.KeepAlive(messageText)
	runtime.KeepAlive(gerror)
}

type fakePipelineRunner struct {
	frozen     atomic.Bool
	ensureInit func(context.Context, int, int) error
	getSegment func(context.Context, int, int) (Segment, error)
	seek       func(float64) bool
}

func (f *fakePipelineRunner) EnsureInit(ctx context.Context, audio int, startIndex int) error {
	if f.ensureInit != nil {
		return f.ensureInit(ctx, audio, startIndex)
	}
	return nil
}

func (f *fakePipelineRunner) GetSegment(ctx context.Context, index int, audio int) (Segment, error) {
	if f.getSegment != nil {
		return f.getSegment(ctx, index, audio)
	}
	return Segment{}, nil
}

func (f *fakePipelineRunner) Seek(seconds float64) bool {
	if f.seek != nil {
		return f.seek(seconds)
	}
	return true
}

func (f *fakePipelineRunner) Frozen() {
	f.frozen.Store(true)
}
func (f *fakePipelineRunner) Dispose() {}
func (f *fakePipelineRunner) IsFrozen() bool {
	return f.frozen.Load()
}

func TestSegmentCatchupCutoffSeconds(t *testing.T) {
	if maxSegmentCatchupSeconds != 60 {
		t.Fatalf("maxSegmentCatchupSeconds = %d, want 60", maxSegmentCatchupSeconds)
	}
}

func TestTaskWithFirstDistantSegmentSeeks(t *testing.T) {
	var seekSeconds []float64
	var pulled []int
	task := &Task{
		LastSentSegment: -1,
		Config:          Config{SegmentSeconds: 6}.normalized(),
	}
	task.runner = &fakePipelineRunner{
		seek: func(seconds float64) bool {
			seekSeconds = append(seekSeconds, seconds)
			return true
		},
		getSegment: func(_ context.Context, index int, _ int) (Segment, error) {
			pulled = append(pulled, index)
			return Segment{Header: []byte{1}}, nil
		},
	}

	if err := task.WithSegment(context.Background(), 100, 0, func(Segment) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if len(seekSeconds) != 1 || seekSeconds[0] != 600 {
		t.Fatalf("seekSeconds=%v, want [600]", seekSeconds)
	}
	if len(pulled) != 1 || pulled[0] != 100 {
		t.Fatalf("pulled=%v, want [100]", pulled)
	}
}

func TestFrozenTaskSeeksToRequestedSegment(t *testing.T) {
	var seekSeconds []float64
	var pulled []int
	runner := &fakePipelineRunner{
		seek: func(seconds float64) bool {
			seekSeconds = append(seekSeconds, seconds)
			return true
		},
		getSegment: func(_ context.Context, index int, _ int) (Segment, error) {
			pulled = append(pulled, index)
			return Segment{Header: []byte{1}}, nil
		},
	}
	runner.frozen.Store(true)
	task := &Task{
		LastSentSegment: 41,
		Config:          Config{SegmentSeconds: 6}.normalized(),
		runner:          runner,
	}

	if err := task.WithSegment(context.Background(), 42, 0, func(Segment) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if len(seekSeconds) != 1 || seekSeconds[0] != 252 {
		t.Fatalf("seekSeconds=%v, want [252]", seekSeconds)
	}
	if len(pulled) != 1 || pulled[0] != 42 {
		t.Fatalf("pulled=%v, want [42]", pulled)
	}
}

func TestTaskRejectsSegmentIndexThatOverflowsSeekTime(t *testing.T) {
	called := false
	task := &Task{
		LastSentSegment: 0,
		Config:          Config{SegmentSeconds: 6}.normalized(),
		runner: &fakePipelineRunner{
			seek: func(float64) bool {
				called = true
				return true
			},
			getSegment: func(context.Context, int, int) (Segment, error) {
				called = true
				return Segment{}, nil
			},
		},
	}

	index := int(^uint(0) >> 1)
	err := task.WithSegment(context.Background(), index, 0, func(Segment) error { return nil })
	if !errors.Is(err, ErrInvalidIdentifier) {
		t.Fatalf("error=%v, want ErrInvalidIdentifier", err)
	}
	if called {
		t.Fatal("pipeline was called for an overflowing segment index")
	}
}

func TestTaskRejectsSegmentPastKnownDuration(t *testing.T) {
	task := &Task{
		LastSentSegment: 0,
		Config:          Config{SegmentSeconds: 6}.normalized(),
		Probe:           ProbeInfo{DurationNS: int64(60 * time.Second)},
		runner:          &fakePipelineRunner{},
	}

	err := task.WithSegment(context.Background(), 10, 0, func(Segment) error { return nil })
	if !errors.Is(err, ErrEndOfStreamExhausted) {
		t.Fatalf("error=%v, want ErrEndOfStreamExhausted", err)
	}
}

func TestTaskWithSegmentCatchesUpWithinCutoff(t *testing.T) {
	var pulled []int
	task := &Task{
		LastSentSegment: 0,
		Config: Config{
			SegmentSeconds: 6,
		}.normalized(),
	}
	task.runner = &fakePipelineRunner{
		getSegment: func(_ context.Context, index int, _ int) (Segment, error) {
			pulled = append(pulled, index)
			return Segment{Header: []byte{1}}, nil
		},
		seek: func(float64) bool {
			t.Fatal("Seek must not be called within catchup cutoff")
			return false
		},
	}

	if err := task.WithSegment(context.Background(), 10, 0, func(Segment) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if len(pulled) != 10 || pulled[0] != 1 || pulled[len(pulled)-1] != 10 {
		t.Fatalf("pulled=%v, want segments 1..10", pulled)
	}
	if task.LastSentSegment != 10 {
		t.Fatalf("LastSentSegment=%d, want 10", task.LastSentSegment)
	}
}

func TestTaskWithSegmentSeeksPastCatchupCutoff(t *testing.T) {
	var pulled []int
	var seekSeconds []float64
	task := &Task{
		LastSentSegment: 0,
		Config: Config{
			SegmentSeconds: 6,
		}.normalized(),
	}
	task.runner = &fakePipelineRunner{
		getSegment: func(_ context.Context, index int, _ int) (Segment, error) {
			pulled = append(pulled, index)
			return Segment{Header: []byte{1}}, nil
		},
		seek: func(seconds float64) bool {
			seekSeconds = append(seekSeconds, seconds)
			return true
		},
	}

	if err := task.WithSegment(context.Background(), 11, 0, func(Segment) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if len(seekSeconds) != 1 || seekSeconds[0] != 66 {
		t.Fatalf("seekSeconds=%v, want [66]", seekSeconds)
	}
	if len(pulled) != 1 || pulled[0] != 11 {
		t.Fatalf("pulled=%v, want only segment 11 after seek", pulled)
	}
}

func TestTaskIsFrozenConcurrentDispose(t *testing.T) {
	task := &Task{
		runner: &fakePipelineRunner{},
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = task.IsFrozen()
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		task.Dispose()
	}()

	wg.Wait()
}

func TestTaskFrozenClearsInitMP4(t *testing.T) {
	runner := &fakePipelineRunner{}
	task := &Task{
		Config: Config{SegmentSeconds: 6}.normalized(),
		runner: runner,
	}
	task.setInitMP4([]byte{1, 2, 3})

	task.Frozen()

	if !runner.IsFrozen() {
		t.Fatal("runner was not frozen")
	}
	if task.hasInitMP4() {
		t.Fatal("init mp4 was retained after freeze")
	}
}

func TestTaskFreezeIfInactiveRechecksLastActive(t *testing.T) {
	runner := &fakePipelineRunner{}
	now := time.Now().UTC()
	task := &Task{
		Config:     Config{SegmentSeconds: 6}.normalized(),
		lastActive: now,
		runner:     runner,
	}

	if task.FreezeIfInactive(now.Add(-time.Minute)) {
		t.Fatal("active task was frozen")
	}
	if runner.IsFrozen() {
		t.Fatal("runner was frozen despite fresh activity")
	}

	task.activeMu.Lock()
	task.lastActive = now.Add(-2 * time.Minute)
	task.activeMu.Unlock()
	if !task.FreezeIfInactive(now.Add(-time.Minute)) {
		t.Fatal("inactive task was not frozen")
	}
}

func TestGstRunnerFreezeReleasesTransientState(t *testing.T) {
	task := &Task{
		Config: Config{
			SegmentSeconds: 6,
			Subtitles:      true,
		}.normalized(),
		Probe: ProbeInfo{Tracks: []TrackInfo{
			{Index: 2, Type: "subtitle", Codec: "subrip"},
		}},
	}
	runner := &gstRunner{task: task}
	task.runner = runner
	runner.ensureTransientState()

	oldReader := runner.reader
	oldStore := runner.subtitleStores[2]
	if oldReader == nil || oldStore == nil {
		t.Fatal("transient state was not initialized")
	}
	task.setInitMP4([]byte{1, 2, 3})
	oldReader.sourcePayload.Write(make([]byte, 1024))
	oldStore.appendVTT("00:00:01.000 --> 00:00:02.000\ntext\n\n", 0, 0)

	runner.freezeAtPosition(12)

	if runner.reader != nil || runner.subtitleStores != nil || runner.subtitleSinks != nil {
		t.Fatal("transient state was retained after freeze")
	}
	if task.hasInitMP4() {
		t.Fatal("init mp4 was retained after runner freeze")
	}
	if !runner.IsFrozen() || runner.position() != 12 || runner.statePlaying {
		t.Fatalf("unexpected frozen state: frozen=%v position=%v playing=%v", runner.IsFrozen(), runner.position(), runner.statePlaying)
	}

	runner.ensureTransientState()
	if runner.reader == nil || runner.reader == oldReader {
		t.Fatal("mp4 reader was not recreated")
	}
	if store := runner.subtitleStores[2]; store == nil || store == oldStore {
		t.Fatal("subtitle store was not recreated")
	}
}

func TestDetachTaskDoesNotRemoveReplacement(t *testing.T) {
	oldTask := &Task{}
	newTask := &Task{}

	service := &Service{
		tasks: map[string]*Task{
			"id": newTask,
		},
	}

	if _, ok := service.detachTask("id", oldTask); ok {
		t.Fatal("old snapshot removed replacement task")
	}

	service.mu.RLock()
	current := service.tasks["id"]
	service.mu.RUnlock()

	if current != newTask {
		t.Fatal("replacement task was removed")
	}
}

func TestProbeCacheReturnsFreshCopy(t *testing.T) {
	service := &Service{probeCache: make(map[string]probeCacheEntry)}
	probe := ProbeInfo{
		Container: "Matroska",
		Tracks: []TrackInfo{
			{Type: "video", CapsName: "video/x-h264"},
		},
	}

	service.setCachedProbe("hash", "1", probe)

	cached, ok := service.getCachedProbe("hash", "1")
	if !ok {
		t.Fatal("cached probe was not returned")
	}
	cached.Tracks[0].CapsName = "video/x-h265"

	cachedAgain, ok := service.getCachedProbe("hash", "1")
	if !ok {
		t.Fatal("cached probe disappeared")
	}
	if cachedAgain.Tracks[0].CapsName != "video/x-h264" {
		t.Fatalf("cached probe was mutated through returned slice: %+v", cachedAgain.Tracks[0])
	}
}

func TestProbeCacheExpires(t *testing.T) {
	service := &Service{
		probeCache: map[string]probeCacheEntry{
			probeCacheKey("hash", "1"): {
				probe:     ProbeInfo{Container: "Matroska"},
				expiresAt: time.Now().UTC().Add(-time.Second),
			},
		},
	}

	if _, ok := service.getCachedProbe("hash", "1"); ok {
		t.Fatal("expired cached probe was returned")
	}
	if len(service.probeCache) != 0 {
		t.Fatal("expired cached probe was not removed")
	}
}

func TestTaskEnsureInitDoesNotRequireSegment(t *testing.T) {
	task := &Task{
		LastSentSegment: -1,
		Config:          Config{SegmentSeconds: 6}.normalized(),
	}

	task.runner = &fakePipelineRunner{
		ensureInit: func(_ context.Context, _ int, startIndex int) error {
			if startIndex != 100 {
				t.Fatalf("startIndex=%d, want 100", startIndex)
			}
			task.setInitMP4([]byte{1, 2, 3})
			return nil
		},
		getSegment: func(context.Context, int, int) (Segment, error) {
			t.Fatal("GetSegment was called for init")
			return Segment{}, nil
		},
	}

	if err := task.EnsureInit(context.Background(), 0, 100); err != nil {
		t.Fatal(err)
	}
	if err := task.WithInitMP4(func(init []byte) error {
		if len(init) != 3 {
			t.Fatal("init mp4 was not stored")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if task.LastSentSegment != 99 {
		t.Fatalf("LastSentSegment=%d, want 99", task.LastSentSegment)
	}
}

func TestTaskWithSegmentReturnsConsumerErrorUnderLock(t *testing.T) {
	sentinel := errors.New("consumer failed")
	task := &Task{
		LastSentSegment: -1,
		Config:          Config{SegmentSeconds: 6}.normalized(),
		runner: &fakePipelineRunner{
			getSegment: func(context.Context, int, int) (Segment, error) {
				return Segment{Header: []byte("segment")}, nil
			},
		},
	}

	lockAttempted := make(chan struct{})
	lockAcquired := make(chan struct{})
	err := task.WithSegment(context.Background(), 0, 0, func(seg Segment) error {
		if seg.Empty() {
			t.Fatal("segment is empty")
		}

		go func() {
			close(lockAttempted)
			task.mu.Lock()
			task.mu.Unlock()
			close(lockAcquired)
		}()

		<-lockAttempted
		select {
		case <-lockAcquired:
			t.Fatal("Task.mu was released before segment consumer returned")
		case <-time.After(20 * time.Millisecond):
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("error=%v, want sentinel", err)
	}

	select {
	case <-lockAcquired:
	case <-time.After(time.Second):
		t.Fatal("Task.mu was not released after segment consumer returned")
	}
}

func TestTaskWithSegmentSkipsConsumerAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	task := &Task{
		LastSentSegment: -1,
		Config:          Config{SegmentSeconds: 6}.normalized(),
		runner: &fakePipelineRunner{
			getSegment: func(context.Context, int, int) (Segment, error) {
				cancel()
				return Segment{Header: []byte("segment")}, nil
			},
		},
	}

	consumerCalled := false
	err := task.WithSegment(ctx, 0, 0, func(Segment) error {
		consumerCalled = true
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v, want context.Canceled", err)
	}
	if consumerCalled {
		t.Fatal("segment consumer was called after request cancellation")
	}
}
