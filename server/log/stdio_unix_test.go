//go:build !windows

package log

import (
	stdlog "log"
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

// Init dup2s the process stdout/stderr onto the log file. Put them back so
// later tests and the race detector still have a live stderr.
func withRestoredStdio(t *testing.T) {
	t.Helper()
	stdout, stderr := os.Stdout, os.Stderr
	out, err := unix.Dup(1)
	if err != nil {
		t.Fatal(err)
	}
	errfd, err := unix.Dup(2)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = unix.Dup2(out, 1)
		_ = unix.Dup2(errfd, 2)
		_ = unix.Close(out)
		_ = unix.Close(errfd)
		os.Stdout = stdout
		os.Stderr = stderr
		stdlog.SetOutput(stderr)
	})
}
