package media

import "testing"

func TestIsHDR(t *testing.T) {
	cases := []struct {
		name   string
		stream Stream
		want   bool
	}{
		{
			name:   "hdr10 pq with bt2020",
			stream: Stream{Kind: KindVideo, Codec: "hevc", BitDepth: 10, ColorTransfer: "smpte2084", ColorPrim: "bt2020"},
			want:   true,
		},
		{
			name:   "hlg broadcast",
			stream: Stream{Kind: KindVideo, Codec: "hevc", BitDepth: 10, ColorTransfer: "arib-std-b67", ColorPrim: "bt2020"},
			want:   true,
		},
		{
			name:   "ten bit with wide gamut",
			stream: Stream{Kind: KindVideo, Codec: "hevc", BitDepth: 10, ColorPrim: "bt2020"},
			want:   true,
		},
		{
			name:   "ten bit without wide gamut is not hdr",
			stream: Stream{Kind: KindVideo, Codec: "h264", BitDepth: 10, ColorPrim: "bt709"},
			want:   false,
		},
		{
			name:   "ordinary sdr",
			stream: Stream{Kind: KindVideo, Codec: "h264", BitDepth: 8, ColorTransfer: "bt709", ColorPrim: "bt709"},
			want:   false,
		},
		{
			name:   "no color metadata",
			stream: Stream{Kind: KindVideo, Codec: "h264", BitDepth: 8},
			want:   false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.stream.IsHDR(); got != c.want {
				t.Errorf("IsHDR = %v, want %v", got, c.want)
			}
		})
	}
}

func TestMediaInfoIsHDR(t *testing.T) {
	hdr := &MediaInfo{Streams: []Stream{
		{Kind: KindVideo, Codec: "hevc", BitDepth: 10, ColorTransfer: "smpte2084"},
	}}
	if !hdr.IsHDR() {
		t.Error("HDR file was not detected")
	}

	sdr := &MediaInfo{Streams: []Stream{
		{Kind: KindVideo, Codec: "h264", BitDepth: 8, ColorTransfer: "bt709"},
	}}
	if sdr.IsHDR() {
		t.Error("SDR file was detected as HDR")
	}

	var empty *MediaInfo
	if empty.IsHDR() {
		t.Error("nil info must not be HDR")
	}
}
