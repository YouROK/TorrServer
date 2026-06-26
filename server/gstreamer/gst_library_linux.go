//go:build linux && (amd64 || arm64)

package gstreamer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ebitengine/purego"
)

func loadGST(conf Config) (*gstAPI, error) {
	preloadHandles := preloadLinuxGSTRuntime(conf)

	glibHandle, err := loadLinuxLibrary(conf, "libglib-2.0.so.0")
	if err != nil {
		return nil, err
	}
	gstHandle, err := loadLinuxLibrary(conf, "libgstreamer-1.0.so.0")
	if err != nil {
		return nil, err
	}
	gstAppHandle, err := loadLinuxLibrary(conf, "libgstapp-1.0.so.0")
	if err != nil {
		return nil, err
	}

	api := &gstAPI{
		handles: append(preloadHandles, glibHandle, gstHandle, gstAppHandle),
	}
	if err := api.bind(gstHandle, gstAppHandle, glibHandle); err != nil {
		return nil, err
	}
	return api, nil
}

func loadLinuxLibrary(conf Config, name string) (uintptr, error) {
	for _, candidate := range linuxLibraryCandidates(conf, name) {
		handle, err := purego.Dlopen(candidate, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err == nil {
			return handle, nil
		}
	}

	handle, err := purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return 0, fmt.Errorf("load %s: %w", name, err)
	}
	return handle, nil
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

func preloadLinuxGSTRuntime(conf Config) []uintptr {
	candidates := linuxPreloadCandidates(conf)
	if len(candidates) == 0 {
		return nil
	}

	var handles []uintptr
	pending := candidates
	for pass := 0; pass < 8 && len(pending) > 0; pass++ {
		next := pending[:0]
		progress := false

		for _, candidate := range pending {
			handle, err := purego.Dlopen(candidate, purego.RTLD_NOW|purego.RTLD_GLOBAL)
			if err == nil {
				handles = append(handles, handle)
				progress = true
				continue
			}
			next = append(next, candidate)
		}

		if !progress {
			break
		}
		pending = next
	}

	return handles
}

func linuxPreloadCandidates(conf Config) []string {
	seen := make(map[string]struct{})
	var candidates []string

	for _, dir := range existingPaths(gstLibraryDirCandidates(gstRuntimeRoots(conf))) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !linuxPreloadLibraryName(entry.Name()) {
				continue
			}

			path := filepath.Join(dir, entry.Name())
			clean := filepath.Clean(path)
			if _, ok := seen[clean]; ok {
				continue
			}
			seen[clean] = struct{}{}
			candidates = append(candidates, clean)
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return linuxPreloadPriority(filepath.Base(candidates[i])) < linuxPreloadPriority(filepath.Base(candidates[j]))
	})

	return candidates
}

func linuxPreloadLibraryName(name string) bool {
	if !strings.Contains(name, ".so") {
		return false
	}
	if strings.HasPrefix(name, ".") {
		return false
	}
	if !strings.HasPrefix(name, "libgst") {
		return true
	}
	return strings.Contains(name, "-1.0.so.")
}

func linuxPreloadPriority(name string) int {
	switch {
	case !strings.HasPrefix(name, "libgst"):
		return 0
	case strings.HasPrefix(name, "libgstreamer-"):
		return 1
	default:
		return 2
	}
}
