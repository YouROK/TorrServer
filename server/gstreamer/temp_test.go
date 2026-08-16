//go:build gst

package gstreamer

import (
	"reflect"
	"testing"
)

func TestConfigDoesNotExposeGStreamerTempFiles(t *testing.T) {
	typ := reflect.TypeOf(Config{})
	for _, field := range []string{"TempFS", "TempFSRing", "AppSinkBuffers"} {
		if _, ok := typ.FieldByName(field); ok {
			t.Fatalf("Config still exposes removed field %s", field)
		}
	}
}
