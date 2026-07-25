//go:build gst

package gstreamer

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type trackingRunner struct {
	disposed atomic.Bool
	frozen   atomic.Bool
}

func (r *trackingRunner) EnsureInit(context.Context, int, int) error {
	return nil
}

func (r *trackingRunner) GetSegment(context.Context, int, int) (Segment, error) {
	return Segment{}, nil
}

func (r *trackingRunner) Seek(float64) bool {
	return true
}

func (r *trackingRunner) Frozen() {
	r.frozen.Store(true)
}

func (r *trackingRunner) Dispose() {
	r.disposed.Store(true)
}

func (r *trackingRunner) IsFrozen() bool {
	return r.frozen.Load()
}

func newTrackedTask(id string, lastActive time.Time) (*Task, *trackingRunner) {
	runner := &trackingRunner{}
	return &Task{
		ID:              id,
		FileID:          "1",
		LastSentSegment: -1,
		lastActive:      lastActive,
		runner:          runner,
	}, runner
}

func TestTaskLimitOneEvictsExistingTaskForProtectedNewTask(t *testing.T) {
	now := time.Now().UTC()
	oldTask, oldRunner := newTrackedTask("old", now.Add(-time.Minute))
	newTask, newRunner := newTrackedTask("new", now)

	service := &Service{
		conf: Config{MaxTasks: 1}.normalized(),
		tasks: map[string]*Task{
			oldTask.ID: oldTask,
			newTask.ID: newTask,
		},
	}

	service.mu.Lock()
	evicted := service.evictTasksForLimitLocked(newTask.ID)
	service.mu.Unlock()
	disposeTasks(evicted)

	if _, ok := service.tasks[oldTask.ID]; ok {
		t.Fatal("old task was not evicted")
	}
	if service.tasks[newTask.ID] != newTask {
		t.Fatal("protected new task was evicted")
	}
	if !oldTask.IsDisposed() || !oldRunner.disposed.Load() {
		t.Fatal("old task was not disposed")
	}
	if newTask.IsDisposed() || newRunner.disposed.Load() {
		t.Fatal("protected new task was disposed")
	}
}

func TestTaskLimitEvictsOldestTask(t *testing.T) {
	now := time.Now().UTC()
	oldTask, oldRunner := newTrackedTask("old", now.Add(-3*time.Minute))
	midTask, midRunner := newTrackedTask("mid", now.Add(-time.Minute))
	newTask, newRunner := newTrackedTask("new", now)

	service := &Service{
		conf: Config{MaxTasks: 2}.normalized(),
		tasks: map[string]*Task{
			oldTask.ID: oldTask,
			midTask.ID: midTask,
			newTask.ID: newTask,
		},
	}

	service.mu.Lock()
	evicted := service.evictTasksForLimitLocked(newTask.ID)
	service.mu.Unlock()
	disposeTasks(evicted)

	if _, ok := service.tasks[oldTask.ID]; ok {
		t.Fatal("oldest task was not evicted")
	}
	if service.tasks[midTask.ID] != midTask {
		t.Fatal("newer task was evicted instead of the oldest one")
	}
	if service.tasks[newTask.ID] != newTask {
		t.Fatal("protected new task was evicted")
	}
	if !oldTask.IsDisposed() || !oldRunner.disposed.Load() {
		t.Fatal("oldest task was not disposed")
	}
	if midTask.IsDisposed() || midRunner.disposed.Load() {
		t.Fatal("newer task was disposed")
	}
	if newTask.IsDisposed() || newRunner.disposed.Load() {
		t.Fatal("protected new task was disposed")
	}
}

func TestRemoveInactiveTaskRechecksLastActive(t *testing.T) {
	now := time.Now().UTC()
	task, runner := newTrackedTask("task", now)
	service := &Service{
		tasks: map[string]*Task{task.ID: task},
	}

	if service.tryRemoveExpectedInactive(task.ID, task, now.Add(-time.Minute)) {
		t.Fatal("active task was removed")
	}
	if service.tasks[task.ID] != task || task.IsDisposed() || runner.disposed.Load() {
		t.Fatal("active task changed during cleanup recheck")
	}

	task.activeMu.Lock()
	task.lastActive = now.Add(-2 * time.Minute)
	task.activeMu.Unlock()
	if !service.tryRemoveExpectedInactive(task.ID, task, now.Add(-time.Minute)) {
		t.Fatal("inactive task was not removed")
	}
	if !task.IsDisposed() || !runner.disposed.Load() {
		t.Fatal("removed task was not disposed")
	}
}

func TestServiceDisposeRejectsFurtherWork(t *testing.T) {
	task, runner := newTrackedTask("task", time.Now().UTC())
	service := &Service{
		tasks:       map[string]*Task{task.ID: task},
		probeCache:  make(map[string]probeCacheEntry),
		stopCleanup: make(chan struct{}),
	}

	service.Dispose()
	service.Dispose()

	if !service.disposed.Load() {
		t.Fatal("service was not marked disposed")
	}
	if !task.IsDisposed() || !runner.disposed.Load() {
		t.Fatal("service task was not disposed")
	}
	if got := service.Get(task.ID); got != nil {
		t.Fatal("disposed service returned a task")
	}
	if _, err := service.GetOrAdd("hash", "1", 0); !errors.Is(err, ErrServiceClosed) {
		t.Fatalf("GetOrAdd error=%v, want ErrServiceClosed", err)
	}
	if _, err := service.Probe("hash", "1"); !errors.Is(err, ErrServiceClosed) {
		t.Fatalf("Probe error=%v, want ErrServiceClosed", err)
	}
}

func TestCachedProbeIsRevalidatedAfterConfigChange(t *testing.T) {
	service := &Service{
		conf:       Config{TranscodeAVI: true}.normalized(),
		tasks:      make(map[string]*Task),
		probeCache: make(map[string]probeCacheEntry),
	}
	service.setCachedProbe("hash", "1", ProbeInfo{
		Container: "AVI",
		Tracks: []TrackInfo{
			{Type: "video", CapsName: "video/x-h264"},
		},
	})

	service.updateConfig(Config{})
	if _, err := service.Probe("hash", "1"); !errors.Is(err, ErrUnsupportedContainer) {
		t.Fatalf("Probe error=%v, want ErrUnsupportedContainer", err)
	}
}
