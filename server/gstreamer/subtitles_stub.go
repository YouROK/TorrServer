//go:build !gst || (!(windows && (amd64 || arm64)) && !(linux && (amd64 || arm64)) && !(darwin && (amd64 || arm64)))

package gstreamer

import (
	"context"
	"time"
)

func extractSubtitleCues(_ context.Context, _ *Task, _ TrackInfo, _ time.Duration, _ time.Duration) ([]subtitleCue, error) {
	return nil, ErrPipelineDisabled
}
