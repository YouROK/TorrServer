//go:build linux && (amd64 || arm64)

package gstreamer

import (
	"os"
	"syscall"
)

const gstReexecEnv = "TORRSERVER_GST_REEXEC"

func ensureGStreamerRuntimeEnv(conf Config) {
	if os.Getenv(gstReexecEnv) == "1" {
		return
	}

	roots := gstRuntimeRoots(conf)
	if len(roots) == 0 || firstExistingPath(gstLibraryDirCandidates(roots)) == "" {
		return
	}

	setupGStreamer(conf)
	_ = os.Setenv(gstReexecEnv, "1")

	exe, err := os.Executable()
	if err != nil || exe == "" {
		return
	}
	_ = syscall.Exec(exe, os.Args, os.Environ())
}
