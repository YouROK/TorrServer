//go:build gst

package gstreamer

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxPendingVTTChars  = 64 * 1024
	subtitleWaitTimeout = 5 * time.Second
)

var vttCuePattern = regexp.MustCompile(`(?ms)(?:^|\n)(?:[^\n]*\n)?(\d{2,}:\d{2}:\d{2}[.,]\d{3})[ \t]+-->[ \t]+(\d{2,}:\d{2}:\d{2}[.,]\d{3})[^\n]*\n(.*?)(?:\n[ \t]*\n)`)

type subtitleCue struct {
	StartNS uint64
	EndNS   uint64
	Text    string
}

type subtitleCueKey struct {
	startNS uint64
	endNS   uint64
	text    string
}

type subtitleStore struct {
	mu          sync.RWMutex
	pending     string
	videoReadTo uint64
	cues        []subtitleCue
	seen        map[subtitleCueKey]struct{}
	updated     chan struct{}
}

func newSubtitleStore() *subtitleStore {
	return &subtitleStore{
		cues:    make([]subtitleCue, 0, 256),
		seen:    make(map[subtitleCueKey]struct{}),
		updated: make(chan struct{}),
	}
}

func (t *Task) setSubtitleStores(stores map[int]*subtitleStore) {
	if t == nil {
		return
	}

	t.subtitleMu.Lock()
	previous := t.subtitleStores
	t.subtitleStores = stores
	t.subtitleMu.Unlock()

	for _, store := range previous {
		store.notify()
	}
}

func (t *Task) subtitleStore(index int) *subtitleStore {
	if t == nil {
		return nil
	}
	t.subtitleMu.RLock()
	store := t.subtitleStores[index]
	t.subtitleMu.RUnlock()
	return store
}

func (s *subtitleStore) notify() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.notifyLocked()
	s.mu.Unlock()
}

func (s *subtitleStore) notifyLocked() {
	if s.updated != nil {
		close(s.updated)
	}
	s.updated = make(chan struct{})
}

func (s *subtitleStore) setVideoReadTo(value uint64) {
	if s == nil {
		return
	}
	s.mu.Lock()
	if s.videoReadTo != value {
		s.videoReadTo = value
		s.notifyLocked()
	}
	s.mu.Unlock()
}

func (s *subtitleStore) advanceVideoReadTo(value uint64) {
	if s == nil || value == 0 {
		return
	}
	s.mu.Lock()
	if value > s.videoReadTo {
		s.videoReadTo = value
		s.notifyLocked()
	}
	s.mu.Unlock()
}

func (r *gstRunner) resetSubtitleProgress(seconds float64) {
	value := subtitleProgressNS(seconds)
	for _, store := range r.subtitleStores {
		store.setVideoReadTo(value)
	}
}

func (r *gstRunner) advanceSubtitleProgress(value uint64) {
	for _, store := range r.subtitleStores {
		store.advanceVideoReadTo(value)
	}
}

func subtitleProgressNS(seconds float64) uint64 {
	if seconds <= 0 || math.IsNaN(seconds) {
		return 0
	}
	maximum := ^uint64(0)
	if math.IsInf(seconds, 1) || seconds >= float64(maximum)/1_000_000_000 {
		return maximum
	}
	return uint64(math.Round(seconds * 1_000_000_000))
}

func supportedSubtitleTrack(track TrackInfo) bool {
	if track.Type != "subtitle" {
		return false
	}
	switch strings.ToLower(track.Codec) {
	case "text", "subrip", "utf8", "ass", "ssa":
		return true
	default:
		return false
	}
}

func (r *gstRunner) drainSubtitles() error {
	if !r.task.Config.Subtitles || len(r.subtitleSinks) == 0 || gstRuntime == nil {
		return nil
	}
	for index, sink := range r.subtitleSinks {
		store := r.subtitleStores[index]
		if store == nil {
			continue
		}
		for {
			sample := gstRuntime.appSinkTryPullSample(sink, 0)
			if sample == 0 {
				break
			}
			var chunk string
			err := gstRuntime.withSampleBytes(sample, func(data []byte) error {
				chunk = string(data)
				return nil
			})
			gstRuntime.sampleUnref(sample)
			if err != nil {
				return err
			}
			store.appendVTT(chunk, r.positionSeekSeconds, r.maxSegmentDurationNS())
		}
	}
	return nil
}

func (r *gstRunner) maxSegmentDurationNS() uint64 {
	if r.task.Cue != nil && r.task.Cue.MaxDurationNS > 0 {
		return r.task.Cue.MaxDurationNS
	}
	return uint64(max(r.task.Config.SegmentSeconds, 1)) * 1_000_000_000
}

func (s *subtitleStore) appendVTT(chunk string, seekSeconds float64, maxBackDiffNS uint64) {
	if s == nil || strings.TrimSpace(chunk) == "" {
		return
	}
	chunk = strings.ReplaceAll(strings.ReplaceAll(chunk, "\r\n", "\n"), "\r", "\n")
	s.mu.Lock()
	defer s.mu.Unlock()

	vtt := s.pending + chunk
	consumed := 0
	for _, match := range vttCuePattern.FindAllStringSubmatchIndex(vtt, -1) {
		startNS, okStart := parseVTTClock(vtt[match[2]:match[3]])
		endNS, okEnd := parseVTTClock(vtt[match[4]:match[5]])
		text := strings.TrimSpace(vtt[match[6]:match[7]])
		consumed = match[1]
		if !okStart || !okEnd || endNS <= startNS || text == "" {
			continue
		}

		seekNS := uint64(0)
		if seekSeconds > 0 {
			seekNS = uint64(seekSeconds * 1_000_000_000)
		}
		minimumExpected := uint64(0)
		if seekNS > maxBackDiffNS {
			minimumExpected = seekNS - maxBackDiffNS
		}
		if seekNS > 0 && startNS < minimumExpected {
			startNS = saturatingAdd(startNS, seekNS)
			endNS = saturatingAdd(endNS, seekNS)
		}
		key := subtitleCueKey{startNS: startNS, endNS: endNS, text: text}
		if _, exists := s.seen[key]; !exists {
			s.seen[key] = struct{}{}
			s.cues = append(s.cues, subtitleCue{StartNS: startNS, EndNS: endNS, Text: text})
		}
	}
	if consumed > 0 {
		s.pending = vtt[consumed:]
	} else {
		s.pending = vtt
	}
	if len(s.pending) > maxPendingVTTChars {
		cut := len(s.pending) - maxPendingVTTChars
		if newline := strings.IndexByte(s.pending[cut:], '\n'); newline >= 0 {
			cut += newline + 1
		}
		s.pending = s.pending[cut:]
	}
	s.notifyLocked()
}

func (t *Task) SubtitleVTT(trackIndex int, segmentIndex int) (string, bool) {
	value, ready, _ := t.subtitleVTTSnapshot(trackIndex, segmentIndex)
	return value, ready
}

func (t *Task) WaitSubtitleVTT(ctx context.Context, trackIndex int, segmentIndex int, timeout time.Duration) (string, error) {
	if t == nil {
		return "", ErrTaskNotFound
	}
	if timeout <= 0 {
		value, _, _ := t.subtitleVTTSnapshot(trackIndex, segmentIndex)
		return value, nil
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		value, ready, updated := t.subtitleVTTSnapshot(trackIndex, segmentIndex)
		if ready {
			return value, nil
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-timer.C:
			value, _, _ = t.subtitleVTTSnapshot(trackIndex, segmentIndex)
			return value, nil
		case <-updated:
		}
	}
}

func (t *Task) subtitleVTTSnapshot(trackIndex int, segmentIndex int) (string, bool, <-chan struct{}) {
	if t == nil {
		return "", false, nil
	}
	fromNS, toNS := t.subtitleRange(segmentIndex)
	if t.disposed.Load() {
		return emptyVTT(fromNS), true, nil
	}

	store := t.subtitleStore(trackIndex)
	if store == nil {
		return emptyVTT(fromNS), true, nil
	}
	return store.renderVTT(fromNS, toNS)
}

func (t *Task) subtitleRange(segmentIndex int) (uint64, uint64) {
	if cue, ok := t.Cue.Segment(segmentIndex); ok {
		return cue.StartNS, cue.EndNS
	}

	fromNS := t.segmentStartNS(segmentIndex)
	durationNS := uint64(max(t.Config.SegmentSeconds, 1)) * 1_000_000_000
	toNS := saturatingAdd(fromNS, durationNS)
	if t.Probe.DurationNS > 0 {
		mediaEndNS := uint64(t.Probe.DurationNS)
		if fromNS >= mediaEndNS {
			toNS = fromNS
		} else if toNS > mediaEndNS {
			toNS = mediaEndNS
		}
	}
	return fromNS, toNS
}

func (s *subtitleStore) renderVTT(fromNS uint64, toNS uint64) (string, bool, <-chan struct{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result strings.Builder
	writeVTTHeader(&result, fromNS)
	for _, cue := range s.cues {
		if cue.EndNS <= fromNS || cue.StartNS >= toNS {
			continue
		}
		startNS := uint64(0)
		if cue.StartNS > fromNS {
			startNS = cue.StartNS - fromNS
		}
		endNS := toNS - fromNS
		if cue.EndNS < toNS {
			endNS = cue.EndNS - fromNS
		}
		if endNS <= startNS {
			continue
		}
		result.WriteString(formatVTTClock(startNS))
		result.WriteString(" --> ")
		result.WriteString(formatVTTClock(endNS))
		result.WriteByte('\n')
		result.WriteString(cue.Text)
		result.WriteString("\n\n")
	}
	ready := toNS <= fromNS || s.videoReadTo >= toNS
	return result.String(), ready, s.updated
}

func emptyVTT(startNS uint64) string {
	var result strings.Builder
	writeVTTHeader(&result, startNS)
	return result.String()
}

func writeVTTHeader(result *strings.Builder, startNS uint64) {
	result.WriteString("WEBVTT\nX-TIMESTAMP-MAP=LOCAL:00:00:00.000,MPEGTS:")
	result.WriteString(strconv.FormatUint(clockTimeToMPEGTS(startNS), 10))
	result.WriteString("\n\n")
}

func parseVTTClock(value string) (uint64, bool) {
	value = strings.TrimSpace(strings.Replace(value, ",", ".", 1))
	parts := strings.Split(value, ":")
	if len(parts) != 3 {
		return 0, false
	}
	hours, errH := strconv.ParseUint(parts[0], 10, 64)
	minutes, errM := strconv.ParseUint(parts[1], 10, 64)
	secondsParts := strings.Split(parts[2], ".")
	if errH != nil || errM != nil || len(secondsParts) != 2 || minutes > 59 {
		return 0, false
	}
	seconds, errS := strconv.ParseUint(secondsParts[0], 10, 64)
	if errS != nil || seconds > 59 {
		return 0, false
	}
	millisecondsText := secondsParts[1]
	if len(millisecondsText) > 3 {
		millisecondsText = millisecondsText[:3]
	}
	millisecondsText += strings.Repeat("0", 3-len(millisecondsText))
	milliseconds, errMS := strconv.ParseUint(millisecondsText, 10, 64)
	if errMS != nil {
		return 0, false
	}
	return (hours*3600+minutes*60+seconds)*1_000_000_000 + milliseconds*1_000_000, true
}

func formatVTTClock(valueNS uint64) string {
	totalMilliseconds := valueNS / 1_000_000
	milliseconds := totalMilliseconds % 1000
	totalSeconds := totalMilliseconds / 1000
	seconds := totalSeconds % 60
	minutes := totalSeconds / 60 % 60
	hours := totalSeconds / 3600
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}

func clockTimeToMPEGTS(valueNS uint64) uint64 {
	return valueNS/1_000_000_000*90_000 + valueNS%1_000_000_000*90_000/1_000_000_000
}

func saturatingAdd(left uint64, right uint64) uint64 {
	if ^uint64(0)-left < right {
		return ^uint64(0)
	}
	return left + right
}
