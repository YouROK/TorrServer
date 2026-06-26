//go:build windows && (amd64 || arm64)

package gstreamer

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func loadGST(conf Config) (*gstAPI, error) {
	for _, root := range gstRuntimeRoots(conf) {
		gstBin := filepath.Join(root, "bin")
		if _, statErr := os.Stat(gstBin); statErr == nil {
			if err := windows.SetDllDirectory(gstBin); err != nil {
				return nil, fmt.Errorf("set gstreamer dll directory %q: %w", gstBin, err)
			}
			break
		}
	}

	glibHandle, err := loadWindowsLibrary(conf, "libglib-2.0-0.dll")
	if err != nil {
		return nil, err
	}
	gstHandle, err := loadWindowsLibrary(conf, "libgstreamer-1.0-0.dll")
	if err != nil {
		return nil, err
	}
	gstAppHandle, err := loadWindowsLibrary(conf, "libgstapp-1.0-0.dll")
	if err != nil {
		return nil, err
	}

	api := &gstAPI{
		handles: []uintptr{uintptr(glibHandle), uintptr(gstHandle), uintptr(gstAppHandle)},
	}
	if err := api.bind(uintptr(gstHandle), uintptr(gstAppHandle), uintptr(glibHandle)); err != nil {
		return nil, err
	}
	return api, nil
}

func loadWindowsLibrary(conf Config, name string) (windows.Handle, error) {
	for _, root := range gstRuntimeRoots(conf) {
		fullPath := filepath.Join(root, "bin", name)
		handle, err := windows.LoadLibraryEx(fullPath, 0, windows.LOAD_WITH_ALTERED_SEARCH_PATH)
		if err == nil {
			return handle, nil
		}
	}

	handle, err := windows.LoadLibrary(name)
	if err != nil {
		return 0, fmt.Errorf("load %s: %w", name, err)
	}
	return handle, nil
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
