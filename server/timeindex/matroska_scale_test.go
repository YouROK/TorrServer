package timeindex

import (
	"os"
	"path/filepath"
	"testing"
)

// The three bytes that mark TimestampScale turn up inside frame data too, and the header
// region a stream is searched over runs well past the real header on a file whose first
// cluster starts a few kilobytes in. Measured on a real 4K episode: the true scale sat at byte
// 4156 and a stray match at byte 2851054 carried 12523939, which multiplied every timestamp in
// the film by twelve and a half. Here the stray is planted deliberately, after the real one.
func TestStrayTimestampScaleInFrameData(t *testing.T) {
	dir := t.TempDir()
	buildSample(t, dir)
	path := filepath.Join(dir, "src.mkv")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	packets, _ := groundTruth(t, path)

	// TimestampScale of 12523939, as found in the wild, planted well inside the picture data
	// but still within the stretch the header is looked for in.
	plant := []byte{0x2A, 0xD7, 0xB1, 0x83, 0xBF, 0x19, 0xA3}
	at := len(data) / 4
	if at >= headerReach {
		at = headerReach / 2
	}
	copy(data[at:], plant)

	ix := New(path)
	feeder := ix.Feeder()
	feedChunks(feeder, data, 0)

	for _, p := range packets {
		if p.pos < int64(at) {
			continue
		}
		got, ok := ix.TimeAt(p.pos)
		if !ok {
			continue
		}
		if got > p.sec+0.001 {
			t.Fatalf("offset %d: index says %.3fs, real time is %.3fs — the stray scale was believed",
				p.pos, got, p.sec)
		}
	}
}
