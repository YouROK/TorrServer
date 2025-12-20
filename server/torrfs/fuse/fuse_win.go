//go:build windows
// +build windows

package fuse

import (
	"server/log"
	"server/settings"
)

func FuseAutoMount() {
	if settings.Args.FusePath != "" {
		log.TLogln("Windows not support FUSE")
	}
}

func FuseCleanup() {
}
