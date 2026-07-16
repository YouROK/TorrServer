//go:build gst && (!embed_gstlib || !windows || (windows && !amd64))

package gstreamer

func embeddedGSTRuntimeRoot() string {
	return ""
}

func embeddedGSTRuntimeStatus() componentStatus {
	return componentStatus{}
}
