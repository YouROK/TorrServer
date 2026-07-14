//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"errors"
	"unsafe"
)

func (g *gstAPI) withSampleBuffer(sample uintptr, consume func(gstBufferSnapshot, []byte) error) error {
	if sample == 0 {
		return nil
	}
	if consume == nil {
		return errors.New("nil gst sample consumer")
	}

	buffer := g.gstSampleGetBuffer(sample)
	if buffer == 0 {
		return errors.New("gst_sample_get_buffer returned nil")
	}

	bufferSize := g.gstBufferGetSize(buffer)
	if bufferSize == 0 {
		return nil
	}
	if err := validateGStreamerSampleSize(bufferSize, 0); err != nil {
		return err
	}

	meta := readGstBuffer(buffer)

	var mapInfo [128]byte
	if g.gstBufferMap(buffer, unsafe.Pointer(&mapInfo[0]), gstMapRead) == 0 {
		return errors.New("gst_buffer_map failed")
	}
	defer g.gstBufferUnmap(buffer, unsafe.Pointer(&mapInfo[0]))

	dataPtr, size := gstMapInfoData(&mapInfo)
	if dataPtr == 0 || size == 0 {
		return nil
	}
	if err := validateGStreamerSampleSize(bufferSize, size); err != nil {
		return err
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(dataPtr)), size)
	return consume(meta, data)
}
