package media

import (
	"fmt"
	"strconv"
	"strings"
)

// Kind описывает тип потока в контейнере.
type Kind string

const (
	KindVideo    Kind = "video"
	KindAudio    Kind = "audio"
	KindSubtitle Kind = "subtitle"
)

// Stream описывает один поток медиафайла.
type Stream struct {
	Index         int     `json:"index"`
	Kind          Kind    `json:"kind"`
	Codec         string  `json:"codec"`
	Profile       string  `json:"profile,omitempty"`
	Level         int     `json:"level,omitempty"`
	Width         int     `json:"width,omitempty"`
	Height        int     `json:"height,omitempty"`
	PixFmt        string  `json:"pix_fmt,omitempty"`
	BitDepth      int     `json:"bit_depth,omitempty"`
	Framerate     float64 `json:"framerate,omitempty"`
	Bitrate       int64   `json:"bitrate,omitempty"`
	Channels      int     `json:"channels,omitempty"`
	ChannelLayout string  `json:"channel_layout,omitempty"`
	SampleRate    int     `json:"sample_rate,omitempty"`
	Language      string  `json:"language,omitempty"`
	Title         string  `json:"title,omitempty"`
	Default       bool    `json:"default"`
	Forced        bool    `json:"forced"`
	IsTextSub     bool    `json:"is_text_subtitle,omitempty"`
	ColorRange    string  `json:"color_range,omitempty"`
	ColorTransfer string  `json:"color_transfer,omitempty"`
	ColorPrim     string  `json:"color_primaries,omitempty"`
	IsInterlaced  bool    `json:"is_interlaced,omitempty"`
	IsAnamorphic  bool    `json:"is_anamorphic,omitempty"`
	RefFrames     int     `json:"ref_frames,omitempty"`
	Duration      float64 `json:"duration,omitempty"`
}

// hdrTransfers перечисляет признаки расширенного динамического диапазона.
// PQ используется в HDR10 и Dolby Vision, HLG - в телевещании.
var hdrTransfers = map[string]bool{
	"smpte2084":    true,
	"arib-std-b67": true,
}

// IsHDR сообщает, что поток содержит расширенный динамический диапазон.
// Такой материал нельзя показывать как SDR: без приведения диапазона
// цвета получаются блёклыми, а яркость завышенной.
func (s Stream) IsHDR() bool {
	if hdrTransfers[strings.ToLower(s.ColorTransfer)] {
		return true
	}

	// Десять бит на канал само по себе не признак HDR,
	// но вместе с широкой цветовой охватностью указывает на него
	if s.BitDepth >= 10 && strings.EqualFold(s.ColorPrim, "bt2020") {
		return true
	}
	return false
}

// MediaInfo описывает разобранный медиафайл целиком.
type MediaInfo struct {
	FormatName string   `json:"format_name"`
	Container  string   `json:"container"`
	Duration   float64  `json:"duration"`
	Size       int64    `json:"size"`
	Bitrate    int64    `json:"bitrate"`
	Streams    []Stream `json:"streams"`
}

// IsHDR сообщает, что видео в файле содержит расширенный динамический диапазон.
func (m *MediaInfo) IsHDR() bool {
	if m == nil {
		return false
	}
	v := m.Video()
	return v != nil && v.IsHDR()
}

// Video возвращает первый видеопоток или nil.
func (m *MediaInfo) Video() *Stream {
	return m.firstOf(KindVideo)
}

// Audio возвращает первый аудиопоток или nil.
func (m *MediaInfo) Audio() *Stream {
	return m.firstOf(KindAudio)
}

// DefaultAudio выбирает аудиопоток по умолчанию: помеченный default, иначе первый.
func (m *MediaInfo) DefaultAudio() *Stream {
	var fallback *Stream
	for i := range m.Streams {
		if m.Streams[i].Kind != KindAudio {
			continue
		}
		if fallback == nil {
			fallback = &m.Streams[i]
		}
		if m.Streams[i].Default {
			return &m.Streams[i]
		}
	}
	return fallback
}

// Count возвращает число потоков указанного типа.
func (m *MediaInfo) Count(k Kind) int {
	n := 0
	for _, s := range m.Streams {
		if s.Kind == k {
			n++
		}
	}
	return n
}

// HasVideo сообщает, есть ли в файле видеопоток.
func (m *MediaInfo) HasVideo() bool { return m.Count(KindVideo) > 0 }

// HasAudio сообщает, есть ли в файле аудиопоток.
func (m *MediaInfo) HasAudio() bool { return m.Count(KindAudio) > 0 }

func (m *MediaInfo) firstOf(k Kind) *Stream {
	for i := range m.Streams {
		if m.Streams[i].Kind == k {
			return &m.Streams[i]
		}
	}
	return nil
}

// ParseRate разбирает дробную запись частоты кадров вида "30000/1001".
func ParseRate(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "0/0" {
		return 0
	}
	num, den, ok := strings.Cut(s, "/")
	if !ok {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return v
	}
	n, err1 := strconv.ParseFloat(strings.TrimSpace(num), 64)
	d, err2 := strconv.ParseFloat(strings.TrimSpace(den), 64)
	if err1 != nil || err2 != nil || d == 0 {
		return 0
	}
	return n / d
}

// ParseFrames разбирает строку частоты кадров и округляет её до двух знаков.
func ParseFrames(s string) float64 {
	v := ParseRate(s)
	if v < 0 {
		return 0
	}
	return v
}

// NormalizeContainer приводит имя контейнера к общепринятому короткому виду.
func NormalizeContainer(format string) string {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		return ""
	}

	containers := map[string]string{
		"matroska":  "mkv",
		"webm":      "webm",
		"mov":       "mp4",
		"mp4":       "mp4",
		"m4a":       "mp4",
		"m4v":       "mp4",
		"3gp":       "mp4",
		"3g2":       "mp4",
		"mj2":       "mp4",
		"mpegts":    "ts",
		"mp3":       "mp3",
		"flac":      "flac",
		"ogg":       "ogg",
		"wav":       "wav",
		"aac":       "aac",
		"ac3":       "ac3",
		"avi":       "avi",
		"asf":       "asf",
		"flv":       "flv",
		"mpeg":      "mpg",
		"mpegvideo": "mpg",
		"image2":    "image",
	}

	parts := strings.Split(format, ",")
	for _, p := range parts {
		if v, ok := containers[strings.TrimSpace(p)]; ok {
			return v
		}
	}
	return strings.TrimSpace(parts[0])
}

// FormatDuration приводит секунды к читаемому виду.
func FormatDuration(seconds float64) string {
	if seconds <= 0 {
		return "unknown"
	}
	total := int64(seconds)
	return fmt.Sprintf("%02d:%02d:%02d", total/3600, (total%3600)/60, total%60)
}
