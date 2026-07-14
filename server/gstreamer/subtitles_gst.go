//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	subtitleExtractTimeout = 15 * time.Second
	subtitleStateTimeout   = 5 * time.Second
	subtitleSegmentOverlap = 500 * time.Millisecond
)

func extractSubtitleCues(ctx context.Context, task *Task, track TrackInfo, start time.Duration, end time.Duration) ([]subtitleCue, error) {
	gstInitOnce.Do(func() {
		initGStreamerRuntime(task.Config)
	})
	if gstInitErr != nil {
		return nil, errors.Join(ErrPipelineDisabled, gstInitErr)
	}
	if gstRuntime == nil {
		return nil, ErrPipelineDisabled
	}

	seekStart := start - subtitleSegmentOverlap
	if seekStart < 0 {
		seekStart = 0
	}
	readUntil := end + subtitleSegmentOverlap

	pipeline, err := gstRuntime.parseLaunch(buildSubtitlePipeline(task, track))
	if err != nil {
		return nil, err
	}

	sink := gstRuntime.binGetByName(pipeline, "subout")
	demux := gstRuntime.binGetByName(pipeline, "d")
	bus := gstRuntime.pipelineGetBus(pipeline)

	cleanup := func() {
		gstRuntime.elementSetState(pipeline, gstStateNull)
		if sink != 0 {
			gstRuntime.objectUnref(sink)
		}
		if demux != 0 {
			gstRuntime.objectUnref(demux)
		}
		if bus != 0 {
			gstRuntime.objectUnref(bus)
		}
		gstRuntime.objectUnref(pipeline)
	}
	defer cleanup()

	if sink == 0 {
		return nil, errors.New("subtitle appsink element is not available")
	}
	if demux == 0 {
		return nil, errors.New("subtitle matroskademux element is not available")
	}

	if err := setSubtitlePipelineState(pipeline, bus, gstStatePaused); err != nil {
		return nil, err
	}

	seekNS := int64(math.Round(float64(seekStart.Nanoseconds())))
	if !gstRuntime.elementSeekSimple(demux, gstFormatTime, gstSeekFlagFlush|gstSeekFlagSnapBefore, seekNS) {
		return nil, errors.New("subtitle seek failed")
	}

	if err := setSubtitlePipelineState(pipeline, bus, gstStatePlaying); err != nil {
		return nil, err
	}

	deadline := time.Now().Add(subtitleExtractTimeout)
	var cues []subtitleCue
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := gstRuntime.popBusError(bus, 0); err != nil {
			return nil, err
		}

		sample := gstRuntime.appSinkTryPullSample(sink, uint64(100*time.Millisecond))
		if sample == 0 {
			if gstRuntime.appSinkIsEOS(sink) {
				return cues, nil
			}
			continue
		}

		var sampleCues []subtitleCue
		err := gstRuntime.withSampleBuffer(sample, func(meta gstBufferSnapshot, data []byte) error {
			cueStart, ok := subtitleBufferTime(meta)
			if !ok {
				cueStart = seekStart
			}
			duration := subtitleBufferDuration(meta)
			sampleCues = subtitleCuesFromSample(track, data, cueStart, duration)
			return nil
		})
		gstRuntime.sampleUnref(sample)
		if err != nil {
			return nil, err
		}

		for _, cue := range sampleCues {
			if cue.Start >= readUntil {
				return cues, nil
			}
			if cue.End > start && cue.Start < end {
				cues = append(cues, cue)
			}
		}
	}

	if err := gstRuntime.popBusError(bus, 0); err != nil {
		return nil, err
	}
	return cues, nil
}

func buildSubtitlePipeline(task *Task, track TrackInfo) string {
	conf := task.Config.normalized()
	gstVersion := effectiveGStreamerVersion(conf)

	var sb strings.Builder
	sb.WriteString("souphttpsrc location=\"")
	sb.WriteString(task.SourceURL)
	sb.WriteString("\" is-live=false keep-alive=true timeout=60 retries=5 ")
	if gstVersion.atLeast(1, 26) {
		sb.WriteString("retry-backoff-factor=0.5 retry-backoff-max=10 ")
	}
	sb.WriteString("! matroskademux name=d ")
	sb.WriteString("d.subtitle_")
	sb.WriteString(strconv.Itoa(track.Index))
	sb.WriteString(" ! queue max-size-buffers=0 max-size-bytes=0 max-size-time=0 ")
	if subtitleNeedsSSAParse(track) {
		sb.WriteString("! ssaparse ")
	}
	sb.WriteString("! webvttenc ")
	sb.WriteString("! appsink name=subout emit-signals=false sync=false max-buffers=1000 drop=false wait-on-eos=false")
	return sb.String()
}

func setSubtitlePipelineState(pipeline uintptr, bus uintptr, state int32) error {
	setResult := gstRuntime.elementSetState(pipeline, state)
	if setResult == gstStateChangeFailure {
		if err := gstRuntime.popBusError(bus, 0); err != nil {
			return err
		}
		return fmt.Errorf("gstreamer failed to request subtitle state %d", state)
	}

	waitResult := gstRuntime.elementGetState(pipeline, subtitleStateTimeout)
	switch waitResult {
	case gstStateChangeSuccess, gstStateChangeNoPreroll:
		return nil
	case gstStateChangeAsync:
		if err := gstRuntime.popBusError(bus, 0); err != nil {
			return err
		}
		return fmt.Errorf("gstreamer subtitle state %d timed out", state)
	case gstStateChangeFailure:
		if err := gstRuntime.popBusError(bus, 0); err != nil {
			return err
		}
		return fmt.Errorf("gstreamer subtitle state %d failed", state)
	default:
		return fmt.Errorf("unexpected subtitle GstStateChangeReturn=%d for state=%d", waitResult, state)
	}
}

func subtitleBufferTime(meta gstBufferSnapshot) (time.Duration, bool) {
	if gstClockTimeIsValid(meta.pts) {
		return time.Duration(meta.pts), true
	}
	if gstClockTimeIsValid(meta.dts) {
		return time.Duration(meta.dts), true
	}
	return 0, false
}

func subtitleBufferDuration(meta gstBufferSnapshot) time.Duration {
	if gstClockTimeIsValid(meta.duration) {
		return time.Duration(meta.duration)
	}
	return 0
}

func subtitleNeedsSSAParse(track TrackInfo) bool {
	codec := strings.ToLower(track.Codec + " " + track.CapsName)
	return strings.Contains(codec, "ass") || strings.Contains(codec, "ssa")
}
