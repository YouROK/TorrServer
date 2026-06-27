package gstreamer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const gstProbeTimeout = 30 * time.Second

var (
	discovererDurationRe  = regexp.MustCompile(`(?i)Duration:\s*(\d+):(\d+):(\d+)(?:\.(\d+))?`)
	discovererContainerRe = regexp.MustCompile(`(?i)^container(?:\s+#\d+)?:\s*(.+)$`)
	discovererStreamRe    = regexp.MustCompile(`(?i)^(video|audio)(?:\s+#(\d+))?:\s*(.+)$`)
	discovererIntRe       = regexp.MustCompile(`-?\d+`)
	discovererRateRe      = regexp.MustCompile(`(\d+)\s*/\s*(\d+)`)
)

type ProbeInfo struct {
	DurationNS int64
	FileSize   int64
	Container  string
	Tracks     []TrackInfo
}

type TrackInfo struct {
	Index   int
	PadName string

	Type     string
	CapsName string

	Title    string
	Language string

	Width    int
	Height   int
	Channels int
	Rate     int

	FrameRateNum int
	FrameRateDen int
}

func (p ProbeInfo) DurationSeconds() int {
	if p.DurationNS <= 0 {
		return 0
	}
	return int(float64(p.DurationNS) / 1_000_000_000.0)
}

func (p ProbeInfo) Video() *TrackInfo {
	for i := range p.Tracks {
		t := &p.Tracks[i]
		if t.Type == "video" ||
			t.CapsName == "video/x-h264" ||
			t.CapsName == "video/x-h265" ||
			t.CapsName == "video/x-av1" ||
			t.CapsName == "video/x-vp9" ||
			t.CapsName == "video/x-vp8" {
			return t
		}
	}
	return nil
}

func (p ProbeInfo) VideoCapsName() string {
	if v := p.Video(); v != nil {
		return v.CapsName
	}
	return ""
}

func (p ProbeInfo) Audio() *TrackInfo {
	for i := range p.Tracks {
		if p.Tracks[i].Type == "audio" {
			return &p.Tracks[i]
		}
	}
	return nil
}

func (p ProbeInfo) HasAudio() bool {
	return p.Audio() != nil
}

func (p ProbeInfo) IsMatroskaContainer() bool {
	container := strings.ToLower(strings.TrimSpace(p.Container))
	return strings.Contains(container, "matroska") ||
		strings.Contains(container, "webm")
}

func (p ProbeInfo) IsH264() bool { return p.VideoCapsName() == "video/x-h264" }
func (p ProbeInfo) IsH265() bool { return p.VideoCapsName() == "video/x-h265" }
func (p ProbeInfo) IsAV1() bool  { return p.VideoCapsName() == "video/x-av1" }
func (p ProbeInfo) IsVP9() bool  { return p.VideoCapsName() == "video/x-vp9" }
func (p ProbeInfo) IsVP8() bool  { return p.VideoCapsName() == "video/x-vp8" }

func probeSource(sourceURL string, conf Config) (ProbeInfo, error) {
	output, err := runGSTDiscoverer(sourceURL, conf, gstProbeTimeout)
	if strings.TrimSpace(output) == "" {
		if err != nil {
			return ProbeInfo{}, err
		}
		return ProbeInfo{}, errors.New("gst-discoverer returned no output")
	}

	probe := probeFromDiscoverer(output)
	if len(probe.Tracks) == 0 {
		if err != nil {
			return ProbeInfo{}, fmt.Errorf("gst-discoverer parse failed: %w", err)
		}
		return ProbeInfo{}, errors.New("gst-discoverer returned no stream info")
	}
	return probe, nil
}

func runGSTDiscoverer(sourceURL string, conf Config, timeout time.Duration) (string, error) {
	bin, err := gstDiscovererPath(conf)
	if err != nil {
		return "", err
	}

	timeoutSeconds := int(timeout.Seconds())
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout+3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, "-v", "-t", strconv.Itoa(timeoutSeconds), sourceURL)
	cmd.Env = gstDiscovererEnv(conf)

	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return string(out), ctx.Err()
	}
	return string(out), err
}

func gstDiscovererPath(conf Config) (string, error) {
	path, _, err := gstDiscovererPathRoot(conf)
	return path, err
}

func gstDiscovererPathRoot(conf Config) (string, string, error) {
	name := "gst-discoverer-1.0"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	for _, root := range gstDiscovererRoots(conf) {
		path := filepath.Join(root, "bin", name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, root, nil
		}
	}

	if path, err := exec.LookPath(name); err == nil {
		return path, "", nil
	}
	return "", "", fmt.Errorf("%s not found", name)
}

func gstDiscovererEnv(conf Config) []string {
	env := os.Environ()
	env = setEnvValue(env, "LANG", "C.UTF-8")
	env = setEnvValue(env, "LC_ALL", "C.UTF-8")
	env = setEnvValue(env, "LANGUAGE", "en")
	env = setEnvValue(env, "GST_DEBUG_NO_COLOR", "1")

	roots := gstDiscovererSelectedRoots(conf)
	pathKey := "PATH"
	if runtime.GOOS == "windows" {
		pathKey = "Path"
	}
	var binDirs []string
	for _, root := range roots {
		binDirs = append(binDirs, filepath.Join(root, "bin"))
	}
	env = prependExistingPathValues(env, pathKey, binDirs)

	switch runtime.GOOS {
	case "linux":
		env = prependExistingPathValues(env, "LD_LIBRARY_PATH", gstDiscovererLibraryCandidates(roots))
	case "darwin":
		env = prependExistingPathValues(env, "DYLD_LIBRARY_PATH", gstDiscovererLibraryCandidates(roots))
	}

	if plugins := gstDiscovererPluginPath(roots); plugins != "" {
		env = setEnvValue(env, "GST_PLUGIN_PATH", plugins)
		env = setEnvValue(env, "GST_PLUGIN_SYSTEM_PATH_1_0", plugins)
	}
	if scanner := gstDiscovererScannerPath(roots); scanner != "" {
		env = setEnvValue(env, "GST_PLUGIN_SCANNER", scanner)
	}
	return env
}

func gstDiscovererSelectedRoots(conf Config) []string {
	_, root, err := gstDiscovererPathRoot(conf)
	if err != nil || root == "" {
		return nil
	}
	return []string{root}
}

func gstDiscovererRoots(conf Config) []string {
	var roots []string
	roots = appendAvailableProbeRoot(roots, conf.GSTPath)
	for _, root := range gstDiscovererDefaultRoots() {
		roots = appendAvailableProbeRoot(roots, root)
	}
	if runtime.GOOS == "windows" {
		if root := gstDiscovererPortableRoot(); root != "" {
			roots = appendAvailableProbeRoot(roots, root)
		}
		if root := embeddedGSTRuntimeRoot(); root != "" {
			roots = appendAvailableProbeRoot(roots, root)
		}
	}
	return roots
}

func gstDiscovererDefaultRoots() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			`C:\Program Files\gstreamer\1.0\mingw_x86_64`,
			`C:\gstreamer\1.0\mingw_x86_64`,
		}
	case "linux":
		return []string{
			"/usr",
			"/usr/local",
			"/opt/gstreamer",
			"/opt/gstreamer/1.0",
		}
	case "darwin":
		return []string{
			"/Library/Frameworks/GStreamer.framework/Versions/1.0",
			"/opt/homebrew",
			"/usr/local",
		}
	default:
		return nil
	}
}

func gstDiscovererPortableRoot() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	root := filepath.Join(filepath.Dir(exe), "gst-lib")
	if info, err := os.Stat(root); err == nil && info.IsDir() {
		return root
	}
	return ""
}

func appendAvailableProbeRoot(paths []string, path string) []string {
	if path == "" || !gstDiscovererRootHasBaseLibrary(path) {
		return paths
	}
	return appendUniqueProbePath(paths, path)
}

func gstDiscovererRootHasBaseLibrary(root string) bool {
	for _, candidate := range gstDiscovererBaseLibraryCandidates(root) {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}

func gstDiscovererBaseLibraryCandidates(root string) []string {
	switch runtime.GOOS {
	case "windows":
		return []string{filepath.Join(root, "bin", "libgstreamer-1.0-0.dll")}
	case "darwin":
		var candidates []string
		for _, dir := range gstDiscovererLibraryCandidates([]string{root}) {
			candidates = append(candidates,
				filepath.Join(dir, "libgstreamer-1.0.0.dylib"),
				filepath.Join(dir, "libgstreamer-1.0.dylib"),
			)
		}
		return candidates
	default:
		var candidates []string
		for _, dir := range gstDiscovererLibraryCandidates([]string{root}) {
			candidates = append(candidates, filepath.Join(dir, "libgstreamer-1.0.so.0"))
		}
		return candidates
	}
}

func gstDiscovererPluginPath(roots []string) string {
	if runtime.GOOS == "windows" {
		for _, root := range roots {
			path := filepath.Join(root, "lib", "gstreamer-1.0")
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				return path
			}
		}
		return ""
	}
	return firstExistingProbePath(gstDiscovererPluginCandidates(roots))
}

func gstDiscovererScannerPath(roots []string) string {
	name := "gst-plugin-scanner"
	if runtime.GOOS == "windows" {
		name += ".exe"
		for _, root := range roots {
			path := filepath.Join(root, "libexec", "gstreamer-1.0", name)
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				return path
			}
		}
		return ""
	}
	return firstExistingProbePath(gstDiscovererScannerCandidates(roots))
}

func gstDiscovererPluginCandidates(roots []string) []string {
	var candidates []string
	for _, root := range roots {
		candidates = append(candidates,
			filepath.Join(root, "lib", "gstreamer-1.0"),
			filepath.Join(root, "lib64", "gstreamer-1.0"),
			filepath.Join(root, "lib", runtime.GOARCH+"-linux-gnu", "gstreamer-1.0"),
			filepath.Join(root, "lib", "x86_64-linux-gnu", "gstreamer-1.0"),
			filepath.Join(root, "lib", "aarch64-linux-gnu", "gstreamer-1.0"),
		)
	}
	return candidates
}

func gstDiscovererLibraryCandidates(roots []string) []string {
	var candidates []string
	for _, root := range roots {
		candidates = append(candidates,
			filepath.Join(root, "lib"),
			filepath.Join(root, "lib64"),
			filepath.Join(root, "lib", runtime.GOARCH+"-linux-gnu"),
			filepath.Join(root, "lib", "x86_64-linux-gnu"),
			filepath.Join(root, "lib", "aarch64-linux-gnu"),
		)
	}
	return candidates
}

func gstDiscovererScannerCandidates(roots []string) []string {
	var candidates []string
	for _, root := range roots {
		candidates = append(candidates,
			filepath.Join(root, "libexec", "gstreamer-1.0", "gst-plugin-scanner"),
			filepath.Join(root, "lib", "gstreamer-1.0", "gst-plugin-scanner"),
			filepath.Join(root, "lib64", "gstreamer-1.0", "gst-plugin-scanner"),
		)
	}
	return candidates
}

func firstExistingProbePath(candidates []string) string {
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func appendUniqueProbePath(paths []string, path string) []string {
	if path == "" {
		return paths
	}
	clean := filepath.Clean(path)
	for _, existing := range paths {
		if strings.EqualFold(filepath.Clean(existing), clean) {
			return paths
		}
	}
	return append(paths, path)
}

func prependExistingPathValues(env []string, key string, values []string) []string {
	existing := make([]string, 0, len(values))
	for _, value := range values {
		if info, err := os.Stat(value); err == nil && info.IsDir() {
			existing = appendUniqueProbePath(existing, value)
		}
	}
	return prependPathValues(env, key, existing)
}

func prependPathValues(env []string, key string, values []string) []string {
	if len(values) == 0 {
		return env
	}

	current := ""
	for _, item := range env {
		name, val, ok := strings.Cut(item, "=")
		if ok && envKeyEqual(name, key) {
			current = val
			break
		}
	}

	separator := string(os.PathListSeparator)
	parts := make([]string, 0, len(values)+1)
	for _, value := range values {
		parts = appendUniqueProbePath(parts, value)
	}
	for _, part := range strings.Split(current, separator) {
		if part != "" {
			parts = appendUniqueProbePath(parts, part)
		}
	}

	return setEnvValue(env, key, strings.Join(parts, separator))
}

func setEnvValue(env []string, key string, value string) []string {
	prefix := key + "="
	for i, item := range env {
		name, _, _ := strings.Cut(item, "=")
		if envKeyEqual(name, key) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

func prependPathValue(env []string, value string) []string {
	pathKey := "PATH"
	if runtime.GOOS == "windows" {
		pathKey = "Path"
	}
	current := ""
	for _, item := range env {
		name, val, ok := strings.Cut(item, "=")
		if ok && envKeyEqual(name, pathKey) {
			current = val
			break
		}
	}
	if current == "" {
		return setEnvValue(env, pathKey, value)
	}
	for _, part := range strings.Split(current, string(os.PathListSeparator)) {
		if strings.EqualFold(part, value) {
			return env
		}
	}
	return setEnvValue(env, pathKey, value+string(os.PathListSeparator)+current)
}

func envKeyEqual(a string, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

func probeFromDiscoverer(text string) ProbeInfo {
	probe := ProbeInfo{DurationNS: parseDiscovererDurationNS(text)}
	var current *TrackInfo

	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		if match := discovererContainerRe.FindStringSubmatch(line); match != nil {
			container := strings.TrimSpace(match[1])
			if probe.Container == "" && container != "" {
				probe.Container = container
			}
			current = nil
			continue
		}

		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "subtitle") ||
			strings.HasPrefix(lower, "subtitles") ||
			strings.HasPrefix(lower, "properties:") {
			current = nil
			continue
		}

		if stream := parseDiscovererStreamHeader(line); stream != nil {
			probe.Tracks = append(probe.Tracks, *stream)
			current = &probe.Tracks[len(probe.Tracks)-1]
			continue
		}

		if current == nil {
			continue
		}
		parseDiscovererTrackLine(current, line)
	}

	videoIndex := 0
	audioIndex := 0
	for i := range probe.Tracks {
		switch probe.Tracks[i].Type {
		case "video":
			probe.Tracks[i].Index = videoIndex
			probe.Tracks[i].PadName = "video_" + strconv.Itoa(videoIndex)
			videoIndex++
		case "audio":
			probe.Tracks[i].Index = audioIndex
			probe.Tracks[i].PadName = "audio_" + strconv.Itoa(audioIndex)
			audioIndex++
		}
	}

	return probe
}

func parseDiscovererStreamHeader(line string) *TrackInfo {
	match := discovererStreamRe.FindStringSubmatch(line)
	if match == nil {
		return nil
	}

	trackType := strings.ToLower(match[1])
	codec := strings.TrimSpace(match[3])
	return &TrackInfo{
		Type:     trackType,
		CapsName: codecToCapsName(trackType, codec),
	}
}

func parseDiscovererTrackLine(track *TrackInfo, line string) {
	switch {
	case startsWithFold(line, "Width:"):
		track.Width = parseIntAfterColon(line)
	case startsWithFold(line, "Height:"):
		track.Height = parseIntAfterColon(line)
	case startsWithFold(line, "Channels:"):
		track.Channels = parseIntAfterColon(line)
	case startsWithFold(line, "Sample rate:"):
		track.Rate = parseIntAfterColon(line)
	case startsWithFold(line, "language code:"):
		track.Language = valueAfterColon(line)
	case startsWithFold(line, "language name:"):
		if track.Language == "" {
			track.Language = valueAfterColon(line)
		}
	case startsWithFold(line, "title:"):
		track.Title = valueAfterColon(line)
	case startsWithFold(line, "audio codec:"):
		if track.CapsName == "" {
			track.CapsName = codecToCapsName(track.Type, valueAfterColon(line))
		}
	case startsWithFold(line, "video codec:"):
		if track.CapsName == "" {
			track.CapsName = codecToCapsName(track.Type, valueAfterColon(line))
		}
	case startsWithFold(line, "Frame rate:"):
		track.FrameRateNum, track.FrameRateDen = parseDiscovererRate(valueAfterColon(line))
	}
}

func parseDiscovererDurationNS(text string) int64 {
	match := discovererDurationRe.FindStringSubmatch(text)
	if match == nil {
		return 0
	}

	hours, _ := strconv.ParseInt(match[1], 10, 64)
	minutes, _ := strconv.ParseInt(match[2], 10, 64)
	seconds, _ := strconv.ParseInt(match[3], 10, 64)

	nsText := match[4]
	if len(nsText) > 9 {
		nsText = nsText[:9]
	}
	nsText += strings.Repeat("0", 9-len(nsText))
	nanos, _ := strconv.ParseInt(nsText, 10, 64)

	return hours*3_600_000_000_000 + minutes*60_000_000_000 + seconds*1_000_000_000 + nanos
}

func codecToCapsName(kind string, values ...string) string {
	codec := strings.ToLower(strings.Join(values, " "))
	if codec == "" {
		return ""
	}

	if kind == "video" {
		switch {
		case strings.Contains(codec, "h264") || strings.Contains(codec, "h.264") || strings.Contains(codec, "avc"):
			return "video/x-h264"
		case strings.Contains(codec, "hevc") || strings.Contains(codec, "h265") || strings.Contains(codec, "h.265"):
			return "video/x-h265"
		case strings.Contains(codec, "av1"):
			return "video/x-av1"
		case strings.Contains(codec, "vp9"):
			return "video/x-vp9"
		case strings.Contains(codec, "vp8"):
			return "video/x-vp8"
		}
	}

	if kind == "audio" {
		switch {
		case strings.Contains(codec, "eac3") || strings.Contains(codec, "e-ac-3") || strings.Contains(codec, "e-ac3"):
			return "audio/x-eac3"
		case strings.Contains(codec, "ac3") || strings.Contains(codec, "ac-3") || strings.Contains(codec, "a/52"):
			return "audio/x-ac3"
		case strings.Contains(codec, "aac"):
			return "audio/mpeg"
		case strings.Contains(codec, "opus"):
			return "audio/x-opus"
		case strings.Contains(codec, "vorbis"):
			return "audio/x-vorbis"
		case strings.Contains(codec, "flac"):
			return "audio/x-flac"
		case strings.Contains(codec, "mpeg") || strings.Contains(codec, "mp3"):
			return "audio/mpeg"
		}
	}

	return ""
}

func parseDiscovererRate(value string) (int, int) {
	match := discovererRateRe.FindStringSubmatch(value)
	if match == nil {
		return 0, 0
	}

	num, err := strconv.Atoi(match[1])
	if err != nil || num <= 0 {
		return 0, 0
	}
	den, err := strconv.Atoi(match[2])
	if err != nil || den <= 0 {
		return 0, 0
	}
	return num, den
}

func parseIntAfterColon(line string) int {
	value := valueAfterColon(line)
	match := discovererIntRe.FindString(value)
	if match == "" {
		return 0
	}
	result, _ := strconv.Atoi(match)
	return result
}

func valueAfterColon(line string) string {
	_, value, ok := strings.Cut(line, ":")
	if !ok {
		return ""
	}
	value = strings.TrimSpace(value)
	if value == "<unknown>" {
		return ""
	}
	return value
}

func startsWithFold(value string, prefix string) bool {
	return len(value) >= len(prefix) && strings.EqualFold(value[:len(prefix)], prefix)
}
