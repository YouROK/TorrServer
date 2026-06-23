package gstreamer

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type pipelineRunner interface {
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

func (t *Task) InitMP4() []byte {
	t.initMu.RLock()
	defer t.initMu.RUnlock()
	return cloneBytes(t.initMP4)
}

func (t *Task) setInitMP4(data []byte) {
	t.initMu.Lock()
	t.initMP4 = cloneBytes(data)
	t.initMu.Unlock()
}

func (t *Task) EnsureInit(ctx context.Context, audio int) error {
	if len(t.InitMP4()) > 0 {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.InitMP4()) > 0 {
		return nil
	}
	if t.runner == nil {
		return ErrTaskNotFound
	}

	_, err := t.runner.GetSegment(ctx, -1, audio)
	return err
}

func (t *Task) Segment(ctx context.Context, index int, audio int) (Segment, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.runner == nil {
		return Segment{}, ErrTaskNotFound
	}

	if t.LastSentSegment != -1 && t.LastSentSegment != index {
		if index != t.LastSentSegment+1 {
			diff := index - t.LastSentSegment
			cutoff := t.Config.PipelineVideoQueue

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
	if t.runner != nil {
		t.runner.Frozen()
	}
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

func (t *Task) IsFrozen() bool {
	return t.runner != nil && t.runner.IsFrozen()
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
