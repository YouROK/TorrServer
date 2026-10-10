package job

import (
	"testing"
	"time"
)

func TestParseProgressLines(t *testing.T) {
	lines := []string{
		"frame=125",
		"fps=25.50",
		"stream_0_0_q=-1.0",
		"bitrate= 1234.5kbits/s",
		"total_size=45678",
		"out_time_us=5000000",
		"out_time_ms=5000000",
		"out_time=00:00:05.000000",
		"dup_frames=0",
		"drop_frames=1",
		"speed=1.25x",
		"progress=continue",
	}

	var p Progress
	for _, line := range lines {
		parseProgressLine(&p, line)
	}

	if p.Frame != 125 {
		t.Errorf("frame = %d, want 125", p.Frame)
	}
	if p.FPS != 25.5 {
		t.Errorf("fps = %v, want 25.5", p.FPS)
	}
	if p.BitrateBps != 1234500 {
		t.Errorf("bitrate = %d, want 1234500", p.BitrateBps)
	}
	if p.TotalSize != 45678 {
		t.Errorf("total size = %d, want 45678", p.TotalSize)
	}
	if p.OutTime != 5*time.Second {
		t.Errorf("out time = %v, want 5s", p.OutTime)
	}
	if p.Speed != 1.25 {
		t.Errorf("speed = %v, want 1.25", p.Speed)
	}
	if p.Done {
		t.Error("progress must not be marked as done")
	}

	parseProgressLine(&p, "progress=end")
	if !p.Done {
		t.Error("progress=end was not recognized")
	}
}

func TestParseProgressIgnoresGarbage(t *testing.T) {
	before := Progress{Frame: 10, FPS: 5}

	p := before
	for _, line := range []string{
		"",
		"some ffmpeg diagnostic output",
		"frame=",
		"=value",
		"[info] stream mapping:",
	} {
		parseProgressLine(&p, line)
	}

	if p.Frame != before.Frame {
		t.Errorf("frame changed to %d, want %d", p.Frame, before.Frame)
	}
	if p.FPS != before.FPS {
		t.Errorf("fps changed to %v, want %v", p.FPS, before.FPS)
	}
}

func TestParseBitrateVariants(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{in: "1234kbits/s", want: 1234000},
		{in: "1234.5kbits/s", want: 1234500},
		{in: "N/A", want: 0},
		{in: "", want: 0},
		{in: "500k", want: 500000},
	}

	for _, c := range cases {
		if got := parseBitrate(c.in); got != c.want {
			t.Errorf("parseBitrate(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestParseTimeVariants(t *testing.T) {
	cases := []struct {
		in      string
		want    time.Duration
		wantErr bool
	}{
		{in: "00:00:05.000000", want: 5 * time.Second},
		{in: "01:02:03.500000", want: time.Hour + 2*time.Minute + 3500*time.Millisecond},
		{in: "bad", wantErr: true},
		{in: "00:00", wantErr: true},
	}

	for _, c := range cases {
		got, err := parseTime(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseTime(%q): expected an error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseTime(%q): %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseTime(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestProgressPercent(t *testing.T) {
	cases := []struct {
		position time.Duration
		duration time.Duration
		want     float64
	}{
		{position: 5 * time.Second, duration: 10 * time.Second, want: 50},
		{position: 20 * time.Second, duration: 10 * time.Second, want: 100},
		{position: -1 * time.Second, duration: 10 * time.Second, want: 0},
		{position: 5 * time.Second, duration: 0, want: 0},
	}

	for _, c := range cases {
		p := Progress{OutTime: c.position}
		if got := p.Percent(c.duration); got != c.want {
			t.Errorf("Percent(%v/%v) = %v, want %v", c.position, c.duration, got, c.want)
		}
	}
}

func TestIsProgressLine(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{line: "frame=10", want: true},
		{line: "out_time_us=1000", want: true},
		{line: "progress=end", want: true},
		{line: "[libx264 @ 0x1] using cpu capabilities", want: false},
		{line: "some random text", want: false},
		{line: "", want: false},
	}

	for _, c := range cases {
		if got := isProgressLine(c.line); got != c.want {
			t.Errorf("isProgressLine(%q) = %v, want %v", c.line, got, c.want)
		}
	}
}

func TestRingBufferKeepsLastLines(t *testing.T) {
	r := newRingBuffer(3)

	for _, line := range []string{"one", "two", "three", "four", "five"} {
		r.Write([]byte(line + "\n"))
	}

	lines := r.Lines()
	if len(lines) != 3 {
		t.Fatalf("lines = %d, want 3", len(lines))
	}
	want := []string{"three", "four", "five"}
	for i, w := range want {
		if lines[i] != w {
			t.Errorf("line %d = %q, want %q", i, lines[i], w)
		}
	}
}

func TestRingBufferWriteReturnsLength(t *testing.T) {
	r := newRingBuffer(2)
	data := []byte("hello\nworld\n")

	n, err := r.Write(data)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(data) {
		t.Errorf("Write returned %d, want %d", n, len(data))
	}
	if len(r.Lines()) != 2 {
		t.Errorf("lines = %d, want 2", len(r.Lines()))
	}
}
