// Package profile описывает профили транскодирования.
package profile

import (
	"fmt"
	"strings"
	"time"

	"silo/internal/ffmpeg/args"
	"silo/internal/ffmpeg/media"
)

// Visibility определяет, кто видит профиль.
type Visibility string

const (
	VisibilityPrivate Visibility = "private" // Только владелец и владелец сервера
	VisibilityPublic  Visibility = "public"  // Все, кто прошёл по рангу
)

// Profile описывает параметры транскодирования для конкретного устройства.
type Profile struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	OwnerID   string        `json:"owner_id"`
	Protocol  args.Protocol `json:"protocol"`
	Container string        `json:"container"`

	Visibility   Visibility `json:"visibility"`
	RankRequired int        `json:"rank_required"` // Минимальный ранг для использования
	IsDefault    bool       `json:"is_default"`    // Профиль по умолчанию для всех

	Video     args.VideoOptions    `json:"video"`
	Audio     args.AudioOptions    `json:"audio"`
	Subtitles args.SubtitleOptions `json:"subtitles"`
	HLS       args.HLSOptions      `json:"hls"`

	// AllowVideoCopy разрешает оставить исходный видеокодек, если он совпадает с целевым
	AllowVideoCopy bool `json:"allow_video_copy"`
	// AllowAudioCopy разрешает оставить исходный аудиокодек, если он совпадает с целевым
	AllowAudioCopy bool `json:"allow_audio_copy"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DefaultID - идентификатор профиля, используемого по умолчанию.
const DefaultID = "default"

// Default возвращает встроенный профиль: h264 и aac в сегментах HLS.
// Сегментирование выбрано основным: оно даёт перемотку и адаптацию под скорость сети.
// Признак профиля по умолчанию не выставлен: его назначает владелец из своих профилей.
func Default() Profile {
	now := time.Now()
	return Profile{
		ID:             DefaultID,
		Name:           "Default",
		Protocol:       args.ProtocolHLS,
		Container:      "ts",
		Visibility:     VisibilityPublic,
		IsDefault:      false,
		Video:          args.DefaultVideoOptions(),
		Audio:          args.DefaultAudioOptions(),
		HLS:            args.DefaultHLSOptions(),
		AllowVideoCopy: true,
		AllowAudioCopy: true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Normalize приводит профиль к согласованному виду.
func (p *Profile) Normalize() {
	if p.Visibility != VisibilityPrivate && p.Visibility != VisibilityPublic {
		p.Visibility = VisibilityPrivate
	}
	if p.RankRequired < 0 {
		p.RankRequired = 0
	}
	if p.Protocol != args.ProtocolHLS && p.Protocol != args.ProtocolProgressive {
		p.Protocol = args.ProtocolProgressive
	}
	if p.HLS.SegmentLength <= 0 {
		p.HLS.SegmentLength = args.DefaultHLSOptions().SegmentLength
	}
	if p.HLS.SegmentType == "" {
		p.HLS.SegmentType = args.DefaultHLSOptions().SegmentType
	}
	if p.Video.Deinterlace == "" {
		p.Video.Deinterlace = args.DefaultVideoOptions().Deinterlace
	}
	if p.Subtitles.Mode == "" {
		p.Subtitles.Mode = args.SubtitleOff
	}
}

// Available сообщает, доступен ли профиль пользователю с указанным рангом.
func (p Profile) Available(owned bool, rank int) bool {
	if p.IsDefault || owned {
		return true
	}
	if p.Visibility != VisibilityPublic {
		return false
	}
	return rank >= p.RankRequired
}

// Validate проверяет, что профиль можно использовать.
func (p Profile) Validate() error {
	if p.Container == "" && p.Protocol != args.ProtocolHLS {
		return fmt.Errorf("container is required")
	}
	if p.Video.Codec == "" {
		return fmt.Errorf("video codec is required")
	}
	if p.Audio.Codec == "" {
		return fmt.Errorf("audio codec is required")
	}
	if p.Protocol == args.ProtocolHLS && p.HLS.SegmentLength <= 0 {
		return fmt.Errorf("segment length is required for hls")
	}
	return nil
}

// Resolve собирает задание для транскодирования с учётом разобранного источника.
func (p Profile) Resolve(input string, info *media.MediaInfo) (args.Options, error) {
	if err := p.Validate(); err != nil {
		return args.Options{}, err
	}

	o := args.Options{
		Protocol:     p.Protocol,
		Input:        input,
		AnalyzeDurMs: 5000,
		ProbeSizeMB:  5,
		Video:        p.Video,
		Audio:        p.Audio,
		Subtitles:    p.Subtitles,
		Container:    p.Container,
		HLS:          p.HLS,
		// Дорожки отбрасываются, если профиль не просит иного
		DropSubs: p.Subtitles.Mode == args.SubtitleOff,
	}

	if o.Protocol == args.ProtocolHLS {
		o.Output = ""
	} else {
		o.Output = "pipe:1"
	}

	if info != nil {
		o.SourceHDR = info.IsHDR()

		// Копирование оставляет диапазон источника как есть. Для HDR это
		// даёт блёклую картинку на обычном экране, поэтому такой материал
		// всегда перекодируется с приведением диапазона
		allowCopy := p.AllowVideoCopy
		if o.Video.Codec == args.CopyCodec {
			allowCopy = false
		}
		if p.Video.TonemapMode.Enabled(o.SourceHDR) && o.SourceHDR {
			allowCopy = false
		}

		o.Video = applySource(p.Video, info.Video(), allowCopy)
		o.Audio = applyAudio(p.Audio, info.DefaultAudio(), p.AllowAudioCopy)

		if !info.HasVideo() {
			o.Video.Codec = "none"
		}
		if !info.HasAudio() {
			o.Audio.Codec = "none"
		}
	}

	return o, nil
}

// applySource оставляет исходный видеокодек, когда он совпадает с целевым.
func applySource(target args.VideoOptions, src *media.Stream, allowCopy bool) args.VideoOptions {
	if !allowCopy || src == nil {
		return target
	}
	if target.Codec != args.CopyCodec && sameCodec(src.Codec, target.Codec) {
		target.Codec = args.CopyCodec
		target.Encoder = ""
	}
	return target
}

// applyAudio оставляет исходный аудиокодек, когда он совпадает с целевым.
func applyAudio(target args.AudioOptions, src *media.Stream, allowCopy bool) args.AudioOptions {
	if !allowCopy || src == nil {
		return target
	}
	if target.Codec != args.CopyCodec && sameCodec(src.Codec, target.Codec) {
		target.Codec = args.CopyCodec
		target.Encoder = ""
	}
	return target
}

// sameCodec сравнивает кодеки с учётом синонимов ffmpeg.
func sameCodec(source, target string) bool {
	source = strings.ToLower(strings.TrimSpace(source))
	target = strings.ToLower(strings.TrimSpace(target))
	if source == "" || target == "" {
		return false
	}
	if source == target {
		return true
	}

	aliases := map[string]string{
		"avc":  "h264",
		"avc1": "h264",
		"x264": "h264",
		"hevc": "h265",
		"hvc1": "h265",
		"hev1": "h265",
		"x265": "h265",
		"aac":  "aac",
		"mp4a": "aac",
	}
	if v, ok := aliases[source]; ok {
		source = v
	}
	if v, ok := aliases[target]; ok {
		target = v
	}
	return source == target
}

// Clone возвращает независимую копию профиля.
func (p Profile) Clone() Profile {
	o := p
	return o
}
