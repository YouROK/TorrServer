package torrshash

import (
	"strings"
	"testing"
)

func BenchmarkTorrshash(b *testing.B) {
	h := New(strings.Repeat("ab", 20))
	h.AddField(TagTitle, "Show Name")
	h.AddField(TagPoster, "http://img.example/poster.jpg")
	h.AddField(TagCategory, "tv")
	h.AddField(TagTracker, "udp://tracker.example:80/announce")
	h.AddField(TagSize, "123456789")
	packed, err := Pack(h)
	if err != nil || packed == "" {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(packed)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s, err := Pack(h)
		if err != nil {
			b.Fatal(err)
		}
		got, err := Unpack(s)
		if err != nil || got.Title() != "Show Name" || got.Hash != h.Hash {
			b.Fatalf("%v %+v", err, got)
		}
	}
}
