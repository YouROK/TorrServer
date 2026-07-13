package gstreamer

import (
	"fmt"

	"server/log"
	"server/settings"
)

func gstDebugf(format string, args ...any) {
	if settings.IsDebug() {
		log.TLogln("[GST]", fmt.Sprintf(format, args...))
	}
}

func gstErrorf(format string, args ...any) {
	log.TLogln("[GST ERROR]", fmt.Sprintf(format, args...))
}
