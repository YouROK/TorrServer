//go:build gst

package gstreamer

import (
	"os"
	"path/filepath"
	"testing"

	"server/settings"
)

func TestNormalizedConfigDefaultsToGStreamer122(t *testing.T) {
	conf := Config{}.normalized()

	if conf.GSTVersion != 1.22 {
		t.Fatalf("GSTVersion=%v, want 1.22", conf.GSTVersion)
	}
	if conf.MaxTasks != 0 {
		t.Fatalf("MaxTasks=%v, want 0 unlimited", conf.MaxTasks)
	}
	if conf.AACChannels != 0 {
		t.Fatalf("AACChannels=%v, want 0 auto", conf.AACChannels)
	}
	if conf.AACSamplerate != 0 {
		t.Fatalf("AACSamplerate=%v, want 0 auto", conf.AACSamplerate)
	}
}

func TestPlatformDefaultsUseReleasePipelinePolicy(t *testing.T) {
	conf := defaultConfigWithoutSettings().normalized()
	if conf.SegmentDiff != 20 || !conf.Subtitles || !conf.HardwareAcceleration || !conf.UseGPU {
		t.Fatalf("unexpected release defaults: %#v", conf)
	}
}

func TestDefaultConfigReadsGstSettingsJSON(t *testing.T) {
	oldPath := settings.Path
	settings.Path = t.TempDir()
	t.Cleanup(func() {
		settings.Path = oldPath
	})

	data := []byte(`{
  "gst": {
    "gstVersion": 1.28,
    "gstPath": "D:\\gst",
    "source": "play",
    "maxTasks": 3,
    "inactiveMinutes": 3,
    "aacBitrateKbps": 192,
    "AACChannels": 6,
    "AACSamplerate": 44100,
    "segmentSeconds": 4,
    "transcodeH264": true,
    "videoBitrate": 8000
  }
}`)

	if err := os.WriteFile(filepath.Join(settings.Path, "settings.json"), data, 0o666); err != nil {
		t.Fatal(err)
	}

	conf := DefaultConfig()

	if conf.GSTVersion != 1.28 ||
		conf.GSTPath != `D:\gst` ||
		conf.Source != "play" ||
		conf.MaxTasks != 3 ||
		conf.InactiveMinutes != 3 ||
		conf.AACBitrateKbps != 192 ||
		conf.AACChannels != 6 ||
		conf.AACSamplerate != 44100 ||
		conf.SegmentSeconds != 4 ||
		!conf.TranscodeH264 ||
		conf.VideoBitrate != 8000 {
		t.Fatalf("DefaultConfig() did not read gst settings: %#v", conf)
	}
}

func TestSourceURLDefaultsToStream(t *testing.T) {
	oldPort := settings.Port
	settings.Port = "8090"
	t.Cleanup(func() {
		settings.Port = oldPort
	})

	got := sourceURL(Config{}.normalized(), "abc def", "1")
	want := "http://127.0.0.1:8090/stream/?link=abc+def&index=1&play"
	if got != want {
		t.Fatalf("sourceURL(default) = %q, want %q", got, want)
	}
}

func TestSourceURLCanUsePlayEndpoint(t *testing.T) {
	oldPort := settings.Port
	settings.Port = "8090"
	t.Cleanup(func() {
		settings.Port = oldPort
	})

	got := sourceURL(Config{Source: "play"}.normalized(), "abc def", "1")
	want := "http://127.0.0.1:8090/play/abc%20def/1"
	if got != want {
		t.Fatalf("sourceURL(play) = %q, want %q", got, want)
	}
}
