//go:build !(windows && (amd64 || arm64)) && !(linux && (amd64 || arm64))

package gstreamer

import "context"

type disabledRunner struct{}

func newPipelineRunner(_ *Task, _ int) (pipelineRunner, error) {
	return nil, ErrPipelineDisabled
}

func (disabledRunner) GetSegment(context.Context, int, int) (Segment, error) {
	return Segment{}, ErrPipelineDisabled
}

func (disabledRunner) Seek(float64) bool { return false }
func (disabledRunner) Frozen()           {}
func (disabledRunner) Dispose()          {}
func (disabledRunner) IsDead() bool      { return false }
func (disabledRunner) IsFrozen() bool    { return false }
