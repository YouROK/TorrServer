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
	GetSegment(ctx context.Context, index int, audio int) (Segment, error)
	Seek(seconds float64) bool
	Frozen()
	Dispose()
	IsFrozen() bool
}

type Task struct {
	ID        string
	Hash      string
	FileID    string
	Audio     int
	SourceURL string
	Probe     ProbeInfo
	Config    Config

	LastSentSegment int

	initMu  sync.RWMutex
	initMP4 []byte

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
	t.mu.Lock()
	defer t.mu.Unlock()

	if startIndex < 0 {
		startIndex = 0
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
	return consume(seg)
}

func (t *Task) segmentLocked(ctx context.Context, index int, audio int) (Segment, error) {
	if t.runner == nil {
		return Segment{}, ErrTaskNotFound
	}

	if t.LastSentSegment != -1 && t.LastSentSegment != index {
		if index != t.LastSentSegment+1 {
			diff := index - t.LastSentSegment
			cutoff := t.Config.PipelineTimeSeconds

			if diff > 0 && maxInt(60, cutoff) >= diff*t.Config.SegmentSeconds {
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
			} else {
				if !t.runner.Seek(float64(index * t.Config.SegmentSeconds)) {
					return Segment{}, ErrSegmentNotReady
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

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
