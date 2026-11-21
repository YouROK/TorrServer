//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// externalAudioSeekElement names the first element after the decoder in the external
// audio branch (audioconvert). A seek of that branch is sent to its source pad, which
// is the audio counterpart of the video seek sent to the source pad of the multiqueue.
const externalAudioSeekElement = "ext_audio_seek"

// externalAudioSeekFlags are the flags of the external audio seek. Every audio frame is
// a key frame, so the audio can start exactly where the video starts.
const externalAudioSeekFlags = gstSeekFlagFlush | gstSeekFlagAccurate

// resolveAudioIndex is the audio index the runner of the task works with: the requested
// external one, if this task was created for it, otherwise the nearest embedded track.
func resolveAudioIndex(task *Task, requested int) int {
	if task.ExternalAudio != nil && isExternalAudio(requested) && requested == task.Audio {
		return requested
	}
	return validAudioIndex(task.Probe, requested)
}

// externalAudioActive reports whether the pipeline takes its audio from a separate file.
func (r *gstRunner) externalAudioActive() bool {
	return r.task.ExternalAudio != nil && isExternalAudio(r.audioIndex)
}

// writeExternalAudioBranch writes a second source, which reads the audio file, and the
// chain that turns it into AAC for the muxer.
func (r *gstRunner) writeExternalAudioBranch(sb *strings.Builder, conf Config, gstVersion gstVersionInfo) {
	external := r.task.ExternalAudio
	sb.WriteString("souphttpsrc location=\"")
	sb.WriteString(external.SourceURL)
	sb.WriteString("\" is-live=false keep-alive=true timeout=60 retries=5 ")
	if gstVersion.atLeast(1, 26) {
		sb.WriteString("retry-backoff-factor=0.5 retry-backoff-max=10 ")
	}
	sb.WriteString("! ")
	r.writeAudioReencode(sb, conf, &external.Track, externalAudioSeekElement)
}

// writeAudioReencode writes the chain that decodes an audio stream and encodes it to AAC
// for the MP4 muxer (up to and including the link to the muxer's audio pad). seekName,
// if not empty, becomes the name of the audioconvert element, whose source pad receives
// the seek events of the branch.
func (r *gstRunner) writeAudioReencode(sb *strings.Builder, conf Config, track *TrackInfo, seekName string) {
	aacChannels := effectiveAACChannels(conf, track)
	aacSampleRate := effectiveAACSampleRate(conf, track)
	aacBitrate := conf.AACBitrateKbps * 1000
	if aacChannels > 2 {
		aacBitrate *= 2
	}

	sb.WriteString("decodebin ! audioconvert")
	if seekName != "" {
		sb.WriteString(" name=")
		sb.WriteString(seekName)
	}
	sb.WriteString(" dithering=none noise-shaping=none ! audioresample quality=2 sinc-filter-mode=full ! audio/x-raw,format=")
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

// externalAudioStartNS is the position the external audio has to be seeked to after the
// video was seeked: the time the video really starts at, if the start probe has seen it,
// otherwise the pipeline position measured after the seek.
func (r *gstRunner) externalAudioStartNS(fallbackSeconds float64) int64 {
	if registration := r.videoStartProbe; registration != nil {
		if state, ok := registration.state.(*videoStartProbeState); ok {
			if actual := state.actualNS.Load(); actual != gstClockTimeNone && actual <= math.MaxInt64 {
				return int64(actual)
			}
		}
	}
	return int64(math.Round(fallbackSeconds * 1_000_000_000))
}

// sendExternalAudioSeekEvent sends a seek event to the external audio branch.
func sendExternalAudioSeekEvent(pipeline uintptr, flags int32, positionNS int64) error {
	pad := gstPipelineElementPad(gstRuntime, pipeline, externalAudioSeekElement, "src")
	if pad == 0 {
		return errors.New("external audio pad is not available")
	}
	defer gstRuntime.objectUnref(pad)

	if !gstRuntime.sendTimeSeekEvent(pad, flags, positionNS) {
		return errors.New("external audio seek event returned false")
	}
	return nil
}

// seekExternalAudio seeks the external audio branch of a paused pipeline and waits until
// the pipeline has prerolled again.
func (r *gstRunner) seekExternalAudio(pipeline uintptr, bus uintptr, positionNS int64) error {
	if positionNS < 0 {
		return errors.New("external audio seek position is negative")
	}
	if err := gstRuntime.popBusError(bus, 0); err != nil {
		return err
	}
	gstRuntime.drainBusMessages(bus, gstMessageAsyncDone|gstMessageEOS)
	if err := sendExternalAudioSeekEvent(pipeline, externalAudioSeekFlags, positionNS); err != nil {
		return err
	}
	asyncDone, err := gstRuntime.waitForSeekDone(bus, pipelineStateTimeout)
	if err != nil {
		return fmt.Errorf("seek did not finish: %w", err)
	}
	if !asyncDone {
		gstTaskDebugf(r.task, "external audio seek ASYNC_DONE not observed; validating pipeline state")
	}

	result := gstRuntime.elementGetState(pipeline, pipelineStateTimeout)
	if result == gstStateChangeSuccess || result == gstStateChangeNoPreroll {
		return nil
	}
	if err := gstRuntime.popBusError(bus, 0); err != nil {
		return err
	}
	return fmt.Errorf("pipeline state=%d after the seek", result)
}
