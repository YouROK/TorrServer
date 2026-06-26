//go:build (windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64))

package gstreamer

import (
	"context"
	"errors"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	gstInitOnce sync.Once
	gstRuntime  *gstAPI
	gstInitErr  error

	gstInitStatus componentStatus
)

const (
	tempFSBlockSeconds       = 30
	tempFSBaseBlocks         = 3
	tempFSFallbackBlockBytes = 32 * 1024 * 1024
)

type gstRunner struct {
	task *Task

	audioIndex int

	statePlaying bool
	readySegment struct {
		index    int
		complete bool
		segment  Segment
	}

	positionSeconds     atomic.Uint64
	positionSeekSeconds float64

	reader *mp4BoxReader

	pipeline uintptr
	bus      uintptr
	sink     uintptr

	frozen atomic.Bool
}

func newPipelineRunner(task *Task, audio int) (pipelineRunner, error) {
	gstInitOnce.Do(func() {
		initGStreamerRuntime(task.Config)
	})
	if gstInitErr != nil {
		return nil, errors.Join(ErrPipelineDisabled, gstInitErr)
	}

	runner := &gstRunner{
		task:       task,
		audioIndex: validAudioIndex(task.Probe, audio),
		reader: Mp4BoxReader(
			func(data []byte) {
				task.setInitMP4(data)
			},
			func(seg Segment) {
				r := task.runner
				if gr, ok := r.(*gstRunner); ok {
					gr.readySegment.segment = seg
					gr.readySegment.complete = true
					if seg.StartSeconds >= 0 {
						gr.setPosition(seg.StartSeconds + gr.positionSeekSeconds)
					}
				}
			},
			float64(task.Config.SegmentSeconds),
		),
	}
	runner.readySegment.index = -1
	return runner, nil
}

func initGStreamerRuntime(conf Config) {
	setupGStreamer(conf)

	var err error
	gstRuntime, err = loadGST(conf)
	if err != nil {
		gstInitErr = err
		return
	}
	gstInitStatus.Found = true

	if err = gstRuntime.init(); err != nil {
		gstInitErr = err
		return
	}
	gstInitStatus.Available = true
	gstInitErr = nil
}

func setupGStreamer(conf Config) {
	_ = os.Setenv("GST_REGISTRY", filepath.Join(os.TempDir(), "torrserver-gstreamer-registry.bin"))

	roots := gstRuntimeRoots(conf)
	if len(roots) == 0 {
		return
	}

	prependExistingEnvPaths("PATH", gstBinDirCandidates(roots))

	switch runtime.GOOS {
	case "linux":
		prependExistingEnvPaths("LD_LIBRARY_PATH", gstLibraryDirCandidates(roots))
	case "darwin":
		prependExistingEnvPaths("DYLD_LIBRARY_PATH", gstLibraryDirCandidates(roots))
	}

	var gstPlugins string
	switch runtime.GOOS {
	case "windows":
		gstPlugins = filepath.Join(roots[0], "lib", "gstreamer-1.0")
	case "linux", "darwin":
		gstPlugins = firstExistingPath(gstPluginCandidates(roots))
	}
	if gstPlugins != "" {
		_ = os.Setenv("GST_PLUGIN_PATH", gstPlugins)
		_ = os.Setenv("GST_PLUGIN_SYSTEM_PATH_1_0", gstPlugins)
	}

	var gstPluginScanner string
	switch runtime.GOOS {
	case "windows":
		gstPluginScanner = filepath.Join(roots[0], "libexec", "gstreamer-1.0", "gst-plugin-scanner.exe")
	case "linux", "darwin":
		gstPluginScanner = firstExistingPath(gstPluginScannerCandidates(roots))
	}
	if gstPluginScanner != "" {
		_ = os.Setenv("GST_PLUGIN_SCANNER", gstPluginScanner)
	}
}

func gstRuntimeRoots(conf Config) []string {
	var roots []string
	if root := embeddedGSTRuntimeRoot(); root != "" {
		roots = appendUniquePath(roots, root)
	}
	if root := portableGSTRuntimeRoot(); root != "" {
		roots = appendUniquePath(roots, root)
	}
	if conf.GSTPath != "" {
		roots = appendUniquePath(roots, conf.GSTPath)
	}
	if runtime.GOOS != "darwin" {
		return roots
	}
	for _, root := range []string{
		"/Library/Frameworks/GStreamer.framework/Versions/1.0",
		"/opt/homebrew",
		"/usr/local",
	} {
		roots = appendUniquePath(roots, root)
	}
	return roots
}

func portableGSTRuntimeRoot() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}

	root := filepath.Join(filepath.Dir(exe), "gst-lib")
	if info, err := os.Stat(root); err == nil && info.IsDir() {
		return root
	}
	return ""
}

func appendUniquePath(paths []string, path string) []string {
	if path == "" {
		return paths
	}
	clean := filepath.Clean(path)
	for _, existing := range paths {
		if strings.EqualFold(filepath.Clean(existing), clean) {
			return paths
		}
	}
	return append(paths, path)
}

func gstBinDirCandidates(roots []string) []string {
	candidates := make([]string, 0, len(roots))
	for _, root := range roots {
		candidates = append(candidates, filepath.Join(root, "bin"))
	}
	return candidates
}

func gstLibraryDirCandidates(roots []string) []string {
	var candidates []string
	for _, root := range roots {
		candidates = append(candidates,
			filepath.Join(root, "lib"),
			filepath.Join(root, "lib64"),
			filepath.Join(root, "lib", runtime.GOARCH+"-linux-gnu"),
			filepath.Join(root, "lib", "x86_64-linux-gnu"),
			filepath.Join(root, "lib", "aarch64-linux-gnu"),
		)
	}
	return candidates
}

func gstPluginCandidates(roots []string) []string {
	var candidates []string
	for _, root := range roots {
		candidates = append(candidates,
			filepath.Join(root, "lib", "gstreamer-1.0"),
			filepath.Join(root, "lib64", "gstreamer-1.0"),
			filepath.Join(root, "lib", runtime.GOARCH+"-linux-gnu", "gstreamer-1.0"),
			filepath.Join(root, "lib", "x86_64-linux-gnu", "gstreamer-1.0"),
			filepath.Join(root, "lib", "aarch64-linux-gnu", "gstreamer-1.0"),
		)
	}
	return candidates
}

func gstPluginScannerCandidates(roots []string) []string {
	var candidates []string
	for _, root := range roots {
		candidates = append(candidates,
			filepath.Join(root, "libexec", "gstreamer-1.0", "gst-plugin-scanner"),
			filepath.Join(root, "lib", "gstreamer-1.0", "gst-plugin-scanner"),
			filepath.Join(root, "lib64", "gstreamer-1.0", "gst-plugin-scanner"),
		)
	}
	return candidates
}

func existingPaths(candidates []string) []string {
	paths := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			paths = appendUniquePath(paths, candidate)
		}
	}
	return paths
}

func firstExistingPath(candidates []string) string {
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func prependExistingEnvPaths(key string, candidates []string) {
	prependEnvPaths(key, existingPaths(candidates))
}

func prependEnvPaths(key string, values []string) {
	if len(values) == 0 {
		return
	}

	separator := string(os.PathListSeparator)
	parts := make([]string, 0, len(values)+1)
	for _, value := range values {
		parts = appendUniqueEnvPath(parts, value)
	}

	current := os.Getenv(key)
	for _, part := range strings.Split(current, separator) {
		if part != "" {
			parts = appendUniqueEnvPath(parts, part)
		}
	}

	_ = os.Setenv(key, strings.Join(parts, separator))
}

func appendUniqueEnvPath(paths []string, path string) []string {
	if path == "" {
		return paths
	}
	for _, existing := range paths {
		if strings.EqualFold(existing, path) {
			return paths
		}
	}
	return append(paths, path)
}

func prependEnvPath(key string, value string) {
	if value == "" {
		return
	}

	current := os.Getenv(key)
	if current == "" {
		_ = os.Setenv(key, value)
		return
	}

	separator := string(os.PathListSeparator)
	for _, part := range strings.Split(current, separator) {
		if strings.EqualFold(part, value) {
			return
		}
	}

	_ = os.Setenv(key, value+separator+current)
}

func (r *gstRunner) createPipelineArgs() string {
	conf := r.task.Config
	probe := r.task.Probe

	queueNS := int64(conf.PipelineTimeSeconds) * int64(time.Second)
	audioQueueBytes := conf.PipelineAudioQueue * 1024 * 1024
	videoQueueBytes := conf.PipelineVideoQueue * 1024 * 1024

	var sb strings.Builder

	sb.WriteString("souphttpsrc ")
	sb.WriteString("location=\"")
	sb.WriteString(r.task.SourceURL)
	sb.WriteString("\" is-live=false keep-alive=true timeout=60 retries=5 ")
	if conf.GSTVersion >= 1.26 {
		sb.WriteString("retry-backoff-factor=0.5 retry-backoff-max=10 ")
	}
	r.writeSourceQueue(&sb, videoQueueBytes, queueNS)
	sb.WriteString(" ! matroskademux name=d ")

	switch {
	case probe.IsH264():
		if conf.TranscodeH264 {
			r.transcodeToH264(&sb, videoQueueBytes, queueNS)
		} else {
			sb.WriteString("d.video_0 ! queue max-size-buffers=0 max-size-bytes=")
			sb.WriteString(strconv.Itoa(videoQueueBytes))
			sb.WriteString(" max-size-time=")
			sb.WriteString(strconv.FormatInt(queueNS, 10))
			sb.WriteString(" leaky=0 ! h264parse config-interval=0 ! h264timestamper ! video/x-h264,stream-format=avc,alignment=au ! mux.video_0 ")
		}

	case probe.IsH265():
		if conf.TranscodeH265 {
			r.transcodeToH264(&sb, videoQueueBytes, queueNS)
		} else {
			sb.WriteString("d.video_0 ! queue max-size-buffers=0 max-size-bytes=")
			sb.WriteString(strconv.Itoa(videoQueueBytes))
			sb.WriteString(" max-size-time=")
			sb.WriteString(strconv.FormatInt(queueNS, 10))
			sb.WriteString(" leaky=0 ! h265parse config-interval=0 ! h265timestamper ! video/x-h265,stream-format=hvc1,alignment=au ! mux.video_0 ")
		}

	case probe.IsAV1():
		if conf.TranscodeAV1 {
			r.transcodeToH264(&sb, videoQueueBytes, queueNS)
		} else {
			sb.WriteString("d.video_0 ! queue max-size-buffers=0 max-size-bytes=")
			sb.WriteString(strconv.Itoa(videoQueueBytes))
			sb.WriteString(" max-size-time=")
			sb.WriteString(strconv.FormatInt(queueNS, 10))
			sb.WriteString(" leaky=0 ! av1parse ! video/x-av1,stream-format=obu-stream,alignment=tu ! mux.video_0 ")
		}

	case probe.IsVP9():
		if conf.TranscodeVP9 {
			r.transcodeToH264(&sb, videoQueueBytes, queueNS)
		} else {
			sb.WriteString("d.video_0 ! queue max-size-buffers=0 max-size-bytes=")
			sb.WriteString(strconv.Itoa(videoQueueBytes))
			sb.WriteString(" max-size-time=")
			sb.WriteString(strconv.FormatInt(queueNS, 10))
			sb.WriteString(" leaky=0 ! vp9parse ! video/x-vp9,alignment=frame ! mux.video_0 ")
		}
	}

	sb.WriteString("d.audio_")
	sb.WriteString(strconv.Itoa(r.audioIndex))
	sb.WriteString(" ! queue max-size-buffers=0 max-size-bytes=")
	sb.WriteString(strconv.Itoa(audioQueueBytes))
	sb.WriteString(" max-size-time=")
	sb.WriteString(strconv.FormatInt(queueNS, 10))
	sb.WriteString(" leaky=0 ! decodebin ! audioconvert dithering=none noise-shaping=none ! audioresample quality=2 sinc-filter-mode=full ! audio/x-raw,format=F32LE,layout=interleaved,rate=48000,channels=2 ! avenc_aac bitrate=")
	sb.WriteString(strconv.Itoa(conf.AACBitrateKbps * 1000))
	sb.WriteString(" ! aacparse ! audio/mpeg,mpegversion=4,stream-format=raw,rate=48000,channels=2 ! mux.audio_0 ")

	sb.WriteString("mp4mux name=mux fragment-duration=")
	sb.WriteString(strconv.Itoa(conf.SegmentSeconds * 1000))
	sb.WriteString(" streamable=true ! appsink name=out emit-signals=false sync=false max-buffers=1")
	if conf.GSTVersion >= 1.28 {
		sb.WriteString(" leaky-type=none")
	} else {
		sb.WriteString(" drop=false")
	}
	sb.WriteString(" wait-on-eos=false")

	return sb.String()
}

func (r *gstRunner) writeSourceQueue(sb *strings.Builder, videoQueueBytes int, queueNS int64) {
	conf := r.task.Config
	sourceQueueBytes := int64(videoQueueBytes)
	if conf.TempFS {
		sourceQueueBytes = r.tempFSBlockBytes()
	}

	sb.WriteString("! queue2 use-buffering=false max-size-buffers=0 max-size-bytes=")
	sb.WriteString(strconv.FormatInt(sourceQueueBytes, 10))
	sb.WriteString(" max-size-time=")
	sb.WriteString(strconv.FormatInt(queueNS, 10))

	if !conf.TempFS {
		return
	}

	ringBlocks := int64(tempFSBaseBlocks + conf.TempFSRing)
	ringBytes := ringBlocks * sourceQueueBytes
	template := gstPath(queue2TempTemplate())

	sb.WriteString(" temp-template=\"")
	sb.WriteString(template)
	sb.WriteString("\" ring-buffer-max-size=")
	sb.WriteString(strconv.FormatInt(ringBytes, 10))
}

func (r *gstRunner) tempFSBlockBytes() int64 {
	probe := r.task.Probe
	durationSeconds := probe.DurationSeconds()
	if probe.FileSize > 0 && durationSeconds > 0 {
		blockBytes := int64(math.Ceil(float64(probe.FileSize) * tempFSBlockSeconds / float64(durationSeconds)))
		if blockBytes > 0 {
			return blockBytes
		}
	}
	return tempFSFallbackBlockBytes
}

func (r *gstRunner) transcodeToH264(sb *strings.Builder, maxQueueBytes int, queueNS int64) {
	conf := r.task.Config
	video := r.task.Probe.Video()

	frameRateNum := 0
	frameRateDen := 0
	if video != nil {
		frameRateNum = video.FrameRateNum
		frameRateDen = video.FrameRateDen
	}

	keyIntMax := 25 * conf.SegmentSeconds
	if frameRateNum > 0 && frameRateDen > 0 {
		keyIntMax = maxInt(1, int(math.Round(float64(frameRateNum*conf.SegmentSeconds)/float64(frameRateDen))))
	}

	sb.WriteString("d.video_0 ! queue max-size-buffers=0 max-size-bytes=")
	sb.WriteString(strconv.Itoa(maxQueueBytes))
	sb.WriteString(" max-size-time=")
	sb.WriteString(strconv.FormatInt(queueNS, 10))
	sb.WriteString(" leaky=0 ! decodebin ! videoconvert ! video/x-raw,format=I420 ! x264enc tune=zerolatency speed-preset=veryfast bitrate=")
	sb.WriteString(strconv.Itoa(conf.VideoBitrate))
	sb.WriteString(" key-int-max=")
	sb.WriteString(strconv.Itoa(keyIntMax))
	sb.WriteString(" bframes=0 byte-stream=false ! video/x-h264,profile=main,stream-format=avc,alignment=au ! h264parse config-interval=0 ! h264timestamper ! video/x-h264,profile=main,stream-format=avc,alignment=au ! mux.video_0 ")
}

func (r *gstRunner) Seek(seconds float64) bool {
	r.stopPipeline()

	r.reader.SeekReset(seconds)
	r.readySegment.index = -1
	r.readySegment.complete = false
	r.readySegment.segment = Segment{}

	if err := r.startPipeline(seconds); err != nil {
		r.freezeAtPosition(seconds)
		return false
	}

	r.frozen.Store(false)
	r.setPosition(seconds)
	r.positionSeekSeconds = seconds
	r.statePlaying = true
	return true
}

func (r *gstRunner) GetSegment(ctx context.Context, index int, audio int) (Segment, error) {
	if r.IsFrozen() {
		if !r.Seek(r.position()) {
			return Segment{}, ErrSegmentNotReady
		}
	} else if !r.statePlaying {
		r.statePlaying = true
		r.audioIndex = validAudioIndex(r.task.Probe, audio)
		startSeconds := 0.0
		if index > 0 {
			startSeconds = float64(index * r.task.Config.SegmentSeconds)
			r.reader.SeekReset(startSeconds)
			r.positionSeekSeconds = startSeconds
			r.setPosition(startSeconds)
		}
		if err := r.startPipeline(startSeconds); err != nil {
			r.freezeAtPosition(startSeconds)
			return Segment{}, err
		}
	}

	if r.readySegment.index == index && r.readySegment.complete {
		return r.readySegment.segment, nil
	}

	r.readySegment.index = -1
	r.readySegment.complete = false
	r.readySegment.segment = Segment{}

	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return Segment{}, ctx.Err()
		}

		sample := gstRuntime.appSinkTryPullSample(r.sink, uint64(100*time.Millisecond))
		if sample == 0 {
			if gstRuntime.appSinkIsEOS(r.sink) {
				completed, err := r.reader.TryProcessDeferred()
				if err != nil {
					r.freezeAtSegment(index)
					return Segment{}, err
				}
				if completed && r.readySegment.complete {
					return r.completeReadySegment(index), nil
				}

				completed, err = r.reader.TryFlushFinalSegment()
				if err != nil {
					r.freezeAtSegment(index)
					return Segment{}, err
				}
				if completed && r.readySegment.complete {
					return r.completeReadySegment(index), nil
				}

				if seg, ok := r.reader.TakePendingSegment(); ok {
					r.readySegment.segment = seg
					r.readySegment.complete = true
					return r.completeReadySegment(index), nil
				}

				r.freezeAtSegment(index)
				return Segment{}, ErrSegmentNotReady
			}
			continue
		}

		err := gstRuntime.withSampleBytes(sample, func(data []byte) error {
			if len(data) == 0 {
				return nil
			}
			return r.reader.Push(data)
		})
		gstRuntime.sampleUnref(sample)
		if err != nil {
			r.freezeAtSegment(index)
			return Segment{}, err
		}

		if r.readySegment.complete {
			return r.completeReadySegment(index), nil
		}
	}

	return Segment{}, ErrSegmentNotReady
}

func (r *gstRunner) completeReadySegment(index int) Segment {
	if index > 0 {
		r.readySegment.index = index
	} else {
		r.readySegment.index = 0
	}
	return r.readySegment.segment
}

func (r *gstRunner) freezeAtSegment(index int) {
	seconds := r.position()
	if index > 0 {
		seconds = float64(index * r.task.Config.SegmentSeconds)
	}

	r.freezeAtPosition(seconds)
}

func (r *gstRunner) freezeAtPosition(seconds float64) {
	r.stopPipeline()
	r.reader.SeekReset(seconds)
	r.readySegment.index = -1
	r.readySegment.complete = false
	r.readySegment.segment = Segment{}
	r.frozen.Store(true)
	r.setPosition(seconds)
	r.positionSeekSeconds = seconds
	r.statePlaying = false
}

func (r *gstRunner) startPipeline(seconds float64) error {
	pipeline, err := gstRuntime.parseLaunch(r.createPipelineArgs())
	if err != nil {
		return err
	}

	sink := gstRuntime.binGetByName(pipeline, "out")
	if sink == 0 {
		gstRuntime.elementSetState(pipeline, gstStateNull)
		gstRuntime.objectUnref(pipeline)
		return errors.New("appsink element is not available")
	}

	bus := gstRuntime.pipelineGetBus(pipeline)

	if seconds > 0 {
		if err := r.setPipelineState(pipeline, bus, gstStatePaused); err != nil {
			gstRuntime.elementSetState(pipeline, gstStateNull)
			gstRuntime.objectUnref(sink)
			gstRuntime.objectUnref(pipeline)
			gstRuntime.objectUnref(bus)
			return err
		}

		if !gstRuntime.elementSeekSimple(pipeline, gstFormatTime, gstSeekFlagFlush|gstSeekFlagKeyUnit|gstSeekFlagSnapAfter, int64(math.Round(seconds*1_000_000_000))) {
			gstRuntime.elementSetState(pipeline, gstStateNull)
			gstRuntime.objectUnref(sink)
			gstRuntime.objectUnref(pipeline)
			gstRuntime.objectUnref(bus)
			return errors.New("gstreamer seek failed")
		}
	}

	if err := r.setPipelineState(pipeline, bus, gstStatePlaying); err != nil {
		gstRuntime.elementSetState(pipeline, gstStateNull)
		gstRuntime.objectUnref(sink)
		gstRuntime.objectUnref(pipeline)
		gstRuntime.objectUnref(bus)
		return err
	}

	r.pipeline = pipeline
	r.bus = bus
	r.sink = sink
	return nil
}

func (r *gstRunner) setPipelineState(pipeline uintptr, bus uintptr, state int32) error {
	if ret := gstRuntime.elementSetState(pipeline, state); ret == gstStateChangeFailure {
		if err := gstRuntime.popBusError(bus, 0); err != nil {
			return err
		}
		return errors.New("gstreamer state change failed")
	}

	if ret := gstRuntime.elementGetState(pipeline, 5*time.Second); ret == gstStateChangeFailure {
		if err := gstRuntime.popBusError(bus, 0); err != nil {
			return err
		}
		return errors.New("gstreamer state wait failed")
	}

	return nil
}

func (r *gstRunner) stopPipeline() {
	if r.pipeline != 0 {
		_ = gstRuntime.elementSetState(r.pipeline, gstStateNull)
	}
	if r.sink != 0 {
		gstRuntime.objectUnref(r.sink)
		r.sink = 0
	}
	if r.bus != 0 {
		gstRuntime.objectUnref(r.bus)
		r.bus = 0
	}
	if r.pipeline != 0 {
		gstRuntime.objectUnref(r.pipeline)
		r.pipeline = 0
	}
}

func (r *gstRunner) Dispose() {
	r.stopPipeline()
	if r.reader != nil {
		r.reader.SeekReset(r.position())
	}
	r.readySegment.index = -1
	r.readySegment.complete = false
	r.readySegment.segment = Segment{}
	r.statePlaying = false
}

func (r *gstRunner) Frozen() {
	r.freezeAtPosition(r.position())
}

func (r *gstRunner) IsFrozen() bool {
	return r.frozen.Load()
}

func (r *gstRunner) setPosition(seconds float64) {
	r.positionSeconds.Store(math.Float64bits(seconds))
}

func (r *gstRunner) position() float64 {
	return math.Float64frombits(r.positionSeconds.Load())
}

func validAudioIndex(probe ProbeInfo, requested int) int {
	for _, track := range probe.Tracks {
		if track.Type == "audio" && track.Index == requested {
			return requested
		}
	}
	return 0
}
