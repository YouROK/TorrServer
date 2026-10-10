package args

import (
	"reflect"
	"strings"
	"testing"
)

// fakeSelector подменяет список доступных кодировщиков.
type fakeSelector struct {
	video map[string]string
	audio map[string]string
}

func (f fakeSelector) AutoVideoEncoder(codec, hwaccel string) string {
	if hwaccel != "" && hwaccel != "none" {
		if enc, ok := f.video[codec+"@"+hwaccel]; ok {
			return enc
		}
	}
	return f.video[codec]
}

func (f fakeSelector) AutoAudioEncoder(codec string) string { return f.audio[codec] }

// defaultSelector описывает сборку с программными кодировщиками.
func defaultSelector() fakeSelector {
	return fakeSelector{
		video: map[string]string{
			"h264":       "libx264",
			"hevc":       "libx265",
			"av1":        "libsvtav1",
			"h264@nvenc": "h264_nvenc",
			"h264@vaapi": "h264_vaapi",
		},
		audio: map[string]string{
			"aac":  "aac",
			"opus": "libopus",
			"mp3":  "libmp3lame",
		},
	}
}

// indexOf возвращает позицию аргумента в списке.
func indexOf(args []string, name string) int {
	for i, a := range args {
		if a == name {
			return i
		}
	}
	return -1
}

// valueOf возвращает значение аргумента ключ-значение.
func valueOf(args []string, name string) string {
	i := indexOf(args, name)
	if i < 0 || i+1 >= len(args) {
		return ""
	}
	return args[i+1]
}

// containsPair проверяет наличие пары ключ-значение.
func containsPair(args []string, name, value string) bool {
	return valueOf(args, name) == value
}

func TestBuildRequiresInput(t *testing.T) {
	if _, err := Build(Options{}, defaultSelector()); err == nil {
		t.Error("expected an error for a missing input")
	}
}

func TestBuildProgressiveMp4(t *testing.T) {
	o := DefaultOptions("http://127.0.0.1:1234/source/1")
	o.Seek = 65.5
	o.Video.MaxWidth = 1280
	o.Video.MaxHeight = 720

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !containsPair(got, "-c:v", "libx264") {
		t.Errorf("video encoder missing: %v", got)
	}
	if !containsPair(got, "-c:a", "aac") {
		t.Errorf("audio encoder missing: %v", got)
	}
	if !containsPair(got, "-crf", "23") {
		t.Errorf("crf missing: %v", got)
	}
	if !containsPair(got, "-preset", "veryfast") {
		t.Errorf("preset missing: %v", got)
	}
	if !containsPair(got, "-f", "mp4") {
		t.Errorf("output format missing: %v", got)
	}
	if !containsPair(got, "-movflags", "frag_keyframe+empty_moov+delay_moov") {
		t.Errorf("movflags missing: %v", got)
	}
	if !containsPair(got, "-ss", "00:01:05.500") {
		t.Errorf("seek argument wrong: %v", valueOf(got, "-ss"))
	}
	if vf := valueOf(got, "-vf"); !strings.Contains(vf, "scale=") || !strings.Contains(vf, "force_original_aspect_ratio=decrease") {
		t.Errorf("scale filter missing: %v", vf)
	}
	if !containsPair(got, "-progress", "pipe:2") {
		t.Errorf("progress argument missing: %v", got)
	}
	if got[len(got)-1] != "pipe:1" {
		t.Errorf("last argument = %q, want output pipe:1", got[len(got)-1])
	}
}

func TestBuildSeekBeforeInput(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Seek = 10

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	seekPos := indexOf(got, "-ss")
	inputPos := indexOf(got, "-i")
	if seekPos < 0 || inputPos < 0 {
		t.Fatalf("seek or input argument missing: %v", got)
	}
	if seekPos > inputPos {
		t.Errorf("-ss must precede -i for fast seek, got %v", got)
	}
}

func TestBuildCopyStreams(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.Codec = CopyCodec
	o.Audio.Codec = CopyCodec

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !containsPair(got, "-c:v", "copy") {
		t.Errorf("video copy missing: %v", got)
	}
	if !containsPair(got, "-c:a", "copy") {
		t.Errorf("audio copy missing: %v", got)
	}
	if indexOf(got, "-crf") >= 0 {
		t.Errorf("copy must not set quality options: %v", got)
	}
}

func TestBuildVideoOnly(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Audio.Codec = "none"

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if indexOf(got, "-an") < 0 {
		t.Errorf("expected -an for a video-only output: %v", got)
	}
}

func TestBuildBitrateMode(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.CRF = 0
	o.Video.BitrateKbps = 4000
	o.Video.MaxBitrateKbps = 6000

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !containsPair(got, "-b:v", "4000k") {
		t.Errorf("target bitrate missing: %v", got)
	}
	if !containsPair(got, "-maxrate", "6000k") {
		t.Errorf("maxrate missing: %v", got)
	}
	if !containsPair(got, "-bufsize", "12000k") {
		t.Errorf("bufsize missing: %v", got)
	}
	if indexOf(got, "-crf") >= 0 {
		t.Errorf("bitrate mode must not set crf: %v", got)
	}
}

func TestBuildHardwareEncoderUsesQp(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.Codec = "h264"
	o.Video.HWAccel = "vaapi"
	o.Video.CRF = 26

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !containsPair(got, "-c:v", "h264_vaapi") {
		t.Errorf("hardware encoder was not selected: %v", got)
	}
	if !containsPair(got, "-qp", "26") {
		t.Errorf("hardware quality option missing: %v", got)
	}
	if indexOf(got, "-crf") >= 0 {
		t.Errorf("hardware encoder must not receive crf: %v", got)
	}
}

func TestBuildMissingEncoder(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.Codec = "av1"

	_, err := Build(o, fakeSelector{})
	if err == nil {
		t.Fatal("expected an error when no encoder is available")
	}
	if !strings.Contains(err.Error(), "av1") {
		t.Errorf("error should name the codec: %v", err)
	}
}

func TestBuildExplicitEncoder(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.Encoder = "libx265"
	o.Audio.Encoder = "libopus"

	got, err := Build(o, fakeSelector{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !containsPair(got, "-c:v", "libx265") {
		t.Errorf("explicit video encoder was ignored: %v", got)
	}
	if !containsPair(got, "-c:a", "libopus") {
		t.Errorf("explicit audio encoder was ignored: %v", got)
	}
}

func TestBuildHLSSegments(t *testing.T) {
	o := DefaultOptions("http://127.0.0.1:1234/source/1")
	o.Protocol = ProtocolHLS
	o.Container = "ts"
	o.Output = "/cache/playlist.m3u8"
	o.HLS = HLSOptions{
		SegmentLength:   6,
		SegmentType:     "mpegts",
		ListSize:        5,
		DeleteThreshold: 1,
		SegmentFilename: "/cache/seg%d.ts",
	}

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !containsPair(got, "-f", "hls") {
		t.Errorf("hls format missing: %v", got)
	}
	if !containsPair(got, "-hls_time", "6") {
		t.Errorf("segment length missing: %v", got)
	}
	if !containsPair(got, "-hls_segment_type", "mpegts") {
		t.Errorf("segment type missing: %v", got)
	}
	if !containsPair(got, "-hls_segment_filename", "/cache/seg%d.ts") {
		t.Errorf("segment filename missing: %v", got)
	}
	if !containsPair(got, "-hls_list_size", "5") {
		t.Errorf("list size missing: %v", got)
	}
	if !containsPair(got, "-hls_delete_threshold", "1") {
		t.Errorf("delete threshold missing: %v", got)
	}
	if !containsPair(got, "-hls_flags", "delete_segments+temp_file") {
		t.Errorf("segment deletion flags missing: %v", got)
	}
	if !containsPair(got, "-force_key_frames", "expr:gte(t,n_forced*6)") {
		t.Errorf("keyframe alignment missing: %v", got)
	}
	if !containsPair(got, "-copyts", "") && indexOf(got, "-copyts") < 0 {
		t.Errorf("copyts missing: %v", got)
	}
	if !containsPair(got, "-avoid_negative_ts", "disabled") {
		t.Errorf("avoid_negative_ts missing: %v", got)
	}
	if got[len(got)-1] != "/cache/playlist.m3u8" {
		t.Errorf("output = %q, want the playlist path", got[len(got)-1])
	}
}

func TestBuildHLSVodKeepsPlaylist(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Protocol = ProtocolHLS
	o.Container = "ts"
	o.Output = "/cache/out.m3u8"
	o.HLS = DefaultHLSOptions()
	o.HLS.PlaylistType = "vod"

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !containsPair(got, "-hls_playlist_type", "vod") {
		t.Errorf("playlist type missing: %v", got)
	}
	if !containsPair(got, "-hls_list_size", "0") {
		t.Errorf("vod playlist must keep all entries: %v", got)
	}
	if indexOf(got, "-hls_flags") >= 0 {
		t.Errorf("vod playlist must not delete segments: %v", got)
	}
}

func TestBuildThreadsAndAnalyze(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Threads = 4
	o.AnalyzeDurMs = 2000
	o.ProbeSizeMB = 10

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !containsPair(got, "-threads", "4") {
		t.Errorf("threads missing: %v", got)
	}
	if !containsPair(got, "-analyzeduration", "2000000") {
		t.Errorf("analyzeduration missing: %v", got)
	}
	if !containsPair(got, "-probesize", "10485760") {
		t.Errorf("probesize missing: %v", got)
	}
}

func TestBuildDropSubtitles(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.DropSubs = true

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if indexOf(got, "-sn") < 0 {
		t.Errorf("expected -sn: %v", got)
	}
}

func TestBuildMaxFramerateAndPixFmt(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.MaxFramerate = 30
	o.Video.PixFmt = "yuv420p"
	o.Video.Profile = "main"
	o.Video.Level = "4.0"

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !containsPair(got, "-r", "30") {
		t.Errorf("framerate limit missing: %v", got)
	}
	if !containsPair(got, "-pix_fmt", "yuv420p") {
		t.Errorf("pixel format missing: %v", got)
	}
	if !containsPair(got, "-profile:v", "main") {
		t.Errorf("profile missing: %v", got)
	}
	if !containsPair(got, "-level", "4.0") {
		t.Errorf("level missing: %v", got)
	}
}

func TestBuildOutputContainer(t *testing.T) {
	cases := []struct {
		container string
		wantFmt   string
	}{
		{container: "mp4", wantFmt: "mp4"},
		{container: "mkv", wantFmt: "matroska"},
		{container: "matroska", wantFmt: "matroska"},
		{container: "ts", wantFmt: "mpegts"},
		{container: "webm", wantFmt: "webm"},
	}

	for _, c := range cases {
		t.Run(c.container, func(t *testing.T) {
			o := DefaultOptions("input.mkv")
			o.Container = c.container

			got, err := Build(o, defaultSelector())
			if err != nil {
				t.Fatalf("Build: %v", err)
			}
			if !containsPair(got, "-f", c.wantFmt) {
				t.Errorf("container %q produced -f %q, want %q", c.container, valueOf(got, "-f"), c.wantFmt)
			}
		})
	}
}

func TestBuildOrderOfSections(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Seek = 5
	o.Video.MaxWidth = 640

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	positions := map[string]int{
		"-analyzeduration": indexOf(got, "-analyzeduration"),
		"-ss":              indexOf(got, "-ss"),
		"-i":               indexOf(got, "-i"),
		"-map":             indexOf(got, "-map"),
		"-c:v":             indexOf(got, "-c:v"),
		"-c:a":             indexOf(got, "-c:a"),
		"-f":               indexOf(got, "-f"),
	}

	order := []string{"-analyzeduration", "-ss", "-i", "-map", "-c:v", "-c:a", "-f"}
	prev := -1
	for _, name := range order {
		pos := positions[name]
		if pos < 0 {
			t.Fatalf("argument %s missing in %v", name, got)
		}
		if pos < prev {
			t.Errorf("argument %s is out of order in %v", name, got)
		}
		prev = pos
	}
}

func TestBuildNoVideoNoAudio(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.Codec = "none"
	o.Audio.Codec = "none"

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if indexOf(got, "-vn") < 0 || indexOf(got, "-an") < 0 {
		t.Errorf("expected both -vn and -an: %v", got)
	}
}

func TestFormatTimestamp(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{in: 0, want: "00:00:00.000"},
		{in: 65.5, want: "00:01:05.500"},
		{in: 3661.25, want: "01:01:01.250"},
		{in: -5, want: "00:00:00.000"},
	}

	for _, c := range cases {
		if got := formatTimestamp(c.in); got != c.want {
			t.Errorf("formatTimestamp(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBuildArgsAreSliceNotString(t *testing.T) {
	o := DefaultOptions("input with spaces.mkv")

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	want := "input with spaces.mkv"
	if !reflect.DeepEqual(got[indexOf(got, "-i")+1], want) {
		t.Errorf("input path was not passed as a single argument: %v", got)
	}
}

func TestBuildHardwareDeviceArgs(t *testing.T) {
	cases := []struct {
		encoder string
		want    string
	}{
		{encoder: "h264_vaapi", want: "vaapi=va:/dev/dri/renderD128"},
		{encoder: "h264_nvenc", want: "cuda=cu:0"},
		{encoder: "h264_qsv", want: "qsv=qs"},
		{encoder: "libx264", want: ""},
		{encoder: "", want: ""},
	}

	for _, c := range cases {
		t.Run(c.encoder, func(t *testing.T) {
			got := buildHwDeviceArgs(hwFromEncoder(c.encoder), "")
			if c.want == "" {
				if len(got) != 0 {
					t.Errorf("args = %v, want none", got)
				}
				return
			}
			if !containsPair(got, "-init_hw_device", c.want) {
				t.Errorf("args = %v, want device %q", got, c.want)
			}
			if indexOf(got, "-filter_hw_device") < 0 {
				t.Errorf("filter device is missing: %v", got)
			}
		})
	}
}

func TestBuildHardwareFilters(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.HWAccel = "vaapi"
	o.Video.MaxWidth, o.Video.MaxHeight = 1280, 720
	o.Video.Codec = "h264"
	o.Video.Encoder = "h264_vaapi"

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Устройство создаётся до ввода
	devicePos := indexOf(got, "-init_hw_device")
	inputPos := indexOf(got, "-i")
	if devicePos < 0 || devicePos > inputPos {
		t.Errorf("-init_hw_device must precede -i, got %v", got)
	}

	vf := valueOf(got, "-vf")
	if !strings.Contains(vf, "hwupload") {
		t.Errorf("hardware filter chain has no hwupload: %q", vf)
	}
	if !strings.Contains(vf, "scale_vaapi") {
		t.Errorf("hardware filter chain has no scale_vaapi: %q", vf)
	}
	if strings.Contains(vf, "scale=w=") {
		t.Errorf("software scale must not be used with hardware encoding: %q", vf)
	}
}

func TestBuildSoftwareFiltersUnchanged(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.MaxWidth, o.Video.MaxHeight = 1280, 720

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if indexOf(got, "-init_hw_device") >= 0 {
		t.Errorf("software path must not create a hw device: %v", got)
	}
	vf := valueOf(got, "-vf")
	if !strings.Contains(vf, "scale=w=") {
		t.Errorf("software scale is missing: %q", vf)
	}
	if strings.Contains(vf, "hwupload") {
		t.Errorf("software path must not upload frames: %q", vf)
	}
}

func TestBuildDeinterlaceBeforeHwUpload(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.HWAccel = "vaapi"
	o.Video.Deinterlace = "yadif"
	o.Video.Encoder = "h264_vaapi"

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	vf := valueOf(got, "-vf")
	deintPos := strings.Index(vf, "yadif")
	uploadPos := strings.Index(vf, "hwupload")

	if deintPos < 0 || uploadPos < 0 {
		t.Fatalf("filter chain = %q, want yadif and hwupload", vf)
	}
	if deintPos > uploadPos {
		t.Errorf("deinterlace must run before upload: %q", vf)
	}
}

func TestHwFromEncoder(t *testing.T) {
	cases := map[string]string{
		"h264_vaapi": "vaapi",
		"hevc_vaapi": "vaapi",
		"h264_nvenc": "cuda",
		"hevc_nvenc": "cuda",
		"h264_qsv":   "qsv",
		"libx264":    "",
		"libx265":    "",
		"":           "",
	}

	for in, want := range cases {
		if got := hwFromEncoder(in); got != want {
			t.Errorf("hwFromEncoder(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestBuildFallsBackToSoftwarePipeline проверяет, что при откате
// на программный кодировщик аппаратные аргументы не добавляются.
func TestBuildFallsBackToSoftwarePipeline(t *testing.T) {
	o := DefaultOptions("input.mkv")
	// Ускорение запрошено, но такого кодировщика в сборке нет
	o.Video.HWAccel = "vaapi"
	o.Video.Codec = "h264"
	o.Video.MaxHeight = 720

	// Селектор вернёт программный libx264, как поступил бы реальный подбор
	sel := fakeSelector{
		video: map[string]string{"h264": "libx264"},
		audio: map[string]string{"aac": "aac"},
	}

	got, err := Build(o, sel)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !containsPair(got, "-c:v", "libx264") {
		t.Errorf("expected a software encoder: %v", got)
	}
	if indexOf(got, "-init_hw_device") >= 0 {
		t.Errorf("software fallback must not create a hw device: %v", got)
	}
	vf := valueOf(got, "-vf")
	if strings.Contains(vf, "hwupload") || strings.Contains(vf, "scale_vaapi") {
		t.Errorf("software fallback must not use hw filters: %q", vf)
	}
	if !strings.Contains(vf, "scale=w=") {
		t.Errorf("software fallback must use the software scale: %q", vf)
	}
}

func TestBuildHwDecodeArgs(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.HWAccel = "vaapi"
	o.Video.HWDevice = "/dev/dri/renderD128"
	o.Video.HWDecode = true
	o.Video.Encoder = "h264_vaapi"

	got, err := Build(o, fakeSelector{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !containsPair(got, "-hwaccel", "vaapi") {
		t.Errorf("hardware decoding is missing: %v", got)
	}
	if !containsPair(got, "-hwaccel_output_format", "vaapi") {
		t.Errorf("output format is missing: %v", got)
	}
	if !containsPair(got, "-hwaccel_device", "/dev/dri/renderD128") {
		t.Errorf("device is missing: %v", got)
	}

	// Декодирование задаётся до ввода
	if indexOf(got, "-hwaccel") > indexOf(got, "-i") {
		t.Errorf("-hwaccel must precede -i: %v", got)
	}
}

func TestBuildHwDecodeSkipsUpload(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.HWAccel = "vaapi"
	o.Video.HWDecode = true
	o.Video.MaxHeight = 720
	o.Video.Encoder = "h264_vaapi"

	got, err := Build(o, fakeSelector{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	vf := valueOf(got, "-vf")
	// Кадры уже в памяти устройства, загружать нечего
	if strings.Contains(vf, "hwupload") {
		t.Errorf("frames are already on the device, hwupload must be skipped: %q", vf)
	}
	if !strings.Contains(vf, "scale_vaapi") {
		t.Errorf("hardware scale is missing: %q", vf)
	}
}

func TestBuildHwDecodeDeinterlaceOnDevice(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.HWAccel = "vaapi"
	o.Video.HWDecode = true
	o.Video.Deinterlace = "yadif"
	o.Video.Encoder = "h264_vaapi"

	got, err := Build(o, fakeSelector{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	vf := valueOf(got, "-vf")
	if strings.Contains(vf, "yadif=deint") {
		t.Errorf("software deinterlace cannot run on device frames: %q", vf)
	}
	if !strings.Contains(vf, "deinterlace_vaapi") {
		t.Errorf("hardware deinterlace is missing: %q", vf)
	}
}

func TestBuildNoHwDecodeByDefault(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.HWAccel = "vaapi"
	o.Video.Encoder = "h264_vaapi"
	// HWDecode не задан

	got, err := Build(o, fakeSelector{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if indexOf(got, "-hwaccel") >= 0 {
		t.Errorf("hardware decoding must be opt-in: %v", got)
	}
	if !strings.Contains(valueOf(got, "-vf"), "hwupload") {
		t.Errorf("software frames must be uploaded to the device: %v", valueOf(got, "-vf"))
	}
}

func TestBuildTransposeSoftware(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.Transpose = "90"
	o.Video.MaxHeight = 720

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	vf := valueOf(got, "-vf")
	if !strings.Contains(vf, "transpose=dir=clock") {
		t.Errorf("transpose filter is missing: %q", vf)
	}
	// Поворот выполняется до масштабирования
	if strings.Index(vf, "transpose") > strings.Index(vf, "scale") {
		t.Errorf("transpose must precede scale: %q", vf)
	}
}

func TestBuildTransposeHardware(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.HWAccel = "vaapi"
	o.Video.Encoder = "h264_vaapi"
	o.Video.Transpose = "270"

	got, err := Build(o, fakeSelector{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !strings.Contains(valueOf(got, "-vf"), "transpose_vaapi=dir=reversal") {
		t.Errorf("hardware transpose is missing: %q", valueOf(got, "-vf"))
	}
}

func TestTransposeFilterModes(t *testing.T) {
	cases := map[string]string{
		"90":       "transpose=dir=clock",
		"180":      "transpose=dir=cclock",
		"270":      "transpose=dir=reversal",
		"clock":    "transpose=dir=clock",
		"cclock":   "transpose=dir=cclock",
		"reversal": "transpose=dir=reversal",
	}

	for in, want := range cases {
		if got := transposeFilter(in, ""); got != want {
			t.Errorf("transposeFilter(%q) = %q, want %q", in, got, want)
		}
	}

	if got := transposeFilter("90", "vaapi"); got != "transpose_vaapi=dir=clock" {
		t.Errorf("hardware transpose = %q", got)
	}
}

func TestHwDeinterlaceFilter(t *testing.T) {
	cases := map[string]string{
		"cuda":  "yadif_cuda=deint=interlaced",
		"qsv":   "deinterlace_qsv=mode=2",
		"vaapi": "deinterlace_vaapi=rate=field:auto=1",
	}

	for hw, want := range cases {
		if got := hwDeinterlaceFilter(hw); got != want {
			t.Errorf("hwDeinterlaceFilter(%q) = %q, want %q", hw, got, want)
		}
	}
}

func TestBuildTonemapSoftware(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.TonemapMode = TonemapOn
	o.Video.MaxHeight = 720

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	vf := valueOf(got, "-vf")
	for _, want := range []string{"zscale=t=linear", "tonemap=tonemap=hable", "zscale=t=bt709"} {
		if !strings.Contains(vf, want) {
			t.Errorf("filter chain %q does not contain %q", vf, want)
		}
	}
	if !strings.Contains(vf, "format=yuv420p") {
		t.Errorf("output must be converted to 8 bit: %q", vf)
	}
	// Тонемаппинг выполняется до масштабирования: ищем именно фильтр масштаба,
	// а не zscale, в имени которого тоже есть "scale"
	tonemapPos := strings.Index(vf, "tonemap=tonemap=")
	scalePos := strings.Index(vf, "scale=w=")
	if tonemapPos < 0 || scalePos < 0 || tonemapPos > scalePos {
		t.Errorf("tonemapping must precede scaling: %q", vf)
	}
}

func TestBuildTonemapAlgorithm(t *testing.T) {
	for _, algo := range []string{"hable", "mobius", "reinhard", "clip"} {
		o := DefaultOptions("input.mkv")
		o.Video.TonemapMode = TonemapOn
		o.Video.TonemapAlgorithm = algo

		got, err := Build(o, defaultSelector())
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		if !strings.Contains(valueOf(got, "-vf"), "tonemap="+algo) {
			t.Errorf("algorithm %q is missing: %q", algo, valueOf(got, "-vf"))
		}
	}
}

func TestBuildTonemapHardware(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Video.HWAccel = "vaapi"
	o.Video.HWDecode = true
	o.Video.TonemapMode = TonemapOn
	o.Video.Encoder = "h264_vaapi"

	got, err := Build(o, fakeSelector{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	vf := valueOf(got, "-vf")
	if !strings.Contains(vf, "tonemap_vaapi") {
		t.Errorf("hardware tonemapping is missing: %q", vf)
	}
	if strings.Contains(vf, "tonemap=tonemap=") {
		t.Errorf("software tonemapping must not be used on device frames: %q", vf)
	}
}

func TestBuildTonemapDisabledByDefault(t *testing.T) {
	// По умолчанию режим auto, поэтому обычный источник не трогается
	o := DefaultOptions("input.mkv")

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if strings.Contains(valueOf(got, "-vf"), "tonemap") {
		t.Errorf("tonemapping must be opt-in: %q", valueOf(got, "-vf"))
	}
}

func TestBuildSubtitleOffByDefault(t *testing.T) {
	o := DefaultOptions("input.mkv")

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if indexOf(got, "-sn") < 0 {
		t.Errorf("subtitles must be dropped by default: %v", got)
	}
	if strings.Contains(valueOf(got, "-vf"), "subtitles=") {
		t.Errorf("no burn filter expected by default: %q", valueOf(got, "-vf"))
	}
}

func TestBuildSubtitleBurn(t *testing.T) {
	o := DefaultOptions("/media/movies/film.mkv")
	o.Subtitles.Mode = SubtitleBurn

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	vf := valueOf(got, "-vf")
	if !strings.Contains(vf, "subtitles=f=") {
		t.Errorf("burn filter is missing: %q", vf)
	}
	// Фильтру нужно имя файла, иначе он не запускается
	if !strings.Contains(vf, "film.mkv") {
		t.Errorf("burn filter must reference the source file: %q", vf)
	}
	if indexOf(got, "-sn") >= 0 {
		t.Errorf("burn-in must not drop subtitle streams: %v", got)
	}
}

func TestBuildSubtitleBurnWithIndex(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Subtitles.Mode = SubtitleBurn
	idx := 3
	o.Subtitles.StreamIndex = &idx

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !strings.Contains(valueOf(got, "-vf"), ":si=3") {
		t.Errorf("stream index is missing: %q", valueOf(got, "-vf"))
	}
}

func TestBuildSubtitleBurnWithStyle(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Subtitles.Mode = SubtitleBurn
	o.Subtitles.ForceStyle = "FontSize=24,PrimaryColour=&H00FFFFFF"

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	vf := valueOf(got, "-vf")
	if !strings.Contains(vf, "force_style=") {
		t.Errorf("style override is missing: %q", vf)
	}
	if !strings.Contains(vf, "FontSize=24") {
		t.Errorf("style value is missing: %q", vf)
	}
}

func TestBuildSubtitleCopy(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Subtitles.Mode = SubtitleCopy

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !hasMapArg(got, "0:s:0?") {
		t.Errorf("subtitle stream is not mapped: %v", got)
	}
	if indexOf(got, "-sn") >= 0 {
		t.Errorf("copy must not drop subtitles: %v", got)
	}
	if strings.Contains(valueOf(got, "-vf"), "subtitles=") {
		t.Errorf("copy must not burn subtitles: %q", valueOf(got, "-vf"))
	}
}

func TestBuildSubtitleCopyWithIndex(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.Subtitles.Mode = SubtitleCopy
	idx := 4
	o.Subtitles.StreamIndex = &idx

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if !hasMapArg(got, "0:4") {
		t.Errorf("explicit subtitle stream is missing: %v", got)
	}
}

func TestEscapeFilterPath(t *testing.T) {
	// Обычный путь оборачивается в кавычки без изменений
	if got := escapeFilterPath("/media/film.mkv"); got != "'/media/film.mkv'" {
		t.Errorf("plain path = %q", got)
	}

	// Двоеточие и обратный слэш экранируются: иначе фильтр разберёт путь неверно
	got := escapeFilterPath("C:\\media\\film.mkv")
	if !strings.Contains(got, "\\\\:") {
		t.Errorf("colon must be escaped: %q", got)
	}
	if !strings.HasPrefix(got, "'") || !strings.HasSuffix(got, "'") {
		t.Errorf("path must be quoted: %q", got)
	}
}

// hasMapArg проверяет наличие значения среди аргументов -map.
func hasMapArg(args []string, want string) bool {
	for i, a := range args {
		if a == "-map" && i+1 < len(args) && args[i+1] == want {
			return true
		}
	}
	return false
}

func TestTonemapModeEnabled(t *testing.T) {
	cases := []struct {
		mode     TonemapMode
		source   bool
		expected bool
	}{
		{mode: TonemapAuto, source: true, expected: true},
		{mode: TonemapAuto, source: false, expected: false},
		{mode: TonemapOn, source: false, expected: true},
		{mode: TonemapOn, source: true, expected: true},
		{mode: TonemapOff, source: true, expected: false},
		{mode: TonemapOff, source: false, expected: false},
		// Пустое значение означает авто
		{mode: "", source: true, expected: true},
		{mode: "", source: false, expected: false},
	}

	for _, c := range cases {
		if got := c.mode.Enabled(c.source); got != c.expected {
			t.Errorf("mode %q with HDR=%v: got %v, want %v", c.mode, c.source, got, c.expected)
		}
	}
}

func TestParseTonemapMode(t *testing.T) {
	cases := map[string]TonemapMode{
		"on": TonemapOn, "true": TonemapOn, "1": TonemapOn, "yes": TonemapOn,
		"off": TonemapOff, "false": TonemapOff, "0": TonemapOff, "no": TonemapOff,
		"auto": TonemapAuto, "": TonemapAuto, "unknown": TonemapAuto,
	}

	for in, want := range cases {
		if got := ParseTonemapMode(in); got != want {
			t.Errorf("ParseTonemapMode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildTonemapAutoWithHDRSource(t *testing.T) {
	// Режим auto: HDR-источник получает приведение диапазона
	o := DefaultOptions("input.mkv")
	o.SourceHDR = true
	o.Video.TonemapMode = TonemapAuto
	o.Video.MaxHeight = 720

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if !strings.Contains(valueOf(got, "-vf"), "tonemap=tonemap=") {
		t.Errorf("HDR source did not get tonemapping: %q", valueOf(got, "-vf"))
	}
}

func TestBuildTonemapAutoWithSDRSource(t *testing.T) {
	// Режим auto: обычный источник не трогается
	o := DefaultOptions("input.mkv")
	o.SourceHDR = false
	o.Video.TonemapMode = TonemapAuto
	o.Video.MaxHeight = 720

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if strings.Contains(valueOf(got, "-vf"), "tonemap") {
		t.Errorf("SDR source must not be tonemapped: %q", valueOf(got, "-vf"))
	}
}

func TestBuildTonemapOffOverridesHDR(t *testing.T) {
	o := DefaultOptions("input.mkv")
	o.SourceHDR = true
	o.Video.TonemapMode = TonemapOff

	got, err := Build(o, defaultSelector())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if strings.Contains(valueOf(got, "-vf"), "tonemap") {
		t.Errorf("explicit off must disable tonemapping: %q", valueOf(got, "-vf"))
	}
}
