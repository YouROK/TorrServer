//go:build !gst || (!(windows && (amd64 || arm64)) && !(linux && (amd64 || arm64)) && !(darwin && (amd64 || arm64)))

package gstreamer

import "context"

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

func (disabledRunner) Seek(float64) bool { return false }
func (disabledRunner) Frozen()           {}
func (disabledRunner) Dispose()          {}
func (disabledRunner) IsFrozen() bool    { return false }
