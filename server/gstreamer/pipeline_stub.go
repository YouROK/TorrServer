//go:build !gst || (!(windows && (amd64 || arm64)) && !(linux && (amd64 || arm64)) && !(darwin && (amd64 || arm64)))

package gstreamer

import (
	"context"
	"time"
)

type disabledRunner struct{}

func newPipelineRunner(_ *Task, _ int) (pipelineRunner, error) {
	return nil, ErrPipelineDisabled
}

func (disabledRunner) EnsureInit(context.Context, int, int) error {
	return ErrPipelineDisabled
}

func (disabledRunner) GetSegment(context.Context, int, int) (Segment, error) {
	return Segment{}, ErrPipelineDisabled
}

func (r disabledRunner) GetSegmentWithTimeout(ctx context.Context, index int, audio int, timeout time.Duration) (Segment, error) {
	return r.GetSegment(ctx, index, audio)
}

func (disabledRunner) Seek(float64) error { return ErrPipelineDisabled }
func (disabledRunner) Frozen()            {}
func (disabledRunner) Dispose()           {}
func (disabledRunner) IsFrozen() bool     { return false }
func (disabledRunner) Position() float64  { return 0 }
