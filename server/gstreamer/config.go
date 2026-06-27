package gstreamer

import (
	"encoding/json"
	"runtime"
	"strings"
	"time"

	"server/settings"
)

type Config struct {
	GSTVersion  float64
	GSTPath     string
	Source      string
	AppSinkMode string

	InactiveMinutes int

	AACBitrateKbps int
	SegmentSeconds int

	TranscodeH264 bool
	TranscodeH265 bool
	TranscodeAV1  bool
	TranscodeVP9  bool
	VideoBitrate  int

	PipelineTimeSeconds int
	PipelineAudioQueue  int
	PipelineVideoQueue  int

	TempFS     bool
	TempFSRing int
}

func DefaultConfig() Config {
	conf := Config{
		GSTVersion:          1.22,
		Source:              "stream",
		AppSinkMode:         "bytes",
		InactiveMinutes:     5,
		AACBitrateKbps:      256,
		SegmentSeconds:      6,
		VideoBitrate:        10_000,
		PipelineTimeSeconds: 18,
		PipelineAudioQueue:  3,
		PipelineVideoQueue:  24,
		TempFS:              false,
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
	if c.VideoBitrate <= 0 {
		c.VideoBitrate = 10_000
	}
	if c.PipelineTimeSeconds <= 0 {
		c.PipelineTimeSeconds = 18
	}
	if c.PipelineAudioQueue <= 0 {
		c.PipelineAudioQueue = 3
	}
	if c.PipelineVideoQueue <= 0 {
		c.PipelineVideoQueue = 24
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
	c.AppSinkMode = strings.ToLower(strings.TrimSpace(c.AppSinkMode))
	switch c.AppSinkMode {
	case "", "bytes", "max-bytes", "max-size-bytes":
		c.AppSinkMode = "bytes"
	case "buffer", "buffers", "max-buffers":
		c.AppSinkMode = "buffers"
	default:
		c.AppSinkMode = "bytes"
	}
	return c
}

func (c Config) inactiveDuration() time.Duration {
	return time.Duration(c.normalized().InactiveMinutes) * time.Minute
}

type storedConfig struct {
	GSTVersion  *float64
	GSTPath     *string
	Source      *string
	AppSinkMode *string

	InactiveMinutes *int

	AACBitrateKbps *int
	SegmentSeconds *int

	TranscodeH264 *bool
	TranscodeH265 *bool
	TranscodeAV1  *bool
	TranscodeVP9  *bool
	VideoBitrate  *int

	PipelineTimeSeconds *int
	PipelineAudioQueue  *int
	PipelineVideoQueue  *int

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
	if stored.AppSinkMode != nil {
		conf.AppSinkMode = *stored.AppSinkMode
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
	if stored.PipelineTimeSeconds != nil {
		conf.PipelineTimeSeconds = *stored.PipelineTimeSeconds
	}
	if stored.PipelineAudioQueue != nil {
		conf.PipelineAudioQueue = *stored.PipelineAudioQueue
	}
	if stored.PipelineVideoQueue != nil {
		conf.PipelineVideoQueue = *stored.PipelineVideoQueue
	}
	if stored.TempFS != nil {
		conf.TempFS = *stored.TempFS
	}
	if stored.TempFSRing != nil {
		conf.TempFSRing = *stored.TempFSRing
	}

	return conf
}
