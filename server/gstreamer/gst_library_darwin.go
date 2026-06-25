//go:build darwin && (amd64 || arm64)

package gstreamer

import (
	"fmt"
	"path/filepath"

	"github.com/ebitengine/purego"
)

func loadGST(conf Config) (*gstAPI, error) {
	glibHandle, err := loadDarwinLibrary(conf, "libglib-2.0.0.dylib", "libglib-2.0.dylib")
	if err != nil {
		return nil, err
	}
	gstHandle, err := loadDarwinLibrary(conf, "libgstreamer-1.0.0.dylib", "libgstreamer-1.0.dylib")
	if err != nil {
		return nil, err
	}
	gstAppHandle, err := loadDarwinLibrary(conf, "libgstapp-1.0.0.dylib", "libgstapp-1.0.dylib")
	if err != nil {
		return nil, err
	}

	api := &gstAPI{
		handles: []uintptr{glibHandle, gstHandle, gstAppHandle},
	}
	if err := api.bind(gstHandle, gstAppHandle, glibHandle); err != nil {
		return nil, err
	}
	return api, nil
}

func loadDarwinLibrary(conf Config, names ...string) (uintptr, error) {
	var lastErr error
	for _, name := range names {
		for _, candidate := range darwinLibraryCandidates(conf, name) {
			handle, err := purego.Dlopen(candidate, purego.RTLD_NOW|purego.RTLD_GLOBAL)
			if err == nil {
				return handle, nil
			}
			lastErr = err
		}

		handle, err := purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err == nil {
			return handle, nil
		}
		lastErr = err
	}

	return 0, fmt.Errorf("load %s: %w", names[0], lastErr)
}

func darwinLibraryCandidates(conf Config, name string) []string {
	roots := gstRuntimeRoots(conf)
	var candidates []string
	for _, dir := range gstLibraryDirCandidates(roots) {
		candidates = append(candidates, filepath.Join(dir, name))
	}
	return candidates
}
