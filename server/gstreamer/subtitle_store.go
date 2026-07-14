package gstreamer

import (
	"strings"
	"time"
)

func (t *Task) subtitleNotifyLocked() chan struct{} {
	if t.subtitleNotify == nil {
		t.subtitleNotify = make(chan struct{})
	}
	return t.subtitleNotify
}

func (t *Task) notifySubtitleWaitersLocked() {
	notify := t.subtitleNotifyLocked()
	close(notify)
	t.subtitleNotify = make(chan struct{})
}

func (t *Task) addSubtitleCues(subtitleIndex int, cues []subtitleCue) {
	if len(cues) == 0 {
		return
	}

	t.subtitleMu.Lock()
	if t.subtitleTracks == nil {
		t.subtitleTracks = make(map[int]*subtitleTrackState)
	}
	state := t.subtitleTracks[subtitleIndex]
	if state == nil {
		state = &subtitleTrackState{seen: make(map[subtitleCueKey]struct{})}
		t.subtitleTracks[subtitleIndex] = state
	}

	added := 0
	for _, cue := range cues {
		cue.Text = strings.TrimSpace(cue.Text)
		if cue.Text == "" || cue.End <= cue.Start {
			continue
		}
		key := subtitleCueKey{Start: cue.Start, End: cue.End, Text: cue.Text}
		if _, exists := state.seen[key]; exists {
			continue
		}
		state.seen[key] = struct{}{}
		state.cues = append(state.cues, cue)
		added++
	}
	total := len(state.cues)
	if added > 0 {
		t.notifySubtitleWaitersLocked()
	}
	t.subtitleMu.Unlock()

	if added > 0 {
		gstDebugf("subtitle store cues task=%s file=%s subtitle=%d added=%d total=%d", t.ID, t.FileID, subtitleIndex, added, total)
	}
}

func (t *Task) subtitleCuesForSegmentLocked(subtitleIndex int, start time.Duration, end time.Duration) []subtitleCue {
	if t.subtitleTracks == nil {
		return nil
	}
	state := t.subtitleTracks[subtitleIndex]
	if state == nil {
		return nil
	}

	cues := make([]subtitleCue, 0)
	for _, cue := range state.cues {
		if cue.End > start && cue.Start < end {
			cues = append(cues, cue)
		}
	}
	return cues
}

func (t *Task) resetSubtitleTimeline(requestedSeconds float64) uint64 {
	generationStart := 0
	segmentSeconds := t.Config.normalized().SegmentSeconds
	if requestedSeconds > 0 && segmentSeconds > 0 {
		generationStart = int(requestedSeconds / float64(segmentSeconds))
	}

	t.subtitleMu.Lock()
	previousGeneration := t.subtitleGeneration
	retainedSegments := len(t.subtitleSegments)
	t.subtitleGeneration++
	generation := t.subtitleGeneration
	if t.subtitleSegments == nil {
		t.subtitleSegments = make(map[int]subtitleSegmentTimeline)
	}
	t.subtitleGenerationStart = generationStart
	t.subtitleFirstReady = -1
	t.subtitleHighestReady = -1
	t.notifySubtitleWaitersLocked()
	t.subtitleMu.Unlock()

	gstDebugf("subtitle timeline reset task=%s file=%s previousGeneration=%d generation=%d requested=%.3f generationStart=%d retainedSegments=%d", t.ID, t.FileID, previousGeneration, generation, requestedSeconds, generationStart, retainedSegments)
	return generation
}

func (t *Task) subtitleSegmentUnavailableLocked(segmentIndex int) (bool, string) {
	if t.subtitleFirstReady >= 0 {
		if segmentIndex < t.subtitleFirstReady {
			return true, "before-first-ready"
		}
		if segmentIndex <= t.subtitleHighestReady {
			return true, "missing-before-highest-ready"
		}
	}
	if t.subtitleGenerationStart >= 0 && segmentIndex < t.subtitleGenerationStart {
		return true, "before-generation-start"
	}
	return false, ""
}

func (t *Task) setSubtitleTimelineOrigin(generation uint64, requestedSeconds float64, actualSeconds float64) {
	t.subtitleMu.Lock()
	currentGeneration := t.subtitleGeneration
	t.subtitleMu.Unlock()
	if generation != currentGeneration {
		gstDebugf("subtitle timeline origin ignored task=%s file=%s generation=%d currentGeneration=%d requested=%.3f actual=%.3f delta=%.3f", t.ID, t.FileID, generation, currentGeneration, requestedSeconds, actualSeconds, actualSeconds-requestedSeconds)
		return
	}
	gstDebugf("subtitle timeline origin task=%s file=%s generation=%d requested=%.3f actual=%.3f delta=%.3f", t.ID, t.FileID, generation, requestedSeconds, actualSeconds, actualSeconds-requestedSeconds)
}

func (t *Task) markSubtitleSegmentReady(segmentIndex int, seg Segment, absoluteEndSeconds float64) {
	if segmentIndex < 0 {
		return
	}
	localDurationSeconds := seg.EndSeconds - seg.StartSeconds
	if localDurationSeconds <= 0 {
		gstErrorf("subtitle timeline segment rejected task=%s file=%s segment=%d localStart=%.3f localEnd=%.3f absoluteEnd=%.3f reason=invalid-duration", t.ID, t.FileID, segmentIndex, seg.StartSeconds, seg.EndSeconds, absoluteEndSeconds)
		return
	}
	if absoluteEndSeconds <= 0 {
		absoluteEndSeconds = seg.EndSeconds
	}
	sourceStartSeconds := absoluteEndSeconds - localDurationSeconds
	sourceEndSeconds := absoluteEndSeconds
	playlistStart, playlistEnd := subtitleSegmentBounds(segmentIndex, t.Config.normalized().SegmentSeconds)
	timeline := subtitleSegmentTimeline{
		SourceStart:   time.Duration(sourceStartSeconds * float64(time.Second)),
		SourceEnd:     time.Duration(sourceEndSeconds * float64(time.Second)),
		PlaylistStart: playlistStart,
		PlaylistEnd:   playlistEnd,
	}

	t.subtitleMu.Lock()
	if t.subtitleSegments == nil {
		t.subtitleSegments = make(map[int]subtitleSegmentTimeline)
	}
	timeline.Generation = t.subtitleGeneration
	t.subtitleSegments[segmentIndex] = timeline
	if t.subtitleFirstReady < 0 {
		t.subtitleFirstReady = segmentIndex
	}
	if t.subtitleHighestReady < segmentIndex {
		t.subtitleHighestReady = segmentIndex
	}
	t.notifySubtitleWaitersLocked()
	t.subtitleMu.Unlock()

	sourceDuration := timeline.SourceEnd - timeline.SourceStart
	playlistDuration := timeline.PlaylistEnd - timeline.PlaylistStart
	gstDebugf("subtitle timeline segment ready task=%s file=%s segment=%d generation=%d sourceStart=%.3f sourceEnd=%.3f sourceDuration=%.3f playlistStart=%.3f playlistEnd=%.3f playlistDuration=%.3f startDelta=%.3f endDelta=%.3f mode=source", t.ID, t.FileID, segmentIndex, timeline.Generation, timeline.SourceStart.Seconds(), timeline.SourceEnd.Seconds(), sourceDuration.Seconds(), timeline.PlaylistStart.Seconds(), timeline.PlaylistEnd.Seconds(), playlistDuration.Seconds(), timeline.SourceStart.Seconds()-timeline.PlaylistStart.Seconds(), timeline.SourceEnd.Seconds()-timeline.PlaylistEnd.Seconds())
}

func (t *Task) closeSubtitleStore() {
	t.subtitleMu.Lock()
	t.subtitleTracks = nil
	t.subtitleSegments = nil
	t.subtitleGeneration++
	t.subtitleGenerationStart = -1
	t.subtitleFirstReady = -1
	t.subtitleHighestReady = -1
	t.notifySubtitleWaitersLocked()
	t.subtitleMu.Unlock()
}
