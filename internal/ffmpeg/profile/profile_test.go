package profile

import (
	"testing"

	"silo/internal/ffmpeg/args"
	"silo/internal/ffmpeg/media"
)

func TestDefaultIsValid(t *testing.T) {
	p := Default()
	if err := p.Validate(); err != nil {
		t.Fatalf("default profile is not valid: %v", err)
	}
	// По умолчанию сегментирование: оно даёт перемотку и адаптацию
	if p.Protocol != args.ProtocolHLS {
		t.Errorf("protocol = %q, want hls", p.Protocol)
	}
	if p.Container != "ts" {
		t.Errorf("container = %q, want ts", p.Container)
	}
	if p.HLS.SegmentLength <= 0 {
		t.Errorf("segment length = %d, want a positive default", p.HLS.SegmentLength)
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*Profile)
		wantErr bool
	}{
		{name: "ok", mutate: func(p *Profile) {}},
		{name: "no container", mutate: func(p *Profile) {
			// Контейнер обязателен только для сплошного потока
			p.Protocol = args.ProtocolProgressive
			p.Container = ""
		}, wantErr: true},
		{name: "no video codec", mutate: func(p *Profile) { p.Video.Codec = "" }, wantErr: true},
		{name: "no audio codec", mutate: func(p *Profile) { p.Audio.Codec = "" }, wantErr: true},
		{name: "hls without segment length", mutate: func(p *Profile) {
			p.Protocol = args.ProtocolHLS
			p.HLS.SegmentLength = 0
		}, wantErr: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := Default()
			c.mutate(&p)
			err := p.Validate()
			if c.wantErr && err == nil {
				t.Error("expected an error")
			}
			if !c.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestResolveKeepsTargetCodec(t *testing.T) {
	p := Default()
	info := testMediaInfo("h264", "aac")

	o, err := p.Resolve("http://127.0.0.1/source/1", info)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	// Исходный кодек совпадает с целевым, поэтому копирование допустимо
	if o.Video.Codec != args.CopyCodec {
		t.Errorf("video codec = %q, want copy", o.Video.Codec)
	}
	if o.Audio.Codec != args.CopyCodec {
		t.Errorf("audio codec = %q, want copy", o.Audio.Codec)
	}
	if o.Video.Encoder != "" {
		t.Errorf("encoder must be cleared for copy, got %q", o.Video.Encoder)
	}
}

func TestResolveTranscodesOtherCodec(t *testing.T) {
	p := Default()
	info := testMediaInfo("hevc", "ac3")

	o, err := p.Resolve("http://127.0.0.1/source/1", info)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	if o.Video.Codec != "h264" {
		t.Errorf("video codec = %q, want h264", o.Video.Codec)
	}
	if o.Audio.Codec != "aac" {
		t.Errorf("audio codec = %q, want aac", o.Audio.Codec)
	}
}

func TestResolveDisablesCopy(t *testing.T) {
	p := Default()
	p.AllowVideoCopy = false
	p.AllowAudioCopy = false
	info := testMediaInfo("h264", "aac")

	o, err := p.Resolve("input", info)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if o.Video.Codec == args.CopyCodec {
		t.Error("video copy must be disabled")
	}
	if o.Audio.Codec == args.CopyCodec {
		t.Error("audio copy must be disabled")
	}
}

func TestResolveVideoOnly(t *testing.T) {
	p := Default()

	info := &media.MediaInfo{
		Container: "mp4",
		Streams: []media.Stream{
			{Index: 0, Kind: media.KindVideo, Codec: "h264", Width: 640, Height: 480},
		},
	}

	o, err := p.Resolve("input", info)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if o.Audio.Codec != "none" {
		t.Errorf("audio codec = %q, want none for a video-only source", o.Audio.Codec)
	}
}

func TestResolveAudioOnly(t *testing.T) {
	p := Default()

	info := &media.MediaInfo{
		Container: "mp3",
		Streams: []media.Stream{
			{Index: 0, Kind: media.KindAudio, Codec: "mp3", Channels: 2, SampleRate: 44100},
		},
	}

	o, err := p.Resolve("input", info)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if o.Video.Codec != "none" {
		t.Errorf("video codec = %q, want none for an audio-only source", o.Video.Codec)
	}
}

func TestResolveOutput(t *testing.T) {
	// Сплошной поток пишется в канал, путь ему не нужен
	p := Default()
	p.Protocol = args.ProtocolProgressive
	p.Container = "mp4"

	progressive, err := p.Resolve("input", nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if progressive.Output != "pipe:1" {
		t.Errorf("progressive output = %q, want pipe:1", progressive.Output)
	}

	// Путь для сегментов подставляет модуль, а не профиль
	p.Protocol = args.ProtocolHLS
	hls, err := p.Resolve("input", nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if hls.Output != "" {
		t.Errorf("hls output = %q, want it to be filled by the module", hls.Output)
	}
}

func TestResolveRequiresInput(t *testing.T) {
	p := Default()
	if _, err := p.Resolve("", nil); err != nil {
		t.Fatalf("Resolve with an empty input must still build options: %v", err)
	}
}

func TestSameCodecAliases(t *testing.T) {
	cases := []struct {
		source string
		target string
		want   bool
	}{
		{source: "h264", target: "h264", want: true},
		{source: "avc1", target: "h264", want: true},
		{source: "h265", target: "hevc", want: true},
		{source: "aac", target: "aac", want: true},
		{source: "h264", target: "hevc", want: false},
		{source: "", target: "h264", want: false},
		{source: "H264", target: "h264", want: true},
	}

	for _, c := range cases {
		if got := sameCodec(c.source, c.target); got != c.want {
			t.Errorf("sameCodec(%q, %q) = %v, want %v", c.source, c.target, got, c.want)
		}
	}
}

// testMediaInfo собирает медиаинформацию с одним видео и одним аудио потоком.
func testMediaInfo(videoCodec, audioCodec string) *media.MediaInfo {
	return &media.MediaInfo{
		Container: "mkv",
		Duration:  120,
		Streams: []media.Stream{
			{Index: 0, Kind: media.KindVideo, Codec: videoCodec, Width: 1920, Height: 1080},
			{Index: 1, Kind: media.KindAudio, Codec: audioCodec, Channels: 2, SampleRate: 48000},
		},
	}
}

func TestResolveHDRIsNeverCopied(t *testing.T) {
	hdrInfo := &media.MediaInfo{
		Container: "mkv",
		Streams: []media.Stream{
			{
				Index: 0, Kind: media.KindVideo, Codec: "hevc",
				Width: 3840, Height: 2160, BitDepth: 10,
				ColorTransfer: "smpte2084", ColorPrim: "bt2020",
			},
			{Index: 1, Kind: media.KindAudio, Codec: "aac", Channels: 2},
		},
	}

	// Профиль просит hevc, кодеки совпадают - обычно поток копируется
	p := Default()
	p.Video.Codec = "hevc"
	p.AllowVideoCopy = true
	p.Video.TonemapMode = args.TonemapAuto

	o, err := p.Resolve("input", hdrInfo)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	// Копирование оставило бы HDR без приведения диапазона
	if o.Video.Codec == args.CopyCodec {
		t.Error("HDR source must not be copied: the picture stays washed out")
	}
	if !o.SourceHDR {
		t.Error("source must be marked as HDR")
	}
	if !o.Video.TonemapMode.Enabled(o.SourceHDR) {
		t.Error("tonemapping must be active for the HDR source")
	}
}

func TestResolveSDRStillCopies(t *testing.T) {
	sdrInfo := &media.MediaInfo{
		Container: "mkv",
		Streams: []media.Stream{
			{Index: 0, Kind: media.KindVideo, Codec: "h264", Width: 1920, Height: 1080, BitDepth: 8, ColorTransfer: "bt709"},
			{Index: 1, Kind: media.KindAudio, Codec: "aac", Channels: 2},
		},
	}

	p := Default()
	p.Video.Codec = "h264"
	p.AllowVideoCopy = true
	p.Video.TonemapMode = args.TonemapAuto

	o, err := p.Resolve("input", sdrInfo)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	// Обычный материал по-прежнему копируется без перекодирования
	if o.Video.Codec != args.CopyCodec {
		t.Errorf("video codec = %q, want copy for a matching SDR source", o.Video.Codec)
	}
	if o.SourceHDR {
		t.Error("SDR source must not be marked as HDR")
	}
}

func TestResolveTonemapOffKeepsCopy(t *testing.T) {
	hdrInfo := &media.MediaInfo{
		Container: "mkv",
		Streams: []media.Stream{
			{
				Index: 0, Kind: media.KindVideo, Codec: "hevc",
				Width: 3840, Height: 2160, BitDepth: 10,
				ColorTransfer: "smpte2084", ColorPrim: "bt2020",
			},
		},
	}

	// Явный отказ от приведения разрешает копирование
	p := Default()
	p.Video.Codec = "hevc"
	p.AllowVideoCopy = true
	p.Video.TonemapMode = args.TonemapOff

	o, err := p.Resolve("input", hdrInfo)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if o.Video.Codec != args.CopyCodec {
		t.Errorf("video codec = %q, want copy when tonemapping is off", o.Video.Codec)
	}
}
