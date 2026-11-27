//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"strings"
	"testing"
)

func pipelineTestTask(audio TrackInfo) *Task {
	return &Task{
		ID:        "h",
		FileID:    "1",
		SourceURL: "http://127.0.0.1:8090/stream/?link=h&index=1&play",
		Config:    Config{}.normalized(),
		Probe: ProbeInfo{
			Container:         "Matroska",
			ContainerCapsName: "video/x-matroska",
			Tracks: []TrackInfo{
				{Type: "video", Index: 0, PadName: "video_0", Codec: "H.264", CapsName: "video/x-h264"},
				audio,
			},
		},
	}
}

func TestPipelineEmbeddedAACIsCopied(t *testing.T) {
	task := pipelineTestTask(TrackInfo{Type: "audio", Index: 0, PadName: "audio_0", Codec: "MPEG-4 AAC", CapsName: "audio/mpeg", Channels: 2, Rate: 48000})
	args := (&gstRunner{task: task, audioIndex: 0}).createPipelineArgs()

	want := "d.audio_0 ! mq.sink_1 mq.src_1 ! aacparse ! audio/mpeg,mpegversion=4,stream-format=raw ! mux.audio_0 "
	if !strings.Contains(args, want) {
		t.Errorf("pipeline lacks %q:\n%s", want, args)
	}
	if n := strings.Count(args, "souphttpsrc"); n != 1 {
		t.Errorf("%d sources, want 1", n)
	}
}

func TestPipelineEmbeddedAudioIsReencoded(t *testing.T) {
	task := pipelineTestTask(TrackInfo{Type: "audio", Index: 0, PadName: "audio_0", Codec: "FLAC", CapsName: "audio/x-flac", Channels: 2, Rate: 48000})
	args := (&gstRunner{task: task, audioIndex: 0}).createPipelineArgs()

	for _, want := range []string{
		"d.audio_0 ! mq.sink_1 mq.src_1 ! decodebin ! audioconvert dithering=none noise-shaping=none ! audioresample quality=2 sinc-filter-mode=full ! audio/x-raw,format=F32LE,layout=interleaved,rate=",
		" ! avenc_aac bitrate=",
		" ! aacparse ! audio/mpeg,mpegversion=4,stream-format=raw,rate=",
		" ! mux.audio_0 ",
	} {
		if !strings.Contains(args, want) {
			t.Errorf("pipeline lacks %q:\n%s", want, args)
		}
	}
	if strings.Contains(args, externalAudioSeekElement) {
		t.Error("an embedded track must not get the name of the external branch")
	}
}

func TestPipelineExternalAudioHasOwnSource(t *testing.T) {
	task := pipelineTestTask(TrackInfo{Type: "audio", Index: 0, PadName: "audio_0", Codec: "MPEG-4 AAC", CapsName: "audio/mpeg", Channels: 2, Rate: 48000})
	task.Audio = 1000
	task.ExternalAudio = &ExternalAudio{
		FileID:    "7",
		SourceURL: "http://127.0.0.1:8090/stream/?link=h&index=7&play",
		Track:     TrackInfo{Type: "audio", Index: 0, Codec: "Vorbis", CapsName: "audio/x-vorbis", Channels: 2, Rate: 44100},
	}
	runner := &gstRunner{task: task, audioIndex: resolveAudioIndex(task, 1000)}
	if !runner.externalAudioActive() {
		t.Fatal("external audio is not active")
	}
	args := runner.createPipelineArgs()

	for _, want := range []string{
		`souphttpsrc location="http://127.0.0.1:8090/stream/?link=h&index=7&play" is-live=false keep-alive=true timeout=60 retries=5 `,
		"! decodebin ! audioconvert name=" + externalAudioSeekElement + " dithering=none noise-shaping=none ! audioresample",
		" ! avenc_aac bitrate=",
		" ! aacparse ! audio/mpeg,mpegversion=4,stream-format=raw,rate=",
		" ! mux.audio_0 ",
	} {
		if !strings.Contains(args, want) {
			t.Errorf("pipeline lacks %q:\n%s", want, args)
		}
	}
	if n := strings.Count(args, "souphttpsrc"); n != 2 {
		t.Errorf("%d sources, want 2 (video and audio)", n)
	}
	// The audio of the video container is not used.
	for _, unwanted := range []string{"mq.sink_1", "mq.src_1", "d.audio_"} {
		if strings.Contains(args, unwanted) {
			t.Errorf("pipeline contains %q:\n%s", unwanted, args)
		}
	}
	if n := strings.Count(args, "mux.audio_0"); n != 1 {
		t.Errorf("%d links to the muxer audio pad, want 1", n)
	}
}

func TestPipelineExternalAudioNotActiveForAnotherTrack(t *testing.T) {
	task := pipelineTestTask(TrackInfo{Type: "audio", Index: 0, PadName: "audio_0", Codec: "MPEG-4 AAC", CapsName: "audio/mpeg", Channels: 2, Rate: 48000})
	task.Audio = 1000
	task.ExternalAudio = &ExternalAudio{SourceURL: "http://127.0.0.1:8090/stream/?link=h&index=7&play", Track: TrackInfo{Type: "audio"}}

	// A request for the embedded track must not start the external branch.
	runner := &gstRunner{task: task, audioIndex: resolveAudioIndex(task, 0)}
	if runner.externalAudioActive() {
		t.Fatal("external audio is active for an embedded track")
	}
	if n := strings.Count(runner.createPipelineArgs(), "souphttpsrc"); n != 1 {
		t.Errorf("%d sources, want 1", n)
	}
}

func TestResolveAudioIndex(t *testing.T) {
	probe := ProbeInfo{Tracks: []TrackInfo{
		{Type: "video", Index: 0},
		{Type: "audio", Index: 0},
		{Type: "audio", Index: 1},
	}}
	embedded := &Task{Audio: 1, Probe: probe}
	external := &Task{Audio: 1000, Probe: probe, ExternalAudio: &ExternalAudio{}}

	for _, test := range []struct {
		name      string
		task      *Task
		requested int
		want      int
	}{
		{"embedded track", embedded, 1, 1},
		{"unknown embedded track falls back to the first", embedded, 7, 0},
		{"external track of an external task", external, 1000, 1000},
		{"embedded track requested from an external task", external, 1, 1},
		{"another external track is not served by this task", external, 1001, 0},
		{"external track without external audio falls back", embedded, 1000, 0},
	} {
		if got := resolveAudioIndex(test.task, test.requested); got != test.want {
			t.Errorf("%s: resolveAudioIndex(%d) = %d, want %d", test.name, test.requested, got, test.want)
		}
	}
}

func TestExternalAudioSeekElementIsUnique(t *testing.T) {
	task := pipelineTestTask(TrackInfo{Type: "audio", Index: 0, PadName: "audio_0", Codec: "FLAC", CapsName: "audio/x-flac", Channels: 2, Rate: 48000})
	task.Audio = 1000
	task.ExternalAudio = &ExternalAudio{SourceURL: "http://127.0.0.1:8090/stream/?link=h&index=7&play", Track: TrackInfo{Type: "audio", Codec: "Vorbis", CapsName: "audio/x-vorbis", Channels: 2, Rate: 44100}}
	runner := &gstRunner{task: task, audioIndex: 1000}

	args := runner.createPipelineArgs()
	// The seek of the audio branch is sent to the element with this name, so it must exist exactly once.
	if n := strings.Count(args, "name="+externalAudioSeekElement+" "); n != 1 {
		t.Errorf("%d elements named %s, want 1:\n%s", n, externalAudioSeekElement, args)
	}
}

func TestExternalAudioStartNSWithoutProbe(t *testing.T) {
	runner := &gstRunner{}
	if got := runner.externalAudioStartNS(12.5); got != 12_500_000_000 {
		t.Errorf("externalAudioStartNS(12.5) = %d, want 12500000000", got)
	}
}

func TestExternalAudioStartNSPrefersProbe(t *testing.T) {
	state := &videoStartProbeState{}
	state.actualNS.Store(34_000_000_000)
	runner := &gstRunner{videoStartProbe: &gstPadProbeRegistration{state: state}}
	if got := runner.externalAudioStartNS(12.5); got != 34_000_000_000 {
		t.Errorf("externalAudioStartNS = %d, want the position the video probe saw", got)
	}

	// A probe that has not seen the start yet must not be trusted.
	state.actualNS.Store(gstClockTimeNone)
	if got := runner.externalAudioStartNS(12.5); got != 12_500_000_000 {
		t.Errorf("externalAudioStartNS = %d, want the fallback", got)
	}
}
