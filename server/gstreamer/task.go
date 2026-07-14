package gstreamer

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type pipelineRunner interface {
	EnsureInit(ctx context.Context, audio int, startIndex int) error
	GetSegmentWithTimeout(ctx context.Context, index int, audio int, timeout time.Duration) (Segment, error)
	DiscardSegment()
	Position() float64
	Seek(seconds float64) error
	Frozen()
	Dispose()
	IsFrozen() bool
}

const (
	forwardGapDrainSegmentLimit = 10
	seekPrerollDrainSegmentLimit = 10
)

type Task struct {
	ID        string
	Hash      string
	FileID    string
	Audio     int
	SourceURL string
	Probe     ProbeInfo
	Config    Config

	LastSentSegment   int
	LastSentSegmentAt time.Time

	initMu  sync.RWMutex
	initMP4 []byte

	subtitleMu         sync.Mutex
	subtitleTracks     map[int]*subtitleTrackState
	subtitleSegments   map[int]subtitleSegmentTimeline
	subtitleNotify     chan struct{}
	subtitleGeneration uint64
	subtitleGenerationStart int
	subtitleFirstReady      int
	subtitleHighestReady    int

	pendingSeek        bool
	pendingSeekRequest float64
	pendingSeekAudio   int
	pendingSeekIndex   int

	activeMu   sync.RWMutex
	lastActive time.Time

	mu     sync.Mutex
	runner pipelineRunner

	disposed atomic.Bool
}

func NewTask(id string, hash string, fileID string, audio int, sourceURL string, probe ProbeInfo, conf Config) (*Task, error) {
	task := &Task{
		ID:              id,
		Hash:            hash,
		FileID:          fileID,
		Audio:           audio,
		SourceURL:       sourceURL,
		Probe:           probe,
		Config:          conf.normalized(),
		LastSentSegment: -1,
		lastActive:      time.Now().UTC(),
		subtitleNotify:  make(chan struct{}),
		subtitleGenerationStart: -1,
		subtitleFirstReady:      -1,
		subtitleHighestReady:    -1,
	}

	runner, err := newPipelineRunner(task, audio)
	if err != nil {
		return nil, err
	}
	task.runner = runner
	return task, nil
}

func (t *Task) UpdateLastActive() {
	t.activeMu.Lock()
	t.lastActive = time.Now().UTC()
	t.activeMu.Unlock()
}

func (t *Task) LastActive() time.Time {
	t.activeMu.RLock()
	defer t.activeMu.RUnlock()
	return t.lastActive
}

func (t *Task) WithInitMP4(consume func([]byte) error) error {
	if consume == nil {
		return errors.New("nil init mp4 consumer")
	}

	t.initMu.RLock()
	defer t.initMu.RUnlock()

	if len(t.initMP4) == 0 {
		return ErrSegmentNotReady
	}
	return consume(t.initMP4)
}

func (t *Task) hasInitMP4() bool {
	t.initMu.RLock()
	defer t.initMu.RUnlock()
	return len(t.initMP4) > 0
}

func (t *Task) setInitMP4(data []byte) {
	t.initMu.Lock()
	t.initMP4 = cloneBytes(data)
	t.initMu.Unlock()
}

func (t *Task) EnsureInit(ctx context.Context, audio int, startIndex int) error {
	started := time.Now()
	t.mu.Lock()
	defer t.mu.Unlock()

	if startIndex < 0 {
		startIndex = 0
	}

	hasInit := t.hasInitMP4()
	gstDebugf("EnsureInit start task=%s file=%s audio=%d startIndex=%d hasInit=%t", t.ID, t.FileID, audio, startIndex, hasInit)
	if hasInit && (startIndex == 0 || t.LastSentSegment != -1) {
		if startIndex > 0 && t.LastSentSegment != -1 {
			seconds := float64(startIndex * t.Config.normalized().SegmentSeconds)
			t.beginPendingSeekLocked(startIndex, audio, seconds)
			gstDebugf("EnsureInit pending explicit seek task=%s file=%s audio=%d startIndex=%d requested=%.3f LastSentSegment=%d", t.ID, t.FileID, audio, startIndex, seconds, t.LastSentSegment)
		}
		gstDebugf("EnsureInit early return task=%s file=%s audio=%d startIndex=%d duration=%s", t.ID, t.FileID, audio, startIndex, time.Since(started))
		return nil
	}
	if t.runner == nil {
		gstErrorf("EnsureInit failed task=%s file=%s audio=%d startIndex=%d err=%v duration=%s", t.ID, t.FileID, audio, startIndex, ErrTaskNotFound, time.Since(started))
		return ErrTaskNotFound
	}

	err := t.runner.EnsureInit(ctx, audio, startIndex)
	if err == nil && startIndex > 0 && t.LastSentSegment == -1 {
		t.LastSentSegment = startIndex - 1
	}
	if err != nil {
		gstErrorf("EnsureInit failed task=%s file=%s audio=%d startIndex=%d err=%v duration=%s", t.ID, t.FileID, audio, startIndex, err, time.Since(started))
	} else {
		gstDebugf("EnsureInit completed task=%s file=%s audio=%d startIndex=%d duration=%s", t.ID, t.FileID, audio, startIndex, time.Since(started))
	}
	return err
}

func (t *Task) WithSegment(ctx context.Context, index int, audio int, consume func(Segment) error) error {
	if consume == nil {
		return errors.New("nil segment consumer")
	}

	t.mu.Lock()
	seg, err := t.segmentLocked(ctx, index, audio)
	if err == nil {
		seg = seg.Clone()
	}
	t.mu.Unlock()

	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		gstDebugf("WithSegment context canceled after segment generation task=%s file=%s audio=%d segment=%d err=%v", t.ID, t.FileID, audio, index, err)
		return err
	}
	return consume(seg)
}

func (t *Task) segmentLocked(ctx context.Context, index int, audio int) (Segment, error) {
	started := time.Now()
	conf := t.Config.normalized()
	playlistStart := float64(index * conf.SegmentSeconds)
	playlistEnd := playlistStart + float64(conf.SegmentSeconds)
	runnerPosition := 0.0
	if t.runner != nil {
		runnerPosition = t.runner.Position()
	}
	lastSentAge := 0.0
	if !t.LastSentSegmentAt.IsZero() {
		lastSentAge = started.Sub(t.LastSentSegmentAt).Seconds()
		if lastSentAge < 0 {
			lastSentAge = 0
		}
	}
	gstDebugf("segment request task=%s file=%s audio=%d segment=%d playlistStart=%.3f playlistEnd=%.3f LastSentSegment=%d lastSentAge=%.3f runnerPosition=%.3f pendingSeek=%t pendingIndex=%d pendingRequested=%.3f", t.ID, t.FileID, audio, index, playlistStart, playlistEnd, t.LastSentSegment, lastSentAge, runnerPosition, t.pendingSeek, t.pendingSeekIndex, t.pendingSeekRequest)
	if t.runner == nil {
		gstErrorf("segmentLocked failed task=%s file=%s audio=%d segment=%d err=%v duration=%s", t.ID, t.FileID, audio, index, ErrTaskNotFound, time.Since(started))
		return Segment{}, ErrTaskNotFound
	}

	if t.pendingSeek && t.pendingSeekAudio == audio {
		if err := t.applyPendingSeekLocked(ctx, index, audio, started); err != nil {
			return Segment{}, err
		}
	}

	if t.LastSentSegment != -1 && t.LastSentSegment != index {
		if index != t.LastSentSegment+1 {
			diff := index - t.LastSentSegment
			gstDebugf("segmentLocked non-sequential task=%s file=%s audio=%d segment=%d LastSentSegment=%d diff=%d", t.ID, t.FileID, audio, index, t.LastSentSegment, diff)

			if diff > 1 {
				virtualTargetSeconds := float64(index * conf.SegmentSeconds)
				currentPosition := t.runner.Position()
				tolerance := float64(2 * conf.SegmentSeconds)
				elapsedSinceLastSent := lastSentAge
				targetAhead := virtualTargetSeconds - currentPosition
				continueCurrentPipeline := false
				continueReason := ""
				if currentPosition > 0 && virtualTargetSeconds < currentPosition-tolerance {
					continueCurrentPipeline = true
					continueReason = "target-behind-current"
				} else if currentPosition > 0 && targetAhead > 0 && elapsedSinceLastSent > 0 && elapsedSinceLastSent+tolerance >= targetAhead {
					continueCurrentPipeline = true
					continueReason = "playback-elapsed"
				}
				if continueCurrentPipeline {
					gstDebugf("segmentLocked forward gap continue current pipeline task=%s file=%s audio=%d segment=%d LastSentSegment=%d diff=%d forwardGapLimit=%d virtualTarget=%.3f currentPosition=%.3f targetAhead=%.3f elapsedSinceLastSent=%.3f tolerance=%.3f reason=%s duration=%s", t.ID, t.FileID, audio, index, t.LastSentSegment, diff, forwardGapDrainSegmentLimit, virtualTargetSeconds, currentPosition, targetAhead, elapsedSinceLastSent, tolerance, continueReason, time.Since(started))
				} else {
					gstDebugf("segmentLocked forward gap request fallback to virtual seek task=%s file=%s audio=%d segment=%d LastSentSegment=%d diff=%d forwardGapLimit=%d virtualTarget=%.3f currentPosition=%.3f targetAhead=%.3f elapsedSinceLastSent=%.3f tolerance=%.3f duration=%s", t.ID, t.FileID, audio, index, t.LastSentSegment, diff, forwardGapDrainSegmentLimit, virtualTargetSeconds, currentPosition, targetAhead, elapsedSinceLastSent, tolerance, time.Since(started))
					if err := t.seekVirtualSegmentLocked(index, audio, started); err != nil {
						return Segment{}, err
					}
				}
			} else if diff < 0 {
				gstDebugf("segmentLocked backward request fallback to virtual seek task=%s file=%s audio=%d segment=%d LastSentSegment=%d diff=%d forwardGapLimit=%d runnerPosition=%.3f duration=%s", t.ID, t.FileID, audio, index, t.LastSentSegment, diff, forwardGapDrainSegmentLimit, t.runner.Position(), time.Since(started))
				if err := t.seekVirtualSegmentLocked(index, audio, started); err != nil {
					return Segment{}, err
				}
			} else {
				gstErrorf("segmentLocked non-sequential without seek target task=%s file=%s audio=%d segment=%d LastSentSegment=%d diff=%d forwardGapLimit=%d duration=%s", t.ID, t.FileID, audio, index, t.LastSentSegment, diff, forwardGapDrainSegmentLimit, time.Since(started))
				return Segment{}, ErrSegmentNotReady
			}
		} else {
			gstDebugf("segmentLocked sequential task=%s file=%s audio=%d segment=%d LastSentSegment=%d", t.ID, t.FileID, audio, index, t.LastSentSegment)
		}
	}

	gstDebugf("segment runner request task=%s file=%s audio=%d segment=%d playlistStart=%.3f playlistEnd=%.3f LastSentSegment=%d runnerPosition=%.3f pendingSeek=%t", t.ID, t.FileID, audio, index, playlistStart, playlistEnd, t.LastSentSegment, t.runner.Position(), t.pendingSeek)
	var seg Segment
	for discarded := 0; ; discarded++ {
		var err error
		seg, err = t.runner.GetSegmentWithTimeout(ctx, index, audio, 0)
		if err != nil {
			gstErrorf("segmentLocked failed task=%s file=%s audio=%d segment=%d prerollDiscarded=%d err=%v duration=%s", t.ID, t.FileID, audio, index, discarded, err, time.Since(started))
			return Segment{}, err
		}

		runnerPosition = t.runner.Position()
		source := "runner-result"
		if discarded > 0 {
			source = "preroll-drain"
		}
		t.logSegmentResultLocked(source, index, audio, seg, runnerPosition, started)

		absoluteEnd := runnerPosition
		if absoluteEnd <= 0 {
			absoluteEnd = seg.EndSeconds
		}
		if index <= 0 || absoluteEnd > playlistStart {
			break
		}

		localDuration := seg.EndSeconds - seg.StartSeconds
		absoluteStart := absoluteEnd - localDuration
		if discarded >= seekPrerollDrainSegmentLimit {
			gstErrorf("segmentLocked preroll drain limit reached task=%s file=%s audio=%d segment=%d playlistStart=%.3f absoluteStart=%.3f absoluteEnd=%.3f discarded=%d limit=%d duration=%s", t.ID, t.FileID, audio, index, playlistStart, absoluteStart, absoluteEnd, discarded, seekPrerollDrainSegmentLimit, time.Since(started))
			return Segment{}, ErrSegmentNotReady
		}

		gstDebugf("segmentLocked preroll discard task=%s file=%s audio=%d segment=%d playlistStart=%.3f absoluteStart=%.3f absoluteEnd=%.3f discard=%d limit=%d duration=%s", t.ID, t.FileID, audio, index, playlistStart, absoluteStart, absoluteEnd, discarded+1, seekPrerollDrainSegmentLimit, time.Since(started))
		t.runner.DiscardSegment()
	}
	t.markSubtitleSegmentReady(index, seg, runnerPosition)
	t.markSegmentSentLocked(index)
	gstDebugf("segmentLocked completed task=%s file=%s audio=%d segment=%d size=%d duration=%s", t.ID, t.FileID, audio, index, seg.Len(), time.Since(started))
	return seg, nil
}

func (t *Task) logSegmentResultLocked(source string, index int, audio int, seg Segment, absoluteEndSeconds float64, started time.Time) {
	conf := t.Config.normalized()
	playlistStart := float64(index * conf.SegmentSeconds)
	playlistEnd := playlistStart + float64(conf.SegmentSeconds)
	localDuration := seg.EndSeconds - seg.StartSeconds
	if localDuration < 0 {
		localDuration = 0
	}
	absoluteStart := absoluteEndSeconds - localDuration
	if absoluteEndSeconds <= 0 {
		absoluteStart = seg.StartSeconds
		absoluteEndSeconds = seg.EndSeconds
	}
	absoluteDuration := absoluteEndSeconds - absoluteStart
	startDelta := absoluteStart - playlistStart
	endDelta := absoluteEndSeconds - playlistEnd
	gstDebugf("segment result source=%s task=%s file=%s audio=%d segment=%d playlistStart=%.3f playlistEnd=%.3f localStart=%.3f localEnd=%.3f localDuration=%.3f absoluteStart=%.3f absoluteEnd=%.3f absoluteDuration=%.3f startDelta=%.3f endDelta=%.3f runnerPosition=%.3f LastSentSegment=%d pendingSeek=%t size=%d elapsed=%s", source, t.ID, t.FileID, audio, index, playlistStart, playlistEnd, seg.StartSeconds, seg.EndSeconds, localDuration, absoluteStart, absoluteEndSeconds, absoluteDuration, startDelta, endDelta, absoluteEndSeconds, t.LastSentSegment, t.pendingSeek, seg.Len(), time.Since(started))
}

func (t *Task) markSegmentSentLocked(index int) {
	t.LastSentSegment = index
	t.LastSentSegmentAt = time.Now().UTC()
}

func (t *Task) applyPendingSeekLocked(ctx context.Context, index int, audio int, started time.Time) error {
	seekIndex := t.pendingSeekIndex
	seekGap := index - seekIndex
	if seekGap < 0 || seekGap > forwardGapDrainSegmentLimit {
		gstErrorf("segmentLocked explicit seek segment mismatch task=%s file=%s audio=%d segment=%d pendingSegment=%d gap=%d forwardGapLimit=%d duration=%s", t.ID, t.FileID, audio, index, seekIndex, seekGap, forwardGapDrainSegmentLimit, time.Since(started))
		return ErrSegmentNotReady
	}

	seconds := t.pendingSeekRequest
	gstDebugf("segmentLocked explicit seek selected task=%s file=%s audio=%d segment=%d pendingSegment=%d requested=%.3f", t.ID, t.FileID, audio, index, seekIndex, seconds)
	if err := t.runner.Seek(seconds); err != nil {
		gstErrorf("segmentLocked explicit seek failed task=%s file=%s audio=%d segment=%d pendingSegment=%d requested=%.3f err=%v duration=%s", t.ID, t.FileID, audio, index, seekIndex, seconds, err, time.Since(started))
		return err
	}

	t.pendingSeek = false
	t.LastSentSegment = seekIndex - 1
	gstDebugf("segmentLocked explicit seek position fixed task=%s file=%s audio=%d segment=%d pendingSegment=%d LastSentSegment=%d", t.ID, t.FileID, audio, index, seekIndex, t.LastSentSegment)

	if index != t.LastSentSegment+1 {
		if err := t.drainForwardGapLocked(ctx, index, audio, started); err != nil {
			return err
		}
	}
	return nil
}

func (t *Task) drainForwardGapLocked(ctx context.Context, targetIndex int, audio int, started time.Time) error {
	for nextIndex := t.LastSentSegment + 1; nextIndex < targetIndex; nextIndex++ {
		gstDebugf("segmentLocked forward gap drain start task=%s file=%s audio=%d segment=%d target=%d LastSentSegment=%d", t.ID, t.FileID, audio, nextIndex, targetIndex, t.LastSentSegment)

		seg, err := t.runner.GetSegmentWithTimeout(ctx, nextIndex, audio, 0)
		if err != nil {
			gstErrorf("segmentLocked forward gap drain failed task=%s file=%s audio=%d segment=%d target=%d err=%v duration=%s", t.ID, t.FileID, audio, nextIndex, targetIndex, err, time.Since(started))
			return err
		}

		t.logSegmentResultLocked("forward-gap-drain", nextIndex, audio, seg, t.runner.Position(), started)
		t.markSubtitleSegmentReady(nextIndex, seg, t.runner.Position())
		t.markSegmentSentLocked(nextIndex)
		gstDebugf("segmentLocked forward gap drain completed task=%s file=%s audio=%d segment=%d target=%d size=%d LastSentSegment=%d duration=%s", t.ID, t.FileID, audio, nextIndex, targetIndex, seg.Len(), t.LastSentSegment, time.Since(started))
	}
	return nil
}

func (t *Task) seekVirtualSegmentLocked(index int, audio int, started time.Time) error {
	conf := t.Config.normalized()
	seconds := float64(index * conf.SegmentSeconds)
	currentPosition := t.runner.Position()
	virtualDelta := currentPosition - seconds
	gstDebugf("segmentLocked virtual seek selected task=%s file=%s audio=%d segment=%d requested=%.3f currentPosition=%.3f delta=%.3f LastSentSegment=%d pendingSeek=%t", t.ID, t.FileID, audio, index, seconds, currentPosition, virtualDelta, t.LastSentSegment, t.pendingSeek)
	if err := t.runner.Seek(seconds); err != nil {
		gstErrorf("segmentLocked virtual seek failed task=%s file=%s audio=%d segment=%d requested=%.3f err=%v duration=%s", t.ID, t.FileID, audio, index, seconds, err, time.Since(started))
		return err
	}

	t.pendingSeek = false
	t.LastSentSegment = index - 1
	gstDebugf("segmentLocked virtual seek position fixed task=%s file=%s audio=%d segment=%d LastSentSegment=%d requested=%.3f", t.ID, t.FileID, audio, index, t.LastSentSegment, seconds)
	return nil
}

func (t *Task) beginPendingSeekLocked(index int, audio int, requestedSeek float64) {
	t.pendingSeek = true
	t.pendingSeekRequest = requestedSeek
	t.pendingSeekAudio = audio
	t.pendingSeekIndex = index
}

func (t *Task) Frozen() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.disposed.Load() || t.runner == nil {
		return
	}
	t.runner.Frozen()
}

func (t *Task) Dispose() {
	if !t.disposed.CompareAndSwap(false, true) {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if t.runner != nil {
		t.runner.Dispose()
		t.runner = nil
	}
	t.setInitMP4(nil)
	t.closeSubtitleStore()
}

func (t *Task) IsDisposed() bool {
	return t.disposed.Load()
}

func (t *Task) IsFrozen() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.disposed.Load() || t.runner == nil {
		return false
	}
	return t.runner.IsFrozen()
}
