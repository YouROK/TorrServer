//go:build gst && linux && (amd64 || arm64)

package gstreamer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ebitengine/purego"
)

func loadGST(conf Config) (*gstAPI, error) {
	roots := gstRuntimeRoots(conf)
	var failures []string
	for _, root := range roots {
		api, err := loadLinuxGSTRoot(root)
		if err == nil {
			return api, nil
		}
		failures = append(failures, fmt.Sprintf("%s: %v", root, err))
	}

	api, err := loadLinuxGST(func(name string) (uintptr, error) {
		return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	})
	if err == nil {
		return api, nil
	}
	failures = append(failures, fmt.Sprintf("system search path: %v", err))
	return nil, fmt.Errorf("load gstreamer runtime: %s", strings.Join(failures, "; "))
}

func loadLinuxGSTRoot(root string) (*gstAPI, error) {
	dirs := gstLibraryDirCandidates([]string{root})
	api, err := loadLinuxGST(func(name string) (uintptr, error) {
		return openLinuxLibrary(dirs, name)
	})
	if err != nil {
		return nil, err
	}
	setupGStreamerRoots([]string{root})
	return api, nil
}

func loadLinuxGST(load func(string) (uintptr, error)) (*gstAPI, error) {
	names := [...]string{"libglib-2.0.so.0", "libgstreamer-1.0.so.0", "libgstapp-1.0.so.0"}
	handles := make([]uintptr, 0, len(names))
	for _, name := range names {
		handle, err := load(name)
		if err != nil {
			closeLinuxLibraries(handles)
			return nil, fmt.Errorf("load %s: %w", name, err)
		}
		handles = append(handles, handle)
	}

	api := &gstAPI{handles: handles}
	if err := api.bind(handles[1], handles[2], handles[0]); err != nil {
		closeLinuxLibraries(handles)
		return nil, err
	}
	return api, nil
}

func openLinuxLibrary(dirs []string, name string) (uintptr, error) {
	var lastErr error
	for _, dir := range dirs {
		handle, err := purego.Dlopen(filepath.Join(dir, name), purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err == nil {
			return handle, nil
		}
		lastErr = err
	}
	return 0, lastErr
}

func closeLinuxLibraries(handles []uintptr) {
	for i := len(handles) - 1; i >= 0; i-- {
		_ = purego.Dlclose(handles[i])
	}
}

func linuxLibraryCandidates(conf Config, name string) []string {
	roots := gstRuntimeRoots(conf)
	candidates := make([]string, 0, len(roots)*5)
	for _, dir := range gstLibraryDirCandidates(roots) {
		candidates = append(candidates, filepath.Join(dir, name))
	}
	return candidates
}

func gstreamerLibraryFound(conf Config) bool {
	for _, candidate := range linuxLibraryCandidates(conf, "libgstreamer-1.0.so.0") {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}
