//go:build gst

package gstreamer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"server/settings"
)

func useSettingsDir(t *testing.T) string {
	t.Helper()
	oldPath, oldRO := settings.Path, settings.ReadOnly
	dir := t.TempDir()
	settings.Path, settings.ReadOnly = dir, false
	t.Cleanup(func() { settings.Path, settings.ReadOnly = oldPath, oldRO })
	return dir
}

func TestSaveConfigRoundTripsThroughDefaultConfig(t *testing.T) {
	useSettingsDir(t)

	want := DefaultConfig()
	want.GSTVersion = 1.26
	want.GSTPath = "/opt/gst"
	want.Source = "play"
	want.MaxTasks = 2
	want.InactiveMinutes = 9
	want.AACBitrateKbps = 128
	want.AACChannels = 2
	want.AACSamplerate = 48000
	want.SegmentSeconds = 4
	want.SegmentDiff = 7
	want.Subtitles = !want.Subtitles
	want.TranscodeH264, want.TranscodeH265, want.TranscodeAV1 = true, true, true
	want.TranscodeVP9, want.TranscodeVP8, want.TranscodeAVI = true, true, true
	want.HDRToSDR, want.X264Ultrafast = true, true
	want.HardwareAcceleration, want.UseGPU = false, false
	want.VideoBitrate = 4321

	if err := SaveConfig(want); err != nil {
		t.Fatal(err)
	}
	if got := DefaultConfig(); got != want {
		t.Fatalf("round trip:\n got %+v\nwant %+v", got, want)
	}
}

func TestSaveConfigNormalizesAndDropsLegacyKeys(t *testing.T) {
	dir := useSettingsDir(t)
	legacy := `{"gst":{"maxTasks":9},"GStreamer":{"maxTasks":8}}`
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := DefaultConfig().MaxTasks; got != 8 {
		t.Fatalf("legacy keys: MaxTasks=%d, want 8 (GStreamer wins over gst)", got)
	}

	err := SaveConfig(Config{Source: " PLAY ", SegmentDiff: -3, VideoBitrate: -1, GSTVersion: 1.0, MaxTasks: -4})
	if err != nil {
		t.Fatal(err)
	}
	got := DefaultConfig()
	if got.Source != "play" || got.SegmentDiff != 0 || got.VideoBitrate != 10_000 || got.GSTVersion != minGSTVersion || got.MaxTasks != 0 {
		t.Fatalf("not normalized on save: %+v", got)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if s := string(raw); strings.Contains(s, `"gst"`) || strings.Contains(s, `"GStreamer"`) {
		t.Fatalf("legacy keys survived SaveConfig: %s", s)
	}
}

func TestSaveConfigRefusals(t *testing.T) {
	dir := useSettingsDir(t)

	settings.ReadOnly = true
	if err := SaveConfig(Config{}); err == nil {
		t.Error("read-only mode must refuse")
	}
	if _, err := os.Stat(filepath.Join(dir, "settings.json")); err == nil {
		t.Error("read-only SaveConfig wrote a file")
	}

	settings.ReadOnly = false
	settings.Path = ""
	if err := SaveConfig(Config{}); err == nil {
		t.Error("empty settings path must refuse")
	}
}

func TestApplySettingsConfigPartialOverrideKeepsDefaults(t *testing.T) {
	dir := useSettingsDir(t)
	if err := os.WriteFile(filepath.Join(dir, "settings.json"),
		[]byte(`{"gstreamer":{"MaxTasks":4,"Subtitles":false}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got := DefaultConfig()
	def := defaultConfigWithoutSettings().normalized()
	if got.MaxTasks != 4 || got.Subtitles {
		t.Fatalf("stored fields not applied: %+v", got)
	}
	got.MaxTasks, got.Subtitles = def.MaxTasks, def.Subtitles
	if got != def {
		t.Fatalf("unrelated fields changed:\n got %+v\nwant %+v", got, def)
	}
}

func TestApplySettingsConfigIgnoresCorruptJSON(t *testing.T) {
	dir := useSettingsDir(t)
	if err := os.WriteFile(filepath.Join(dir, "settings.json"),
		[]byte(`{"gstreamer":{"MaxTasks":"many"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, want := DefaultConfig(), defaultConfigWithoutSettings().normalized(); got != want {
		t.Fatalf("corrupt stored config leaked into defaults: %+v", got)
	}
}

func TestNormalizedClampsAndInactiveDuration(t *testing.T) {
	c := Config{InactiveMinutes: -1, AACBitrateKbps: 0, AACChannels: -2, AACSamplerate: -1, SegmentSeconds: 0, Source: "Garbage"}.normalized()
	if c.InactiveMinutes != 5 || c.AACBitrateKbps != 256 || c.AACChannels != 0 || c.AACSamplerate != 0 || c.SegmentSeconds != 6 || c.Source != "stream" {
		t.Fatalf("clamps: %+v", c)
	}
	if got := (Config{InactiveMinutes: 3}).inactiveDuration(); got != 3*time.Minute {
		t.Errorf("inactiveDuration = %v", got)
	}
	if got := (Config{}).inactiveDuration(); got != 5*time.Minute {
		t.Errorf("default inactiveDuration = %v", got)
	}
}
