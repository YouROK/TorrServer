//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"fmt"
	"sync"
	"time"
)

const hardwareProbeTimeout = 5 * time.Second

type hardwareCandidate struct {
	name  string
	build func(bitrate int, keyIntMax int) string
}

type hardwareVideoState struct {
	sync.Once
	selected *hardwareCandidate
	width    int
	height   int
}

var hardwareVideo hardwareVideoState

var hardwareCandidates = []hardwareCandidate{
	{
		name: "Direct3D11 + Media Foundation",
		build: func(bitrate int, keyIntMax int) string {
			return fmt.Sprintf("d3d11upload ! d3d11convert ! video/x-raw(memory:D3D11Memory),format=NV12 ! mfh264enc name=video_encoder bitrate=%d gop-size=%d low-latency=true rc-mode=cbr ! ", bitrate, keyIntMax)
		},
	},
	{
		name: "NVIDIA NVENC",
		build: func(bitrate int, keyIntMax int) string {
			return fmt.Sprintf("videoconvert ! video/x-raw,format=NV12 ! nvh264enc name=video_encoder bitrate=%d gop-size=%d bframes=0 zerolatency=true rc-mode=cbr ! ", bitrate, keyIntMax)
		},
	},
	{
		name: "Intel Quick Sync",
		build: func(bitrate int, keyIntMax int) string {
			return fmt.Sprintf("videoconvert ! video/x-raw,format=NV12 ! qsvh264enc name=video_encoder bitrate=%d gop-size=%d b-frames=0 low-latency=true rate-control=cbr ! ", bitrate, keyIntMax)
		},
	},
	{
		name: "AMD AMF",
		build: func(bitrate int, keyIntMax int) string {
			return fmt.Sprintf("videoconvert ! video/x-raw,format=NV12 ! amfh264enc name=video_encoder bitrate=%d gop-size=%d b-frames=0 usage=low-latency preset=speed rate-control=cbr ! ", bitrate, keyIntMax)
		},
	},
	{
		name: "Direct3D12",
		build: func(bitrate int, keyIntMax int) string {
			return fmt.Sprintf("videoconvert ! video/x-raw,format=NV12 ! d3d12h264enc name=video_encoder bitrate=%d gop-size=%d rate-control=cbr ! ", bitrate, keyIntMax)
		},
	},
}

func hardwareH264Pipeline(width int, height int, bitrate int, keyIntMax int) string {
	if gstRuntime == nil {
		return ""
	}
	hardwareVideo.Do(probeHardwareVideo)
	selected := hardwareVideo.selected
	if selected == nil || width < 128 || height < 128 || width&1 != 0 || height&1 != 0 ||
		width > hardwareVideo.width || height > hardwareVideo.height {
		return ""
	}
	return selected.build(bitrate, keyIntMax)
}

func probeHardwareVideo() {
	for i := range hardwareCandidates {
		candidate := &hardwareCandidates[i]
		if probeHardwareCandidate(candidate, 3840, 2160) {
			hardwareVideo.selected = candidate
			hardwareVideo.width = 3840
			hardwareVideo.height = 2160
			return
		}
		if probeHardwareCandidate(candidate, 1920, 1080) {
			hardwareVideo.selected = candidate
			hardwareVideo.width = 1920
			hardwareVideo.height = 1080
			return
		}
	}
}

func probeHardwareCandidate(candidate *hardwareCandidate, width int, height int) bool {
	bufferSize := width * height * 3 / 2
	description := fmt.Sprintf(
		"fakesrc num-buffers=2 sizetype=fixed sizemax=%d do-timestamp=true format=time ! video/x-raw,format=NV12,width=%d,height=%d,framerate=30/1 ! %s h264parse ! video/x-h264,profile=main,stream-format=avc,alignment=au ! appsink name=probe_out emit-signals=false sync=false max-buffers=1 drop=false wait-on-eos=false",
		bufferSize,
		width,
		height,
		candidate.build(14_000, 270),
	)
	pipeline, err := gstRuntime.parseLaunch(description)
	if err != nil {
		return false
	}
	defer gstRuntime.objectUnref(pipeline)
	defer gstRuntime.elementSetState(pipeline, gstStateNull)

	sink := gstRuntime.binGetByName(pipeline, "probe_out")
	if sink == 0 {
		return false
	}
	defer gstRuntime.objectUnref(sink)

	if gstRuntime.elementSetState(pipeline, gstStatePlaying) == gstStateChangeFailure {
		return false
	}
	sample := gstRuntime.appSinkTryPullSample(sink, uint64(hardwareProbeTimeout))
	if sample == 0 {
		return false
	}
	gstRuntime.sampleUnref(sample)
	return true
}
