//go:build gst && darwin && (amd64 || arm64)

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
		api, err := loadDarwinGSTRoot(root)
		if err == nil {
			return api, nil
		}
		failures = append(failures, fmt.Sprintf("%s: %v", root, err))
	}

	api, err := loadDarwinGST(func(names []string) (uintptr, error) {
		return openDarwinLibrary(nil, names)
	})
	if err == nil {
		return api, nil
	}
	failures = append(failures, fmt.Sprintf("system search path: %v", err))
	return nil, fmt.Errorf("load gstreamer runtime: %s", strings.Join(failures, "; "))
}

func loadDarwinGSTRoot(root string) (*gstAPI, error) {
	dirs := gstLibraryDirCandidates([]string{root})
	api, err := loadDarwinGST(func(names []string) (uintptr, error) {
		return openDarwinLibrary(dirs, names)
	})
	if err != nil {
		return nil, err
	}
	setupGStreamerRoots([]string{root})
	return api, nil
}

func loadDarwinGST(load func([]string) (uintptr, error)) (*gstAPI, error) {
	names := [...][2]string{
		{"libglib-2.0.0.dylib", "libglib-2.0.dylib"},
		{"libgstreamer-1.0.0.dylib", "libgstreamer-1.0.dylib"},
		{"libgstapp-1.0.0.dylib", "libgstapp-1.0.dylib"},
	}
	handles := make([]uintptr, 0, len(names))
	for _, aliases := range names {
		handle, err := load(aliases[:])
		if err != nil {
			closeDarwinLibraries(handles)
			return nil, fmt.Errorf("load %s: %w", aliases[0], err)
		}
		handles = append(handles, handle)
	}

	api := &gstAPI{handles: handles}
	if err := api.bind(handles[1], handles[2], handles[0]); err != nil {
		closeDarwinLibraries(handles)
		return nil, err
	}
	return api, nil
}

func openDarwinLibrary(dirs []string, names []string) (uintptr, error) {
	var lastErr error
	for _, name := range names {
		for _, dir := range dirs {
			handle, err := purego.Dlopen(filepath.Join(dir, name), purego.RTLD_NOW|purego.RTLD_GLOBAL)
			if err == nil {
				return handle, nil
			}
			lastErr = err
		}

		if len(dirs) == 0 {
			handle, err := purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
			if err == nil {
				return handle, nil
			}
			lastErr = err
		}
	}

	return 0, lastErr
}

func closeDarwinLibraries(handles []uintptr) {
	for i := len(handles) - 1; i >= 0; i-- {
		_ = purego.Dlclose(handles[i])
	}
}

func darwinLibraryCandidates(conf Config, name string) []string {
	roots := gstRuntimeRoots(conf)
	var candidates []string
	for _, dir := range gstLibraryDirCandidates(roots) {
		candidates = append(candidates, filepath.Join(dir, name))
	}
	return candidates
}

func gstreamerLibraryFound(conf Config) bool {
	for _, name := range []string{"libgstreamer-1.0.0.dylib", "libgstreamer-1.0.dylib"} {
		for _, candidate := range darwinLibraryCandidates(conf, name) {
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return true
			}
		}
	}
	return false
}
