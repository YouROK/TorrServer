//go:build gst

package gstreamer

import (
	"context"
	"errors"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type pipelineRunner interface {
	EnsureInit(ctx context.Context, audio int, startIndex int) error
	GetSegment(ctx context.Context, index int, audio int) (Segment, error)
	Seek(seconds float64) bool
	Frozen()
	Dispose()
	IsFrozen() bool
}

type Task struct {
	ID        string
	FileID    string
	Audio     int
	SourceURL string
	Probe     ProbeInfo
	Cue       *CueTimeline
	Config    Config

	LastSentSegment int

	initMu  sync.RWMutex
	initMP4 []byte
	variant *HLSVariantInfo

	activeMu   sync.RWMutex
	lastActive time.Time

	mu     sync.Mutex
	runner pipelineRunner

	subtitleMu     sync.RWMutex
	subtitleStores map[int]*subtitleStore

	disposed atomic.Bool
}

func NewTask(id string, fileID string, audio int, sourceURL string, probe ProbeInfo, cue *CueTimeline, conf Config) (*Task, error) {
	task := &Task{
		ID:              id,
		FileID:          fileID,
		Audio:           audio,
		SourceURL:       sourceURL,
		Probe:           probe,
		Cue:             cue,
		Config:          conf.normalized(),
		LastSentSegment: -1,
		lastActive:      time.Now().UTC(),
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

func (t *Task) segmentStartNS(index int) uint64 {
	if cue, ok := t.Cue.Segment(index); ok {
		return cue.StartNS
	}
	if index <= 0 {
		return 0
	}
	return uint64(index) * uint64(max(t.Config.SegmentSeconds, 1)) * 1_000_000_000
}

func (t *Task) startIndexForSeconds(seconds int) int {
	if seconds <= 0 {
		return 0
	}
	segmentSeconds := max(t.Config.SegmentSeconds, 1)
	if t.Cue != nil {
		targetNS := uint64(seconds) * 1_000_000_000
		for i, segment := range t.Cue.Segments {
			if targetNS < segment.EndNS {
				return i
			}
		}
		return len(t.Cue.Segments)
	}
	duration := t.Probe.DurationSeconds()
	count := 0
	if duration > 0 {
		count = 1 + (duration-1)/segmentSeconds
	}
	return startSegmentIndex(seconds, segmentSeconds, count)
}

func (t *Task) setInitMP4(data []byte) {
	variant := readMP4InitInfo(data)
	if variant != nil {
		if video := t.Probe.Video(); video != nil {
			if variant.Width <= 0 {
				variant.Width = video.Width
			}
			if variant.Height <= 0 {
				variant.Height = video.Height
			}
			if video.FrameRateNum > 0 && video.FrameRateDen > 0 {
				variant.FrameRate = float64(video.FrameRateNum) / float64(video.FrameRateDen)
			}
			if variant.VideoRange == "" {
				variant.VideoRange = strings.ToUpper(video.VideoTransfer)
			}
			if t.Config.HDRToSDR && video.IsHDRVideo() {
				variant.VideoRange = "SDR"
			}
		}
	}

	t.initMu.Lock()
	t.initMP4 = data
	t.variant = variant
	t.initMu.Unlock()
}

func (t *Task) clearInitMP4() {
	t.initMu.Lock()
	t.initMP4 = nil
	t.variant = nil
	t.initMu.Unlock()
}

func (t *Task) hlsVariant() (HLSVariantInfo, bool) {
	t.initMu.RLock()
	defer t.initMu.RUnlock()
	if t.variant == nil {
		return HLSVariantInfo{}, false
	}
	return *t.variant, true
}

func (t *Task) EnsureInit(ctx context.Context, audio int, startIndex int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if startIndex < 0 {
		startIndex = 0
	}
	if err := t.validateSegmentIndex(startIndex); err != nil {
		return err
	}
	if t.hasInitMP4() && (startIndex == 0 || t.LastSentSegment != -1) {
		return nil
	}
	if t.runner == nil {
		return ErrTaskNotFound
	}

	err := t.runner.EnsureInit(ctx, audio, startIndex)
	if err == nil && startIndex > 0 && t.LastSentSegment == -1 {
		t.LastSentSegment = startIndex - 1
	}
	return err
}

func (t *Task) WithSegment(ctx context.Context, index int, audio int, consume func(Segment) error) error {
	if consume == nil {
		return errors.New("nil segment consumer")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	seg, err := t.segmentLocked(ctx, index, audio)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return consume(seg)
}

func (t *Task) segmentLocked(ctx context.Context, index int, audio int) (Segment, error) {
	if t.runner == nil {
		return Segment{}, ErrTaskNotFound
	}
	if err := t.validateSegmentIndex(index); err != nil {
		return Segment{}, err
	}

	if t.runner.IsFrozen() {
		if err := t.seekToSegmentLocked(index); err != nil {
			return Segment{}, err
		}
	} else if t.LastSentSegment == -1 && index > 0 {
		if err := t.seekToSegmentLocked(index); err != nil {
			return Segment{}, err
		}
	} else if t.LastSentSegment != -1 && t.LastSentSegment != index {
		if index != t.LastSentSegment+1 {
			conf := t.Config.normalized()
			seekRequired := true
			if t.Cue == nil {
				diff := index - t.LastSentSegment

				if diff > 0 && diff <= maxSegmentCatchupSeconds/conf.SegmentSeconds {
					for i := 0; i < diff-1; i++ {
						if ctx.Err() != nil {
							return Segment{}, ctx.Err()
						}

						t.LastSentSegment++
						if _, err := t.runner.GetSegment(ctx, t.LastSentSegment, audio); err != nil {
							t.LastSentSegment--
							return Segment{}, err
						}
					}
					seekRequired = false
				}
			}
			if seekRequired {
				if err := t.seekToSegmentLocked(index); err != nil {
					return Segment{}, err
				}
			}
		}
	}

	seg, err := t.runner.GetSegment(ctx, index, audio)
	if err != nil {
		return Segment{}, err
	}

	t.LastSentSegment = index
	return seg, nil
}

func (t *Task) seekToSegmentLocked(index int) error {
	if cue, ok := t.Cue.Segment(index); ok {
		if !t.runner.Seek(float64(cue.StartNS) / 1_000_000_000) {
			return ErrSegmentNotReady
		}
		return nil
	}
	if t.Cue != nil {
		return ErrEndOfStreamExhausted
	}
	seconds := float64(index) * float64(t.Config.normalized().SegmentSeconds)
	if !t.runner.Seek(seconds) {
		return ErrSegmentNotReady
	}
	return nil
}

func (t *Task) validateSegmentIndex(index int) error {
	if index < 0 {
		return ErrInvalidIdentifier
	}
	if t.Cue != nil {
		if index >= len(t.Cue.Segments) {
			return ErrEndOfStreamExhausted
		}
		return nil
	}

	segmentSeconds := int64(t.Config.normalized().SegmentSeconds)
	if segmentSeconds <= 0 || segmentSeconds > math.MaxInt64/int64(time.Second) {
		return ErrInvalidIdentifier
	}
	segmentDurationNS := segmentSeconds * int64(time.Second)
	if int64(index) > math.MaxInt64/segmentDurationNS {
		return ErrInvalidIdentifier
	}
	if t.Probe.DurationNS > 0 {
		count := 1 + (t.Probe.DurationNS-1)/segmentDurationNS
		if int64(index) >= count {
			return ErrEndOfStreamExhausted
		}
	}
	return nil
}

const maxSegmentCatchupSeconds = 60

func (t *Task) Frozen() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.freezeLocked()
}

func (t *Task) FreezeIfInactive(cutoff time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.activeMu.Lock()
	defer t.activeMu.Unlock()
	if !t.lastActive.Before(cutoff) {
		return false
	}

	return t.freezeLocked()
}

func (t *Task) freezeLocked() bool {
	if t.disposed.Load() || t.runner == nil || t.runner.IsFrozen() {
		return false
	}

	t.runner.Frozen()
	t.clearInitMP4()
	return true
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
	t.clearInitMP4()
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
