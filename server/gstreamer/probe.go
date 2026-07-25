//go:build gst

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
	discovererContainerRe = regexp.MustCompile(`(?i)^(?:container(?:\s+#\d+)?|container[\s-]+format)\s*:\s*(.+)$`)
	discovererStreamRe    = regexp.MustCompile(`(?i)^(video|audio|subtitle|subtitles)(?:\s+#(\d+))?:\s*(.+)$`)
	discovererIntRe       = regexp.MustCompile(`-?\d+`)
	discovererRateRe      = regexp.MustCompile(`(\d+)\s*/\s*(\d+)`)
)

type ProbeInfo struct {
	DurationNS        int64
	FileSize          int64
	Container         string
	ContainerCapsName string
	Tracks            []TrackInfo
}

type TrackInfo struct {
	Index   int
	PadName string

	Type     string
	Codec    string
	CapsName string

	Title    string
	Language string

	Width    int
	Height   int
	Channels int
	Rate     int

	FrameRateNum int
	FrameRateDen int

	Colorimetry             string
	Transfer                string
	Primaries               string
	Matrix                  string
	BitDepth                int
	HasMasteringDisplayInfo bool
	HasContentLightLevel    bool
	IsDolbyVision           bool
	DolbyVisionProfile      int
	VideoTransfer           string
}

func (p ProbeInfo) DurationSeconds() int {
	if p.DurationNS <= 0 {
		return 0
	}
	return int(p.DurationNS / int64(time.Second))
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

func (p ProbeInfo) AudioTrack(index int) *TrackInfo {
	fallback := -1
	for i := range p.Tracks {
		if p.Tracks[i].Type != "audio" {
			continue
		}
		if fallback < 0 {
			fallback = i
		}
		if p.Tracks[i].Index == index {
			return &p.Tracks[i]
		}
	}
	if fallback >= 0 {
		return &p.Tracks[fallback]
	}
	return nil
}

func (p ProbeInfo) HasAudio() bool {
	return p.Audio() != nil
}

func (t TrackInfo) IsAACAudio() bool {
	if t.Type != "audio" {
		return false
	}
	codec := strings.ToLower(t.Codec)
	return strings.Contains(codec, "aac") ||
		strings.Contains(codec, "mp4a") ||
		strings.Contains(codec, "mpeg-4") ||
		strings.Contains(codec, "mpegversion=(int)4") ||
		strings.Contains(codec, "mpegversion=4")
}

func (t TrackInfo) IsHDRVideo() bool {
	return t.Type == "video" && (t.IsDolbyVision || t.VideoTransfer == "pq" || t.VideoTransfer == "hlg")
}

func (p ProbeInfo) IsMatroskaContainer() bool {
	container := strings.ToLower(strings.TrimSpace(p.Container + " " + p.ContainerCapsName))
	return strings.Contains(container, "matroska") ||
		strings.Contains(container, "webm")
}

func (p ProbeInfo) IsAVIContainer() bool {
	container := strings.ToLower(strings.TrimSpace(p.Container + " " + p.ContainerCapsName))
	return strings.Contains(container, "video/x-msvideo") || strings.Contains(container, "avi")
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
	name := gstDiscovererExecutableName()

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

func gstDiscovererExecutableName() string {
	if runtime.GOOS == "windows" {
		return "gst-discoverer-1.0.exe"
	}
	return "gst-discoverer-1.0"
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
		env = prependExistingPathValues(env, "LD_LIBRARY_PATH", gstLibraryDirCandidates(roots))
	case "darwin":
		env = prependExistingPathValues(env, "DYLD_LIBRARY_PATH", gstLibraryDirCandidates(roots))
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
	roots = appendAvailableGSTRoot(roots, conf.GSTPath)
	for _, root := range gstDefaultRuntimeRoots() {
		roots = appendAvailableGSTRoot(roots, root)
	}
	if runtime.GOOS == "windows" {
		if root := portableGSTRuntimeRoot(); root != "" {
			roots = appendAvailableGSTRoot(roots, root)
		}
		if !gstDiscovererExecutableAvailable(roots) {
			root := embeddedGSTRuntimeRoot()
			roots = appendAvailableGSTRoot(roots, root)
		}
	}
	return roots
}

func gstDiscovererExecutableAvailable(roots []string) bool {
	name := gstDiscovererExecutableName()
	for _, root := range roots {
		if info, err := os.Stat(filepath.Join(root, "bin", name)); err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}

func gstDiscovererPluginPath(roots []string) string {
	return firstExistingPath(gstPluginCandidates(roots))
}

func gstDiscovererScannerPath(roots []string) string {
	return firstExistingPath(gstPluginScannerCandidates(roots))
}

func prependExistingPathValues(env []string, key string, values []string) []string {
	existing := make([]string, 0, len(values))
	for _, value := range values {
		if info, err := os.Stat(value); err == nil && info.IsDir() {
			existing = appendUniquePath(existing, value)
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
		parts = appendUniquePath(parts, value)
	}
	for _, part := range strings.Split(current, separator) {
		if part != "" {
			parts = appendUniquePath(parts, part)
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
				probe.ContainerCapsName = containerToCapsName(container)
			}
			current = nil
			continue
		}
		if caps := containerCapsFromLine(line); caps != "" {
			probe.Container = caps
			probe.ContainerCapsName = caps
			current = nil
			continue
		}

		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "properties:") {
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
	subtitleIndex := 0
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
		case "subtitle":
			probe.Tracks[i].Index = subtitleIndex
			probe.Tracks[i].PadName = "subtitle_" + strconv.Itoa(subtitleIndex)
			subtitleIndex++
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
	if trackType == "subtitles" {
		trackType = "subtitle"
	}
	codec := strings.TrimSpace(match[3])
	if trackType == "subtitle" {
		codec = subtitleCodec(codec)
	}
	return &TrackInfo{
		Type:     trackType,
		Codec:    codec,
		CapsName: codecToCapsName(trackType, codec),
	}
}

func parseDiscovererTrackLine(track *TrackInfo, line string) {
	parseVideoMetadata(track, line)
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
		if track.Codec == "" {
			track.Codec = valueAfterColon(line)
		}
		if track.CapsName == "" {
			track.CapsName = codecToCapsName(track.Type, track.Codec)
		}
	case startsWithFold(line, "video codec:"):
		if track.Codec == "" {
			track.Codec = valueAfterColon(line)
		}
		if track.CapsName == "" {
			track.CapsName = codecToCapsName(track.Type, track.Codec)
		}
	case startsWithFold(line, "subtitle codec:"):
		track.Codec = subtitleCodec(valueAfterColon(line))
		track.CapsName = codecToCapsName(track.Type, track.Codec)
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

	if kind == "subtitle" {
		switch subtitleCodec(codec) {
		case "text", "subrip", "utf8":
			return "text/x-raw"
		case "ass":
			return "application/x-ass"
		case "ssa":
			return "application/x-ssa"
		case "pgs":
			return "subpicture/x-pgs"
		case "dvd":
			return "subpicture/x-dvd"
		case "kate":
			return "subtitle/x-kate"
		default:
			return "application/x-subtitle-unknown"
		}
	}

	return ""
}

func subtitleCodec(value string) string {
	codec := strings.ToLower(strings.TrimSpace(value))
	switch {
	case strings.Contains(codec, "subrip") || strings.Contains(codec, "srt"):
		return "subrip"
	case strings.Contains(codec, "utf8") || strings.Contains(codec, "utf-8"):
		return "utf8"
	case strings.Contains(codec, "ass"):
		return "ass"
	case strings.Contains(codec, "ssa"):
		return "ssa"
	case strings.Contains(codec, "pgs"):
		return "pgs"
	case strings.Contains(codec, "dvd"):
		return "dvd"
	case strings.Contains(codec, "kate"):
		return "kate"
	case strings.Contains(codec, "text"):
		return "text"
	default:
		return codec
	}
}

func parseVideoMetadata(track *TrackInfo, line string) {
	if track == nil || track.Type != "video" {
		return
	}
	if track.Colorimetry == "" {
		track.Colorimetry = metadataField(line, "colorimetry")
	}
	if track.Transfer == "" {
		track.Transfer = firstNonEmpty(
			metadataField(line, "transfer"),
			metadataField(line, "transfer-characteristics"),
			metadataField(line, "transfer-function"),
		)
	}
	if track.Primaries == "" {
		track.Primaries = firstNonEmpty(metadataField(line, "primaries"), metadataField(line, "color-primaries"))
	}
	if track.Matrix == "" {
		track.Matrix = firstNonEmpty(metadataField(line, "matrix"), metadataField(line, "matrix-coefficients"))
	}
	if track.BitDepth == 0 {
		for _, name := range []string{"bit-depth-luma", "bit-depth", "bits-per-component"} {
			if value := metadataField(line, name); value != "" {
				track.BitDepth, _ = strconv.Atoi(firstInteger(value))
				if track.BitDepth > 0 {
					break
				}
			}
		}
		upper := strings.ToUpper(line)
		if track.BitDepth == 0 && (strings.Contains(upper, "P010") || strings.Contains(upper, "10LE") || strings.Contains(upper, "10BE")) {
			track.BitDepth = 10
		} else if track.BitDepth == 0 && (strings.Contains(upper, "P012") || strings.Contains(upper, "12LE") || strings.Contains(upper, "12BE")) {
			track.BitDepth = 12
		}
	}
	lower := strings.ToLower(line)
	track.HasMasteringDisplayInfo = track.HasMasteringDisplayInfo || strings.Contains(lower, "mastering-display-info")
	track.HasContentLightLevel = track.HasContentLightLevel || strings.Contains(lower, "content-light-level") || strings.Contains(lower, "max-cll")
	track.IsDolbyVision = track.IsDolbyVision || strings.Contains(lower, "dolby vision") || strings.Contains(lower, "video/x-dolby-vision") || strings.Contains(lower, "dovi")
	if track.DolbyVisionProfile == 0 {
		if profile := metadataField(line, "profile"); track.IsDolbyVision && profile != "" {
			track.DolbyVisionProfile, _ = strconv.Atoi(firstInteger(profile))
		}
	}
	transfer := strings.ToLower(track.Transfer + " " + track.Colorimetry + " " + line)
	switch {
	case strings.Contains(transfer, "smpte2084") || strings.Contains(transfer, "st2084") || strings.Contains(transfer, "pq"):
		track.VideoTransfer = "pq"
	case strings.Contains(transfer, "arib-std-b67") || strings.Contains(transfer, "hlg"):
		track.VideoTransfer = "hlg"
	}
}

func metadataField(line string, name string) string {
	lower := strings.ToLower(line)
	needle := strings.ToLower(name)
	index := strings.Index(lower, needle)
	if index < 0 {
		return ""
	}
	rest := strings.TrimSpace(line[index+len(name):])
	rest = strings.TrimLeft(rest, " =:")
	if strings.HasPrefix(rest, "(") {
		if close := strings.Index(rest, ")"); close >= 0 {
			rest = strings.TrimSpace(rest[close+1:])
		}
	}
	if comma := strings.IndexByte(rest, ','); comma >= 0 {
		rest = rest[:comma]
	}
	return strings.Trim(strings.TrimSpace(rest), "\"'")
}

func firstInteger(value string) string {
	return discovererIntRe.FindString(value)
}

func containerToCapsName(value string) string {
	if direct := containerCapsFromLine(value); direct != "" {
		return direct
	}
	lower := strings.ToLower(value)
	switch {
	case strings.Contains(lower, "webm"):
		return "video/webm"
	case strings.Contains(lower, "matroska"):
		return "video/x-matroska"
	case strings.Contains(lower, "avi"):
		return "video/x-msvideo"
	case strings.Contains(lower, "quicktime") || strings.Contains(lower, "mp4"):
		return "video/quicktime"
	default:
		return ""
	}
}

func containerCapsFromLine(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	if comma := strings.IndexByte(lower, ','); comma >= 0 {
		lower = lower[:comma]
	}
	lower = strings.Trim(lower, "\"'")
	for _, caps := range []string{
		"audio/x-matroska", "video/x-matroska", "video/x-matroska-3d",
		"audio/webm", "video/webm", "video/x-msvideo", "video/quicktime",
		"video/mp4", "video/mpegts", "application/ogg", "video/x-flv",
	} {
		if lower == caps {
			return caps
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
