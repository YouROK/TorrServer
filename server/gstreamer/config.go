//go:build gst

package gstreamer

import (
	"encoding/json"
	"errors"
	"runtime"
	"strings"
	"time"

	"server/settings"
)

const gstreamerSettingsKey = "gstreamer"

const minGSTVersion = 1.22

type Config struct {
	GSTVersion float64 `json:"GSTVersion"`
	GSTPath    string  `json:"GSTPath"`
	Source     string  `json:"Source"`
	MaxTasks   int     `json:"MaxTasks"`

	InactiveMinutes int `json:"InactiveMinutes"`

	AACBitrateKbps int  `json:"AACBitrateKbps"`
	AACChannels    int  `json:"AACChannels"`
	AACSamplerate  int  `json:"AACSamplerate"`
	SegmentSeconds int  `json:"SegmentSeconds"`
	SegmentDiff    int  `json:"SegmentDiff"`
	Subtitles      bool `json:"Subtitles"`

	TranscodeH264        bool `json:"TranscodeH264"`
	TranscodeH265        bool `json:"TranscodeH265"`
	TranscodeAV1         bool `json:"TranscodeAV1"`
	TranscodeVP9         bool `json:"TranscodeVP9"`
	TranscodeVP8         bool `json:"TranscodeVP8"`
	TranscodeAVI         bool `json:"TranscodeAVI"`
	HDRToSDR             bool `json:"HDRToSDR"`
	HardwareAcceleration bool `json:"HardwareAcceleration"`
	UseGPU               bool `json:"UseGPU"`
	X264Ultrafast        bool `json:"X264Ultrafast"`
	VideoBitrate         int  `json:"VideoBitrate"`
}

func DefaultConfig() Config {
	return applySettingsConfig(defaultConfigWithoutSettings()).normalized()
}

func defaultConfigWithoutSettings() Config {
	conf := Config{
		GSTVersion:           minGSTVersion,
		Source:               "stream",
		InactiveMinutes:      5,
		AACBitrateKbps:       256,
		SegmentSeconds:       6,
		SegmentDiff:          20,
		Subtitles:            true,
		HardwareAcceleration: true,
		UseGPU:               true,
		VideoBitrate:         10_000,
	}

	if runtime.GOOS == "windows" {
		conf.GSTVersion = 1.28
		conf.GSTPath = `C:\Program Files\gstreamer\1.0\mingw_x86_64`
	}

	return conf
}

func (c Config) normalized() Config {
	if c.InactiveMinutes <= 0 {
		c.InactiveMinutes = 5
	}
	if c.AACBitrateKbps <= 0 {
		c.AACBitrateKbps = 256
	}
	if c.AACChannels < 0 {
		c.AACChannels = 0
	}
	if c.AACSamplerate < 0 {
		c.AACSamplerate = 0
	}
	if c.MaxTasks < 0 {
		c.MaxTasks = 0
	}
	if c.SegmentSeconds <= 0 {
		c.SegmentSeconds = 6
	}
	if c.SegmentDiff < 0 {
		c.SegmentDiff = 0
	}
	if c.VideoBitrate <= 0 {
		c.VideoBitrate = 10_000
	}
	if c.GSTVersion < minGSTVersion {
		c.GSTVersion = minGSTVersion
	}
	c.Source = strings.ToLower(strings.TrimSpace(c.Source))
	if c.Source != "play" {
		c.Source = "stream"
	}
	return c
}

func (c Config) inactiveDuration() time.Duration {
	return time.Duration(c.normalized().InactiveMinutes) * time.Minute
}

type storedConfig struct {
	GSTVersion *float64
	GSTPath    *string
	Source     *string
	MaxTasks   *int

	InactiveMinutes *int

	AACBitrateKbps *int
	AACChannels    *int
	AACSamplerate  *int
	SegmentSeconds *int
	SegmentDiff    *int
	Subtitles      *bool

	TranscodeH264        *bool
	TranscodeH265        *bool
	TranscodeAV1         *bool
	TranscodeVP9         *bool
	TranscodeVP8         *bool
	TranscodeAVI         *bool
	HDRToSDR             *bool
	HardwareAcceleration *bool
	UseGPU               *bool
	X264Ultrafast        *bool
	VideoBitrate         *int
}

func applySettingsConfig(conf Config) Config {
	if settings.Path == "" {
		return conf
	}

	db := settings.NewJsonDB()
	if db == nil {
		return conf
	}

	var data []byte
	for _, name := range []string{gstreamerSettingsKey, "GStreamer", "gst"} {
		data = db.Get("Settings", name)
		if len(data) > 0 {
			break
		}
	}
	if len(data) == 0 {
		return conf
	}

	var stored storedConfig
	if err := json.Unmarshal(data, &stored); err != nil {
		return conf
	}

	if stored.GSTVersion != nil {
		conf.GSTVersion = *stored.GSTVersion
	}
	if stored.GSTPath != nil {
		conf.GSTPath = *stored.GSTPath
	}
	if stored.Source != nil {
		conf.Source = *stored.Source
	}
	if stored.MaxTasks != nil {
		conf.MaxTasks = *stored.MaxTasks
	}
	if stored.InactiveMinutes != nil {
		conf.InactiveMinutes = *stored.InactiveMinutes
	}
	if stored.AACBitrateKbps != nil {
		conf.AACBitrateKbps = *stored.AACBitrateKbps
	}
	if stored.AACChannels != nil {
		conf.AACChannels = *stored.AACChannels
	}
	if stored.AACSamplerate != nil {
		conf.AACSamplerate = *stored.AACSamplerate
	}
	if stored.SegmentSeconds != nil {
		conf.SegmentSeconds = *stored.SegmentSeconds
	}
	if stored.SegmentDiff != nil {
		conf.SegmentDiff = *stored.SegmentDiff
	}
	if stored.Subtitles != nil {
		conf.Subtitles = *stored.Subtitles
	}
	if stored.TranscodeH264 != nil {
		conf.TranscodeH264 = *stored.TranscodeH264
	}
	if stored.TranscodeH265 != nil {
		conf.TranscodeH265 = *stored.TranscodeH265
	}
	if stored.TranscodeAV1 != nil {
		conf.TranscodeAV1 = *stored.TranscodeAV1
	}
	if stored.TranscodeVP9 != nil {
		conf.TranscodeVP9 = *stored.TranscodeVP9
	}
	if stored.TranscodeVP8 != nil {
		conf.TranscodeVP8 = *stored.TranscodeVP8
	}
	if stored.TranscodeAVI != nil {
		conf.TranscodeAVI = *stored.TranscodeAVI
	}
	if stored.HDRToSDR != nil {
		conf.HDRToSDR = *stored.HDRToSDR
	}
	if stored.HardwareAcceleration != nil {
		conf.HardwareAcceleration = *stored.HardwareAcceleration
	}
	if stored.UseGPU != nil {
		conf.UseGPU = *stored.UseGPU
	}
	if stored.X264Ultrafast != nil {
		conf.X264Ultrafast = *stored.X264Ultrafast
	}
	if stored.VideoBitrate != nil {
		conf.VideoBitrate = *stored.VideoBitrate
	}
	return conf
}

func SaveConfig(conf Config) error {
	if settings.ReadOnly {
		return errors.New("read-only mode")
	}
	if settings.Path == "" {
		return errors.New("settings path is not configured")
	}

	conf = conf.normalized()
	db := settings.NewJsonDB()
	if db == nil {
		return errors.New("json db unavailable")
	}

	data, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	db.Set("Settings", gstreamerSettingsKey, data)
	db.Rem("Settings", "gst")
	db.Rem("Settings", "GStreamer")
	return nil
}
