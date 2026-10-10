package probe

import (
	"strconv"
	"strings"

	"silo/internal/ffmpeg/media"
)

// normalize превращает результат ffprobe в доменную модель.
func normalize(res *result) *media.MediaInfo {
	info := &media.MediaInfo{
		FormatName: res.Format.FormatName,
		Container:  media.NormalizeContainer(res.Format.FormatName),
		Duration:   res.Format.Duration.Float64(),
		Size:       res.Format.Size.Int64(),
		Bitrate:    res.Format.BitRate.Int64(),
	}

	for _, s := range res.Streams {
		if s.Disposition.AttachedPic.Bool() {
			continue
		}
		stream := normalizeStream(s)
		if stream == nil {
			continue
		}
		info.Streams = append(info.Streams, *stream)
	}

	if info.Duration <= 0 {
		info.Duration = durationFromStreams(info.Streams)
	}
	if info.Duration <= 0 {
		info.Duration = parseDurationTag(res.Format.Tags.get("duration"))
	}

	estimateBitrates(info)
	return info
}

// normalizeStream переводит один поток ffprobe в доменную модель.
func normalizeStream(s streamInfo) *media.Stream {
	var kind media.Kind
	switch s.CodecType {
	case "video":
		kind = media.KindVideo
	case "audio":
		kind = media.KindAudio
	case "subtitle":
		kind = media.KindSubtitle
	case "data", "attachment":
		return nil
	default:
		return nil
	}

	if kind == media.KindVideo && (s.Width.Int() <= 0 || s.Height.Int() <= 0) {
		return nil
	}

	stream := &media.Stream{
		Index:         s.Index.Int(),
		Kind:          kind,
		Codec:         strings.ToLower(s.CodecName),
		Profile:       s.Profile,
		Level:         s.Level.Int(),
		Width:         s.Width.Int(),
		Height:        s.Height.Int(),
		PixFmt:        s.PixFmt,
		BitDepth:      bitDepth(s),
		Bitrate:       s.BitRate.Int64(),
		Channels:      s.Channels.Int(),
		ChannelLayout: s.ChannelLayout,
		SampleRate:    s.SampleRate.Int(),
		Language:      s.Tags.get("language"),
		Title:         s.Tags.get("title", "handler_name"),
		Default:       s.Disposition.Default.Bool(),
		Forced:        s.Disposition.Forced.Bool(),
		ColorRange:    s.ColorRange,
		ColorTransfer: s.ColorTransfer,
		ColorPrim:     s.ColorPrimaries,
		IsInterlaced:  isInterlaced(s.FieldOrder),
		RefFrames:     s.Refs.Int(),
		Duration:      s.Duration.Float64(),
	}

	rate := media.ParseFrames(s.AvgFrameRate)
	if rate <= 0 {
		rate = media.ParseFrames(s.RFrameRate)
	}
	stream.Framerate = rate

	if aspect := strings.TrimSpace(s.SampleAspect); aspect != "" && aspect != "0:1" && aspect != "1:1" {
		stream.IsAnamorphic = true
	}

	if kind == media.KindSubtitle {
		stream.IsTextSub = isTextSubtitle(stream.Codec)
	}

	return stream
}

// bitDepth определяет разрядность потока.
func bitDepth(s streamInfo) int {
	if v := s.BitsPerRawSample.Int(); v > 0 {
		return v
	}
	if v := s.BitsPerSample.Int(); v > 0 {
		return v
	}
	switch {
	case strings.Contains(s.PixFmt, "12"):
		return 12
	case strings.Contains(s.PixFmt, "10"):
		return 10
	case strings.Contains(s.PixFmt, "16"):
		return 16
	case strings.HasPrefix(s.PixFmt, "yuv") || strings.HasPrefix(s.PixFmt, "rgb"):
		return 8
	case s.SampleFmt == "s16" || s.SampleFmt == "u16":
		return 16
	case s.SampleFmt == "s32" || s.SampleFmt == "flt" || s.SampleFmt == "fltp":
		return 32
	}
	return 0
}

// isInterlaced определяет чересстрочную развёртку по порядку полей.
func isInterlaced(fieldOrder string) bool {
	switch strings.ToLower(fieldOrder) {
	case "", "unknown", "progressive":
		return false
	default:
		return true
	}
}

// isTextSubtitle показывает, является ли субтитр текстовым.
func isTextSubtitle(codec string) bool {
	switch codec {
	case "subrip", "srt", "ass", "ssa", "mov_text", "webvtt", "text", "eia_608", "subviewer", "microdvd":
		return true
	}
	return false
}

// durationFromStreams берёт максимальную длительность среди потоков.
func durationFromStreams(streams []media.Stream) float64 {
	var max float64
	for _, s := range streams {
		if s.Duration > max {
			max = s.Duration
		}
	}
	return max
}

// parseDurationTag разбирает тег duration формата "00:12:34.567".
func parseDurationTag(v string) float64 {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f
	}

	parts := strings.Split(v, ":")
	if len(parts) != 3 {
		return 0
	}
	hours, err1 := strconv.ParseFloat(parts[0], 64)
	minutes, err2 := strconv.ParseFloat(parts[1], 64)
	seconds, err3 := strconv.ParseFloat(strings.Replace(parts[2], ",", ".", 1), 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0
	}
	return hours*3600 + minutes*60 + seconds
}

// estimateBitrates восстанавливает отсутствующие битрейты по данным контейнера.
func estimateBitrates(info *media.MediaInfo) {
	if info.Duration <= 0 || info.Size <= 0 {
		return
	}

	total := int64(0)
	unknown := 0
	for i := range info.Streams {
		if info.Streams[i].Bitrate > 0 {
			total += info.Streams[i].Bitrate
		} else {
			unknown++
		}
	}

	if info.Bitrate <= 0 {
		info.Bitrate = int64(float64(info.Size) * 8 / info.Duration)
	}

	if unknown == 0 || info.Bitrate <= 0 {
		return
	}

	rest := info.Bitrate - total
	if rest <= 0 {
		return
	}

	perStream := rest / int64(unknown)
	for i := range info.Streams {
		if info.Streams[i].Bitrate <= 0 {
			info.Streams[i].Bitrate = perStream
		}
	}
}
