package ffmpeg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"silo/internal/log"
)

const (
	Tag = "[FFmpeg]"
)

// versionRe разбирает версию из первой строки вывода ffmpeg -version.
var versionRe = regexp.MustCompile(`^ffmpeg version n?((?:[0-9]+\.?)+)`)

var (
	ErrNotInstalled = errors.New("ffmpeg not installed")
	ErrTooOld       = errors.New("ffmpeg version is too old")
)

// Binary описывает найденный исполняемый файл ffmpeg вместе с его возможностями.
type Binary struct {
	FFmpeg  string
	FFprobe string

	Version    Version
	VersionStr string

	Encoders map[string]struct{}
	Filters  map[string]struct{}
	Hwaccels map[string]struct{}

	// PauseSupport показывает, принимает ли сборка клавиши p и u в stdin.
	PauseSupport bool

	pauseOnce sync.Once
}

// Version - разобранная версия ffmpeg.
type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// AtLeast сравнивает версию с минимально допустимой, записанной строкой.
func (v Version) AtLeast(min string) bool {
	mv, err := ParseVersion(min)
	if err != nil {
		return true
	}
	return v.compare(mv) >= 0
}

func (v Version) compare(o Version) int {
	if v.Major != o.Major {
		return v.Major - o.Major
	}
	if v.Minor != o.Minor {
		return v.Minor - o.Minor
	}
	return v.Patch - o.Patch
}

// ParseVersion разбирает строку вида "6.1", "8.1.3" или "n6.0".
func ParseVersion(s string) (Version, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "n")
	parts := strings.Split(s, ".")
	var v Version
	for i, p := range parts {
		if i > 2 {
			break
		}
		n, err := strconv.Atoi(strings.TrimFunc(p, func(r rune) bool { return r < '0' || r > '9' }))
		if err != nil {
			return Version{}, fmt.Errorf("invalid version %q: %w", s, err)
		}
		switch i {
		case 0:
			v.Major = n
		case 1:
			v.Minor = n
		case 2:
			v.Patch = n
		}
	}
	if v.Major == 0 && v.Minor == 0 && v.Patch == 0 {
		return Version{}, fmt.Errorf("invalid version %q", s)
	}
	return v, nil
}

// candidateSource описывает место, где найден бинарник.
type candidateSource string

const (
	sourceConfig candidateSource = "config"
	sourcePath   candidateSource = "PATH"
	sourceNear   candidateSource = "executable dir"
)

type candidate struct {
	path   string
	source candidateSource
}

// binaryName добавляет расширение .exe для Windows.
func binaryName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// executableDir возвращает каталог запущенного файла.
func executableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err == nil {
		exe = resolved
	}
	return filepath.Dir(exe)
}

// findCandidates собирает пути к ffmpeg в порядке приоритета.
func findCandidates(explicit string) []candidate {
	var list []candidate
	seen := make(map[string]struct{})

	add := func(path string, source candidateSource) {
		if path == "" {
			return
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			abs = path
		}
		if _, ok := seen[abs]; ok {
			return
		}
		if st, err := os.Stat(abs); err != nil || st.IsDir() {
			return
		}
		seen[abs] = struct{}{}
		list = append(list, candidate{path: abs, source: source})
	}

	if explicit != "" {
		add(explicit, sourceConfig)
	}

	if dir := executableDir(); dir != "" {
		add(filepath.Join(dir, binaryName("ffmpeg")), sourceNear)
		add(filepath.Join(dir, "bin", binaryName("ffmpeg")), sourceNear)
	}

	if p, err := exec.LookPath("ffmpeg"); err == nil {
		add(p, sourcePath)
	}

	return list
}

// findProbePath ищет ffprobe рядом с ffmpeg и в PATH.
func findProbePath(explicit, ffmpegPath string) string {
	if explicit != "" {
		if abs, err := filepath.Abs(explicit); err == nil {
			if st, err := os.Stat(abs); err == nil && !st.IsDir() {
				return abs
			}
		}
		return ""
	}

	sibling := filepath.Join(filepath.Dir(ffmpegPath), binaryName("ffprobe"))
	if st, err := os.Stat(sibling); err == nil && !st.IsDir() {
		return sibling
	}

	if p, err := exec.LookPath("ffprobe"); err == nil {
		return p
	}

	if dir := executableDir(); dir != "" {
		near := filepath.Join(dir, binaryName("ffprobe"))
		if st, err := os.Stat(near); err == nil && !st.IsDir() {
			return near
		}
	}

	return ""
}

// probeBinary ищет подходящий ffmpeg и проверяет его возможности.
func probeBinary() (*Binary, error) {
	candidates := findCandidates("")
	if len(candidates) == 0 {
		return nil, ErrNotInstalled
	}

	var lastErr error
	cfg := DefaultConfig()
	for _, c := range candidates {
		bin, err := inspect(c.path, cfg)
		if err == nil {
			return bin, nil
		}
		lastErr = err
		log.Warnf("%s Candidate %s rejected: %v", Tag, c.path, err)
	}

	if lastErr == nil {
		lastErr = ErrNotInstalled
	}
	return nil, lastErr
}

// inspect проверяет один бинарник и собирает список его возможностей.
func inspect(path string, cfg Config) (*Binary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultStartTimeout)
	defer cancel()

	versionOut, err := runOutput(ctx, path, "-hide_banner", "-version")
	if err != nil {
		return nil, fmt.Errorf("failed to run -version: %w", err)
	}
	if strings.Contains(versionOut, "Libav developers") {
		return nil, errors.New("avconv is not supported")
	}

	version, versionStr, err := parseVersionOutput(versionOut)
	if err != nil {
		return nil, err
	}
	if !version.AtLeast(DefaultMinVersion) {
		return nil, fmt.Errorf("%w: %s is below minimum %s", ErrTooOld, version.String(), DefaultMinVersion)
	}

	probePath := findProbePath("", path)
	if probePath == "" {
		return nil, errors.New("ffprobe not found next to ffmpeg")
	}

	bin := &Binary{
		FFmpeg:     path,
		FFprobe:    probePath,
		Version:    version,
		VersionStr: versionStr,
	}

	encOut, err := runOutput(ctx, path, "-hide_banner", "-loglevel", "error", "-encoders")
	if err != nil {
		return nil, fmt.Errorf("failed to list encoders: %w", err)
	}
	bin.Encoders = parseCodecList(encOut)
	if len(bin.Encoders) == 0 {
		return nil, errors.New("no encoders reported")
	}

	filterOut, err := runOutput(ctx, path, "-hide_banner", "-loglevel", "error", "-filters")
	if err == nil {
		bin.Filters = parseFilterList(filterOut)
	}

	hwOut, err := runOutput(ctx, path, "-hide_banner", "-loglevel", "error", "-hwaccels")
	if err == nil {
		bin.Hwaccels = parseHwaccelList(hwOut)
	}

	return bin, nil
}

// parseVersionOutput достаёт версию из вывода ffmpeg -version.
func parseVersionOutput(out string) (Version, string, error) {
	line := out
	if idx := strings.IndexByte(out, '\n'); idx >= 0 {
		line = out[:idx]
	}
	line = strings.TrimSpace(line)

	m := versionRe.FindStringSubmatch(line)
	if m == nil {
		return Version{}, "", fmt.Errorf("cannot parse version from %q", line)
	}
	v, err := ParseVersion(m[1])
	if err != nil {
		return Version{}, "", err
	}
	return v, m[1], nil
}

// parseCodecList разбирает вывод -encoders или -decoders.
func parseCodecList(out string) map[string]struct{} {
	res := make(map[string]struct{})
	lines := strings.Split(out, "\n")
	started := false
	for _, line := range lines {
		if strings.HasPrefix(line, " ------") {
			started = true
			continue
		}
		if !started {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 || len(fields[0]) != 6 {
			continue
		}
		res[fields[1]] = struct{}{}
	}
	return res
}

// parseFilterList разбирает вывод -filters.
func parseFilterList(out string) map[string]struct{} {
	res := make(map[string]struct{})
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 || len(fields[0]) != 3 {
			continue
		}
		name := strings.Trim(fields[1], ".")
		if name != "" {
			res[name] = struct{}{}
		}
	}
	return res
}

// parseHwaccelList разбирает вывод -hwaccels.
func parseHwaccelList(out string) map[string]struct{} {
	res := make(map[string]struct{})
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Hardware acceleration") {
			continue
		}
		res[line] = struct{}{}
	}
	return res
}

// runOutput запускает команду и возвращает объединённый вывод.
func runOutput(ctx context.Context, bin string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return buf.String(), err
	}
	return buf.String(), nil
}

// SupportsEncoder проверяет наличие кодировщика.
func (b *Binary) SupportsEncoder(name string) bool {
	if b == nil {
		return false
	}
	_, ok := b.Encoders[name]
	return ok
}

// SupportsFilter проверяет наличие фильтра.
func (b *Binary) SupportsFilter(name string) bool {
	if b == nil {
		return false
	}
	_, ok := b.Filters[name]
	return ok
}

// SupportsHwaccel проверяет наличие метода аппаратного ускорения.
func (b *Binary) SupportsHwaccel(name string) bool {
	if b == nil {
		return false
	}
	_, ok := b.Hwaccels[name]
	return ok
}

// pauseKeyProbe проверяет, принимает ли сборка клавишу p в stdin.
// Заглушка кодируется одну секунду, в stdin отправляется "?" для вывода справки.
func (b *Binary) pauseKeyProbe() bool {
	if b == nil || b.FFmpeg == "" {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultStartTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, b.FFmpeg,
		"-hide_banner", "-re", "-f", "lavfi", "-i", "nullsrc=s=1x1:r=1:d=2", "-f", "null", "-")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return false
	}
	if err := cmd.Start(); err != nil {
		return false
	}

	_, _ = io.WriteString(stdin, "?")
	time.Sleep(300 * time.Millisecond)
	_, _ = io.WriteString(stdin, "q")
	_ = stdin.Close()

	_ = cmd.Wait()
	return strings.Contains(buf.String(), "pause transcoding")
}

// DetectPause определяет поддержку паузы через stdin и кэширует результат.
func (b *Binary) DetectPause() bool {
	if b == nil {
		return false
	}
	b.pauseOnce.Do(func() {
		b.PauseSupport = b.pauseKeyProbe()
	})
	return b.PauseSupport
}

// AutoVideoEncoder подбирает лучший доступный кодировщик для кодека.
func (b *Binary) AutoVideoEncoder(codec string, hwaccel string) string {
	if b == nil {
		return ""
	}

	if hwaccel != "" && hwaccel != "none" {
		hw := map[string]map[string]string{
			"vaapi": {"h264": "h264_vaapi", "hevc": "hevc_vaapi", "av1": "av1_vaapi", "vp9": "vp9_vaapi"},
			"nvenc": {"h264": "h264_nvenc", "hevc": "hevc_nvenc", "av1": "av1_nvenc"},
		}
		if m, ok := hw[hwaccel]; ok {
			if enc, ok := m[codec]; ok && b.SupportsEncoder(enc) {
				return enc
			}
		}
	}

	software := map[string][]string{
		"h264": {"libx264", "libopenh264", "h264"},
		"hevc": {"libx265", "hevc"},
		"av1":  {"libsvtav1", "libaom-av1", "librav1e", "av1"},
		"vp9":  {"libvpx-vp9", "vp9"},
		"vp8":  {"libvpx", "vp8"},
	}
	for _, name := range software[codec] {
		if b.SupportsEncoder(name) {
			return name
		}
	}
	return ""
}

// AutoAudioEncoder подбирает лучший доступный аудиокодировщик для кодека.
func (b *Binary) AutoAudioEncoder(codec string) string {
	if b == nil {
		return ""
	}
	software := map[string][]string{
		"aac":    {"libfdk_aac", "aac"},
		"opus":   {"libopus", "opus"},
		"mp3":    {"libmp3lame", "libshine", "mp3"},
		"ac3":    {"ac3", "eac3"},
		"eac3":   {"eac3"},
		"flac":   {"flac"},
		"vorbis": {"libvorbis", "vorbis"},
	}
	for _, name := range software[codec] {
		if b.SupportsEncoder(name) {
			return name
		}
	}
	return ""
}

// hwEncoderNames сопоставляет метод ускорения и кодек с именем кодировщика.
var hwEncoderNames = map[string]map[string]string{
	"vaapi": {"h264": "h264_vaapi", "hevc": "hevc_vaapi", "av1": "av1_vaapi", "vp9": "vp9_vaapi", "mpeg2video": "mpeg2_vaapi"},
	"nvenc": {"h264": "h264_nvenc", "hevc": "hevc_nvenc", "av1": "av1_nvenc"},
	"qsv":   {"h264": "h264_qsv", "hevc": "hevc_qsv", "av1": "av1_qsv", "vp9": "vp9_qsv", "mpeg2video": "mpeg2_qsv"},
	"amf":   {"h264": "h264_amf", "hevc": "hevc_amf", "av1": "av1_amf"},
}

// hwDevicePaths сопоставляет метод ускорения с устройствами для проверки.
var hwDevicePaths = map[string][]string{
	"vaapi": {"/dev/dri/renderD128", "/dev/dri/renderD129"},
	"nvenc": {"/dev/nvidia0", "/dev/nvidiactl"},
	"qsv":   {"/dev/dri/renderD128"},
}

// HWCapability описывает проверенную возможность аппаратного ускорения.
type HWCapability struct {
	// Method - метод ускорения: vaapi, nvenc, qsv
	Method string `json:"method"`
	// Device - устройство, прошедшее проверку
	Device string `json:"device,omitempty"`
	// Codecs - кодеки, для которых доступен аппаратный кодировщик
	Codecs []string `json:"codecs"`
	// Decode - доступно ли аппаратное декодирование
	Decode bool `json:"decode"`
}

// DetectHardware проверяет доступные методы ускорения и возвращает годный.
// Проверяется не только наличие кодировщика, но и работоспособность устройства:
// сборка может заявлять поддержку, не имея подходящего железа.
func (b *Binary) DetectHardware() []HWCapability {
	if b == nil {
		return nil
	}

	// Порядок предпочтения: сначала дискретные карты, затем встроенные
	order := []string{"nvenc", "vaapi", "qsv"}
	var found []HWCapability

	for _, method := range order {
		if !b.SupportsHwaccel(method) && !b.hasEncoderFor(method) {
			continue
		}

		device := b.workingDevice(method)
		if device == "" {
			continue
		}

		cap := HWCapability{
			Method: method,
			Device: device,
			Decode: b.supportsHwDecode(method),
		}
		for codec := range hwEncoderNames[method] {
			if b.SupportsEncoder(hwEncoderNames[method][codec]) {
				cap.Codecs = append(cap.Codecs, codec)
			}
		}
		sort.Strings(cap.Codecs)
		if len(cap.Codecs) > 0 {
			found = append(found, cap)
		}
	}
	return found
}

// hasEncoderFor сообщает, есть ли хотя бы один кодировщик метода.
func (b *Binary) hasEncoderFor(method string) bool {
	for _, name := range hwEncoderNames[method] {
		if b.SupportsEncoder(name) {
			return true
		}
	}
	return false
}

// supportsHwDecode проверяет наличие аппаратного декодирования.
func (b *Binary) supportsHwDecode(method string) bool {
	switch method {
	case "vaapi":
		return b.SupportsHwaccel("vaapi") && b.SupportsFilter("hwupload_vaapi")
	case "nvenc":
		return b.SupportsHwaccel("cuda")
	case "qsv":
		return b.SupportsHwaccel("qsv")
	default:
		return false
	}
}

// workingDevice находит устройство, на котором ускорение действительно работает.
func (b *Binary) workingDevice(method string) string {
	paths := hwDevicePaths[method]
	if len(paths) == 0 {
		return ""
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if b.probeDevice(method, path) {
			return path
		}
	}
	return ""
}

// probeDevice проверяет, что на устройстве можно создать контекст.
func (b *Binary) probeDevice(method, device string) bool {
	if b.FFmpeg == "" {
		return false
	}

	var args []string
	switch method {
	case "vaapi":
		args = []string{"-hide_banner", "-v", "error", "-init_hw_device", "vaapi=probe:" + device, "-f", "lavfi", "-i", "nullsrc=s=64x64:d=1", "-f", "null", "-"}
	case "nvenc":
		args = []string{"-hide_banner", "-v", "error", "-init_hw_device", "cuda=probe:0", "-f", "lavfi", "-i", "nullsrc=s=64x64:d=1", "-f", "null", "-"}
	case "qsv":
		args = []string{"-hide_banner", "-v", "error", "-init_hw_device", "qsv=probe", "-f", "lavfi", "-i", "nullsrc=s=64x64:d=1", "-f", "null", "-"}
	default:
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultStartTimeout)
	defer cancel()

	if _, err := runOutput(ctx, b.FFmpeg, args...); err != nil {
		log.Debugf("%s Hardware %s on %s is not usable: %v", Tag, method, device, err)
		return false
	}
	return true
}

// PreferredHWAccel возвращает лучшее доступное ускорение.
func (b *Binary) PreferredHWAccel() string {
	caps := b.DetectHardware()
	if len(caps) == 0 {
		return "none"
	}
	return caps[0].Method
}
