package subs

import "testing"

func BenchmarkSubsParse(b *testing.B) {
	files := []FileInfo{
		{Index: 1, Path: "Show/Season 1/episode.mkv", Size: 1 << 30},
		{Index: 2, Path: "Show/Season 1/episode.en.srt", Size: 4096},
		{Index: 3, Path: "Show/Subs/episode.ass", Size: 8192},
		{Index: 4, Path: "Show/Season 1/other.mkv", Size: 1 << 30},
	}
	srt := []byte("1\n00:00:01,000 --> 00:00:02,000\nHello\n\n2\n00:00:03,000 --> 00:00:04,000\nWorld\n")
	b.Run("match", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if len(Match(files[0], files)) == 0 {
				b.Fatal("no tracks")
			}
		}
	})
	b.Run("vtt", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(srt)))
		for i := 0; i < b.N; i++ {
			if _, err := ToVTT(srt, "srt"); err != nil {
				b.Fatal(err)
			}
		}
	})
}
