//go:build (windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64))

package gstreamer

func checkGStreamer(conf Config) componentStatus {
	gstInitOnce.Do(func() {
		initGStreamerRuntime(conf)
	})

	status := gstInitStatus
	if gstInitErr != nil {
		status.Error = gstInitErr.Error()
	}
	if !status.Available || gstRuntime == nil {
		return status
	}

	status.Works = true
	return status
}
