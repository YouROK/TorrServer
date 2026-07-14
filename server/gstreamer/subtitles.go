package gstreamer

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrSubtitleNotFound        = errors.New("subtitle track not found")
	ErrUnsupportedSubtitle     = errors.New("unsupported subtitle codec")
	ErrSubtitleTimelineChanged = errors.New("subtitle timeline changed")
	assOverrideTagRe          = regexp.MustCompile(`\{[^}]*\}`)
	srtTimingLineRe           = regexp.MustCompile(`(?m)^\s*\d{1,2}:\d{2}:\d{2}[,.]\d{1,3}\s+-->\s+\d{1,2}:\d{2}:\d{2}[,.]\d{1,3}.*$`)
	srtIndexLineRe            = regexp.MustCompile(`(?m)^\s*\d+\s*$`)
)

type subtitleCue struct {
	Start time.Duration
	End   time.Duration
	Text  string
}

type subtitleCueKey struct {
	Start time.Duration
	End   time.Duration
	Text  string
}

type subtitleTrackState struct {
	cues []subtitleCue
	seen map[subtitleCueKey]struct{}
}

type subtitleSegmentTimeline struct {
	Generation    uint64
	SourceStart   time.Duration
	SourceEnd     time.Duration
	PlaylistStart time.Duration
	PlaylistEnd   time.Duration
}

func (t *Task) SubtitleSegment(ctx context.Context, subtitleIndex int, segmentIndex int) ([]byte, error) {
	if segmentIndex < 0 {
		return nil, errors.New("invalid subtitle segment index")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	track := t.Probe.SubtitleTrack(subtitleIndex)
	if track == nil {
		return nil, ErrSubtitleNotFound
	}
	if !track.IsSupportedSubtitle() {
		return nil, ErrUnsupportedSubtitle
	}

	started := time.Now()
	waits := 0
	lastWaitLog := time.Time{}
	var generation uint64
	for {
		t.subtitleMu.Lock()
		currentGeneration := t.subtitleGeneration
		if generation == 0 && currentGeneration > 0 {
			generation = currentGeneration
		}
		timeline, ready := t.subtitleSegments[segmentIndex]
		unavailable := false
		unavailableReason := ""
		if !ready {
			unavailable, unavailableReason = t.subtitleSegmentUnavailableLocked(segmentIndex)
		}
		var cues []subtitleCue
		if ready {
			cues = t.subtitleCuesForSegmentLocked(subtitleIndex, timeline.SourceStart, timeline.SourceEnd)
		}
		generationStart := t.subtitleGenerationStart
		firstReady := t.subtitleFirstReady
		highestReady := t.subtitleHighestReady
		notify := t.subtitleNotifyLocked()
		disposed := t.disposed.Load()
		t.subtitleMu.Unlock()

		if disposed {
			return nil, ErrTaskNotFound
		}
		if ready {
			sourceCueStart, sourceCueEnd := subtitleCueRange(cues)
			data := buildWebVTTSegment(cues, timeline.SourceStart, timeline.SourceEnd)
			sourceDuration := timeline.SourceEnd - timeline.SourceStart
			playlistDuration := timeline.PlaylistEnd - timeline.PlaylistStart
			gstDebugf("subtitle timeline response task=%s file=%s subtitle=%d segment=%d generation=%d sourceStart=%.3f sourceEnd=%.3f sourceDuration=%.3f playlistStart=%.3f playlistEnd=%.3f playlistDuration=%.3f mode=source sourceCues=%d sourceCueStart=%.3f sourceCueEnd=%.3f bytes=%d waits=%d duration=%s", t.ID, t.FileID, subtitleIndex, segmentIndex, timeline.Generation, timeline.SourceStart.Seconds(), timeline.SourceEnd.Seconds(), sourceDuration.Seconds(), timeline.PlaylistStart.Seconds(), timeline.PlaylistEnd.Seconds(), playlistDuration.Seconds(), len(cues), sourceCueStart.Seconds(), sourceCueEnd.Seconds(), len(data), waits, time.Since(started))
			return data, nil
		}

		if unavailable {
			data := emptyWebVTTSegment()
			gstDebugf("subtitle timeline skipped task=%s file=%s subtitle=%d segment=%d requestGeneration=%d currentGeneration=%d generationStart=%d firstReady=%d highestReady=%d reason=%s bytes=%d waits=%d duration=%s", t.ID, t.FileID, subtitleIndex, segmentIndex, generation, currentGeneration, generationStart, firstReady, highestReady, unavailableReason, len(data), waits, time.Since(started))
			return data, nil
		}
		if generation > 0 && currentGeneration != generation {
			gstDebugf("subtitle timeline request rebound task=%s file=%s subtitle=%d segment=%d previousGeneration=%d currentGeneration=%d waits=%d duration=%s", t.ID, t.FileID, subtitleIndex, segmentIndex, generation, currentGeneration, waits, time.Since(started))
			generation = currentGeneration
		}

		waits++
		now := time.Now()
		if lastWaitLog.IsZero() || now.Sub(lastWaitLog) >= time.Second {
			lastWaitLog = now
			gstDebugf("subtitle timeline wait task=%s file=%s subtitle=%d segment=%d requestGeneration=%d currentGeneration=%d waits=%d elapsed=%s", t.ID, t.FileID, subtitleIndex, segmentIndex, generation, currentGeneration, waits, now.Sub(started))
		}
		select {
		case <-ctx.Done():
			gstDebugf("subtitle timeline request context done task=%s file=%s subtitle=%d segment=%d generation=%d waits=%d err=%v duration=%s", t.ID, t.FileID, subtitleIndex, segmentIndex, generation, waits, ctx.Err(), time.Since(started))
			return nil, ctx.Err()
		case <-notify:
		}
	}
}

func subtitleCueRange(cues []subtitleCue) (time.Duration, time.Duration) {
	if len(cues) == 0 {
		return 0, 0
	}
	start := cues[0].Start
	end := cues[0].End
	for _, cue := range cues[1:] {
		if cue.Start < start {
			start = cue.Start
		}
		if cue.End > end {
			end = cue.End
		}
	}
	return start, end
}

func subtitleSegmentBounds(segmentIndex int, segmentSeconds int) (time.Duration, time.Duration) {
	if segmentSeconds <= 0 {
		segmentSeconds = 6
	}
	start := time.Duration(segmentIndex*segmentSeconds) * time.Second
	return start, start + time.Duration(segmentSeconds)*time.Second
}

func subtitleTracksForPlaylist(probe ProbeInfo) []TrackInfo {
	var tracks []TrackInfo
	for _, track := range probe.SubtitleTracks() {
		if track.IsSupportedSubtitle() {
			tracks = append(tracks, track)
		}
	}
	return tracks
}

func subtitleTrackName(track TrackInfo) string {
	name := strings.TrimSpace(track.Title)
	if name == "" {
		name = fmt.Sprintf("Subtitle %d", track.Index)
	}

	if language := subtitleTrackDisplayLanguage(track); language != "" {
		name += " [" + language + "]"
	}
	return name
}

func subtitleTrackDisplayLanguage(track TrackInfo) string {
	language := strings.TrimSpace(track.Language)
	if language == "" || strings.EqualFold(language, "und") {
		return ""
	}
	upper := strings.ToUpper(language)
	if len([]rune(upper)) <= 8 {
		return upper
	}
	return language
}

func subtitleTrackLanguage(track TrackInfo) string {
	language := strings.TrimSpace(track.Language)
	if language == "" {
		return "und"
	}
	return language
}

func hlsQuote(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}

func buildSubtitlePlaylist(segmentSeconds int, count int, subtitleIndex int) string {
	if segmentSeconds <= 0 {
		segmentSeconds = 6
	}
	if count < 0 {
		count = 0
	}

	var playlist strings.Builder
	playlist.WriteString("#EXTM3U\n")
	playlist.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")
	playlist.WriteString("#EXT-X-VERSION:3\n")
	playlist.WriteString("#EXT-X-TARGETDURATION:")
	playlist.WriteString(strconv.Itoa(segmentSeconds))
	playlist.WriteByte('\n')
	playlist.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")

	for i := 0; i < count; i++ {
		playlist.WriteString("#EXTINF:")
		playlist.WriteString(strconv.Itoa(segmentSeconds))
		playlist.WriteString(".000,\n")
		playlist.WriteString(strconv.Itoa(subtitleIndex))
		playlist.WriteByte('/')
		playlist.WriteString(strconv.Itoa(i))
		playlist.WriteString(".vtt\n")
	}

	playlist.WriteString("#EXT-X-ENDLIST\n")
	return playlist.String()
}

func emptyWebVTTSegment() []byte {
	return []byte("WEBVTT\n\n")
}

func buildWebVTTSegment(cues []subtitleCue, segmentStart time.Duration, segmentEnd time.Duration) []byte {
	sort.SliceStable(cues, func(i, j int) bool {
		if cues[i].Start == cues[j].Start {
			if cues[i].End == cues[j].End {
				return cues[i].Text < cues[j].Text
			}
			return cues[i].End < cues[j].End
		}
		return cues[i].Start < cues[j].Start
	})

	var out strings.Builder
	out.WriteString("WEBVTT\n\n")

	seen := make(map[string]bool, len(cues))
	for _, cue := range cues {
		cue.Text = strings.TrimSpace(cue.Text)
		if cue.Text == "" || cue.End <= segmentStart || cue.Start >= segmentEnd {
			continue
		}
		key := fmt.Sprintf("%d:%d:%s", cue.Start, cue.End, cue.Text)
		if seen[key] {
			continue
		}
		seen[key] = true

		out.WriteString(formatVTTTime(cue.Start))
		out.WriteString(" --> ")
		out.WriteString(formatVTTTime(cue.End))
		out.WriteByte('\n')
		out.WriteString(cue.Text)
		out.WriteString("\n\n")
	}

	return []byte(out.String())
}

func subtitleCuesFromSample(track TrackInfo, data []byte, start time.Duration, duration time.Duration) []subtitleCue {
	text := normalizeSubtitleSampleText(track, data)
	if text == "" {
		return nil
	}

	if duration <= 0 {
		duration = 4 * time.Second
	}
	end := start + duration
	if end <= start {
		end = start + 4*time.Second
	}

	return []subtitleCue{{
		Start: start,
		End:   end,
		Text:  text,
	}}
}

func normalizeSubtitleSampleText(track TrackInfo, data []byte) string {
	text := string(data)
	text = strings.TrimPrefix(text, "\ufeff")
	text = strings.ReplaceAll(text, "\x00", "")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.TrimSpace(text)

	if strings.EqualFold(text, "WEBVTT") {
		return ""
	}
	if strings.HasPrefix(strings.ToUpper(text), "WEBVTT") {
		lines := strings.Split(text, "\n")
		for len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
			lines = lines[1:]
		}
		text = strings.TrimSpace(strings.Join(lines, "\n"))
	}

	codec := strings.ToLower(track.Codec + " " + track.CapsName)
	if strings.Contains(codec, "ass") || strings.Contains(codec, "ssa") {
		text = cleanASSSubtitleText(text)
	}

	text = strings.ReplaceAll(text, `\N`, "\n")
	text = strings.ReplaceAll(text, `\n`, "\n")
	text = assOverrideTagRe.ReplaceAllString(text, "")
	text = srtTimingLineRe.ReplaceAllString(text, "")
	text = srtIndexLineRe.ReplaceAllString(text, "")
	return strings.TrimSpace(compactSubtitleText(text))
}

func cleanASSSubtitleText(text string) string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(trimmed), "dialogue:") {
			trimmed = assDialogueText(trimmed)
		}
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return strings.Join(lines, "\n")
}

func assDialogueText(line string) string {
	_, rest, ok := strings.Cut(line, ":")
	if !ok {
		return line
	}
	parts := strings.SplitN(strings.TrimSpace(rest), ",", 10)
	if len(parts) < 10 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(parts[9])
}

func compactSubtitleText(text string) string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func formatVTTTime(value time.Duration) string {
	if value < 0 {
		value = 0
	}
	totalMilliseconds := value.Milliseconds()
	hours := totalMilliseconds / 3_600_000
	totalMilliseconds %= 3_600_000
	minutes := totalMilliseconds / 60_000
	totalMilliseconds %= 60_000
	seconds := totalMilliseconds / 1_000
	milliseconds := totalMilliseconds % 1_000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}
