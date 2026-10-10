package args

import "strings"

// Protocol задаёт способ доставки транскодированного потока.
type Protocol string

const (
	ProtocolProgressive Protocol = "progressive"
	ProtocolHLS         Protocol = "hls"
)

// CopyCodec означает копирование потока без перекодирования.
const CopyCodec = "copy"

// VideoOptions описывает параметры видеопотока.
type VideoOptions struct {
	Codec          string  `json:"codec"`        // copy, h264, hevc, av1, vp9
	Encoder        string  `json:"encoder"`      // Пусто - подбирается автоматически
	HWAccel        string  `json:"hwaccel"`      // none, vaapi, nvenc
	CRF            int     `json:"crf"`          // Используется, когда BitrateKbps равен нулю
	BitrateKbps    int     `json:"bitrate_kbps"` // Целевой битрейт
	MaxBitrateKbps int     `json:"max_bitrate_kbps"`
	Preset         string  `json:"preset"`  // ultrafast .. placebo
	Profile        string  `json:"profile"` // Базовый, main, high
	Level          string  `json:"level"`   // Например, 4.0
	PixFmt         string  `json:"pix_fmt"`
	MaxWidth       int     `json:"max_width"`     // 0 - без ограничения
	MaxHeight      int     `json:"max_height"`    // 0 - без ограничения
	MaxFramerate   float64 `json:"max_framerate"` // 0 - без ограничения
	Deinterlace    string  `json:"deinterlace"`   // off, auto, yadif
	StreamIndex    *int    `json:"stream_index"`  // Явный номер потока, nil - первый подходящий
	// HWDevice - путь к устройству ускорения для VAAPI, пусто - узел по умолчанию
	HWDevice string `json:"hw_device,omitempty"`
	// HWDecode включает декодирование на устройстве, а не только кодирование
	HWDecode bool `json:"hw_decode,omitempty"`
	// Transpose поворачивает видео: 90, 180, 270
	Transpose string `json:"transpose,omitempty"`
	// TonemapMode управляет приведением HDR к SDR
	TonemapMode TonemapMode `json:"tonemap_mode,omitempty"`
	// TonemapAlgorithm задаёт алгоритм отображения: hable, mobius, reinhard, clip
	TonemapAlgorithm string `json:"tonemap_algorithm,omitempty"`
}

// AudioOptions описывает параметры аудиопотока.
type AudioOptions struct {
	Codec       string `json:"codec"`   // copy, aac, opus, mp3, ac3, flac
	Encoder     string `json:"encoder"` // Пусто - подбирается автоматически
	BitrateKbps int    `json:"bitrate_kbps"`
	Channels    int    `json:"channels"`     // 0 - как в источнике
	SampleRate  int    `json:"sample_rate"`  // 0 - как в источнике
	StreamIndex *int   `json:"stream_index"` // Явный номер потока, nil - первый подходящий
}

// HLSOptions описывает параметры сегментирования.
type HLSOptions struct {
	SegmentLength   int    `json:"segment_length"`   // Длительность сегмента в секундах
	SegmentType     string `json:"segment_type"`     // mpegts, fmp4
	ListSize        int    `json:"list_size"`        // Сколько сегментов держать в плейлисте
	DeleteThreshold int    `json:"delete_threshold"` // Сколько сегментов держать после выпадения из плейлиста
	StartNumber     int    `json:"start_number"`     // Номер первого сегмента
	PlaylistType    string `json:"playlist_type"`    // vod, event, пусто - скользящее окно
	SegmentFilename string `json:"segment_filename"` // Шаблон с %d
	// InitFilename - имя начального сегмента fmp4 с описанием дорожек
	InitFilename string `json:"init_filename,omitempty"`
}

// DefaultInitSegment - имя начального сегмента fmp4 по умолчанию.
const DefaultInitSegment = "init.mp4"

// Options описывает задание на транскодирование.
type Options struct {
	Protocol Protocol `json:"protocol"`

	Input        string   `json:"input"`
	InputArgs    []string `json:"input_args"` // Параметры, добавляемые до -i
	Seek         float64  `json:"seek"`       // Позиция старта в секундах
	AnalyzeDurMs int      `json:"analyze_dur_ms"`
	ProbeSizeMB  int      `json:"probe_size_mb"`
	Realtime     bool     `json:"realtime"` // Читать вход с естественной скоростью

	Video     VideoOptions    `json:"video"`
	Audio     AudioOptions    `json:"audio"`
	Subtitles SubtitleOptions `json:"subtitles"`

	Container string     `json:"container"` // mp4, mkv, ts, webm
	Threads   int        `json:"threads"`   // 0 - авто
	CopyTS    bool       `json:"copy_ts"`
	Output    string     `json:"output"` // Путь или pipe:1
	HLS       HLSOptions `json:"hls"`
	// SourceHDR показывает, что источник содержит расширенный динамический диапазон.
	// Нужен для автоматического приведения диапазона.
	SourceHDR bool     `json:"source_hdr,omitempty"`
	DropSubs  bool     `json:"drop_subs"`
	ExtraArgs []string `json:"extra_args"`
}

// DefaultVideoOptions возвращает разумные параметры видеокодирования.
func DefaultVideoOptions() VideoOptions {
	return VideoOptions{
		Codec:       "h264",
		CRF:         23,
		Preset:      "veryfast",
		Deinterlace: "auto",
	}
}

// DefaultAudioOptions возвращает разумные параметры аудиокодирования.
func DefaultAudioOptions() AudioOptions {
	return AudioOptions{
		Codec:       "aac",
		BitrateKbps: 192,
	}
}

// DefaultHLSOptions возвращает параметры сегментирования по умолчанию.
func DefaultHLSOptions() HLSOptions {
	return HLSOptions{
		SegmentLength:   4,
		SegmentType:     "mpegts",
		ListSize:        0,
		DeleteThreshold: 1,
	}
}

// DefaultOptions собирает тестовое задание: h264 в mp4 с аудио aac.
func DefaultOptions(input string) Options {
	return Options{
		Protocol:     ProtocolProgressive,
		Input:        input,
		AnalyzeDurMs: 5000,
		ProbeSizeMB:  5,
		Video:        DefaultVideoOptions(),
		Audio:        DefaultAudioOptions(),
		Container:    "mp4",
		Output:       "pipe:1",
		Subtitles:    SubtitleOptions{Mode: SubtitleOff},
		DropSubs:     true,
	}
}

// TonemapMode задаёт способ приведения расширенного диапазона к обычному.
type TonemapMode string

const (
	// TonemapAuto включает приведение только для HDR-материала
	TonemapAuto TonemapMode = "auto"
	// TonemapOn включает приведение всегда
	TonemapOn TonemapMode = "on"
	// TonemapOff отключает приведение
	TonemapOff TonemapMode = "off"
)

// TonemapEnabled сообщает, нужно ли приводить диапазон для указанного источника.
func (m TonemapMode) Enabled(sourceIsHDR bool) bool {
	switch m {
	case TonemapOn:
		return true
	case TonemapOff:
		return false
	default:
		return sourceIsHDR
	}
}

// ParseTonemapMode разбирает значение режима, неизвестное означает авто.
func ParseTonemapMode(raw string) TonemapMode {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "on", "true", "yes", "1":
		return TonemapOn
	case "off", "false", "no", "0":
		return TonemapOff
	default:
		return TonemapAuto
	}
}

// SubtitleMode задаёт способ обработки субтитров.
type SubtitleMode string

const (
	SubtitleOff  SubtitleMode = "off"  // Субтитры не попадают в результат
	SubtitleCopy SubtitleMode = "copy" // Дорожки копируются как есть
	SubtitleBurn SubtitleMode = "burn" // Субтитры впечатываются в видео
)

// SubtitleOptions описывает обработку субтитров.
type SubtitleOptions struct {
	Mode SubtitleMode `json:"mode"`
	// StreamIndex задаёт номер дорожки в источнике, nil - первая подходящая
	StreamIndex *int `json:"stream_index,omitempty"`
	// ForceStyle включает переопределение стиля при впечатывании
	ForceStyle string `json:"force_style,omitempty"`
}
