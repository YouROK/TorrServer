package timeindex

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

// The parsers sit in the path of every byte a client reads, so what matters is how fast they
// chew through data that holds nothing for them — the ordinary case, picture data between
// two timestamps.
func BenchmarkFeed(b *testing.B) {
	payload := make([]byte, 64<<10)
	rng := rand.New(rand.NewSource(1))
	rng.Read(payload)

	for _, name := range []string{"x.mkv", "x.ts", "x.mp4", "x.vob", "x.flv", "x.avi"} {
		b.Run(name, func(b *testing.B) {
			feeder := New(name).Feeder()
			b.SetBytes(int64(len(payload)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				feeder.Feed(int64(i)*int64(len(payload)), payload)
			}
		})
	}
}

// The same measurement over real files, where the parsers that have to find a packet
// boundary find one and then stride over the data instead of searching it.
func BenchmarkFeedRealFiles(b *testing.B) {
	dir := b.TempDir()
	buildSample(b, dir)

	for _, name := range []string{"src.mkv", "src.ts", "src.mp4", "src.vob", "src.flv", "src.avi"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			b.Fatal(err)
		}
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				feeder := New(name).Feeder()
				feedChunks(feeder, data, 0)
			}
		})
	}
}
