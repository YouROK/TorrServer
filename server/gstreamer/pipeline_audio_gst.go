//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"strconv"
	"strings"
)

// externalAudioParseName names the last element of the external audio branch before
// the muxer. A seek of that branch is sent to its source pad.
const externalAudioParseName = "ext_audio_parse"

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
	r.writeAudioReencode(sb, conf, &external.Track, externalAudioParseName)
}

// writeAudioReencode writes the chain that decodes an audio stream and encodes it to AAC
// for the MP4 muxer (up to and including the link to the muxer's audio pad). parseName,
// if not empty, becomes the name of the aacparse element.
func (r *gstRunner) writeAudioReencode(sb *strings.Builder, conf Config, track *TrackInfo, parseName string) {
	aacChannels := effectiveAACChannels(conf, track)
	aacSampleRate := effectiveAACSampleRate(conf, track)
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
	sb.WriteString(" ! aacparse")
	if parseName != "" {
		sb.WriteString(" name=")
		sb.WriteString(parseName)
	}
	sb.WriteString(" ! audio/mpeg,mpegversion=4,stream-format=raw,rate=")
	sb.WriteString(strconv.Itoa(aacSampleRate))
	sb.WriteString(",channels=")
	sb.WriteString(strconv.Itoa(aacChannels))
	sb.WriteString(" ! mux.audio_0 ")
}
