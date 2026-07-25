//go:build gst

package gstreamer

import (
	"bytes"
	"io"
	"testing"
)

func TestSegmentWriteToWritesHeaderAndPayloads(t *testing.T) {
	seg := Segment{
		Header:   []byte("head"),
		Payloads: [][]byte{[]byte("video"), []byte("audio")},
	}

	var out bytes.Buffer
	written, err := seg.WriteTo(&out)
	if err != nil {
		t.Fatal(err)
	}
	if written != int64(seg.Len()) {
		t.Fatalf("WriteTo() wrote %d bytes, want %d", written, seg.Len())
	}

	if got, want := out.String(), "headvideoaudio"; got != want {
		t.Fatalf("WriteTo() = %q, want %q", got, want)
	}
	if got, want := seg.Len(), len("headvideoaudio"); got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
}

func TestSegmentWriteRangeCrossesParts(t *testing.T) {
	seg := Segment{
		Header:   []byte("head"),
		Payloads: [][]byte{[]byte("video"), []byte("audio")},
	}

	var out bytes.Buffer
	if err := seg.WriteRange(&out, 2, 8); err != nil {
		t.Fatal(err)
	}

	if got, want := out.String(), "advideoa"; got != want {
		t.Fatalf("WriteRange() = %q, want %q", got, want)
	}
}

func TestSegmentWriteRangeDoesNotAllocateParts(t *testing.T) {
	seg := Segment{
		Header:   []byte("head"),
		Payloads: [][]byte{[]byte("video"), []byte("audio")},
	}

	allocs := testing.AllocsPerRun(1000, func() {
		if err := seg.WriteRange(io.Discard, 2, 8); err != nil {
			t.Fatal(err)
		}
	})
	if allocs != 0 {
		t.Fatalf("WriteRange allocations = %v, want 0", allocs)
	}
}

type shortWriter struct{}

func (shortWriter) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	return len(data) - 1, nil
}

func TestSegmentWriteToRejectsShortWrite(t *testing.T) {
	seg := Segment{Header: []byte("header")}
	if _, err := seg.WriteTo(shortWriter{}); err != io.ErrShortWrite {
		t.Fatalf("WriteTo error=%v, want io.ErrShortWrite", err)
	}
}

func TestSegmentWriteRangeRejectsShortWrite(t *testing.T) {
	seg := Segment{Header: []byte("header")}
	if err := seg.WriteRange(shortWriter{}, 1, 3); err != io.ErrShortWrite {
		t.Fatalf("WriteRange error=%v, want io.ErrShortWrite", err)
	}
}
