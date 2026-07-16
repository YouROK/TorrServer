//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"context"
	"errors"
	"fmt"
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
	defaultAACChannels   = 2
	defaultAACSampleRate = 48000
	pipelineReadTimeout  = 20 * time.Second
	pipelineStateTimeout = 5 * time.Second
	gstPollInterval      = 100 * time.Millisecond
)

var aacEncoderRates = [...]int{7350, 8000, 11025, 12000, 16000, 22050, 24000, 32000, 44100, 48000, 64000, 88200, 96000}

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
	watchGeneration     atomic.Uint64

	reader *mp4BoxReader

	pipeline       uintptr
	bus            uintptr
	sink           uintptr
	subtitleSinks  map[int]uintptr
	subtitleStores map[int]*subtitleStore

	watchMu     sync.Mutex
	watchCancel chan struct{}
	watchDone   chan struct{}
	busErrorMu  sync.Mutex
	busError    error

	frozen atomic.Bool
}

func newPipelineRunner(task *Task, audio int) (pipelineRunner, error) {
	gstInitOnce.Do(func() {
		initGStreamerRuntime(task.Config)
	})
	if gstInitErr != nil {
		return nil, errors.Join(ErrPipelineUnavailable, gstInitErr)
	}
	if video := task.Probe.Video(); task.Config.HDRToSDR && video != nil && video.IsHDRVideo() && !gstElementAvailable("hdrtonemap") {
		return nil, errors.New("HDR tone mapping backend is not available")
	}

	runner := &gstRunner{
		task:       task,
		audioIndex: validAudioIndex(task.Probe, audio),
	}
	runner.ensureTransientState()
	runner.readySegment.index = -1
	return runner, nil
}

func (r *gstRunner) ensureTransientState() {
	if r.subtitleSinks == nil {
		r.subtitleSinks = make(map[int]uintptr)
	}
	if r.subtitleStores == nil {
		r.subtitleStores = make(map[int]*subtitleStore)
		if r.task.Config.Subtitles {
			for _, track := range r.task.Probe.Tracks {
				if supportedSubtitleTrack(track) {
					r.subtitleStores[track.Index] = newSubtitleStore()
				}
			}
		}
	}
	if r.reader != nil {
		return
	}

	segmentDiff := r.task.Config.SegmentDiff
	if videoIsTranscoded(r.task.Config, r.task.Probe) {
		segmentDiff = 0
	}
	r.reader = Mp4BoxReader(
		func(data []byte) {
			r.task.setInitMP4(data)
		},
		func(seg Segment) {
			r.acceptSegment(seg)
		},
		float64(r.task.Config.SegmentSeconds),
		segmentDiff,
		r.task.Cue != nil,
	)
}

func (r *gstRunner) releaseTransientState() {
	r.reader = nil
	r.subtitleStores = nil
	r.subtitleSinks = nil
	_ = r.takeBusError()
}

func gstElementAvailable(name string) bool {
	if gstRuntime == nil || name == "" {
		return false
	}
	element, err := gstRuntime.parseLaunch(name)
	if err != nil {
		return false
	}
	gstRuntime.objectUnref(element)
	return true
}

func initGStreamerRuntime(conf Config) {
	setupGStreamer(conf)
	gstInitStatus = componentStatus{Found: gstreamerLibraryFound(conf)}

	var err error
	gstRuntime, err = loadGST(conf)
	if err != nil {
		gstInitErr = err
		gstErrorf("runtime load failed: %v", err)
		return
	}
	gstInitStatus.Found = true

	if err = gstRuntime.init(); err != nil {
		gstInitErr = err
		gstErrorf("runtime initialization failed: %v", err)
		return
	}
	gstInitStatus.Available = true
	gstInitErr = nil
	gstDebugf("runtime initialized")
}

func setupGStreamer(_ Config) {
	_ = os.Setenv("GST_REGISTRY", filepath.Join(os.TempDir(), "torrserver-gstreamer-registry.bin"))
}

func setupGStreamerRoots(roots []string) {
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

	pluginRoots := append([]string(nil), roots...)
	if runtime.GOOS == "windows" {
		pluginRoots = appendUniquePath(pluginRoots, portableGSTRuntimeRoot())
		pluginRoots = appendUniquePath(pluginRoots, embeddedGSTRuntimeRoot())
	}
	if extraPlugins := existingPaths(gstExtraPluginCandidates(pluginRoots)); len(extraPlugins) > 0 {
		prependExistingEnvPaths("PATH", extraPlugins)
		prependEnvPaths("GST_PLUGIN_PATH", extraPlugins)
	}
	if gstPlugins := firstExistingPath(gstPluginCandidates(roots)); gstPlugins != "" {
		_ = os.Setenv("GST_PLUGIN_SYSTEM_PATH_1_0", gstPlugins)
	}

	if gstPluginScanner := firstExistingPath(gstPluginScannerCandidates(roots)); gstPluginScanner != "" {
		_ = os.Setenv("GST_PLUGIN_SCANNER", gstPluginScanner)
	}
}

func gstRuntimeRoots(conf Config) []string {
	var roots []string
	roots = appendAvailableGSTRoot(roots, conf.GSTPath)
	for _, root := range gstDefaultRuntimeRoots() {
		roots = appendAvailableGSTRoot(roots, root)
	}
	if runtime.GOOS == "windows" {
		if root := portableGSTRuntimeRoot(); root != "" {
			roots = appendAvailableGSTRoot(roots, root)
		}
		if len(roots) == 0 {
			root := embeddedGSTRuntimeRoot()
			roots = appendAvailableGSTRoot(roots, root)
		}
	}
	return roots
}

func gstDefaultRuntimeRoots() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			`C:\Program Files\gstreamer\1.0\mingw_x86_64`,
			`C:\gstreamer\1.0\mingw_x86_64`,
		}
	case "linux":
		return []string{
			"/usr",
			"/usr/local",
			"/opt/gstreamer",
			"/opt/gstreamer/1.0",
		}
	case "darwin":
		return []string{
			"/Library/Frameworks/GStreamer.framework/Versions/1.0",
			"/opt/homebrew",
			"/usr/local",
		}
	default:
		return nil
	}
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

func appendAvailableGSTRoot(paths []string, path string) []string {
	if path == "" || !gstRootHasBaseLibrary(path) {
		return paths
	}
	return appendUniquePath(paths, path)
}

func gstRootHasBaseLibrary(root string) bool {
	for _, candidate := range gstBaseLibraryCandidates(root) {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}

func gstBaseLibraryCandidates(root string) []string {
	switch runtime.GOOS {
	case "windows":
		return []string{filepath.Join(root, "bin", "libgstreamer-1.0-0.dll")}
	case "darwin":
		var candidates []string
		for _, dir := range gstLibraryDirCandidates([]string{root}) {
			candidates = append(candidates,
				filepath.Join(dir, "libgstreamer-1.0.0.dylib"),
				filepath.Join(dir, "libgstreamer-1.0.dylib"),
			)
		}
		return candidates
	default:
		var candidates []string
		for _, dir := range gstLibraryDirCandidates([]string{root}) {
			candidates = append(candidates, filepath.Join(dir, "libgstreamer-1.0.so.0"))
		}
		return candidates
	}
}

func appendUniquePath(paths []string, path string) []string {
	if path == "" {
		return paths
	}
	if containsPath(paths, path) {
		return paths
	}
	return append(paths, path)
}

func containsPath(paths []string, path string) bool {
	clean := filepath.Clean(path)
	for _, existing := range paths {
		if strings.EqualFold(filepath.Clean(existing), clean) {
			return true
		}
	}
	return false
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

func gstExtraPluginCandidates(roots []string) []string {
	var candidates []string
	for _, root := range roots {
		candidates = append(candidates, filepath.Join(root, "torrserver-plugins"))
	}
	return candidates
}

func gstPluginScannerCandidates(roots []string) []string {
	var candidates []string
	name := "gst-plugin-scanner"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	for _, root := range roots {
		candidates = append(candidates,
			filepath.Join(root, "libexec", "gstreamer-1.0", name),
			filepath.Join(root, "lib", "gstreamer-1.0", name),
			filepath.Join(root, "lib64", "gstreamer-1.0", name),
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

func (r *gstRunner) createPipelineArgs() string {
	conf := r.task.Config.normalized()
	probe := r.task.Probe
	gstVersion := effectiveGStreamerVersion(conf)

	var sb strings.Builder

	sb.WriteString("souphttpsrc ")
	sb.WriteString("location=\"")
	sb.WriteString(r.task.SourceURL)
	sb.WriteString("\" is-live=false keep-alive=true timeout=60 retries=5 ")
	if gstVersion.atLeast(1, 26) {
		sb.WriteString("retry-backoff-factor=0.5 retry-backoff-max=10 ")
	}
	if probe.IsAVIContainer() {
		sb.WriteString(" ! avidemux name=d ")
	} else {
		sb.WriteString(" ! matroskademux name=d ")
	}
	sb.WriteString("multiqueue name=mq use-buffering=false max-size-buffers=5 max-size-bytes=0 max-size-time=0 ")

	sb.WriteString("d.video_0 ! mq.sink_0 ")

	if videoIsTranscoded(conf, probe) {
		r.transcodeToH264(&sb)
	} else {
		switch {
		case probe.IsH264():
			sb.WriteString("mq.src_0 ! h264parse config-interval=0 ! h264timestamper name=video_timestamper ! video/x-h264,stream-format=avc,alignment=au ! mux.video_0 ")

		case probe.IsH265():
			sb.WriteString("mq.src_0 ! h265parse config-interval=0 ! h265timestamper name=video_timestamper ! video/x-h265,stream-format=hvc1,alignment=au ! mux.video_0 ")

		case probe.IsAV1():
			sb.WriteString("mq.src_0 ! av1parse ! video/x-av1,stream-format=obu-stream,alignment=tu ! mux.video_0 ")

		case probe.IsVP9():
			sb.WriteString("mq.src_0 ! vp9parse ! video/x-vp9,alignment=frame ! mux.video_0 ")
		}
	}

	if audioTrack := probe.AudioTrack(r.audioIndex); audioTrack != nil {
		sb.WriteString("d.audio_")
		sb.WriteString(strconv.Itoa(audioTrack.Index))
		sb.WriteString(" ! mq.sink_1 mq.src_1 ! ")
		if audioTrack.IsAACAudio() {
			sb.WriteString("aacparse ! audio/mpeg,mpegversion=4,stream-format=raw ! mux.audio_0 ")
		} else {
			aacChannels := effectiveAACChannels(conf, audioTrack)
			aacSampleRate := effectiveAACSampleRate(conf, audioTrack)
			aacBitrate := conf.AACBitrateKbps * 1000
			if aacChannels > 2 {
				aacBitrate *= 2
			}

			sb.WriteString("decodebin ! audioconvert dithering=none noise-shaping=none ! audioresample quality=2 sinc-filter-mode=full ! audio/x-raw,format=")
			sb.WriteString(aacRawFormat())
			sb.WriteString(",layout=interleaved,rate=")
			sb.WriteString(strconv.Itoa(aacSampleRate))
			sb.WriteString(",channels=")
			sb.WriteString(strconv.Itoa(aacChannels))
			if aacChannels == 6 {
				// Normalize 5.1(side) decoders to the browser-compatible AAC 5.1 layout.
				sb.WriteString(",channel-mask=(bitmask)0x000000000000003f")
			}
			sb.WriteString(" ! ")
			sb.WriteString(r.aacEncoder())
			sb.WriteString(" bitrate=")
			sb.WriteString(strconv.Itoa(aacBitrate))
			sb.WriteString(" ! aacparse ! audio/mpeg,mpegversion=4,stream-format=raw,rate=")
			sb.WriteString(strconv.Itoa(aacSampleRate))
			sb.WriteString(",channels=")
			sb.WriteString(strconv.Itoa(aacChannels))
			sb.WriteString(" ! mux.audio_0 ")
		}
	}
	r.writeSubtitleBranches(&sb, gstVersion)

	sb.WriteString("mp4mux name=mux fragment-mode=dash-or-mss fragment-duration=")
	if r.task.Cue != nil {
		sb.WriteString("1")
	} else {
		sb.WriteString(strconv.Itoa(conf.SegmentSeconds * 1000))
	}
	r.writeAppSink(&sb, gstVersion)

	return sb.String()
}

func (r *gstRunner) writeSubtitleBranches(sb *strings.Builder, gstVersion gstVersionInfo) {
	if !r.task.Config.Subtitles {
		return
	}
	for _, track := range r.task.Probe.Tracks {
		if !supportedSubtitleTrack(track) {
			continue
		}
		sb.WriteString("d.")
		sb.WriteString(track.PadName)
		sb.WriteString(" ! queue max-size-buffers=16 max-size-bytes=0 max-size-time=0 ! ")
		if track.Codec == "ass" || track.Codec == "ssa" {
			sb.WriteString("ssaparse ! ")
		}
		sb.WriteString("webvttenc ! appsink name=subs_")
		sb.WriteString(strconv.Itoa(track.Index))
		sb.WriteString(" emit-signals=false sync=false async=false max-buffers=16")
		if gstVersion.atLeast(1, 28) {
			sb.WriteString(" leaky-type=none")
		} else {
			sb.WriteString(" drop=false")
		}
		sb.WriteString(" wait-on-eos=false ")
	}
}

func (r *gstRunner) writeAppSink(sb *strings.Builder, gstVersion gstVersionInfo) {
	sb.WriteString(" streamable=true ! appsink name=out emit-signals=false sync=false max-buffers=1")
	if gstVersion.atLeast(1, 28) {
		sb.WriteString(" leaky-type=none")
	} else {
		sb.WriteString(" drop=false")
	}
	sb.WriteString(" wait-on-eos=false")
}

func effectiveGStreamerVersion(conf Config) gstVersionInfo {
	if gstRuntime != nil && gstRuntime.version.valid() {
		return gstRuntime.version
	}
	if conf.GSTVersion < minGSTVersion {
		conf.GSTVersion = minGSTVersion
	}

	major := uint32(conf.GSTVersion)
	minor := uint32(math.Round((conf.GSTVersion - float64(major)) * 100))
	if minor >= 100 {
		major += minor / 100
		minor %= 100
	}
	return gstVersionInfo{major: major, minor: minor}
}

func (r *gstRunner) aacEncoder() string {
	return "avenc_aac"
}

func aacRawFormat() string {
	return "F32LE"
}

func effectiveAACChannels(conf Config, track *TrackInfo) int {
	channels := conf.AACChannels
	if channels <= 0 && track != nil {
		channels = track.Channels
	}
	if channels <= 0 {
		channels = defaultAACChannels
	}
	return min(max(channels, 1), 8)
}

func effectiveAACSampleRate(conf Config, track *TrackInfo) int {
	rate := conf.AACSamplerate
	if rate <= 0 && track != nil {
		rate = track.Rate
	}
	if rate <= 0 {
		rate = defaultAACSampleRate
	}

	best := aacEncoderRates[0]
	bestDistance := absInt(rate - best)
	for _, candidate := range aacEncoderRates[1:] {
		if distance := absInt(rate - candidate); distance < bestDistance {
			best = candidate
			bestDistance = distance
		}
	}
	return best
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func (r *gstRunner) transcodeToH264(sb *strings.Builder) {
	conf := r.task.Config
	video := r.task.Probe.Video()
	toneMapHDR := conf.HDRToSDR && video != nil && video.IsHDRVideo()

	frameRateNum := 0
	frameRateDen := 0
	if video != nil {
		frameRateNum = video.FrameRateNum
		frameRateDen = video.FrameRateDen
	}

	keyIntMax := 25 * conf.SegmentSeconds
	if frameRateNum > 0 && frameRateDen > 0 {
		keyIntMax = int(math.Round(float64(frameRateNum*conf.SegmentSeconds) / float64(frameRateDen)))
		if keyIntMax < 1 {
			keyIntMax = 1
		}
	}

	sb.WriteString("mq.src_0 ! decodebin ! ")
	if toneMapHDR {
		transfer := "pq"
		if video.VideoTransfer == "hlg" {
			transfer = "hlg"
		}
		sb.WriteString("hdrtonemap transfer=")
		sb.WriteString(transfer)
		sb.WriteString(" use-opencl=")
		sb.WriteString(strconv.FormatBool(conf.UseGPU))
		sb.WriteString(" ! ")
	} else {
		// The selected encoder adds the converter it requires.
	}
	encoderPipeline := ""
	if conf.UseGPU && conf.HardwareAcceleration && video != nil {
		encoderPipeline = hardwareH264Pipeline(video.Width, video.Height, conf.VideoBitrate, keyIntMax)
	}
	if encoderPipeline == "" {
		speedPreset := "veryfast"
		if conf.X264Ultrafast {
			speedPreset = "ultrafast"
		}
		if !toneMapHDR {
			sb.WriteString("videoconvert ! ")
		}
		sb.WriteString("video/x-raw,format=I420 ! x264enc name=video_encoder tune=zerolatency speed-preset=")
		sb.WriteString(speedPreset)
		sb.WriteString(" bitrate=")
		sb.WriteString(strconv.Itoa(conf.VideoBitrate))
		sb.WriteString(" key-int-max=")
		sb.WriteString(strconv.Itoa(keyIntMax))
		sb.WriteString(" bframes=0 byte-stream=false ! video/x-h264,profile=main,stream-format=avc,alignment=au ! ")
	} else {
		sb.WriteString(encoderPipeline)
	}
	sb.WriteString("h264parse config-interval=0 ! h264timestamper name=video_timestamper ! video/x-h264,profile=main,stream-format=avc,alignment=au ! mux.video_0 ")
}

func videoIsTranscoded(conf Config, probe ProbeInfo) bool {
	if conf.HDRToSDR && probe.Video() != nil && probe.Video().IsHDRVideo() {
		return true
	}
	if conf.TranscodeAVI && probe.IsAVIContainer() {
		return true
	}
	switch {
	case probe.IsH264():
		return conf.TranscodeH264
	case probe.IsH265():
		return conf.TranscodeH265
	case probe.IsAV1():
		return conf.TranscodeAV1
	case probe.IsVP9():
		return conf.TranscodeVP9
	case probe.IsVP8():
		return conf.TranscodeVP8
	default:
		return false
	}
}

func (r *gstRunner) Seek(seconds float64) bool {
	r.ensureTransientState()
	r.discardReadySegment()
	wasFrozen := r.IsFrozen()
	reuse := r.pipeline != 0
	accurate := r.task.Cue != nil
	gstTaskDebugf(r.task, "seek requested=%.3fs reuse=%t accurate=%t", seconds, reuse, accurate)

	var actualSeconds float64
	var err error
	if reuse {
		actualSeconds, err = r.reusePipeline(seconds, accurate)
	} else {
		r.reader.SeekReset(seconds)
		actualSeconds, err = r.startPipeline(seconds)
	}
	if err != nil {
		gstTaskErrorf(r.task, "seek requested=%.3fs failed: %v", seconds, err)
		r.freezeAtPosition(seconds)
		return false
	}
	r.reader.SeekReset(actualSeconds)

	r.frozen.Store(false)
	r.setPosition(actualSeconds)
	r.positionSeekSeconds = actualSeconds
	r.statePlaying = true
	gstTaskDebugf(r.task, "seek completed requested=%.3fs actual=%.3fs reuse=%t", seconds, actualSeconds, reuse)
	if wasFrozen {
		gstTaskDebugf(r.task, "pipeline thawed position=%.3fs", actualSeconds)
	}
	return true
}

func (r *gstRunner) reusePipeline(seconds float64, accurate bool) (float64, error) {
	if r.pipeline == 0 || r.bus == 0 || r.sink == 0 {
		return 0, errors.New("pipeline cannot be reused")
	}
	r.stopBusWatch()
	if err := r.setPipelineState(r.pipeline, r.bus, gstStatePaused); err != nil {
		return 0, fmt.Errorf("pause pipeline before seek: %w", err)
	}

	mux := gstRuntime.binGetByName(r.pipeline, "mux")
	if mux == 0 {
		return 0, errors.New("mp4 mux is not available for seek reset")
	}
	defer gstRuntime.objectUnref(mux)

	videoTimestamper := gstRuntime.binGetByName(r.pipeline, "video_timestamper")
	if videoTimestamper != 0 {
		defer gstRuntime.objectUnref(videoTimestamper)
	}
	videoEncoder := gstRuntime.binGetByName(r.pipeline, "video_encoder")
	if videoEncoder != 0 {
		defer gstRuntime.objectUnref(videoEncoder)
	}

	for _, element := range []uintptr{r.sink, mux, videoTimestamper, videoEncoder} {
		if element != 0 && gstRuntime.elementSetState(element, gstStateReady) == gstStateChangeFailure {
			return 0, errors.New("reset pipeline child before seek")
		}
	}
	for _, element := range []uintptr{videoEncoder, videoTimestamper, mux, r.sink} {
		if element != 0 && gstRuntime.elementSetState(element, gstStatePaused) == gstStateChangeFailure {
			return 0, errors.New("pause pipeline child after seek reset")
		}
	}

	r.reader.SeekReset(seconds)
	r.positionSeekSeconds = seconds
	r.setPosition(seconds)

	flags := gstSeekFlagFlush | gstSeekFlagKeyUnit | gstSeekFlagSnapAfter
	if accurate {
		flags |= gstSeekFlagAccurate
	}
	seekNS := int64(math.Round(seconds * 1_000_000_000))
	if err := sendVideoSeekEvent(r.pipeline, flags, seekNS); err != nil {
		return 0, fmt.Errorf("gstreamer seek failed while reusing pipeline: %w", err)
	}

	waitResult := gstRuntime.elementGetState(r.pipeline, pipelineStateTimeout)
	if waitResult != gstStateChangeSuccess && waitResult != gstStateChangeNoPreroll {
		if err := gstRuntime.popBusError(r.bus, 0); err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("gstreamer seek state=%d while reusing pipeline", waitResult)
	}

	actualSeconds := seconds
	if positionNS, ok := gstRuntime.elementQueryPosition(r.pipeline); ok && positionNS >= 0 {
		actualSeconds = float64(positionNS) / 1_000_000_000
	}
	if err := r.setPipelineState(r.pipeline, r.bus, gstStatePlaying); err != nil {
		return 0, fmt.Errorf("resume pipeline after seek: %w", err)
	}
	r.startBusWatch()
	return actualSeconds, nil
}

func sendVideoSeekEvent(pipeline uintptr, flags int32, positionNS int64) error {
	multiqueue := gstRuntime.binGetByName(pipeline, "mq")
	if multiqueue == 0 {
		return errors.New("multiqueue is not available")
	}
	defer gstRuntime.objectUnref(multiqueue)

	videoPad := gstRuntime.elementGetStaticPad(multiqueue, "src_0")
	if videoPad == 0 {
		return errors.New("multiqueue video pad is not available")
	}
	defer gstRuntime.objectUnref(videoPad)

	if !gstRuntime.sendTimeSeekEvent(videoPad, flags, positionNS) {
		return errors.New("video seek event returned false")
	}
	return nil
}

func (r *gstRunner) EnsureInit(ctx context.Context, audio int, startIndex int) error {
	if startIndex < 0 {
		startIndex = 0
	}
	r.ensureTransientState()

	startSeconds := r.segmentStartSeconds(startIndex)

	if r.IsFrozen() {
		if !r.Seek(startSeconds) {
			return ErrSegmentNotReady
		}
	} else if !r.statePlaying {
		r.statePlaying = true
		r.audioIndex = validAudioIndex(r.task.Probe, audio)
		if startSeconds > 0 {
			r.reader.SeekReset(startSeconds)
			r.positionSeekSeconds = startSeconds
			r.setPosition(startSeconds)
		}
		actualSeconds, err := r.startPipeline(startSeconds)
		if err != nil {
			r.freezeAtPosition(startSeconds)
			return err
		}
		if startSeconds > 0 {
			r.reader.SeekReset(actualSeconds)
			r.positionSeekSeconds = actualSeconds
			r.setPosition(actualSeconds)
		}
	} else if startIndex > 0 && math.Abs(r.position()-startSeconds) > 0.001 {
		if !r.Seek(startSeconds) {
			return ErrSegmentNotReady
		}
	}

	if r.task.hasInitMP4() {
		if r.readySegment.complete {
			r.completeReadySegment(startIndex)
		}
		return nil
	}

	deadline := time.Now().Add(pipelineReadTimeout)
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return err
		}

		eos, err := r.pullOutputSample()
		if err != nil {
			r.freezeAtSegment(startIndex)
			return err
		}
		if eos {
			r.freezeAtSegment(startIndex)
			return ErrSegmentNotReady
		}

		if r.task.hasInitMP4() {
			if r.readySegment.complete {
				r.completeReadySegment(startIndex)
			}
			return nil
		}
	}

	if err := r.pollPipelineError(); err != nil {
		r.freezeAtSegment(startIndex)
		return err
	}

	return ErrSegmentNotReady
}

func (r *gstRunner) GetSegment(ctx context.Context, index int, audio int) (Segment, error) {
	r.ensureTransientState()
	if r.IsFrozen() {
		if !r.Seek(r.position()) {
			return Segment{}, ErrSegmentNotReady
		}
	} else if !r.statePlaying {
		r.statePlaying = true
		r.audioIndex = validAudioIndex(r.task.Probe, audio)
		startSeconds := r.segmentStartSeconds(index)
		if startSeconds > 0 {
			r.reader.SeekReset(startSeconds)
			r.positionSeekSeconds = startSeconds
			r.setPosition(startSeconds)
		}
		actualSeconds, err := r.startPipeline(startSeconds)
		if err != nil {
			r.freezeAtPosition(startSeconds)
			return Segment{}, err
		}
		if startSeconds > 0 {
			r.reader.SeekReset(actualSeconds)
			r.positionSeekSeconds = actualSeconds
			r.setPosition(actualSeconds)
		}
	}

	if r.readySegment.index == index && r.readySegment.complete {
		return r.readySegment.segment, nil
	}

	r.discardReadySegment()
	if cue, ok := r.task.Cue.Segment(index); ok {
		if err := r.reader.SetTargetSegment(cue.StartNS, cue.EndNS, max(r.task.Cue.TimestampScaleNS, 1)); err != nil {
			r.freezeAtSegment(index)
			return Segment{}, err
		}
	} else if r.task.Cue != nil {
		return Segment{}, ErrEndOfStreamExhausted
	}

	deadline := time.Now().Add(pipelineReadTimeout)
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return Segment{}, err
		}

		eos, err := r.pullOutputSample()
		if err != nil {
			r.freezeAtSegment(index)
			return Segment{}, err
		}
		if eos {
			return r.finishEndOfStream(index)
		}

		if r.readySegment.complete {
			return r.completeReadySegment(index), nil
		}
	}

	if err := r.pollPipelineError(); err != nil {
		r.freezeAtSegment(index)
		return Segment{}, err
	}

	return Segment{}, ErrSegmentNotReady
}

func (r *gstRunner) pullOutputSample() (bool, error) {
	if gstRuntime == nil || r.sink == 0 || r.reader == nil {
		return false, errors.New("gstreamer output is not available")
	}
	if err := r.pollPipelineError(); err != nil {
		return false, err
	}
	if err := r.drainSubtitles(); err != nil {
		return false, err
	}

	sample := gstRuntime.appSinkTryPullSample(r.sink, uint64(gstPollInterval))
	if sample == 0 {
		if err := r.pollPipelineError(); err != nil {
			return false, err
		}
		return gstRuntime.appSinkIsEOS(r.sink), nil
	}
	defer gstRuntime.sampleUnref(sample)

	if err := gstRuntime.withSampleBytes(sample, func(data []byte) error {
		if len(data) == 0 {
			return nil
		}
		return r.reader.Push(data)
	}); err != nil {
		return false, fmt.Errorf("mp4 parser: %w", err)
	}
	if err := r.pollPipelineError(); err != nil {
		return false, err
	}
	if err := r.drainSubtitles(); err != nil {
		return false, err
	}
	return false, nil
}

func (r *gstRunner) pollPipelineError() error {
	if err := r.takeBusError(); err != nil {
		return err
	}
	if r.bus == 0 || gstRuntime == nil {
		return nil
	}
	if r.busWatchActive() {
		return nil
	}
	return gstRuntime.popBusError(r.bus, 0)
}

func (r *gstRunner) startBusWatch() {
	r.stopBusWatch()
	_ = r.takeBusError()
	if r.bus == 0 || gstRuntime == nil || gstRuntime.gstBusTimedPopFiltered == nil {
		return
	}

	cancel := make(chan struct{})
	done := make(chan struct{})
	generation := r.watchGeneration.Add(1)
	bus := r.bus

	r.watchMu.Lock()
	r.watchCancel = cancel
	r.watchDone = done
	r.watchMu.Unlock()

	go r.watchBus(bus, generation, cancel, done)
}

func (r *gstRunner) watchBus(bus uintptr, generation uint64, cancel <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	for {
		select {
		case <-cancel:
			return
		default:
		}

		if err := gstRuntime.popBusError(bus, gstPollInterval); err != nil {
			gstTaskErrorf(r.task, "pipeline bus failed at %.3fs: %v", r.position(), err)
			r.storeBusError(err)
			go r.freezeBusGeneration(generation)
			return
		}
	}
}

func (r *gstRunner) freezeBusGeneration(generation uint64) {
	if r.watchGeneration.Load() != generation || r.task == nil {
		return
	}

	r.task.mu.Lock()
	defer r.task.mu.Unlock()
	if r.watchGeneration.Load() != generation || r.task.disposed.Load() || r.task.runner != r {
		return
	}
	r.freezeAtPosition(r.position())
}

func (r *gstRunner) stopBusWatch() {
	r.watchMu.Lock()
	cancel := r.watchCancel
	done := r.watchDone
	r.watchCancel = nil
	r.watchDone = nil
	if cancel != nil {
		close(cancel)
	}
	r.watchGeneration.Add(1)
	r.watchMu.Unlock()

	if done != nil {
		<-done
	}
}

func (r *gstRunner) busWatchActive() bool {
	r.watchMu.Lock()
	defer r.watchMu.Unlock()
	return r.watchDone != nil
}

func (r *gstRunner) storeBusError(err error) {
	if err == nil {
		return
	}
	r.busErrorMu.Lock()
	r.busError = err
	r.busErrorMu.Unlock()
}

func (r *gstRunner) takeBusError() error {
	r.busErrorMu.Lock()
	defer r.busErrorMu.Unlock()
	err := r.busError
	r.busError = nil
	return err
}

func (r *gstRunner) drainEndOfStream(index int) (Segment, error) {
	if r.reader == nil {
		return Segment{}, ErrSegmentNotReady
	}

	completed, err := r.reader.TryProcessDeferred()
	if err != nil {
		if len(r.reader.video) > 0 && !r.reader.video[0].startsWithSync {
			return Segment{}, r.reader.undecodableEOSRemainderError()
		}
		return Segment{}, err
	}
	if completed {
		if !r.readySegment.complete {
			return Segment{}, errors.New("mp4 reader completed a segment without onSegment callback")
		}
		return r.completeReadySegment(index), nil
	}

	completed, err = r.reader.TryBuildEndOfStreamRemainder()
	if err != nil {
		return Segment{}, err
	}
	if completed {
		if !r.readySegment.complete {
			return Segment{}, errors.New("mp4 reader completed EOS remainder without onSegment callback")
		}
		return r.completeReadySegment(index), nil
	}

	if err := r.reader.EndOfStreamError(); err != nil {
		return Segment{}, err
	}

	return Segment{}, ErrEndOfStreamExhausted
}

func (r *gstRunner) finishEndOfStream(index int) (Segment, error) {
	seg, err := r.drainEndOfStream(index)
	if err != nil && !errors.Is(err, ErrEndOfStreamExhausted) {
		gstTaskErrorf(r.task, "mp4 EOS drain failed for segment=%d: %v", index, err)
		r.freezeAtSegment(index)
	}
	return seg, err
}

func (r *gstRunner) completeReadySegment(index int) Segment {
	if index > 0 {
		r.readySegment.index = index
	} else {
		r.readySegment.index = 0
	}
	return r.readySegment.segment
}

func (r *gstRunner) acceptSegment(seg Segment) {
	r.readySegment.segment = seg
	r.readySegment.complete = true
	if seg.EndSeconds >= seg.StartSeconds {
		r.setPosition(seg.EndSeconds)
	} else if seg.StartSeconds >= 0 {
		r.setPosition(seg.StartSeconds)
	}
}

func (r *gstRunner) discardReadySegment() {
	hadReady := r.readySegment.complete

	r.readySegment.index = -1
	r.readySegment.complete = false
	r.readySegment.segment = Segment{}

	if hadReady && r.reader != nil {
		r.reader.ReleaseSegment()
	}
}

func (r *gstRunner) freezeAtSegment(index int) {
	seconds := r.position()
	if index >= 0 {
		seconds = r.segmentStartSeconds(index)
	}

	r.freezeAtPosition(seconds)
}

func (r *gstRunner) segmentStartSeconds(index int) float64 {
	if cue, ok := r.task.Cue.Segment(index); ok {
		return float64(cue.StartNS) / 1_000_000_000
	}
	if index > 0 {
		return float64(index * r.task.Config.SegmentSeconds)
	}
	return 0
}

func (r *gstRunner) freezeAtPosition(seconds float64) {
	wasFrozen := r.frozen.Load()
	hadPipeline := r.pipeline != 0
	r.stopPipeline()
	r.discardReadySegment()
	r.releaseTransientState()
	if r.task != nil {
		r.task.clearInitMP4()
	}
	r.frozen.Store(true)
	r.setPosition(seconds)
	r.positionSeekSeconds = seconds
	r.statePlaying = false
	if !wasFrozen || hadPipeline {
		gstTaskDebugf(r.task, "pipeline frozen position=%.3fs resources_released=true", seconds)
	}
}

func (r *gstRunner) startPipeline(seconds float64) (float64, error) {
	r.ensureTransientState()
	gstTaskDebugf(r.task, "pipeline start requested=%.3fs audio=%d", seconds, r.audioIndex)
	pipeline, err := gstRuntime.parseLaunch(r.createPipelineArgs())
	if err != nil {
		return 0, err
	}

	sink := gstRuntime.binGetByName(pipeline, "out")
	if sink == 0 {
		gstRuntime.elementSetState(pipeline, gstStateNull)
		gstRuntime.objectUnref(pipeline)
		return 0, errors.New("appsink element is not available")
	}

	bus := gstRuntime.pipelineGetBus(pipeline)
	if bus == 0 {
		gstRuntime.elementSetState(pipeline, gstStateNull)
		gstRuntime.objectUnref(sink)
		gstRuntime.objectUnref(pipeline)
		return 0, errors.New("gstreamer bus is not available")
	}
	subtitleSinks := make(map[int]uintptr)
	for index := range r.subtitleStores {
		if subSink := gstRuntime.binGetByName(pipeline, "subs_"+strconv.Itoa(index)); subSink != 0 {
			subtitleSinks[index] = subSink
		}
	}
	actualStartSeconds := seconds

	cleanup := func() {
		gstRuntime.elementSetState(pipeline, gstStateNull)
		gstRuntime.objectUnref(sink)
		gstRuntime.objectUnref(pipeline)
		gstRuntime.objectUnref(bus)
		for _, subSink := range subtitleSinks {
			gstRuntime.objectUnref(subSink)
		}
	}

	if seconds > 0 {
		if err := r.setPipelineState(pipeline, bus, gstStatePaused); err != nil {
			cleanup()
			return 0, err
		}

		flags := gstSeekFlagFlush | gstSeekFlagKeyUnit | gstSeekFlagSnapAfter
		if r.task.Cue != nil {
			flags |= gstSeekFlagAccurate
		}
		if err := sendVideoSeekEvent(pipeline, flags, int64(math.Round(seconds*1_000_000_000))); err != nil {
			cleanup()
			return 0, fmt.Errorf("gstreamer seek failed: %w", err)
		}

		waitResult := gstRuntime.elementGetState(pipeline, pipelineStateTimeout)
		switch waitResult {
		case gstStateChangeSuccess, gstStateChangeNoPreroll:
		case gstStateChangeAsync:
			if err := gstRuntime.popBusError(bus, 0); err != nil {
				cleanup()
				return 0, err
			}
			cleanup()
			return 0, fmt.Errorf("gstreamer seek to %.3fs timed out", seconds)
		case gstStateChangeFailure:
			if err := gstRuntime.popBusError(bus, 0); err != nil {
				cleanup()
				return 0, err
			}
			cleanup()
			return 0, fmt.Errorf("gstreamer seek to %.3fs failed", seconds)
		default:
			cleanup()
			return 0, fmt.Errorf("unexpected GstStateChangeReturn=%d after seek", waitResult)
		}

		if positionNS, ok := gstRuntime.elementQueryPosition(pipeline); ok {
			actualStartSeconds = float64(positionNS) / 1_000_000_000.0
		} else {
			gstTaskDebugf(r.task, "position query unavailable after seek; using requested=%.3fs", seconds)
		}
	}

	if err := r.setPipelineState(pipeline, bus, gstStatePlaying); err != nil {
		cleanup()
		return 0, err
	}

	r.pipeline = pipeline
	r.bus = bus
	r.sink = sink
	r.subtitleSinks = subtitleSinks
	r.startBusWatch()
	gstTaskDebugf(r.task, "pipeline started requested=%.3fs actual=%.3fs audio=%d", seconds, actualStartSeconds, r.audioIndex)
	return actualStartSeconds, nil
}

func (r *gstRunner) setPipelineState(pipeline uintptr, bus uintptr, state int32) error {
	setResult := gstRuntime.elementSetState(pipeline, state)
	if setResult == gstStateChangeFailure {
		if err := gstRuntime.popBusError(bus, 0); err != nil {
			return err
		}
		return fmt.Errorf("gstreamer failed to request state change to %d", state)
	}

	waitResult := gstRuntime.elementGetState(pipeline, pipelineStateTimeout)
	switch waitResult {
	case gstStateChangeSuccess, gstStateChangeNoPreroll:
		return nil

	case gstStateChangeAsync:
		if err := gstRuntime.popBusError(bus, 0); err != nil {
			return err
		}
		return fmt.Errorf("gstreamer state change to %d timed out", state)

	case gstStateChangeFailure:
		if err := gstRuntime.popBusError(bus, 0); err != nil {
			return err
		}
		return fmt.Errorf("gstreamer state change to %d failed", state)

	default:
		return fmt.Errorf("unexpected GstStateChangeReturn=%d for state=%d", waitResult, state)
	}
}

func (r *gstRunner) stopPipeline() {
	r.stopBusWatch()
	if r.pipeline != 0 {
		_ = gstRuntime.elementSetState(r.pipeline, gstStateNull)
	}
	if r.sink != 0 {
		gstRuntime.objectUnref(r.sink)
		r.sink = 0
	}
	for index, sink := range r.subtitleSinks {
		if sink != 0 {
			gstRuntime.objectUnref(sink)
		}
		delete(r.subtitleSinks, index)
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
	gstTaskDebugf(r.task, "pipeline dispose position=%.3fs", r.position())
	r.stopPipeline()
	r.discardReadySegment()
	r.releaseTransientState()
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
	fallback := -1

	for _, track := range probe.Tracks {
		if track.Type != "audio" {
			continue
		}
		if fallback < 0 {
			fallback = track.Index
		}
		if track.Index == requested {
			return requested
		}
	}

	return fallback
}
