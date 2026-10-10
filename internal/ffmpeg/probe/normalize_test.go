package probe

import (
	"os"
	"path/filepath"
	"testing"

	"silo/internal/ffmpeg/media"
)

func loadFixture(t *testing.T, name string) *media.MediaInfo {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	res, err := decode(data)
	if err != nil {
		t.Fatalf("failed to decode fixture: %v", err)
	}
	return normalize(res)
}

func TestNormalizeMatroska(t *testing.T) {
	info := loadFixture(t, "mkv_av.json")

	if info.Container != "mkv" {
		t.Errorf("container = %q, want %q", info.Container, "mkv")
	}
	if info.Duration < 12 || info.Duration > 12.1 {
		t.Errorf("duration = %v, want about 12", info.Duration)
	}
	if !info.HasVideo() || !info.HasAudio() {
		t.Fatalf("expected video and audio streams, got %d", len(info.Streams))
	}

	video := info.Video()
	if video.Codec != "h264" {
		t.Errorf("video codec = %q, want h264", video.Codec)
	}
	if video.Width != 640 || video.Height != 480 {
		t.Errorf("resolution = %dx%d, want 640x480", video.Width, video.Height)
	}
	if video.Framerate < 29.9 || video.Framerate > 30.0 {
		t.Errorf("framerate = %v, want about 29.97", video.Framerate)
	}
	if video.BitDepth != 8 {
		t.Errorf("bit depth = %d, want 8", video.BitDepth)
	}

	audio := info.Audio()
	if audio.Codec != "aac" {
		t.Errorf("audio codec = %q, want aac", audio.Codec)
	}
	if audio.Channels != 1 {
		t.Errorf("audio channels = %d, want 1", audio.Channels)
	}
	if audio.SampleRate != 44100 {
		t.Errorf("sample rate = %d, want 44100", audio.SampleRate)
	}
	if audio.Language != "eng" {
		t.Errorf("audio language = %q, want eng", audio.Language)
	}
}

func TestNormalizeMp4(t *testing.T) {
	info := loadFixture(t, "mp4_v.json")

	if info.Container != "mp4" {
		t.Errorf("container = %q, want mp4", info.Container)
	}
	if !info.HasVideo() || info.HasAudio() {
		t.Errorf("expected video only, got %d streams", len(info.Streams))
	}
	if info.Video().Framerate != 25 {
		t.Errorf("framerate = %v, want 25", info.Video().Framerate)
	}
}

func TestNormalizeAudioOnly(t *testing.T) {
	info := loadFixture(t, "mp3_a.json")

	if info.Container != "mp3" {
		t.Errorf("container = %q, want mp3", info.Container)
	}
	if info.HasVideo() {
		t.Error("mp3 must not report a video stream")
	}
	if !info.HasAudio() {
		t.Fatal("expected an audio stream")
	}
	if info.Duration < 2.9 || info.Duration > 3.1 {
		t.Errorf("duration = %v, want about 3", info.Duration)
	}
}

func TestNormalizeMpegts(t *testing.T) {
	info := loadFixture(t, "ts_v.json")

	if info.Container != "ts" {
		t.Errorf("container = %q, want ts", info.Container)
	}
	if !info.HasVideo() {
		t.Fatal("expected a video stream")
	}
	if got := info.Count(media.KindVideo); got != 1 {
		t.Errorf("video streams = %d, want 1", got)
	}
}

func TestNormalizeMultipleStreams(t *testing.T) {
	info := loadFixture(t, "mkv_multi.json")

	if got := len(info.Streams); got != 2 {
		t.Fatalf("streams = %d, want 2", got)
	}
	if got := info.Count(media.KindVideo); got != 1 {
		t.Errorf("video streams = %d, want 1", got)
	}
	if got := info.Count(media.KindAudio); got != 1 {
		t.Errorf("audio streams = %d, want 1", got)
	}
	if info.DefaultAudio() == nil {
		t.Error("default audio stream was not picked")
	}
}

func TestNormalizeBitrateEstimation(t *testing.T) {
	info := loadFixture(t, "mkv_av.json")

	if info.Bitrate <= 0 {
		t.Fatalf("container bitrate = %d, want positive", info.Bitrate)
	}
	for _, s := range info.Streams {
		if s.Bitrate <= 0 {
			t.Errorf("stream %d (%s) bitrate = %d, want positive", s.Index, s.Kind, s.Bitrate)
		}
	}
}

func TestNormalizeAttachedPictureSkipped(t *testing.T) {
	raw := `{
		"streams": [
			{"index": 0, "codec_name": "mjpeg", "codec_type": "video", "width": 600, "height": 600,
			 "disposition": {"attached_pic": 1}},
			{"index": 1, "codec_name": "mp3", "codec_type": "audio", "channels": 2, "sample_rate": 44100}
		],
		"format": {"format_name": "mp3", "duration": "10.0"}
	}`

	res, err := decode([]byte(raw))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	info := normalize(res)

	if info.HasVideo() {
		t.Error("attached picture must not be treated as a video stream")
	}
	if got := len(info.Streams); got != 1 {
		t.Errorf("streams = %d, want 1", got)
	}
}

func TestNormalizeStringNumbers(t *testing.T) {
	raw := `{
		"streams": [
			{"index": "0", "codec_name": "h264", "codec_type": "video", "width": "1920", "height": "1080",
			 "avg_frame_rate": "24000/1001", "bit_rate": "5000000",
			 "disposition": {"default": "1"}}
		],
		"format": {"format_name": "matroska,webm", "duration": "120.5", "size": "90000000"}
	}`

	res, err := decode([]byte(raw))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	info := normalize(res)

	if len(info.Streams) != 1 {
		t.Fatalf("streams = %d, want 1", len(info.Streams))
	}
	s := info.Streams[0]
	if s.Width != 1920 || s.Height != 1080 {
		t.Errorf("resolution = %dx%d, want 1920x1080", s.Width, s.Height)
	}
	if !s.Default {
		t.Error("default flag was not parsed from a string value")
	}
	if s.Framerate < 23.9 || s.Framerate > 24.0 {
		t.Errorf("framerate = %v, want about 23.976", s.Framerate)
	}
}

func TestNormalizeDurationFallback(t *testing.T) {
	raw := `{
		"streams": [
			{"index": 0, "codec_name": "h264", "codec_type": "video", "width": 320, "height": 240,
			 "duration": "42.5"}
		],
		"format": {"format_name": "matroska,webm"}
	}`

	res, err := decode([]byte(raw))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	info := normalize(res)

	if info.Duration != 42.5 {
		t.Errorf("duration = %v, want 42.5 from the stream", info.Duration)
	}
}

func TestDecodeBadJSON(t *testing.T) {
	if _, err := decode([]byte("not json")); err == nil {
		t.Error("expected a parse error")
	}
}
