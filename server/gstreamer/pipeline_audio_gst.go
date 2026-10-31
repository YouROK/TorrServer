//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"strconv"
	"strings"
)

// writeAudioReencode writes the chain that decodes an audio stream and encodes it to AAC
// for the MP4 muxer (up to and including the link to the muxer's audio pad).
func (r *gstRunner) writeAudioReencode(sb *strings.Builder, conf Config, track *TrackInfo) {
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
	sb.WriteString(" ! aacparse ! audio/mpeg,mpegversion=4,stream-format=raw,rate=")
	sb.WriteString(strconv.Itoa(aacSampleRate))
	sb.WriteString(",channels=")
	sb.WriteString(strconv.Itoa(aacChannels))
	sb.WriteString(" ! mux.audio_0 ")
}
