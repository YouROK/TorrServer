//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

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

	status.Version = gstRuntime.version.String()
	status.Works = true
	return status
}

func checkHDRToneMapping(gstreamer componentStatus) componentStatus {
	if !gstreamer.Works {
		return componentStatus{}
	}
	status := componentStatus{Found: gstElementAvailable("hdrtonemap")}
	status.Available = status.Found
	status.Works = status.Found
	if !status.Found {
		status.Error = "hdrtonemap element is not available"
	}
	return status
}
