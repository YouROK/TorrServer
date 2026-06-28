//go:build (windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64))

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

	var gstPlugins string
	switch runtime.GOOS {
	case "windows":
		gstPlugins = firstExistingPath(gstPluginCandidates(roots))
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
		gstPluginScanner = firstExistingPath(gstPluginScannerCandidates(roots))
	case "linux", "darwin":
		gstPluginScanner = firstExistingPath(gstPluginScannerCandidates(roots))
	}
	if gstPluginScanner != "" {
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

func setExistingEnvPaths(key string, candidates []string) {
	values := existingPaths(candidates)
	if len(values) == 0 {
		return
	}
	_ = os.Setenv(key, strings.Join(values, string(os.PathListSeparator)))
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
	conf := r.task.Config.normalized()
	probe := r.task.Probe
	gstVersion := effectiveGStreamerVersion(conf)

	queueNS := int64(conf.SegmentSeconds*2) * int64(time.Second)
	var sb strings.Builder

	sb.WriteString("souphttpsrc ")
	sb.WriteString("location=\"")
	sb.WriteString(r.task.SourceURL)
	sb.WriteString("\" is-live=false keep-alive=true timeout=60 retries=5 ")
	if gstVersion.atLeast(1, 26) {
		sb.WriteString("retry-backoff-factor=0.5 retry-backoff-max=10 ")
	}
	r.writeSourceQueue(&sb)
	sb.WriteString(" ! matroskademux name=d multiqueue name=mq use-buffering=false max-size-buffers=0 max-size-bytes=0 max-size-time=")
	sb.WriteString(strconv.FormatInt(queueNS, 10))
	sb.WriteString(" ")

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

	if audioTrack := probe.AudioTrack(r.audioIndex); audioTrack != nil {
		sb.WriteString("d.audio_")
		sb.WriteString(strconv.Itoa(audioTrack.Index))
		sb.WriteString(" ! mq.sink_1 mq.src_1 ! ")
		if audioTrack.IsAACAudio() {
			sb.WriteString("aacparse ! audio/mpeg,mpegversion=4,stream-format=raw ! mux.audio_0 ")
		} else {
			aacEncoder := r.aacEncoder()

			sb.WriteString("decodebin ! audioconvert dithering=none noise-shaping=none ! audioresample quality=2 sinc-filter-mode=full ! audio/x-raw,format=")
			sb.WriteString(aacRawFormat())
			sb.WriteString(",layout=interleaved,rate=48000,channels=2 ! ")
			sb.WriteString(aacEncoder)
			sb.WriteString(" bitrate=")
			sb.WriteString(strconv.Itoa(conf.AACBitrateKbps * 1000))
			sb.WriteString(" ! aacparse ! audio/mpeg,mpegversion=4,stream-format=raw,rate=48000,channels=2 ! mux.audio_0 ")
		}
	}

	sb.WriteString("mp4mux name=mux fragment-duration=")
	sb.WriteString(strconv.Itoa(conf.SegmentSeconds * 1000))
	sb.WriteString(" streamable=true ! appsink name=out emit-signals=false sync=false max-buffers=1")
	if gstVersion.atLeast(1, 28) {
		sb.WriteString(" leaky-type=none")
	} else {
		sb.WriteString(" drop=false")
	}
	sb.WriteString(" wait-on-eos=false")

	return sb.String()
}

func effectiveGStreamerVersion(conf Config) gstVersionInfo {
	if gstRuntime != nil && gstRuntime.version.valid() {
		return gstRuntime.version
	}
	if conf.GSTVersion <= 0 {
		conf.GSTVersion = 1.22
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
		keyIntMax = maxInt(1, int(math.Round(float64(frameRateNum*conf.SegmentSeconds)/float64(frameRateDen))))
	}

	sb.WriteString("mq.src_0 ! decodebin ! videoconvert ! video/x-raw,format=I420 ! x264enc tune=zerolatency speed-preset=veryfast bitrate=")
	sb.WriteString(strconv.Itoa(conf.VideoBitrate))
	sb.WriteString(" key-int-max=")
	sb.WriteString(strconv.Itoa(keyIntMax))
	sb.WriteString(" bframes=0 byte-stream=false ! video/x-h264,profile=main,stream-format=avc,alignment=au ! h264parse config-interval=0 ! h264timestamper ! video/x-h264,profile=main,stream-format=avc,alignment=au ! mux.video_0 ")
}

func (r *gstRunner) Seek(seconds float64) bool {
	r.stopPipeline()

	r.discardReadySegment()
	r.reader.SeekReset(seconds)

	actualSeconds, err := r.startPipeline(seconds)
	if err != nil {
		r.freezeAtPosition(seconds)
		return false
	}
	r.reader.SeekReset(actualSeconds)

	r.frozen.Store(false)
	r.setPosition(actualSeconds)
	r.positionSeekSeconds = actualSeconds
	r.statePlaying = true
	return true
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

	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return err
		}

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
				return ErrSegmentNotReady
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
			r.freezeAtSegment(startIndex)
			return err
		}

		if err := r.pollPipelineError(); err != nil {
			r.freezeAtSegment(startIndex)
			return err
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

	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return Segment{}, err
		}

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
				seg, err := r.drainEndOfStream(index)
				if err != nil {
					return Segment{}, err
				}
				return seg, nil
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

		if err := r.pollPipelineError(); err != nil {
			r.freezeAtSegment(index)
			return Segment{}, err
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

func (r *gstRunner) pollPipelineError() error {
	if r.bus == 0 || gstRuntime == nil {
		return nil
	}
	return gstRuntime.popBusError(r.bus, 0)
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
		r.setPosition(seg.EndSeconds + r.positionSeekSeconds)
	} else if seg.StartSeconds >= 0 {
		r.setPosition(seg.StartSeconds + r.positionSeekSeconds)
	}
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

func (r *gstRunner) freezeAtSegment(index int) {
	seconds := r.position()
	if index > 0 {
		seconds = float64(index * r.task.Config.SegmentSeconds)
	}

	r.freezeAtPosition(seconds)
}

func (r *gstRunner) freezeAtPosition(seconds float64) {
	r.stopPipeline()
	r.discardReadySegment()
	r.reader.SeekReset(seconds)
	r.frozen.Store(true)
	r.setPosition(seconds)
	r.positionSeekSeconds = seconds
	r.statePlaying = false
}

func (r *gstRunner) startPipeline(seconds float64) (float64, error) {
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
	actualStartSeconds := seconds

	cleanup := func() {
		gstRuntime.elementSetState(pipeline, gstStateNull)
		gstRuntime.objectUnref(sink)
		gstRuntime.objectUnref(pipeline)
		gstRuntime.objectUnref(bus)
	}

	if seconds > 0 {
		if err := r.setPipelineState(pipeline, bus, gstStatePaused); err != nil {
			cleanup()
			return 0, err
		}

		if !gstRuntime.elementSeekSimple(pipeline, gstFormatTime, gstSeekFlagFlush|gstSeekFlagKeyUnit|gstSeekFlagSnapAfter, int64(math.Round(seconds*1_000_000_000))) {
			cleanup()
			return 0, errors.New("gstreamer seek failed")
		}

		waitResult := gstRuntime.elementGetState(pipeline, 5*time.Second)
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

		positionNS, ok := gstRuntime.elementQueryPosition(pipeline)
		if !ok {
			cleanup()
			return 0, errors.New("gstreamer position query failed after seek")
		}
		actualStartSeconds = float64(positionNS) / 1_000_000_000.0
	}

	if err := r.setPipelineState(pipeline, bus, gstStatePlaying); err != nil {
		cleanup()
		return 0, err
	}

	r.pipeline = pipeline
	r.bus = bus
	r.sink = sink
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

	waitResult := gstRuntime.elementGetState(pipeline, 5*time.Second)
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
	r.discardReadySegment()
	if r.reader != nil {
		r.reader.SeekReset(r.position())
	}
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
