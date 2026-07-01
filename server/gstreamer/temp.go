package gstreamer

import (
	"os"
	"path/filepath"
	"strings"

	"server/settings"
)

const queue2TempPrefix = "gst-"

func queue2TempTemplate() string {
	return filepath.Join(queue2TempDir(), queue2TempPrefix+"XXXXXX")
}

func queue2TempDir() string {
	if dir := torrServerWorkDir(); dir != "" {
		return dir
	}

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

func torrServerWorkDir() string {
	for _, dir := range []string{settings.Path, argsPath()} {
		if dir == "" {
			continue
		}
		if abs, err := filepath.Abs(dir); err == nil {
			dir = abs
		}
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return ""
}

func argsPath() string {
	if settings.Args == nil {
		return ""
	}
	return settings.Args.Path
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
