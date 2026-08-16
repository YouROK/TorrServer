//go:build gst && ((windows && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (darwin && (amd64 || arm64)))

package gstreamer

import (
	"math"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"
)

func TestGSTVersionInfoFormatting(t *testing.T) {
	version := gstVersionInfo{major: 1, minor: 26, micro: 2}
	if got := version.String(); got != "1.26.2" {
		t.Fatalf("String()=%q, want 1.26.2", got)
	}
	if got := version.configValue(); math.Abs(got-1.26) > 0.000001 {
		t.Fatalf("configValue()=%v, want 1.26", got)
	}
}

func TestLoadGStreamerRuntimeIfAvailable(t *testing.T) {
	conf := DefaultConfig()
	setupGStreamer(conf)

	api, err := loadGST(conf)
	if err != nil {
		t.Skipf("gstreamer runtime is not available: %v", err)
	}
	if err := api.init(); err != nil {
		t.Fatalf("gst init failed: %v", err)
	}

	pipeline, err := api.parseLaunch("fakesrc num-buffers=1 ! fakesink")
	if err != nil {
		t.Fatalf("parse launch failed: %v", err)
	}
	defer api.objectUnref(pipeline)

	if ret := api.elementSetState(pipeline, gstStateNull); ret == gstStateChangeFailure {
		t.Fatal("failed to set test pipeline to NULL")
	}
}

func TestParseLaunchRejectsPartialPipeline(t *testing.T) {
	message := append([]byte("partial pipeline parse error"), 0)
	gerror := make([]byte, 16)
	*(*uintptr)(unsafe.Pointer(&gerror[8])) = uintptr(unsafe.Pointer(&message[0]))

	var unrefed uintptr
	api := &gstAPI{
		gstParseLaunch: func(_ string, errOut unsafe.Pointer) uintptr {
			*(*uintptr)(errOut) = uintptr(unsafe.Pointer(&gerror[0]))
			return 123
		},
		gstObjectUnref: func(value uintptr) {
			unrefed = value
		},
		gErrorFree: func(uintptr) {},
	}

	pipeline, err := api.parseLaunch("broken ! pipeline")
	if err == nil {
		t.Fatal("partial parse error was ignored")
	}
	if pipeline != 0 {
		t.Fatalf("pipeline=%d, want 0", pipeline)
	}
	if unrefed != 123 {
		t.Fatalf("partial pipeline was not unrefed")
	}

	runtime.KeepAlive(message)
	runtime.KeepAlive(gerror)
}

func TestWithSampleBytesReturnsMapError(t *testing.T) {
	api := &gstAPI{
		gstSampleGetBuffer: func(uintptr) uintptr { return 2 },
		gstBufferGetSize:   func(uintptr) uintptr { return 10 },
		gstBufferMap:       func(uintptr, unsafe.Pointer, int32) int32 { return 0 },
	}

	err := api.withSampleBytes(1, func([]byte) error {
		t.Fatal("consume must not be called")
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "gst_buffer_map failed") {
		t.Fatalf("error=%v, want gst_buffer_map failed", err)
	}
}

func TestValidateGStreamerSampleSizeOK(t *testing.T) {
	if err := validateGStreamerSampleSize(1024, 1024); err != nil {
		t.Fatal(err)
	}
}

func TestGStreamerSampleSafetyLimit(t *testing.T) {
	const want = 256 * 1024 * 1024
	if maxGStreamerSampleBytes != want {
		t.Fatalf("maxGStreamerSampleBytes=%d, want %d", maxGStreamerSampleBytes, want)
	}
	if err := validateGStreamerSampleSize(want, want); err != nil {
		t.Fatalf("sample at safety limit was rejected: %v", err)
	}
}

func TestValidateGStreamerSampleSizeRejectsHugeBuffer(t *testing.T) {
	err := validateGStreamerSampleSize(uintptr(maxGStreamerSampleBytes+1), 1024)
	if err == nil || !strings.Contains(err.Error(), "gst buffer exceeds safety limit") {
		t.Fatalf("error=%v, want safety limit error", err)
	}
}

func TestValidateGStreamerSampleSizeRejectsHugeMap(t *testing.T) {
	err := validateGStreamerSampleSize(1024, maxGStreamerSampleBytes+1)
	if err == nil || !strings.Contains(err.Error(), "gst map size exceeds safety limit") {
		t.Fatalf("error=%v, want map safety limit error", err)
	}
}

func TestValidateGStreamerSampleSizeRejectsMapLargerThanBuffer(t *testing.T) {
	err := validateGStreamerSampleSize(1024, 2048)
	if err == nil || !strings.Contains(err.Error(), "gst map size exceeds buffer size") {
		t.Fatalf("error=%v, want map larger than buffer error", err)
	}
}

func TestValidateGStreamerSampleSizeRejectsNegativeMap(t *testing.T) {
	err := validateGStreamerSampleSize(1024, -1)
	if err == nil || !strings.Contains(err.Error(), "gst map size is negative") {
		t.Fatalf("error=%v, want negative map error", err)
	}
}

func TestWaitForSeekDoneConsumesAsyncDone(t *testing.T) {
	var filters []int32
	var unrefed uintptr
	api := &gstAPI{
		gstBusTimedPopFiltered: func(_ uintptr, _ uint64, filter int32) uintptr {
			filters = append(filters, filter)
			if filter == gstMessageAsyncDone {
				return 77
			}
			return 0
		},
		gstMiniObjectUnref: func(message uintptr) {
			unrefed = message
		},
	}

	done, err := api.waitForSeekDone(1, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !done {
		t.Fatal("ASYNC_DONE was not reported")
	}
	if unrefed != 77 {
		t.Fatalf("unrefed message=%d, want 77", unrefed)
	}
	want := []int32{gstMessageError, gstMessageEOS, gstMessageAsyncDone}
	if len(filters) != len(want) {
		t.Fatalf("filters=%v, want %v", filters, want)
	}
	for i := range want {
		if filters[i] != want[i] {
			t.Fatalf("filters=%v, want %v", filters, want)
		}
	}
}

func TestWaitForSeekDoneTreatsMissingAsyncDoneAsSoftTimeout(t *testing.T) {
	api := &gstAPI{
		gstBusTimedPopFiltered: func(uintptr, uint64, int32) uintptr { return 0 },
	}

	done, err := api.waitForSeekDone(1, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if done {
		t.Fatal("ASYNC_DONE was reported on timeout")
	}
}
