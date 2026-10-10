package ffmpeg

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"silo/internal/ffmpeg/args"
	"silo/internal/ffmpeg/media"
)

// realFFmpegPath возвращает путь к настоящему ffmpeg, если он доступен.
func realFFmpegPath(t *testing.T) string {
	t.Helper()

	candidates := []string{
		filepath.Join("..", "..", "dist", binaryName("ffmpeg")),
		"/usr/bin/ffmpeg",
	}

	for _, path := range candidates {
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			return path
		}
	}
	return ""
}

// setupRealModule поднимает модуль на настоящем ffmpeg.
func setupRealModule(t *testing.T) *FFmpeg {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping real ffmpeg test in short mode")
	}
	path := realFFmpegPath(t)
	if path == "" {
		t.Skip("ffmpeg binary is not available")
	}

	f := newWithBinary(DefaultConfig(), t.TempDir(), path)
	if !f.Enabled() {
		t.Skipf("transcoding is unavailable: %s", f.Reason())
	}
	t.Cleanup(func() {
		f.StopAll()
		f.Wait()
	})
	return f
}

// makeTestMedia создаёт короткий файл для проверки транскодирования.
func makeTestMedia(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "src.mkv")

	cmd := newTestCommand(realFFmpegPath(t),
		"-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc=s=320x240:r=25:d=4",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=4",
		"-c:v", "libx264", "-preset", "ultrafast",
		"-c:a", "aac",
		"-y", path,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create test media: %v (%s)", err, out)
	}
	return path
}

func TestRealProbe(t *testing.T) {
	f := setupRealModule(t)
	input := makeTestMedia(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := f.Probe(ctx, input)
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}

	if info.Container != "mkv" {
		t.Errorf("container = %q, want mkv", info.Container)
	}
	if !info.HasVideo() || !info.HasAudio() {
		t.Fatalf("expected video and audio, got %d streams", len(info.Streams))
	}
	if v := info.Video(); v.Width != 320 || v.Height != 240 {
		t.Errorf("resolution = %dx%d, want 320x240", v.Width, v.Height)
	}
	if v := info.Video(); v.Codec != "h264" {
		t.Errorf("video codec = %q, want h264", v.Codec)
	}
	if a := info.Audio(); a.Codec != "aac" {
		t.Errorf("audio codec = %q, want aac", a.Codec)
	}
	if info.Duration < 3.9 || info.Duration > 4.1 {
		t.Errorf("duration = %v, want about 4", info.Duration)
	}
}

func TestRealProbeCache(t *testing.T) {
	f := setupRealModule(t)
	input := makeTestMedia(t)

	ctx := context.Background()
	first, err := f.Probe(ctx, input)
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	second, err := f.Probe(ctx, input)
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	if first != second {
		t.Error("cached probe did not return the same result")
	}
}

func TestRealTranscodeToPipe(t *testing.T) {
	f := setupRealModule(t)
	input := makeTestMedia(t)

	o := f.DefaultTestOptions(input)
	o.Video.MaxWidth, o.Video.MaxHeight = 160, 120
	o.Container = "mp4"

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	j, err := f.Start(ctx, StartRequest{
		UserID:   "owner",
		Hash:     "aabbccddeeff00112233445566778899aabbccdd",
		FileIdx:  0,
		Options:  o,
		Duration: 4 * time.Second,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Поток вычитывается до конца: ffmpeg пишет в канал и встаёт, если его не читать
	total := 0
	buf := make([]byte, 32*1024)
	out := j.Output()
	for {
		n, err := out.Read(buf)
		total += n
		if err != nil {
			break
		}
	}
	if total < 1024 {
		t.Fatalf("read only %d bytes from the transcoded stream", total)
	}

	select {
	case <-j.Done():
	case <-time.After(60 * time.Second):
		t.Fatal("transcoding did not finish in time")
	}

	if j.State() != "done" {
		t.Errorf("state = %s, want done (stderr: %v)", j.State(), j.Stderr())
	}
	if j.Snapshot().OutputSize == 0 {
		t.Error("output size was not reported")
	}
}

func TestRealTranscodeScaleMatches(t *testing.T) {
	f := setupRealModule(t)
	input := makeTestMedia(t)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "scaled.mp4")

	o := f.DefaultTestOptions(input)
	o.Video.MaxWidth, o.Video.MaxHeight = 160, 120
	o.Container = "mp4"
	o.Output = outPath

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	j, err := f.Start(ctx, StartRequest{UserID: "owner", Hash: "h", Options: o, Duration: 4 * time.Second})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	select {
	case <-j.Done():
	case <-time.After(60 * time.Second):
		t.Fatal("transcoding did not finish in time")
	}
	if j.State() != "done" {
		t.Fatalf("state = %s, want done (stderr: %v)", j.State(), j.Stderr())
	}

	info, err := f.Probe(ctx, outPath)
	if err != nil {
		t.Fatalf("Probe result: %v", err)
	}
	v := info.Video()
	if v == nil {
		t.Fatal("result has no video stream")
	}
	if v.Width != 160 || v.Height != 120 {
		t.Errorf("result resolution = %dx%d, want 160x120", v.Width, v.Height)
	}
}

func TestRealHLSSegments(t *testing.T) {
	f := setupRealModule(t)
	input := makeTestMedia(t)

	o := f.DefaultTestOptions(input)
	o.Protocol = args.ProtocolHLS
	o.Container = "ts"
	o.HLS = args.DefaultHLSOptions()
	o.HLS.SegmentLength = 2
	o.HLS.ListSize = 3

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	j, err := f.Start(ctx, StartRequest{
		ID:       "hls-segments",
		UserID:   "owner",
		Hash:     "h",
		Options:  o,
		Duration: 4 * time.Second,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	select {
	case <-j.Done():
	case <-time.After(60 * time.Second):
		t.Fatal("transcoding did not finish in time")
	}
	if j.State() != "done" {
		t.Fatalf("state = %s, want done (stderr: %v)", j.State(), j.Stderr())
	}

	// Пути задаёт модуль: плейлист и сегменты лежат в каталоге сессии
	if j.PlaylistPath == "" {
		t.Fatal("playlist path was not set")
	}
	if _, err := os.Stat(j.PlaylistPath); err != nil {
		t.Fatalf("playlist was not created: %v", err)
	}

	segments, err := filepath.Glob(filepath.Join(j.WorkDir, "seg*.ts"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(segments) == 0 {
		t.Fatal("no segments were produced")
	}
	if len(segments) > 6 {
		t.Errorf("segments on disk = %d, expected the sliding window to cap them", len(segments))
	}
}

func TestRealModuleInfo(t *testing.T) {
	f := setupRealModule(t)

	info := f.Info()
	if !info.Enabled {
		t.Fatalf("module is disabled: %s", info.Reason)
	}
	if info.FFmpegPath == "" || info.FFprobePath == "" {
		t.Error("binary paths were not reported")
	}
	if info.Version == "" {
		t.Error("version was not reported")
	}
	if info.CacheDir == "" {
		t.Error("cache dir was not reported")
	}
	if len(info.Encoders) == 0 {
		t.Error("encoders were not reported")
	}
	if !f.Enabled() {
		t.Error("Enabled returned false for a working module")
	}
}

func TestRealEncoderSelection(t *testing.T) {
	f := setupRealModule(t)

	if enc := f.SelectVideoEncoder("h264", "none"); enc == "" {
		t.Error("no h264 encoder was selected")
	}
	if enc := f.SelectAudioEncoder("aac"); enc == "" {
		t.Error("no aac encoder was selected")
	}
	if enc := f.SelectVideoEncoder("unknown-codec", "none"); enc != "" {
		t.Errorf("unexpected encoder %q for an unknown codec", enc)
	}
}

func TestRealDisabledModule(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = false

	f := New(cfg, t.TempDir())
	if f.Enabled() {
		t.Error("module must be disabled when enabled=false")
	}
	if f.Reason() == "" {
		t.Error("reason must explain why transcoding is unavailable")
	}
	if _, err := f.Probe(context.Background(), "x"); err == nil {
		t.Error("Probe must fail when transcoding is disabled")
	}
	if _, err := f.Start(context.Background(), StartRequest{}); err == nil {
		t.Error("Start must fail when transcoding is disabled")
	}
}

// containsString проверяет наличие строки в списке.
func containsString(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

// hwDeviceAvailable проверяет наличие узла рендеринга.
func hwDeviceAvailable(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TestRealVaapiEncode проверяет аппаратное кодирование через VAAPI.
func TestRealVaapiEncode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping hardware test in short mode")
	}

	device := "/dev/dri/renderD128"
	if !hwDeviceAvailable(device) {
		t.Skipf("no VAAPI render node at %s", device)
	}

	f := setupRealModule(t)
	if !containsString(f.Info().Hwaccels, "vaapi") {
		t.Skip("this ffmpeg build has no vaapi")
	}

	input := makeTestMedia(t)

	o := f.DefaultTestOptions(input)
	o.Video.Codec = "h264"
	o.Video.HWAccel = "vaapi"
	o.Video.HWDevice = device
	o.Video.Encoder = "h264_vaapi"
	o.Video.CRF = 0
	o.Video.BitrateKbps = 2000
	o.Video.MaxWidth, o.Video.MaxHeight = 320, 240
	o.Container = "mkv"

	dir := t.TempDir()
	out := filepath.Join(dir, "hw.mkv")
	o.Output = out

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	j, err := f.Start(ctx, StartRequest{
		UserID:   "owner",
		Hash:     "0011223344556677889900aabbccddeeff001122",
		Options:  o,
		Duration: 4 * time.Second,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	select {
	case <-j.Done():
	case <-time.After(60 * time.Second):
		t.Fatal("hardware encoding did not finish in time")
	}

	if j.State() != "done" {
		t.Fatalf("state = %s, want done (stderr: %v)", j.State(), j.Stderr())
	}

	info, err := f.Probe(ctx, out)
	if err != nil {
		t.Fatalf("Probe result: %v", err)
	}
	v := info.Video()
	if v == nil {
		t.Fatal("result has no video stream")
	}
	if v.Codec != "h264" {
		t.Errorf("codec = %q, want h264", v.Codec)
	}
	if v.Width != 320 || v.Height != 240 {
		t.Errorf("resolution = %dx%d, want 320x240", v.Width, v.Height)
	}
	if v.Height == 0 {
		t.Error("result has no height")
	}
}

// TestRealNvencEncode проверяет аппаратное кодирование через NVENC.
func TestRealNvencEncode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping hardware test in short mode")
	}
	if !hwDeviceAvailable("/dev/nvidia0") {
		t.Skip("no NVIDIA device")
	}

	f := setupRealModule(t)
	if !containsString(f.Info().Hwaccels, "cuda") {
		t.Skip("this ffmpeg build has no cuda")
	}

	input := makeTestMedia(t)

	o := f.DefaultTestOptions(input)
	o.Video.Codec = "h264"
	o.Video.HWAccel = "nvenc"
	o.Video.Encoder = "h264_nvenc"
	o.Video.CRF = 0
	o.Video.BitrateKbps = 2000
	o.Video.MaxWidth, o.Video.MaxHeight = 320, 240
	o.Container = "mkv"

	dir := t.TempDir()
	o.Output = filepath.Join(dir, "nvenc.mkv")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	j, err := f.Start(ctx, StartRequest{
		UserID:   "owner",
		Hash:     "11223344556677889900aabbccddeeff00112233",
		Options:  o,
		Duration: 4 * time.Second,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	select {
	case <-j.Done():
	case <-time.After(60 * time.Second):
		t.Fatal("nvenc encoding did not finish in time")
	}

	if j.State() != "done" {
		t.Fatalf("state = %s, want done (stderr: %v)", j.State(), j.Stderr())
	}

	info, err := f.Probe(ctx, o.Output)
	if err != nil {
		t.Fatalf("Probe result: %v", err)
	}
	if v := info.Video(); v == nil || v.Codec != "h264" {
		t.Errorf("result video = %+v, want h264", v)
	}
}

// TestRealTonemapEncode проверяет приведение HDR к SDR на реальном файле.
func TestRealTonemapEncode(t *testing.T) {
	f := setupRealModule(t)

	dir := t.TempDir()
	hdr := filepath.Join(dir, "hdr.mkv")

	// Источник в 10 битах с признаками HDR
	cmd := newTestCommand(realFFmpegPath(t),
		"-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc2=s=320x180:r=25:d=2",
		"-pix_fmt", "yuv420p10le",
		"-c:v", "libx265", "-preset", "ultrafast",
		"-x265-params", "hdr-opt=1:colorprim=bt2020:transfer=smpte2084:colormatrix=bt2020nc",
		"-y", hdr,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("cannot create an HDR source: %v (%s)", err, out)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Режим auto: приведение диапазона включается по признаку HDR в источнике
	info, err := f.Probe(ctx, hdr)
	if err != nil {
		t.Fatalf("Probe HDR source: %v", err)
	}
	if !info.IsHDR() {
		t.Fatalf("test source was not detected as HDR: %+v", info.Video())
	}

	o := f.DefaultTestOptions(hdr)
	o.Video.Codec = "h264"
	o.Video.Encoder = f.SelectVideoEncoder("h264", "")
	o.SourceHDR = info.IsHDR()
	o.Video.TonemapMode = args.TonemapAuto
	o.Video.MaxWidth, o.Video.MaxHeight = 160, 90
	o.Container = "mkv"

	out := filepath.Join(dir, "sdr.mkv")
	o.Output = out

	j, err := f.Start(ctx, StartRequest{
		UserID:   "owner",
		Hash:     "223344556677889900aabbccddeeff0011223344",
		Options:  o,
		Duration: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	select {
	case <-j.Done():
	case <-time.After(60 * time.Second):
		t.Fatal("tonemapping did not finish in time")
	}
	if j.State() != "done" {
		t.Fatalf("state = %s, want done (stderr: %v)", j.State(), j.Stderr())
	}

	result, err := f.Probe(ctx, out)
	if err != nil {
		t.Fatalf("Probe result: %v", err)
	}
	v := result.Video()
	if v == nil {
		t.Fatal("result has no video stream")
	}
	// Результат должен быть 8-битным SDR
	if v.BitDepth > 8 {
		t.Errorf("bit depth = %d, want at most 8 after tonemapping", v.BitDepth)
	}
	if v.IsHDR() {
		t.Error("result is still marked as HDR")
	}
}

// TestRealSubtitleBurn проверяет впечатывание субтитров в видео.
func TestRealSubtitleBurn(t *testing.T) {
	f := setupRealModule(t)

	dir := t.TempDir()
	source := filepath.Join(dir, "sub.mkv")
	subPath := filepath.Join(dir, "subs.srt")

	srt := "1\n00:00:00,200 --> 00:00:01,800\nSilo test subtitle\n\n"
	if err := os.WriteFile(subPath, []byte(srt), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newTestCommand(realFFmpegPath(t),
		"-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc=s=320x180:r=25:d=2",
		"-i", subPath,
		"-c:v", "libx264", "-preset", "ultrafast",
		"-c:s", "srt",
		"-y", source,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("cannot create a source with subtitles: %v (%s)", err, out)
	}

	o := f.DefaultTestOptions(source)
	o.Video.Codec = "h264"
	o.Video.Encoder = f.SelectVideoEncoder("h264", "")
	o.Audio.Codec = "none"
	o.Subtitles.Mode = args.SubtitleBurn
	o.DropSubs = false
	o.Container = "mkv"

	out := filepath.Join(dir, "burned.mkv")
	o.Output = out

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	j, err := f.Start(ctx, StartRequest{
		UserID:   "owner",
		Hash:     "3344556677889900aabbccddeeff001122334455",
		Options:  o,
		Duration: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	select {
	case <-j.Done():
	case <-time.After(60 * time.Second):
		t.Fatal("subtitle burn did not finish in time")
	}
	if j.State() != "done" {
		t.Fatalf("state = %s, want done (stderr: %v)", j.State(), j.Stderr())
	}

	if _, err := os.Stat(out); err != nil {
		t.Fatalf("output file was not created: %v", err)
	}

	info, err := f.Probe(ctx, out)
	if err != nil {
		t.Fatalf("Probe result: %v", err)
	}
	if !info.HasVideo() {
		t.Error("result has no video stream")
	}
	// Впечатанные субтитры не остаются отдельными дорожками
	if info.Count(media.KindSubtitle) != 0 {
		t.Errorf("burned-in subtitles must not remain as tracks: %d", info.Count(media.KindSubtitle))
	}
}

// TestRealSubtitleCopy проверяет копирование дорожки субтитров.
func TestRealSubtitleCopy(t *testing.T) {
	f := setupRealModule(t)

	dir := t.TempDir()
	source := filepath.Join(dir, "sub.mkv")
	subPath := filepath.Join(dir, "subs.srt")

	srt := "1\n00:00:00,200 --> 00:00:01,800\nCopied subtitle\n\n"
	if err := os.WriteFile(subPath, []byte(srt), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := newTestCommand(realFFmpegPath(t),
		"-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc=s=320x180:r=25:d=2",
		"-i", subPath,
		"-c:v", "libx264", "-preset", "ultrafast",
		"-c:s", "srt",
		"-y", source,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("cannot create a source with subtitles: %v (%s)", err, out)
	}

	o := f.DefaultTestOptions(source)
	o.Video.Codec = "h264"
	o.Video.Encoder = f.SelectVideoEncoder("h264", "")
	o.Audio.Codec = "none"
	o.Subtitles.Mode = args.SubtitleCopy
	o.DropSubs = false
	o.Container = "mkv"

	out := filepath.Join(dir, "copied.mkv")
	o.Output = out

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	j, err := f.Start(ctx, StartRequest{
		UserID:   "owner",
		Hash:     "44556677889900aabbccddeeff00112233445566",
		Options:  o,
		Duration: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	select {
	case <-j.Done():
	case <-time.After(60 * time.Second):
		t.Fatal("subtitle copy did not finish in time")
	}
	if j.State() != "done" {
		t.Fatalf("state = %s, want done (stderr: %v)", j.State(), j.Stderr())
	}

	info, err := f.Probe(ctx, out)
	if err != nil {
		t.Fatalf("Probe result: %v", err)
	}
	if info.Count(media.KindSubtitle) == 0 {
		t.Error("copied subtitle track is missing")
	}
}

// TestRealTonemapSkippedForSDR проверяет, что обычный материал не трогается.
func TestRealTonemapSkippedForSDR(t *testing.T) {
	f := setupRealModule(t)
	input := makeTestMedia(t)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	info, err := f.Probe(ctx, input)
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	if info.IsHDR() {
		t.Skip("test media is HDR, nothing to check")
	}

	o := f.DefaultTestOptions(input)
	o.Video.Codec = "h264"
	o.Video.Encoder = f.SelectVideoEncoder("h264", "")
	o.SourceHDR = info.IsHDR()
	// Режим auto при обычном источнике не добавляет приведение диапазона
	o.Video.TonemapMode = args.TonemapAuto
	o.Container = "mkv"

	dir := t.TempDir()
	o.Output = filepath.Join(dir, "plain.mkv")

	j, err := f.Start(ctx, StartRequest{
		UserID:   "owner",
		Hash:     "556677889900aabbccddeeff0011223344556677",
		Options:  o,
		Duration: 4 * time.Second,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	select {
	case <-j.Done():
	case <-time.After(60 * time.Second):
		t.Fatal("transcoding did not finish in time")
	}
	if j.State() != "done" {
		t.Fatalf("state = %s, want done (stderr: %v)", j.State(), j.Stderr())
	}

	result, err := f.Probe(ctx, o.Output)
	if err != nil {
		t.Fatalf("Probe result: %v", err)
	}
	if v := result.Video(); v == nil {
		t.Fatal("result has no video stream")
	} else if v.BitDepth > 8 {
		t.Errorf("bit depth = %d, SDR source must stay 8 bit", v.BitDepth)
	}
}
