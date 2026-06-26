//go:build !embed_gstlib || (!windows && !linux) || (windows && !amd64) || (linux && !amd64 && !arm64)

package gstreamer

func embeddedGSTRuntimeRoot() string {
	return ""
}
