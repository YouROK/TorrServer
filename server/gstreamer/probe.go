package gstreamer

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"server/ffprobe"

	probedata "gopkg.in/vansante/go-ffprobe.v2"
)

const gstProbeTimeout = 30 * time.Second

type ProbeInfo struct {
	DurationNS int64
	Tracks     []TrackInfo
}

type TrackInfo struct {
	Index   int
	PadName string

	Type     string
	CapsName string

	Title    string
	Language string

	Width    int
	Height   int
	Channels int
	Rate     int

	FrameRateNum int
	FrameRateDen int
}

func (p ProbeInfo) DurationSeconds() int {
	if p.DurationNS <= 0 {
		return 0
	}
	return int(float64(p.DurationNS) / 1_000_000_000.0)
}

func (p ProbeInfo) Video() *TrackInfo {
	for i := range p.Tracks {
		t := &p.Tracks[i]
		if t.Type == "video" ||
			t.CapsName == "video/x-h264" ||
			t.CapsName == "video/x-h265" ||
			t.CapsName == "video/x-av1" ||
			t.CapsName == "video/x-vp9" ||
			t.CapsName == "video/x-vp8" {
			return t
		}
	}
	return nil
}

func (p ProbeInfo) VideoCapsName() string {
	if v := p.Video(); v != nil {
		return v.CapsName
	}
	return ""
}

func (p ProbeInfo) IsH264() bool { return p.VideoCapsName() == "video/x-h264" }
func (p ProbeInfo) IsH265() bool { return p.VideoCapsName() == "video/x-h265" }
func (p ProbeInfo) IsAV1() bool  { return p.VideoCapsName() == "video/x-av1" }
func (p ProbeInfo) IsVP9() bool  { return p.VideoCapsName() == "video/x-vp9" }
func (p ProbeInfo) IsVP8() bool  { return p.VideoCapsName() == "video/x-vp8" }

func probeSource(sourceURL string) (ProbeInfo, error) {
	data, err := ffprobe.ProbeUrlWithTimeout(sourceURL, gstProbeTimeout)
	if err != nil {
		return ProbeInfo{}, err
	}
	return probeFromFFProbe(data), nil
}

func probeFromFFProbe(data *probedata.ProbeData) ProbeInfo {
	var probe ProbeInfo
	if data == nil {
		return probe
	}

	if data.Format != nil && data.Format.DurationSeconds > 0 {
		probe.DurationNS = int64(data.Format.DurationSeconds * 1_000_000_000)
	}

	videoIndex := 0
	audioIndex := 0

	for _, stream := range data.Streams {
		if stream == nil {
			continue
		}

		switch stream.CodecType {
		case "video":
			track := TrackInfo{
				Index:        videoIndex,
				PadName:      "video_" + strconv.Itoa(videoIndex),
				Type:         "video",
				CapsName:     codecToCapsName("video", stream.CodecName, stream.CodecLongName),
				Width:        stream.Width,
				Height:       stream.Height,
				Title:        tagString(stream.TagList, "title"),
				Language:     tagString(stream.TagList, "language"),
				FrameRateNum: 0,
				FrameRateDen: 0,
			}
			track.FrameRateNum, track.FrameRateDen = parseRate(firstNonEmpty(stream.AvgFrameRate, stream.RFrameRate))
			probe.Tracks = append(probe.Tracks, track)
			videoIndex++

		case "audio":
			rate, _ := strconv.Atoi(stream.SampleRate)
			track := TrackInfo{
				Index:    audioIndex,
				PadName:  "audio_" + strconv.Itoa(audioIndex),
				Type:     "audio",
				CapsName: codecToCapsName("audio", stream.CodecName, stream.CodecLongName),
				Channels: stream.Channels,
				Rate:     rate,
				Title:    tagString(stream.TagList, "title"),
				Language: tagString(stream.TagList, "language"),
			}
			probe.Tracks = append(probe.Tracks, track)
			audioIndex++
		}
	}

	return probe
}

func codecToCapsName(kind string, values ...string) string {
	codec := strings.ToLower(strings.Join(values, " "))
	if codec == "" {
		return ""
	}

	if kind == "video" {
		switch {
		case strings.Contains(codec, "h264") || strings.Contains(codec, "h.264") || strings.Contains(codec, "avc"):
			return "video/x-h264"
		case strings.Contains(codec, "hevc") || strings.Contains(codec, "h265") || strings.Contains(codec, "h.265"):
			return "video/x-h265"
		case strings.Contains(codec, "av1"):
			return "video/x-av1"
		case strings.Contains(codec, "vp9"):
			return "video/x-vp9"
		case strings.Contains(codec, "vp8"):
			return "video/x-vp8"
		}
	}

	if kind == "audio" {
		switch {
		case strings.Contains(codec, "eac3") || strings.Contains(codec, "e-ac-3") || strings.Contains(codec, "e-ac3"):
			return "audio/x-eac3"
		case strings.Contains(codec, "ac3") || strings.Contains(codec, "ac-3") || strings.Contains(codec, "a/52"):
			return "audio/x-ac3"
		case strings.Contains(codec, "aac"):
			return "audio/mpeg"
		case strings.Contains(codec, "opus"):
			return "audio/x-opus"
		case strings.Contains(codec, "vorbis"):
			return "audio/x-vorbis"
		case strings.Contains(codec, "flac"):
			return "audio/x-flac"
		case strings.Contains(codec, "mpeg") || strings.Contains(codec, "mp3"):
			return "audio/mpeg"
		}
	}

	return ""
}

func parseRate(value string) (int, int) {
	value = strings.TrimSpace(value)
	if value == "" || value == "0/0" {
		return 0, 0
	}

	parts := strings.SplitN(value, "/", 2)
	num, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || num <= 0 {
		return 0, 0
	}

	den := 1
	if len(parts) == 2 {
		den, err = strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || den <= 0 {
			return 0, 0
		}
	}

	if math.IsInf(float64(num)/float64(den), 0) {
		return 0, 0
	}

	return num, den
}

func tagString(tags probedata.Tags, key string) string {
	if tags == nil {
		return ""
	}
	if value, err := tags.GetString(key); err == nil {
		return value
	}
	for k, raw := range tags {
		if strings.EqualFold(k, key) {
			return strings.TrimSpace(strings.Trim(rawString(raw), `"`))
		}
	}
	return ""
}

func rawString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
