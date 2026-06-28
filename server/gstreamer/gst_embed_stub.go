//go:build !embed_gstlib || !windows || (windows && !amd64)

package gstreamer

func embeddedGSTRuntimeRoot() string {
	return ""
}
