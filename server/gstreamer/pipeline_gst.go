//go:build (windows && (amd64 || arm64)) || (linux && (amd64 || arm64))

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

	reader *Mp4BoxReader

	pipeline uintptr
	bus      uintptr
	sink     uintptr

	dead   atomic.Bool
	frozen atomic.Bool
	eos    atomic.Bool
}

func newPipelineRunner(task *Task, audio int) (pipelineRunner, error) {
	gstInitOnce.Do(func() {
		setupGStreamer(task.Config)
		gstRuntime, gstInitErr = loadGST(task.Config)
		if gstInitErr != nil {
			return
		}
		gstInitErr = gstRuntime.init()
	})
	if gstInitErr != nil {
		return nil, errors.Join(ErrPipelineDisabled, gstInitErr)
	}

	runner := &gstRunner{
		task:       task,
		audioIndex: validAudioIndex(task.Probe, audio),
		reader: NewMp4BoxReader(
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
		),
	}
	runner.readySegment.index = -1
	return runner, nil
}

func setupGStreamer(conf Config) {
	_ = os.Setenv("GST_REGISTRY", filepath.Join(os.TempDir(), "torrserver-gstreamer-registry.bin"))

	if conf.GSTPath == "" {
		return
	}

	gstBin := filepath.Join(conf.GSTPath, "bin")
	if runtime.GOOS == "windows" {
		if _, err := os.Stat(gstBin); err == nil {
			prependEnvPath("PATH", gstBin)
		}
	}

	var gstPlugins string
	switch runtime.GOOS {
	case "windows":
		gstPlugins = filepath.Join(conf.GSTPath, "lib", "gstreamer-1.0")
	case "linux":
		for _, candidate := range []string{
			filepath.Join(conf.GSTPath, "lib", "gstreamer-1.0"),
			filepath.Join(conf.GSTPath, "lib64", "gstreamer-1.0"),
			filepath.Join(conf.GSTPath, "lib", runtime.GOARCH+"-linux-gnu", "gstreamer-1.0"),
			filepath.Join(conf.GSTPath, "lib", "x86_64-linux-gnu", "gstreamer-1.0"),
		} {
			if _, err := os.Stat(candidate); err == nil {
				gstPlugins = candidate
				break
			}
		}
	}
	if gstPlugins != "" {
		_ = os.Setenv("GST_PLUGIN_PATH", gstPlugins)
		_ = os.Setenv("GST_PLUGIN_SYSTEM_PATH_1_0", gstPlugins)
	}

	var gstPluginScanner string
	switch runtime.GOOS {
	case "windows":
		gstPluginScanner = filepath.Join(conf.GSTPath, "libexec", "gstreamer-1.0", "gst-plugin-scanner.exe")
	case "linux":
		for _, candidate := range []string{
			filepath.Join(conf.GSTPath, "libexec", "gstreamer-1.0", "gst-plugin-scanner"),
			filepath.Join(conf.GSTPath, "lib", "gstreamer-1.0", "gst-plugin-scanner"),
			filepath.Join(conf.GSTPath, "lib64", "gstreamer-1.0", "gst-plugin-scanner"),
		} {
			if _, err := os.Stat(candidate); err == nil {
				gstPluginScanner = candidate
				break
			}
		}
	}
	if gstPluginScanner != "" {
		_ = os.Setenv("GST_PLUGIN_SCANNER", gstPluginScanner)
	}
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
	sb.WriteString("! queue2 use-buffering=false max-size-buffers=0 max-size-bytes=")
	sb.WriteString(strconv.Itoa(16 * 1024 * 1024))
	sb.WriteString(" max-size-time=")
	sb.WriteString(strconv.FormatInt(queueNS, 10))
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
			sb.WriteString(" leaky=0 ! h264parse config-interval=-1 ! h264timestamper ! video/x-h264,stream-format=avc,alignment=au ! mux.video_0 ")
		}

	case probe.IsH265():
		if conf.TranscodeH265 {
			r.transcodeToH264(&sb, videoQueueBytes, queueNS)
		} else {
			sb.WriteString("d.video_0 ! queue max-size-buffers=0 max-size-bytes=")
			sb.WriteString(strconv.Itoa(videoQueueBytes))
			sb.WriteString(" max-size-time=")
			sb.WriteString(strconv.FormatInt(queueNS, 10))
			sb.WriteString(" leaky=0 ! h265parse config-interval=-1 ! h265timestamper ! video/x-h265,stream-format=hvc1,alignment=au ! mux.video_0 ")
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
	sb.WriteString(" leaky=0 ! decodebin ! audioconvert ! audioresample ! audio/x-raw,rate=48000,channels=2 ! avenc_aac bitrate=")
	sb.WriteString(strconv.Itoa(conf.AACBitrateKbps * 1000))
	sb.WriteString(" ! aacparse ! audio/mpeg,mpegversion=4,stream-format=raw,rate=48000,channels=2 ! mux.audio_0 ")

	sb.WriteString("mp4mux name=mux fragment-duration=")
	sb.WriteString(strconv.Itoa(conf.SegmentSeconds * 1000))
	sb.WriteString(" streamable=true ! appsink name=out emit-signals=false sync=false max-buffers=1 max-bytes=0 max-time=0")
	if conf.GSTVersion >= 1.28 {
		sb.WriteString(" leaky-type=none")
	} else {
		sb.WriteString(" drop=false")
	}
	sb.WriteString(" wait-on-eos=false")

	return sb.String()
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
	sb.WriteString(" bframes=0 byte-stream=false ! video/x-h264,profile=main,stream-format=avc,alignment=au ! h264parse config-interval=-1 ! h264timestamper ! video/x-h264,profile=main,stream-format=avc,alignment=au ! mux.video_0 ")
}

func (r *gstRunner) Seek(seconds float64) bool {
	if r.IsDead() || !r.statePlaying {
		return false
	}

	r.stopPipeline()

	if err := r.startPipeline(seconds); err != nil {
		r.dead.Store(true)
		r.Dispose()
		return false
	}

	r.reader.ResetSegment()
	r.reader.SeekReset(seconds)
	r.readySegment.index = -1
	r.readySegment.complete = false
	r.readySegment.segment = Segment{}

	r.frozen.Store(false)
	r.eos.Store(false)
	r.setPosition(seconds)
	r.positionSeekSeconds = seconds
	return true
}

func (r *gstRunner) GetSegment(ctx context.Context, index int, audio int) (Segment, error) {
	if r.IsDead() {
		return Segment{}, ErrSegmentNotReady
	}

	if !r.statePlaying {
		r.statePlaying = true
		r.audioIndex = validAudioIndex(r.task.Probe, audio)
		if err := r.startPipeline(0); err != nil {
			r.dead.Store(true)
			r.Dispose()
			return Segment{}, err
		}
	} else if r.IsFrozen() {
		if !r.Seek(r.position()) {
			r.dead.Store(true)
			r.Dispose()
			return Segment{}, ErrSegmentNotReady
		}
	}

	if r.readySegment.index == index && r.readySegment.complete {
		return r.readySegment.segment, nil
	}

	r.reader.ResetSegment()
	r.readySegment.index = -1
	r.readySegment.complete = false
	r.readySegment.segment = Segment{}

	completed, err := r.reader.TryProcessDeferred()
	if err != nil {
		r.dead.Store(true)
		r.Dispose()
		return Segment{}, err
	}
	if completed && r.readySegment.complete {
		if index > 0 {
			r.readySegment.index = index
		} else {
			r.readySegment.index = 0
		}
		return r.readySegment.segment, nil
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return Segment{}, ctx.Err()
		}
		if r.IsDead() {
			return Segment{}, ErrSegmentNotReady
		}

		sample := gstRuntime.appSinkTryPullSample(r.sink, uint64(100*time.Millisecond))
		if sample == 0 {
			if gstRuntime.appSinkIsEOS(r.sink) {
				r.eos.Store(true)
				return Segment{}, ErrSegmentNotReady
			}
			continue
		}

		data := gstRuntime.sampleBytes(sample)
		gstRuntime.sampleUnref(sample)
		if len(data) == 0 {
			continue
		}

		if err := r.reader.Push(data); err != nil {
			r.dead.Store(true)
			r.Dispose()
			return Segment{}, err
		}

		if r.readySegment.complete {
			if index > 0 {
				r.readySegment.index = index
			} else {
				r.readySegment.index = 0
			}
			return r.readySegment.segment, nil
		}
	}

	return Segment{}, ErrSegmentNotReady
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
}

func (r *gstRunner) Frozen() {
	if r.pipeline == 0 {
		return
	}
	r.frozen.Store(true)
	r.stopPipeline()
}

func (r *gstRunner) IsDead() bool {
	return r.dead.Load()
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
