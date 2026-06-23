//go:build (windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64))

package gstreamer

import "time"

func checkGStreamer(conf Config) componentStatus {
	gstInitOnce.Do(func() {
		initGStreamerRuntime(conf)
	})

	status := gstInitStatus
	if !status.Available || gstRuntime == nil {
		return status
	}

	if checkGStreamerPipeline(gstRuntime) == nil {
		status.Works = true
	}
	return status
}

func checkGStreamerPipeline(api *gstAPI) error {
	pipeline, err := api.parseLaunch("fakesrc num-buffers=1 ! fakesink")
	if err != nil {
		return err
	}
	defer api.objectUnref(pipeline)

	bus := api.pipelineGetBus(pipeline)
	if bus != 0 {
		defer api.objectUnref(bus)
	}

	if ret := api.elementSetState(pipeline, gstStatePlaying); ret == gstStateChangeFailure {
		_ = api.elementSetState(pipeline, gstStateNull)
		return api.popBusError(bus, 0)
	}
	if ret := api.elementGetState(pipeline, 5*time.Second); ret == gstStateChangeFailure {
		_ = api.elementSetState(pipeline, gstStateNull)
		return api.popBusError(bus, 0)
	}

	if ret := api.elementSetState(pipeline, gstStateNull); ret == gstStateChangeFailure {
		return api.popBusError(bus, 0)
	}
	return nil
}
