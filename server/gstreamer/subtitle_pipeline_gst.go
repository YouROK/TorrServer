//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (r *gstRunner) writeSubtitleBranches(sb *strings.Builder) {
	for _, track := range subtitleTracksForPlaylist(r.task.Probe) {
		padName := track.PadName
		if padName == "" {
			padName = "subtitle_" + strconv.Itoa(track.Index)
		}

		sb.WriteString("d.")
		sb.WriteString(padName)
		sb.WriteString(" ! queue max-size-buffers=0 max-size-bytes=0 max-size-time=0 ")
		if subtitleNeedsSSAParse(track) {
			sb.WriteString("! ssaparse ")
		}
		sb.WriteString("! webvttenc ! appsink name=subs_")
		sb.WriteString(strconv.Itoa(track.Index))
		sb.WriteString(" emit-signals=false sync=false async=false max-buffers=256 drop=false wait-on-eos=false ")
	}
}

func (r *gstRunner) lookupSubtitleSinks(pipeline uintptr) (map[int]uintptr, error) {
	tracks := subtitleTracksForPlaylist(r.task.Probe)
	if len(tracks) == 0 {
		return nil, nil
	}

	sinks := make(map[int]uintptr, len(tracks))
	for _, track := range tracks {
		name := "subs_" + strconv.Itoa(track.Index)
		sink := gstRuntime.binGetByName(pipeline, name)
		if sink == 0 {
			for _, acquired := range sinks {
				gstRuntime.objectUnref(acquired)
			}
			return nil, fmt.Errorf("subtitle appsink %s is not available", name)
		}
		sinks[track.Index] = sink
	}
	return sinks, nil
}

func (r *gstRunner) drainSubtitleSinks() {
	if len(r.subtitleSinks) == 0 || gstRuntime == nil {
		return
	}

	for subtitleIndex, sink := range r.subtitleSinks {
		track := r.task.Probe.SubtitleTrack(subtitleIndex)
		if track == nil {
			continue
		}

		samples := 0
		cuesAdded := 0
		var collected []subtitleCue
		for {
			sample := gstRuntime.appSinkTryPullSample(sink, 0)
			if sample == 0 {
				break
			}
			samples++

			var sampleCues []subtitleCue
			var rawCueStart time.Duration
			var absoluteCueStart time.Duration
			var cueDuration time.Duration
			var hasTimestamp bool
			var sampleBytes int
			err := gstRuntime.withSampleBuffer(sample, func(meta gstBufferSnapshot, data []byte) error {
				sampleBytes = len(data)
				rawCueStart, hasTimestamp = subtitleBufferTime(meta)
				if !hasTimestamp {
					rawCueStart = time.Duration(r.position() * float64(time.Second))
				}
				absoluteCueStart = r.absoluteSubtitleTime(rawCueStart)
				cueDuration = subtitleBufferDuration(meta)
				sampleCues = subtitleCuesFromSample(*track, data, absoluteCueStart, cueDuration)
				return nil
			})
			gstRuntime.sampleUnref(sample)
			if err != nil {
				gstErrorf("subtitle appsink sample failed task=%s file=%s subtitle=%d err=%v", r.task.ID, r.task.FileID, subtitleIndex, err)
				continue
			}
			if len(sampleCues) > 0 {
				collected = append(collected, sampleCues...)
				cuesAdded += len(sampleCues)
				gstDebugf("subtitle cue sample task=%s file=%s subtitle=%d timestamp=%t rawStart=%.3f absoluteStart=%.3f duration=%.3f cueStart=%.3f cueEnd=%.3f runnerPosition=%.3f seekBase=%.3f bytes=%d", r.task.ID, r.task.FileID, subtitleIndex, hasTimestamp, rawCueStart.Seconds(), absoluteCueStart.Seconds(), cueDuration.Seconds(), sampleCues[0].Start.Seconds(), sampleCues[0].End.Seconds(), r.position(), r.positionSeekSeconds, sampleBytes)
			}
		}

		if len(collected) > 0 {
			r.task.addSubtitleCues(subtitleIndex, collected)
		}
		if samples > 0 {
			gstDebugf("subtitle appsink drained task=%s file=%s subtitle=%d samples=%d cues=%d", r.task.ID, r.task.FileID, subtitleIndex, samples, cuesAdded)
		}
	}
}

func (r *gstRunner) absoluteSubtitleTime(value time.Duration) time.Duration {
	seekBase := time.Duration(r.positionSeekSeconds * float64(time.Second))
	if seekBase <= 0 {
		return value
	}

	tolerance := time.Duration(r.task.Config.normalized().SegmentSeconds*2) * time.Second
	if value < seekBase-tolerance {
		return value + seekBase
	}
	return value
}
