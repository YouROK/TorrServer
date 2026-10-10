package probe

import (
	"encoding/json"
	"strconv"
	"strings"
)

// flexInt принимает число как в виде числа, так и в виде строки.
type flexInt int64

func (f *flexInt) UnmarshalJSON(data []byte) error {
	s := strings.Trim(strings.TrimSpace(string(data)), `"`)
	if s == "" || s == "N/A" || s == "null" {
		*f = 0
		return nil
	}
	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		*f = flexInt(v)
		return nil
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		*f = flexInt(int64(v))
		return nil
	}
	*f = 0
	return nil
}

func (f flexInt) Int64() int64 { return int64(f) }
func (f flexInt) Int() int     { return int(f) }

// flexFloat принимает дробное число как в виде числа, так и в виде строки.
type flexFloat float64

func (f *flexFloat) UnmarshalJSON(data []byte) error {
	s := strings.Trim(strings.TrimSpace(string(data)), `"`)
	if s == "" || s == "N/A" || s == "null" {
		*f = 0
		return nil
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		*f = flexFloat(v)
		return nil
	}
	*f = 0
	return nil
}

func (f flexFloat) Float64() float64 { return float64(f) }

// flexBool принимает флаг как 0/1, "0"/"1" или как настоящее логическое значение.
type flexBool bool

func (f *flexBool) UnmarshalJSON(data []byte) error {
	s := strings.Trim(strings.TrimSpace(string(data)), `"`)
	switch s {
	case "1", "true", "True", "yes":
		*f = true
	default:
		*f = false
	}
	return nil
}

func (f flexBool) Bool() bool { return bool(f) }

type result struct {
	Streams  []streamInfo `json:"streams"`
	Format   formatInfo   `json:"format"`
	Chapters []chapter    `json:"chapters"`
}

type chapter struct {
	StartTime flexFloat `json:"start_time"`
	EndTime   flexFloat `json:"end_time"`
	Tags      tagsMap   `json:"tags"`
}

type formatInfo struct {
	Filename   string    `json:"filename"`
	NbStreams  flexInt   `json:"nb_streams"`
	FormatName string    `json:"format_name"`
	StartTime  flexFloat `json:"start_time"`
	Duration   flexFloat `json:"duration"`
	Size       flexInt   `json:"size"`
	BitRate    flexInt   `json:"bit_rate"`
	Tags       tagsMap   `json:"tags"`
}

type streamInfo struct {
	Index            flexInt   `json:"index"`
	CodecName        string    `json:"codec_name"`
	CodecLongName    string    `json:"codec_long_name"`
	CodecType        string    `json:"codec_type"`
	Profile          string    `json:"profile"`
	Level            flexInt   `json:"level"`
	Width            flexInt   `json:"width"`
	Height           flexInt   `json:"height"`
	CodedWidth       flexInt   `json:"coded_width"`
	CodedHeight      flexInt   `json:"coded_height"`
	PixFmt           string    `json:"pix_fmt"`
	BitsPerRawSample flexInt   `json:"bits_per_raw_sample"`
	BitsPerSample    flexInt   `json:"bits_per_sample"`
	RFrameRate       string    `json:"r_frame_rate"`
	AvgFrameRate     string    `json:"avg_frame_rate"`
	BitRate          flexInt   `json:"bit_rate"`
	Channels         flexInt   `json:"channels"`
	ChannelLayout    string    `json:"channel_layout"`
	SampleRate       flexInt   `json:"sample_rate"`
	SampleFmt        string    `json:"sample_fmt"`
	DisplayAspect    string    `json:"display_aspect_ratio"`
	SampleAspect     string    `json:"sample_aspect_ratio"`
	ColorRange       string    `json:"color_range"`
	ColorSpace       string    `json:"color_space"`
	ColorTransfer    string    `json:"color_transfer"`
	ColorPrimaries   string    `json:"color_primaries"`
	FieldOrder       string    `json:"field_order"`
	Refs             flexInt   `json:"refs"`
	Duration         flexFloat `json:"duration"`
	HasBFrames       flexInt   `json:"has_b_frames"`
	Disposition      struct {
		Default         flexBool `json:"default"`
		Forced          flexBool `json:"forced"`
		AttachedPic     flexBool `json:"attached_pic"`
		HearingImpaired flexBool `json:"hearing_impaired"`
	} `json:"disposition"`
	Tags tagsMap `json:"tags"`
}

// tagsMap хранит теги без учёта регистра ключей.
type tagsMap map[string]string

func (t *tagsMap) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	res := make(map[string]string, len(raw))
	for k, v := range raw {
		switch val := v.(type) {
		case string:
			res[strings.ToLower(k)] = val
		case nil:
			res[strings.ToLower(k)] = ""
		default:
			res[strings.ToLower(k)] = strings.Trim(strings.TrimSpace(string(mustJSON(val))), `"`)
		}
	}
	*t = res
	return nil
}

func (t tagsMap) get(keys ...string) string {
	for _, k := range keys {
		if v, ok := t[strings.ToLower(k)]; ok && v != "" {
			return v
		}
	}
	return ""
}

func mustJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return data
}
