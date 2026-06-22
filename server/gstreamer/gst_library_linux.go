//go:build linux && (amd64 || arm64)

package gstreamer

import (
	"fmt"
	"path/filepath"

	"github.com/ebitengine/purego"
)

func loadGST(conf Config) (*gstAPI, error) {
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
		handles: []uintptr{glibHandle, gstHandle, gstAppHandle},
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
	if conf.GSTPath == "" {
		return nil
	}

	return []string{
		filepath.Join(conf.GSTPath, "lib", name),
		filepath.Join(conf.GSTPath, "lib64", name),
		filepath.Join(conf.GSTPath, "lib", "x86_64-linux-gnu", name),
		filepath.Join(conf.GSTPath, "lib", "aarch64-linux-gnu", name),
	}
}
