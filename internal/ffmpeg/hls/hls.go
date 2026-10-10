// Package hls читает плейлисты ffmpeg и готовит их к раздаче клиенту.
package hls

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	ErrBadPath    = errors.New("invalid segment path")
	ErrNotReady   = errors.New("playlist is not ready")
	ErrNoSegments = errors.New("playlist has no segments")
)

// Segment описывает один сегмент из плейлиста ffmpeg.
type Segment struct {
	Name     string
	Duration float64
}

// Playlist описывает разобранный плейлист ffmpeg.
type Playlist struct {
	TargetDuration int
	MediaSequence  int
	Segments       []Segment
	Ended          bool
	// Version - версия протокола из плейлиста
	Version int
	// InitSegment - имя начального сегмента fmp4 из EXT-X-MAP
	InitSegment string
}

// Parse читает плейлист ffmpeg с диска.
func Parse(path string) (*Playlist, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotReady
		}
		return nil, err
	}
	defer func() { _ = file.Close() }()

	p := &Playlist{}
	scanner := bufio.NewScanner(file)

	var pendingDuration float64
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "#EXT-X-TARGETDURATION:"):
			p.TargetDuration = parseTagInt(line)

		case strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"):
			p.MediaSequence = parseTagInt(line)

		case strings.HasPrefix(line, "#EXTINF:"):
			pendingDuration = parseTagFloat(line)

		case strings.HasPrefix(line, "#EXT-X-VERSION:"):
			p.Version = parseTagInt(line)

		case strings.HasPrefix(line, "#EXT-X-MAP:"):
			p.InitSegment = parseInitSegment(line)

		case line == "#EXT-X-ENDLIST":
			p.Ended = true

		case strings.HasPrefix(line, "#"):
			// Остальные теги не влияют на раздачу сегментов

		default:
			p.Segments = append(p.Segments, Segment{
				Name:     line,
				Duration: pendingDuration,
			})
			pendingDuration = 0
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(p.Segments) == 0 {
		return nil, ErrNoSegments
	}
	return p, nil
}

// LocalName возвращает имя сегмента без префиксов и параметров запроса.
func (s Segment) LocalName() string {
	name := s.Name
	if idx := strings.IndexAny(name, "?#"); idx >= 0 {
		name = name[:idx]
	}
	return filepath.Base(name)
}

// Rewrite собирает плейлист с подставленными префиксом и суффиксом сегментов.
// Суффикс нужен для параметров доступа, которые плеер обязан повторить.
func (p *Playlist) Rewrite(prefix, suffix string) string {
	var b strings.Builder

	b.WriteString("#EXTM3U\n")

	version := p.Version
	if version <= 0 {
		version = 3
	}
	fmt.Fprintf(&b, "#EXT-X-VERSION:%d\n", version)
	b.WriteString("#EXT-X-PLAYLIST-TYPE:EVENT\n")

	target := p.TargetDuration
	if target <= 0 {
		target = 1
	}
	fmt.Fprintf(&b, "#EXT-X-TARGETDURATION:%d\n", target)
	fmt.Fprintf(&b, "#EXT-X-MEDIA-SEQUENCE:%d\n", p.MediaSequence)

	// Начальный сегмент fmp4 должен быть доступен плееру по нашему адресу
	if p.InitSegment != "" {
		fmt.Fprintf(&b, "#EXT-X-MAP:URI=\"%s%s%s\"\n", prefix, p.InitSegment, suffix)
	}

	for _, seg := range p.Segments {
		fmt.Fprintf(&b, "#EXTINF:%.6f,\n", seg.Duration)
		fmt.Fprintf(&b, "%s%s%s\n", prefix, seg.LocalName(), suffix)
	}

	if p.Ended {
		b.WriteString("#EXT-X-ENDLIST\n")
	}
	return b.String()
}

// LastSegmentIndex возвращает номер последнего сегмента в плейлисте.
func (p *Playlist) LastSegmentIndex() int {
	if len(p.Segments) == 0 {
		return p.MediaSequence - 1
	}
	return p.MediaSequence + len(p.Segments) - 1
}

// WaitForSegment ждёт появления сегмента и следующего за ним.
// Сегмент считается готовым, когда записан следующий или процесс завершился.
func WaitForSegment(playlistPath, baseDir, name string, timeout time.Duration, done <-chan struct{}) error {
	full, err := SafeSegmentPath(baseDir, name)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(timeout)
	for {
		if _, err := os.Stat(full); err == nil {
			// Сегмент дописан, если существует следующий или запись завершена
			if nextSegmentExists(baseDir, name) {
				return nil
			}
			if done != nil {
				select {
				case <-done:
					return nil
				default:
				}
			}
			// Размер перестал расти - сегмент дописан
			if stableSize(full) {
				return nil
			}
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("%w: segment %s", ErrNotReady, name)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// nextSegmentExists проверяет наличие сегмента со следующим номером.
func nextSegmentExists(dir, name string) bool {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	index, ok := trailingNumber(base)
	if !ok {
		return false
	}
	next := filepath.Join(dir, fmt.Sprintf("%s%d%s", strings.TrimRight(base, "0123456789"), index+1, ext))

	_, err := os.Stat(next)
	return err == nil
}

// trailingNumber возвращает число в конце строки.
func trailingNumber(s string) (int, bool) {
	end := len(s)
	start := end
	for start > 0 && s[start-1] >= '0' && s[start-1] <= '9' {
		start--
	}
	if start == end {
		return 0, false
	}

	v, err := strconv.Atoi(s[start:end])
	if err != nil {
		return 0, false
	}
	return v, true
}

// stableSize дважды проверяет размер файла, чтобы отличить растущий сегмент от готового.
func stableSize(path string) bool {
	first, err := os.Stat(path)
	if err != nil {
		return false
	}
	time.Sleep(50 * time.Millisecond)

	second, err := os.Stat(path)
	if err != nil {
		return false
	}
	return first.Size() == second.Size() && second.Size() > 0
}

// parseInitSegment извлекает имя начального сегмента из тега EXT-X-MAP.
func parseInitSegment(line string) string {
	_, value, ok := strings.Cut(line, ":")
	if !ok {
		return ""
	}

	_, after, found := strings.Cut(value, "URI=")
	if !found {
		return ""
	}

	// Значение имеет вид URI="init.mp4", после имени могут идти другие параметры
	after = strings.TrimSpace(after)
	if idx := strings.IndexByte(after, '"'); idx >= 0 {
		after = after[idx+1:]
		if end := strings.IndexByte(after, '"'); end >= 0 {
			return after[:end]
		}
	}
	return strings.Trim(after, "\"")
}

// SegmentIndex возвращает номер сегмента из имени файла.
func SegmentIndex(name string) (int, bool) {
	name = Segment{Name: name}.LocalName()
	base := strings.TrimSuffix(name, filepath.Ext(name))
	return trailingNumber(base)
}

// SafeSegmentPath проверяет, что путь лежит внутри каталога сессии.
func SafeSegmentPath(baseDir, name string) (string, error) {
	if name == "" {
		return "", ErrBadPath
	}
	// Имя сегмента не должно содержать разделителей пути или переходов вверх
	if strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return "", ErrBadPath
	}

	base, err := filepath.Abs(baseDir)
	if err != nil {
		return "", ErrBadPath
	}

	full := filepath.Join(base, name)
	rel, err := filepath.Rel(base, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", ErrBadPath
	}
	return full, nil
}

// BuildMaster собирает главный плейлист с одним вариантом.
func BuildMaster(variantURI string, bandwidth int, codecs string) string {
	var b strings.Builder

	b.WriteString("#EXTM3U\n")
	b.WriteString("#EXT-X-VERSION:3\n")
	fmt.Fprintf(&b, "#EXT-X-STREAM-INF:BANDWIDTH=%d", bandwidth)
	if codecs != "" {
		fmt.Fprintf(&b, ",CODECS=\"%s\"", codecs)
	}
	b.WriteString("\n")
	b.WriteString(variantURI)
	b.WriteString("\n")

	return b.String()
}

// parseTagInt разбирает числовое значение тега плейлиста.
func parseTagInt(line string) int {
	_, value, ok := strings.Cut(line, ":")
	if !ok {
		return 0
	}
	v, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return v
}

// parseTagFloat разбирает дробное значение тега плейлиста.
func parseTagFloat(line string) float64 {
	_, value, ok := strings.Cut(line, ":")
	if !ok {
		return 0
	}
	value = strings.TrimSuffix(strings.TrimSpace(value), ",")
	if idx := strings.IndexByte(value, ','); idx >= 0 {
		value = value[:idx]
	}

	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return v
}
