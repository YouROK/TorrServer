package gstreamer

import (
	"os"
	"path/filepath"
	"strings"
)

const queue2TempPrefix = "gst-"

func queue2TempTemplate() string {
	return filepath.Join(queue2TempDir(), queue2TempPrefix+"XXXXXX")
}

func queue2TempDir() string {
	dir := os.TempDir()
	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exe = resolved
		}
		if exeDir := filepath.Dir(exe); exeDir != "." && exeDir != "" {
			dir = exeDir
		}
	}
	return dir
}

func gstPath(path string) string {
	return strings.ReplaceAll(filepath.ToSlash(path), `"`, `\"`)
}

func cleanupGSTTempFiles() {
	dir := queue2TempDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !isGSTTempFileName(entry.Name()) {
			continue
		}
		_ = os.Remove(filepath.Join(dir, entry.Name()))
	}
}

func isGSTTempFileName(name string) bool {
	return len(name) == len(queue2TempPrefix)+6 && strings.HasPrefix(name, queue2TempPrefix)
}
