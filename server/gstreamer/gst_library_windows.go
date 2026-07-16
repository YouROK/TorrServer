//go:build gst && windows && (amd64 || arm64)

package gstreamer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

func loadGST(conf Config) (*gstAPI, error) {
	roots := gstRuntimeRoots(conf)
	var failures []string
	for _, root := range roots {
		api, err := loadWindowsGSTRoot(root)
		if err == nil {
			return api, nil
		}
		failures = append(failures, fmt.Sprintf("%s: %v", root, err))
	}

	if embedded := embeddedGSTRuntimeRoot(); embedded != "" && !containsPath(roots, embedded) {
		api, err := loadWindowsGSTRoot(embedded)
		if err == nil {
			return api, nil
		}
		failures = append(failures, fmt.Sprintf("%s: %v", embedded, err))
	}

	_ = windows.SetDllDirectory("")
	api, err := loadWindowsGST(func(name string) (windows.Handle, error) {
		return windows.LoadLibrary(name)
	})
	if err == nil {
		return api, nil
	}
	failures = append(failures, fmt.Sprintf("system search path: %v", err))
	return nil, fmt.Errorf("load gstreamer runtime: %s", strings.Join(failures, "; "))
}

func loadWindowsGSTRoot(root string) (*gstAPI, error) {
	bin := filepath.Join(root, "bin")
	if err := windows.SetDllDirectory(bin); err != nil {
		return nil, fmt.Errorf("set DLL directory: %w", err)
	}
	api, err := loadWindowsGST(func(name string) (windows.Handle, error) {
		return windows.LoadLibraryEx(filepath.Join(bin, name), 0, windows.LOAD_WITH_ALTERED_SEARCH_PATH)
	})
	if err != nil {
		return nil, err
	}
	setupGStreamerRoots([]string{root})
	return api, nil
}

func loadWindowsGST(load func(string) (windows.Handle, error)) (*gstAPI, error) {
	names := [...]string{"libglib-2.0-0.dll", "libgstreamer-1.0-0.dll", "libgstapp-1.0-0.dll"}
	handles := make([]windows.Handle, 0, len(names))
	for _, name := range names {
		handle, err := load(name)
		if err != nil {
			freeWindowsLibraries(handles)
			return nil, fmt.Errorf("load %s: %w", name, err)
		}
		handles = append(handles, handle)
	}

	api := &gstAPI{
		handles: []uintptr{uintptr(handles[0]), uintptr(handles[1]), uintptr(handles[2])},
	}
	if err := api.bind(uintptr(handles[1]), uintptr(handles[2]), uintptr(handles[0])); err != nil {
		freeWindowsLibraries(handles)
		return nil, err
	}
	return api, nil
}

func freeWindowsLibraries(handles []windows.Handle) {
	for i := len(handles) - 1; i >= 0; i-- {
		_ = windows.FreeLibrary(handles[i])
	}
}

func gstreamerLibraryFound(conf Config) bool {
	for _, root := range gstRuntimeRoots(conf) {
		fullPath := filepath.Join(root, "bin", "libgstreamer-1.0-0.dll")
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}
