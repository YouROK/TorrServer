package log

import (
	"path/filepath"
	"testing"
)

func TestInitSharedLogPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "combined.log")

	Init(path, path)
	defer Close()

	if logFile == nil {
		t.Fatal("expected logFile to be set")
	}
	if webLogFile != logFile {
		t.Fatal("expected webLogFile to share logFile when paths are equal")
	}
	if webLog == nil {
		t.Fatal("expected webLog to be set")
	}
}

func TestInitSeparateLogPaths(t *testing.T) {
	dir := t.TempDir()
	serverPath := filepath.Join(dir, "server.log")
	webPath := filepath.Join(dir, "web.log")

	Init(serverPath, webPath)
	defer Close()

	if logFile == nil || webLogFile == nil {
		t.Fatal("expected both log files to be set")
	}
	if logFile == webLogFile {
		t.Fatal("expected separate file handles for different paths")
	}
}

func TestCloseSharedLogPathOnce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "combined.log")

	Init(path, path)
	Close()

	if logFile != nil || webLogFile != nil || webLog != nil {
		t.Fatal("expected log handles to be cleared after Close")
	}
}
