//go:build gst

package gstreamer

import (
	"context"
	"errors"
	"fmt"

	"server/log"
	"server/settings"
)

func gstDebugf(format string, args ...any) {
	if !settings.IsDebug() {
		return
	}
	log.TLogln("[GStreamer] debug:", fmt.Sprintf(format, args...))
}

func gstErrorf(format string, args ...any) {
	log.TLogln("[GStreamer] error:", fmt.Sprintf(format, args...))
}

func gstTaskDebugf(task *Task, format string, args ...any) {
	if !settings.IsDebug() {
		return
	}
	gstDebugf("%s %s", gstTaskLogPrefix(task), fmt.Sprintf(format, args...))
}

func gstTaskErrorf(task *Task, format string, args ...any) {
	gstErrorf("%s %s", gstTaskLogPrefix(task), fmt.Sprintf(format, args...))
}

func gstTaskFailure(task *Task, operation string, err error) {
	if err == nil {
		return
	}
	if errors.Is(err, context.Canceled) {
		gstTaskDebugf(task, "%s canceled", operation)
		return
	}
	gstTaskErrorf(task, "%s failed: %v", operation, err)
}

func gstSourceFailure(hash string, fileID string, audio int, operation string, err error) {
	if err == nil {
		return
	}
	if errors.Is(err, context.Canceled) {
		gstDebugf("hash=%s file=%s audio=%d %s canceled", hash, fileID, audio, operation)
		return
	}
	gstErrorf("hash=%s file=%s audio=%d %s failed: %v", hash, fileID, audio, operation, err)
}

func gstTaskLogPrefix(task *Task) string {
	if task == nil {
		return "hash=<nil>"
	}
	return fmt.Sprintf("hash=%s file=%s audio=%d", task.ID, task.FileID, task.Audio)
}
