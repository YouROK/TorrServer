//go:build !linux || (!amd64 && !arm64)

package gstreamer

func ensureGStreamerRuntimeEnv(_ Config) {}
