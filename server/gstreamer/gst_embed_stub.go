//go:build !windows || !amd64 || !embed_gstlib

package gstreamer

func embeddedGSTRuntimeRoot() string {
	return ""
}
