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
	tempFSBlockSeconds       = 30
	tempFSBaseBlocks         = 3
	tempFSFallbackBlockBytes = 32 * 1024 * 1024
	appSinkMaxBytes          = 136 * 1024 * 1024
	defaultAACChannels       = 2
	defaultAACSampleRate     = 48000
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
	subtitleSinks map[int]uintptr
	probes   *gstPipelineProbes

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
	}
	runner.reader = Mp4BoxReader(
		func(data []byte) {
			task.setInitMP4(data)
		},
		func(seg Segment) {
			runner.acceptSegment(seg)
		},
		float64(task.Config.SegmentSeconds),
	)
	runner.readySegment.index = -1
	return runner, nil
}

func initGStreamerRuntime(conf Config) {
	setupGStreamer(conf)
	gstInitStatus = componentStatus{Found: gstreamerLibraryFound(conf)}

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

	if gstPlugins := firstExistingPath(gstPluginCandidates(roots)); gstPlugins != "" {
		_ = os.Setenv("GST_PLUGIN_PATH", gstPlugins)
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
		if root := embeddedGSTRuntimeRoot(); root != "" {
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
	started := time.Now()
	conf := r.task.Config.normalized()
	probe := r.task.Probe
	gstVersion := effectiveGStreamerVersion(conf)
	audioTrack := probe.AudioTrack(r.audioIndex)
	video := probe.Video()

	var sb strings.Builder

	sb.WriteString("souphttpsrc ")
	sb.WriteString("location=\"")
	sb.WriteString(r.task.SourceURL)
	sb.WriteString("\" is-live=false keep-alive=true timeout=60 retries=5 ")
	if gstVersion.atLeast(1, 26) {
		sb.WriteString("retry-backoff-factor=0.5 retry-backoff-max=10 ")
	}
	r.writeSourceQueue(&sb)
	sb.WriteString(" ! matroskademux name=d multiqueue name=mq use-buffering=false max-size-buffers=5 ")

	sb.WriteString("d.video_0 ! mq.sink_0 ")

	switch {
	case probe.IsH264():
		if conf.TranscodeH264 {
			r.transcodeToH264(&sb)
		} else {
			sb.WriteString("mq.src_0 ! h264parse config-interval=0 ! h264timestamper ! video/x-h264,stream-format=avc,alignment=au ! mux.video_0 ")
		}

	case probe.IsH265():
		if conf.TranscodeH265 {
			r.transcodeToH264(&sb)
		} else {
			sb.WriteString("mq.src_0 ! h265parse config-interval=0 ! h265timestamper ! video/x-h265,stream-format=hvc1,alignment=au ! mux.video_0 ")
		}

	case probe.IsAV1():
		if conf.TranscodeAV1 {
			r.transcodeToH264(&sb)
		} else {
			sb.WriteString("mq.src_0 ! av1parse ! video/x-av1,stream-format=obu-stream,alignment=tu ! mux.video_0 ")
		}

	case probe.IsVP9():
		if conf.TranscodeVP9 {
			r.transcodeToH264(&sb)
		} else {
			sb.WriteString("mq.src_0 ! vp9parse ! video/x-vp9,alignment=frame ! mux.video_0 ")
		}
	}

	if audioTrack != nil {
		sb.WriteString("d.audio_")
		sb.WriteString(strconv.Itoa(audioTrack.Index))
		sb.WriteString(" ! mq.sink_1 mq.src_1 ! ")
		if audioTrack.IsAACAudio() {
			sb.WriteString("aacparse ! audio/mpeg,mpegversion=4,stream-format=raw ! mux.audio_0 ")
		} else {
			aacEncoder := r.aacEncoder()
			aacChannels := effectiveAACChannels(conf, audioTrack)
			aacSampleRate := effectiveAACSampleRate(conf, audioTrack)

			sb.WriteString("decodebin ! audioconvert dithering=none noise-shaping=none ! ")
			sb.WriteString("audioresample quality=2 sinc-filter-mode=full ! audio/x-raw,format=")
			sb.WriteString(aacRawFormat())
			sb.WriteString(",layout=interleaved,rate=")
			sb.WriteString(strconv.Itoa(aacSampleRate))
			sb.WriteString(",channels=")
			sb.WriteString(strconv.Itoa(aacChannels))
			sb.WriteString(" ! ")
			sb.WriteString(aacEncoder)
			sb.WriteString(" bitrate=")
			sb.WriteString(strconv.Itoa(conf.AACBitrateKbps * 1000))
			sb.WriteString(" ! aacparse ! audio/mpeg,mpegversion=4,stream-format=raw,rate=")
			sb.WriteString(strconv.Itoa(aacSampleRate))
			sb.WriteString(",channels=")
			sb.WriteString(strconv.Itoa(aacChannels))
			sb.WriteString(" ! mux.audio_0 ")
		}
	}

	r.writeSubtitleBranches(&sb)

	sb.WriteString("mp4mux name=mux fragment-duration=")
	sb.WriteString(strconv.Itoa(conf.SegmentSeconds * 1000))
	r.writeAppSink(&sb, conf, gstVersion)

	args := sb.String()
	videoCodec := ""
	if video != nil {
		videoCodec = firstNonEmpty(video.Codec, video.CapsName)
	}
	audioCodec := ""
	audioChannels := 0
	audioRate := 0
	if audioTrack != nil {
		audioCodec = firstNonEmpty(audioTrack.Codec, audioTrack.CapsName)
		audioChannels = audioTrack.Channels
		audioRate = audioTrack.Rate
	}
	gstDebugf("create pipeline task=%s file=%s audio=%d selectedAudio=%d videoCodec=%q audioCodec=%q audioChannels=%d audioRate=%d segmentSeconds=%d appsinkBuffers=%d tempfs=%t tempfsRing=%d source=%s transcodeVideo=%t aacChannels=%d aacRate=%d aacBitrate=%d duration=%s",
		r.task.ID, r.task.FileID, r.audioIndex, r.audioIndex, videoCodec, audioCodec, audioChannels, audioRate, conf.SegmentSeconds, conf.AppSinkBuffers, conf.TempFS, conf.TempFSRing, conf.Source, r.transcodesVideo(conf, probe), effectiveAACChannels(conf, audioTrack), effectiveAACSampleRate(conf, audioTrack), conf.AACBitrateKbps, time.Since(started))
	gstDebugf("create pipeline args task=%s file=%s audio=%d pipeline=%s", r.task.ID, r.task.FileID, r.audioIndex, args)
	return args
}

func (r *gstRunner) transcodesVideo(conf Config, probe ProbeInfo) bool {
	switch {
	case probe.IsH264():
		return conf.TranscodeH264
	case probe.IsH265():
		return conf.TranscodeH265
	case probe.IsAV1():
		return conf.TranscodeAV1
	case probe.IsVP9():
		return conf.TranscodeVP9
	default:
		return false
	}
}

func (r *gstRunner) writeAppSink(sb *strings.Builder, conf Config, gstVersion gstVersionInfo) {
	buffers := conf.normalized().AppSinkBuffers
	if buffers <= 1 {
		buffers = 1
	}

	sb.WriteString(" streamable=true ! appsink name=out emit-signals=false sync=false max-buffers=")
	sb.WriteString(strconv.Itoa(buffers))
	if gstVersion.atLeast(1, 24) && buffers > 1 {
		sb.WriteString(" max-bytes=")
		sb.WriteString(strconv.Itoa(appSinkMaxBytes))
	}
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
	if conf.AACChannels > 0 {
		return conf.AACChannels
	}
	if track != nil && track.Channels > 0 {
		return track.Channels
	}
	return defaultAACChannels
}

func effectiveAACSampleRate(conf Config, track *TrackInfo) int {
	if conf.AACSamplerate > 0 {
		return conf.AACSamplerate
	}
	if track != nil && track.Rate > 0 {
		return track.Rate
	}
	return defaultAACSampleRate
}

func (r *gstRunner) writeSourceQueue(sb *strings.Builder) {
	conf := r.task.Config

	if !conf.TempFS {
		return
	}

	ringBlocks := int64(tempFSBaseBlocks + conf.TempFSRing)
	blockBytes := r.tempFSBlockBytes()
	ringBytes := ringBlocks*blockBytes + 1024*1024
	template := gstPath(queue2TempTemplate())

	sb.WriteString(" ! queue2 use-buffering=false temp-template=\"")
	sb.WriteString(template)
	sb.WriteString("\" temp-remove=true ring-buffer-max-size=")
	sb.WriteString(strconv.FormatInt(ringBytes, 10))
	sb.WriteString(" max-size-bytes=")
	sb.WriteString(strconv.FormatInt(blockBytes, 10))
	sb.WriteString(" max-size-buffers=0 max-size-time=0")
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

func (r *gstRunner) transcodeToH264(sb *strings.Builder) {
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
		keyIntMax = int(math.Round(float64(frameRateNum*conf.SegmentSeconds) / float64(frameRateDen)))
		if keyIntMax < 1 {
			keyIntMax = 1
		}
	}

	sb.WriteString("mq.src_0 ! decodebin ! videoconvert ! video/x-raw,format=I420 ! x264enc tune=zerolatency speed-preset=veryfast bitrate=")
	sb.WriteString(strconv.Itoa(conf.VideoBitrate))
	sb.WriteString(" key-int-max=")
	sb.WriteString(strconv.Itoa(keyIntMax))
	sb.WriteString(" bframes=0 byte-stream=false ! video/x-h264,profile=main,stream-format=avc,alignment=au ! h264parse config-interval=0 ! h264timestamper ! video/x-h264,profile=main,stream-format=avc,alignment=au ! mux.video_0 ")
}

func (r *gstRunner) Seek(seconds float64) error {
	started := time.Now()
	oldPosition := r.position()
	gstDebugf("Seek start task=%s file=%s audio=%d requested=%.3f oldPosition=%.3f", r.task.ID, r.task.FileID, r.audioIndex, seconds, oldPosition)
	gstDebugf("Seek stopping old pipeline task=%s file=%s audio=%d requested=%.3f", r.task.ID, r.task.FileID, r.audioIndex, seconds)
	r.stopPipeline()

	r.discardReadySegment()
	r.reader.SeekReset(seconds)

	actualSeconds, err := r.startPipeline(seconds)
	if err != nil {
		r.freezeAtPosition(seconds)
		err = fmt.Errorf("seek to %.3fs: %w", seconds, err)
		gstErrorf("Seek failed task=%s file=%s audio=%d requested=%.3f actual=%.3f err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, seconds, actualSeconds, err, time.Since(started))
		return err
	}
	r.reader.SeekReset(actualSeconds)

	r.frozen.Store(false)
	r.setPosition(actualSeconds)
	r.positionSeekSeconds = actualSeconds
	r.statePlaying = true
	gstDebugf("Seek completed task=%s file=%s audio=%d requested=%.3f actual=%.3f delta=%.3f duration=%s", r.task.ID, r.task.FileID, r.audioIndex, seconds, actualSeconds, actualSeconds-seconds, time.Since(started))
	return nil
}

func (r *gstRunner) EnsureInit(ctx context.Context, audio int, startIndex int) error {
	if startIndex < 0 {
		startIndex = 0
	}

	startSeconds := 0.0
	if startIndex > 0 {
		startSeconds = float64(startIndex * r.task.Config.SegmentSeconds)
	}

	if r.IsFrozen() {
		if err := r.Seek(startSeconds); err != nil {
			return err
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
		if err := r.Seek(startSeconds); err != nil {
			return err
		}
	}

	r.drainSubtitleSinks()
	if r.task.hasInitMP4() {
		if r.readySegment.complete {
			r.completeReadySegment(startIndex)
		}
		return nil
	}

	started := time.Now()
	deadline := time.Now().Add(20 * time.Second)
	lastWaitLog := time.Time{}
	samples := 0
	bytes := 0
	gstDebugf("runner EnsureInit wait start task=%s file=%s audio=%d startIndex=%d position=%.3f deadline=%s", r.task.ID, r.task.FileID, audio, startIndex, r.position(), deadline.Format(time.RFC3339Nano))
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			gstErrorf("runner EnsureInit context canceled task=%s file=%s audio=%d startIndex=%d err=%v duration=%s", r.task.ID, r.task.FileID, audio, startIndex, err, time.Since(started))
			return err
		}

		r.drainSubtitleSinks()
		if err := r.pollPipelineError(); err != nil {
			r.freezeAtSegment(startIndex)
			return err
		}

		sample := gstRuntime.appSinkTryPullSample(r.sink, uint64(100*time.Millisecond))
		if sample == 0 {
			if err := r.pollPipelineError(); err != nil {
				r.freezeAtSegment(startIndex)
				return err
			}
			if gstRuntime.appSinkIsEOS(r.sink) {
				gstDebugf("runner EnsureInit EOS task=%s file=%s audio=%d startIndex=%d samples=%d bytes=%d duration=%s", r.task.ID, r.task.FileID, audio, startIndex, samples, bytes, time.Since(started))
				return ErrSegmentNotReady
			}
			r.logWaitSummary("EnsureInit", started, &lastWaitLog, samples, bytes)
			continue
		}

		samples++
		sampleBytes := 0
		err := gstRuntime.withSampleBytes(sample, func(data []byte) error {
			sampleBytes = len(data)
			if len(data) == 0 {
				return nil
			}
			return r.reader.Push(data)
		})
		gstRuntime.sampleUnref(sample)
		bytes += sampleBytes
		if err != nil {
			gstErrorf("runner EnsureInit MP4 reader failed task=%s file=%s audio=%d startIndex=%d err=%v samples=%d bytes=%d duration=%s", r.task.ID, r.task.FileID, audio, startIndex, err, samples, bytes, time.Since(started))
			r.freezeAtSegment(startIndex)
			return err
		}
		r.drainSubtitleSinks()

		if err := r.pollPipelineError(); err != nil {
			r.freezeAtSegment(startIndex)
			return err
		}

		if r.task.hasInitMP4() {
			if r.readySegment.complete {
				r.completeReadySegment(startIndex)
			}
			gstDebugf("runner EnsureInit init MP4 ready task=%s file=%s audio=%d startIndex=%d samples=%d bytes=%d duration=%s", r.task.ID, r.task.FileID, audio, startIndex, samples, bytes, time.Since(started))
			return nil
		}
		r.logWaitSummary("EnsureInit", started, &lastWaitLog, samples, bytes)
	}

	if err := r.pollPipelineError(); err != nil {
		r.freezeAtSegment(startIndex)
		return err
	}

	gstErrorf("runner EnsureInit timeout task=%s file=%s audio=%d startIndex=%d samples=%d bytes=%d deadline=%s duration=%s", r.task.ID, r.task.FileID, audio, startIndex, samples, bytes, deadline.Format(time.RFC3339Nano), time.Since(started))
	return ErrSegmentNotReady
}

func (r *gstRunner) GetSegment(ctx context.Context, index int, audio int) (Segment, error) {
	return r.getSegmentWithTimeout(ctx, index, audio, 20*time.Second)
}

func (r *gstRunner) GetSegmentWithTimeout(ctx context.Context, index int, audio int, timeout time.Duration) (Segment, error) {
	return r.getSegmentWithTimeout(ctx, index, audio, timeout)
}

func (r *gstRunner) getSegmentWithTimeout(ctx context.Context, index int, audio int, timeout time.Duration) (Segment, error) {
	if timeout <= 0 {
		timeout = 20 * time.Second
	}

	if r.IsFrozen() {
		conf := r.task.Config.normalized()
		startSeconds := float64(index * conf.SegmentSeconds)
		if err := r.Seek(startSeconds); err != nil {
			return Segment{}, err
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

	r.drainSubtitleSinks()
	if r.readySegment.index == index && r.readySegment.complete {
		seg := r.readySegment.segment
		return seg, nil
	}

	r.discardReadySegment()

	started := time.Now()
	deadline := time.Now().Add(timeout)
	lastWaitLog := time.Time{}
	samples := 0
	bytes := 0
	gstDebugf("runner GetSegment wait start task=%s file=%s audio=%d segment=%d position=%.3f timeout=%s deadline=%s", r.task.ID, r.task.FileID, audio, index, r.position(), timeout, deadline.Format(time.RFC3339Nano))
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			gstErrorf("runner GetSegment context canceled task=%s file=%s audio=%d segment=%d err=%v duration=%s", r.task.ID, r.task.FileID, audio, index, err, time.Since(started))
			return Segment{}, err
		}

		r.drainSubtitleSinks()
		if err := r.pollPipelineError(); err != nil {
			r.freezeAtSegment(index)
			return Segment{}, err
		}

		sample := gstRuntime.appSinkTryPullSample(r.sink, uint64(100*time.Millisecond))
		if sample == 0 {
			if err := r.pollPipelineError(); err != nil {
				r.freezeAtSegment(index)
				return Segment{}, err
			}

			if gstRuntime.appSinkIsEOS(r.sink) {
				r.drainSubtitleSinks()
				gstDebugf("runner GetSegment EOS task=%s file=%s audio=%d segment=%d samples=%d bytes=%d elapsed=%s", r.task.ID, r.task.FileID, audio, index, samples, bytes, time.Since(started))
				seg, err := r.drainEndOfStream(index)
				if err != nil {
					return Segment{}, err
				}
				return seg, nil
			}
			r.logWaitSummary("GetSegment", started, &lastWaitLog, samples, bytes)
			continue
		}

		samples++
		sampleBytes := 0
		err := gstRuntime.withSampleBytes(sample, func(data []byte) error {
			sampleBytes = len(data)
			if len(data) == 0 {
				return nil
			}
			return r.reader.Push(data)
		})
		gstRuntime.sampleUnref(sample)
		bytes += sampleBytes
		if err != nil {
			gstErrorf("runner GetSegment MP4 reader failed task=%s file=%s audio=%d segment=%d err=%v samples=%d bytes=%d duration=%s", r.task.ID, r.task.FileID, audio, index, err, samples, bytes, time.Since(started))
			r.freezeAtSegment(index)
			return Segment{}, err
		}
		r.drainSubtitleSinks()

		if err := r.pollPipelineError(); err != nil {
			r.freezeAtSegment(index)
			return Segment{}, err
		}

		if r.readySegment.complete {
			seg := r.completeReadySegment(index)
			r.drainSubtitleSinks()
			gstDebugf("runner GetSegment ready task=%s file=%s audio=%d segment=%d samples=%d bytes=%d size=%d duration=%s", r.task.ID, r.task.FileID, audio, index, samples, bytes, seg.Len(), time.Since(started))
			return seg, nil
		}
		r.logWaitSummary("GetSegment", started, &lastWaitLog, samples, bytes)
	}

	if err := r.pollPipelineError(); err != nil {
		r.freezeAtSegment(index)
		return Segment{}, err
	}

	gstErrorf("runner GetSegment timeout task=%s file=%s audio=%d segment=%d samples=%d bytes=%d timeout=%s deadline=%s duration=%s", r.task.ID, r.task.FileID, audio, index, samples, bytes, timeout, deadline.Format(time.RFC3339Nano), time.Since(started))
	return Segment{}, ErrSegmentNotReady
}

func (r *gstRunner) pollPipelineError() error {
	if r.bus == 0 || gstRuntime == nil {
		return nil
	}
	err := gstRuntime.popBusError(r.bus, 0)
	if err != nil {
		gstErrorf("GStreamer bus error task=%s file=%s audio=%d err=%v", r.task.ID, r.task.FileID, r.audioIndex, err)
	}
	return err
}

func (r *gstRunner) drainEndOfStream(index int) (Segment, error) {
	started := time.Now()
	gstDebugf("drainEndOfStream start task=%s file=%s audio=%d segment=%d", r.task.ID, r.task.FileID, r.audioIndex, index)
	r.drainSubtitleSinks()
	if r.reader == nil {
		gstErrorf("drainEndOfStream failed task=%s file=%s audio=%d segment=%d err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, index, ErrSegmentNotReady, time.Since(started))
		return Segment{}, ErrSegmentNotReady
	}

	completed, err := r.reader.TryProcessDeferred()
	if err != nil {
		if len(r.reader.video) > 0 && !r.reader.video[0].startsWithSync {
			err = r.reader.undecodableEOSRemainderError()
			gstErrorf("drainEndOfStream MP4 reader failed task=%s file=%s audio=%d segment=%d err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, index, err, time.Since(started))
			return Segment{}, err
		}
		gstErrorf("drainEndOfStream MP4 reader failed task=%s file=%s audio=%d segment=%d err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, index, err, time.Since(started))
		return Segment{}, err
	}
	if completed {
		if !r.readySegment.complete {
			err := errors.New("mp4 reader completed a segment without onSegment callback")
			gstErrorf("drainEndOfStream MP4 reader failed task=%s file=%s audio=%d segment=%d err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, index, err, time.Since(started))
			return Segment{}, err
		}
		seg := r.completeReadySegment(index)
		r.drainSubtitleSinks()
		gstDebugf("drainEndOfStream completed deferred task=%s file=%s audio=%d segment=%d size=%d duration=%s", r.task.ID, r.task.FileID, r.audioIndex, index, seg.Len(), time.Since(started))
		return seg, nil
	}

	completed, err = r.reader.TryBuildEndOfStreamRemainder()
	if err != nil {
		gstErrorf("drainEndOfStream MP4 reader failed task=%s file=%s audio=%d segment=%d err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, index, err, time.Since(started))
		return Segment{}, err
	}
	if completed {
		if !r.readySegment.complete {
			err := errors.New("mp4 reader completed EOS remainder without onSegment callback")
			gstErrorf("drainEndOfStream MP4 reader failed task=%s file=%s audio=%d segment=%d err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, index, err, time.Since(started))
			return Segment{}, err
		}
		seg := r.completeReadySegment(index)
		r.drainSubtitleSinks()
		gstDebugf("drainEndOfStream completed remainder task=%s file=%s audio=%d segment=%d size=%d duration=%s", r.task.ID, r.task.FileID, r.audioIndex, index, seg.Len(), time.Since(started))
		return seg, nil
	}

	if err := r.reader.EndOfStreamError(); err != nil {
		gstErrorf("drainEndOfStream MP4 reader failed task=%s file=%s audio=%d segment=%d err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, index, err, time.Since(started))
		return Segment{}, err
	}

	gstErrorf("drainEndOfStream exhausted task=%s file=%s audio=%d segment=%d err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, index, ErrEndOfStreamExhausted, time.Since(started))
	return Segment{}, ErrEndOfStreamExhausted
}

func (r *gstRunner) completeReadySegment(index int) Segment {
	requestedIndex := index
	if index > 0 {
		r.readySegment.index = index
	} else {
		r.readySegment.index = 0
	}
	assignedIndex := r.readySegment.index
	seg := r.readySegment.segment
	absoluteStart := seg.StartSeconds + r.positionSeekSeconds
	absoluteEnd := seg.EndSeconds + r.positionSeekSeconds
	gstDebugf("completeReadySegment task=%s file=%s audio=%d requestedIndex=%d assignedIndex=%d size=%d start=%.3f end=%.3f positionSeek=%.3f absoluteStart=%.3f absoluteEnd=%.3f runnerPosition=%.3f", r.task.ID, r.task.FileID, r.audioIndex, requestedIndex, assignedIndex, seg.Len(), seg.StartSeconds, seg.EndSeconds, r.positionSeekSeconds, absoluteStart, absoluteEnd, r.position())
	return r.readySegment.segment
}

func (r *gstRunner) acceptSegment(seg Segment) {
	r.readySegment.segment = seg
	r.readySegment.complete = true
	if seg.EndSeconds >= seg.StartSeconds {
		r.setPosition(seg.EndSeconds + r.positionSeekSeconds)
	} else if seg.StartSeconds >= 0 {
		r.setPosition(seg.StartSeconds + r.positionSeekSeconds)
	}
	gstDebugf("acceptSegment task=%s file=%s audio=%d StartSeconds=%.3f EndSeconds=%.3f duration=%.3f size=%d payloads=%d positionSeek=%.3f absoluteStart=%.3f absoluteEnd=%.3f position=%.3f readyIndex=%d", r.task.ID, r.task.FileID, r.audioIndex, seg.StartSeconds, seg.EndSeconds, seg.EndSeconds-seg.StartSeconds, seg.Len(), len(seg.Payloads), r.positionSeekSeconds, seg.StartSeconds+r.positionSeekSeconds, seg.EndSeconds+r.positionSeekSeconds, r.position(), r.readySegment.index)
}

func (r *gstRunner) discardReadySegment() {
	hadReady := r.readySegment.complete

	r.readySegment.index = -1
	r.readySegment.complete = false
	r.readySegment.segment = Segment{}

	if hadReady && r.reader != nil {
		r.reader.ReclaimPayloads()
	}
}

func (r *gstRunner) DiscardSegment() {
	r.discardReadySegment()
}

func (r *gstRunner) freezeAtSegment(index int) {
	position := r.position()
	virtualSeconds := position
	if index > 0 {
		conf := r.task.Config.normalized()
		virtualSeconds = float64(index * conf.SegmentSeconds)
	}

	seconds := position
	action := "freeze-current-position"
	if index > 0 {
		seconds = virtualSeconds
		action = "freeze-segment-position"
	} else if seconds <= 0 {
		seconds = virtualSeconds
	}

	gstDebugf("freezeAtSegment task=%s file=%s audio=%d segment=%d position=%.3f virtual=%.3f freezeAt=%.3f action=%s", r.task.ID, r.task.FileID, r.audioIndex, index, position, virtualSeconds, seconds, action)
	r.freezeAtPosition(seconds)
}

func (r *gstRunner) freezeAtPosition(seconds float64) {
	gstDebugf("freezeAtPosition task=%s file=%s audio=%d position=%.3f", r.task.ID, r.task.FileID, r.audioIndex, seconds)
	r.stopPipeline()
	r.discardReadySegment()
	r.reader.SeekReset(seconds)
	r.frozen.Store(true)
	r.setPosition(seconds)
	r.positionSeekSeconds = seconds
	r.statePlaying = false
}

func (r *gstRunner) startPipeline(seconds float64) (float64, error) {
	started := time.Now()
	timelineGeneration := r.task.resetSubtitleTimeline(seconds)
	gstDebugf("startPipeline start task=%s file=%s audio=%d requested=%.3f subtitleGeneration=%d", r.task.ID, r.task.FileID, r.audioIndex, seconds, timelineGeneration)
	stageStarted := time.Now()
	pipeline, err := gstRuntime.parseLaunch(r.createPipelineArgs())
	if err != nil {
		gstErrorf("startPipeline parseLaunch failed task=%s file=%s audio=%d requested=%.3f err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, seconds, err, time.Since(stageStarted))
		return 0, err
	}
	gstDebugf("startPipeline parseLaunch completed task=%s file=%s audio=%d duration=%s", r.task.ID, r.task.FileID, r.audioIndex, time.Since(stageStarted))

	stageStarted = time.Now()
	sink := gstRuntime.binGetByName(pipeline, "out")
	if sink == 0 {
		gstRuntime.elementSetState(pipeline, gstStateNull)
		gstRuntime.objectUnref(pipeline)
		err := errors.New("appsink element is not available")
		gstErrorf("startPipeline appsink lookup failed task=%s file=%s audio=%d err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, err, time.Since(stageStarted))
		return 0, err
	}
	gstDebugf("startPipeline appsink lookup completed task=%s file=%s audio=%d duration=%s", r.task.ID, r.task.FileID, r.audioIndex, time.Since(stageStarted))

	conf := r.task.Config.normalized()
	fragmentDuration := time.Duration(conf.SegmentSeconds) * time.Second
	pipelineGen := nextPipelineGen.Add(1)
	filterHEVCLeadingPictures := seconds > 0 && r.task.Probe.IsH265() && !r.transcodesVideo(conf, r.task.Probe)
	probes := newGSTPipelineProbes(r, pipeline, pipelineGen, filterHEVCLeadingPictures, fragmentDuration)

	stageStarted = time.Now()
	bus := gstRuntime.pipelineGetBus(pipeline)
	gstDebugf("startPipeline bus acquired task=%s file=%s audio=%d bus=%t duration=%s", r.task.ID, r.task.FileID, r.audioIndex, bus != 0, time.Since(stageStarted))
	actualStartSeconds := seconds
	var demux uintptr
	var subtitleSinks map[int]uintptr

	cleanup := func() {
		probes.release()
		gstRuntime.elementSetState(pipeline, gstStateNull)
		if demux != 0 {
			gstRuntime.objectUnref(demux)
		}
		for _, subtitleSink := range subtitleSinks {
			gstRuntime.objectUnref(subtitleSink)
		}
		gstRuntime.objectUnref(sink)
		gstRuntime.objectUnref(pipeline)
		gstRuntime.objectUnref(bus)
	}

	subtitleSinks, err = r.lookupSubtitleSinks(pipeline)
	if err != nil {
		cleanup()
		gstErrorf("startPipeline subtitle appsink lookup failed task=%s file=%s audio=%d err=%v", r.task.ID, r.task.FileID, r.audioIndex, err)
		return 0, err
	}
	gstDebugf("startPipeline subtitle appsinks acquired task=%s file=%s audio=%d count=%d", r.task.ID, r.task.FileID, r.audioIndex, len(subtitleSinks))

	if seconds > 0 {
		stageStarted = time.Now()
		demux = gstRuntime.binGetByName(pipeline, "d")
		if demux == 0 {
			cleanup()
			err := errors.New("matroskademux element is not available")
			gstErrorf("startPipeline demux lookup failed task=%s file=%s audio=%d err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, err, time.Since(stageStarted))
			return 0, err
		}
		gstDebugf("startPipeline demux lookup completed task=%s file=%s audio=%d duration=%s", r.task.ID, r.task.FileID, r.audioIndex, time.Since(stageStarted))

		stageStarted = time.Now()
		if err := r.setPipelineState(pipeline, bus, gstStatePaused); err != nil {
			cleanup()
			gstErrorf("startPipeline PAUSED failed task=%s file=%s audio=%d requested=%.3f err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, seconds, err, time.Since(stageStarted))
			return 0, err
		}
		gstDebugf("startPipeline PAUSED completed task=%s file=%s audio=%d duration=%s", r.task.ID, r.task.FileID, r.audioIndex, time.Since(stageStarted))

		probes.resetForSeek()

		stageStarted = time.Now()
		gstDebugf("startPipeline seek send task=%s file=%s audio=%d target=demux snap=before requested=%.3f flags=%d", r.task.ID, r.task.FileID, r.audioIndex, seconds, gstSeekFlagFlush|gstSeekFlagKeyUnit|gstSeekFlagSnapBefore)
		if !gstRuntime.elementSeekSimple(demux, gstFormatTime, gstSeekFlagFlush|gstSeekFlagKeyUnit|gstSeekFlagSnapBefore, int64(math.Round(seconds*1_000_000_000))) {
			cleanup()
			err := errors.New("gstreamer seek failed")
			gstErrorf("startPipeline seek send failed task=%s file=%s audio=%d requested=%.3f err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, seconds, err, time.Since(stageStarted))
			return 0, err
		}
		gstDebugf("startPipeline seek sent task=%s file=%s audio=%d requested=%.3f duration=%s", r.task.ID, r.task.FileID, r.audioIndex, seconds, time.Since(stageStarted))

		stageStarted = time.Now()
		waitResult := gstRuntime.elementGetState(pipeline, 5*time.Second)
		gstDebugf("startPipeline seek wait result task=%s file=%s audio=%d requested=%.3f GstStateChangeReturn=%d duration=%s", r.task.ID, r.task.FileID, r.audioIndex, seconds, waitResult, time.Since(stageStarted))
		switch waitResult {
		case gstStateChangeSuccess, gstStateChangeNoPreroll:
		case gstStateChangeAsync:
			if err := gstRuntime.popBusError(bus, 0); err != nil {
				cleanup()
				gstErrorf("startPipeline seek wait bus error task=%s file=%s audio=%d requested=%.3f err=%v", r.task.ID, r.task.FileID, r.audioIndex, seconds, err)
				return 0, err
			}
			cleanup()
			err := fmt.Errorf("gstreamer seek to %.3fs timed out", seconds)
			gstErrorf("startPipeline seek wait timeout task=%s file=%s audio=%d requested=%.3f err=%v", r.task.ID, r.task.FileID, r.audioIndex, seconds, err)
			return 0, err
		case gstStateChangeFailure:
			if err := gstRuntime.popBusError(bus, 0); err != nil {
				cleanup()
				gstErrorf("startPipeline seek wait bus error task=%s file=%s audio=%d requested=%.3f err=%v", r.task.ID, r.task.FileID, r.audioIndex, seconds, err)
				return 0, err
			}
			cleanup()
			err := fmt.Errorf("gstreamer seek to %.3fs failed", seconds)
			gstErrorf("startPipeline seek wait failed task=%s file=%s audio=%d requested=%.3f err=%v", r.task.ID, r.task.FileID, r.audioIndex, seconds, err)
			return 0, err
		default:
			cleanup()
			err := fmt.Errorf("unexpected GstStateChangeReturn=%d after seek", waitResult)
			gstErrorf("startPipeline seek wait unexpected task=%s file=%s audio=%d requested=%.3f err=%v", r.task.ID, r.task.FileID, r.audioIndex, seconds, err)
			return 0, err
		}

		stageStarted = time.Now()
		positionNS, ok := gstRuntime.elementQueryPosition(pipeline)
		if ok {
			actualStartSeconds = float64(positionNS) / 1_000_000_000.0
			gstDebugf("startPipeline position query completed task=%s file=%s audio=%d requested=%.3f actual=%.3f delta=%.3f duration=%s", r.task.ID, r.task.FileID, r.audioIndex, seconds, actualStartSeconds, actualStartSeconds-seconds, time.Since(stageStarted))
		} else {
			gstDebugf("startPipeline position query unavailable task=%s file=%s audio=%d requested=%.3f fallback=%.3f duration=%s", r.task.ID, r.task.FileID, r.audioIndex, seconds, actualStartSeconds, time.Since(stageStarted))
		}
	}

	stageStarted = time.Now()
	if err := r.setPipelineState(pipeline, bus, gstStatePlaying); err != nil {
		cleanup()
		gstErrorf("startPipeline PLAYING failed task=%s file=%s audio=%d requested=%.3f err=%v duration=%s", r.task.ID, r.task.FileID, r.audioIndex, seconds, err, time.Since(stageStarted))
		return 0, err
	}
	gstDebugf("startPipeline PLAYING completed task=%s file=%s audio=%d duration=%s", r.task.ID, r.task.FileID, r.audioIndex, time.Since(stageStarted))

	if demux != 0 {
		gstRuntime.objectUnref(demux)
		demux = 0
	}

	r.pipeline = pipeline
	r.bus = bus
	r.sink = sink
	r.subtitleSinks = subtitleSinks
	r.probes = probes
	r.task.setSubtitleTimelineOrigin(timelineGeneration, seconds, actualStartSeconds)
	gstDebugf("startPipeline completed task=%s file=%s audio=%d requested=%.3f actual=%.3f delta=%.3f subtitleGeneration=%d duration=%s", r.task.ID, r.task.FileID, r.audioIndex, seconds, actualStartSeconds, actualStartSeconds-seconds, timelineGeneration, time.Since(started))
	return actualStartSeconds, nil
}

func (r *gstRunner) setPipelineState(pipeline uintptr, bus uintptr, state int32) error {
	started := time.Now()
	setResult := gstRuntime.elementSetState(pipeline, state)
	gstDebugf("setPipelineState state=%d setResult=%d", state, setResult)
	if setResult == gstStateChangeFailure {
		if err := gstRuntime.popBusError(bus, 0); err != nil {
			return err
		}
		return fmt.Errorf("gstreamer failed to request state change to %d", state)
	}

	waitStarted := time.Now()
	waitResult := gstRuntime.elementGetState(pipeline, 5*time.Second)
	waitDuration := time.Since(waitStarted)
	gstDebugf("setPipelineState state=%d waitResult=%d GstStateChangeReturn=%d waitDuration=%s duration=%s", state, waitResult, waitResult, waitDuration, time.Since(started))
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
	if r.pipeline != 0 || r.sink != 0 || r.bus != 0 {
		gstDebugf("stopPipeline task=%s file=%s audio=%d pipeline=%t sink=%t bus=%t", r.task.ID, r.task.FileID, r.audioIndex, r.pipeline != 0, r.sink != 0, r.bus != 0)
	}
	if r.probes != nil {
		r.probes.release()
		r.probes = nil
	}
	if r.pipeline != 0 {
		_ = gstRuntime.elementSetState(r.pipeline, gstStateNull)
	}
	for subtitleIndex, subtitleSink := range r.subtitleSinks {
		if subtitleSink != 0 {
			gstRuntime.objectUnref(subtitleSink)
		}
		delete(r.subtitleSinks, subtitleIndex)
	}
	r.subtitleSinks = nil
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
	gstDebugf("Dispose task=%s file=%s audio=%d position=%.3f", r.task.ID, r.task.FileID, r.audioIndex, r.position())
	r.stopPipeline()
	r.discardReadySegment()
	if r.reader != nil {
		r.reader.SeekReset(r.position())
	}
	r.statePlaying = false
}

func (r *gstRunner) Frozen() {
	gstDebugf("Frozen task=%s file=%s audio=%d position=%.3f", r.task.ID, r.task.FileID, r.audioIndex, r.position())
	r.freezeAtPosition(r.position())
}

func (r *gstRunner) logWaitSummary(operation string, started time.Time, last *time.Time, samples int, bytes int) {
	now := time.Now()
	if !last.IsZero() && now.Sub(*last) < time.Second {
		return
	}
	*last = now
	gstDebugf("%s waiting: task=%s file=%s audio=%d elapsed=%s samples=%d bytes=%d position=%.3f", operation, r.task.ID, r.task.FileID, r.audioIndex, now.Sub(started), samples, bytes, r.position())
}

func (r *gstRunner) IsFrozen() bool {
	return r.frozen.Load()
}

func (r *gstRunner) Position() float64 {
	return r.position()
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
