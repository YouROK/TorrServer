//go:build gst

package gstreamer

import "testing"

func TestProbeFromDiscoverer(t *testing.T) {
	text := `
Properties:
  Duration: 1:34:09.920000000
  Seekable: yes
  container: Matroska
    video #1: H.264 (High Profile)
      Width: 1920
      Height: 800
      Frame rate: 24000/1001
      language code: und
    audio #2: AC-3 (ATSC A/52)
      Sample rate: 48000
      Channels: 6
      language code: ru
      title: DUB
    audio #3: E-AC-3 (ATSC A/52B)
      Sample rate: 48000
      Channels: 6
      language code: en
      title: Original
`

	probe := probeFromDiscoverer(text)
	if probe.DurationNS != 5_649_920_000_000 {
		t.Fatalf("DurationNS = %d", probe.DurationNS)
	}
	if probe.Container != "Matroska" {
		t.Fatalf("Container = %q, want Matroska", probe.Container)
	}
	if len(probe.Tracks) != 3 {
		t.Fatalf("Tracks len = %d", len(probe.Tracks))
	}

	video := probe.Tracks[0]
	if video.Type != "video" || video.Index != 0 || video.PadName != "video_0" || video.CapsName != "video/x-h264" {
		t.Fatalf("video track = %+v", video)
	}
	if video.Width != 1920 || video.Height != 800 || video.FrameRateNum != 24000 || video.FrameRateDen != 1001 {
		t.Fatalf("video details = %+v", video)
	}

	audio0 := probe.Tracks[1]
	if audio0.Type != "audio" || audio0.Index != 0 || audio0.PadName != "audio_0" || audio0.CapsName != "audio/x-ac3" {
		t.Fatalf("audio0 track = %+v", audio0)
	}
	if audio0.Rate != 48000 || audio0.Channels != 6 || audio0.Language != "ru" || audio0.Title != "DUB" {
		t.Fatalf("audio0 details = %+v", audio0)
	}

	audio1 := probe.Tracks[2]
	if audio1.Type != "audio" || audio1.Index != 1 || audio1.PadName != "audio_1" || audio1.CapsName != "audio/x-eac3" {
		t.Fatalf("audio1 track = %+v", audio1)
	}
	if audio1.Rate != 48000 || audio1.Channels != 6 || audio1.Language != "en" || audio1.Title != "Original" {
		t.Fatalf("audio1 details = %+v", audio1)
	}
}

func TestProbeParsesMatroskaContainer(t *testing.T) {
	probe := probeFromDiscoverer(`
Properties:
  Duration: 0:01:00.000000000
  container #0: Matroska
    video #1: H.264 (High Profile)
    audio #2: AAC
`)

	if probe.Container != "Matroska" {
		t.Fatalf("Container=%q", probe.Container)
	}
	if !probe.IsMatroskaContainer() {
		t.Fatal("Matroska container was not accepted")
	}
	if !probe.HasAudio() {
		t.Fatal("audio track was not detected")
	}
}

func TestProbeParsesAACAudioCodec(t *testing.T) {
	probe := probeFromDiscoverer(`
Properties:
  Duration: 0:01:00.000000000
  container #0: Matroska
    video #1: H.264
    audio #2: MPEG-4 AAC
`)

	audio := probe.AudioTrack(0)
	if audio == nil {
		t.Fatal("audio track was not detected")
	}
	if audio.Codec != "MPEG-4 AAC" || audio.CapsName != "audio/mpeg" {
		t.Fatalf("audio track = %+v", *audio)
	}
	if !audio.IsAACAudio() {
		t.Fatal("AAC audio track was not detected")
	}
}

func TestAACAudioDetectionAcceptsMPEG4Caps(t *testing.T) {
	track := TrackInfo{
		Type:  "audio",
		Codec: "audio/mpeg, mpegversion=(int)4, stream-format=(string)raw",
	}

	if !track.IsAACAudio() {
		t.Fatal("MPEG-4 audio caps were not detected as AAC")
	}
}

func TestProbeParsesVideoOnlyWebM(t *testing.T) {
	probe := probeFromDiscoverer(`
Properties:
  Duration: 0:00:10.000000000
  container #0: WebM
    video #1: VP9
`)

	if !probe.IsMatroskaContainer() {
		t.Fatal("WebM container was not accepted")
	}
	if probe.HasAudio() {
		t.Fatal("video-only source unexpectedly has audio")
	}
}

func TestValidateProbeRejectsNonMatroska(t *testing.T) {
	err := validateProbe(ProbeInfo{
		Container: "Quicktime",
		Tracks: []TrackInfo{
			{Type: "video", CapsName: "video/x-h264"},
		},
	}, Config{})
	if err == nil {
		t.Fatal("non-Matroska source was accepted")
	}
}

func TestValidateProbeAcceptsAVIOnlyWithTranscode(t *testing.T) {
	probe := ProbeInfo{
		Container:         "AVI",
		ContainerCapsName: "video/x-msvideo",
		Tracks:            []TrackInfo{{Type: "video", CapsName: "video/x-h264"}},
	}
	if err := validateProbe(probe, Config{}); err == nil {
		t.Fatal("AVI was accepted without TranscodeAVI")
	}
	if err := validateProbe(probe, Config{TranscodeAVI: true}); err != nil {
		t.Fatalf("AVI with TranscodeAVI was rejected: %v", err)
	}
}

func TestValidateProbeAcceptsVP8OnlyWithTranscode(t *testing.T) {
	probe := ProbeInfo{
		Container: "WebM",
		Tracks:    []TrackInfo{{Type: "video", CapsName: "video/x-vp8"}},
	}
	if err := validateProbe(probe, Config{}); err == nil {
		t.Fatal("VP8 was accepted without TranscodeVP8")
	}
	if err := validateProbe(probe, Config{TranscodeVP8: true}); err != nil {
		t.Fatalf("VP8 with TranscodeVP8 was rejected: %v", err)
	}
}
