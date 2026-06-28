package gstreamer

import (
	"encoding/json"
	"runtime"
	"strings"
	"time"

	"server/settings"
)

type Config struct {
	GSTVersion float64
	GSTPath    string
	Source     string

	InactiveMinutes int

	AACBitrateKbps int
	SegmentSeconds int
	AppSinkBuffers int

	TranscodeH264 bool
	TranscodeH265 bool
	TranscodeAV1  bool
	TranscodeVP9  bool
	VideoBitrate  int

	TempFS     bool
	TempFSRing int
}

func DefaultConfig() Config {
	conf := Config{
		GSTVersion:      1.22,
		Source:          "stream",
		InactiveMinutes: 5,
		AACBitrateKbps:  256,
		SegmentSeconds:  6,
		AppSinkBuffers:  1000,
		VideoBitrate:    10_000,
		TempFS:          false,
	}

	if runtime.GOOS == "windows" {
		conf.GSTVersion = 1.28
		conf.GSTPath = `C:\Program Files\gstreamer\1.0\mingw_x86_64`
	}

	return applySettingsConfig(conf).normalized()
}

func (c Config) normalized() Config {
	if c.InactiveMinutes <= 0 {
		c.InactiveMinutes = 5
	}
	if c.AACBitrateKbps <= 0 {
		c.AACBitrateKbps = 256
	}
	if c.SegmentSeconds <= 0 {
		c.SegmentSeconds = 6
	}
	if c.AppSinkBuffers <= 0 {
		c.AppSinkBuffers = 1000
	}
	if c.VideoBitrate <= 0 {
		c.VideoBitrate = 10_000
	}
	if c.TempFSRing < 0 {
		c.TempFSRing = 0
	}
	if c.GSTVersion <= 0 {
		c.GSTVersion = 1.22
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

	InactiveMinutes *int

	AACBitrateKbps *int
	SegmentSeconds *int
	AppSinkBuffers *int `json:"appsinkBuffers"`

	TranscodeH264 *bool
	TranscodeH265 *bool
	TranscodeAV1  *bool
	TranscodeVP9  *bool
	VideoBitrate  *int

	TempFS     *bool `json:"tempfs"`
	TempFSRing *int  `json:"tempfs_ring"`
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
	for _, name := range []string{"gst", "GStreamer"} {
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
	if stored.InactiveMinutes != nil {
		conf.InactiveMinutes = *stored.InactiveMinutes
	}
	if stored.AACBitrateKbps != nil {
		conf.AACBitrateKbps = *stored.AACBitrateKbps
	}
	if stored.SegmentSeconds != nil {
		conf.SegmentSeconds = *stored.SegmentSeconds
	}
	if stored.AppSinkBuffers != nil {
		conf.AppSinkBuffers = *stored.AppSinkBuffers
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
	if stored.VideoBitrate != nil {
		conf.VideoBitrate = *stored.VideoBitrate
	}
	if stored.TempFS != nil {
		conf.TempFS = *stored.TempFS
	}
	if stored.TempFSRing != nil {
		conf.TempFSRing = *stored.TempFSRing
	}

	return conf
}
