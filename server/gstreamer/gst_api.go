//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"errors"
	"fmt"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	gstStateNull    int32 = 1
	gstStateReady   int32 = 2
	gstStatePaused  int32 = 3
	gstStatePlaying int32 = 4

	gstStateChangeFailure   int32 = 0
	gstStateChangeSuccess   int32 = 1
	gstStateChangeAsync     int32 = 2
	gstStateChangeNoPreroll int32 = 3

	gstFormatTime int32 = 3

	gstSeekFlagFlush     int32 = 1
	gstSeekFlagAccurate  int32 = 2
	gstSeekFlagKeyUnit   int32 = 4
	gstSeekFlagSnapAfter int32 = 64

	gstSeekTypeNone int32 = 0
	gstSeekTypeSet  int32 = 1

	gstMapRead int32 = 1

	gstMessageError int32 = 1 << 1
)

const maxGStreamerSampleBytes = 128 * 1024 * 1024

type gstVersionInfo struct {
	major uint32
	minor uint32
	micro uint32
	nano  uint32
}

func (v gstVersionInfo) valid() bool {
	return v.major != 0
}

func (v gstVersionInfo) atLeast(major uint32, minor uint32) bool {
	if v.major != major {
		return v.major > major
	}
	return v.minor >= minor
}

func (v gstVersionInfo) configValue() float64 {
	return float64(v.major) + float64(v.minor)/100
}

func (v gstVersionInfo) String() string {
	if !v.valid() {
		return ""
	}
	if v.nano != 0 {
		return fmt.Sprintf("%d.%d.%d.%d", v.major, v.minor, v.micro, v.nano)
	}
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.micro)
}

type gstAPI struct {
	handles []uintptr

	gstInitCheck            func(argc unsafe.Pointer, argv unsafe.Pointer, err unsafe.Pointer) int32
	gstVersion              func(major unsafe.Pointer, minor unsafe.Pointer, micro unsafe.Pointer, nano unsafe.Pointer)
	gstParseLaunch          func(description string, err unsafe.Pointer) uintptr
	gstBinGetByName         func(bin uintptr, name string) uintptr
	gstObjectUnref          func(obj uintptr)
	gstMiniObjectUnref      func(obj uintptr)
	gstElementSetState      func(element uintptr, state int32) int32
	gstElementGetState      func(element uintptr, state unsafe.Pointer, pending unsafe.Pointer, timeout uint64) int32
	gstElementGetStaticPad  func(element uintptr, name string) uintptr
	gstElementQueryPosition func(element uintptr, format int32, cur unsafe.Pointer) int32
	gstEventNewSeek         func(rate float64, format int32, flags int32, startType int32, start int64, stopType int32, stop int64) uintptr
	gstPadSendEvent         func(pad uintptr, event uintptr) int32
	gstPipelineGetBus       func(pipeline uintptr) uintptr
	gstBusTimedPopFiltered  func(bus uintptr, timeout uint64, types int32) uintptr
	gstMessageParseError    func(msg uintptr, err unsafe.Pointer, debug unsafe.Pointer)
	gstSampleGetBuffer      func(sample uintptr) uintptr
	gstSampleUnref          func(sample uintptr)
	gstBufferGetSize        func(buffer uintptr) uintptr
	gstBufferMap            func(buffer uintptr, mapInfo unsafe.Pointer, flags int32) int32
	gstBufferUnmap          func(buffer uintptr, mapInfo unsafe.Pointer)

	gstAppSinkTryPullSample func(sink uintptr, timeout uint64) uintptr
	gstAppSinkIsEOS         func(sink uintptr) int32

	gErrorFree func(err uintptr)
	gFree      func(ptr uintptr)

	version gstVersionInfo
}

func (g *gstAPI) bind(gstHandle uintptr, gstAppHandle uintptr, glibHandle uintptr) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("gstreamer symbol bind failed: %v", recovered)
		}
	}()

	purego.RegisterLibFunc(&g.gstInitCheck, gstHandle, "gst_init_check")
	purego.RegisterLibFunc(&g.gstVersion, gstHandle, "gst_version")
	purego.RegisterLibFunc(&g.gstParseLaunch, gstHandle, "gst_parse_launch")
	purego.RegisterLibFunc(&g.gstBinGetByName, gstHandle, "gst_bin_get_by_name")
	purego.RegisterLibFunc(&g.gstObjectUnref, gstHandle, "gst_object_unref")
	purego.RegisterLibFunc(&g.gstMiniObjectUnref, gstHandle, "gst_mini_object_unref")
	purego.RegisterLibFunc(&g.gstElementSetState, gstHandle, "gst_element_set_state")
	purego.RegisterLibFunc(&g.gstElementGetState, gstHandle, "gst_element_get_state")
	purego.RegisterLibFunc(&g.gstElementGetStaticPad, gstHandle, "gst_element_get_static_pad")
	purego.RegisterLibFunc(&g.gstElementQueryPosition, gstHandle, "gst_element_query_position")
	purego.RegisterLibFunc(&g.gstEventNewSeek, gstHandle, "gst_event_new_seek")
	purego.RegisterLibFunc(&g.gstPadSendEvent, gstHandle, "gst_pad_send_event")
	purego.RegisterLibFunc(&g.gstPipelineGetBus, gstHandle, "gst_pipeline_get_bus")
	purego.RegisterLibFunc(&g.gstBusTimedPopFiltered, gstHandle, "gst_bus_timed_pop_filtered")
	purego.RegisterLibFunc(&g.gstMessageParseError, gstHandle, "gst_message_parse_error")
	purego.RegisterLibFunc(&g.gstSampleGetBuffer, gstHandle, "gst_sample_get_buffer")
	purego.RegisterLibFunc(&g.gstSampleUnref, gstHandle, "gst_sample_unref")
	purego.RegisterLibFunc(&g.gstBufferGetSize, gstHandle, "gst_buffer_get_size")
	purego.RegisterLibFunc(&g.gstBufferMap, gstHandle, "gst_buffer_map")
	purego.RegisterLibFunc(&g.gstBufferUnmap, gstHandle, "gst_buffer_unmap")

	purego.RegisterLibFunc(&g.gstAppSinkTryPullSample, gstAppHandle, "gst_app_sink_try_pull_sample")
	purego.RegisterLibFunc(&g.gstAppSinkIsEOS, gstAppHandle, "gst_app_sink_is_eos")

	purego.RegisterLibFunc(&g.gErrorFree, glibHandle, "g_error_free")
	purego.RegisterLibFunc(&g.gFree, glibHandle, "g_free")

	return nil
}

func (g *gstAPI) init() error {
	var errPtr uintptr
	if g.gstInitCheck(nil, nil, unsafe.Pointer(&errPtr)) == 0 {
		msg := g.takeGError(errPtr)
		if msg == "" {
			msg = "gst_init_check failed"
		}
		return errors.New(msg)
	}
	if g.gstVersion != nil {
		var major, minor, micro, nano uint32
		g.gstVersion(
			unsafe.Pointer(&major),
			unsafe.Pointer(&minor),
			unsafe.Pointer(&micro),
			unsafe.Pointer(&nano),
		)
		g.version = gstVersionInfo{
			major: major,
			minor: minor,
			micro: micro,
			nano:  nano,
		}
	}
	return nil
}

func (g *gstAPI) parseLaunch(description string) (uintptr, error) {
	var errPtr uintptr
	pipeline := g.gstParseLaunch(description, unsafe.Pointer(&errPtr))

	if errPtr != 0 {
		msg := g.takeGError(errPtr)
		if pipeline != 0 {
			g.objectUnref(pipeline)
		}
		if msg == "" {
			msg = "gst_parse_launch reported an error"
		}
		return 0, errors.New(msg)
	}

	if pipeline == 0 {
		return 0, errors.New("gst_parse_launch failed without GError")
	}

	return pipeline, nil
}

func (g *gstAPI) binGetByName(bin uintptr, name string) uintptr {
	if bin == 0 {
		return 0
	}
	return g.gstBinGetByName(bin, name)
}

func (g *gstAPI) objectUnref(obj uintptr) {
	if obj != 0 {
		g.gstObjectUnref(obj)
	}
}

func (g *gstAPI) miniObjectUnref(obj uintptr) {
	if obj != 0 {
		g.gstMiniObjectUnref(obj)
	}
}

func (g *gstAPI) elementSetState(element uintptr, state int32) int32 {
	if element == 0 {
		return gstStateChangeFailure
	}
	return g.gstElementSetState(element, state)
}

func (g *gstAPI) elementGetState(element uintptr, timeout time.Duration) int32 {
	if element == 0 {
		return gstStateChangeFailure
	}
	return g.gstElementGetState(element, nil, nil, uint64(timeout))
}

func (g *gstAPI) elementGetStaticPad(element uintptr, name string) uintptr {
	if element == 0 || g.gstElementGetStaticPad == nil {
		return 0
	}
	return g.gstElementGetStaticPad(element, name)
}

func (g *gstAPI) sendTimeSeekEvent(pad uintptr, flags int32, position int64) bool {
	if pad == 0 || g.gstEventNewSeek == nil || g.gstPadSendEvent == nil {
		return false
	}
	event := g.gstEventNewSeek(
		1.0,
		gstFormatTime,
		flags,
		gstSeekTypeSet,
		position,
		gstSeekTypeNone,
		-1,
	)
	if event == 0 {
		return false
	}

	// gst_pad_send_event takes ownership of the event.
	return g.gstPadSendEvent(pad, event) != 0
}

func (g *gstAPI) elementQueryPosition(element uintptr) (int64, bool) {
	if element == 0 || g.gstElementQueryPosition == nil {
		return 0, false
	}
	var position int64
	if g.gstElementQueryPosition(element, gstFormatTime, unsafe.Pointer(&position)) == 0 {
		return 0, false
	}
	if position < 0 {
		return 0, false
	}
	return position, true
}

func (g *gstAPI) pipelineGetBus(pipeline uintptr) uintptr {
	if pipeline == 0 {
		return 0
	}
	return g.gstPipelineGetBus(pipeline)
}

func (g *gstAPI) appSinkTryPullSample(sink uintptr, timeout uint64) uintptr {
	if sink == 0 {
		return 0
	}
	return g.gstAppSinkTryPullSample(sink, timeout)
}

func (g *gstAPI) appSinkIsEOS(sink uintptr) bool {
	return sink != 0 && g.gstAppSinkIsEOS(sink) != 0
}

func (g *gstAPI) sampleUnref(sample uintptr) {
	if sample != 0 {
		g.gstSampleUnref(sample)
	}
}

func (g *gstAPI) withSampleBytes(sample uintptr, consume func([]byte) error) error {
	if sample == 0 {
		return nil
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
	return consume(data)
}

func validateGStreamerSampleSize(bufferSize uintptr, mapSize int) error {
	maxInt := int(^uint(0) >> 1)

	if bufferSize > uintptr(maxInt) {
		return fmt.Errorf("gst buffer is too large: %d bytes", bufferSize)
	}
	if bufferSize > uintptr(maxGStreamerSampleBytes) {
		return fmt.Errorf("gst buffer exceeds safety limit: %d bytes", bufferSize)
	}
	if mapSize < 0 {
		return errors.New("gst map size is negative")
	}
	if mapSize > maxGStreamerSampleBytes {
		return fmt.Errorf("gst map size exceeds safety limit: %d bytes", mapSize)
	}
	if uintptr(mapSize) > bufferSize {
		return fmt.Errorf("gst map size exceeds buffer size: map=%d buffer=%d", mapSize, bufferSize)
	}
	return nil
}

func (g *gstAPI) popBusError(bus uintptr, timeout time.Duration) error {
	if bus == 0 {
		return nil
	}

	msg := g.gstBusTimedPopFiltered(bus, uint64(timeout), gstMessageError)
	if msg == 0 {
		return nil
	}

	defer g.miniObjectUnref(msg)
	message := g.parseMessageError(msg)
	if message == "" {
		message = "gstreamer bus error"
	}
	return errors.New(message)
}

func (g *gstAPI) parseMessageError(msg uintptr) string {
	var errPtr uintptr
	var debugPtr uintptr
	g.gstMessageParseError(msg, unsafe.Pointer(&errPtr), unsafe.Pointer(&debugPtr))

	message := g.takeGError(errPtr)
	if debug := cString(debugPtr); debug != "" {
		if message != "" {
			message += ": " + debug
		} else {
			message = debug
		}
	}
	if debugPtr != 0 {
		g.gFree(debugPtr)
	}
	return message
}

func (g *gstAPI) takeGError(errPtr uintptr) string {
	if errPtr == 0 {
		return ""
	}
	messagePtr := *(*uintptr)(unsafe.Pointer(errPtr + 8))
	message := cString(messagePtr)
	g.gErrorFree(errPtr)
	return message
}

func gstMapInfoData(mapInfo *[128]byte) (uintptr, int) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	dataOffset := alignTo(uintptr(ptrSize)+4, uintptr(ptrSize))
	sizeOffset := dataOffset + uintptr(ptrSize)

	dataPtr := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(&mapInfo[0])) + dataOffset))
	size := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(&mapInfo[0])) + sizeOffset))
	return dataPtr, int(size)
}

func alignTo(value uintptr, alignment uintptr) uintptr {
	if alignment == 0 {
		return value
	}
	remainder := value % alignment
	if remainder == 0 {
		return value
	}
	return value + alignment - remainder
}

func cString(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}

	var out []byte
	for offset := uintptr(0); ; offset++ {
		b := *(*byte)(unsafe.Pointer(ptr + offset))
		if b == 0 {
			break
		}
		out = append(out, b)
	}
	return string(out)
}
