//go:build ios
// +build ios

package fuse

import (
	"server/log"
	"server/settings"
)

func FuseAutoMount() {
	if settings.Args != nil && settings.Args.FusePath != "" {
		log.TLogln("iOS does not support FUSE")
	}
}

func FuseCleanup() {
}
